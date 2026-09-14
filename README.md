# tbook

Um leitor de EPUB que mora no terminal. Abre o livro, mostra a capa, desenha as imagens em blocos coloridos e retoma a leitura exatamente de onde você parou — sem app pesado, sem janelas, sem distração.

![tbook lendo um EPUB no terminal](assets/image.png)

## A experiência

Você abre o terminal, digita `tbook livro.epub` e o livro está lá. Setas trocam de capítulo, `/` busca um trecho, `b` marca o ponto, `t` troca o tema. Ao sair, o tbook guarda posição, bookmarks e tempo de leitura — na próxima vez, você cai no mesmo parágrafo.

- **Leitura sem distração** — uma coluna centralizada de 40 a 120 colunas (padrão 76), em tema `dark`, `light` ou `sepia`.
- **Capa e ilustrações** — a capa abre a primeira página; as imagens do livro viram arte ANSI (24 bits, com fallback para 256 cores).
- **Busca e bookmarks** — `/` encontra, `n`/`N` pulam entre ocorrências; `b` marca o ponto e `B` leva direto até ele.
- **Memória de leitura** — posição, bookmarks e tempo por livro; `tbook stats` mostra seu histórico.
- **Sua estante no terminal** — `tbook library` navega pela biblioteca com capa e progresso ao lado; `tbook read harry` abre pelo nome.
- **Clássicos de graça** — `tbook search` e `tbook download` buscam e baixam EPUBs do Project Gutenberg direto para a sua biblioteca.

## Começando em 30 segundos

Sem nenhum livro no computador? Busque um clássico e baixe direto para a estante:

```sh
tbook search "pride and prejudice"
  1342  Pride and Prejudice — Austen, Jane

tbook download 1342
baixando Pride and Prejudice (Austen, Jane)...
salvo em ~/Books/Pride and Prejudice.epub

tbook library
```

Na tela da biblioteca: `j`/`k` para navegar, `enter` abre o livro, `q` sai. Já tem um arquivo? É só apontar:

```sh
tbook livro.epub
```

## Durante a leitura

| Tecla                     | Ação                                        |
| ------------------------- | ------------------------------------------- |
| `←` / `→` ou `h` / `l`    | capítulo anterior / próximo                 |
| `c`                       | lista de capítulos                          |
| `j` / `k` ou `↓` / `↑`    | rolar uma linha                             |
| `d` / `u`                 | meia página para baixo / para cima          |
| `espaço` / `pgup`         | página para baixo / para cima               |
| `/`                       | buscar no capítulo                          |
| `n` / `N`                 | próxima / anterior ocorrência da busca      |
| `b`                       | marcar / desmarcar bookmark no ponto atual  |
| `B`                       | abrir a lista de bookmarks                  |
| `t`                       | trocar o tema                               |
| `+` / `-`                 | alargar / estreitar a coluna                |
| `?`                       | ajuda com todos os atalhos                  |
| `q` ou `ctrl+c`           | sair                                        |

Nas listas (capítulos, bookmarks e ajuda): `j`/`k` movem, `enter` abre e `esc` fecha.

### Sua estante

```sh
tbook list                 # lista os livros da biblioteca
tbook read harry           # abre pelo nome, sem digitar o caminho
tbook toc livro.epub       # capítulos numerados
tbook cat livro.epub -c 3  # imprime o capítulo 3 no stdout
tbook stats                # tempo de leitura por livro
```

A biblioteca é o diretório `library_dir` do config (padrão `~/Books`); `read`, `toc` e `cat` aceitam tanto um caminho quanto um nome de lá.

### Onde ficam os dados

Em `os.UserConfigDir()/tbook` (no Linux, `~/.config/tbook`):

- `config.json` — tema, largura da coluna e `library_dir`.
- `state.json` — posição de leitura, bookmarks e segundos lidos por livro.

## Instalação

Binários para Linux e macOS ficam nas releases do GitHub. Com Go 1.27+ também dá para instalar direto:

```sh
go install github.com/bernardofernandezz/tui-ebook-reader/cmd/tbook@latest
```

O binário `tbook` vai para `$(go env GOPATH)/bin` — mantenha essa pasta no `PATH`.

A partir do código:

```sh
git clone https://github.com/bernardofernandezz/tui-ebook-reader
cd tui-ebook-reader
go build -o tbook ./cmd/tbook
```

O tbook lê EPUB 2 e 3 **sem DRM** — arquivos protegidos não abrem. O projeto não distribui livros: seus `.epub` ficam só com você (nada disso é versionado).

## Por baixo do capô

Para quem gosta dos detalhes:

- **Capítulos**: o NCX do livro manda; sem ele, os documentos do spine decidem, ignorando os muito curtos (capa, copyright).
- **Renderização**: Markdown via glamour com temas embutidos em `internal/ui/themes/` (via `go:embed`), com cache por capítulo/largura.
- **Imagens**: cada referência é reduzida (até 48×24) e desenhada com `▀`.
- **Sessão**: `internal/core` guarda capítulo, posição, bookmarks e tempo — sem nada de terminal; a TUI é só uma consumidora.
- **Persistência**: JSON puro, sem banco nem framework.

| Pacote           | Papel                                                                                            |
| ---------------- | ------------------------------------------------------------------------------------------------ |
| `internal/core`  | Sessão de leitura: capítulo atual, posição, bookmarks e tempo — independente de terminal.        |
| `internal/cli`   | Comandos (`read`, `toc`, `cat`, `list`, `library`, `search`, `download`, `stats`).               |
| `internal/epub`  | Wrapper do parser `raitucarp/epub`: capítulos via NCX, capa, imagens e limpeza do Markdown.       |
| `internal/ui`    | TUI Bubble Tea: overlays, temas embutidos, busca, layout centralizado e render ANSI.              |
| `internal/store` | Config e estado em JSON (posição, bookmarks, tempo), sem banco nem framework.                     |

## Próximos passos

Posição de leitura que sobrevive à troca de tela, API HTTP local e um leitor no navegador do celular: é para onde o núcleo aponta agora.

## Desenvolvimento

```sh
go run ./cmd/tbook seu-livro.epub   # roda sem gerar binário
go test ./...                       # testes
go vet ./...                        # checagem estática
go build ./...                      # compila todos os pacotes
```

## Feito com

- [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles) e [Lip Gloss](https://github.com/charmbracelet/lipgloss) — TUI.
- [Glamour](https://github.com/charmbracelet/glamour) — renderização de Markdown no terminal.
- [raitucarp/epub](https://github.com/raitucarp/epub) — parsing de EPUB.
- [Cobra](https://github.com/spf13/cobra) — linha de comando.
- [Gutendex](https://gutendex.com) — catálogo do Project Gutenberg.
