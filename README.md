# tbook

Leitor de EPUB no terminal, escrito em Go com Bubble Tea.

## Instalação

Com Go 1.27+ instalado:

```sh
go install github.com/bernardofernandezz/tui-ebook-reader/cmd/tbook@latest
```

O binário vai para `$(go env GOPATH)/bin` — deixe essa pasta no `PATH`.

## Uso

```sh
tbook livro.epub            # abre o leitor
tbook read livro.epub       # idem
tbook toc livro.epub        # lista os capítulos
tbook cat livro.epub -c 3   # imprime o capítulo 3 no stdout
```

## Atalhos no leitor

| tecla                  | ação                                          |
| ---------------------- | --------------------------------------------- |
| `←` / `→` ou `h` / `l` | capítulo anterior / próximo                   |
| `j` / `k` ou `↑` / `↓` | rolar linha a linha                           |
| `d` / `u`              | meia página                                   |
| `espaço` / `pgup`      | página inteira                                |
| `b`                    | marcar / desmarcar bookmark no ponto atual    |
| `B`                    | abrir a lista de bookmarks (`enter` vai até)  |
| `q`                    | sair                                          |

Os bookmarks ficam em `~/.config/tbook/bookmarks.json`. Imagens do EPUB são
desenhadas no próprio texto com blocos coloridos.
