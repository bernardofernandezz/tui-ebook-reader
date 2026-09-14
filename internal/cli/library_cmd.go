package cli

import (
	"fmt"

	"github.com/bernardofernandezz/tui-ebook-reader/internal/core"
	"github.com/bernardofernandezz/tui-ebook-reader/internal/store"
	"github.com/bernardofernandezz/tui-ebook-reader/internal/ui"
	"github.com/spf13/cobra"
)

var libraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Navega pela biblioteca com capa e progresso",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := store.LoadConfig()

		paths, err := listBooks(cfg.LibraryDir)
		if err != nil {
			return err
		}
		if len(paths) == 0 {
			fmt.Printf("nenhum EPUB em %s\n", cfg.LibraryDir)
			return nil
		}

		st := store.LoadState()
		book, err := ui.RunLibrary(paths, st)
		if err != nil {
			return err
		}
		if book == nil {
			return nil
		}
		return ui.Run(core.New(book, st), cfg)
	},
}

func init() {
	rootCmd.AddCommand(libraryCmd)
}
