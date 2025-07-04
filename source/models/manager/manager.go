package manager

import (
	"os"
	"path/filepath"

	"github.com/MHunterDev/explorer/source/models/manager/handlers"
	"github.com/MHunterDev/explorer/source/models/paths"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	viewPosStyle = lipgloss.NewStyle().Align(lipgloss.Left).Padding(0, 1, 0, 0)
	inputStyle   = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1).MarginLeft(2)
	vpStyle      = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1).MarginLeft(2)
	inUse        = lipgloss.NewStyle().Align(lipgloss.Left).Padding(0, 1, 0, 0).Foreground(lipgloss.Color("13"))
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
	m.Portal.Width = 83

	return nil
}

func clampViewportOffset(vp *viewport.Model) {
	maxOffset := vp.TotalLineCount() - vp.Height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if vp.YOffset > maxOffset {
		vp.SetYOffset(maxOffset)
	}
	if vp.YOffset < 0 {
		vp.SetYOffset(0)
	}
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
			case "left":
				m.Selection = 1 // Switch to Viewer
				m.Input.Blur()
				return m, nil
			case "pgup":
				m.Portal.SetYOffset(m.Portal.YOffset - m.Portal.Height)
				clampViewportOffset(&m.Portal)
				return m, nil
			case "pgdown":
				m.Portal.SetYOffset(m.Portal.YOffset + m.Portal.Height)
				clampViewportOffset(&m.Portal)
				return m, nil
			case "home":
				m.Portal.GotoTop()
				return m, nil
			case "end":
				m.Portal.GotoBottom()
				return m, nil
			case "enter":
				if m.Input.Value() != "" {
					m.Portal.SetContent(m.Portal.View() + "\n> " + m.Input.Value())
					// Handle the command input and return the command
					cmd = handlers.HandleCMD(m.Input.Value())
					m.Input.Reset()
					clampViewportOffset(&m.Portal)
					return m, cmd
				}
			}
		case handlers.HandlerMsg:
			cur := m.Portal.View() + "\n"
			if msg.Error() != nil {
				m.Portal.SetContent(cur + "Error: " + msg.Error().Error())
			} else {
				m.Portal.SetContent(cur + msg.Msg)
			}
			m.Portal.GotoBottom()
			clampViewportOffset(&m.Portal)
			return m, nil
		}

		m.Input, cmd = m.Input.Update(msg)
		m.Portal, _ = m.Portal.Update(msg)
		clampViewportOffset(&m.Portal)
		return m, cmd

	case 1: // Viewer mode
		// Handle mode switching keys first
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "right":
				m.Selection = 0 // Switch to Input
				m.Input.Focus()
				return m, nil
			case "pgup":
				m.Portal.SetYOffset(m.Portal.YOffset - m.Portal.Height)
				clampViewportOffset(&m.Portal)
				return m, nil
			case "pgdown":
				m.Portal.SetYOffset(m.Portal.YOffset + m.Portal.Height)
				clampViewportOffset(&m.Portal)
				return m, nil
			case "home":
				m.Portal.GotoTop()
				return m, nil
			case "end":
				m.Portal.GotoBottom()
				return m, nil
			}
		}

		switch msg := msg.(type) {
		case paths.PortalMsg:
			if msg.Error() != nil {
				m.Portal.SetContent("Error: " + msg.Error().Error())
			} else {
				data, err := os.ReadFile(filepath.Join("", msg.String()))
				if err != nil {
					m.Portal.SetContent("Error reading file: " + err.Error())
				} else {
					m.Portal.SetContent(string(data))
				}
			}
			clampViewportOffset(&m.Portal)
			return m, nil
		}

		// Send message to Viewer and get updated model
		updatedViewer, cmd := m.Viewer.Update(msg)
		if viewer, ok := updatedViewer.(*paths.Viewer); ok {
			m.Viewer = viewer
		}
		return m, cmd
	}
	return m, nil
}

func (m *Manager) View() string {
	var v string
	if m.Selection == 1 {
		v = inUse.Render(m.Viewer.View())
	} else {
		v = viewPosStyle.Render(m.Viewer.View())
	}
	input := inputStyle.Render(m.Input.View())
	vp := vpStyle.Render(m.Portal.View() + "\n")

	main := lipgloss.JoinVertical(lipgloss.Left, vp, input)
	output := lipgloss.JoinHorizontal(lipgloss.Top, v, main)

	return output
}
