package epub

import (
	"testing"

	"github.com/raitucarp/epub/ncx"
)

func TestStripFrontmatter(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"sem frontmatter", "# Título\n\ntexto", "# Título\n\ntexto"},
		{"com frontmatter", "---\ntitle: \"x\"\n---\n# Título\n", "\n# Título\n"},
		{"frontmatter sem fim", "---\ntitle: \"x\"", "---\ntitle: \"x\""},
	}
	for _, tc := range cases {
		if got := stripFrontmatter(tc.in); got != tc.want {
			t.Errorf("%s: stripFrontmatter() = %q, quer %q", tc.name, got, tc.want)
		}
	}
}

func TestFirstHeading(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"h1", "# Título\n", "Título"},
		{"h4 com travessões", "#### — CAPÍTULO UM —\n", "CAPÍTULO UM"},
		{"h4 em itálico", "#### *O menino que sobreviveu*\n", "O menino que sobreviveu"},
		{"frontmatter ignorado", "---\ntitle: \"livro\"\n---\n# Capítulo\n", "Capítulo"},
		{"fallback de texto", "Primeira linha útil\n", "Primeira linha útil"},
		{"imagem ignorada", "![](../Images/logo.jpg)\n", ""},
	}
	for _, tc := range cases {
		if got := firstHeading(tc.in); got != tc.want {
			t.Errorf("%s: firstHeading() = %q, quer %q", tc.name, got, tc.want)
		}
	}
}

func TestChapterFromNavPoint(t *testing.T) {
	byHref := map[string]string{"Text/08_c1.xhtml": "body008"}
	byBase := map[string]string{"09_c2": "body009"}

	point := ncx.NavPoint{
		NavLabel: ncx.NavLabel{Text: "— CAPÍTULO UM —"},
		Content:  ncx.Content{Src: "Text/08_c1.xhtml#top"},
	}
	ch, ok := chapterFromNavPoint(point, byHref, byBase, map[string]bool{})
	if !ok || ch.ID != "body008" || ch.Title != "CAPÍTULO UM" {
		t.Fatalf("href exato: %+v, ok=%v", ch, ok)
	}

	// resolve pelo nome do arquivo quando a extensão/caminho divergem
	point.Content.Src = "09_c2.xhtml"
	if ch, ok := chapterFromNavPoint(point, byHref, byBase, map[string]bool{}); !ok || ch.ID != "body009" {
		t.Fatalf("fallback por nome: %+v, ok=%v", ch, ok)
	}

	// documento repetido ou sem título é ignorado
	if _, ok := chapterFromNavPoint(point, byHref, byBase, map[string]bool{"body009": true}); ok {
		t.Fatal("documento repetido deveria ser ignorado")
	}
	point.NavLabel.Text = "  "
	if _, ok := chapterFromNavPoint(point, byHref, byBase, map[string]bool{}); ok {
		t.Fatal("navpoint sem título deveria ser ignorado")
	}
}
