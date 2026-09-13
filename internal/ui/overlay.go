package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// overlay é uma lista navegável mostrada no lugar do texto: sumário,
// bookmarks e ajuda. Com pick == nil a lista é apenas informativa.
type overlay struct {
	title  string
	items  []string
	cursor int
	pick   func(i int)
}

func (o *overlay) move(delta int) {
	o.cursor = min(max(0, o.cursor+delta), max(0, len(o.items)-1))
}

func (o *overlay) view(width, textWidth int) string {
	lines := make([]string, len(o.items))
	for i, item := range o.items {
		if o.pick != nil && i == o.cursor {
			lines[i] = lipgloss.NewStyle().Bold(true).Render("▸ " + item)
		} else {
			lines[i] = "  " + item
		}
	}

	block := lipgloss.NewStyle().
		Width(textWidth).
		Align(lipgloss.Left).
		Render(strings.Join(lines, "\n"))

	return lipgloss.PlaceHorizontal(width, lipgloss.Center, block)
}
