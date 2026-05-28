package main

import "github.com/charmbracelet/lipgloss"

var (
	colorBg      = lipgloss.Color("#0a0c12")
	colorPrimary = lipgloss.Color("#f5f7ff")
	colorTitle   = lipgloss.Color("#c8d0e8")
	colorMuted   = lipgloss.Color("#8b94ab")

	// vivid neon cyan + magenta palette
	colorCyan    = lipgloss.Color("#00f0ff")
	colorMint    = lipgloss.Color("#3df5c4")
	colorPurple  = lipgloss.Color("#c44dff")
	colorMagenta = lipgloss.Color("#ff3df1")
	colorPink    = lipgloss.Color("#ff4d9d")
	colorAmber   = lipgloss.Color("#ffbe5e")

	// one loud color per tab (about, experience, skills, education, contact)
	tabColors = []lipgloss.Color{colorCyan, colorMint, colorPurple, colorMagenta, colorPink}

	styleBase = lipgloss.NewStyle().
			Background(colorBg).
			Foreground(colorPrimary)

	styleMuted = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleAmber = lipgloss.NewStyle().
			Foreground(colorAmber)

	styleEntryTitle = lipgloss.NewStyle().
			Foreground(colorTitle).
			Bold(true)

	styleEntryMeta = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)

	styleHelp = lipgloss.NewStyle().
			Foreground(colorMuted)
)
