package store

import "testing"

func TestConfigRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg := LoadConfig()
	if cfg.Theme != DefaultTheme || cfg.Width != DefaultWidth {
		t.Fatalf("padrões errados: %+v", cfg)
	}
	if cfg.LibraryDir == "" {
		t.Fatal("diretório da biblioteca vazio")
	}

	cfg.Theme = "sepia"
	cfg.Width = 60
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}

	again := LoadConfig()
	if again.Theme != "sepia" || again.Width != 60 {
		t.Fatalf("config não persistiu: %+v", again)
	}
}

func TestStateBookmarks(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	st := LoadState()
	book := st.Book("/tmp/livro.epub")

	book.ToggleBookmark(Bookmark{Chapter: 3, Percent: 20, Label: "CAPITULO TRES"})
	if !book.HasBookmark(3) {
		t.Fatal("bookmark não foi adicionado")
	}
	book.ToggleBookmark(Bookmark{Chapter: 3})
	if book.HasBookmark(3) {
		t.Fatal("bookmark não foi removido")
	}

	book.ToggleBookmark(Bookmark{Chapter: 5, Percent: 10, Label: "CINCO"})
	book.Position = Position{Chapter: 5, Percent: 10}
	book.Seconds = 120
	if err := st.Save(); err != nil {
		t.Fatal(err)
	}

	again := LoadState().Book("/tmp/livro.epub")
	if !again.HasBookmark(5) {
		t.Fatal("bookmark não persistiu")
	}
	if again.Position.Chapter != 5 || again.Seconds != 120 {
		t.Fatalf("estado não persistiu: %+v", again)
	}
	if marks := again.SortedBookmarks(); len(marks) != 1 || marks[0].Chapter != 5 {
		t.Fatalf("bookmarks fora de ordem: %+v", marks)
	}
}
