package organization

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/real-erik/platui/process"
	"github.com/real-erik/platui/tui/list"
)

type Model struct {
	loading  bool
	list     list.Model
	Selected process.Result
}

func NewModel() Model {
	return Model{
		loading: true,
		list:    list.NewModel("Organizations"),
	}
}

type ForwardMsg struct {
	Payload process.Result
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg.(type) {
	case []process.Result:
		m.loading = false
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	if cmd != nil {
		listMsg := cmd()
		switch listMsg.(type) {
		case list.Msg:
			listMsg := listMsg.(list.Msg)
			switch listMsg.Direction {
			case list.Forward:
				m.Selected = listMsg.Item
				cmd = func() tea.Msg {
					return ForwardMsg{
						Payload: listMsg.Item,
					}
				}
			}
		default:
			// this is a command from bubbletea list, let it pass through
		}
	}

	return m, cmd

}

type item struct {
	title, desc string
}

func (m Model) View() string {
	if m.loading {
		return fmt.Sprintf("Loading Organizations...")
	}

	return lipgloss.NewStyle().Render(m.list.View())
}
