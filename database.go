package main

import (
	"database/sql"
	"embed"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed sql/*.sql
var sqlFiles embed.FS

// loadSQL loads a SQL file from the embedded filesystem
func loadSQL(filename string) (string, error) {
	content, err := sqlFiles.ReadFile("sql/" + filename)
	if err != nil {
		return "", fmt.Errorf("failed to load SQL file %s: %w", filename, err)
	}
	return strings.TrimSpace(string(content)), nil
}

type Todo struct {
	ID             int
	Title          string
	Description    string
	Done           bool
	DueDate        *time.Time // Optional due date
	ScheduledStart *time.Time // Time block start
	ScheduledEnd   *time.Time // Time block end
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type DB struct {
	conn *sql.DB

	// Prepared statements
	insertTodo     *sql.Stmt
	getAllTodos    *sql.Stmt
	getInboxTodos  *sql.Stmt
	getDateTodos   *sql.Stmt
	getRangeTodos  *sql.Stmt
	getMonthTodos  *sql.Stmt
	updateTodo     *sql.Stmt
	deleteTodo     *sql.Stmt
	toggleTodo     *sql.Stmt
}

func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}

	// Enable WAL mode for better concurrency and performance
	if err := db.enableWAL(); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	if err := db.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	if err := db.prepareStatements(); err != nil {
		return nil, fmt.Errorf("failed to prepare statements: %w", err)
	}

	return db, nil
}

func (db *DB) enableWAL() error {
	// Enable WAL mode
	_, err := db.conn.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		return err
	}

	// Set other performance optimizations
	_, err = db.conn.Exec("PRAGMA synchronous=NORMAL")
	if err != nil {
		return err
	}

	_, err = db.conn.Exec("PRAGMA cache_size=1000")
	if err != nil {
		return err
	}

	_, err = db.conn.Exec("PRAGMA temp_store=MEMORY")
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) createTables() error {
	query, err := loadSQL("create_tables.sql")
	if err != nil {
		return err
	}

	_, err = db.conn.Exec(query)
	if err != nil {
		return err
	}

	// Migrate existing tables to add new columns if they don't exist
	return db.migrateTables()
}

func (db *DB) migrateTables() error {
	// Check which columns exist
	rows, err := db.conn.Query("PRAGMA table_info(todos)")
	if err != nil {
		return err
	}
	defer rows.Close()

	existingColumns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk bool
		var defaultValue interface{}
		
		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			return err
		}
		
		existingColumns[name] = true
	}

	// Add missing columns
	columnsToAdd := []string{
		"due_date DATETIME",
		"scheduled_start DATETIME", 
		"scheduled_end DATETIME",
	}

	for _, column := range columnsToAdd {
		parts := strings.Fields(column)
		columnName := parts[0]
		
		if !existingColumns[columnName] {
			_, err = db.conn.Exec(fmt.Sprintf("ALTER TABLE todos ADD COLUMN %s", column))
			if err != nil {
				return fmt.Errorf("failed to add column %s: %w", columnName, err)
			}
		}
	}

	return nil
}

func (db *DB) prepareStatements() error {
	var err error

	// Load SQL queries from files
	insertTodoSQL, err := loadSQL("insert_todo.sql")
	if err != nil {
		return err
	}
	db.insertTodo, err = db.conn.Prepare(insertTodoSQL)
	if err != nil {
		return err
	}

	getAllTodosSQL, err := loadSQL("get_all_todos.sql")
	if err != nil {
		return err
	}
	db.getAllTodos, err = db.conn.Prepare(getAllTodosSQL)
	if err != nil {
		return err
	}

	getInboxTodosSQL, err := loadSQL("get_inbox_todos.sql")
	if err != nil {
		return err
	}
	db.getInboxTodos, err = db.conn.Prepare(getInboxTodosSQL)
	if err != nil {
		return err
	}

	getDateTodosSQL, err := loadSQL("get_date_todos.sql")
	if err != nil {
		return err
	}
	db.getDateTodos, err = db.conn.Prepare(getDateTodosSQL)
	if err != nil {
		return err
	}

	getRangeTodosSQL, err := loadSQL("get_range_todos.sql")
	if err != nil {
		return err
	}
	db.getRangeTodos, err = db.conn.Prepare(getRangeTodosSQL)
	if err != nil {
		return err
	}

	getMonthTodosSQL, err := loadSQL("get_month_todos.sql")
	if err != nil {
		return err
	}
	db.getMonthTodos, err = db.conn.Prepare(getMonthTodosSQL)
	if err != nil {
		return err
	}

	updateTodoSQL, err := loadSQL("update_todo.sql")
	if err != nil {
		return err
	}
	db.updateTodo, err = db.conn.Prepare(updateTodoSQL)
	if err != nil {
		return err
	}

	deleteTodoSQL, err := loadSQL("delete_todo.sql")
	if err != nil {
		return err
	}
	db.deleteTodo, err = db.conn.Prepare(deleteTodoSQL)
	if err != nil {
		return err
	}

	toggleTodoSQL, err := loadSQL("toggle_todo.sql")
	if err != nil {
		return err
	}
	db.toggleTodo, err = db.conn.Prepare(toggleTodoSQL)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) AddTodo(title, description string, dueDate, scheduledStart, scheduledEnd *time.Time) error {
	_, err := db.insertTodo.Exec(title, description, dueDate, scheduledStart, scheduledEnd)
	return err
}

func (db *DB) GetAllTodos() ([]Todo, error) {
	rows, err := db.getAllTodos.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Done,
			&todo.DueDate,
			&todo.ScheduledStart,
			&todo.ScheduledEnd,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, rows.Err()
}

func (db *DB) GetInboxTodos() ([]Todo, error) {
	rows, err := db.getInboxTodos.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Done,
			&todo.DueDate,
			&todo.ScheduledStart,
			&todo.ScheduledEnd,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, rows.Err()
}

func (db *DB) GetDateTodos(date time.Time) ([]Todo, error) {
	// Format date as YYYY-MM-DD for SQLite DATE() function
	dateStr := date.Format("2006-01-02")
	
	rows, err := db.getDateTodos.Query(dateStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Done,
			&todo.DueDate,
			&todo.ScheduledStart,
			&todo.ScheduledEnd,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, rows.Err()
}

// GetTodayTodos is a convenience method for getting today's todos
func (db *DB) GetTodayTodos() ([]Todo, error) {
	return db.GetDateTodos(time.Now())
}

func (db *DB) GetRangeTodos(startDate, endDate time.Time) ([]Todo, error) {
	// Format dates as YYYY-MM-DD for SQLite DATE() function
	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")
	
	rows, err := db.getRangeTodos.Query(startStr, endStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Done,
			&todo.DueDate,
			&todo.ScheduledStart,
			&todo.ScheduledEnd,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, rows.Err()
}

func (db *DB) GetMonthTodos(date time.Time) ([]Todo, error) {
	// Format date as YYYY-MM-DD for SQLite strftime function
	dateStr := date.Format("2006-01-02")
	
	rows, err := db.getMonthTodos.Query(dateStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Done,
			&todo.DueDate,
			&todo.ScheduledStart,
			&todo.ScheduledEnd,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, rows.Err()
}

func (db *DB) UpdateTodo(id int, title, description string, dueDate, scheduledStart, scheduledEnd *time.Time) error {
	_, err := db.updateTodo.Exec(title, description, dueDate, scheduledStart, scheduledEnd, id)
	return err
}

func (db *DB) DeleteTodo(id int) error {
	_, err := db.deleteTodo.Exec(id)
	return err
}

func (db *DB) ToggleTodo(id int) error {
	_, err := db.toggleTodo.Exec(id)
	return err
}

// ScheduleTodo updates only the scheduled time for a todo
func (db *DB) ScheduleTodo(id int, scheduledStart, scheduledEnd *time.Time) error {
	scheduleSQL, err := loadSQL("schedule_todo.sql")
	if err != nil {
		return err
	}
	
	_, err = db.conn.Exec(scheduleSQL, scheduledStart, scheduledEnd, id)
	return err
}

func (db *DB) Close() error {
	if db.insertTodo != nil {
		db.insertTodo.Close()
	}
	if db.getAllTodos != nil {
		db.getAllTodos.Close()
	}
	if db.getInboxTodos != nil {
		db.getInboxTodos.Close()
	}
	if db.getDateTodos != nil {
		db.getDateTodos.Close()
	}
	if db.getRangeTodos != nil {
		db.getRangeTodos.Close()
	}
	if db.getMonthTodos != nil {
		db.getMonthTodos.Close()
	}
	if db.updateTodo != nil {
		db.updateTodo.Close()
	}
	if db.deleteTodo != nil {
		db.deleteTodo.Close()
	}
	if db.toggleTodo != nil {
		db.toggleTodo.Close()
	}

	return db.conn.Close()
}
