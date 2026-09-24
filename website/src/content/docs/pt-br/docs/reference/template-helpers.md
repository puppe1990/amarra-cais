---
title: Template helpers
description: As funções de template que o Amarra registra no view renderer para CSRF, formulários, flash, links e i18n.
sidebar:
  order: 7
---

O Amarra registra um conjunto de funções helper no view renderer para que os templates de página continuem declarativos. Elas cobrem campos CSRF, campos de formulário e seus erros, mensagens flash, links e localização.

## Referência de helpers

| Helper            | Chamada de exemplo                                                                        | Renderiza                                                      |
| ----------------- | ----------------------------------------------------------------------------------------- | -------------------------------------------------------------- |
| `csrfField`       | `{{ csrfField .CSRFToken }}`                                                              | Um input CSRF oculto (token double-submit).                    |
| `fieldError`      | `{{ fieldError .Errors "name" }}`                                                         | A string de erro de um campo de um map `validate.FieldErrors`. |
| `makeField`       | `makeField "name" "Name" .Name "text" true .Errors`                                       | Um valor `forms.FieldData` passado para `fieldInput`.          |
| `fieldInput`      | `{{ fieldInput (makeField ...) }}`                                                        | Um input, textarea ou checkbox mais o seu erro.                |
| `fieldPassword`   | `{{ fieldPassword ... }}`                                                                 | Um input de senha com a alternância de mostrar/ocultar (olho). |
| `makeSelectField` | `makeSelectField "category_id" "Category" .Item.CategoryID .CategoryOptions true .Errors` | Um valor `forms.FieldData` de select.                          |
| `fieldSelect`     | `{{ fieldSelect (makeSelectField ...) }}`                                                 | Um `<select>` de chave estrangeira mais o seu erro.            |
| `flashMessage`    | `{{ flashMessage .Flash }}`                                                               | O aviso flash de uso único.                                    |
| `linkTo`          | `{{ linkTo "/items/1" "Delete" (dict "method" "delete" "confirm" "Delete this item?") }}` | Um `<a>` simples que o Drive intercepta.                       |
| `t`               | `{{ t "key" }}`                                                                           | Uma string localizada do catálogo.                             |
| `dict`            | `(dict "method" "delete" "confirm" "Sure?" "frame" "cart")`                               | Um map, usado para passar opções.                              |

## CSRF

Os formulários combinam o cookie double-submit `cais_csrf` com um campo de token ou o header `X-CSRF-Token`. Nos templates, você mesmo emite o campo oculto:

```html
<form action="/contact" method="post">{{ csrfField .CSRFToken }} ...</form>
```

O `<.form>` do kit injeta o campo CSRF a partir dos dados raiz (`$.CSRFToken`), então não adicione `csrfField` dentro do slot dele. O layout renderiza `<meta name="csrf-token">`, e o [amarra.js](/amarra-cais/pt-br/docs/reference/amarra-js/) envia `X-CSRF-Token` nas requisições Drive.

## Campos de formulário e validação

Colete os erros em um map `validate.FieldErrors` e então passe-o para a página como `.Errors`:

```go
var errs validate.FieldErrors
if item.Name == "" {
  errs.Add("name", "Name is required")
}
if errs.Any() {
  // re-render the form with errs as .Errors
}
```

Verificações de campo único usam `validate.Email`, `validate.URL`, `validate.Required`, `validate.MinLength` e `validate.MaxLength`.

Renderize um campo e o seu erro com os helpers agrupados. `makeField` constrói um valor `forms.FieldData` e `fieldInput` renderiza input, textarea ou checkbox mais o erro:

```html
{{ fieldInput (makeField "name" "Name" .Name "text" true .Errors) }}
```

Use sempre `fieldPassword` para campos de senha — a alternância de mostrar/ocultar (olho) é o padrão.

Os selects de chave estrangeira usam `makeSelectField` (as opções vêm de um método `List*Options` do store) e `fieldSelect`:

```html
{{ fieldSelect (makeSelectField "category_id" "Category" .Item.CategoryID .CategoryOptions true
.Errors) }}
```

O [generator de resource](/amarra-cais/pt-br/docs/reference/generators/) conecta esses helpers mais o `<.form>` e o `<.button>` do kit aos formulários do admin. Veja [Formulários e validação](/amarra-cais/pt-br/docs/how-to/forms-and-validation/).

## Mensagens flash

`flash.Set(w, "notice", "Saved!", cfg.CookieSecure())` escreve um cookie de uso único; leia-o na próxima requisição com `flash.MessageFromRequest(r)` e passe-o para os dados da página. No template, sempre o renderize com `flashMessage`:

```html
{{ flashMessage .Flash }}
```

:::caution
Nunca imprima `{{ .Flash }}` — isso transforma a struct em string. Mantenha `<.flash />` dentro de `#amarra-main` para que um morph do Drive preserve o aviso.
:::

## Links

`linkTo` emite um `<a>` simples; o Drive o intercepta por padrão. Passe um `dict` para adicionar opções:

```html
{{ linkTo "/items/1" "Delete" (dict "method" "delete" "confirm" "Delete this item?") }}
```

Opções: `"method"` (por exemplo, `delete`), `"confirm"` (texto de confirmação) e `"frame"` (um id de frame, como `cart`).

## i18n

`{{ t "key" }}` procura uma string no catálogo de locale. O layout também usa `{{ htmlLang }}` e `{{ ogLocale }}`. Chaves ausentes devolvem a própria chave, o que fica visível em desenvolvimento.

O catálogo é selecionado pela env var `LOCALE` (`en` é o padrão, `pt` é suportado) e conectado como `cais.Config.Locale`. Os handlers recebem o catálogo para validação e strings flash (`catalog.T("contact.title")`), e as mesmas funções chegam a páginas e partials. Veja [i18n](/amarra-cais/pt-br/docs/how-to/i18n/) para os arquivos de locale.
