package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
)

type tuiState int

const (
	tuiTodayView tuiState = iota
	tuiInboxView
	tuiCalendarView
	tuiCaptureView
	tuiAddView
	tuiEditView
)

type tuiModel struct {
	db             *DB
	todos          []Todo
	cursor         int
	state          tuiState
	previousState  tuiState
	input          string
	inputDesc      string
	inputDue       string
	inputScheduled string
	editingID      int
	err            error
	width          int
	height         int
	inputField     int // 0: title, 1: description, 2: due date, 3: scheduled time
	calendar       *Calendar
	keys           keyMap
	help           help.Model
}

type keyMap struct {
	Today    key.Binding
	Inbox    key.Binding
	Calendar key.Binding
	Capture  key.Binding
	Up       key.Binding
	Down     key.Binding
	Toggle   key.Binding
	New      key.Binding
	Edit     key.Binding
	Delete   key.Binding
	Quit     key.Binding
}

var defaultKeyMap = keyMap{
	Today: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "today"),
	),
	Inbox: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "inbox"),
	),
	Calendar: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "calendar"),
	),
	Capture: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "capture"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Toggle: key.NewBinding(
		key.WithKeys(" ", "enter"),
		key.WithHelp("space/enter", "toggle"),
	),
	New: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new todo"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Toggle, k.New, k.Today, k.Inbox, k.Calendar, k.Capture, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Toggle, k.New, k.Edit, k.Delete},
		{k.Today, k.Inbox, k.Calendar, k.Capture, k.Quit},
	}
}

func NewTuiModel(db *DB) tuiModel {
	todos, err := db.GetTodayTodos()
	if err != nil {
		return tuiModel{db: db, err: err}
	}

	return tuiModel{
		db:       db,
		todos:    todos,
		state:    tuiTodayView,
		calendar: NewCalendar(db, time.Now(), MonthView),
		keys:     defaultKeyMap,
		help:     help.New(),
	}
}

func (m tuiModel) Init() tea.Cmd {
	return nil
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		// Handle global navigation keys first
		switch {
		case key.Matches(msg, m.keys.Today) && m.state != tuiAddView && m.state != tuiEditView && m.state != tuiCaptureView:
			m.state = tuiTodayView
			todos, _ := m.db.GetTodayTodos()
			m.todos = todos
			m.cursor = 0
			return m, nil
		case key.Matches(msg, m.keys.Inbox) && m.state != tuiAddView && m.state != tuiEditView && m.state != tuiCaptureView:
			m.state = tuiInboxView
			todos, _ := m.db.GetInboxTodos()
			m.todos = todos
			m.cursor = 0
			return m, nil
		case key.Matches(msg, m.keys.Calendar) && m.state != tuiAddView && m.state != tuiEditView && m.state != tuiCaptureView:
			m.state = tuiCalendarView
			return m, nil
		case key.Matches(msg, m.keys.Capture) && m.state != tuiAddView && m.state != tuiEditView:
			m.state = tuiCaptureView
			m.input = ""
			return m, nil
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}

		// Handle state-specific keys
		switch m.state {
		case tuiInboxView:
			return m.updateInbox(msg)
		case tuiTodayView:
			return m.updateToday(msg)
		case tuiCalendarView:
			return m.updateCalendar(msg)
		case tuiCaptureView:
			return m.updateCapture(msg)
		case tuiAddView:
			return m.updateAdd(msg)
		case tuiEditView:
			return m.updateEdit(msg)
		}
	}
	return m, nil
}

func (m tuiModel) updateInbox(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.todos)-1 {
			m.cursor++
		}
	case "n":
		m.previousState = m.state
		m.state = tuiAddView
		m.input = ""
		m.inputDesc = ""
		m.inputDue = ""
		m.inputScheduled = ""
		m.inputField = 0
	case "enter", " ":
		if len(m.todos) > 0 {
			todo := m.todos[m.cursor]
			m.db.ToggleTodo(todo.ID)
			todos, _ := m.db.GetInboxTodos()
			m.todos = todos
		}
	case "d":
		if len(m.todos) > 0 {
			todo := m.todos[m.cursor]
			m.db.DeleteTodo(todo.ID)
			todos, _ := m.db.GetInboxTodos()
			m.todos = todos
			if m.cursor >= len(m.todos) && len(m.todos) > 0 {
				m.cursor = len(m.todos) - 1
			}
		}
	case "e":
		if len(m.todos) > 0 {
			todo := m.todos[m.cursor]
			m.previousState = m.state
			m.state = tuiEditView
			m.editingID = todo.ID
			m.input = todo.Title
			m.inputDesc = todo.Description

			// Format current due date and scheduled time for editing
			m.inputDue = ""
			if todo.DueDate != nil {
				m.inputDue = todo.DueDate.Format("2006-01-02")
			}

			m.inputScheduled = ""
			if todo.ScheduledStart != nil {
				if todo.ScheduledEnd != nil {
					m.inputScheduled = fmt.Sprintf("%s-%s",
						todo.ScheduledStart.Format("3:04pm"),
						todo.ScheduledEnd.Format("3:04pm"))
				} else {
					m.inputScheduled = todo.ScheduledStart.Format("3:04pm")
				}
			}

			m.inputField = 0
		}
	}
	return m, nil
}

