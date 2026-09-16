# Sidebar fixa pós-login — design

Escolhas validadas com o usuário (brainstorm 2026-09-16): direção **A**
(sidebar fixa completa), itens **mínimo + patch**, escopo **todos os scaffolds**
(full, minimal, blank).

## 1. Shell (`layouts/app.html`)

- A nav horizontal (`#amarra-nav`, `top-[57px]`) sai; entra sidebar fixa:
  `fixed left-0 top-[57px] bottom-0 w-60`, abaixo do header existente
  (header com logo + login continua igual).
- `#amarra-main` ganha `lg:ml-60`; mobile sem margem (drawer sobrepõe).
- Header, `<.flash />`, `#amarra-toast-host` e morph do Drive não mudam —
  a sidebar vive fora do `#amarra-main`, como a nav atual.
- A faixa vazia do `--minimal`/`--blank` (nav sem links, só borda) deixa
  de existir: a sidebar sempre nasce com conteúdo (seção 2).

## 2. Itens + contrato de patch

- Sidebar nasce só com **Dashboard + logout** (logout é o `<.form action="/logout">`
  existente, com o botão estilizado como item da sidebar); sem Home/Contact.
- O marcador `<!-- cais:nav -->` continua dentro do container da sidebar,
  então `g resource --public` e `destroy` funcionam sem mudança
  (`resource_patch.go` injeta/remova no mesmo ponto).
- Full, minimal e blank nascem idênticos nesse ponto.

## 3. Mobile + Drive

- Mobile: sidebar vira drawer aberto por hamburger no header, sem JS —
  checkbox + `peer-checked` (padrão Tailwind já usado no projeto).
- `amarra-hook="nav"` + `data-amarra-nav-on/off` vão no container da
  sidebar: o sync de link ativo após morph do Drive continua funcionando.
- Sem mudança no `amarra.js` nem nos handlers; só template + CSS.

## Fora de escopo

- Segundo layout "dashboard" (opção C descartada).
- Sidebar colapsável por preferência persistida (só drawer no mobile).
- Kit `<.sidebar>` separado — markup vive no layout, como a nav atual.
