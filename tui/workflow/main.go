package workflow

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/real-erik/platui/process"

	"github.com/real-erik/platui/tui/list"
)

type Model struct {
	loading bool
	list    list.Model
	items   []process.Result
}

func NewModel() Model {
	return Model{
		list:    list.NewModel("Workflows"),
		loading: true,
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
	case []process.Result:
		m.loading = false

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
				cmd = func() tea.Msg {
					return ForwardMsg{
						Payload: m.items[listMsg.Item],
					}
				}
			case list.Back:
				m.loading = true
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
	if m.loading {
		return fmt.Sprintf("Loading workflows...")
	}
	return lipgloss.NewStyle().Render(m.list.View())
}
