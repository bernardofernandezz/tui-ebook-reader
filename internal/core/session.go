// Package core guarda a sessão de leitura: livro, capítulo, posição e
// bookmarks. Nada aqui conhece terminal.
package core

import (
	"time"

	"github.com/bernardofernandezz/tui-ebook-reader/internal/epub"
	"github.com/bernardofernandezz/tui-ebook-reader/internal/store"
)

type Session struct {
	book      *epub.Book
	state     *store.State
	chapter   int
	startedAt time.Time
}

// New retoma a posição salva quando o capítulo ainda existe.
func New(book *epub.Book, state *store.State) *Session {
	s := &Session{book: book, state: state, startedAt: time.Now()}
	if pos := state.Book(book.Path).Position; pos.Chapter >= 0 && pos.Chapter < len(book.Chapters) {
		s.chapter = pos.Chapter
	}
	return s
}

func (s *Session) Book() *epub.Book { return s.book }

func (s *Session) Chapter() int { return s.chapter }

// GoTo devolve false se o capítulo não existir.
func (s *Session) GoTo(chapter int) bool {
	if chapter < 0 || chapter >= len(s.book.Chapters) {
		return false
	}
	s.chapter = chapter
	return true
}

// Position devolve o último ponto gravado para o livro.
func (s *Session) Position() store.Position {
	return s.state.Book(s.book.Path).Position
}

func (s *Session) UpdateProgress(percent int) {
	s.state.Book(s.book.Path).Position = store.Position{
		Chapter: s.chapter,
		Percent: percent,
	}
}

func (s *Session) ToggleBookmark(percent int, label string) {
	s.state.Book(s.book.Path).ToggleBookmark(store.Bookmark{
		Chapter: s.chapter,
		Percent: percent,
		Label:   label,
	})
}

func (s *Session) HasBookmark(chapter int) bool {
	return s.state.Book(s.book.Path).HasBookmark(chapter)
}

// Bookmarks devolve em ordem de leitura.
func (s *Session) Bookmarks() []store.Bookmark {
	return s.state.Book(s.book.Path).SortedBookmarks()
}

func (s *Session) Save() error {
	return s.state.Save()
}

// Close soma o tempo de leitura desde a abertura e grava o estado.
func (s *Session) Close() error {
	s.state.Book(s.book.Path).Seconds += int(time.Since(s.startedAt).Seconds())
	return s.state.Save()
}
