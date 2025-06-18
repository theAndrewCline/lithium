package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// CLI handles all command-line interface operations
type CLI struct {
	db *DB
}

// NewCLI creates a new CLI instance
func NewCLI(db *DB) *CLI {
	return &CLI{db: db}
}

// HandleCommand processes the given command and arguments
func (c *CLI) HandleCommand(command string, args []string) {
	switch command {
	case "add", "a":
		c.handleAdd(args)
	case "list", "ls", "l":
		c.handleList()
	case "inbox", "i":
		c.handleInbox()
	case "today", "tod":
		c.handleToday()
	case "day", "date":
		c.handleDate(args)
	case "calendar", "cal":
		c.handleCalendar(args)
	case "toggle", "t":
		c.handleToggle(args)
	case "delete", "del", "d":
		c.handleDelete(args)
	case "edit", "e":
		c.handleEdit(args)
	case "schedule", "s":
		c.handleSchedule(args)
	case "ui":
		c.handleUI()
	case "help", "h":
		c.printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		c.printUsage()
	}
}

func (c *CLI) handleAdd(args []string) {
	if len(args) == 0 {
		fmt.Println(errorStyle.Render("Error: Title is required"))
		fmt.Println(styleCommand("Usage: li add <title> [description]"))
		return
	}

	title := args[0]
	description := ""
	if len(args) > 1 {
		description = strings.Join(args[1:], " ")
	}

	err := c.db.AddTodo(title, description, nil, nil, nil)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error adding todo: %v", err)))
		return
	}

	fmt.Println(successStyle.Render(fmt.Sprintf("✅ Added todo: %s", title)))
}

func (c *CLI) handleInbox() {
	todos, err := c.db.GetInboxTodos()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error listing inbox: %v", err)))
		return
	}

	c.renderTodoList(todos, "📥 Inbox (Unscheduled):", "No unscheduled todos! Everything is planned. ✅")
}

func (c *CLI) handleToday() {
	c.handleDate([]string{"today"})
}

func (c *CLI) handleDate(args []string) {
	var targetDate time.Time
	var title string
	var emptyMessage string
	
	if len(args) == 0 || args[0] == "today" {
		targetDate = time.Now()
		title = "📅 Today's Schedule:"
		emptyMessage = "Nothing scheduled for today. Use " + styleCommand("li schedule <id> \"<time>\"") + " to plan your day."
	} else {
		// Parse the date argument
		dateStr := strings.Join(args, " ")
		
		// Handle relative dates
		switch strings.ToLower(dateStr) {
		case "tomorrow":
			targetDate = time.Now().AddDate(0, 0, 1)
			title = "📅 Tomorrow's Schedule:"
			emptyMessage = "Nothing scheduled for tomorrow. Use " + styleCommand("li schedule <id> \"<time>\"") + " to plan ahead."
		case "yesterday":
			targetDate = time.Now().AddDate(0, 0, -1)
			title = "📅 Yesterday's Schedule:"
			emptyMessage = "Nothing was scheduled for yesterday."
		default:
			// Try to parse as a regular date
			parsedDate, err := parseScheduleDate(dateStr)
			if err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("Error parsing date: %v", err)))
				fmt.Println("Examples: today, tomorrow, yesterday, Monday, Dec 25, 2024-12-25")
				return
			}
			targetDate = *parsedDate
			title = fmt.Sprintf("📅 Schedule for %s:", targetDate.Format("Jan 2, 2006"))
			emptyMessage = fmt.Sprintf("Nothing scheduled for %s.", targetDate.Format("Jan 2, 2006"))
		}
	}
	
	todos, err := c.db.GetDateTodos(targetDate)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error listing todos for %s: %v", targetDate.Format("Jan 2"), err)))
		return
	}

	c.renderTodoList(todos, title, emptyMessage)
}

// parseScheduleDate parses various date formats for schedule queries
func parseScheduleDate(dateStr string) (*time.Time, error) {
	now := time.Now()
	dateStr = strings.TrimSpace(strings.ToLower(dateStr))
	
	// Handle weekdays
	weekdays := map[string]time.Weekday{
		"sunday": time.Sunday, "sun": time.Sunday,
		"monday": time.Monday, "mon": time.Monday,
		"tuesday": time.Tuesday, "tue": time.Tuesday, "tues": time.Tuesday,
		"wednesday": time.Wednesday, "wed": time.Wednesday,
		"thursday": time.Thursday, "thu": time.Thursday, "thur": time.Thursday, "thurs": time.Thursday,
		"friday": time.Friday, "fri": time.Friday,
		"saturday": time.Saturday, "sat": time.Saturday,
	}
	
	if weekday, exists := weekdays[dateStr]; exists {
		daysUntil := (int(weekday) - int(now.Weekday()) + 7) % 7
		if daysUntil == 0 {
			daysUntil = 7 // Next week if it's the same day
		}
		date := now.AddDate(0, 0, daysUntil)
		return &date, nil
	}
	
	// Try standard date parsing from ParseDueDate
	return ParseDueDate(dateStr)
}

