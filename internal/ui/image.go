package ui

import (
	"fmt"
	"image"
	"strings"

	"github.com/muesli/termenv"
	"golang.org/x/image/draw"
)

const (
	maxImageWidth  = 48 // colunas
	maxImageHeight = 24 // linhas de caractere (2 pixels cada)
)

// renderImage converte a imagem em blocos ANSI: cada caractere "▀" mostra
// dois pixels, o de cima na cor do texto e o de baixo na cor de fundo.
func renderImage(img image.Image, profile termenv.Profile, maxWidth int) string {
	bounds := img.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()

	width := min(maxImageWidth, max(10, maxWidth))
	scale := min(
		float64(width)/float64(srcW),
		float64(maxImageHeight*2)/float64(srcH),
		1,
	)
	w := max(1, int(float64(srcW)*scale))
	h := max(2, int(float64(srcH)*scale))
	if h%2 != 0 {
		h++
	}

	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	var sb strings.Builder
	for y := 0; y < h; y += 2 {
		sb.WriteString("  ") // mesma margem que o glamour usa no texto
		for x := 0; x < w; x++ {
			r1, g1, b1, _ := dst.At(x, y).RGBA()
			r2, g2, b2, _ := dst.At(x, y+1).RGBA()
			sb.WriteString(colorCode(uint8(r1>>8), uint8(g1>>8), uint8(b1>>8), false, profile))
			sb.WriteString(colorCode(uint8(r2>>8), uint8(g2>>8), uint8(b2>>8), true, profile))
			sb.WriteString("▀")
		}
		sb.WriteString("\x1b[0m\n")
	}
	return sb.String()
}

// colorCode devolve a sequência ANSI da cor no perfil do terminal.
func colorCode(r, g, b uint8, background bool, profile termenv.Profile) string {
	layer := "38" // frente
	if background {
		layer = "48" // fundo
	}

	switch profile {
	case termenv.Ascii:
		return ""
	case termenv.TrueColor:
		return fmt.Sprintf("\x1b[%s;2;%d;%d;%dm", layer, r, g, b)
	default:
		return fmt.Sprintf("\x1b[%s;5;%dm", layer, rgbTo256(r, g, b))
	}
}

// rgbTo256 aproxima a cor para o cubo 6x6x6 do xterm.
func rgbTo256(r, g, b uint8) int {
	toCube := func(v uint8) int { return (int(v)*5 + 127) / 255 }
	return 16 + 36*toCube(r) + 6*toCube(g) + toCube(b)
}
