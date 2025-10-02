# Documentation

This directory contains the documentation for go-dbt2lookml, built with [Hugo](https://gohugo.io/) using the [hugo-book](https://github.com/alex-shpak/hugo-book) theme.

## 🚀 Quick Start

### Prerequisites

- [Hugo Extended](https://gohugo.io/installation/) v0.112+ installed

### Local Development

```bash
# Serve documentation locally at http://localhost:1313
make docs-serve

# Or directly with Hugo
cd docs && hugo server -D
```

The site will auto-reload when you edit documentation files.

### Build Documentation

```bash
# Build static site to public/ directory
make docs-build

# Or directly
cd docs && hugo --minify
```

## 📁 Structure

```
docs/
├── hugo.yaml              # Hugo configuration
├── content/               # Documentation content
│   ├── _index.md         # Home page
│   ├── getting-started.md
│   ├── configuration.md
│   ├── cli-reference.md
│   ├── error-handling.md
│   ├── api/              # Auto-generated API docs (gitignored)
│   │   ├── models.md
│   │   ├── parsers.md
│   │   ├── generators.md
│   │   └── ...
│   └── development/      # Development guides
│       ├── contributing.md
│       └── testing.md
├── themes/
│   └── hugo-book/        # Hugo Book theme (submodule)
└── public/               # Generated site (ignored by git)
```

## ✍️ Writing Documentation

### Page Front Matter

All content pages should have front matter:

```markdown
---
title: Page Title
weight: 10
bookToc: true
---

# Page Content
```

### Hugo Shortcodes

**Tabs:**
```markdown
{{< tabs "unique-id" >}}
{{< tab "Tab 1" >}}
Content for tab 1
{{< /tab >}}
{{< tab "Tab 2" >}}
Content for tab 2
{{< /tab >}}
{{< /tabs >}}
```

**Hints (Callouts):**
```markdown
{{< hint info >}}
**Info**  
This is an info box.
{{< /hint >}}

{{< hint warning >}}
**Warning**  
This is a warning.
{{< /hint >}}

{{< hint danger >}}
**Danger**  
This is dangerous!
{{< /hint >}}
```

**Buttons:**
```markdown
{{< button href="https://example.com" >}}Click Me{{< /button >}}
```

### Adding New Pages

1. Create a new `.md` file in `content/`
2. Add front matter with title and weight
3. Hugo will automatically add it to the menu

## 📚 API Documentation

Generate API documentation from Go source code:

```bash
# Generate API docs from pkg/ packages
make docs-api

# Generate API docs + build Hugo site
make docs-full
```

This reads your godoc comments and creates markdown files in `content/api/`.

**Packages documented:**
- `pkg/models` - Core data models
- `pkg/parsers` - Parsing logic
- `pkg/generators` - LookML generation
- `pkg/enums` - Enumerations
- `pkg/utils` - Utility functions

## 🚀 Deployment

Documentation is automatically deployed to GitHub Pages when changes are pushed to `main`.

The workflow:
1. GitHub Actions generates API docs from Go code
2. Builds the site with Hugo
3. Uploads artifact to GitHub Pages
4. Site is available at: https://magnus-ffcg.github.io/go-dbt2lookml/

## 📝 Configuration

All configuration is in `hugo.yaml`.

Key settings:
- `baseURL` - Site URL
- `theme` - Theme name (hugo-book)
- `params` - Theme-specific parameters
- `menu` - Additional menu items

## 🎨 Theme Features

Hugo Book theme includes:
- ✅ Light/dark mode toggle
- ✅ Search functionality
- ✅ Table of contents
- ✅ Mobile responsive
- ✅ Code syntax highlighting
- ✅ Git integration (edit links)

## 🛠️ Make Commands

```bash
make docs-serve    # Serve locally with live reload
make docs-build    # Build static site
make docs-clean    # Clean build artifacts
```

## 🔗 Resources

- [Hugo Documentation](https://gohugo.io/documentation/)
- [Hugo Book Theme](https://github.com/alex-shpak/hugo-book)
- [Hugo Shortcodes](https://gohugo.io/content-management/shortcodes/)
