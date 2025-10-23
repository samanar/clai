package components

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	padding  = 2
	maxWidth = 80
)

var (
	currentPkgNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	doneStyle           = lipgloss.NewStyle().Margin(1, 2)
	checkMark           = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	errorStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
)

type progressMsg struct {
	downloaded int64
	total      int64
}

type progressErrMsg struct {
	err error
}

type downloadCompleteMsg struct {
	path string
}

// DownloadModel represents the download progress TUI model
type DownloadModel struct {
	url          string
	destPath     string
	progress     progress.Model
	err          error
	done         bool
	progressChan chan progressMsg
}

// Error returns the error if download failed
func (m DownloadModel) Error() error {
	return m.err
}

// NewDownloadModel creates a new download model
func NewDownloadModel(url, destPath string) DownloadModel {
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = maxWidth - padding*2 - 4

	return DownloadModel{
		url:          url,
		destPath:     destPath,
		progress:     prog,
		progressChan: make(chan progressMsg, 100),
	}
}

// Init initializes the download model
func (m DownloadModel) Init() tea.Cmd {
	return tea.Batch(
		m.progress.Init(),
		m.downloadFile,
		m.waitForProgress,
	)
}

// waitForProgress waits for progress updates from the channel
func (m DownloadModel) waitForProgress() tea.Msg {
	return <-m.progressChan
}

// Update handles messages for the download model
func (m DownloadModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - padding*2 - 4
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil

	case progressMsg:
		var cmds []tea.Cmd
		if msg.total > 0 {
			percent := float64(msg.downloaded) / float64(msg.total)
			cmd := m.progress.SetPercent(percent)
			cmds = append(cmds, cmd)
		}
		// Continue waiting for more progress updates
		cmds = append(cmds, m.waitForProgress)
		return m, tea.Batch(cmds...)

	case downloadCompleteMsg:
		m.done = true
		close(m.progressChan)
		return m, tea.Quit

	case progressErrMsg:
		m.err = msg.err
		close(m.progressChan)
		return m, tea.Quit

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	default:
		return m, nil
	}

	return m, nil
}

// View renders the download progress view
func (m DownloadModel) View() string {
	if m.err != nil {
		return doneStyle.Render(fmt.Sprintf("%s Download failed: %v\n", errorStyle.Render("✗"), m.err))
	}

	if m.done {
		return doneStyle.Render(fmt.Sprintf("%s Downloaded successfully to: %s\n", checkMark, m.destPath))
	}

	pad := strings.Repeat(" ", padding)
	return "\n" +
		pad + currentPkgNameStyle.Render(fmt.Sprintf("Downloading: %s", filepath.Base(m.url))) + "\n" +
		pad + m.progress.View() + "\n\n" +
		pad + "Press q or ctrl+c to cancel\n"
}

// downloadFile performs the actual download
func (m DownloadModel) downloadFile() tea.Msg {
	// Create the destination directory if it doesn't exist
	dir := filepath.Dir(m.destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return progressErrMsg{err: fmt.Errorf("failed to create directory: %w", err)}
	}

	// Create the HTTP request
	resp, err := http.Get(m.url)
	if err != nil {
		return progressErrMsg{err: fmt.Errorf("failed to download: %w", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return progressErrMsg{err: fmt.Errorf("bad status: %s", resp.Status)}
	}

	// Create the destination file
	out, err := os.Create(m.destPath)
	if err != nil {
		return progressErrMsg{err: fmt.Errorf("failed to create file: %w", err)}
	}
	defer out.Close()

	// Get the total size
	total := resp.ContentLength

	// Create a progress writer that sends updates to the channel
	pw := &progressWriter{
		total:        total,
		downloaded:   0,
		progressChan: m.progressChan,
	}

	// Copy with progress
	_, err = io.Copy(out, io.TeeReader(resp.Body, pw))
	if err != nil {
		return progressErrMsg{err: fmt.Errorf("failed to save file: %w", err)}
	}

	return downloadCompleteMsg{path: m.destPath}
}

// progressWriter wraps an io.Writer and tracks progress
type progressWriter struct {
	total        int64
	downloaded   int64
	progressChan chan progressMsg
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.downloaded += int64(n)

	// Send progress update through channel
	select {
	case pw.progressChan <- progressMsg{downloaded: pw.downloaded, total: pw.total}:
	default:
		// Channel full, skip this update
	}

	return n, nil
}

func (pw *progressWriter) Downloaded() int64 {
	return pw.downloaded
}

// Download starts a download with progress display
func Download(url, destPath string) error {
	m := NewDownloadModel(url, destPath)
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	if m.err != nil {
		return m.err
	}

	return nil
}
