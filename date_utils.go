package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseDueDate parses a due date string into a time.Time pointer
// Supports formats: "2024-12-25", "Dec 25", "tomorrow", "next week", etc.
func ParseDueDate(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}
	
	dateStr = strings.TrimSpace(strings.ToLower(dateStr))
	now := time.Now()
	
	// Handle relative dates
	switch dateStr {
	case "today":
		date := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
		return &date, nil
	case "tomorrow":
		date := now.AddDate(0, 0, 1)
		date = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, date.Location())
		return &date, nil
	case "next week":
		date := now.AddDate(0, 0, 7)
		date = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, date.Location())
		return &date, nil
	}
	
	// Try standard date formats
	formats := []string{
		"2006-01-02",           // YYYY-MM-DD
		"2006/01/02",           // YYYY/MM/DD
		"01/02/2006",           // MM/DD/YYYY
		"Jan 2, 2006",          // Jan 2, 2006
		"January 2, 2006",      // January 2, 2006
		"2 Jan 2006",           // 2 Jan 2006
		"2 January 2006",       // 2 January 2006
	}
	
	for _, format := range formats {
		if parsed, err := time.Parse(format, dateStr); err == nil {
			// Set to end of day
			date := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, now.Location())
			return &date, nil
		}
	}
	
	// Try month/day without year (assume current year)
	monthDayFormats := []string{
		"Jan 2",       // Jan 2
		"January 2",   // January 2
		"01/02",       // MM/DD
		"1/2",         // M/D
	}
	
	for _, format := range monthDayFormats {
		if parsed, err := time.Parse(format, dateStr); err == nil {
			// Use current year and set to end of day
			date := time.Date(now.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, now.Location())
			// If the date has already passed this year, use next year
			if date.Before(now) {
				date = date.AddDate(1, 0, 0)
			}
			return &date, nil
		}
	}
	
	return nil, fmt.Errorf("unable to parse date: %s", dateStr)
}

// FormatDueDate formats a due date for display
func FormatDueDate(dueDate *time.Time) string {
	if dueDate == nil {
		return ""
	}
	
	now := time.Now()
	diff := dueDate.Sub(now)
	
	// If it's today
	if dueDate.Year() == now.Year() && dueDate.YearDay() == now.YearDay() {
		return "Today"
	}
	
	// If it's tomorrow
	tomorrow := now.AddDate(0, 0, 1)
	if dueDate.Year() == tomorrow.Year() && dueDate.YearDay() == tomorrow.YearDay() {
		return "Tomorrow"
	}
	
	// If it's overdue
	if diff < 0 {
		days := int(-diff.Hours() / 24)
		if days == 1 {
			return "1 day overdue"
		}
		return fmt.Sprintf("%d days overdue", days)
	}
	
	// If it's within a week
	if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "In 1 day"
		}
		return fmt.Sprintf("In %d days", days)
	}
	
	// Otherwise show the date
	if dueDate.Year() == now.Year() {
		return dueDate.Format("Jan 2")
	}
	return dueDate.Format("Jan 2, 2006")
}

// GetDueDateColor returns appropriate color for due date status
func GetDueDateColor(dueDate *time.Time) string {
	if dueDate == nil {
		return ColorGray
	}
	
	now := time.Now()
	diff := dueDate.Sub(now)
	
	// Overdue - red
	if diff < 0 {
		return ColorRed
	}
	
	// Due today - orange
	if dueDate.Year() == now.Year() && dueDate.YearDay() == now.YearDay() {
		return ColorOrange
	}
	
	// Due tomorrow - yellow
	tomorrow := now.AddDate(0, 0, 1)
	if dueDate.Year() == tomorrow.Year() && dueDate.YearDay() == tomorrow.YearDay() {
		return ColorYellow
	}
	
	// Due within a week - blue
	if diff < 7*24*time.Hour {
		return ColorBlue
	}
	
	// Future date - gray
	return ColorGray
}

// TimeBlock represents a scheduled time block
type TimeBlock struct {
	Start *time.Time
	End   *time.Time
}

