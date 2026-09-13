package ui

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/bernardofernandezz/tui-ebook-reader/internal/bookmarks"
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

type mode int

const (
	modeReading mode = iota
	modeBookmarks
)

type model struct {
	book     *epub.Book
	chapter  int
	viewport viewport.Model
	renderer *glamour.TermRenderer
	store    *bookmarks.Store
	marks    []bookmarks.Bookmark
	mode     mode
	cursor   int
	offset   int // rolagem salva ao abrir a lista de bookmarks
	ready    bool
	width    int
	height   int
}

func New(book *epub.Book) model {
	store := bookmarks.Load()
	return model{
		book:  book,
		store: store,
		marks: store.List(book.Path),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.mode == modeBookmarks {
			return m.updateBookmarks(msg)
		}
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
		case "b":
			// "b" também é page-up no viewport; por isso não repassamos a tecla
			m.toggleBookmark()
			return m, nil
		case "B":
			m.openBookmarks()
			return m, nil
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

var imageRefPattern = regexp.MustCompile(`!\[[^\]]*\]\(([^)\s]+)[^)]*\)`)

func (m *model) setContent() {
	ch := m.book.Chapters[m.chapter]
	text := m.book.Text(ch)

	if m.renderer != nil {
		text = m.renderMarkdown(ch, text)
	}

	// centraliza a coluna de leitura no terminal
	centered := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(text)

	m.viewport.SetContent(centered)
	m.viewport.GotoTop()
}

// renderMarkdown renderiza o capítulo e troca cada referência de imagem pelo
// desenho em blocos ANSI correspondente.
func (m *model) renderMarkdown(ch epub.Chapter, md string) string {
	var out strings.Builder
	last := 0
	for _, loc := range imageRefPattern.FindAllStringSubmatchIndex(md, -1) {
		m.renderSegment(&out, md[last:loc[0]])

		ref := md[loc[2]:loc[3]]
		if img, err := m.book.Image(ch, ref); err == nil {
			out.WriteString(renderImage(img))
		}
		last = loc[1]
	}
	m.renderSegment(&out, md[last:])
	return out.String()
}

func (m *model) renderSegment(out *strings.Builder, md string) {
	if strings.TrimSpace(md) == "" {
		return
	}
	rendered, err := m.renderer.Render(md)
	if err != nil {
		out.WriteString(md)
		return
	}
	out.WriteString(rendered)
}

func (m model) updateBookmarks(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "B":
		m.closeBookmarks()
	case "j", "down":
		if m.cursor < len(m.marks)-1 {
			m.cursor++
			m.setBookmarkList()
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
			m.setBookmarkList()
		}
	case "enter":
		m.jumpToBookmark()
	}
	return m, nil
}

func (m *model) toggleBookmark() {
	m.store.Toggle(m.book.Path, bookmarks.Bookmark{
		Chapter: m.chapter,
		Percent: int(m.viewport.ScrollPercent() * 100),
		Label:   m.book.Chapters[m.chapter].Title,
	})
	m.marks = m.store.List(m.book.Path)
}

func (m *model) openBookmarks() {
	m.mode = modeBookmarks
	m.offset = m.viewport.YOffset
	m.cursor = 0
	m.setBookmarkList()
}

func (m *model) closeBookmarks() {
	m.mode = modeReading
	m.setContent()
	m.viewport.SetYOffset(m.offset)
}

func (m *model) jumpToBookmark() {
	if len(m.marks) == 0 {
		return
	}
	bm := m.marks[m.cursor]
	m.chapter = bm.Chapter
	m.mode = modeReading
	m.setContent()
	m.scrollToPercent(bm.Percent)
}

func (m *model) scrollToPercent(percent int) {
	maxOffset := m.viewport.TotalLineCount() - m.viewport.VisibleLineCount()
	if maxOffset < 0 {
		maxOffset = 0
	}
	m.viewport.SetYOffset(percent * maxOffset / 100)
}

func (m *model) setBookmarkList() {
	var lines []string
	if len(m.marks) == 0 {
		lines = append(lines, "nenhum bookmark ainda; pressione b durante a leitura")
	}
	for i, bm := range m.marks {
		line := fmt.Sprintf("%s · capítulo %d · %d%%", bm.Label, bm.Chapter+1, bm.Percent)
		if i == m.cursor {
			line = lipgloss.NewStyle().Bold(true).Render("▸ " + line)
		} else {
			line = "  " + line
		}
		lines = append(lines, line)
	}

	content := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(strings.Join(lines, "\n"))

	m.viewport.SetContent(content)
	m.viewport.GotoTop()
}

func (m model) hasBookmark(chapter int) bool {
	for _, b := range m.marks {
		if b.Chapter == chapter {
			return true
		}
	}
	return false
}

func (m model) View() string {
	if !m.ready {
		return "carregando..."
	}

	ch := m.book.Chapters[m.chapter]

	var header, footer string
	if m.mode == modeBookmarks {
		header = chromeStyle.Render(fmt.Sprintf("bookmarks · %s", m.book.Title))
		footer = chromeStyle.Render("j/k mover · enter abrir · esc fechar")
	} else {
		star := ""
		if m.hasBookmark(m.chapter) {
			star = " · ★"
		}
		header = chromeStyle.Render(fmt.Sprintf("%s · %s", m.book.Title, ch.Title))
		footer = chromeStyle.Render(fmt.Sprintf(
			"capítulo %d/%d · %.0f%%%s · b marca · B lista · q sair",
			m.chapter+1,
			len(m.book.Chapters),
			m.viewport.ScrollPercent()*100,
			star,
		))
	}

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
