package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Issue struct {
	ID       string
	Title    string
	Body     string
	Expanded bool
}

type model struct {
	inputs        []textinput.Model
	viewport      viewport.Model
	focusIndex    int
	listIndex     int
	issues        []Issue
	width, height int
	ready         bool
}

var (
	inputBoxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	activeBox     = lipgloss.NewStyle().BorderForeground(lipgloss.Color("205"))
	issueBoxStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).MarginBottom(1).Padding(0, 1)
	activeIssue   = lipgloss.NewStyle().BorderForeground(lipgloss.Color("62")).Background(lipgloss.Color("235"))
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("62"))
)

func initialModel() model {
	m := model{
		focusIndex: 0,
		listIndex:  -1,
		inputs:     make([]textinput.Model, 2),
	}
	for i := range m.inputs {
		t := textinput.New()
		t.Placeholder = []string{"Name/ID", "Description"}[i]
		if i == 0 {
			t.Focus()
		}
		m.inputs[i] = t
	}
	return m
}

func (m model) Init() tea.Cmd { return textinput.Blink }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Initialize/Update Viewport Size
		// We subtract space for header, inputs, and footer
		headerHeight := 8
		footerHeight := 2
		if !m.ready {
			m.viewport = viewport.New(msg.Width-6, msg.Height-headerHeight-footerHeight)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 6
			m.viewport.Height = msg.Height - headerHeight - footerHeight
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab":
			m.listIndex = -1
			m.inputs[m.focusIndex].Blur()
			m.focusIndex = (m.focusIndex + 1) % 2
			m.inputs[m.focusIndex].Focus()

		case "down":
			if m.listIndex < len(m.issues)-1 {
				m.inputs[m.focusIndex].Blur()
				if m.listIndex != -1 {
					m.issues[m.listIndex].Expanded = false
				}
				m.listIndex++
				// Sync viewport to focus
				m.viewport.SetContent(m.renderList())
			}
		case "up":
			if m.listIndex > 0 {
				m.issues[m.listIndex].Expanded = false
				m.listIndex--
				m.viewport.SetContent(m.renderList())
			} else if m.listIndex == 0 {
				m.listIndex = -1
				m.inputs[m.focusIndex].Focus()
			}

		case "ctrl+d":
			if m.listIndex != -1 {
				m.issues[m.listIndex].Expanded = !m.issues[m.listIndex].Expanded
				m.viewport.SetContent(m.renderList())
			}

		case "ctrl+p":
			newIssue := Issue{
				ID:    fmt.Sprintf("#%d", len(m.issues)+101),
				Title: m.inputs[0].Value(),
				Body:  m.inputs[1].Value() + " " + strings.Repeat("Extended log detail for scrolling test. ", 8),
			}
			m.issues = append([]Issue{newIssue}, m.issues...)
			m.viewport.SetContent(m.renderList())
			m.viewport.GotoTop()
		}
	}

	// Update the appropriate component
	if m.listIndex == -1 {
		m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.viewport, cmd = m.viewport.Update(msg) // Handles Mouse Wheel scrolling
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// Helper to render the string used inside the viewport
func (m model) renderList() string {
	if len(m.issues) == 0 {
		return "\n\n  No issues yet. Use Shift+Enter to post dummy data."
	}

	var listItems []string
	for i, issue := range m.issues {
		style := issueBoxStyle.Copy().Width(m.width - 10)
		if i == m.listIndex {
			style = style.Inherit(activeIssue)
		}

		content := issue.Body
		if !issue.Expanded {
			if len(content) > 60 {
				content = content[:60] + "..."
			}
		}

		display := fmt.Sprintf("%s %s\n%s", titleStyle.Render(issue.ID), issue.Title, content)
		listItems = append(listItems, style.Render(display))
	}
	return strings.Join(listItems, "\n")
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// 1. Inputs
	boxWidth := (m.width / 2) - 4
	var inputViews []string
	for i := range m.inputs {
		style := inputBoxStyle.Copy().Width(boxWidth)
		if i == m.focusIndex && m.listIndex == -1 {
			style = style.Inherit(activeBox)
		}
		inputViews = append(inputViews, style.Render(m.inputs[i].View()))
	}
	inputArea := lipgloss.JoinHorizontal(lipgloss.Top, inputViews...)

	// 2. Assembly
	header := titleStyle.Render(" SYSTEM FEED ")
	mainContent := lipgloss.JoinVertical(lipgloss.Center,
		header,
		"\n",
		inputArea,
		"\n",
		m.viewport.View(), // This is the scrollable area
	)

	footer := lipgloss.NewStyle().Background(lipgloss.Color("235")).Width(m.width).
		Render(" Shift+Enter: Post • Up/Down: Select • Ctrl+D: Expand • Mouse Scroll: Enabled")

	gap := m.height - lipgloss.Height(mainContent) - lipgloss.Height(footer)
	if gap < 0 {
		gap = 0
	}

	return mainContent + strings.Repeat("\n", gap) + footer
}

// func main() {
// 	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
// 	if _, err := p.Run(); err != nil {
// 		fmt.Printf("Error: %v", err)
// 	}
// }
