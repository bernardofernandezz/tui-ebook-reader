# AGENTS.md — TUI-Ebook-Reader

## O que é este projeto

TUI ebook reader em Go. Objetivo: abrir arquivos EPUB no terminal, navegar capítulos e ler o texto com teclado.

Stack principal:
- Go
- `github.com/raitucarp/epub` (parser)
- Charm (Bubble Tea, Bubbles, Lip Gloss) para a interface

## Contexto atual

Projeto de estudos. O autor está aprendendo:
- módulos Go (`go.mod`, imports, alias)
- estrutura `cmd/` + `internal/`
- parsing de EPUB (spine, TOC, metadados)
- TUI com Bubble Tea

Estado aproximado:
- leitura no terminal já funciona: biblioteca, temas, busca, bookmarks e stats
- `internal/core` concentra a sessão de leitura (capítulo, posição, bookmarks, tempo), independente de terminal
- o roadmap de evolução (posição portável, API HTTP, leitor no celular) segue em aberto

## Como o assistente deve responder

Este repositório é para **estudo e prática**.

Portanto:

1. **Não entregue implementações completas** como solução final pronta para copiar sem pensar.
2. Prefira **guias passo a passo**, com:
   - o que fazer
   - por que fazer
   - o que observar no resultado
3. Pode mostrar **trechos de código curtos** quando ajudarem a entender um conceito.
4. Evite dumps grandes de arquivos inteiros, salvo quando o usuário pedir explicitamente um esqueleto mínimo.
5. Quando houver erro, explique a **causa** e o **ajuste pontual**, não reescreva o projeto inteiro.
6. Mantenha o escopo no MVP enquanto o usuário não pedir avanço:
   - abrir EPUB
   - listar/navegar capítulos
   - ler texto no terminal
7. Ensine trade-offs simples (ex.: spine ≠ TOC, valor vs ponteiro, wrapper de biblioteca).
8. Responda em português, de forma direta e organizada.

## O que evitar

- Implementar features extras sem pedido (busca, biblioteca, sync, temas complexos)
- Trocar a stack sem necessidade
- Respostas longas e confusas com vários caminhos ao mesmo tempo
- Assumir que o usuário quer código final de produção

## Forma preferida de ajuda

- Um objetivo claro por resposta
- Estrutura de arquivos sugerida (quando fizer sentido)
- Próximo passo único e verificável
- Perguntas de checagem do tipo: “compilou?”, “o que apareceu no terminal?”

## Resumo da postura

Guiar. Explicar. Sugerir.  
Não substituir o aprendizado por implementação pronta.