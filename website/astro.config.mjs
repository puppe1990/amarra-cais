import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import sitemap from "@astrojs/sitemap";

export default defineConfig({
  site: "https://puppe1990.github.io",
  base: "/amarra-cais",
  prefetch: { prefetchAll: true, defaultStrategy: "hover" },
  integrations: [
    starlight({
      title: "Amarra",
      description:
        "HTML-first Go framework for mini apps: Amarra Views + Drive, Tailwind and SQLite — with a Rails-style CLI.",
      defaultLocale: "root",
      locales: {
        root: { label: "English", lang: "en" },
        "pt-br": { label: "Português", lang: "pt-BR" },
      },
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: "https://github.com/puppe1990/amarra-cais",
        },
      ],
      editLink: {
        baseUrl: "https://github.com/puppe1990/amarra-cais/edit/main/website/",
      },
      customCss: ["./src/styles/starlight.css"],
      sidebar: [
        {
          label: "Getting started",
          translations: { "pt-BR": "Começando" },
          items: [{ autogenerate: { directory: "docs/getting-started" } }],
        },
        {
          label: "How-to guides",
          translations: { "pt-BR": "Guias práticos" },
          items: [{ autogenerate: { directory: "docs/how-to" } }],
        },
        {
          label: "Reference",
          translations: { "pt-BR": "Referência" },
          items: [{ autogenerate: { directory: "docs/reference" } }],
        },
        {
          label: "Explanation",
          translations: { "pt-BR": "Explicação" },
          items: [{ autogenerate: { directory: "docs/explanation" } }],
        },
      ],
    }),
    sitemap(),
  ],
});
