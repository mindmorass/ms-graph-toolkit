# GitHub Actions Workflows

This directory contains GitHub Actions workflows for building and releasing the components of this monorepo.

## Workflows

### `client.yml`
Builds and releases the Go client for multiple platforms.

**Triggers:**
- Changes in `client/` directory
- Manual dispatch

**Builds:**
- Linux (amd64)
- macOS (amd64, arm64)

**Releases:**
- Creates a GitHub release with all platform binaries
- Only on pushes to main/master branch

### `extension.yml`
Builds and releases the Chrome extension.

**Triggers:**
- Changes in `extension/` directory
- Manual dispatch

**Builds:**
- Packages extension as ZIP file
- Validates manifest.json and required files

**Releases:**
- Creates a GitHub release with the extension ZIP
- Only on pushes to main/master branch

### `ci.yml`
Runs CI checks for both components.

**Triggers:**
- All pushes and pull requests

**Jobs:**
- `client-test`: Runs Go tests and linters (only if client/ changed)
- `extension-validate`: Validates extension files (only if extension/ changed)

## Path-Based Triggers

All workflows use path filters to only run when relevant files change:

- `client/**` - Triggers Go client workflows
- `extension/**` - Triggers extension workflows

This ensures builds only run when necessary, saving CI/CD resources.

## Manual Triggers

All workflows support `workflow_dispatch` for manual triggering:

1. Go to Actions tab in GitHub
2. Select the workflow
3. Click "Run workflow"
4. Choose branch and click "Run workflow"

## Release Process

### Automatic Releases
- Triggered on pushes to main/master branch
- Only when files in the respective directory change
- Creates a new release with version based on date and commit SHA

### Manual Releases
- Use `workflow_dispatch` to trigger a release
- Or create a tag to trigger a release with that tag name

## Versioning

- **Automatic**: `vYYYYMMDD-COMMITSHA` (e.g., `v20240105-a1b2c3d4`)
- **Tagged**: Uses the tag name (e.g., `v1.0.0`)

Releases are tagged with prefixes:
- Go client: `client-v*`
- Extension: `extension-v*`

## Permissions

The release workflows require `contents: write` permission to create GitHub releases. This is automatically granted via the `permissions` block in each workflow.

If you encounter 403 errors when creating releases, check:

1. **Repository Settings** → **Actions** → **General** → **Workflow permissions**
   - Should be set to "Read and write permissions" (not "Read repository contents and packages permissions")

2. **Repository Settings** → **Actions** → **General** → **Allow GitHub Actions to create and approve pull requests**
   - Should be enabled if you want workflows to create releases

3. The workflows include explicit permissions:
   ```yaml
   permissions:
     contents: write
     actions: read
   ```

