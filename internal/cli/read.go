package cli

import (
	"github.com/bernardofernandezz/tui-ebook-reader/internal/epub"
	"github.com/bernardofernandezz/tui-ebook-reader/internal/ui"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read <arquivo.epub>",
	Short: "Abre o leitor no terminal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReader(args[0])
	},
}

func runReader(path string) error {
	book, err := epub.Open(path)
	if err != nil {
		return err
	}
	return ui.Run(book)
}

func init() {
	rootCmd.AddCommand(readCmd)
}