func (m tuiModel) updateToday(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.todos)-1 {
			m.cursor++
		}
	case "n":
		m.previousState = m.state
		m.state = tuiAddView
		m.input = ""
		m.inputDesc = ""
		m.inputDue = ""
		m.inputScheduled = ""
		m.inputField = 0
	case "enter", " ":
		if len(m.todos) > 0 {
			todo := m.todos[m.cursor]
			m.db.ToggleTodo(todo.ID)
			todos, _ := m.db.GetTodayTodos()
			m.todos = todos
		}
	case "d":
		if len(m.todos) > 0 {
			todo := m.todos[m.cursor]
			m.db.DeleteTodo(todo.ID)
			todos, _ := m.db.GetTodayTodos()
			m.todos = todos
			if m.cursor >= len(m.todos) && len(m.todos) > 0 {
				m.cursor = len(m.todos) - 1
			}
		}
	case "e":
		if len(m.todos) > 0 {
			todo := m.todos[m.cursor]
			m.previousState = m.state
			m.state = tuiEditView
			m.editingID = todo.ID
			m.input = todo.Title
			m.inputDesc = todo.Description

			// Format current due date and scheduled time for editing
			m.inputDue = ""
			if todo.DueDate != nil {
				m.inputDue = todo.DueDate.Format("2006-01-02")
			}

			m.inputScheduled = ""
			if todo.ScheduledStart != nil {
				if todo.ScheduledEnd != nil {
					m.inputScheduled = fmt.Sprintf("%s-%s",
						todo.ScheduledStart.Format("3:04pm"),
						todo.ScheduledEnd.Format("3:04pm"))
				} else {
					m.inputScheduled = todo.ScheduledStart.Format("3:04pm")
				}
			}

			m.inputField = 0
		}
	}
	return m, nil
}

func (m tuiModel) updateCalendar(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		// Navigate to previous month
		m.calendar = NewCalendar(m.db, m.calendar.GetDate().AddDate(0, -1, 0), m.calendar.GetView())
		m.calendar.LoadTodos()
	case "right", "l":
		// Navigate to next month
		m.calendar = NewCalendar(m.db, m.calendar.GetDate().AddDate(0, 1, 0), m.calendar.GetView())
		m.calendar.LoadTodos()
	case "m":
		m.calendar = NewCalendar(m.db, m.calendar.GetDate(), MonthView)
		m.calendar.LoadTodos()
	case "w":
		m.calendar = NewCalendar(m.db, m.calendar.GetDate(), WeekView)
		m.calendar.LoadTodos()
	}
	return m, nil
}

func (m tuiModel) updateCapture(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		m.state = tuiInboxView
		todos, _ := m.db.GetInboxTodos()
		m.todos = todos
		m.cursor = 0
		return m, nil
	case "enter":
		if strings.TrimSpace(m.input) != "" {
			// Parse the input - look for " -- " separator for description
			parts := strings.SplitN(m.input, " -- ", 2)
			title := strings.TrimSpace(parts[0])
			desc := ""
			if len(parts) > 1 {
				desc = strings.TrimSpace(parts[1])
			}

			// Add todo to inbox (no scheduling)
			m.db.AddTodo(title, desc, nil, nil, nil)
			m.input = "" // Clear for next todo
		}
	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	default:
		if len(msg.String()) == 1 { // Only add printable characters
			m.input += msg.String()
		}
	}
	return m, nil
}

// returnToPreviousState returns to the state before entering add/edit mode
func (m *tuiModel) returnToPreviousState() {
	m.state = m.previousState
	switch m.previousState {
	case tuiTodayView:
		todos, _ := m.db.GetTodayTodos()
		m.todos = todos
	case tuiInboxView:
		todos, _ := m.db.GetInboxTodos()
		m.todos = todos
	case tuiCalendarView:
		// Calendar doesn't need todo reloading since it manages its own data
	}
}

