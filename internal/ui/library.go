package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bernardofernandezz/tui-ebook-reader/internal/epub"
	"github.com/bernardofernandezz/tui-ebook-reader/internal/store"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// libraryModel é a lista de livros: à esquerda o acervo, à direita a capa e o
// progresso do livro destacado. Cada EPUB é aberto só quando o cursor chega
// nele, e a capa renderizada fica em cache.
type libraryModel struct {
	paths   []string
	cursor  int
	offset  int
	chosen  *epub.Book
	books   map[string]*epub.Book // abertos sob demanda
	covers  map[string]string     // capas já renderizadas
	openErr string
	st      *store.State
	ready   bool
	width   int
	height  int
}

// RunLibrary abre a biblioteca e devolve o livro escolhido, ou nil se o
// usuário saiu sem escolher.
func RunLibrary(paths []string, st *store.State) (*epub.Book, error) {
	m := &libraryModel{
		paths:  paths,
		st:     st,
		books:  map[string]*epub.Book{},
		covers: map[string]string{},
	}
	if len(paths) > 0 {
		m.ensureBook(paths[0])
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return nil, err
	}
	return m.chosen, nil
}

func (m *libraryModel) Init() tea.Cmd { return nil }

func (m *libraryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.ready = true
		m.scroll()
		m.ensureCover()
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "j", "down":
			m.move(1)
		case "k", "up":
			m.move(-1)
		case "enter":
			if book := m.ensureBook(m.paths[m.cursor]); book != nil {
				m.chosen = book
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m *libraryModel) move(delta int) {
	m.cursor = min(max(m.cursor+delta, 0), len(m.paths)-1)
	m.ensureBook(m.paths[m.cursor])
	m.ensureCover()
	m.scroll()
}

// scroll mantém o cursor dentro da janela visível (2 linhas por livro).
func (m *libraryModel) scroll() {
	rows := m.rows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+rows {
		m.offset = m.cursor - rows + 1
	}
	if maxOffset := len(m.paths) - rows; m.offset > maxOffset {
		m.offset = max(0, maxOffset)
	}
}

func (m *libraryModel) rows() int {
	return max(1, (m.height-4)/2)
}

func (m *libraryModel) listWidth() int {
	return min(40, max(24, m.width/2))
}

func (m *libraryModel) ensureBook(path string) *epub.Book {
	if book, ok := m.books[path]; ok {
		return book
	}
	book, err := epub.Open(path)
	if err != nil {
		m.openErr = err.Error()
		return nil
	}
	m.openErr = ""
	m.books[path] = book
	return book
}

func (m *libraryModel) ensureCover() {
	if !m.ready || len(m.paths) == 0 {
		return
	}
	path := m.paths[m.cursor]
	if _, ok := m.covers[path]; ok {
		return
	}
	book := m.books[path]
	if book == nil {
		return
	}
	cover := book.Cover()
	if cover == nil {
		m.covers[path] = ""
		return
	}
	m.covers[path] = renderImage(cover, lipgloss.ColorProfile(), m.width-m.listWidth()-4)
}

// percent lê o progresso direto do state, sem criar entradas novas.
func (m *libraryModel) percent(path string) int {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	if b, ok := m.st.Books[abs]; ok {
		return b.Position.Percent
	}
	return 0
}

func (m *libraryModel) View() string {
	if !m.ready {
		return "carregando..."
	}

	width := m.listWidth()
	end := min(len(m.paths), m.offset+m.rows())

	var list strings.Builder
	for i := m.offset; i < end; i++ {
		path := m.paths[i]
		book := m.books[path]

		title := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		author := ""
		if book != nil {
			if book.Title != "" {
				title = book.Title
			}
			author = book.Author
		}

		marker := "  "
		style := lipgloss.NewStyle().Width(width - 2).MaxWidth(width - 2)
		if i == m.cursor {
			marker = "▸ "
			style = style.Bold(true)
		}
		fmt.Fprintf(&list, "%s%s\n", marker, style.Render(title))

		sub := fmt.Sprintf("%d%%", m.percent(path))
		if author != "" {
			sub = author + " · " + sub
		}
		fmt.Fprintf(&list, "  %s\n", chromeStyle.Render(style.Render(sub)))
	}

	footer := "j/k mover · enter abrir · q sair"
	if m.openErr != "" {
		footer = m.openErr + " · " + footer
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, list.String(), "  ", m.preview()) +
		"\n" + chromeStyle.Render(footer)
}

func (m *libraryModel) preview() string {
	if len(m.paths) == 0 {
		return ""
	}
	path := m.paths[m.cursor]
	book := m.books[path]
	if book == nil {
		return chromeStyle.Render("não foi possível abrir este livro")
	}

	var sb strings.Builder
	if art, ok := m.covers[path]; ok && art != "" {
		sb.WriteString(art)
		sb.WriteString("\n")
	}
	sb.WriteString(book.Title + "\n")
	if book.Author != "" {
		sb.WriteString(book.Author + "\n")
	}
	percent := m.percent(path)
	fmt.Fprintf(&sb, "%s %d%%\n", progressBar(float64(percent)/100, barWidth), percent)
	return sb.String()
}
