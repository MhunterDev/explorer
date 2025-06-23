package tree

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/MhunterDev/log4"
	"github.com/charmbracelet/lipgloss"
)

var (
	l         = log4.NewChannelLoggerWithConfig(log4.DefaultConfig())
	DirColor  = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	Udir      = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Underline(true)
	FileColor = lipgloss.NewStyle().Foreground(lipgloss.Color("110"))
	Ufile     = lipgloss.NewStyle().Foreground(lipgloss.Color("110")).Bold(true).Underline(true)
	ExeColor  = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Italic(true)
	Uexec     = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Italic(true).Underline(true)
	Border    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Foreground(lipgloss.Color("12")).Padding(1, 2, 0, 2)
)

type status struct {
	code int
	msg  string
}

func (s *status) Error() error {
	return fmt.Errorf("%s", s.msg)
}

func dirInfo(path string) ([]os.DirEntry, status) {

	dir, err := os.ReadDir(path)
	if err != nil {
		l.Error("base", err.Error())
		return nil, status{
			code: 1,
			msg:  fmt.Sprintf("Error reading root directory: %v", err),
		}
	}

	return dir, status{
		code: 0,
		msg:  "Root directory read successfully",
	}
}

func pathInfo(d []os.DirEntry) ([]string, status) {
	var paths []string
	for _, entry := range d {
		if entry.IsDir() {
			paths = append(paths, entry.Name())
		}
	}

	if len(paths) == 0 {
		return nil, status{
			code: 2,
			msg:  "No directories found in root",
		}
	}

	return paths, status{
		code: 0,
		msg:  "Directories retrieved successfully",
	}
}

func fileInfo(d []os.DirEntry) ([]string, []string, status) {

	var files []string
	var executables []string

	for _, entry := range d {
		if !entry.IsDir() {
			if isExecutable(entry) {
				executables = append(executables, entry.Name())
			} else {
				files = append(files, entry.Name())
			}
		}
	}
	return files, executables, status{
		code: 0,
		msg:  "Files retrieved successfully",
	}
}

func isExecutable(file os.DirEntry) bool {
	info, err := file.Info()
	if err != nil {
		l.Error("base", err.Error())
		return false
	}

	// Check if the file has execute permission for the user, group, or others
	mode := info.Mode()
	return mode&0111 != 0 // Check if any execute bit is set
}

type FileView struct {
	Name  string
	Dirs  []string
	Files []string
	Execs []string
}

type cachedFileView struct {
	view      *FileView
	timestamp time.Time
}

var viewCache = make(map[string]cachedFileView)

const cacheTimeout = 5 * time.Second

func NewFileView(path string) *FileView {
	if cached, exists := viewCache[path]; exists {
		if time.Since(cached.timestamp) < cacheTimeout {
			return cached.view
		}
	}

	// Validate path
	if path == "" {
		l.Error("base", "Empty path provided")
		return nil
	}

	// Clean the path
	path = filepath.Clean(path)

	d, s := dirInfo(path)
	if s.code != 0 {
		l.Error("base", s.msg)
		return nil
	}

	dirs, s := pathInfo(d)
	if s.code == 2 {
		// No directories is not an error, just empty slice
		dirs = []string{}
	} else if s.code != 0 {
		l.Error("Failed to get directories: %s", s.msg)
		return nil
	}

	files, execs, s := fileInfo(d)
	if s.code != 0 {
		l.Error("Failed to get files: %s", s.msg)
		return nil
	}

	view := &FileView{
		Name:  path, // Store full path instead of base
		Dirs:  dirs,
		Files: files,
		Execs: execs,
	}
	viewCache[path] = cachedFileView{view: view, timestamp: time.Now()}
	return view
}

func NewPath(path string) *FileView {
	d, err := os.ReadDir(path)
	if err != nil {
		l.Error("base", err.Error())
		return nil
	}

	dirs, s := pathInfo(d)
	if s.code != 0 {
		l.Error("Failed to get directories: %s", s.msg)
		return nil
	}

	files, execs, s := fileInfo(d)
	if s.code != 0 {
		l.Error("Failed to get files: %s", s.msg)
		return nil
	}

	return &FileView{
		Dirs:  dirs,
		Files: files,
		Execs: execs,
	}
}

func (fv *FileView) TypeBreak() (int, int, int) {
	d := len(fv.Dirs)
	f := d + len(fv.Files)
	e := f + len(fv.Execs)
	return d, f, e
}

func (fv *FileView) Expand(dir string) (nfv *FileView, s status) {
	var sub FileView

	// Use absolute path instead of relative to current working directory
	subDir := filepath.Join(fv.Name, dir)
	if !filepath.IsAbs(subDir) {
		current, err := os.Getwd()
		if err != nil {
			l.Error("base", err.Error())
			return nil, status{
				code: 1,
				msg:  fmt.Sprintf("Error getting current directory: %v", err),
			}
		}
		subDir = filepath.Join(current, subDir)
	}

	d, err := os.ReadDir(subDir)
	if err != nil {
		l.Error("base", err.Error())
		return nil, status{
			code: 1,
			msg:  fmt.Sprintf("Error reading directory %s: %v", dir, err),
		}
	}
	sub.Dirs, s = pathInfo(d)
	if s.code != 0 {
		l.Error("base", s.msg)
		return nil, s
	}
	sub.Files, sub.Execs, s = fileInfo(d)
	if s.code != 0 {
		l.Error("base", s.msg)
		return nil, s
	}
	sub.Name = subDir
	return &sub, status{
		code: 0,
		msg:  fmt.Sprintf("Directory %s expanded successfully", dir),
	}
}
