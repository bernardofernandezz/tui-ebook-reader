package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bernardofernandezz/tui-ebook-reader/internal/epub"
	"github.com/bernardofernandezz/tui-ebook-reader/internal/store"
)

// openBook resolve o nome na biblioteca (ou um caminho direto) e abre o EPUB.
func openBook(name string) (*epub.Book, error) {
	path, err := resolveBook(name, store.LoadConfig().LibraryDir)
	if err != nil {
		return nil, err
	}
	return epub.Open(path)
}

// resolveBook aceita um caminho existente ou procura o nome na biblioteca.
func resolveBook(name, libraryDir string) (string, error) {
	if info, err := os.Stat(name); err == nil && !info.IsDir() {
		return name, nil
	}
	if strings.ContainsRune(name, os.PathSeparator) {
		return "", fmt.Errorf("arquivo não encontrado: %s", name)
	}

	books, err := listBooks(libraryDir)
	if err != nil {
		return "", err
	}

	var matches []string
	for _, book := range books {
		base := filepath.Base(book)
		if strings.EqualFold(base, name) {
			return book, nil
		}
		if strings.Contains(strings.ToLower(base), strings.ToLower(name)) {
			matches = append(matches, base)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("nenhum livro encontrado para %q em %s", name, libraryDir)
	case 1:
		return filepath.Join(libraryDir, matches[0]), nil
	default:
		return "", fmt.Errorf("%q combina com %d livros: %s", name, len(matches), strings.Join(matches, ", "))
	}
}

// listBooks devolve os arquivos .epub do diretório em ordem alfabética.
func listBooks(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("biblioteca %s: %w", dir, err)
	}

	var books []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".epub") {
			continue
		}
		books = append(books, filepath.Join(dir, entry.Name()))
	}
	return books, nil
}
