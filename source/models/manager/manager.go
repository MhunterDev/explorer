package manager

import (
	"github.com/MHunterDev/explorer/source/models/paths"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Manager struct {
	Selection int
	Viewer    *paths.Viewer
	Input     textinput.Model
	Portal    viewport.Model
}

func NewManager() tea.Model {
	return &Manager{
		Selection: 0,
		Viewer:    paths.NewViewer(),
		Input:     textinput.New(),
		Portal:    viewport.New(0, 0),
	}
}

func (m *Manager) Init() tea.Cmd {
	m.Input.Placeholder = "Enter command..."
	m.Input.Focus()
	m.Input.CharLimit = 256
	m.Input.Width = 80
	m.Portal.Height = 20
	m.Portal.Width = 80

	return nil
}

func (m *Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.Selection {
	case 0: // Input mode
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				m.Input.Reset()
				return m, nil
			case "left":
				m.Selection = 1 // Switch to Viewer
				return m, nil
			}
		}
		m.Input, cmd = m.Input.Update(msg)
		return m, cmd

	case 1: // Viewer mode
		// Handle mode switching keys first
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "right":
				m.Selection = 0 // Switch to Input
				return m, nil
			}
		}

		// Send message to Viewer and get updated model
		updatedViewer, cmd := m.Viewer.Update(msg)
		m.Viewer = updatedViewer.(*paths.Viewer)
		return m, cmd
	}

	return m, nil
}

func (m *Manager) View() string {
	viewPos := lipgloss.NewStyle().Align(lipgloss.Left).Padding(0, 1, 0, 0)
	v := viewPos.Render(m.Viewer.View())
	input := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Margin(1, 2).Render(m.Input.View())
	vp := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Margin(1, 2).Render(m.Portal.View())

	main := lipgloss.JoinVertical(lipgloss.Left, vp, input)
	output := lipgloss.JoinHorizontal(lipgloss.Top, v, main)

	return output
}
