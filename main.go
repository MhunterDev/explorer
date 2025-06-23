package main

import (
	"github.com/MHunterDev/explorer/source/models/manager"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	a := tea.NewProgram(manager.NewManager(), tea.WithAltScreen())
	a.Run()
}
