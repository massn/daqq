# daqq documentation site

Built with [Hugo](https://gohugo.io/) (extended) and the [Hextra](https://imfing.github.io/hextra/) theme, deployed to **Cloudflare Pages**.

## Local development

Prerequisites: Hugo extended (`brew install hugo`) and Go (already required by the chain itself).

```bash
cd docs
hugo server                  # http://localhost:1313
hugo --minify                # build into ./public
```

The theme is pulled in via **Hugo Modules** (see `go.mod`). To update it:

```bash
hugo mod get -u github.com/imfing/hextra
hugo mod tidy
```

## Authoring

- Content lives under `content/`. Each Markdown file's front-matter controls title, weight (ordering), etc.
- Mermaid diagrams: write a ```` ```mermaid ```` fenced block — Hextra renders them automatically with dark-mode support.
- Math: KaTeX is enabled. Inline `\( ... \)`, block `$$ ... $$`.
- Callouts: use the Hextra shortcode `{{< callout type="info|warning" >}} ... {{< /callout >}}`.

## Cloudflare Pages deployment (Git integration)

1. In the Cloudflare dashboard: **Workers & Pages → Create → Pages → Connect to Git** → pick this repository.
2. Set the project settings:

   | Field | Value |
   |---|---|
   | Production branch | `main` |
   | Framework preset | `Hugo` |
   | Build command | `hugo --minify` |
   | Build output directory | `public` |
   | Root directory | `docs` |
   | Environment variable: `HUGO_VERSION` | `0.160.1` (or whatever matches local) |
   | Environment variable: `HUGO_ENV` | `production` |

3. Cloudflare will install Hugo and Go automatically; Hugo Modules will resolve Hextra at build time.
4. The first build publishes to `https://<project>.pages.dev`. Custom domains can be attached later under **Custom domains** in the project settings.

Preview deployments are automatically created for non-production branches and PRs.
