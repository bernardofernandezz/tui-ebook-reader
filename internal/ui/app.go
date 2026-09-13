package ui

import (
	"fmt"
	"strings"

	"github.com/bernardofernandezz/tui-ebook-reader/internal/epub"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	book     *epub.Book
	chapter  int
	viewport viewport.Model
	ready    bool
	width    int
	height   int
}

func New(book *epub.Book) model {
	return model{
		book:    book,
		chapter: 0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "right", "l", "n":
			if m.chapter < len(m.book.Chapters)-1 {
				m.chapter++
				m.setContent()
			}
		case "left", "h", "p":
			if m.chapter > 0 {
				m.chapter--
				m.setContent()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := 3
		footerHeight := 2
		verticalMargin := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMargin)
			m.viewport.YPosition = headerHeight
			m.ready = true
			m.setContent()
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMargin
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *model) setContent() {
	ch := m.book.Chapters[m.chapter]
	text := m.book.Text(ch)
	m.viewport.SetContent(text)
	m.viewport.GotoTop()
}

func (m model) View() string {
	if !m.ready {
		return "carregando..."
	}

	ch := m.book.Chapters[m.chapter]

	header := lipgloss.NewStyle().
		Bold(true).
		Render(fmt.Sprintf("%s — %s", m.book.Title, ch.Title))

	footer := lipgloss.NewStyle().
		Faint(true).
		Render(fmt.Sprintf(
			"capítulo %d/%d  |  ←/→ ou h/l  |  q sair  |  %3.f%%",
			m.chapter+1,
			len(m.book.Chapters),
			m.viewport.ScrollPercent()*100,
		))

	return strings.Join([]string{
		header,
		strings.Repeat("─", max(10, m.width)),
		m.viewport.View(),
		strings.Repeat("─", max(10, m.width)),
		footer,
	}, "\n")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Run(book *epub.Book) error {
	p := tea.NewProgram(
		New(book),
		tea.WithAltScreen(),
	)
	_, err := p.Run()
	return err
}
