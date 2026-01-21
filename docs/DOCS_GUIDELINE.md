# Docusaurus Setup & Deployment Guide

This guide provides a standardized approach to setting up, configuring, and deploying Docusaurus documentation sites to GitHub Pages, based on best practices and lessons learned from the OpenGIN project.

## 1. Prerequisites & Environment
*   **Node.js Version**: Docusaurus 3.x requires Node.js **>= 18.0**. We recommend enforcing **Node.js 20 (LTS)** or higher in the project to avoid version mismatches.
*   **Package Manager**: `npm` is standard, key dependencies must be locked.

### Recommended Configuration
Add or update the `engines` field in your `package.json` to enforce the Node version:

```json
"engines": {
  "node": ">=20.0"
}
```

Create a `.nvmrc` file in the root of your docs directory:
```text
20
```

## 2. Configuration (`docusaurus.config.js`)

### Essential GitHub Pages Settings
Configuring `url` and `baseUrl` correctly is critical for assets to load on GitHub Pages.

```javascript
// docusaurus.config.js
const config = {
  // ...
  url: 'https://<ORG_NAME>.github.io', // Your GitHub Pages domain
  baseUrl: '/<REPO_NAME>/',           // The name of your repository with slashes
  
  // GitHub Deployment Config
  organizationName: '<ORG_NAME>', 
  projectName: '<REPO_NAME>', 
  
  // Handling Broken Links
  onBrokenLinks: 'throw', // Recommended: Break build on broken links
  onBrokenMarkdownLinks: 'warn',

  // ...
};
```

### Dynamic Base URL (Optional but Recommended)
For PR previews or local testing where the path might differ, you can make `baseUrl` dynamic:

```javascript
baseUrl: process.env.BASE_URL || '/<REPO_NAME>/',
```

## 3. GitHub Actions Workflows

We use two primary workflows: one for **production deployment** (on push to main) and one for **PR previews**.

### A. Production Deployment (`.github/workflows/deploy-docs.yml`)

<details>
<summary>Click to see template</summary>

```yaml
name: Deploy to GitHub Pages

on:
  push:
    branches:
      - main
    paths:
      - 'docs/**' # specific directory trigger
  workflow_dispatch:

permissions:
  contents: write # REQUIRED for pushing to gh-pages branch

jobs:
  deploy:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: docs # Set if docs are in a subdir
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: npm
          cache-dependency-path: docs/package-lock.json

      - name: Install dependencies
        run: npm ci

      - name: Build website
        run: npm run build

      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./docs/build
```
</details>

### B. PR Preview (`.github/workflows/preview-docs.yml`)

<details>
<summary>Click to see template</summary>

```yaml
name: Deploy PR Preview

on:
  pull_request_target:
    types: [opened, reopened, synchronize, closed]
    paths: ['docs/**']

permissions:
  contents: write
  pull-requests: write
  issues: write

jobs:
  deploy-preview:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: docs
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }} # Checkout PR code

      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: npm
          cache-dependency-path: docs/package-lock.json

      - name: Install dependencies
        run: npm ci

      - name: Deploy preview
        uses: rossjrw/pr-preview-action@v1
        with:
          source_dir: docs/build
          preview_branch: gh-pages
          umbrella_dir: pr-preview
          action: auto
          build_script: npm run build
        env:
          # Crucial for assets to load in the preview sub-path
          BASE_URL: /<REPO_NAME>/pr-preview/${{ github.event.number }}/
```
</details>

## 4. Common Pitfalls & Solutions

### "Remote Rejected ... Repository Rule Violations"
**Cause:** The `gh-pages` branch (used for deployment) has "Branch Protection Rules" enabled that prevent the GitHub Actions bot from pushing.
**Fix:**
1.  Go to Repo Settings -> Rules -> Rulesets (or Branches).
2.  Add a bypass for the `github-actions` app on the `gh-pages` branch.
3.  Alternatively, disable "Require pull request before merging" for `gh-pages`.

### "Exit Code 1" during Build (Node Version)
**Cause:** Docusaurus 3.x failing on Node 16 or 18.
**Fix:** Ensure `actions/setup-node` uses `node-version: 20` and `package.json` engines are set.

## 5. Master Prompt for AI Agents
Use the following prompt to instruct an AI assistant to set this up for a new project.

---

**Copy & Paste this Prompt:**

> I need to set up Docusaurus documentation for this project with automatic deployment to GitHub Pages. Please follow these specific guidelines based on our organization's standards:
>
> 1.  **Project Structure**:
>     *   Initialize a new Docusaurus project in a `docs/` subdirectory if one doesn't exist.
>     *   Use `npm` for package management.
>     *   Ensure `package.json` enforces `engines: { "node": ">=20.0" }`.
>     *   Create an `.nvmrc` file with `20`.
>
> 2.  **Configuration (`docusaurus.config.js`)**:
>     *   Set `url` to `https://<ORG_NAME>.github.io`.
>     *   Set `baseUrl` to `/<REPO_NAME>/` but allow it to be overridden by `process.env.BASE_URL` (for PR previews).
>     *   Set `onBrokenLinks` to `'throw'` and `onBrokenMarkdownLinks` to `'warn'`.
>     *   Ensure `organizationName` and `projectName` are set correctly.
>
> 3.  **CI/CD Workflows (GitHub Actions)**:
>     *   Create `.github/workflows/deploy-docs.yml`:
>         *   Trigger on `push` to `main` (filtered to `docs/` path).
>         *   Use `permissions: contents: write`.
>         *   Use `actions/setup-node@v4` with version `20`.
>         *   Use `peaceiris/actions-gh-pages@v3` for deployment.
>     *   Create `.github/workflows/preview-docs.yml` (Optional):
>         *   Trigger on `pull_request` to `main`.
>         *   Use `rossjrw/pr-preview-action@v1`.
>         *   Pass the correct `BASE_URL` environment variable to the build script to ensure assets load in the preview sub-path.
>
> 4.  **Verification**:
>     *   Remind me to check "Branch Protection Rules" for the `gh-pages` branch if the push fails.
>     *   Verify that the build command (`npm run build`) succeeds locally with Node 20.
>
> Please analyze the current repository structure and generate the necessary files and changes.

---
