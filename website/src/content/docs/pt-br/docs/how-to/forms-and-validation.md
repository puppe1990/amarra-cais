---
title: Formulários e validação
description: Use as tags de formulário do kit incluído, CSRF, erros de campo e upload de arquivos em um app Amarra.
sidebar:
  order: 3
---

Amarra inclui um kit pequeno para formulários, então você escreve markup em vez de montar inputs na mão. As tags do kit são expandidas no boot, e `<.form>` injeta o campo CSRF para você.

## As tags de formulário do kit

```html
{{ define "content" }}
<.form action="/items" method="post">
  <.input name="title" label="Title" value="{{ .Item.Title }}" error="{{ fieldError .Errors "title" }}" />
  <.button type="submit">Save</.button>
</.form>
{{ end }}
```

- `<.form>` renderiza o elemento `<form>` e injeta o campo `csrf_token` a partir dos dados raiz (`$.CSRFToken`). Não adicione um segundo campo CSRF dentro do slot.
- `<.input>` aceita `name`, `label`, `value` e `error`. Renderize o erro com `{{ fieldError .Errors "title" }}`.
- `<.button>` renderiza um botão de submit ou de ação.

O kit também inclui `<.select>`, `<.textarea>` e `<.checkbox>` para esses tipos de controle; dê a cada um o `name` e o `label` do campo, e passe `error` da mesma forma que faz para `<.input>`. Todo componente aceita overrides do app em `web/templates/components/`, indexados pelo nome-base do arquivo, então você pode re-estilizar o contrato real em vez de recriá-lo.

## CSRF

CSRF é um cookie double-submit (`cais_csrf`): `middleware.CSRF(cfg)` valida todo `POST`, `PUT`, `DELETE` e `PATCH`. O `<.form>` do kit injeta o campo automaticamente. Em formulários escritos à mão, use `{{ csrfField .CSRFToken }}` ou um input oculto `csrf_token`. Requisições do Drive enviam o token como `X-CSRF-Token`, alimentado pelo `<meta name="csrf-token">` do layout.

## Validação no lado do servidor

Valide no handler e re-renderize a mesma página com `422` quando algo estiver errado. Para um único campo, use os helpers em `pkg/cais/validate` — `validate.Email`, `validate.URL`, `validate.Required`, `validate.MinLength`, `validate.MaxLength`. Para um formulário inteiro, colete os erros em `validate.FieldErrors`:

```go
var errs validate.FieldErrors
if item.Name == "" {
  errs.Add("name", "Name is required")
}
if errs.Any() {
  writeView(w, r, h.views, h.cfg, "app", "item", amarraData(r, h.site, map[string]any{
    "Errors": errs,
    "Item":   item,
  }), http.StatusUnprocessableEntity)
  return
}
```

Passe os erros como `.Errors` nos dados da página; os inputs os leem com `{{ fieldError .Errors "name" }}`.

## Helpers de template

`pkg/cais/forms` registra helpers no renderer de views para construir campos a partir de valores Go:

```html
{{ fieldInput (makeField "name" "Name" .Name "text" true .Errors) }}
```

`makeField` retorna um `forms.FieldData`, e `fieldInput` renderiza o controle correto (input, textarea ou checkbox) mais o seu erro.

:::tip
Campos de senha sempre usam `fieldPassword`, para que recebam o toggle de mostrar/ocultar com o olho. Selects de chave estrangeira usam `fieldSelect` com `makeSelectField`:

```html
{{ fieldSelect (makeSelectField "category_id" "Category" .Item.CategoryID .CategoryOptions true
.Errors) }}
```

:::

## Upload de arquivos

Defina `enctype="multipart/form-data"` no formulário e use `<.input type="file" />`. Não defina `value` em um input de arquivo.

```html
<.form action="/upload" method="post" enctype="multipart/form-data">
  <.input type="file" name="avatar" accept="image/*" label="Avatar" />
  <.button type="submit">Upload</.button>
</.form>
```

## Próximos passos

- [Páginas e views](/amarra-cais/pt-br/docs/how-to/pages-and-views/) — como o handler renderiza e re-renderiza a página.
- [Helpers de template](/amarra-cais/pt-br/docs/reference/template-helpers/) — a lista completa de helpers.
