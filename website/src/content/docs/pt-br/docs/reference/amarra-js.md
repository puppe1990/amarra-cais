---
title: amarra.js
description: O contrato HTML público para Drive, frames e os builtins de amarra-hook inclusos.
sidebar:
  order: 6
---

`amarra.js` é o único script que um app gerado inclui (`/static/js/amarra.js`). Ele alimenta o **Drive** (interceptando navegações e fazendo morph da página) e os comportamentos embutidos de `amarra-hook`. O HTMX não é uma dependência pública.

## Como o Drive funciona

O Drive intercepta **todos os cliques e submits same-origin por padrão** — não existe atributo de opt-in. `<a>`, `<form>`, `{{ linkTo }}` e `<.form>` do kit são todos capturados, transformados em um `fetch` com o header `Amarra-Drive: true` mais CSRF, e a resposta faz morph de `#amarra-main`.

O primeiro carregamento, o `curl` e os crawlers recebem o layout completo. Remova um elemento do Drive com `data-amarra-skip`.

O layout renderiza `<meta name="csrf-token">`, e o Drive envia `X-CSRF-Token` em suas requisições (em par com o cookie double-submit `cais_csrf`).

## Atributos

| Atributo                        | Aplica-se a               | Efeito                                                                        |
| ------------------------------- | ------------------------- | ----------------------------------------------------------------------------- |
| `data-amarra-skip`              | `<a>`, `<form>`           | Remove esse elemento do Drive; ele navega ou envia normalmente.               |
| `data-amarra-confirm`           | click / submit            | Pede confirmação (o texto é o valor) antes de executar a requisição.          |
| `data-amarra-method`            | link                      | Usa este método HTTP para o link (por exemplo, `delete`).                     |
| `data-amarra-disable-with`      | submit / button           | Rótulo para o qual o controle muda enquanto sua requisição está em andamento. |
| `data-amarra-frame`             | link / form               | Aponta para um frame nomeado em vez do morph padrão de `#amarra-main`.        |
| `amarra-click`                  | element                   | Vincula uma ação de click.                                                    |
| `amarra-change`                 | element                   | Vincula uma ação de change.                                                   |
| `amarra-submit`                 | element                   | Vincula uma ação de submit.                                                   |
| `amarra-debounce`               | element                   | Aplica debounce à ação vinculada.                                             |
| `amarra-hook`                   | element                   | Anexa um hook embutido pelo nome (veja abaixo).                               |
| `amarra-live`                   | element                   | Habilita o elemento no hub Live via WebSocket.                                |
| `<amarra-frame loading="lazy">` | elemento `<amarra-frame>` | Elemento de frame; `loading="lazy"` o adia até que seja necessário.           |

## Hooks embutidos (`amarra-hook`)

Defina `amarra-hook="<name>"` em um container para ativar um comportamento embutido.

