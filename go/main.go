package main

import (
	"C"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"time"
)

const (
	Add = iota
	Update
	Remove
)

type PyMsg struct {
	Type     int
	ID       int
	Value    int
	Label    string
	ParentID int
}

type Bar struct {
	ID       int
	Current  int
	Total    int
	Label    string
	ParentID int
	Children []*Bar
}

type model struct {
	bars     []*Bar
	barMap   map[int]*Bar
	rootBars []*Bar
}

var eventChan chan PyMsg

func (m model) Init() tea.Cmd {
	return waitForPyMsg(eventChan)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case PyMsg:
		newModel := m.handlePyMsg(msg)
		return newModel, waitForPyMsg(eventChan)
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) handlePyMsg(msg PyMsg) model {
	switch msg.Type {
	case 0:
		bar := &Bar{
			ID:       msg.ID,
			Current:  0,
			Total:    msg.Value,
			Label:    msg.Label,
			ParentID: msg.ParentID,
			Children: make([]*Bar, 0),
		}
		m.barMap[msg.ID] = bar
		if msg.ParentID >= 0 {
			if parent, exists := m.barMap[msg.ParentID]; exists {
				parent.Children = append(parent.Children, bar)
			}
		} else {
			m.rootBars = append(m.rootBars, bar)
		}
		m.bars = append(m.bars, bar)
	case 1:
		if bar, exists := m.barMap[msg.ID]; exists {
			bar.Current += msg.Value
			if bar.Current > bar.Total {
				bar.Current = bar.Total
			}
		}
	case 2:
		if bar, exists := m.barMap[msg.ID]; exists {
			for len(bar.Children) > 0 {
				childID := bar.Children[0].ID
				if _, childExists := m.barMap[childID]; childExists {
					bar.Children = bar.Children[1:]
					for i, b := range m.bars {
						if b.ID == childID {
							m.bars = append(m.bars[:i], m.bars[i+1:]...)
							break
						}
					}
					delete(m.barMap, childID)
				}
			}
			if bar.ParentID >= 0 {
				if parent, exists := m.barMap[bar.ParentID]; exists {
					for i, child := range parent.Children {
						if child.ID == msg.ID {
							parent.Children = append(parent.Children[:i], parent.Children[i+1:]...)
							break
						}
					}
				}
			} else {
				for i, rootBar := range m.rootBars {
					if rootBar.ID == msg.ID {
						m.rootBars = append(m.rootBars[:i], m.rootBars[i+1:]...)
						break
					}
				}
			}
			for i, b := range m.bars {
				if b.ID == msg.ID {
					m.bars = append(m.bars[:i], m.bars[i+1:]...)
					break
				}
			}
			delete(m.barMap, msg.ID)
		}
	}
	return m
}

func (m model) View() string {
	if len(m.rootBars) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true).
			Padding(1)
		return emptyStyle.Render("⏳ No active progress bars\n")
	}
	var sb strings.Builder
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("62")).
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)
	sb.WriteString(headerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	sb.WriteString("\n")
	sb.WriteString(headerStyle.Render("  📊 Progress Monitor"))
	sb.WriteString("\n")
	sb.WriteString(headerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	sb.WriteString("\n\n")
	m.renderBarsRecursive(&sb, m.rootBars, 0)
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		PaddingTop(1)
	sb.WriteString("\n")
	sb.WriteString(footerStyle.Render("  Press 'q' or Ctrl+C to quit"))
	return sb.String()
}

func (m model) renderBarsRecursive(sb *strings.Builder, bars []*Bar, indent int) {
	for i, bar := range bars {
		var prefix string
		if indent > 0 {
			prefix = strings.Repeat("  ", indent-1)
			if i == len(bars)-1 {
				prefix += "└─ "
			} else {
				prefix += "├─ "
			}
		} else {
			prefix = "  "
		}
		sb.WriteString(prefix + renderBar(bar, indent) + "\n")
		if len(bar.Children) > 0 {
			m.renderBarsRecursive(sb, bar.Children, indent+1)
		}
	}
}

func waitForPyMsg(ch chan PyMsg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func renderBar(b *Bar, indent int) string {
	var percent float64
	var filled, width int
	width = 45 - (indent * 2)
	if width < 20 {
		width = 20
	}
	if b.Total <= 0 {
		percent = 0
		filled = 0
	} else {
		percent = float64(b.Current) / float64(b.Total) * 100
		if percent > 100 {
			percent = 100
		}
		filled = int(float64(b.Current) / float64(b.Total) * float64(width))
		if filled > width {
			filled = width
		}
	}
	var barColor, labelColor string
	if indent == 0 {
		if percent >= 100 {
			barColor = "42"
			labelColor = "42"
		} else if percent >= 75 {
			barColor = "39"
			labelColor = "39"
		} else if percent >= 50 {
			barColor = "33"
			labelColor = "33"
		} else {
			barColor = "62"
			labelColor = "62"
		}
	} else {
		if percent >= 100 {
			barColor = "46"
			labelColor = "46"
		} else if percent >= 75 {
			barColor = "36"
			labelColor = "36"
		} else if percent >= 50 {
			barColor = "178"
			labelColor = "178"
		} else {
			barColor = "135"
			labelColor = "135"
		}
	}
	var barStr strings.Builder
	empty := width - filled
	if filled > 0 {
		barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(barColor))
		filledChars := strings.Repeat("█", filled)
		barStr.WriteString(barStyle.Render(filledChars))
	}
	if filled < width && filled > 0 {
		indicatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(barColor))
		barStr.WriteString(indicatorStyle.Render("▌"))
		empty--
	}
	if empty > 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
		barStr.WriteString(emptyStyle.Render(strings.Repeat("░", empty)))
	}
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(labelColor)).
		Bold(indent == 0)
	label := labelStyle.Render(b.Label)
	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	stats := statsStyle.Render(fmt.Sprintf("%d/%d", b.Current, b.Total))
	percentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(barColor)).
		Bold(true)
	percentStr := percentStyle.Render(fmt.Sprintf("%5.1f%%", percent))
	barContainer := fmt.Sprintf("[%s]", barStr.String())
	return fmt.Sprintf("%s %s %s %s", label, barContainer, stats, percentStr)
}

//export StartEngine
func StartEngine() {
	eventChan = make(chan PyMsg, 1000)
	go func() {
		initialModel := model{
			bars:     make([]*Bar, 0),
			barMap:   make(map[int]*Bar),
			rootBars: make([]*Bar, 0),
		}
		p := tea.NewProgram(initialModel, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
		}
	}()
	time.Sleep(50 * time.Millisecond)
}

//export AddBar
func AddBar(id int, total int, label *C.char) {
	eventChan <- PyMsg{Type: 0, ID: id, Value: total, Label: C.GoString(label), ParentID: -1}
}

//export AddBarWithParent
func AddBarWithParent(id int, total int, label *C.char, parentID int) {
	eventChan <- PyMsg{Type: 0, ID: id, Value: total, Label: C.GoString(label), ParentID: parentID}
}

//export UpdateBar
func UpdateBar(id int, val int) {
	eventChan <- PyMsg{Type: 1, ID: id, Value: val}
}

//export RemoveBar
func RemoveBar(id int) {
	eventChan <- PyMsg{Type: 2, ID: id}
}

func main() {}
