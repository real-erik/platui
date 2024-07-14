package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/real-erik/platui/process"
	"github.com/real-erik/platui/tui/artifact"
	"github.com/real-erik/platui/tui/environment"
	"github.com/real-erik/platui/tui/filepicker"
	"github.com/real-erik/platui/tui/organization"
	"github.com/real-erik/platui/tui/repository"
	"github.com/real-erik/platui/tui/spinner"
	"github.com/real-erik/platui/tui/styles"
	"github.com/real-erik/platui/tui/workflow"
)

type modeStack []mode

func (s modeStack) GoForward(v mode) modeStack {
	return append(s, v)
}

func (s modeStack) GoBack() modeStack {
	l := len(s)
	return s[:l-1]
}

func (s modeStack) GetCurrent() mode {
	return s[len(s)-1]
}

func (m model) GoForward(mode mode) model {
	m.mode = m.mode.GoForward(mode)
	m.loadingMessage = ""
	return m
}

type model struct {
	mode           modeStack
	loadingMessage string
	process        process.Process
	spinner        spinner.Model
	environment    environment.Model
	organization   organization.Model
	repository     repository.Model
	workflow       workflow.Model
	artifact       artifact.Model
	filepicker     filepicker.Model
}

func NewModel(process process.Process) model {
	return model{
		process:      process,
		mode:         modeStack{Environment},
		spinner:      spinner.NewModel(),
		environment:  environment.NewModel(),
		organization: organization.NewModel(),
		repository:   repository.NewModel(),
		workflow:     workflow.NewModel(),
		artifact:     artifact.NewModel(),
		filepicker:   filepicker.NewModel(),
	}
}

type mode int

const (
	Environment mode = iota
	Organization
	Repository
	Workflow
	Artifact
	Filepicker
)

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case organizationDataMsg:
		m = m.GoForward(Organization)
		m.organization, _ = m.organization.Update(msg.Payload)
		return m, nil

	case repositoryDataMsg:
		m = m.GoForward(Repository)
		m.repository, _ = m.repository.Update(msg.Payload)
		return m, nil

	case workflowDataMsg:
		m = m.GoForward(Workflow)
		m.workflow, _ = m.workflow.Update(msg.Payload)
		return m, nil

	case artifactDataMsg:
		m = m.GoForward(Artifact)
		m.artifact, _ = m.artifact.Update(msg.Payload)
		return m, nil

	case filepickerDataMsg:
		m = m.GoForward(Filepicker)
		m.filepicker, cmd = m.filepicker.Update(filepicker.ArtifactMsg(msg.Payload))
		return m, cmd

	case environment.ForwardMsg:
		switch msg.Payload.Name {
		case "Github":
			m.loadingMessage = "Loading organizations"
			startLoading := m.spinner.Init()
			cmd := m.getOrganizationsCmd()
			return m, tea.Batch(startLoading, cmd)

		case "Local":
			m = m.GoForward(Filepicker)
			m.filepicker, cmd = m.filepicker.Update(filepicker.LocalMsg{})
			return m, cmd
		}

	case organization.ForwardMsg:
		m.loadingMessage = "Loading repositories"
		startLoading := m.spinner.Init()
		cmd = m.getRepositoriesCmd(msg.Payload.Name)
		return m, tea.Batch(startLoading, cmd)

	case organization.BackMsg:
		m.mode = m.mode.GoBack()
		return m, nil

	case repository.ForwardMsg:
		m.loadingMessage = "Loading workflows"
		cmd = m.getWorkflowsCmd(msg.Payload.Name)
		startLoading := m.spinner.Init()
		return m, tea.Batch(startLoading, cmd)

	case repository.BackMsg:
		m.mode = m.mode.GoBack()
		return m, nil

	case workflow.ForwardMsg:
		m.loadingMessage = "Loading artifacts"
		cmd = m.getArtifactsCmd(msg.Payload.ID)
		startLoading := m.spinner.Init()
		return m, tea.Batch(startLoading, cmd)

	case workflow.BackMsg:
		m.mode = m.mode.GoBack()
		return m, nil

	case artifact.ForwardMsg:
		m.loadingMessage = "Downloading files"
		cmd = m.downloadArtifactCmd(msg.Payload.ID)
		startLoading := m.spinner.Init()
		return m, tea.Batch(startLoading, cmd)

	case artifact.BackMsg:
		m.mode = m.mode.GoBack()
		return m, nil

	case filepicker.SelectedMsg:
		// TODO: change mode?
		cmd = m.runFileCmd(msg.Payload)
		return m, cmd

	case filepicker.BackMsg:
		m.mode = m.mode.GoBack()
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.environment, _ = m.environment.Update(msg)
		m.organization, _ = m.organization.Update(msg)
		m.repository, _ = m.repository.Update(msg)
		m.workflow, _ = m.workflow.Update(msg)
		m.artifact, _ = m.artifact.Update(msg)
		m.filepicker, _ = m.filepicker.Update(msg)

		return m, nil
	}

	// TODO: why can't I place this as default?
	if m.loadingMessage != "" {
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	switch m.mode.GetCurrent() {
	case Environment:
		m.environment, cmd = m.environment.Update(msg)
	case Organization:
		m.organization, cmd = m.organization.Update(msg)
	case Repository:
		m.repository, cmd = m.repository.Update(msg)
	case Workflow:
		m.workflow, cmd = m.workflow.Update(msg)
	case Artifact:
		m.artifact, cmd = m.artifact.Update(msg)
	case Filepicker:
		m.filepicker, cmd = m.filepicker.Update(msg)
	}

	return m, cmd

}

func (m model) View() string {
	if m.loadingMessage != "" {
		return styles.DocStyle.Render(m.spinner.View() + " " + m.loadingMessage + "...")
	}

	switch m.mode.GetCurrent() {
	case Environment:
		return m.environment.View()
	case Organization:
		return m.organization.View()
	case Repository:
		return m.repository.View()
	case Workflow:
		return m.workflow.View()
	case Artifact:
		return m.artifact.View()
	case Filepicker:
		return m.filepicker.View()
	}

	return ""
}

func main() {
	p := process.NewProcess(os.Getenv("GITHUB_TOKEN"))

	t := tea.NewProgram(NewModel(p), tea.WithAltScreen())
	if _, err := t.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
