package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// CalendarView represents different calendar view modes
type CalendarView int

const (
	MonthView CalendarView = iota
	WeekView
)

// Calendar represents a calendar with todos
type Calendar struct {
	db      *DB
	date    time.Time
	view    CalendarView
	todos   []Todo
	todoMap map[string][]Todo // Key: YYYY-MM-DD, Value: todos for that day
}

// NewCalendar creates a new calendar instance
func NewCalendar(db *DB, date time.Time, view CalendarView) *Calendar {
	return &Calendar{
		db:      db,
		date:    date,
		view:    view,
		todoMap: make(map[string][]Todo),
	}
}

// LoadTodos loads todos for the calendar period
func (c *Calendar) LoadTodos() error {
	var todos []Todo
	var err error

	switch c.view {
	case MonthView:
		todos, err = c.db.GetMonthTodos(c.date)
	case WeekView:
		startOfWeek := c.getStartOfWeek(c.date)
		endOfWeek := startOfWeek.AddDate(0, 0, 6)
		todos, err = c.db.GetRangeTodos(startOfWeek, endOfWeek)
	}

	if err != nil {
		return err
	}

	c.todos = todos
	c.buildTodoMap()
	return nil
}

// buildTodoMap organizes todos by date
func (c *Calendar) buildTodoMap() {
	c.todoMap = make(map[string][]Todo)
	for _, todo := range c.todos {
		if todo.ScheduledStart != nil {
			dateKey := todo.ScheduledStart.Format("2006-01-02")
			c.todoMap[dateKey] = append(c.todoMap[dateKey], todo)
		}
	}
}

// GetDate returns the current date of the calendar
func (c *Calendar) GetDate() time.Time {
	return c.date
}

// GetView returns the current view mode of the calendar
func (c *Calendar) GetView() CalendarView {
	return c.view
}

// Render displays the calendar
func (c *Calendar) Render() string {
	switch c.view {
	case MonthView:
		return c.renderMonth()
	case WeekView:
		return c.renderWeek()
	default:
		return "Unknown calendar view"
	}
}

// renderMonth renders a monthly calendar view
func (c *Calendar) renderMonth() string {
	var s strings.Builder

	// Calendar title
	title := fmt.Sprintf("📅 %s", c.date.Format("January 2006"))
	s.WriteString(titleStyle.Render(title))
	s.WriteString("\n\n")

	// Get first day of month and calculate padding
	firstDay := time.Date(c.date.Year(), c.date.Month(), 1, 0, 0, 0, 0, c.date.Location())
	startDay := c.getStartOfWeek(firstDay)

	// Days of week header
	weekdays := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorBlue)).
		Bold(true).
		Width(10).
		Align(lipgloss.Center)

	for _, day := range weekdays {
		s.WriteString(headerStyle.Render(day))
	}
	s.WriteString("\n")

	// Calendar grid
	current := startDay
	for week := 0; week < 6; week++ {
		weekEmpty := true
		var weekLine strings.Builder

		for day := 0; day < 7; day++ {
			nextMonth := c.date.AddDate(0, 1, 0).Month()
			if current.Month() == c.date.Month() || week < 1 || (week == 5 && current.Month() == nextMonth) {
				weekEmpty = false
			}

			dayStr := c.renderDay(current)
			weekLine.WriteString(dayStr)
			current = current.AddDate(0, 0, 1)
		}

		if !weekEmpty || week < 4 {
			s.WriteString(weekLine.String())
			s.WriteString("\n")
		}

		// Stop if we've gone past the month and filled at least 4 weeks
		if week >= 3 && current.Month() != c.date.Month() {
			break
		}
	}

	// Todo list for the month
	if len(c.todos) > 0 {
		s.WriteString("\n")
		s.WriteString(titleStyle.Render("Scheduled for " + c.date.Format("January 2006") + ":"))
		s.WriteString("\n\n")

		currentDate := ""
		for _, todo := range c.todos {
			if todo.ScheduledStart != nil {
				todoDate := todo.ScheduledStart.Format("Jan 2")
				if todoDate != currentDate {
					currentDate = todoDate
					s.WriteString(lipgloss.NewStyle().
						Foreground(lipgloss.Color(ColorBlue)).
						Bold(true).
						Render(todoDate + ":"))
					s.WriteString("\n")
				}

				timeBlock := FormatTimeBlock(todo.ScheduledStart, todo.ScheduledEnd)
				timeStr := ""
				if timeBlock != "" {
					timeStr = strings.TrimPrefix(timeBlock, "Scheduled: ")
				}

				todoText := fmt.Sprintf("  • %s", todo.Title)
				if timeStr != "" {
					todoText += fmt.Sprintf(" (%s)", timeStr)
				}

				if todo.Done {
					todoText = completedStyle.Render(todoText)
				}

				s.WriteString(todoText)
				s.WriteString("\n")
			}
		}
	}

	return s.String()
}

