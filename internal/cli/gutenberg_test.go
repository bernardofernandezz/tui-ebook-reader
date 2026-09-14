package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGutendex(t *testing.T) {
	const body = `{"id":1342,"title":"Pride and Prejudice","authors":[{"name":"Austen, Jane"}],"formats":{"text/html":"https://x/1342.html","application/epub+zip":"https://x/1342.epub"}}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/books":
			io.WriteString(w, `{"results":[`+body+`]}`)
		case "/books/1342":
			io.WriteString(w, body)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	old := gutendexBase
	gutendexBase = srv.URL + "/books"
	defer func() { gutendexBase = old }()

	books, err := searchGutenberg("pride and prejudice")
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 1 || books[0].ID != 1342 || books[0].author() != "Austen, Jane" {
		t.Fatalf("busca inesperada: %+v", books)
	}

	book, err := fetchGutenberg(1342)
	if err != nil {
		t.Fatal(err)
	}
	if url, err := book.epubURL(); err != nil || url != "https://x/1342.epub" {
		t.Fatalf("epubURL: got %q, err %v", url, err)
	}

	if _, err := fetchGutenberg(999); err == nil {
		t.Fatal("id inexistente deveria dar erro")
	}
}

func TestSanitizeFilename(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Pride and Prejudice", "Pride and Prejudice"},
		{"Dom Casmurro", "Dom Casmurro"},
		{"A/B:C*D?", "A-B-C-D"},
		{"  ...  ", ""},
	}
	for _, c := range cases {
		if got := sanitizeFilename(c.in); got != c.want {
			t.Errorf("sanitizeFilename(%q) = %q, quer %q", c.in, got, c.want)
		}
	}
}
