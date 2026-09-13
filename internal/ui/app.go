package ui

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/bernardofernandezz/tui-ebook-reader/internal/epub"
	"github.com/bernardofernandezz/tui-ebook-reader/internal/store"
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
	barWidth      = 12
)

var chromeStyle = lipgloss.NewStyle().Faint(true)

var helpLines = []string{
	"←/→ ou h/l   capítulo anterior / próximo",
	"c            lista de capítulos",
	"j/k ou ↑/↓   rolar uma linha",
	"d/u          meia página",
	"espaço       página para baixo",
	"pgup         página para cima",
	"b            marcar / desmarcar bookmark",
	"B            lista de bookmarks",
	"?            esta ajuda",
	"q            sair",
}

type model struct {
	book     *epub.Book
	cfg      *store.Config
	st       *store.State
	chapter  int
	viewport viewport.Model
	renderer *glamour.TermRenderer
	overlay  *overlay
	coverArt string
	offset   int // rolagem salva ao abrir um overlay
	cacheCh  int
	cacheW   int
	cache    string
	ready    bool
	width    int
	height   int
}

func New(book *epub.Book, cfg *store.Config, st *store.State) *model {
	m := &model{
		book:    book,
		cfg:     cfg,
		st:      st,
		cacheCh: -1,
	}
	if cover := book.Cover(); cover != nil {
		m.coverArt = renderImage(cover)
	}
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.overlay != nil {
			return m.updateOverlay(msg)
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "right", "l", "n":
			m.goToChapter(m.chapter + 1)
		case "left", "h", "p":
			m.goToChapter(m.chapter - 1)
		case "b":
			// "b" também é page-up no viewport; por isso não repassamos a tecla
			m.toggleBookmark()
			return m, nil
		case "B":
			m.openBookmarks()
			return m, nil
		case "c":
			m.openChapters()
			return m, nil
		case "?":
			m.openHelp()
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

		// o renderer depende da largura; ao mudar, re-renderiza o conteúdo
		if widthChanged {
			if renderer, err := newRenderer(m.width); err == nil {
				m.renderer = renderer
			}
			if m.overlay != nil {
				m.setOverlayContent()
			} else {
				m.setContent()
			}
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

func (m *model) goToChapter(chapter int) {
	if chapter < 0 || chapter >= len(m.book.Chapters) {
		return
	}
	m.chapter = chapter
	m.setContent()
	m.savePosition()
}

var imageRefPattern = regexp.MustCompile(`!\[[^\]]*\]\(([^)\s]+)[^)]*\)`)

func (m *model) setContent() {
	content := m.chapterContent()

	if m.chapter == 0 && m.coverArt != "" {
		content = m.coverArt + "\n" + content
	}

	// centraliza a coluna de leitura no terminal
	centered := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(content)

	m.viewport.SetContent(centered)
	m.viewport.GotoTop()
}

// chapterContent renderiza o capítulo atual, com cache por capítulo/largura.
func (m *model) chapterContent() string {
	if m.cacheCh == m.chapter && m.cacheW == m.width {
		return m.cache
	}

	ch := m.book.Chapters[m.chapter]
	text := m.book.Text(ch)
	if m.renderer != nil {
		text = m.renderMarkdown(ch, text)
	}

	m.cacheCh, m.cacheW, m.cache = m.chapter, m.width, text
	return text
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

func (m *model) updateOverlay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.closeOverlay()
	case "j", "down":
		m.overlay.move(1)
		m.setOverlayContent()
	case "k", "up":
		m.overlay.move(-1)
		m.setOverlayContent()
	case "enter":
		if m.overlay.pick != nil {
			m.overlay.pick(m.overlay.cursor)
		}
	default:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *model) openOverlay(title string, items []string, pick func(i int)) {
	m.offset = m.viewport.YOffset
	m.overlay = &overlay{title: title, items: items, pick: pick}
	m.setOverlayContent()
}

func (m *model) closeOverlay() {
	m.overlay = nil
	m.setContent()
	m.viewport.SetYOffset(m.offset)
}

func (m *model) setOverlayContent() {
	m.viewport.SetContent(m.overlay.view(m.width))

	// mantém o item selecionado visível
	cursor := m.overlay.cursor
	switch {
	case cursor < m.viewport.YOffset:
		m.viewport.SetYOffset(cursor)
	case cursor >= m.viewport.YOffset+m.viewport.Height:
		m.viewport.SetYOffset(cursor - m.viewport.Height + 1)
	}
}

func (m *model) openChapters() {
	items := make([]string, len(m.book.Chapters))
	for i, ch := range m.book.Chapters {
		items[i] = fmt.Sprintf("%2d  %s", i+1, ch.Title)
	}
	m.openOverlay("capítulos", items, func(i int) {
		m.overlay = nil
		m.goToChapter(i)
	})
}

func (m *model) openBookmarks() {
	marks := m.st.Book(m.book.Path).SortedBookmarks()
	if len(marks) == 0 {
		m.openOverlay("bookmarks", []string{"nenhum bookmark ainda; pressione b durante a leitura"}, nil)
		return
	}

	items := make([]string, len(marks))
	for i, bm := range marks {
		items[i] = fmt.Sprintf("%s · capítulo %d · %d%%", bm.Label, bm.Chapter+1, bm.Percent)
	}
	m.openOverlay("bookmarks", items, func(i int) {
		bm := marks[i]
		m.overlay = nil
		m.chapter = bm.Chapter
		m.setContent()
		m.scrollToPercent(bm.Percent)
		m.savePosition()
	})
}

func (m *model) openHelp() {
	m.openOverlay("ajuda", helpLines, nil)
}

func (m *model) toggleBookmark() {
	ch := m.book.Chapters[m.chapter]
	m.st.Book(m.book.Path).ToggleBookmark(store.Bookmark{
		Chapter: m.chapter,
		Percent: int(m.viewport.ScrollPercent() * 100),
		Label:   ch.Title,
	})
	_ = m.st.Save()
}

func (m *model) scrollToPercent(percent int) {
	maxOffset := m.viewport.TotalLineCount() - m.viewport.VisibleLineCount()
	if maxOffset < 0 {
		maxOffset = 0
	}
	m.viewport.SetYOffset(percent * maxOffset / 100)
}

func (m model) hasBookmark(chapter int) bool {
	return m.st.Book(m.book.Path).HasBookmark(chapter)
}

func (m *model) savePosition() {
	m.st.Book(m.book.Path).Position = store.Position{
		Chapter: m.chapter,
		Percent: int(m.viewport.ScrollPercent() * 100),
	}
	_ = m.st.Save()
}

func progressBar(percent float64, width int) string {
	filled := min(width, int(percent*float64(width)+0.5))
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}

func (m model) View() string {
	if !m.ready {
		return "carregando..."
	}

	var header, footer string
	if m.overlay != nil {
		header = fmt.Sprintf("%s · %s", m.overlay.title, m.book.Title)
		if m.overlay.pick != nil {
			footer = "j/k mover · enter abrir · esc fechar"
		} else {
			footer = "esc fechar"
		}
	} else {
		ch := m.book.Chapters[m.chapter]
		star := ""
		if m.hasBookmark(m.chapter) {
			star = " · ★"
		}
		header = fmt.Sprintf("%s · %s", m.book.Title, ch.Title)
		footer = fmt.Sprintf("capítulo %d/%d · %s %.0f%%%s · ? ajuda",
			m.chapter+1,
			len(m.book.Chapters),
			progressBar(m.viewport.ScrollPercent(), barWidth),
			m.viewport.ScrollPercent()*100,
			star,
		)
	}

	return strings.Join([]string{
		lipgloss.PlaceHorizontal(m.width, lipgloss.Center, chromeStyle.Render(header)),
		"",
		m.viewport.View(),
		"",
		lipgloss.PlaceHorizontal(m.width, lipgloss.Center, chromeStyle.Render(footer)),
	}, "\n")
}

func Run(book *epub.Book, cfg *store.Config, st *store.State) error {
	start := time.Now()

	p := tea.NewProgram(New(book, cfg, st), tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return err
	}

	if m, ok := final.(*model); ok {
		m.savePosition()
	}
	st.Book(book.Path).Seconds += int(time.Since(start).Seconds())
	return st.Save()
}
