package epub

import (
	"path/filepath"
	"strings"
	"testing"

	raitu "github.com/raitucarp/epub"
)

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

func TestOpenFallsBackToSpine(t *testing.T) {
	w := raitu.New("urn:test:sem-toc")
	w.Title("Sem sumário")
	w.Languages("pt")
	body := strings.Repeat("palavra ", 120)

	w.AddContent("ch1.xhtml", []byte(xhtml("Capitulo Um", body)))
	w.AddContent("ch2.xhtml", []byte(xhtml("Capitulo Dois", body)))

	// um sumário com um único item não é suficiente; o spine decide
	toc := raitu.TOC{Title: "Sumário", Items: []raitu.TOC{{Title: "Um", Href: "ch1.xhtml"}}}
	if err := w.TableOfContents("toc", toc); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "sem-toc.epub")
	if err := w.Write(path); err != nil {
		t.Fatal(err)
	}

	book, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(book.Chapters) != 2 {
		t.Fatalf("capítulos = %d, quer 2", len(book.Chapters))
	}
	if got := book.Chapters[0].Title; got != "Capitulo Um" {
		t.Fatalf("primeiro título = %q, quer %q", got, "Capitulo Um")
	}
}

func TestOpenUsesTOCWhenAvailable(t *testing.T) {
	w := raitu.New("urn:test:com-toc")
	w.Title("Com sumário")
	w.Languages("pt")
	body := strings.Repeat("palavra ", 120)

	w.AddContent("ch1.xhtml", []byte(xhtml("Ignorado 1", body)))
	w.AddContent("ch2.xhtml", []byte(xhtml("Ignorado 2", body)))

	toc := raitu.TOC{
		Title: "Sumário",
		Items: []raitu.TOC{
			{Title: "Um", Href: "ch1.xhtml"},
			{Title: "Dois", Href: "ch2.xhtml"},
		},
	}
	if err := w.TableOfContents("toc", toc); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "com-toc.epub")
	if err := w.Write(path); err != nil {
		t.Fatal(err)
	}

	book, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(book.Chapters) != 2 {
		t.Fatalf("capítulos = %d, quer 2", len(book.Chapters))
	}
	if got := book.Chapters[0].Title; got != "Um" {
		t.Fatalf("primeiro título = %q, quer vir do sumário (%q)", got, "Um")
	}
}
