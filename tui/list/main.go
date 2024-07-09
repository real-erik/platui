package list

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/real-erik/platui/tui/styles"
)

type Model struct {
	title string
	list  list.Model

	height int
	width  int
}

func NewModel(title string) Model {
	var listItems []list.Item

	return Model{
		title: title,
		list:  list.New(listItems, list.NewDefaultDelegate(), 0, 0),
	}
}

type Direction = int

const (
	Forward Direction = iota
	Back
)

type Item struct {
	Title string
	Description string
}

type Msg struct {
	Item      int
	Direction Direction
}

// FIXME: why doesn't this work?
func (m Model) setListSize() {
	h, v := styles.DocStyle.GetFrameSize()
	m.list.SetSize(m.width-h, m.height-v)
}


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
	case []Item:
		items := []list.Item{}
		for i, resultItem := range msg {
			newItem := item{
				title: resultItem.Title,
				desc: resultItem.Description,
				id:    i,
			}
			items = append(items, newItem)
		}

		m.list = list.New(items, list.NewDefaultDelegate(), 0, 0)
		m.list.Title = m.title

		// FIXME: why doesn't this work?
		// m.setListSize()
		h, v := styles.DocStyle.GetFrameSize()
		m.list.SetSize(m.width-h, m.height-v)
		return m, nil

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

		// FIXME: why doesn't this work?
		// m.setListSize()
		h, v := styles.DocStyle.GetFrameSize()
		m.list.SetSize(m.width-h, m.height-v)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.list.FilterState() != list.Filtering {
				selected := m.list.SelectedItem().(item)
				cmd := func() tea.Msg {
					return Msg{
						Item:      selected.id,
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

	return styles.DocStyle.Render(m.list.View())
}
