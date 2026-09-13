package ui

import (
	"fmt"
	"strings"

	"github.com/bernardofernandezz/tui-ebook-reader/internal/epub"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

const (
	headerHeight  = 3
	footerHeight  = 2
	verticalSpace = headerHeight + footerHeight
	maxLineWidth  = 96 // largura máxima da coluna de leitura
)

type model struct {
	book     *epub.Book
	chapter  int
	viewport viewport.Model
	renderer *glamour.TermRenderer
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
		widthChanged := msg.Width != m.width
		m.width = msg.Width
		m.height = msg.Height

		viewportHeight := max(1, m.height-verticalSpace)
		if !m.ready {
			m.viewport = viewport.New(m.width, viewportHeight)
			m.viewport.YPosition = headerHeight
			m.ready = true
		} else {
			m.viewport.Width = m.width
			m.viewport.Height = viewportHeight
		}

		// o renderer depende da largura; ao mudar, re-renderiza o capítulo
		if widthChanged {
			if renderer, err := newRenderer(m.width); err == nil {
				m.renderer = renderer
			}
			m.setContent()
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func newRenderer(width int) (*glamour.TermRenderer, error) {
	return glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(wrapWidth(width)),
	)
}

// wrapWidth limita a largura do texto para uma leitura confortável.
func wrapWidth(termWidth int) int {
	return min(max(10, termWidth-4), maxLineWidth)
}

func (m *model) setContent() {
	ch := m.book.Chapters[m.chapter]
	text := m.book.Text(ch)

	if m.renderer != nil {
		if rendered, err := m.renderer.Render(text); err == nil {
			text = rendered
		}
	}

	// centraliza a coluna de leitura no terminal
	centered := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(text)

	m.viewport.SetContent(centered)
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
			"capítulo %d/%d  |  ←/→ capítulo  |  j/k, d/u, espaço rolam  |  q sair  |  %3.f%%",
			m.chapter+1,
			len(m.book.Chapters),
			m.viewport.ScrollPercent()*100,
		))

	rule := strings.Repeat("─", max(10, m.width))

	return strings.Join([]string{
		header,
		rule,
		m.viewport.View(),
		rule,
		footer,
	}, "\n")
}

func Run(book *epub.Book) error {
	p := tea.NewProgram(
		New(book),
		tea.WithAltScreen(),
	)
	_, err := p.Run()
	return err
}
