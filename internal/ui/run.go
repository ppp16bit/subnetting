package ui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func Run(colorMode ColorMode) error {
	renderer := newRenderer(os.Stdout, colorMode, noColorRequested())
	program := tea.NewProgram(newModel(renderer), tea.WithAltScreen())
	_, err := program.Run()
	return err
}
