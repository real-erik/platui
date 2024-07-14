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

type model struct {
	mode           mode
	previousMode   mode
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
	Loading mode = iota
	Environment
	Organization
	Repository
	Workflow
	Artifact
	Filepicker
)

func (m model) SetMode(mode mode) model {
	m.previousMode = m.mode
	m.mode = mode

	return m
}

func (m model) Init() tea.Cmd {
	return m.environment.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case environment.EnvironmentDataMsg:
		m.mode = Environment
		m.environment, _ = m.environment.Update(msg)
		return m, nil

	case organizationDataMsg:
		m.mode = Organization
		m.organization, _ = m.organization.Update(msg.Payload)
		return m, nil

	case repositoryDataMsg:
		m.mode = Repository
		m.repository, _ = m.repository.Update(msg.Payload)
		return m, nil

	case workflowDataMsg:
		m.mode = Workflow
		m.workflow, _ = m.workflow.Update(msg.Payload)
		return m, nil

	case artifactDataMsg:
		m.mode = Artifact
		m.artifact, _ = m.artifact.Update(msg.Payload)
		return m, nil

	case filepickerDataMsg:
		m.mode = Filepicker
		m.filepicker, cmd = m.filepicker.Update(filepicker.ArtifactMsg(msg.Payload))
		return m, cmd

	case environment.ForwardMsg:
		switch msg.Payload.Name {
		case "Github":
			m.mode = Loading
			m.loadingMessage = "Loading organizations"
			startLoading := m.spinner.Init()
			cmd := m.getOrganizationsCmd()
			return m, tea.Batch(startLoading, cmd)

		case "Local":
			m = m.SetMode(Filepicker)
			m.filepicker, cmd = m.filepicker.Update(filepicker.LocalMsg{})
			return m, cmd
		}

	case organization.ForwardMsg:
		m.mode = Loading
		m.loadingMessage = "Loading repositories"
		startLoading := m.spinner.Init()
		cmd = m.getRepositoriesCmd(msg.Payload.Name)
		return m, tea.Batch(startLoading, cmd)

	case organization.BackMsg:
		m.mode = Environment
		return m, nil

	case repository.ForwardMsg:
		m.mode = Loading
		m.loadingMessage = "Loading workflows"
		cmd = m.getWorkflowsCmd(msg.Payload.Name)
		startLoading := m.spinner.Init()
		return m, tea.Batch(startLoading, cmd)

	case repository.BackMsg:
		m.mode = Organization
		return m, nil

	case workflow.ForwardMsg:
		m.mode = Loading
		m.loadingMessage = "Loading artifacts"
		cmd = m.getArtifactsCmd(msg.Payload.ID)
		startLoading := m.spinner.Init()
		return m, tea.Batch(startLoading, cmd)

	case workflow.BackMsg:
		m.mode = Repository
		return m, nil

	case artifact.ForwardMsg:
		m.mode = Loading
		m.loadingMessage = "Downloading files"
		cmd = m.downloadArtifactCmd(msg.Payload.ID)
		startLoading := m.spinner.Init()
		return m, tea.Batch(startLoading, cmd)

	case artifact.BackMsg:
		m.mode = Workflow
		return m, nil

	case filepicker.SelectedMsg:
		// TODO: change mode?
		cmd = m.runFileCmd(msg.Payload)
		return m, cmd

	case filepicker.BackMsg:
		m.mode = m.previousMode
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
	switch m.mode {
	case Loading:
		m.spinner, cmd = m.spinner.Update(msg)
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
	switch m.mode {
	case Loading:
		return styles.DocStyle.Render(m.spinner.View() + " " + m.loadingMessage + "...")
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
