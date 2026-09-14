package cli

import (
	"github.com/bernardofernandezz/tui-ebook-reader/internal/core"
	"github.com/bernardofernandezz/tui-ebook-reader/internal/epub"
	"github.com/bernardofernandezz/tui-ebook-reader/internal/store"
	"github.com/bernardofernandezz/tui-ebook-reader/internal/ui"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read <arquivo.epub|nome>",
	Short: "Abre o leitor no terminal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReader(args[0])
	},
}

func runReader(name string) error {
	cfg := store.LoadConfig()

	path, err := resolveBook(name, cfg.LibraryDir)
	if err != nil {
		return err
	}

	book, err := epub.Open(path)
	if err != nil {
		return err
	}
	return ui.Run(core.New(book, store.LoadState()), cfg)
}

func init() {
	rootCmd.AddCommand(readCmd)
}
