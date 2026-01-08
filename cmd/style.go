package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 14
const defaultWidth = 20

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Italic(true).
	PaddingLeft(2).
	PaddingTop(2).
	PaddingBottom(1).
	Foreground(lipgloss.Color("#EE6FF8")).
	BorderForeground(lipgloss.Color("240"))

var filterStyle = lipgloss.NewStyle().
	Bold(true).
	PaddingLeft(2).
	Foreground(lipgloss.Color("#EE6FF8")).
	BorderForeground(lipgloss.Color("240"))

var welcomeStyle = lipgloss.NewStyle().
	Bold(true).
	PaddingTop(1).
	PaddingLeft(2).
	AlignHorizontal(lipgloss.Right)
