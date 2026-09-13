package ui

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/bernardofernandezz/tui-ebook-reader/internal/epub"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

//go:embed theme.json
var themeJSON []byte

const (
	headerHeight  = 2
	footerHeight  = 2
	verticalSpace = headerHeight + footerHeight
	maxTextWidth  = 76 // largura do texto; o glamour adiciona 2 colunas de margem
)

var chromeStyle = lipgloss.NewStyle().Faint(true)

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
		glamour.WithStylesFromJSONBytes(themeJSON),
		glamour.WithWordWrap(wrapWidth(width)),
	)
}

// wrapWidth limita a largura do texto para uma leitura confortável.
func wrapWidth(termWidth int) int {
	return min(max(20, termWidth-4), maxTextWidth)
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

	header := chromeStyle.Render(fmt.Sprintf("%s · %s", m.book.Title, ch.Title))
	footer := chromeStyle.Render(fmt.Sprintf(
		"capítulo %d/%d · %.0f%% · q sair",
		m.chapter+1,
		len(m.book.Chapters),
		m.viewport.ScrollPercent()*100,
	))

	return strings.Join([]string{
		lipgloss.PlaceHorizontal(m.width, lipgloss.Center, header),
		"",
		m.viewport.View(),
		"",
		lipgloss.PlaceHorizontal(m.width, lipgloss.Center, footer),
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
