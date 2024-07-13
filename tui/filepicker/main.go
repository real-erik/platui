package filepicker

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	filepicker         filepicker.Model
	selectedFile       string
	quitting           bool
	err                error
	clicksAwayFromRoot int
}

func NewModel() Model {

	fp := filepicker.New()
	fp.AllowedTypes = []string{".zip", ".webm", ".png"}

	return Model{
		filepicker:         fp,
		clicksAwayFromRoot: 0,
	}
}

type BackMsg struct{}

type SelectedMsg struct {
	Payload string
}

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func getCurrentDirectory(artifactId int64) string {
	wd, _ := os.Getwd()
	wd = wd + "/output/" + strconv.FormatInt(artifactId, 10)
	return wd
}

func (m Model) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	case int64:
		wd := getCurrentDirectory(msg)
		m.filepicker.CurrentDirectory = wd
		cmd := m.Init()
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// HACK: add to clicks and remove later if it was a file select
			m.clicksAwayFromRoot++
		case "esc":
			if m.clicksAwayFromRoot == 0 {
				m.selectedFile = ""
				cmd := func() tea.Msg {
					return BackMsg{}
				}
				return m, cmd
			}

			m.clicksAwayFromRoot--
		}

	case clearErrorMsg:
		m.err = nil

	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	if didSelect, path := m.filepicker.DidSelectDisabledFile(msg); didSelect {
		m.clicksAwayFromRoot--
		m.err = errors.New(path + " is not valid.")
		m.selectedFile = ""
		return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {

		m.clicksAwayFromRoot--
		m.selectedFile = path
		cmd = func() tea.Msg {
			return SelectedMsg{
				Payload: path,
			}
		}
		return m, cmd
	}

	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder
	s.WriteString("\n  ")
	if m.err != nil {
		s.WriteString(m.filepicker.Styles.DisabledFile.Render(m.err.Error()))
	} else if m.selectedFile == "" {
		s.WriteString("Pick a file:")
	} else {
		s.WriteString("Selected file: " + m.filepicker.Styles.Selected.Render(m.selectedFile))
	}
	s.WriteString("\n\n" + m.filepicker.View() + "\n")
	return s.String()
}
