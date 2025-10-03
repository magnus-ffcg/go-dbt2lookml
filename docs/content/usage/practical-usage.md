---
title: Practical Usage
weight: 15
---

# Practical Usage

go-dbt2lookml is a standalone binary that can be integrated into your existing dbt workflow wherever it runs.

## Overview

Since go-dbt2lookml is just a binary with no runtime dependencies, you can use it:

- **Locally** on your development machine
- **In CI/CD pipelines** (GitHub Actions, GitLab CI, etc.)
- **In Docker containers**
- **On any server** where dbt runs

## Integration with dbt Workflow

In your current dbt workflow you just need to run "dbt docs generate" or other commands that generate manifest.json and catalog.json before running dbt2lookml command.

You do not need to run "dbt compile" before "dbt docs generate" if using a separate workflow (we dont dont use that extra information such as raw sql)

## Output from dbt2lookml to Looker

Once dbt2lookml has generated lookml views, you need make it available to looker. Easiest is using a git-repo where you just commit the files to and use "imported_project" feature in Looker, either a remote or local project. 

See: [Looker Docs](https://cloud.google.com/looker/docs/importing-projects#viewing_files_from_an_imported_project)

You could also add files through looker api, but its likely a slow process if you have many views.



