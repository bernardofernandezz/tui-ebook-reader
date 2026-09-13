package epub

import (
	"fmt"
	"image"
	"path"
	"path/filepath"
	"strings"

	raitu "github.com/raitucarp/epub"
	"github.com/raitucarp/epub/ncx"
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

	b.Chapters = b.loadChapters()
	if len(b.Chapters) == 0 {
		return nil, fmt.Errorf("nenhum capítulo legível encontrado")
	}

	return b, nil
}

// Cover devolve a imagem de capa do livro, se houver.
func (b *Book) Cover() image.Image {
	if cover := b.reader.Cover(); cover != nil {
		return *cover
	}
	return nil
}

// loadChapters prefere o sumário do próprio EPUB (NCX); sem ele, cai na
// heurística sobre os documentos do spine.
func (b *Book) loadChapters() []Chapter {
	if chapters := b.chaptersFromTOC(); len(chapters) >= 2 {
		return chapters
	}
	return b.chaptersFromSpine()
}

func (b *Book) chaptersFromTOC() []Chapter {
	doc := b.reader.NavigationCenterExtended()
	if doc == nil {
		return nil
	}

	byHref := map[string]string{}
	byBase := map[string]string{}
	for _, res := range b.reader.Resources() {
		href := path.Clean(res.Href)
		byHref[href] = res.ID
		base := strings.TrimSuffix(path.Base(href), path.Ext(href))
		if _, ok := byBase[base]; !ok {
			byBase[base] = res.ID
		}
	}

	var chapters []Chapter
	seen := map[string]bool{}

	var walk func(points []ncx.NavPoint)
	walk = func(points []ncx.NavPoint) {
		for _, point := range points {
			if ch, ok := chapterFromNavPoint(point, byHref, byBase, seen); ok {
				seen[ch.ID] = true
				chapters = append(chapters, ch)
			}
			walk(point.NavPoints)
		}
	}
	walk(doc.NavMap.NavPoints)

	return chapters
}

func chapterFromNavPoint(point ncx.NavPoint, byHref, byBase map[string]string, seen map[string]bool) (Chapter, bool) {
	title := strings.Trim(strings.TrimSpace(point.NavLabel.Text), "— ")
	if title == "" {
		return Chapter{}, false
	}

	src := strings.SplitN(point.Content.Src, "#", 2)[0]
	if src == "" {
		return Chapter{}, false
	}

	id := byHref[path.Clean(src)]
	if id == "" {
		id = byBase[strings.TrimSuffix(path.Base(src), path.Ext(src))]
	}
	if id == "" || seen[id] {
		return Chapter{}, false
	}

	return Chapter{ID: id, Title: title, Index: len(seen)}, true
}

func (b *Book) chaptersFromSpine() []Chapter {
	ids := b.reader.ListContentDocumentIds()
	chapters := make([]Chapter, 0, len(ids))

	for i, id := range ids {
		md := b.reader.ReadContentMarkdownById(id)

		// pula documentos muito curtos (capa, copyright, avisos)
		if len(strings.TrimSpace(md)) < 400 {
			continue
		}

		title := firstHeading(md)
		if title == "" {
			title = fmt.Sprintf("Seção %d", i+1)
		}

		chapters = append(chapters, Chapter{
			ID:    id,
			Title: title,
			Index: i,
		})
	}
	return chapters
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
