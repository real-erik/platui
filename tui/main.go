package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/real-erik/platui/process"
	"github.com/real-erik/platui/tui/artifact"
	"github.com/real-erik/platui/tui/filepicker"
	"github.com/real-erik/platui/tui/organization"
	"github.com/real-erik/platui/tui/repository"
	"github.com/real-erik/platui/tui/workflow"
)

type model struct {
	mode         mode
	process      process.Process
	organization organization.Model
	repository   repository.Model
	workflow     workflow.Model
	artifact     artifact.Model
	filepicker   filepicker.Model
}

func NewModel(process process.Process) model {
	return model{
		process:      process,
		mode:         Organization,
		organization: organization.NewModel(),
		repository:   repository.NewModel(),
		workflow:     workflow.NewModel(),
		artifact:     artifact.NewModel(),
		filepicker:   filepicker.NewModel(),
	}
}

type mode int

const (
	Organization mode = iota
	Repository
	Workflow
	Artifact
	Filepicker
)

func (m model) Init() tea.Cmd {
	cmd := m.getOrganizationsCmd()
	return cmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case organizationDataMsg:
		m.organization, _ = m.organization.Update(msg.Payload)
		return m, nil

	case repositoryDataMsg:
		m.repository, _ = m.repository.Update(msg.Payload)
		return m, nil

	case workflowDataMsg:
		m.workflow, _ = m.workflow.Update(msg.Payload)
		return m, nil

	case artifactDataMsg:
		m.artifact, _ = m.artifact.Update(msg.Payload)
		return m, nil

	case filepickerDataMsg:
		m.filepicker, cmd = m.filepicker.Update(msg.Payload)
		return m, cmd

	case organization.ForwardMsg:
		m.mode = Repository
		cmd = m.getRepositoriesCmd(msg.Payload.Name)
		startLoading := m.repository.Init()
		return m, tea.Batch(startLoading, cmd)

	case repository.ForwardMsg:
		m.mode = Workflow
		cmd = m.getWorkflowsCmd(msg.Payload.Name)
		startLoading := m.workflow.Init()
		return m, tea.Batch(startLoading, cmd)

	case repository.BackMsg:
		m.mode = Organization
		return m, nil

	case workflow.ForwardMsg:
		m.mode = Artifact
		cmd = m.getArtifactsCmd(msg.Payload.ID)
		return m, cmd

	case workflow.BackMsg:
		m.mode = Repository
		return m, nil

	case artifact.ForwardMsg:
		m.mode = Filepicker
		cmd = m.downloadArtifactCmd(msg.Payload.ID)
		return m, cmd

	case artifact.BackMsg:
		m.mode = Workflow
		return m, nil

	case filepicker.SelectedMsg:
		// TODO: change mode?
		cmd = m.runFileCmd(msg.Payload)
		return m, cmd

	case filepicker.BackMsg:
		m.mode = Artifact
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.organization, _ = m.organization.Update(msg)
		m.repository, _ = m.repository.Update(msg)
		m.workflow, _ = m.workflow.Update(msg)
		m.artifact, _ = m.artifact.Update(msg)
		m.filepicker, _ = m.filepicker.Update(msg)
		return m, nil
	}

	switch m.mode {
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
	switch m.mode {
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