func (c *CLI) handleCalendar(args []string) {
	var targetDate time.Time
	var view CalendarView

	// Default to current month
	targetDate = time.Now()
	view = MonthView

	// Parse arguments
	if len(args) > 0 {
		switch strings.ToLower(args[0]) {
		case "week", "w":
			view = WeekView
			if len(args) > 1 {
				// Parse date for week
				if parsedDate, err := parseScheduleDate(strings.Join(args[1:], " ")); err == nil {
					targetDate = *parsedDate
				}
			}
		case "month", "m":
			view = MonthView
			if len(args) > 1 {
				// Parse date for month
				if parsedDate, err := parseScheduleDate(strings.Join(args[1:], " ")); err == nil {
					targetDate = *parsedDate
				}
			}
		default:
			// Assume it's a date for month view
			if parsedDate, err := parseScheduleDate(strings.Join(args, " ")); err == nil {
				targetDate = *parsedDate
				view = MonthView
			} else {
				fmt.Println(errorStyle.Render(fmt.Sprintf("Error parsing date: %v", err)))
				fmt.Println("Usage:")
				fmt.Println("  " + styleCommand("li calendar") + " - Show current month")
				fmt.Println("  " + styleCommand("li calendar week") + " - Show current week")
				fmt.Println("  " + styleCommand("li calendar month Dec") + " - Show December")
				fmt.Println("  " + styleCommand("li calendar week Monday") + " - Show week containing Monday")
				return
			}
		}
	}

	// Create and load calendar
	calendar := NewCalendar(c.db, targetDate, view)
	err := calendar.LoadTodos()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error loading calendar: %v", err)))
		return
	}

	// Render calendar
	fmt.Print(calendar.Render())
}

func (c *CLI) handleList() {
	todos, err := c.db.GetAllTodos()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error listing todos: %v", err)))
		return
	}

	if len(todos) == 0 {
		fmt.Println(descStyle.Render("No todos found. Add one with: ") + styleCommand("li add <title>"))
		return
	}

	fmt.Println(titleStyle.Render("⚡ Your Todos:"))
	fmt.Println()

	for _, todo := range todos {
		status, statusColor := createStatusStyle(todo.Done)

		id := idStyle.Render(fmt.Sprintf("[%d]", todo.ID))
		statusStyled := lipgloss.NewStyle().Foreground(statusColor).Render(status)

		todoText := todo.Title
		if todo.Done {
			todoText = completedStyle.Render(todoText)
		}

		line := fmt.Sprintf("%s %s %s", id, statusStyled, todoText)

		if todo.Description != "" {
			descText := todo.Description
			if todo.Done {
				descText = completedStyle.Render(descText)
			} else {
				descText = descStyle.Render(descText)
			}
			line += fmt.Sprintf(" - %s", descText)
		}

		// Add time block info if scheduled
		if timeBlock := FormatTimeBlock(todo.ScheduledStart, todo.ScheduledEnd); timeBlock != "" {
			timeBlockStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue)).Render(fmt.Sprintf(" [%s]", timeBlock))
			line += timeBlockStyled
		}

		fmt.Println(todoStyle.Render(line))
	}
}

func (c *CLI) handleToggle(args []string) {
	if len(args) == 0 {
		fmt.Println(errorStyle.Render("Error: Todo ID is required"))
		fmt.Println(styleCommand("Usage: li toggle <id>"))
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: Invalid ID '%s'. Must be a number.", args[0])))
		return
	}

	err = c.db.ToggleTodo(id)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error toggling todo: %v", err)))
		return
	}

	fmt.Println(successStyle.Render(fmt.Sprintf("✅ Toggled todo %d", id)))
}

func (c *CLI) handleDelete(args []string) {
	if len(args) == 0 {
		fmt.Println(errorStyle.Render("Error: Todo ID is required"))
		fmt.Println(styleCommand("Usage: li delete <id>"))
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: Invalid ID '%s'. Must be a number.", args[0])))
		return
	}

	err = c.db.DeleteTodo(id)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error deleting todo: %v", err)))
		return
	}

	fmt.Println(successStyle.Render(fmt.Sprintf("🗑️  Deleted todo %d", id)))
}

func (c *CLI) handleEdit(args []string) {
	if len(args) < 2 {
		fmt.Println(errorStyle.Render("Error: Todo ID and title are required"))
		fmt.Println(styleCommand("Usage: li edit <id> <title> [description]"))
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: Invalid ID '%s'. Must be a number.", args[0])))
		return
	}

	title := args[1]
	description := ""
	if len(args) > 2 {
		description = strings.Join(args[2:], " ")
	}

	err = c.db.UpdateTodo(id, title, description, nil, nil, nil)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error updating todo: %v", err)))
		return
	}

	fmt.Println(successStyle.Render(fmt.Sprintf("✏️  Updated todo %d: %s", id, title)))
}

