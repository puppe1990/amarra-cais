---
title: Views e o kit incluso
description: O contrato do loader de templates do Amarra e os componentes de kit embutidos que você pode usar e sobrescrever.
sidebar:
  order: 5
---

O Amarra renderiza HTML no servidor. Os handlers chamam `view.Write`, que resolve um layout mais o corpo de uma página e expande quaisquer tags de kit `<.component>`. O boot carrega todos os templates uma vez com `view.Load`; uma tag `<.component>` desconhecida falha no boot, e não na primeira requisição.

## Contrato do loader de templates

`view.Load` faz glob de `web/templates/` uma vez no boot (`pkg/amarra/view/renderer.go`). O contrato é fixado por `pkg/amarra/view/renderer_contract_test.go`.

| Glob                             | Endereçável como               | Observações                                                                                                              |
| -------------------------------- | ------------------------------ | ------------------------------------------------------------------------------------------------------------------------ |
| `layouts/*.html`                 | `view.Page{Layout: "app"}`     | Plano. Pelo menos um layout é obrigatório.                                                                               |
| `pages/*.html`, `pages/*/*.html` | `view.Page{Name: "blog/post"}` | O Name é o caminho sob `pages/` sem o `.html`. Um nível de aninhamento, então um blog pode manter um arquivo por artigo. |
| `partials/*.html`                | `{{ template "card" . }}`      | Apenas plano. `partials/posts/card.html` nunca é carregado, e o erro aparece em tempo de renderização, não no boot.      |
| `components/*.html`              | override de `<.input>`         | Apenas plano. Kit incluso mais overrides do app, indexados pelo stem do arquivo. Um `<.x>` desconhecido falha no boot.   |

:::caution
`partials/` é apenas plano. Um partial aninhado como `partials/posts/card.html` é ignorado silenciosamente por `view.Load`, e a chamada `{{ template }}` só falha quando a página que precisa dele é renderizada.
:::

## Renderizando uma página

`view.Page` carrega o nome do layout, o nome da página e o map de dados:

```go
view.Write(w, r, h.views, view.Page{
  Layout: "app",
  Name:   "contact",
  Data: map[string]any{
    "Title":     h.catalog.T("contact.title"),
    "Site":      meta.ForRequest(h.site, r),
    "CSRFToken": csrf.TokenFromRequest(r),
    "Flash":     flashMsg,
  },
}, h.cfg)
```

Requisições Drive ainda renderizam o layout completo para que o `amarra.js` possa fazer morph de `#amarra-main`. Requisições Frame (`Amarra-Frame: <id>`) renderizam apenas o bloco `{{ define "frame:<id>" }}` correspondente.

Os layouts fornecem os marcadores de shell `#amarra-nav`, `#amarra-main` e `#amarra-toast-host`, e carregam um único script em `/static/js/amarra.js`.

## Corpos de página e frames

Toda página define um bloco `content` e pode definir blocos de frame nomeados. O layout renderiza o `content` da página dentro de `#amarra-main`.

```html
{{ define "content" }}
<h1>{{ .Title }}</h1>
<.form action="/items" method="post">
  <.input name="title" label="Title" value="{{ .Item.Title }}" error="{{ fieldError .Errors "title" }}" />
  <.button type="submit">Save</.button>
</.form>
{{ end }}

{{ define "frame:cart" }}
<div id="cart">{{ .CartTotal }}</div>
{{ end }}
```

Uma requisição de frame devolve o bloco `frame:<id>` sozinho, então o Drive pode trocar uma região menor em vez de todo o `#amarra-main`.

## Kit incluso

O kit fica em `web/templates/components/`. Sobrescreva um componente escrevendo um arquivo com o mesmo stem (por exemplo, `locale-toggle.html`); `amarra-cais g component <kit-name>` faz o seed do markup incluso e `amarra-cais g component --list` imprime os nomes que podem ser sobrescritos. Veja [Generators](/amarra-cais/pt-br/docs/reference/generators/).

