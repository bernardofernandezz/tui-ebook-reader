package ui

import (
	"embed"
	"io/fs"
	"strings"
)

//go:embed themes/*.json
var themesFS embed.FS

// theme é um estilo do glamour embutido no binário.
type theme struct {
	name    string
	glamour []byte
}

// themes fica na ordem alfabética dos arquivos (dark, light, sepia).
var themes = loadThemes()

func loadThemes() []theme {
	entries, err := fs.ReadDir(themesFS, "themes")
	if err != nil {
		return nil
	}

	var out []theme
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := themesFS.ReadFile("themes/" + entry.Name())
		if err != nil {
			continue
		}
		out = append(out, theme{
			name:    strings.TrimSuffix(entry.Name(), ".json"),
			glamour: data,
		})
	}
	return out
}

// themeIndex devolve a posição do tema pelo nome; sem correspondência, usa o primeiro.
func themeIndex(name string) int {
	for i, t := range themes {
		if t.name == name {
			return i
		}
	}
	return 0
}