func (m tuiModel) updateAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		// Return to the previous view
		m.returnToPreviousState()
	case "enter":
		if strings.TrimSpace(m.input) != "" {
			// Parse due date and scheduled time
			var dueDate, scheduledStart, scheduledEnd *time.Time

			if m.inputDue != "" {
				if parsed, err := ParseDueDate(m.inputDue); err == nil {
					dueDate = parsed
				}
			}

			if m.inputScheduled != "" {
				if timeBlock, err := ParseTimeBlock(m.inputScheduled); err == nil && timeBlock != nil {
					scheduledStart = timeBlock.Start
					scheduledEnd = timeBlock.End
				}
			}

			m.db.AddTodo(m.input, m.inputDesc, dueDate, scheduledStart, scheduledEnd)
			// Return to previous view after adding
			m.returnToPreviousState()
		}
	case "tab":
		m.inputField = (m.inputField + 1) % 4
	case "backspace":
		switch m.inputField {
		case 0:
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		case 1:
			if len(m.inputDesc) > 0 {
				m.inputDesc = m.inputDesc[:len(m.inputDesc)-1]
			}
		case 2:
			if len(m.inputDue) > 0 {
				m.inputDue = m.inputDue[:len(m.inputDue)-1]
			}
		case 3:
			if len(m.inputScheduled) > 0 {
				m.inputScheduled = m.inputScheduled[:len(m.inputScheduled)-1]
			}
		}
	default:
		if len(msg.String()) == 1 { // Only add printable characters
			switch m.inputField {
			case 0:
				m.input += msg.String()
			case 1:
				m.inputDesc += msg.String()
			case 2:
				m.inputDue += msg.String()
			case 3:
				m.inputScheduled += msg.String()
			}
		}
	}
	return m, nil
}

func (m tuiModel) updateEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		// Return to previous view
		m.returnToPreviousState()
	case "enter":
		if strings.TrimSpace(m.input) != "" {
			// Parse due date and scheduled time
			var dueDate, scheduledStart, scheduledEnd *time.Time

			if m.inputDue != "" {
				if parsed, err := ParseDueDate(m.inputDue); err == nil {
					dueDate = parsed
				}
			}

			if m.inputScheduled != "" {
				if timeBlock, err := ParseTimeBlock(m.inputScheduled); err == nil && timeBlock != nil {
					scheduledStart = timeBlock.Start
					scheduledEnd = timeBlock.End
				}
			}

			m.db.UpdateTodo(m.editingID, m.input, m.inputDesc, dueDate, scheduledStart, scheduledEnd)
			// Return to previous view after editing
			m.returnToPreviousState()
		}
	case "tab":
		m.inputField = (m.inputField + 1) % 4
	case "backspace":
		switch m.inputField {
		case 0:
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		case 1:
			if len(m.inputDesc) > 0 {
				m.inputDesc = m.inputDesc[:len(m.inputDesc)-1]
			}
		case 2:
			if len(m.inputDue) > 0 {
				m.inputDue = m.inputDue[:len(m.inputDue)-1]
			}
		case 3:
			if len(m.inputScheduled) > 0 {
				m.inputScheduled = m.inputScheduled[:len(m.inputScheduled)-1]
			}
		}
	default:
		if len(msg.String()) == 1 { // Only add printable characters
			switch m.inputField {
			case 0:
				m.input += msg.String()
			case 1:
				m.inputDesc += msg.String()
			case 2:
				m.inputDue += msg.String()
			case 3:
				m.inputScheduled += msg.String()
			}
		}
	}
	return m, nil
}

func (m tuiModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	var content string
	switch m.state {
	case tuiInboxView:
		content = m.viewInbox()
	case tuiTodayView:
		content = m.viewToday()
	case tuiCalendarView:
		content = m.viewCalendar()
	case tuiCaptureView:
		return m.viewCapture() // This view has its own help
	case tuiAddView:
		return m.viewAdd() // These views have their own help
	case tuiEditView:
		return m.viewEdit() // These views have their own help
	}

	// Add help at bottom for main views
	helpView := m.help.View(m.keys)
	return content + "\n" + helpView
}

