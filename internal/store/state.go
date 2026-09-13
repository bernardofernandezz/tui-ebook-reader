package store

import "sort"

// Position é o ponto onde a leitura parou.
type Position struct {
	Chapter int `json:"chapter"`
	Percent int `json:"percent"`
}

// Bookmark marca um ponto de um capítulo.
type Bookmark struct {
	Chapter int    `json:"chapter"`
	Percent int    `json:"percent"`
	Label   string `json:"label"`
}

// BookState são os dados de leitura de um livro.
type BookState struct {
	Position  Position   `json:"position"`
	Bookmarks []Bookmark `json:"bookmarks,omitempty"`
	Seconds   int        `json:"seconds"`
}

// State é o estado de leitura de todos os livros, indexado pelo caminho.
type State struct {
	Books map[string]*BookState `json:"books"`
}

// LoadState lê o estado do disco; se não existir, começa vazio.
func LoadState() *State {
	s := &State{}
	_ = readJSON(path(stateFile), s)
	if s.Books == nil {
		s.Books = map[string]*BookState{}
	}
	return s
}

// Save grava o estado no disco.
func (s *State) Save() error {
	return writeJSON(path(stateFile), s)
}

// Book devolve o estado do livro, criando a entrada se ainda não existir.
func (s *State) Book(bookPath string) *BookState {
	b, ok := s.Books[bookPath]
	if !ok {
		b = &BookState{}
		s.Books[bookPath] = b
	}
	return b
}

// ToggleBookmark adiciona o bookmark do capítulo ou remove o existente.
func (b *BookState) ToggleBookmark(bm Bookmark) {
	for i, cur := range b.Bookmarks {
		if cur.Chapter == bm.Chapter {
			b.Bookmarks = append(b.Bookmarks[:i], b.Bookmarks[i+1:]...)
			return
		}
	}
	b.Bookmarks = append(b.Bookmarks, bm)
}

// HasBookmark informa se o capítulo tem bookmark.
func (b *BookState) HasBookmark(chapter int) bool {
	for _, bm := range b.Bookmarks {
		if bm.Chapter == chapter {
			return true
		}
	}
	return false
}

// SortedBookmarks devolve os bookmarks em ordem de leitura.
func (b *BookState) SortedBookmarks() []Bookmark {
	marks := append([]Bookmark(nil), b.Bookmarks...)
	sort.Slice(marks, func(i, j int) bool {
		if marks[i].Chapter != marks[j].Chapter {
			return marks[i].Chapter < marks[j].Chapter
		}
		return marks[i].Percent < marks[j].Percent
	})
	return marks
}
