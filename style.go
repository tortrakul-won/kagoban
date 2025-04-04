package main

import "github.com/charmbracelet/lipgloss"

const (
	CardBorderColor    = "#FFBF00"
	CardBackgroudColor = "#ffd75f"
	ForegroundColor    = "#000000"
)

var systemStyle = lipgloss.NewStyle().
	// BorderStyle(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("#33ffaa")).
	// Background(lipgloss.Color("#71797E")).
	// Foreground(lipgloss.Color("#ffffff"))
	Padding(1, 2)

var cardStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.HiddenBorder()).BorderForeground(lipgloss.Color(CardBackgroudColor)).
	Background(lipgloss.Color(CardBackgroudColor)).
	Foreground(lipgloss.Color(ForegroundColor)).
	Padding(1, 2, 1, 2).
	Height(5).Width(20)

var sectionHeaderStyle = lipgloss.NewStyle().Bold(true).
	Background(lipgloss.Color(CardBackgroudColor)).Padding(0, 1).Foreground(lipgloss.Color(ForegroundColor))
