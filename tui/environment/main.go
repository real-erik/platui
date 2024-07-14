package environment

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/real-erik/platui/process"
	"github.com/real-erik/platui/tui/list"
)

type Model struct {
	list  list.Model
	items []process.Result
}

func NewModel() Model {
	items := []process.Result{
		{Name: "Github"},
		{Name: "Local"},
	}

	listItems := []list.Item{}
	for _, resultItem := range items {
		newItem := list.Item{
			Title: resultItem.Name,
		}
		listItems = append(listItems, newItem)
	}

	list := list.NewModel("Environment")
	list, _ = list.Update(listItems)

	return Model{
		list:  list,
		items: items,
	}
}

type ForwardMsg struct {
	Payload process.Result
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list, _ = m.list.Update(msg)
		return m, nil
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
				cmd = func() tea.Msg {
					return ForwardMsg{
						Payload: m.items[listMsg.Item],
					}
				}
			}
		default:
			// this is a command from bubbletea list, let it pass through
		}
	}

	return m, cmd

}

func (m Model) View() string {
	return lipgloss.NewStyle().Render(m.list.View())
}