// ParseTimeBlock parses time block expressions like:
// "Monday 2pm-4pm", "Dec 25 9am-11am", "tomorrow 3pm for 2 hours"
func ParseTimeBlock(input string) (*TimeBlock, error) {
	if input == "" {
		return nil, nil
	}
	
	input = strings.TrimSpace(input)
	now := time.Now()
	
	// Pattern: "DAY TIME-TIME" or "DATE TIME-TIME"
	timeRangePattern := regexp.MustCompile(`^(.+?)\s+(\d{1,2}(?::\d{2})?(?:am|pm)?)-(\d{1,2}(?::\d{2})?(?:am|pm)?)$`)
	
	// Pattern: "DAY TIME for DURATION"
	durationPattern := regexp.MustCompile(`^(.+?)\s+(\d{1,2}(?::\d{2})?(?:am|pm)?)\s+for\s+(\d+)\s*(hour|hr|h|minute|min|m)s?$`)
	
	if matches := timeRangePattern.FindStringSubmatch(input); matches != nil {
		datePart := strings.TrimSpace(matches[1])
		startTime := strings.TrimSpace(matches[2])
		endTime := strings.TrimSpace(matches[3])
		
		baseDate, err := parseBaseDate(datePart, now)
		if err != nil {
			return nil, err
		}
		
		start, err := parseTimeOnDate(startTime, baseDate)
		if err != nil {
			return nil, err
		}
		
		end, err := parseTimeOnDate(endTime, baseDate)
		if err != nil {
			return nil, err
		}
		
		// Handle overnight times (end < start means next day)
		if end.Before(*start) {
			nextDay := end.AddDate(0, 0, 1)
			end = &nextDay
		}
		
		return &TimeBlock{Start: start, End: end}, nil
	}
	
	if matches := durationPattern.FindStringSubmatch(input); matches != nil {
		datePart := strings.TrimSpace(matches[1])
		startTime := strings.TrimSpace(matches[2])
		durationValue := matches[3]
		durationUnit := matches[4]
		
		baseDate, err := parseBaseDate(datePart, now)
		if err != nil {
			return nil, err
		}
		
		start, err := parseTimeOnDate(startTime, baseDate)
		if err != nil {
			return nil, err
		}
		
		duration, err := strconv.Atoi(durationValue)
		if err != nil {
			return nil, err
		}
		
		var durationTime time.Duration
		switch strings.ToLower(string(durationUnit[0])) {
		case "h":
			durationTime = time.Duration(duration) * time.Hour
		case "m":
			durationTime = time.Duration(duration) * time.Minute
		default:
			return nil, fmt.Errorf("unsupported duration unit: %s", durationUnit)
		}
		
		end := start.Add(durationTime)
		return &TimeBlock{Start: start, End: &end}, nil
	}
	
	return nil, fmt.Errorf("unable to parse time block: %s", input)
}

// parseBaseDate parses the date part (Monday, tomorrow, Dec 25, etc.)
func parseBaseDate(datePart string, now time.Time) (*time.Time, error) {
	datePart = strings.ToLower(datePart)
	
	// Handle relative dates
	switch datePart {
	case "today":
		date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return &date, nil
	case "tomorrow":
		date := now.AddDate(0, 0, 1)
		date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		return &date, nil
	}
	
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
	
	if weekday, exists := weekdays[datePart]; exists {
		daysUntil := (int(weekday) - int(now.Weekday()) + 7) % 7
		if daysUntil == 0 && now.Hour() >= 12 { // If it's the same day but afternoon, assume next week
			daysUntil = 7
		}
		date := now.AddDate(0, 0, daysUntil)
		date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		return &date, nil
	}
	
	// Try to parse as regular date
	return ParseDueDate(datePart)
}

// parseTimeOnDate parses time like "2pm", "14:30" on a specific date
func parseTimeOnDate(timeStr string, baseDate *time.Time) (*time.Time, error) {
	timeStr = strings.ToLower(timeStr)
	
	// Handle AM/PM format
	amPmPattern := regexp.MustCompile(`^(\d{1,2})(?::(\d{2}))?(?:\s*)?(am|pm)$`)
	if matches := amPmPattern.FindStringSubmatch(timeStr); matches != nil {
		hour, _ := strconv.Atoi(matches[1])
		minute := 0
		if matches[2] != "" {
			minute, _ = strconv.Atoi(matches[2])
		}
		ampm := matches[3]
		
		if ampm == "pm" && hour != 12 {
			hour += 12
		} else if ampm == "am" && hour == 12 {
			hour = 0
		}
		
		result := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), hour, minute, 0, 0, baseDate.Location())
		return &result, nil
	}
	
	// Handle 24-hour format
	hourMinPattern := regexp.MustCompile(`^(\d{1,2}):(\d{2})$`)
	if matches := hourMinPattern.FindStringSubmatch(timeStr); matches != nil {
		hour, _ := strconv.Atoi(matches[1])
		minute, _ := strconv.Atoi(matches[2])
		
		result := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), hour, minute, 0, 0, baseDate.Location())
		return &result, nil
	}
	
	// Handle hour only (assume PM if < 8, AM if >= 8)
	hourPattern := regexp.MustCompile(`^(\d{1,2})$`)
	if matches := hourPattern.FindStringSubmatch(timeStr); matches != nil {
		hour, _ := strconv.Atoi(matches[1])
		
		// Smart defaulting: 1-7 = PM, 8-12 = AM, 13-24 = 24hr format
		if hour >= 1 && hour <= 7 {
			hour += 12
		}
		
		result := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), hour, 0, 0, 0, baseDate.Location())
		return &result, nil
	}
	
	return nil, fmt.Errorf("unable to parse time: %s", timeStr)
}

// FormatTimeBlock formats a time block for display
func FormatTimeBlock(start, end *time.Time) string {
	if start == nil {
		return ""
	}
	
	if end == nil {
		return fmt.Sprintf("Scheduled: %s", start.Format("Jan 2 3:04pm"))
	}
	
	// Same day
	if start.Year() == end.Year() && start.YearDay() == end.YearDay() {
		return fmt.Sprintf("Scheduled: %s %s-%s", 
			start.Format("Jan 2"), 
			start.Format("3:04pm"), 
			end.Format("3:04pm"))
	}
	
	// Different days
	return fmt.Sprintf("Scheduled: %s - %s", 
		start.Format("Jan 2 3:04pm"), 
		end.Format("Jan 2 3:04pm"))
}
