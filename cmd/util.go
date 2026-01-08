package main

import (
	"sort"

	kube "github.com/alexei-ozerov/kestrel/internal/kube"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

func extractNameFromGVR(allGVR []kube.ApiResource) []string {
	var allGVRNameList []string
	for _, res := range allGVR {
		allGVRNameList = append(allGVRNameList, res.Name)
	}

	return allGVRNameList
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