// renderWeek renders a weekly calendar view
func (c *Calendar) renderWeek() string {
	var s strings.Builder

	startOfWeek := c.getStartOfWeek(c.date)
	endOfWeek := startOfWeek.AddDate(0, 0, 6)

	// Week title
	title := fmt.Sprintf("📅 Week of %s - %s",
		startOfWeek.Format("Jan 2"),
		endOfWeek.Format("Jan 2, 2006"))
	s.WriteString(titleStyle.Render(title))
	s.WriteString("\n\n")

	// Render each day of the week
	current := startOfWeek
	for i := 0; i < 7; i++ {
		dayName := current.Format("Monday")
		dayDate := current.Format("Jan 2")

		isToday := current.Year() == time.Now().Year() &&
			current.YearDay() == time.Now().YearDay()

		dayTitle := fmt.Sprintf("%s, %s", dayName, dayDate)
		if isToday {
			dayTitle += " (Today)"
		}

		dayStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorBlue)).
			Bold(true)
		if isToday {
			dayStyle = dayStyle.Foreground(lipgloss.Color(ColorYellow))
		}

		s.WriteString(dayStyle.Render(dayTitle))
		s.WriteString("\n")

		// Show todos for this day
		dateKey := current.Format("2006-01-02")
		if dayTodos, exists := c.todoMap[dateKey]; exists {
			for _, todo := range dayTodos {
				timeBlock := FormatTimeBlock(todo.ScheduledStart, todo.ScheduledEnd)
				timeStr := ""
				if timeBlock != "" {
					timeStr = strings.TrimPrefix(timeBlock, "Scheduled: ")
					// Remove date prefix since we're already showing the date
					parts := strings.Split(timeStr, " ")
					if len(parts) > 2 {
						timeStr = strings.Join(parts[2:], " ")
					}
				}

				todoText := fmt.Sprintf("  • %s", todo.Title)
				if timeStr != "" {
					todoText += fmt.Sprintf(" (%s)", timeStr)
				}

				if todo.Done {
					todoText = completedStyle.Render(todoText)
				}

				s.WriteString(todoText)
				s.WriteString("\n")
			}
		} else {
			s.WriteString(descStyle.Render("  No todos scheduled"))
			s.WriteString("\n")
		}

		s.WriteString("\n")
		current = current.AddDate(0, 0, 1)
	}

	return s.String()
}

// renderDay renders a single day cell for the month view
func (c *Calendar) renderDay(date time.Time) string {
	dayNum := date.Day()
	dateKey := date.Format("2006-01-02")

	isCurrentMonth := date.Month() == c.date.Month()
	isToday := date.Year() == time.Now().Year() &&
		date.YearDay() == time.Now().YearDay()
	hasTodos := len(c.todoMap[dateKey]) > 0

	// Base style
	cellStyle := lipgloss.NewStyle().
		Width(10).
		Height(1).
		Align(lipgloss.Center)

	dayText := fmt.Sprintf("%2d", dayNum)

	if !isCurrentMonth {
		// Grayed out for other months
		cellStyle = cellStyle.Foreground(lipgloss.Color(ColorGray))
	} else if isToday {
		// Highlight today
		cellStyle = cellStyle.
			Foreground(lipgloss.Color(ColorWhite)).
			Background(lipgloss.Color(ColorYellow)).
			Bold(true)
	} else if hasTodos {
		// Highlight days with todos
		cellStyle = cellStyle.
			Foreground(lipgloss.Color(ColorBlue)).
			Bold(true)
		dayText += " •"
	} else {
		// Regular day
		cellStyle = cellStyle.Foreground(lipgloss.Color(ColorWhite))
	}

	return cellStyle.Render(dayText)
}

// getStartOfWeek returns the start of the week (Sunday) for a given date
func (c *Calendar) getStartOfWeek(date time.Time) time.Time {
	weekday := int(date.Weekday())
	start := date.AddDate(0, 0, -weekday)
	return time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
}

// Next moves to the next period (month or week)
func (c *Calendar) Next() {
	switch c.view {
	case MonthView:
		c.date = c.date.AddDate(0, 1, 0)
	case WeekView:
		c.date = c.date.AddDate(0, 0, 7)
	}
}

// Previous moves to the previous period (month or week)
func (c *Calendar) Previous() {
	switch c.view {
	case MonthView:
		c.date = c.date.AddDate(0, -1, 0)
	case WeekView:
		c.date = c.date.AddDate(0, 0, -7)
	}
}
