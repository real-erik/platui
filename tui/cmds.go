package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/real-erik/platui/process"
)

type organizationDataMsg struct {
	Payload []process.Result
}

type repositoryDataMsg struct {
	Payload []process.Result
}

type workflowDataMsg struct {
	Payload []process.Result
}

type artifactDataMsg struct {
	Payload []process.Result
}

type filepickerDataMsg struct {
	Payload int64
}

type errorMsg struct{ err error }

func (m model) getOrganizationsCmd() tea.Cmd {
	return func() tea.Msg {
		organizations, err := m.process.GetOrganizations()

		if err != nil {
			return errorMsg{err}
		}

		return organizationDataMsg{Payload: organizations}
	}
}

func (m model) getRepositoriesCmd(organization string) tea.Cmd {
	return func() tea.Msg {
		repositories, err := m.process.GetRepositories(organization)

		if err != nil {
			return errorMsg{err}
		}

		return repositoryDataMsg{Payload: repositories}
	}
}

func (m model) getWorkflowsCmd(repository string) tea.Cmd {
	return func() tea.Msg {
		workflows, err := m.process.GetWorkflowRuns(m.organization.Selected.Name, repository)

		if err != nil {
			return errorMsg{err}
		}

		return workflowDataMsg{Payload: workflows}
	}
}

func (m model) getArtifactsCmd(workflowId int64) tea.Cmd {
	return func() tea.Msg {
		artifacts, err := m.process.GetArtifacts(m.organization.Selected.Name, m.repository.Selected.Name, workflowId)

		if err != nil {
			return errorMsg{err}
		}

		return artifactDataMsg{Payload: artifacts}
	}
}

func (m model) downloadArtifactCmd(artifactId int64) tea.Cmd {
	return func() tea.Msg {
		err := m.process.DownloadArtifact(m.organization.Selected.Name, m.repository.Selected.Name, artifactId)

		if err != nil {
			return errorMsg{err}
		}

		return filepickerDataMsg{Payload: artifactId}
	}
}

func (m model) runFileCmd(filePath string) tea.Cmd {
	return func() tea.Msg {
		err := m.process.Run(filePath)

		if err != nil {
			return errorMsg{err}
		}

		// TODO: update screen to running state?
		return nil
	}
}
