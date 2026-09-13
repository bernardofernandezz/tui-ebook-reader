package cli

import (
	"fmt"

	"github.com/bernardofernandezz/tui-ebook-reader/internal/epub"
	"github.com/spf13/cobra"
)

var tocCmd = &cobra.Command{
	Use:   "toc <arquivo.epub>",
	Short: "Lista os capítulos do EPUB",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		book, err := epub.Open(args[0])
		if err != nil {
			return err
		}
		for i, ch := range book.Chapters {
			fmt.Printf("%3d  %s\n", i+1, ch.Title)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tocCmd)
}
