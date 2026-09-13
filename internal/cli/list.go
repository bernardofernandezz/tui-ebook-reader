package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bernardofernandezz/tui-ebook-reader/internal/store"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista os livros da biblioteca",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := store.LoadConfig().LibraryDir

		books, err := listBooks(dir)
		if err != nil {
			return err
		}
		if len(books) == 0 {
			fmt.Printf("nenhum EPUB em %s\n", dir)
			return nil
		}

		for _, book := range books {
			fmt.Println(strings.TrimSuffix(filepath.Base(book), filepath.Ext(book)))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
