package list

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/real-erik/platui/process"
)

type Model struct {
	title string
	items []process.Result
	list  list.Model

	height int
	width  int
}

func NewModel(title string) Model {
	var listItems []list.Item

	return Model{
		title: title,
		items: []process.Result{},
		list:  list.New(listItems, list.NewDefaultDelegate(), 0, 0),
	}
}

type Direction = int

const (
	Forward Direction = iota
	Back
)

type Msg struct {
	Item      process.Result
	Direction Direction
}

// FIXME: why doesn't this work?
func (m Model) setListSize() {
	h, v := docStyle.GetFrameSize()
	m.list.SetSize(m.width-h, m.height-v)
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	title, desc string
	id          int
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case []process.Result:
		m.items = msg

		items := []list.Item{}
		for i, resultItem := range m.items {
			newItem := item{
				title: resultItem.Name,
				id:    i,
			}
			items = append(items, newItem)
		}

		m.list = list.New(items, list.NewDefaultDelegate(), 0, 0)
		m.list.Title = m.title

		// FIXME: why doesn't this work?
		// m.setListSize()
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(m.width-h, m.height-v)
		return m, nil

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

		// FIXME: why doesn't this work?
		// m.setListSize()
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(m.width-h, m.height-v)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.list.FilterState() != list.Filtering {
				selected := m.list.SelectedItem().(item)
				cmd := func() tea.Msg {
					return Msg{
						Item:      m.items[selected.id],
						Direction: Forward,
					}
				}
				return m, cmd
			}
		case "esc":
			if m.list.FilterState() == list.Unfiltered {
				cmd := func() tea.Msg {
					return Msg{
						Direction: Back,
					}
				}
				return m, cmd
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {

	return docStyle.Render(m.list.View())
}
