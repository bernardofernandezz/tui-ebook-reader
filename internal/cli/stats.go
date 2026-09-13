package cli

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/bernardofernandezz/tui-ebook-reader/internal/store"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Mostra o tempo de leitura por livro",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		st := store.LoadState()
		if len(st.Books) == 0 {
			fmt.Println("nenhuma leitura registrada ainda")
			return nil
		}

		type row struct {
			path string
			book *store.BookState
		}
		rows := make([]row, 0, len(st.Books))
		for path, book := range st.Books {
			rows = append(rows, row{path: path, book: book})
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].book.Seconds > rows[j].book.Seconds })

		total := 0
		for _, r := range rows {
			total += r.book.Seconds
			fmt.Printf("%8s  %s", formatSeconds(r.book.Seconds), filepath.Base(r.path))
			if n := len(r.book.Bookmarks); n > 0 {
				fmt.Printf("  (%d bookmarks)", n)
			}
			fmt.Println()
		}
		fmt.Printf("\ntotal: %s\n", formatSeconds(total))
		return nil
	},
}

func formatSeconds(seconds int) string {
	switch {
	case seconds < 60:
		return fmt.Sprintf("%ds", seconds)
	case seconds < 3600:
		return fmt.Sprintf("%dmin", seconds/60)
	default:
		return fmt.Sprintf("%dh%02dmin", seconds/3600, (seconds%3600)/60)
	}
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
