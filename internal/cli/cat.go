package cli

import (
	"fmt"

	"github.com/bernardofernandezz/tui-ebook-reader/internal/epub"
	"github.com/spf13/cobra"
)

var catChapter int

var catCmd = &cobra.Command{
	Use:   "cat <arquivo.epub>",
	Short: "Imprime um capítulo no stdout",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		book, err := epub.Open(args[0])
		if err != nil {
			return err
		}
		if catChapter < 1 || catChapter > len(book.Chapters) {
			return fmt.Errorf("capítulo %d não existe (o livro tem %d)", catChapter, len(book.Chapters))
		}
		fmt.Print(book.Text(book.Chapters[catChapter-1]))
		return nil
	},
}

func init() {
	catCmd.Flags().IntVarP(&catChapter, "cap", "c", 1, "número do capítulo")
	rootCmd.AddCommand(catCmd)
}
