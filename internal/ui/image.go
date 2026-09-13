package ui

import (
	"fmt"
	"image"
	"strings"

	"golang.org/x/image/draw"
)

const (
	maxImageWidth  = 48 // colunas
	maxImageHeight = 24 // linhas de caractere (2 pixels cada)
)

// renderImage converte a imagem em blocos ANSI: cada caractere "▀" mostra
// dois pixels, o de cima na cor do texto e o de baixo na cor de fundo.
func renderImage(img image.Image) string {
	bounds := img.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()

	scale := min(
		float64(maxImageWidth)/float64(srcW),
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
			fmt.Fprintf(&sb,
				"\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀",
				r1>>8, g1>>8, b1>>8,
				r2>>8, g2>>8, b2>>8,
			)
		}
		sb.WriteString("\x1b[0m\n")
	}
	return sb.String()
}
