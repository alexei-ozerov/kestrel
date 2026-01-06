package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	kube "github.com/alexei-ozerov/kestrel/internal/kube"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	viewport     viewport.Model
	content      []string
	contentSelection string

	resourceList []kube.ApiResource
	cursorIndex  int

	ready        bool
}

func initializeModel(resources []kube.ApiResource) model {
	view := viewport.New(20, 20)
	
	var content []string
	for _, res := range resources {
		content = append(content, res.Name)
	}

	return model{
		viewport: view,
		resourceList: resources,
		content: content,
		contentSelection: content[len(content) - 1],
		cursorIndex:  len(resources) - 1, // Start at the bottom
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) renderContent() string {
	var b strings.Builder
	for i, item := range m.content {
		if i == m.cursorIndex {
			// Highlighted cursor line
			b.WriteString(selectedItemStyle.Render(fmt.Sprintf("%s", string(item))))
		} else {
			// Normal line
			b.WriteString(itemStyle.Render(fmt.Sprintf("%s", string(item))))
		}
		if i < len(m.content)-1 {
			b.WriteRune('\n')
		}
	}

	return b.String()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursorIndex > 0 {
				m.cursorIndex--
			}

		case "down", "j":
			if m.cursorIndex < len(m.content)-1 {
				m.cursorIndex++
			}

		case "G":
			m.cursorIndex = len(m.content) - 1
			m.syncScroll()
		}

		m.contentSelection = m.content[m.cursorIndex]
		m.viewport.SetContent(m.renderContent())
		m.syncScroll()

	case tea.WindowSizeMsg:
		headerHeight := 4
		footerHeight := 1
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.SetContent(m.renderContent())
			m.viewport.GotoBottom()
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) syncScroll() {
    if m.cursorIndex < m.viewport.YOffset {
        m.viewport.YOffset = m.cursorIndex
    }

	if m.viewport.YOffset < 0 {
    	m.viewport.YOffset = 0
	}

    bottomBoundary := m.viewport.YOffset + m.viewport.Height
    if m.cursorIndex >= bottomBoundary {
        m.viewport.YOffset = m.cursorIndex - m.viewport.Height + 1
    }
    
    m.viewport.SetYOffset(m.viewport.YOffset)
}

func (m model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("Kestrel"),
		m.viewport.View(),
		welcomeStyle.Render("[ Resources ] ") + fmt.Sprintf("%s", m.contentSelection),
	)
}

func main() {
	resources, err := kube.GetK8sDiscoveredResourcesList()
	if err != nil {
		log.Fatal(err)
	}

	p := tea.NewProgram(initializeModel(resources), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
