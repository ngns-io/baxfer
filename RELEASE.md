# Release Process

This document describes how to create and deploy releases for Baxfer.

## Version Numbering

Baxfer follows [Semantic Versioning](https://semver.org/):

- **MAJOR.MINOR.PATCH** (e.g., `v1.2.3`)
- **Pre-release**: append `-alpha.N`, `-beta.N`, or `-rc.N` (e.g., `v1.0.0-alpha.1`)

### Version Guidelines

| Change Type | Version Bump | Example |
|-------------|--------------|---------|
| Breaking API/CLI changes | MAJOR | v1.0.0 → v2.0.0 |
| New features (backward compatible) | MINOR | v1.0.0 → v1.1.0 |
| Bug fixes | PATCH | v1.0.0 → v1.0.1 |
| Pre-release testing | Pre-release suffix | v1.0.0-alpha.1 |

### Pre-release Stages

1. **Alpha** (`-alpha.N`): Early testing, may have bugs, API may change
2. **Beta** (`-beta.N`): Feature complete, testing for stability
3. **Release Candidate** (`-rc.N`): Final testing before stable release
4. **Stable**: No suffix, production-ready

## CI/CD Workflows

### Continuous Integration (`ci.yml`)

Runs on every push to `main` and on pull requests:
- **Lint**: Runs golangci-lint for static analysis
- **Test**: Runs all tests with race detection and coverage
- **Build**: Verifies the project builds successfully

### Release (`release.yml`)

Triggered automatically when a version tag is pushed:
- Runs tests
- Builds binaries for all platforms (darwin/linux/windows × amd64/arm64)
- Generates SHA256 checksums
- Creates GitHub Release with release notes

## Creating a Release

### Prerequisites

1. All tests passing on `main` branch
2. All desired changes merged to `main`
3. `git` configured with appropriate permissions

### Step-by-Step Instructions

#### 1. Ensure main is up to date

```bash
git checkout main
git pull origin main
```

#### 2. Verify tests pass locally

```bash
go test -race ./...
```

#### 3. Determine the version number

Check the current version:
```bash
git tag -l --sort=-v:refname | head -5
```

Decide on the next version based on the changes:
- Bug fixes only → increment PATCH (v0.1.0 → v0.1.1)
- New features → increment MINOR (v0.1.0 → v0.2.0)
- Breaking changes → increment MAJOR (v0.1.0 → v1.0.0)
- Pre-release → add/increment suffix (v0.1.0-alpha.5 → v0.1.0-alpha.6)

#### 4. Create and push the tag

```bash
# Set the version (adjust as needed)
VERSION="v0.1.0-alpha.6"

# Create an annotated tag
git tag -a "$VERSION" -m "Release $VERSION"

# Push the tag to trigger the release workflow
git push origin "$VERSION"
```

#### 5. Monitor the release workflow

```bash
# Watch the workflow progress
gh run watch

# Or view in browser
gh run list --workflow=release.yml
```

#### 6. Verify the release

```bash
# List releases
gh release list

# View the specific release
gh release view "$VERSION"
```

The release should include:
- `baxfer_darwin_amd64` (macOS Intel)
- `baxfer_darwin_arm64` (macOS Apple Silicon)
- `baxfer_linux_amd64` (Linux x86_64)
- `baxfer_linux_arm64` (Linux ARM64)
- `baxfer_windows_amd64.exe` (Windows)
- `checksums_sha256.txt` (SHA256 checksums)

### Quick Release (One-liner)

For experienced maintainers:

```bash
VERSION="v0.1.0-alpha.6" && git tag -a "$VERSION" -m "Release $VERSION" && git push origin "$VERSION"
```

## Graduating from Pre-release to Stable

When ready to release v1.0.0 (stable):

1. **Create release candidate(s)**:
   ```bash
   git tag -a "v1.0.0-rc.1" -m "Release v1.0.0-rc.1"
   git push origin "v1.0.0-rc.1"
   ```

2. **Test thoroughly** on all platforms

3. **Fix any issues** found, create new RC if needed:
   ```bash
   git tag -a "v1.0.0-rc.2" -m "Release v1.0.0-rc.2"
   git push origin "v1.0.0-rc.2"
   ```

4. **Create stable release**:
   ```bash
   git tag -a "v1.0.0" -m "Release v1.0.0"
   git push origin "v1.0.0"
   ```

5. **Update README** if needed (remove/update "development" warnings)

## Fixing a Failed Release

If the release workflow fails:

### Delete the tag and retry

```bash
# Delete remote tag
git push --delete origin "$VERSION"

# Delete local tag
git tag -d "$VERSION"

# Fix the issue, then re-tag
git tag -a "$VERSION" -m "Release $VERSION"
git push origin "$VERSION"
```

### Delete a release (if created)

```bash
gh release delete "$VERSION" --yes
```

## Manual Release (Emergency)

If automated release fails and you need to release manually:

```bash
VERSION="v0.1.0"

# Build all binaries
GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags="-s -w -X github.com/ngns-io/baxfer/internal/cli.version=$VERSION" -o dist/baxfer_darwin_amd64 ./cmd/baxfer
GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w -X github.com/ngns-io/baxfer/internal/cli.version=$VERSION" -o dist/baxfer_darwin_arm64 ./cmd/baxfer
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w -X github.com/ngns-io/baxfer/internal/cli.version=$VERSION" -o dist/baxfer_linux_amd64 ./cmd/baxfer
GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w -X github.com/ngns-io/baxfer/internal/cli.version=$VERSION" -o dist/baxfer_linux_arm64 ./cmd/baxfer
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w -X github.com/ngns-io/baxfer/internal/cli.version=$VERSION" -o dist/baxfer_windows_amd64.exe ./cmd/baxfer

# Generate checksums
cd dist && sha256sum baxfer_* > checksums_sha256.txt && cd ..

# Create release
gh release create "$VERSION" dist/* --title "$VERSION" --notes "Release $VERSION"
```

## Post-Release Tasks

After a successful release:

1. **Announce the release** (if applicable)
2. **Update documentation** if there are new features
3. **Close related GitHub issues/milestones**
4. **Monitor for user-reported issues**

## Troubleshooting

### "tag already exists"

The tag exists locally or remotely. Delete it first:
```bash
git tag -d "$VERSION"
git push --delete origin "$VERSION"
```

### Workflow not triggering

Verify the tag format starts with `v`:
```bash
git tag -l | grep "^v"
```

### Wrong version in binary

Check the ldflags path matches the code:
```bash
grep "var version" internal/cli/cli.go
# Should show: var version = "dev"
```

The build uses `-X github.com/ngns-io/baxfer/internal/cli.version=$VERSION`

### Missing platform binary

Check the build matrix in `.github/workflows/release.yml` includes the platform.
