package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color constants
const (
	ColorOrange = "#FFA500"
	ColorGreen  = "#04B575"
	ColorBlue   = "#4A90E2"
	ColorWhite  = "#FFFFFF"
	ColorRed    = "#FF5F87"
	ColorYellow = "#F4D03F"
	ColorGray   = "#626262"
)

// CLI Styles
var (
	// Title styles
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorYellow)).
			Bold(true).
			Padding(0, 1)

	// Command styles
	commandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWhite)).
			Bold(true)

	// Description styles
	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWhite))

	// Variable styles (for <id>, <title>, etc.)
	variableStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorOrange)).
			Italic(true)

	// Success message styles
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorGreen)).
			Bold(true)

	// Error message styles
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorRed)).
			Bold(true)

	// Todo item styles
	todoStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	completedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorGreen)).
			Strikethrough(true)

	idStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorBlue)).
		Bold(true)
)

// TUI Styles
var (
	tuiTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorYellow)).
			Bold(true).
			Padding(0, 1).
			MarginBottom(1)

	tuiSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorWhite)).
				Background(lipgloss.Color(ColorBlue)).
				Bold(true).
				Padding(0, 1)

	tuiDoneStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorGreen)).
			Strikethrough(true)

	tuiHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorGray)).
			Padding(1, 0)

	tuiInputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWhite)).
			Bold(true)
	
	tuiLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorBlue)).
			Bold(true)

	tuiContainerStyle = lipgloss.NewStyle().
				Padding(1, 2)
)

// Helper function to style command text with orange variables
func styleCommand(text string) string {
	// Replace variables in angle brackets with orange styling
	result := text

	// Define variable patterns
	variables := []string{"<title>", "<description>", "<id>"}

	for _, variable := range variables {
		if strings.Contains(result, variable) {
			styledVar := variableStyle.Render(variable)
			result = strings.ReplaceAll(result, variable, styledVar)
		}
	}

	return result
}

// Helper function to create status styling
func createStatusStyle(done bool) (string, lipgloss.Color) {
	if done {
		return "☑", lipgloss.Color(ColorGreen)
	}
	return "☐", lipgloss.Color(ColorOrange)
}
