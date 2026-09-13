package cli

import (
	"github.com/bernardofernandezz/tui-ebook-reader/internal/epub"
	"github.com/bernardofernandezz/tui-ebook-reader/internal/ui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui <arquivo.epub>",
	Short: "Abre o leitor interativo no terminal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		book, err := epub.Open(args[0])
		if err != nil {
			return err
		}
		return ui.Run(book)
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