// renderTabHeader creates a tab-style header for navigation
func (m tuiModel) renderTabHeader() string {
	var tabs []string

	todayTab := "📅 Today"
	inboxTab := "📥 Inbox"
	calendarTab := "🗓️ Calendar"
	captureTab := "🎯 Capture"

	// Highlight active tab
	switch m.state {
	case tuiTodayView:
		todayTab = tuiSelectedStyle.Render(todayTab)
	case tuiInboxView:
		inboxTab = tuiSelectedStyle.Render(inboxTab)
	case tuiCalendarView:
		calendarTab = tuiSelectedStyle.Render(calendarTab)
	case tuiCaptureView:
		captureTab = tuiSelectedStyle.Render(captureTab)
	}

	tabs = append(tabs, todayTab, inboxTab, calendarTab, captureTab)

	header := strings.Join(tabs, " | ")
	return tuiTitleStyle.Render("⚡ Lithium") + "\n" + header + "\n\n"
}

func (m tuiModel) viewAdd() string {
	var s strings.Builder

	s.WriteString(tuiTitleStyle.Render("⚡ Add New Todo"))
	s.WriteString("\n\n")

	// Title field with cursor indicator
	titleLabel := tuiLabelStyle.Render("Title: ")
	titleValue := tuiInputStyle.Render(m.input)
	if m.inputField == 0 {
		titleValue += tuiInputStyle.Render("█") // Cursor
	}
	s.WriteString(fmt.Sprintf("%s%s\n", titleLabel, titleValue))

	// Description field with cursor indicator
	descLabel := tuiLabelStyle.Render("Description: ")
	descValue := tuiInputStyle.Render(m.inputDesc)
	if m.inputField == 1 {
		descValue += tuiInputStyle.Render("█") // Cursor
	}
	s.WriteString(fmt.Sprintf("%s%s\n", descLabel, descValue))

	// Due date field with cursor indicator
	dueLabel := tuiLabelStyle.Render("Due Date: ")
	dueValue := tuiInputStyle.Render(m.inputDue)
	if m.inputField == 2 {
		dueValue += tuiInputStyle.Render("█") // Cursor
	}
	s.WriteString(fmt.Sprintf("%s%s\n", dueLabel, dueValue))

	// Scheduled time field with cursor indicator
	schedLabel := tuiLabelStyle.Render("Scheduled: ")
	schedValue := tuiInputStyle.Render(m.inputScheduled)
	if m.inputField == 3 {
		schedValue += tuiInputStyle.Render("█") // Cursor
	}
	s.WriteString(fmt.Sprintf("%s%s\n", schedLabel, schedValue))

	s.WriteString(tuiHelpStyle.Render("\nTab: switch fields, Enter: save, Esc: cancel"))
	s.WriteString(tuiHelpStyle.Render("\nDue Date: today, tomorrow, 2024-12-25"))
	s.WriteString(tuiHelpStyle.Render("\nScheduled: today 2pm-4pm, Monday 9am for 2 hours"))

	return tuiContainerStyle.Render(s.String())
}

func (m tuiModel) viewEdit() string {
	var s strings.Builder

	s.WriteString(tuiTitleStyle.Render("⚡ Edit Todo"))
	s.WriteString("\n\n")

	// Title field with cursor indicator
	titleLabel := tuiLabelStyle.Render("Title: ")
	titleValue := tuiInputStyle.Render(m.input)
	if m.inputField == 0 {
		titleValue += tuiInputStyle.Render("█") // Cursor
	}
	s.WriteString(fmt.Sprintf("%s%s\n", titleLabel, titleValue))

	// Description field with cursor indicator
	descLabel := tuiLabelStyle.Render("Description: ")
	descValue := tuiInputStyle.Render(m.inputDesc)
	if m.inputField == 1 {
		descValue += tuiInputStyle.Render("█") // Cursor
	}
	s.WriteString(fmt.Sprintf("%s%s\n", descLabel, descValue))

	// Due date field with cursor indicator
	dueLabel := tuiLabelStyle.Render("Due Date: ")
	dueValue := tuiInputStyle.Render(m.inputDue)
	if m.inputField == 2 {
		dueValue += tuiInputStyle.Render("█") // Cursor
	}
	s.WriteString(fmt.Sprintf("%s%s\n", dueLabel, dueValue))

	// Scheduled time field with cursor indicator
	schedLabel := tuiLabelStyle.Render("Scheduled: ")
	schedValue := tuiInputStyle.Render(m.inputScheduled)
	if m.inputField == 3 {
		schedValue += tuiInputStyle.Render("█") // Cursor
	}
	s.WriteString(fmt.Sprintf("%s%s\n", schedLabel, schedValue))

	s.WriteString(tuiHelpStyle.Render("\nTab: switch fields, Enter: save, Esc: cancel"))
	s.WriteString(tuiHelpStyle.Render("\nDue Date: today, tomorrow, 2024-12-25"))
	s.WriteString(tuiHelpStyle.Render("\nScheduled: today 2pm-4pm, Monday 9am for 2 hours"))

	return tuiContainerStyle.Render(s.String())
}

