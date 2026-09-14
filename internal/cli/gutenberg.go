package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var gutendexBase = "https://gutendex.com/books"
var gutenbergClient = &http.Client{Timeout: 30 * time.Second}

// gutenbergBook é o subconjunto dos campos do Gutendex que usamos.
type gutenbergBook struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Authors []struct {
		Name string `json:"name"`
	} `json:"authors"`
	Formats map[string]string `json:"formats"`
}

type gutenbergResults struct {
	Results []gutenbergBook `json:"results"`
}

func searchGutenberg(query string) ([]gutenbergBook, error) {
	var out gutenbergResults
	if err := getJSON(gutendexBase+"?search="+url.QueryEscape(query), &out); err != nil {
		return nil, err
	}
	return out.Results, nil
}

func fetchGutenberg(id int) (gutenbergBook, error) {
	var out gutenbergBook
	err := getJSON(fmt.Sprintf("%s/%d", gutendexBase, id), &out)
	return out, err
}

func getJSON(rawURL string, v any) error {
	resp, err := gutenbergClient.Get(rawURL)
	if err != nil {
		return fmt.Errorf("gutendex: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gutendex: %s devolveu %s", rawURL, resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("gutendex: resposta inválida: %w", err)
	}
	return nil
}

func (b gutenbergBook) author() string {
	if len(b.Authors) == 0 {
		return "autor desconhecido"
	}
	return b.Authors[0].Name
}

func (b gutenbergBook) epubURL() (string, error) {
	if u := b.Formats["application/epub+zip"]; u != "" {
		return u, nil
	}
	return "", fmt.Errorf("livro %d (%s) sem versão EPUB", b.ID, b.Title)
}
