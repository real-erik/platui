package repository

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/real-erik/platui/process"
	"github.com/real-erik/platui/tui/list"
)

type Model struct {
	list     list.Model
	items    []process.Result
	Selected process.Result
}

func NewModel() Model {
	return Model{
		list:    list.NewModel("Repositories"),
	}
}

type BackMsg struct{}

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

	case []process.Result:

		m.items = msg
		items := []list.Item{}
		for _, resultItem := range msg {
			newItem := list.Item{
				Title: resultItem.Name,
			}
			items = append(items, newItem)
		}
		m.list, _ = m.list.Update(items)
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
				m.Selected = m.items[listMsg.Item]
				cmd = func() tea.Msg {
					return ForwardMsg{
						Payload: m.Selected,
					}
				}
			case list.Back:
				cmd = func() tea.Msg {
					return BackMsg{}
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
	return lipgloss.NewStyle().Render(m.list.View())
}
