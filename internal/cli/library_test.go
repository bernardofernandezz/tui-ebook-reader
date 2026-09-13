package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveBook(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"harry-potter.epub", "harry-potter-2.epub", "outro-livro.epub"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	direct := filepath.Join(dir, "harry-potter.epub")
	if got, err := resolveBook(direct, dir); err != nil || got != direct {
		t.Fatalf("caminho direto: got %q, err %v", got, err)
	}

	if got, err := resolveBook("outro-livro.epub", dir); err != nil || filepath.Base(got) != "outro-livro.epub" {
		t.Fatalf("nome exato: got %q, err %v", got, err)
	}

	if _, err := resolveBook("harry", dir); err == nil {
		t.Fatal("nome ambíguo deveria dar erro")
	}

	if _, err := resolveBook("inexistente", dir); err == nil {
		t.Fatal("livro inexistente deveria dar erro")
	}

	if _, err := resolveBook("nao/existe.epub", dir); err == nil {
		t.Fatal("caminho inexistente deveria dar erro")
	}
}
