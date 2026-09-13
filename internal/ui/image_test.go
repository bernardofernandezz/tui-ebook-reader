package ui

import (
	"image"
	"image/color"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/muesli/termenv"
)

func testImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 4, 2))
	for x := 0; x < 4; x++ {
		img.Set(x, 0, color.RGBA{R: 255, A: 255}) // linha de cima vermelha
		img.Set(x, 1, color.RGBA{B: 255, A: 255}) // linha de baixo azul
	}
	return img
}

func TestRenderImageTrueColor(t *testing.T) {
	out := renderImage(testImage(), termenv.TrueColor, 80)

	if lines := strings.Count(out, "\n"); lines != 1 {
		t.Fatalf("renderImage() gerou %d linhas, quer 1", lines)
	}
	if !strings.Contains(out, "\x1b[38;2;") || !strings.Contains(out, "\x1b[48;2;") {
		t.Fatal("faltou cor RGB de 24 bits")
	}

	plain := ansiPattern.ReplaceAllString(out, "")
	line, _, _ := strings.Cut(plain, "\n")
	if width := utf8.RuneCountInString(line); width != 6 {
		t.Fatalf("largura visível = %d, quer 6 (4 colunas + 2 de margem)", width)
	}
}

func TestRenderImageANSI256(t *testing.T) {
	out := renderImage(testImage(), termenv.ANSI256, 80)
	if !strings.Contains(out, "\x1b[38;5;") {
		t.Fatal("faltou cor de 256 cores")
	}
	if strings.Contains(out, "\x1b[38;2;") {
		t.Fatal("não deveria usar cor de 24 bits no perfil 256")
	}
}

func TestRenderImageAscii(t *testing.T) {
	out := renderImage(testImage(), termenv.Ascii, 80)
	if strings.Contains(out, "\x1b[38") || strings.Contains(out, "\x1b[48") {
		t.Fatal("perfil sem cor não deveria emitir código de cor")
	}
}

func TestRenderImageRespectsColumn(t *testing.T) {
	wide := image.NewRGBA(image.Rect(0, 0, 40, 2))
	out := renderImage(wide, termenv.TrueColor, 10)

	plain := ansiPattern.ReplaceAllString(out, "")
	line, _, _ := strings.Cut(plain, "\n")
	if width := utf8.RuneCountInString(line); width > 12 {
		t.Fatalf("largura visível = %d, quer no máximo 12 (10 + margem)", width)
	}
}

func TestRGBTo256(t *testing.T) {
	if got := rgbTo256(0, 0, 0); got != 16 {
		t.Errorf("preto = %d, quer 16", got)
	}
	if got := rgbTo256(255, 255, 255); got != 231 {
		t.Errorf("branco = %d, quer 231", got)
	}
}
