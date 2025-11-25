# Fix GoReleaser SBOM Generation

**Date**: November 25, 2025  
**Type**: Bug Fix / CI/CD  
**Impact**: High - Unblocks release workflow

## Summary

Fixed the GoReleaser workflow failure by installing the `syft` tool, which is required for generating Software Bill of Materials (SBOMs) as configured in `.goreleaser.yaml`. The release workflow was failing with the error "syft: executable file not found in $PATH" because the GitHub Actions runner didn't have the tool installed.

## Problem Statement

The GoReleaser workflow was configured to generate SBOMs for all release artifacts but was failing during the release process.

### Error Observed
```
Error: The process '/opt/hostedtoolcache/goreleaser-action/2.12.7/x64/goreleaser' 
failed with exit code 1

catalogs artifacts: syft failed: exec: "syft": executable file not found in $PATH
```

### Root Cause

1. **SBOM Configuration Present**: The `.goreleaser.yaml` file includes SBOM generation:
   ```yaml
   sboms:
     - artifacts: archive
   ```

2. **Missing Tool**: GoReleaser delegates SBOM generation to Anchore's `syft` tool, which must be available in the system PATH

3. **No Installation Step**: The GitHub Actions workflow (`.github/workflows/release.yml`) didn't include a step to install `syft` before running GoReleaser

### Impact

- All release workflows were failing at the SBOM generation step
- Unable to create new releases with tags (e.g., `v1.0.1`)
- Release artifacts (binaries, archives, checksums) could not be published to GitHub Releases
- Docker images could still be built (separate job), but the complete release process was blocked

## Solution

Added an installation step for `syft` in the release workflow using the official Anchore GitHub Action.

## Changes Made

### Updated `.github/workflows/release.yml`

Added a new step in the `goreleaser` job between "Set up Go" and "Run GoReleaser":

```yaml
- name: Install syft
  uses: anchore/sbom-action/download-syft@v0
```

### Complete Job Structure (After)

```yaml
jobs:
  goreleaser:
    name: GoReleaser
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Install syft
        uses: anchore/sbom-action/download-syft@v0

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Technical Details

### What is syft?

`syft` is a CLI tool from Anchore that generates Software Bill of Materials (SBOM) from container images and filesystems. SBOMs are important for:
- **Supply Chain Security**: Track all components and dependencies in releases
- **Vulnerability Management**: Identify known vulnerabilities in dependencies
- **Compliance**: Meet regulatory requirements for software transparency
- **Transparency**: Provide users with detailed information about software composition

### Why anchore/sbom-action/download-syft?

- **Official Action**: Maintained by Anchore, the creators of syft
- **Lightweight**: Only downloads syft binary without running a full scan
- **PATH Integration**: Automatically adds syft to the runner's PATH
- **Version Management**: Ensures compatible version with GoReleaser
- **No Configuration**: Works out-of-the-box with GoReleaser's SBOM generation

### SBOM Output

With this fix, GoReleaser will now generate SBOM files for each release archive:
- `mcp-server-planton_1.0.1_Linux_x86_64.tar.gz.sbom`
- `mcp-server-planton_1.0.1_Linux_arm64.tar.gz.sbom`
- `mcp-server-planton_1.0.1_Darwin_x86_64.tar.gz.sbom`
- `mcp-server-planton_1.0.1_Darwin_arm64.tar.gz.sbom`
- `mcp-server-planton_1.0.1_Windows_x86_64.zip.sbom`
- `mcp-server-planton_1.0.1_Windows_arm64.zip.sbom`

These SBOM files will be uploaded as release artifacts alongside the binaries and archives.

## Benefits

1. **Unblocked Releases**: Release workflow now completes successfully
2. **Enhanced Security**: SBOMs provide transparency about software components
3. **Supply Chain Visibility**: Users can inspect what's included in each release
4. **Compliance Ready**: Meets modern software distribution best practices
5. **Automated Process**: SBOMs generated automatically for every release

## Verification

To verify the fix works:

1. **Create a Test Tag**:
   ```bash
   git tag v1.0.2
   git push origin v1.0.2
   ```

2. **Monitor Workflow**: Check GitHub Actions for the release workflow
   - The "Install syft" step should complete successfully
   - GoReleaser should generate SBOMs without errors
   - Release artifacts should include `.sbom` files

3. **Check Release Assets**: On the GitHub Releases page, verify:
   - All binary archives are present
   - Each archive has a corresponding `.sbom` file
   - Checksums file is generated
   - No errors in the workflow logs

## Migration Notes

- **No Breaking Changes**: This is purely an addition to the workflow
- **No Configuration Changes**: `.goreleaser.yaml` remains unchanged
- **Backward Compatible**: Previous releases are unaffected
- **Automatic**: Next tag push will use the fixed workflow

## Rollback Plan

If issues arise, the syft installation step can be temporarily removed:

```yaml
# Comment out or remove this step
# - name: Install syft
#   uses: anchore/sbom-action/download-syft@v0
```

And disable SBOM generation in `.goreleaser.yaml`:

```yaml
# Comment out the sboms section
# sboms:
#   - artifacts: archive
```

However, this would mean releases don't include SBOMs, which is not recommended for production releases.

## References

- **Anchore SBOM Action**: https://github.com/anchore/sbom-action
- **syft Documentation**: https://github.com/anchore/syft
- **GoReleaser SBOM Docs**: https://goreleaser.com/customization/sbom/
- **SBOM Standards**: SPDX and CycloneDX formats supported

## Conclusion

This fix restores the release workflow to working order while maintaining the security and transparency benefits of SBOM generation. The solution uses the official Anchore action, which is lightweight and requires no additional configuration. All future releases will now include comprehensive SBOMs for all platform artifacts.
