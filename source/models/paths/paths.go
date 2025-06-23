package paths

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MHunterDev/explorer/source/tree"

	tea "github.com/charmbracelet/bubbletea"
)

type FileViewNode struct {
	View     *tree.FileView
	Name     string
	Parent   *FileViewNode
	Children []*FileViewNode
}

type Viewer struct {
	Cursor  int
	Current *FileViewNode
	Root    *FileViewNode
}

func (v Viewer) Init() tea.Cmd {

	return nil
}
func NewViewer() *Viewer {
	rootView := tree.NewFileView("/")
	rootNode := &FileViewNode{
		View: rootView,
		Name: "/",
	}
	return &Viewer{
		Cursor:  0,
		Current: rootNode,
		Root:    rootNode,
	}
}

func (v *Viewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if v.Cursor > 0 {
				v.Cursor--
			}
		case "down":
			maxItems := len(v.Current.View.Dirs) + len(v.Current.View.Files) + len(v.Current.View.Execs)
			if maxItems > 0 && v.Cursor < maxItems-1 {
				v.Cursor++
			}
		case "enter":
			if v.Cursor < len(v.Current.View.Dirs) {
				dirName := v.Current.View.Dirs[v.Cursor]
				newView := tree.NewFileView(filepath.Join(v.Current.Name, dirName))
				if newView != nil {
					// Check if already expanded to avoid duplicate children
					var existing *FileViewNode
					for _, child := range v.Current.Children {
						if child.Name == filepath.Join(v.Current.Name, dirName) {
							existing = child
							break
						}
					}
					if existing != nil {
						v.Current = existing
					} else {
						childNode := &FileViewNode{
							View:   newView,
							Name:   filepath.Join(v.Current.Name, dirName),
							Parent: v.Current,
						}
						v.Current.Children = append(v.Current.Children, childNode)
						v.Current = childNode
					}
					v.Cursor = 0
				}
				return v, nil
			}
			if v.Cursor >= len(v.Current.View.Dirs) && v.Cursor < (len(v.Current.View.Files)+len(v.Current.View.Dirs)) {
				fileName := v.Current.View.Files[v.Cursor-len(v.Current.View.Dirs)]
				filePath := filepath.Join(v.Current.Name, fileName)

				// Check file size before reading (limit to 10MB)
				fileInfo, err := os.Stat(filePath)
				if err != nil {
					g := PortalMsg{Content: "", Err: err}
					return v, g.UpdatePortal()
				}

				const maxFileSize = 10 * 1024 * 1024 // 10MB
				if fileInfo.Size() > maxFileSize {
					g := PortalMsg{Content: "", Err: fmt.Errorf("file too large (%d bytes), maximum allowed is %d bytes", fileInfo.Size(), maxFileSize)}
					return v, g.UpdatePortal()
				}

				// Read file content directly instead of using less
				content, err := os.ReadFile(filePath)
				if err != nil {
					g := PortalMsg{Content: "", Err: err}
					return v, g.UpdatePortal()
				}

				// Create temp file with content
				file, err := os.CreateTemp("/tmp", "explorer-*.txt")
				if err != nil {
					g := PortalMsg{Content: "", Err: err}
					return v, g.UpdatePortal()
				}
				defer file.Close()

				if _, err := file.Write(content); err != nil {
					g := PortalMsg{Content: "", Err: err}
					return v, g.UpdatePortal()
				}

				g := PortalMsg{Content: file.Name(), Err: nil}
				return v, g.UpdatePortal()
			}

		case "left", "esc":
			if v.Current.Parent != nil {
				v.Current = v.Current.Parent
				v.Cursor = 0
			} else {
				return v, tea.Quit
			}
			return v, nil
		case "ctrl+c":
			return v, tea.Quit
		}
		return v, nil
	}
	return v, nil
}

func (v *Viewer) View() string {
	var output string
	var options []string
	dirCount := len(v.Current.View.Dirs)
	fileCount := len(v.Current.View.Files)

	// Build options list
	options = append(options, v.Current.View.Dirs...)
	options = append(options, v.Current.View.Files...)
	options = append(options, v.Current.View.Execs...)

	for i, opt := range options {
		var line string
		switch {
		case i < dirCount:
			if i == v.Cursor {
				line = "|--" + tree.Udir.Render(opt)
			} else {
				line = "|--" + tree.DirColor.Render(opt)
			}
		case i < dirCount+fileCount:
			if i == v.Cursor {
				line = "|-" + tree.Ufile.Render(opt)
			} else {
				line = "|-" + tree.FileColor.Render(opt)
			}
		default: // Executables
			if i == v.Cursor {
				line = "|-" + tree.Uexec.Render(opt)
			} else {
				line = "|-" + tree.ExeColor.Render(opt)
			}
		}
		output += line + "\n"
	}
	ins := "use arrow keys to navigate\nuse enter to open a directory\nuse left/ESC to go back\nuse Ctrl+C to exit.\n"

	return tree.Border.Render(ins + "\n\n" + output)
}

type PortalMsg struct {
	Content string
	Err     error
}

func (p PortalMsg) Error() error {
	return p.Err
}
func (p PortalMsg) String() string {
	return p.Content
}

func (p *PortalMsg) UpdatePortal() tea.Cmd {
	return func() tea.Msg {
		return PortalMsg{
			Content: p.Content,
			Err:     p.Err,
		}
	}
}
