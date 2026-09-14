package ui

import (
	"path/filepath"
	"testing"

	"github.com/bernardofernandezz/tui-ebook-reader/internal/epub"
	"github.com/bernardofernandezz/tui-ebook-reader/internal/store"
)

func TestLibraryMoveAndScroll(t *testing.T) {
	m := &libraryModel{
		paths:  []string{"a.epub", "b.epub", "c.epub"},
		books:  map[string]*epub.Book{},
		covers: map[string]string{},
		st:     &store.State{Books: map[string]*store.BookState{}},
		height: 8, // rows() = 2
	}

	m.move(1)
	if m.cursor != 1 || m.offset != 0 {
		t.Fatalf("cursor %d, offset %d; quer 1, 0", m.cursor, m.offset)
	}
	m.move(1)
	if m.cursor != 2 || m.offset != 1 {
		t.Fatalf("cursor %d, offset %d; quer 2, 1", m.cursor, m.offset)
	}
	m.move(5)
	if m.cursor != 2 || m.offset != 1 {
		t.Fatalf("clamp fim: cursor %d, offset %d; quer 2, 1", m.cursor, m.offset)
	}
	m.move(-5)
	if m.cursor != 0 || m.offset != 0 {
		t.Fatalf("clamp início: cursor %d, offset %d; quer 0, 0", m.cursor, m.offset)
	}
}

func TestLibraryPercent(t *testing.T) {
	abs, err := filepath.Abs("a.epub")
	if err != nil {
		t.Fatal(err)
	}
	m := &libraryModel{
		st: &store.State{Books: map[string]*store.BookState{
			abs: {Position: store.Position{Percent: 42}},
		}},
	}

	if got := m.percent("a.epub"); got != 42 {
		t.Fatalf("percent = %d, quer 42", got)
	}
	if got := m.percent("nao-existe.epub"); got != 0 {
		t.Fatalf("percent = %d, quer 0", got)
	}
}
