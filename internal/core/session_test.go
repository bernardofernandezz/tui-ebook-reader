package core

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	raitu "github.com/raitucarp/epub"

	"github.com/bernardofernandezz/tui-ebook-reader/internal/epub"
	"github.com/bernardofernandezz/tui-ebook-reader/internal/store"
)

// fakeBook monta um EPUB de teste com os títulos informados.
func fakeBook(t *testing.T, titles ...string) *epub.Book {
	t.Helper()

	w := raitu.New("urn:test:core")
	w.Title("Livro de teste")
	w.Languages("pt")

	body := strings.Repeat("palavra ", 120)
	items := make([]raitu.TOC, len(titles))
	for i, title := range titles {
		name := fmt.Sprintf("ch%d.xhtml", i+1)
		w.AddContent(name, []byte(xhtml(title, body)))
		items[i] = raitu.TOC{Title: title, Href: name}
	}
	if err := w.TableOfContents("toc", raitu.TOC{Title: "Sumário", Items: items}); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "livro.epub")
	if err := w.Write(path); err != nil {
		t.Fatal(err)
	}

	book, err := epub.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	return book
}

func xhtml(title, body string) string {
	return `<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<body>
<h1>` + title + `</h1>
<p>` + body + `</p>
</body>
</html>`
}

func TestSessionRestoresSavedPosition(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	book := fakeBook(t, "Um", "Dois", "Três")
	st := store.LoadState()
	st.Book(book.Path).Position = store.Position{Chapter: 2, Percent: 40}

	s := New(book, st)
	if s.Chapter() != 2 {
		t.Fatalf("capítulo = %d, quer 2", s.Chapter())
	}
	if pos := s.Position(); pos.Percent != 40 {
		t.Fatalf("percent = %d, quer 40", pos.Percent)
	}
}

func TestSessionIgnoresInvalidSavedChapter(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	book := fakeBook(t, "Um", "Dois")
	st := store.LoadState()
	st.Book(book.Path).Position = store.Position{Chapter: 99, Percent: 10}

	if s := New(book, st); s.Chapter() != 0 {
		t.Fatalf("capítulo = %d, quer 0", s.Chapter())
	}
}

func TestGoToRespectsLimits(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	s := New(fakeBook(t, "Um", "Dois"), store.LoadState())

	if !s.GoTo(1) || s.Chapter() != 1 {
		t.Fatal("GoTo(1) devia mover para o capítulo 1")
	}
	for _, invalid := range []int{-1, 2, 99} {
		if s.GoTo(invalid) || s.Chapter() != 1 {
			t.Fatalf("GoTo(%d) devia ser recusado", invalid)
		}
	}
}

func TestUpdateProgressPersistsOnSave(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	book := fakeBook(t, "Um", "Dois")
	s := New(book, store.LoadState())

	if !s.GoTo(1) {
		t.Fatal("GoTo falhou")
	}
	s.UpdateProgress(35)
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	pos := store.LoadState().Book(book.Path).Position
	if pos.Chapter != 1 || pos.Percent != 35 {
		t.Fatalf("posição persistida = %+v, quer capítulo 1 em 35%%", pos)
	}
}

func TestBookmarkTogglePersists(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	book := fakeBook(t, "Um", "Dois")
	s := New(book, store.LoadState())

	s.ToggleBookmark(10, "Um")
	if !s.HasBookmark(0) {
		t.Fatal("bookmark devia existir no capítulo 0")
	}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	bookmarks := store.LoadState().Book(book.Path).SortedBookmarks()
	if len(bookmarks) != 1 || bookmarks[0].Label != "Um" {
		t.Fatalf("bookmarks persistidos = %+v", bookmarks)
	}

	s.ToggleBookmark(30, "Um")
	if s.HasBookmark(0) {
		t.Fatal("segundo toggle devia remover o bookmark")
	}
	if len(s.Bookmarks()) != 0 {
		t.Fatalf("bookmarks = %+v, quer vazio", s.Bookmarks())
	}
}

func TestCloseAccumulatesReadingTime(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	book := fakeBook(t, "Um", "Dois")
	s := New(book, store.LoadState())
	s.startedAt = s.startedAt.Add(-90 * time.Second)

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	if seconds := store.LoadState().Book(book.Path).Seconds; seconds < 90 {
		t.Fatalf("segundos = %d, quer pelo menos 90", seconds)
	}
}
