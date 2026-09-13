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
	return stripFrontmatter(b.reader.ReadContentMarkdownById(ch.ID))
}

// stripFrontmatter remove o bloco YAML inicial (--- ... ---) que alguns
// EPUBs colocam no começo de cada documento.
func stripFrontmatter(md string) string {
	if !strings.HasPrefix(md, "---") {
		return md
	}
	if i := strings.Index(md[3:], "---"); i >= 0 {
		return md[3+i+3:]
	}
	return md
}

func firstHeading(md string) string {
	md = stripFrontmatter(md)

	for _, line := range strings.Split(md, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "---") {
			continue
		}
		// heading de qualquer nível (# até ######)
		if h := strings.TrimLeft(line, "#"); len(h) < len(line) && strings.HasPrefix(h, " ") {
			if title := strings.Trim(h, "*— "); title != "" {
				return title
			}
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