func (m tuiModel) viewInbox() string {
	var s strings.Builder

	s.WriteString(m.renderTabHeader())

	if len(m.todos) == 0 {
		s.WriteString("No unscheduled todos. All todos are scheduled!\n")
	} else {
		for i, todo := range m.todos {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}

			status := "[ ]"
			title := todo.Title
			if todo.Done {
				status = "[x]"
				title = tuiDoneStyle.Render(title)
			}

			line := fmt.Sprintf("%s %s %s", cursor, status, title)
			if todo.Description != "" {
				desc := todo.Description
				if todo.Done {
					desc = tuiDoneStyle.Render(desc)
				} else {
					desc = descStyle.Render(desc)
				}
				line += fmt.Sprintf(" - %s", desc)
			}

			if m.cursor == i {
				line = tuiSelectedStyle.Render(line)
			}

			s.WriteString(line)
			s.WriteString("\n")
		}
	}

	return tuiContainerStyle.Render(s.String())
}

func (m tuiModel) viewToday() string {
	var s strings.Builder

	s.WriteString(m.renderTabHeader())

	if len(m.todos) == 0 {
		s.WriteString("No todos scheduled for today.\n")
	} else {
		for i, todo := range m.todos {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}

			status := "[ ]"
			title := todo.Title
			if todo.Done {
				status = "[x]"
				title = tuiDoneStyle.Render(title)
			}

			line := fmt.Sprintf("%s %s %s", cursor, status, title)

			// Show time if scheduled
			if todo.ScheduledStart != nil {
				timeStr := todo.ScheduledStart.Format("15:04")
				if todo.ScheduledEnd != nil {
					timeStr += "-" + todo.ScheduledEnd.Format("15:04")
				}
				line += fmt.Sprintf(" [%s]", timeStr)
			}

			if todo.Description != "" {
				desc := todo.Description
				if todo.Done {
					desc = tuiDoneStyle.Render(desc)
				} else {
					desc = descStyle.Render(desc)
				}
				line += fmt.Sprintf(" - %s", desc)
			}

			if m.cursor == i {
				line = tuiSelectedStyle.Render(line)
			}

			s.WriteString(line)
			s.WriteString("\n")
		}
	}

	return tuiContainerStyle.Render(s.String())
}

func (m tuiModel) viewCalendar() string {
	var s strings.Builder

	s.WriteString(m.renderTabHeader())

	// Load and render calendar
	err := m.calendar.LoadTodos()
	if err != nil {
		s.WriteString("Error loading calendar data")
	} else {
		// Render calendar
		calendarStr := m.calendar.Render()
		s.WriteString(calendarStr)
	}

	s.WriteString("\n")

	return tuiContainerStyle.Render(s.String())
}

func (m tuiModel) viewCapture() string {
	var s strings.Builder

	s.WriteString(tuiTitleStyle.Render("🎯 Capture - Brain Dump Mode"))
	s.WriteString("\n\n")

	s.WriteString(tuiLabelStyle.Render("Type todos quickly, one per line:"))
	s.WriteString("\n\n")

	// Current input line
	inputLine := tuiInputStyle.Render(m.input)
	inputLine += tuiInputStyle.Render("█") // Cursor
	s.WriteString(fmt.Sprintf("> %s\n", inputLine))

	s.WriteString("\n")
	s.WriteString(tuiHelpStyle.Render("Enter: save todo and start next"))
	s.WriteString("\n")
	s.WriteString(tuiHelpStyle.Render("Use ' -- ' to separate title and description"))
	s.WriteString("\n")
	s.WriteString(tuiHelpStyle.Render("Example: 'Buy milk -- get the organic kind'"))
	s.WriteString("\n")
	s.WriteString(tuiHelpStyle.Render("Esc: return to inbox"))

	return tuiContainerStyle.Render(s.String())
}

func RunTUI(db *DB) error {
	p := tea.NewProgram(NewTuiModel(db), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
