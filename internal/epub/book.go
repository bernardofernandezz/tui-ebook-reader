package epub

import (
	"fmt"
	"image"
	"path"
	"path/filepath"
	"strings"

	raitu "github.com/raitucarp/epub"
)

type Book struct {
	Title    string
	Author   string
	Chapters []Chapter
	Path     string
	reader   raitu.Reader
	images   map[string]image.Image
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

	b := &Book{
		reader: r,
		images: map[string]image.Image{},
	}
	if abs, err := filepath.Abs(path); err == nil {
		b.Path = abs
	} else {
		b.Path = path
	}

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

// Image carrega a imagem referenciada no markdown do capítulo (ex.:
// ../Images/x.jpg), resolvendo o caminho relativo ao documento.
func (b *Book) Image(ch Chapter, ref string) (image.Image, error) {
	href := b.resolveImageHref(ch, ref)
	if img, ok := b.images[href]; ok {
		return img, nil
	}
	for _, try := range []string{href, ref} {
		if img := b.reader.ReadImageByHref(try); img != nil {
			b.images[href] = *img
			return *img, nil
		}
	}
	return nil, fmt.Errorf("imagem não encontrada: %s", ref)
}

func (b *Book) resolveImageHref(ch Chapter, ref string) string {
	for _, res := range b.reader.Resources() {
		if res.ID == ch.ID {
			return path.Join(path.Dir(res.Href), ref)
		}
	}
	return ref
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
		// fallback: primeira linha útil (ignora referências de imagem)
		if len(line) > 3 && !strings.HasPrefix(line, "!") {
			if len(line) > 60 {
				return line[:60] + "..."
			}
			return line
		}
	}
	return ""
}
