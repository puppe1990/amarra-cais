# Amarra docs site

Starlight (Astro) site published at <https://puppe1990.github.io/amarra-cais/>.

## Commands

```bash
npm install
npm run dev      # http://localhost:4321/amarra-cais/
npm run check    # astro check (content collections + types)
npm run build    # static build into dist/
npm run preview
```

From the repo root: `make docs` (install + dev) and `make docs-build` (ci + build).

## Content

English (root locale) lives in `src/content/docs/`, pt-BR in `src/content/docs/pt-br/` — the same filenames in both trees, so Starlight pairs them as translations. Sidebar order comes from the `sidebar.order` frontmatter; the groups themselves are `autogenerate` entries in `astro.config.mjs`.

## Internal links must include the base path

The site is served from `/amarra-cais/` (GitHub Pages project site). Astro does **not** rewrite relative links, and root-relative links get no base — both are emitted as written and would 404 in production. Always write:

```md
[CLI reference](/amarra-cais/docs/reference/cli/) <!-- EN -->
[Referência da CLI](/amarra-cais/pt-br/docs/reference/cli/) <!-- pt-BR -->
```

A link like `../reference/cli/` or `/docs/reference/cli/` looks fine locally but breaks on Pages. Hero `link:` values in frontmatter follow the same rule.

If the site ever moves to a custom domain (base becomes `/`), drop the `/amarra-cais` prefix from content links and update `site`/`base` in `astro.config.mjs`.

## Deploy

`.github/workflows/deploy-docs.yml` builds and publishes `dist/` to GitHub Pages on pushes to `main` that touch `website/` (or by manual dispatch). Enable Settings → Pages → Source: GitHub Actions once in the repository.
