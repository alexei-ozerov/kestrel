package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	kube "github.com/alexei-ozerov/kestrel/internal/kube"

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
	cursor int
	ready  bool
	state  sessionState

	viewportContent []string
	filteredContent []string

	modeText string

	selectedGVR      string
	selectedInstance string

	viewport    viewport.Model
	searchInput textinput.Model

	rawGVRData []kube.ApiResource

	version string
}

func initializeModel(resources []kube.ApiResource) model {
	view := viewport.New(20, 20)

	initialSelection := ""
	if len(resources) > 0 {
		initialSelection = extractNameFromGVR(resources)[len(resources)-1]
	}

	return model{
		version:         "0.0.1", // TODO: extract from config or somewhere else
		viewport:        view,
		viewportContent: extractNameFromGVR(resources),
		cursor:          len(resources) - 1, // Start at the bottom
		searchInput:     textinput.New(),
		rawGVRData:      resources,
		modeText:        "[ RESOURCES ]",
		selectedGVR:     initialSelection,
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

func (m *model) ResetViewport() {
	m.viewportContent = extractNameFromGVR(m.rawGVRData)
	m.cursor = len(m.rawGVRData) - 1
	m.selectedGVR = m.viewportContent[m.cursor]

	m.viewport.SetContent(m.renderContent())
	m.syncScroll()
}

func (m *model) RefreshViewport() {
	m.syncScroll()
	m.selectedGVR = m.viewportContent[m.cursor]
	m.viewport.SetContent(m.renderContent())
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

				m.modeText = "[ RESOURCES ]"
				m.ResetViewport()

				m.viewport, cmd = m.viewport.Update(msg)
				cmds = append(cmds, cmd)

				return m, tea.Batch(cmds...)
			case "enter":
				if len(m.filteredContent) > 0 {
					m.searchInput.Blur()
					// TODO (ozerova): should I do a m.searchInput.Reset() here?

					m.viewportContent = m.filteredContent
					m.cursor = 0
					m.modeText = "[ RESOURCES ]"

					m.RefreshViewport()

					m.viewport, cmd = m.viewport.Update(msg)
					cmds = append(cmds, cmd)

					return m, tea.Batch(cmds...)
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
		m.RefreshViewport()

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
				m.RefreshViewport()
			}

		case "down", "j":
			if m.cursor < len(m.viewportContent)-1 {
				m.cursor++
				m.RefreshViewport()
			}

		case "g":
			// TODO: implement double tap logic
			m.cursor = 0
			m.RefreshViewport()

		case "G":
			m.cursor = len(m.viewportContent) - 1
			m.RefreshViewport()

		case "/":
			m.modeText = "[ SEARCH ]"
			return m, m.searchInput.Focus()

		// Reset after search
		case "esc":
			m.ResetViewport()
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

func (m model) View() string {
	header := titleStyle.Render("\nKestrel") + fmt.Sprintf(" v%s\n", m.version)

	var footer string
	if m.searchInput.Focused() {
		footer = welcomeStyle.Render(m.modeText) + fmt.Sprintf(" %s", m.searchInput.Value())
	} else {
		footer = welcomeStyle.Render(m.modeText) + fmt.Sprintf(" %s", m.selectedGVR)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		m.viewport.View(),
		footer,
	)
}

func main() {
	resources, err := kube.GetK8sDiscoveredResourcesList()
	if err != nil {
		log.Fatal(err)
	}

	p := tea.NewProgram(initializeModel(resources))
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
