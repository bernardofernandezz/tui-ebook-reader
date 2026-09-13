// Package bookmarks guarda as marcas de leitura por livro em um JSON.
package bookmarks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// Bookmark aponta para um ponto do livro: capítulo e percentual rolado.
type Bookmark struct {
	Chapter int    `json:"chapter"`
	Percent int    `json:"percent"`
	Label   string `json:"label"`
}

// Store mantém os bookmarks de todos os livros, indexados pelo caminho.
type Store struct {
	path  string
	Books map[string][]Bookmark `json:"books"`
}

// Load lê o arquivo do usuário; se não existir, começa vazio.
func Load() *Store {
	s := &Store{Books: map[string][]Bookmark{}}

	dir, err := os.UserConfigDir()
	if err != nil {
		return s
	}
	s.path = filepath.Join(dir, "tbook", "bookmarks.json")

	data, err := os.ReadFile(s.path)
	if err != nil {
		return s
	}
	if err := json.Unmarshal(data, s); err != nil || s.Books == nil {
		s.Books = map[string][]Bookmark{}
	}
	return s
}

// List devolve os bookmarks do livro em ordem de leitura.
func (s *Store) List(bookPath string) []Bookmark {
	marks := append([]Bookmark(nil), s.Books[bookPath]...)
	sort.Slice(marks, func(i, j int) bool {
		if marks[i].Chapter != marks[j].Chapter {
			return marks[i].Chapter < marks[j].Chapter
		}
		return marks[i].Percent < marks[j].Percent
	})
	return marks
}

// Toggle adiciona o bookmark do capítulo ou remove o existente.
func (s *Store) Toggle(bookPath string, bm Bookmark) {
	for i, b := range s.Books[bookPath] {
		if b.Chapter == bm.Chapter {
			s.Books[bookPath] = append(s.Books[bookPath][:i], s.Books[bookPath][i+1:]...)
			s.save()
			return
		}
	}
	s.Books[bookPath] = append(s.Books[bookPath], bm)
	s.save()
}

func (s *Store) save() {
	if s.path == "" {
		return
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return
	}
	_ = os.WriteFile(s.path, data, 0o644)
}
