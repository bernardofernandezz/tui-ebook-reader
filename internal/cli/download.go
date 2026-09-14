package cli

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/bernardofernandezz/tui-ebook-reader/internal/store"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download <id>",
	Short: "Baixa um EPUB do Project Gutenberg para a biblioteca",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("id inválido: %q (use o número do Project Gutenberg, ex.: 1342)", args[0])
		}

		book, err := fetchGutenberg(id)
		if err != nil {
			return err
		}
		src, err := book.epubURL()
		if err != nil {
			return err
		}

		cfg := store.LoadConfig()
		if err := os.MkdirAll(cfg.LibraryDir, 0o755); err != nil {
			return err
		}

		name := sanitizeFilename(book.Title)
		if name == "" {
			name = fmt.Sprintf("gutenberg-%d", book.ID)
		}
		dest := filepath.Join(cfg.LibraryDir, name+".epub")
		if _, err := os.Stat(dest); err == nil {
			fmt.Printf("já existe: %s\n", dest)
			return nil
		}

		fmt.Printf("baixando %s (%s)...\n", book.Title, book.author())
		if err := downloadFile(src, dest); err != nil {
			return err
		}
		fmt.Printf("salvo em %s\n", dest)
		return nil
	},
}

// downloadFile grava num arquivo temporário e só então renomeia, para não
// deixar EPUB pela metade se a conexão cair.
func downloadFile(rawURL, dest string) error {
	resp, err := gutenbergClient.Get(rawURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download: %s devolveu %s", rawURL, resp.Status)
	}

	tmp := dest + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dest)
}

// sanitizeFilename deixa no título apenas o que é seguro em nome de arquivo.
func sanitizeFilename(title string) string {
	safe := strings.Map(func(r rune) rune {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			return r
		case r == ' ', r == '-', r == '_', r == '.', r == ',':
			return r
		default:
			return '-'
		}
	}, title)
	safe = strings.Join(strings.Fields(safe), " ")
	return strings.Trim(safe, " -.")
}

func init() {
	rootCmd.AddCommand(downloadCmd)
}
