package epub

import (
	"fmt"
	"strings"

	raitu "github.com/raitucarp/epub"
)

type Book struct {
	Title    string
	Author   string
	Chapters []Chapter
	reader   raitu.Reader
}

type Chapter struct {
	ID    string
	Title string
	Index int
}

func Open(path string) (*Book, error) {
	r, err := raitu.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("abrir epub: %w", err)
	}

	b := &Book{reader: r}

	if t := r.Title(); len(t) > 0 {
		b.Title = t[0]
	}
	if a := r.Author(); len(a) > 0 {
		b.Author = a[0]
	}

	ids := r.ListContentDocumentIds()
	b.Chapters = make([]Chapter, 0, len(ids))

	for i, id := range ids {
		md := r.ReadContentMarkdownById(id)

		// pula documentos muito curtos (capa, copyright, avisos)
		if len(strings.TrimSpace(md)) < 400 {
			continue
		}

		title := firstHeading(md)
		if title == "" {
			title = fmt.Sprintf("Seção %d", i+1)
		}

		b.Chapters = append(b.Chapters, Chapter{
			ID:    id,
			Title: title,
			Index: i,
		})
	}

	if len(b.Chapters) == 0 {
		return nil, fmt.Errorf("nenhum capítulo legível encontrado")
	}

	return b, nil
}

func (b *Book) Text(ch Chapter) string {
	return b.reader.ReadContentMarkdownById(ch.ID)
}

func firstHeading(md string) string {
	for _, line := range strings.Split(md, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "---") {
			continue
		}
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(line[2:])
		}
		if strings.HasPrefix(line, "## ") {
			return strings.TrimSpace(line[3:])
		}
		// fallback: primeira linha útil
		if len(line) > 3 {
			if len(line) > 60 {
				return line[:60] + "..."
			}
			return line
		}
	}
	return ""
}