| Hook        | Markup esperado                                                                                                                                                                     | Comportamento                                                                                                                                                                                                                                                                                            |
| ----------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `bulk`      | container `amarra-hook="bulk"`, `input[data-amarra-bulk-all]`, checkboxes das linhas `input[data-amarra-bulk-row]`, `[data-amarra-bulk-bar]` + `[data-amarra-bulk-count]` opcionais | Selecionar tudo na página atual. O header mostra `indeterminate` quando algumas linhas estão selecionadas, e a barra opcional aparece quando a contagem é maior que zero. A ação em si continua um POST/DELETE do Drive com os ids selecionados.                                                         |
| `clipboard` | `data-amarra-copy`                                                                                                                                                                  | Copia o texto informado para a área de transferência.                                                                                                                                                                                                                                                    |
| `dialog`    | container `amarra-hook="dialog"`, `button[data-amarra-dialog-open]`, `dialog[data-amarra-dialog-target]`, `button[data-amarra-dialog-close]` opcional                               | `<dialog>` nativo via `showModal()`, então Esc, o focus trap e `::backdrop` vêm de graça. `aria-modal="true"` é definido na conexão e o foco volta para o opener ao fechar. O `<.modal>` renderiza o alvo do dialog.                                                                                     |
| `dropdown`  | container `amarra-hook="dropdown"`, `button[data-amarra-dropdown-button aria-expanded]`, `div[data-amarra-dropdown-menu hidden]`                                                    | Ações de linha / menus de overflow. O clique alterna; Esc e cliques fora fecham; um clique em item de menu fecha após agir.                                                                                                                                                                              |
| `nav`       | container nav `amarra-hook="nav"` com as listas de classes `data-amarra-nav-on` / `data-amarra-nav-off`                                                                             | Ressincroniza o link ativo após um morph do Drive. Compara `a.pathname === location.pathname` (query ignorada), define `aria-current="page"` e roda na conexão, em `amarra:morphed` e em `popstate`.                                                                                                     |
| `password`  | controle mais `data-amarra-password-for` opcional                                                                                                                                   | Alterna `type` e `aria-pressed`. Usa como fallback o `input` irmão mais próximo quando `data-amarra-password-for` está ausente; troca `aria-label` via `data-amarra-label-show` / `data-amarra-label-hide`. Os ícones aceitam `data-amarra-password-icon` (`data-cais-password-icon` é um alias legado). |
| `reveal`    | controle com `data-amarra-reveal-show` e `data-amarra-reveal-target`                                                                                                                | Mostra ou oculta o alvo quando o valor do controle corresponde — sem ida e volta ao Drive.                                                                                                                                                                                                               |
| `theme`     | controle com `amarra-hook="theme"`                                                                                                                                                  | Alterna `html.light`, persiste `localStorage["amarra-theme"]` e atualiza o meta `theme-color` opcional.                                                                                                                                                                                                  |

### Exemplos de hook

```html
<div amarra-hook="clipboard" data-amarra-copy="hi">Copy</div>

<button
  type="button"
  amarra-hook="password"
  data-amarra-password-for="#password"
  data-amarra-label-show="Show password"
  data-amarra-label-hide="Hide password"
>
  Show
</button>

<select
  amarra-hook="reveal"
  data-amarra-reveal-show="access_keys"
  data-amarra-reveal-target="#aws-keys"
>
  <option value="default_chain">Default chain</option>
  <option value="access_keys">Access keys</option>
</select>
<div id="aws-keys" hidden>…keys…</div>

<button
  type="button"
  amarra-hook="theme"
  data-amarra-theme-key="app-theme"
  data-amarra-theme-color="#f5f5f4"
  data-amarra-theme-color-off="#0f172a"
  data-amarra-theme-on-label="Dark mode"
  data-amarra-theme-off-label="Light mode"
>
  Light mode
</button>
```

O hook `theme` é configurável por elemento:

| Atributo                                                     | Padrão         | Finalidade                                      |
| ------------------------------------------------------------ | -------------- | ----------------------------------------------- |
| `data-amarra-theme-key`                                      | `amarra-theme` | Chave do `localStorage`.                        |
| `data-amarra-theme-class`                                    | `light`        | Classe alternada no `html`.                     |
| `data-amarra-theme-color` / `data-amarra-theme-color-off`    | —              | Hex para o meta `theme-color`.                  |
| `data-amarra-theme-on-label` / `data-amarra-theme-off-label` | —              | Troca o texto do botão (mantém `aria-pressed`). |

## Snippet de FOUC do tema

Coloque isto no `<head>` do layout antes do CSS para que um tema claro armazenado não pisque escuro. Se você definir um `data-amarra-theme-key` personalizado, copie essa chave para o snippet para que a restauração corresponda.

```html
<script>
  try {
    if (localStorage.getItem("amarra-theme") === "light")
      document.documentElement.classList.add("light");
  } catch (e) {}
</script>
```

:::note
A Content Security Policy mantém `script-src 'self' 'unsafe-inline'` para que este snippet de FOUC e o init inline do Drive rodem; o escaping é a defesa principal. Veja o [modelo de segurança](/amarra-cais/pt-br/docs/explanation/security-model/).
:::

:::tip
`amarra-live` é opt-in e só é necessário para páginas em tempo real; o CRUD fica no Drive. Veja [Formulários e validação](/amarra-cais/pt-br/docs/how-to/forms-and-validation/) para o lado do formulário no contrato.
:::
