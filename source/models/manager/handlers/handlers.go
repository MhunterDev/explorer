package handlers

import (
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type HandlerMsg struct {
	Input string
	Msg   string
	Err   error
}

func (h HandlerMsg) Error() error {
	return h.Err
}

func HandleCMD(s string) tea.Cmd {
	return func() tea.Msg {
		// Validate input
		s = strings.TrimSpace(s)
		if s == "" {
			return HandlerMsg{
				Input: s,
				Msg:   "",
				Err:   nil,
			}
		}

		output, err := exec.Command("sh", "-c", s).Output()
		if err != nil {
			return HandlerMsg{
				Input: s,
				Msg:   "",
				Err:   err,
			}
		}
		return HandlerMsg{
			Input: s,
			Msg:   string(output),
			Err:   nil,
		}
	}
}
