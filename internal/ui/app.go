package ui

import (
	"fmt"
	"image"
	"regexp"
	"strings"
	"time"

	"github.com/bernardofernandezz/tui-ebook-reader/internal/epub"
	"github.com/bernardofernandezz/tui-ebook-reader/internal/store"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

const (
	headerHeight  = 2
	footerHeight  = 2
	verticalSpace = headerHeight + footerHeight
	barWidth      = 12

	minReadWidth = 40
	maxReadWidth = 120
	widthStep    = 4
)

var (
	chromeStyle = lipgloss.NewStyle().Faint(true)
	ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

var helpLines = []string{
	"←/→ ou h/l   capítulo anterior / próximo",
	"c            lista de capítulos",
	"j/k ou ↑/↓   rolar uma linha",
	"d/u          meia página",
	"espaço       página para baixo",
	"pgup         página para cima",
	"/            buscar no capítulo",
	"n/N          próxima / anterior ocorrência",
	"b            marcar / desmarcar bookmark",
	"B            lista de bookmarks",
	"t            trocar o tema",
	"+/-          alargar / estreitar a coluna",
	"?            esta ajuda",
	"q            sair",
}

type model struct {
	book      *epub.Book
	cfg       *store.Config
	st        *store.State
	chapter   int
	viewport  viewport.Model
	renderer  *glamour.TermRenderer
	overlay   *overlay
	input     textinput.Model
	profile   termenv.Profile
	themeIdx  int
	cover     image.Image
	coverArt  string
	offset    int // rolagem salva ao abrir um overlay
	cacheCh   int
	cacheW    int
	cache     string
	resume    int // percentual a restaurar depois do primeiro layout
	searching bool
	query     string
	matches   []int
	matchIdx  int
	ready     bool
	width     int
	height    int
}

func New(book *epub.Book, cfg *store.Config, st *store.State) *model {
	m := &model{
		book:     book,
		cfg:      cfg,
		st:       st,
		profile:  lipgloss.ColorProfile(),
		themeIdx: themeIndex(cfg.Theme),
		cacheCh:  -1,
	}
	if cover := book.Cover(); cover != nil {
		m.cover = cover
		m.coverArt = renderImage(cover, m.profile, cfg.Width)
	}

	if pos := st.Book(book.Path).Position; pos.Chapter >= 0 && pos.Chapter < len(book.Chapters) {
		m.chapter = pos.Chapter
		m.resume = pos.Percent
	}

	m.input = textinput.New()
	m.input.Prompt = "/ "
	m.input.Placeholder = "buscar no capítulo"
	m.input.CharLimit = 64

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
		if m.searching {
			return m.updateSearch(msg)
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "right", "l":
			m.goToChapter(m.chapter + 1)
		case "left", "h", "p":
			m.goToChapter(m.chapter - 1)
		case "n":
			if m.query != "" {
				m.jumpMatch(1)
			} else {
				m.goToChapter(m.chapter + 1)
			}
		case "N":
			m.jumpMatch(-1)
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
		case "t":
			m.cycleTheme()
			return m, nil
		case "+", "=":
			m.adjustWidth(widthStep)
			return m, nil
		case "-", "_":
			m.adjustWidth(-widthStep)
			return m, nil
		case "?":
			m.openHelp()
			return m, nil
		case "/":
			m.searching = true
			m.input.SetValue("")
			return m, m.input.Focus()
		case "esc":
			m.clearSearch()
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
			if renderer, err := newRenderer(themes[m.themeIdx].glamour, m.width, m.cfg.Width); err == nil {
				m.renderer = renderer
			}
			if m.overlay != nil {
				m.setOverlayContent()
			} else {
				m.setContent()
			}
			if m.query != "" {
				m.clearSearch()
			}
			if m.resume > 0 {
				m.scrollToPercent(m.resume)
				m.resume = 0
			}
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func newRenderer(style []byte, termWidth, maxWidth int) (*glamour.TermRenderer, error) {
	return glamour.NewTermRenderer(
		glamour.WithStylesFromJSONBytes(style),
		glamour.WithWordWrap(wrapWidth(termWidth, maxWidth)),
	)
}

// wrapWidth limita a largura do texto para uma leitura confortável.
func wrapWidth(termWidth, maxWidth int) int {
	if maxWidth <= 0 {
		maxWidth = store.DefaultWidth
	}
	return min(max(20, termWidth-4), maxWidth)
}

// applyRenderer recria o renderer (tema/largura) e redesenha o conteúdo.
func (m *model) applyRenderer() {
	if renderer, err := newRenderer(themes[m.themeIdx].glamour, m.width, m.cfg.Width); err == nil {
		m.renderer = renderer
	}
	m.cacheCh = -1
	if m.cover != nil {
		m.coverArt = renderImage(m.cover, m.profile, m.cfg.Width)
	}

	if m.overlay != nil {
		m.setOverlayContent()
	} else {
		m.setContent()
	}
}

func (m *model) cycleTheme() {
	m.themeIdx = (m.themeIdx + 1) % len(themes)
	m.cfg.Theme = themes[m.themeIdx].name
	_ = m.cfg.Save()
	m.applyRenderer()
}

func (m *model) adjustWidth(delta int) {
	m.cfg.Width = min(maxReadWidth, max(minReadWidth, m.cfg.Width+delta))
	_ = m.cfg.Save()
	m.applyRenderer()
}

func (m *model) goToChapter(chapter int) {
	if chapter < 0 || chapter >= len(m.book.Chapters) {
		return
	}
	m.clearSearch()
	m.chapter = chapter
	m.setContent()
	m.savePosition()
}

var (
	imageRefPattern = regexp.MustCompile(`!\[[^\]]*\]\(([^)\s]+)[^)]*\)`)
	linkPattern     = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
)

func (m *model) setContent() {
	// centraliza a coluna de leitura no terminal
	centered := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(m.chapterText())

	m.viewport.SetContent(centered)
	m.viewport.GotoTop()
}

// chapterText é o conteúdo do capítulo atual, com a capa na primeira página.
func (m *model) chapterText() string {
	text := m.chapterContent()
	if m.chapter == 0 && m.coverArt != "" {
		text = m.coverArt + "\n" + text
	}
	return text
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
			out.WriteString(renderImage(img, m.profile, m.cfg.Width))
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

	// no leitor o texto do link basta; a URL só atrapalha
	md = linkPattern.ReplaceAllString(md, "$1")

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
	m.viewport.SetContent(m.overlay.view(m.width, m.cfg.Width))

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

func (m *model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.stopSearch()
		return m, nil
	case "enter":
		m.startSearch()
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// startSearch procura a consulta nas linhas renderizadas do capítulo atual.
func (m *model) startSearch() {
	m.searching = false
	m.input.Blur()

	m.query = strings.TrimSpace(m.input.Value())
	m.matches = nil
	m.matchIdx = 0
	if m.query == "" {
		return
	}

	needle := strings.ToLower(m.query)
	for i, line := range strings.Split(m.chapterText(), "\n") {
		plain := strings.ToLower(ansiPattern.ReplaceAllString(line, ""))
		if strings.Contains(plain, needle) {
			m.matches = append(m.matches, i)
		}
	}
	m.jumpMatch(0)
}

func (m *model) stopSearch() {
	m.searching = false
	m.input.Blur()
}

func (m *model) clearSearch() {
	m.stopSearch()
	m.query = ""
	m.matches = nil
	m.matchIdx = 0
}

// jumpMatch anda para a próxima ocorrência (delta 1) ou anterior (delta -1).
func (m *model) jumpMatch(delta int) {
	if len(m.matches) == 0 {
		return
	}
	m.matchIdx = (m.matchIdx + delta + len(m.matches)) % len(m.matches)
	m.viewport.SetYOffset(max(0, m.matches[m.matchIdx]-m.viewport.Height/2))
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
	switch {
	case m.overlay != nil:
		header = fmt.Sprintf("%s · %s", m.overlay.title, m.book.Title)
		if m.overlay.pick != nil {
			footer = "j/k mover · enter abrir · esc fechar"
		} else {
			footer = "esc fechar"
		}
	case m.searching:
		header = fmt.Sprintf("%s · %s", m.book.Title, m.book.Chapters[m.chapter].Title)
		footer = m.input.View() + "   enter busca · esc cancela"
	case m.query != "":
		header = fmt.Sprintf("%s · %s", m.book.Title, m.book.Chapters[m.chapter].Title)
		if len(m.matches) == 0 {
			footer = fmt.Sprintf("busca %q · nenhuma ocorrência · esc limpa", m.query)
		} else {
			footer = fmt.Sprintf("busca %q · %d/%d · n/N pular · esc limpa",
				m.query, m.matchIdx+1, len(m.matches))
		}
	default:
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
