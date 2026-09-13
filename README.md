# tbook

Leitor de EPUB no terminal, escrito em Go. Abre arquivos `.epub`, lista capítulos, renderiza o texto em uma coluna confortável de ler e desenha as imagens do livro com blocos coloridos — tudo sem sair do terminal.

![tbook lendo um EPUB no terminal](assets/image.png)

## Recursos

- Leitura de EPUB 2 e EPUB 3 direto no terminal (Bubble Tea + viewport).
- Capítulos vindos do sumário do próprio livro (NCX) e, sem ele, dos documentos do spine.
- Texto renderizado como Markdown via glamour, com temas `dark`, `light` e `sepia`.
- Coluna de leitura ajustável e centralizada (40–120 colunas, padrão 76).
- Imagens do livro convertidas para ANSI (24 bits, com fallback para 256 cores).
- Busca dentro do capítulo, bookmarks persistentes e retomada automática da leitura.
- Biblioteca: `tbook list` e abertura pelo nome (`tbook read harry`).
- Barra de progresso e estatísticas de tempo de leitura.

## Instalação

Binários para Linux e macOS ficam nas releases do GitHub. Com Go 1.27+ também dá para instalar direto:

```sh
go install github.com/bernardofernandezz/tui-ebook-reader/cmd/tbook@latest
```

O binário `tbook` vai para `$(go env GOPATH)/bin` — mantenha essa pasta no `PATH`.

### A partir do código

```sh
git clone https://github.com/bernardofernandezz/tui-ebook-reader
cd tui-ebook-reader
go build -o tbook ./cmd/tbook
```

O projeto não distribui livros: use seus próprios arquivos `.epub` (eles não são versionados).

## Uso

```sh
tbook livro.epub            # abre o leitor
tbook read harry            # procura "harry" na biblioteca
tbook toc livro.epub        # lista os capítulos numerados
tbook cat livro.epub -c 3   # imprime o capítulo 3 no stdout
tbook list                  # lista os livros da biblioteca
tbook stats                 # tempo de leitura por livro
tbook --version             # versão do binário
tbook --help                # ajuda de qualquer comando
```

A biblioteca é o diretório `library_dir` do config (padrão `~/Books`); `read`, `toc` e `cat` aceitam tanto um caminho quanto um nome de arquivo de lá.

### Atalhos no leitor

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

### Onde ficam os dados

Em `os.UserConfigDir()/tbook` (no Linux, `~/.config/tbook`):

- `config.json` — tema, largura da coluna e `library_dir`.
- `state.json` — posição de leitura, bookmarks e segundos lidos por livro.

## Como funciona

O código é organizado em `cmd/` + `internal/`, com responsabilidades separadas:

| Pacote           | Papel                                                                                            |
| ---------------- | ------------------------------------------------------------------------------------------------ |
| `internal/cli`   | Comandos (`read`, `toc`, `cat`, `list`, `stats`); o comando raiz abre o leitor.                   |
| `internal/epub`  | Wrapper do parser `raitucarp/epub`: capítulos via NCX, capa, imagens e limpeza do Markdown.       |
| `internal/ui`    | TUI Bubble Tea: overlays, temas embutidos, busca, layout centralizado e render ANSI.              |
| `internal/store` | Config e estado em JSON (posição, bookmarks, tempo), sem banco nem framework.                     |

Detalhes de implementação:

- **Capítulos**: o NCX do livro é a fonte principal; sem pelo menos dois destinos válidos, usa-se os documentos do spine, ignorando os muito curtos (capa, copyright).
- **Renderização**: o Markdown passa pelo glamour com temas embutidos em `internal/ui/themes/` (via `go:embed`). A largura da coluna é limitada e centralizada; o capítulo fica em cache por capítulo/largura.
- **Imagens**: cada referência é reduzida (até 48×24) e desenhada com `▀`; 24 bits quando disponível, 256 cores como fallback e sem cor em terminais ASCII.
- **Busca**: a consulta roda sobre as linhas já renderizadas (sem os códigos ANSI) e o viewport pula até a ocorrência.
- **Persistência**: posição, bookmarks e tempo vão para `state.json`; abrir um livro continua de onde a leitura parou.

### Estrutura do projeto

```
cmd/tbook/main.go          # entrada do binário
internal/cli/              # comandos Cobra
internal/epub/             # parsing, capítulos, capa e imagens
internal/store/            # config e estado em JSON
internal/ui/               # TUI, overlays e temas (themes/*.json)
.github/workflows/         # CI e release
.goreleaser.yml            # binários para Linux/macOS
assets/image.png           # screenshot usado neste README
```

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
