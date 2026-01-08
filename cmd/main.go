package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	kube "github.com/alexei-ozerov/kestrel/internal/kube"

	"github.com/lithammer/fuzzysearch/fuzzy"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState int

const (
	resourceTypeView sessionState = iota
	resourceListView
)

type model struct {
	cursor          int
	ready           bool
	state           sessionState
	viewportContent []string
	filteredContent []string

	selectedGVR      string
	selectedInstance string

	viewport    viewport.Model
	searchInput textinput.Model

	rawGVRData []kube.ApiResource
}

func extractNameFromGVR(allGVR []kube.ApiResource) []string {
	var allGVRNameList []string
	for _, res := range allGVR {
		allGVRNameList = append(allGVRNameList, res.Name)
	}

	return allGVRNameList
}

func initializeModel(resources []kube.ApiResource) model {
	view := viewport.New(20, 20)

	return model{
		viewport:        view,
		viewportContent: extractNameFromGVR(resources),
		cursor:          len(resources) - 1, // Start at the bottom
		searchInput:     textinput.New(),
		rawGVRData:      resources,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) renderContent() string {
	var b strings.Builder
	for i, item := range m.viewportContent {
		if i == m.cursor {
			b.WriteString(selectedItemStyle.Render(fmt.Sprintf("%s", string(item))))
		} else {
			b.WriteString(itemStyle.Render(fmt.Sprintf("%s", string(item))))
		}
		if i < len(m.viewportContent)-1 {
			b.WriteRune('\n')
		}
	}

	return b.String()
}

// TODO (ozerova): Improve this!
func fuzzyFind(searchTerm string, data []string) []string {
	rankedList := fuzzy.RankFind(searchTerm, data)
	sort.Sort(rankedList)

	var sortedList []string
	for _, entry := range rankedList {
		sortedList = append(sortedList, entry.Target)
	}

	return sortedList
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if m.searchInput.Focused() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.searchInput.Blur()
				m.searchInput.Reset()
				m.viewportContent = extractNameFromGVR(m.rawGVRData) // Should this data be here?
				m.viewport.SetContent(m.renderContent())
				return m, nil

			case "enter":
				if len(m.filteredContent) > 0 {
					m.searchInput.Blur()
					m.viewportContent = m.filteredContent
					m.cursor = 0
					m.viewport.SetContent(m.renderContent())
					return m, nil
				}
			}
		}

		m.searchInput, cmd = m.searchInput.Update(msg)

		term := m.searchInput.Value()
		if term == "" {
			m.viewportContent = extractNameFromGVR(m.rawGVRData)
		} else {
			m.filteredContent = fuzzyFind(term, extractNameFromGVR(m.rawGVRData))
			m.viewportContent = m.filteredContent
		}

		m.cursor = 0
		m.viewport.SetContent(m.renderContent())

		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.syncScroll()
				m.selectedGVR = m.viewportContent[m.cursor]
				m.viewport.SetContent(m.renderContent())
			}

		case "down", "j":
			if m.cursor < len(m.viewportContent)-1 {
				m.cursor++
				m.syncScroll()
				m.selectedGVR = m.viewportContent[m.cursor]
				m.viewport.SetContent(m.renderContent())
			}

		case "G":
			m.cursor = len(m.viewportContent) - 1
			m.syncScroll()
			m.selectedGVR = m.viewportContent[m.cursor]
			m.viewport.SetContent(m.renderContent())

		case "/":
			return m, m.searchInput.Focus()

		// Reset after search
		case "esc":
			m.viewportContent = extractNameFromGVR(m.rawGVRData)
			m.viewport.SetContent(m.renderContent())
			m.cursor = len(m.viewportContent) - 1
		}

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
	if m.cursor < m.viewport.YOffset {
		m.viewport.YOffset = m.cursor
	}

	if m.viewport.YOffset < 0 {
		m.viewport.YOffset = 0
	}

	bottomBoundary := m.viewport.YOffset + m.viewport.Height
	if m.cursor >= bottomBoundary {
		m.viewport.YOffset = m.cursor - m.viewport.Height + 1
	}

	m.viewport.SetYOffset(m.viewport.YOffset)
}

func (m model) View() string {
	var footer string
	if m.searchInput.Focused() {
		footer = welcomeStyle.Render("[ Resources ] ") + fmt.Sprintf("%s", m.searchInput.Value())
	} else {
		footer = welcomeStyle.Render("[ Resources ] ") + fmt.Sprintf("%s", m.selectedGVR)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("Kestrel"),
		m.viewport.View(),
		footer,
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