func (c *CLI) handleSchedule(args []string) {
	if len(args) < 2 {
		fmt.Println(errorStyle.Render("Error: Todo ID and time block are required"))
		fmt.Println(styleCommand("Usage: li schedule <id> \"<time block>\""))
		fmt.Println("Examples:")
		fmt.Println("  " + styleCommand("li schedule 1 \"Monday 2pm-4pm\""))
		fmt.Println("  " + styleCommand("li schedule 1 \"tomorrow 9am for 2 hours\""))
		fmt.Println("  " + styleCommand("li schedule 1 \"Dec 25 10am-12pm\""))
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: Invalid ID '%s'. Must be a number.", args[0])))
		return
	}

	timeBlockStr := strings.Join(args[1:], " ")
	timeBlock, err := ParseTimeBlock(timeBlockStr)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error parsing time block: %v", err)))
		fmt.Println("Examples of valid time blocks:")
		fmt.Println("  \"Monday 2pm-4pm\"")
		fmt.Println("  \"tomorrow 9am for 2 hours\"")
		fmt.Println("  \"Dec 25 10am-12pm\"")
		return
	}

	err = c.db.ScheduleTodo(id, timeBlock.Start, timeBlock.End)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error scheduling todo: %v", err)))
		return
	}

	fmt.Println(successStyle.Render(fmt.Sprintf("📅 Scheduled todo %d: %s", id, FormatTimeBlock(timeBlock.Start, timeBlock.End))))
}

func (c *CLI) handleUI() {
	fmt.Println(titleStyle.Render("🚀 Launching TUI mode..."))
	err := RunTUI(c.db)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error running TUI: %v", err)))
	}
}

func (c *CLI) printUsage() {
	fmt.Println(titleStyle.Render("⚡ Lithium"))
	fmt.Println()
	fmt.Println(commandStyle.Render("Usage:"))

	commands := [][]string{
		{"li add <title> [description]", "Add a new todo"},
		{"li list", "List all todos"},
		{"li inbox", "List unscheduled todos"},
		{"li today", "List today's scheduled todos"},
		{"li day <date>", "List todos for a specific date"},
		{"li calendar [month|week] [date]", "Show calendar view"},
		{"li toggle <id>", "Toggle todo completion"},
		{"li delete <id>", "Delete a todo"},
		{"li edit <id> <title> [description]", "Edit a todo"},
		{"li schedule <id> \"<time block>\"", "Schedule a time block for a todo"},
		{"li ui", "Launch interactive TUI mode"},
		{"li help", "Show this help"},
	}

	for _, cmd := range commands {
		// Style the command with orange variables
		styledCmd := styleCommand(cmd[0])

		// Calculate spacing based on the original length (without styling)
		spaces := 40 - len(cmd[0])
		if spaces < 1 {
			spaces = 1
		}

		fmt.Printf("  %s %s %s\n",
			styledCmd,
			strings.Repeat(" ", spaces),
			descStyle.Render(cmd[1]))
	}

	fmt.Println()
	fmt.Println(commandStyle.Render("Aliases:"))
	aliases := [][]string{
		{"a, add", "l, ls, list", "i, inbox", "tod, today", "day, date", "cal, calendar", "t, toggle", "d, del, delete", "e, edit", "s, schedule"},
	}

	for _, aliasGroup := range aliases {
		fmt.Print("  ")
		for i, alias := range aliasGroup {
			if i > 0 {
				fmt.Print("     ")
			}
			fmt.Print(descStyle.Render(alias))
		}
		fmt.Println()
	}
}

func (c *CLI) renderTodoList(todos []Todo, title, emptyMessage string) {
	if len(todos) == 0 {
		fmt.Println(descStyle.Render(emptyMessage))
		return
	}

	fmt.Println(titleStyle.Render(title))
	fmt.Println()
	
	for _, todo := range todos {
		status, statusColor := createStatusStyle(todo.Done)

		id := idStyle.Render(fmt.Sprintf("[%d]", todo.ID))
		statusStyled := lipgloss.NewStyle().Foreground(statusColor).Render(status)
		
		todoText := todo.Title
		if todo.Done {
			todoText = completedStyle.Render(todoText)
		}

		line := fmt.Sprintf("%s %s %s", id, statusStyled, todoText)

		if todo.Description != "" {
			descText := todo.Description
			if todo.Done {
				descText = completedStyle.Render(descText)
			} else {
				descText = descStyle.Render(descText)
			}
			line += fmt.Sprintf(" - %s", descText)
		}

		// Add time block info if scheduled
		if timeBlock := FormatTimeBlock(todo.ScheduledStart, todo.ScheduledEnd); timeBlock != "" {
			timeBlockStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue)).Render(fmt.Sprintf(" [%s]", timeBlock))
			line += timeBlockStyled
		}

		fmt.Println(todoStyle.Render(line))
	}
}