| Componente      | Finalidade                                                        | Atributos documentados                                                                                                                                                |
| --------------- | ----------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `form`          | Shell de formulário                                               | `action`, `method`, `enctype`. Injeta `csrf_token` a partir de `.CSRFToken`.                                                                                          |
| `input`         | Input de texto/email/arquivo                                      | `name`, `type`, `label`, `value`, `error`; `accept` em inputs de arquivo.                                                                                             |
| `password`      | Input de senha com alternância do olho (`amarra-hook="password"`) | `name`, `label`, `value`, `error`, `autocomplete`, `required`.                                                                                                        |
| `select`        | Campo select                                                      | O corpo é o contrato do markup.                                                                                                                                       |
| `textarea`      | Input de múltiplas linhas                                         | —                                                                                                                                                                     |
| `checkbox`      | Checkbox                                                          | —                                                                                                                                                                     |
| `button`        | Botão                                                             | `type` (por exemplo, `type="submit"`).                                                                                                                                |
| `flash`         | Aviso flash de uso único                                          | `<.flash />` autocontido é permitido.                                                                                                                                 |
| `nav`           | Container de navegação                                            | —                                                                                                                                                                     |
| `pagination`    | Paginação de listas                                               | `base`.                                                                                                                                                               |
| `modal`         | Alvo do dialog para o hook `dialog`                               | —                                                                                                                                                                     |
| `locale-toggle` | Seletor de locale / idioma                                        | —                                                                                                                                                                     |
| `stat`          | Card de KPI                                                       | `label`, `value`; `delta`, `hint`, `href` opcionais.                                                                                                                  |
| `empty`         | Estado vazio                                                      | `title`, slot de corpo; link `href` ou `action` opcional.                                                                                                             |
| `table`         | Tabela de dados                                                   | As linhas vão no slot `.Inner`; `cols` vêm dos dados da página (maps `Field`/`Label`/`Sortable`); `sort`/`dir` para links de ordenação no servidor; `frame` opcional. |
| `filters`       | Formulário de filtro GET                                          | `action`; `frame`, `submit`, `clear` href opcionais. Os campos `.Inner` viram query params; inputs ocultos preservam `sort`/`page`.                                   |

Todo componente tem exatamente um slot: `.Inner`. O `<.form>` injeta o campo CSRF para você, então não adicione outro dentro do slot. Para uploads, use `<.input type="file" accept="image/*" />` dentro de `<.form enctype="multipart/form-data">` e nunca defina `value` em um input de arquivo.

## Interpolação de atributos do kit

Os atributos do kit interpolam como markup comum, então um único valor pode misturar texto estático com dados da página:

```html
<.stat label="Potência" value="{{ .Power }} kWp" />
```

renderiza `5 kWp`. Um `{{ .X }}` sozinho mantém a expressão bruta, então dados que não são string continuam funcionando dentro dos próprios blocos `if`/`range` do componente.

Não coloque uma ação de controle em um valor de atributo: um `{{ if }}` ou `{{ range }}` dentro de um atributo falha no boot em vez de imprimir `{{ … }}` para o usuário.

:::note
Ambas as regras de falha no boot são intencionais. Uma tag `<.x>` desconhecida e uma ação de controle dentro de um valor de atributo param o processo na inicialização, então o erro nunca chega a uma página renderizada.
:::

:::tip
O Drive devolve o layout inteiro enquanto os frames devolvem um bloco — veja [amarra.js](/amarra-cais/pt-br/docs/reference/amarra-js/) para as regras de interrupção e [template helpers](/amarra-cais/pt-br/docs/reference/template-helpers/) para `csrfField`, `linkTo` e os helpers de formulário. A justificativa por trás do modelo HTML-first está em [Views e Drive](/amarra-cais/pt-br/docs/explanation/views-and-drive/).
:::
