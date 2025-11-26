<!-- 65a46357-9b83-427f-9906-76d5e016da28 eccac136-ad74-4b35-bde7-b9beea426fd0 -->
# Fix Release Process

## Problem Analysis

1. **Tag Conflict Issue**: The tag `v1.0.6` exists both locally and remotely (remote wasn't actually deleted as expected)
2. **No Build Validation**: Current `make release` doesn't check if code builds or is properly formatted, causing format issues in releases

## Solution

### 1. Fix Current Tag Conflict (v1.0.6)

Delete the existing tag both locally and remotely:

```bash
git tag -d v1.0.6                    # Delete local tag
git push origin :refs/tags/v1.0.6    # Delete remote tag
```

### 2. Update Makefile Release Target

Modify the `release` target in `Makefile` (lines 83-97) to:

**Add build validation as prerequisite:**

- Change line 84: `release:` → `release: build`
- This automatically runs `fmt-check` (line 15) before building
- Build will fail if formatting is incorrect or code doesn't compile

**Add force-delete option for tag conflicts:**

- Add a `force` parameter to handle existing tags
- If tag exists and `force=true`, delete it locally and remotely before creating new one
- If tag exists and `force=false`, show error with helpful message

### 3. Enhanced Release Target Features

The updated release target will:

1. Run `fmt-check` to validate Go code formatting
2. Run `build` to ensure code compiles successfully  
3. Check if tag already exists
4. Handle existing tags based on `force` parameter
5. Create and push the new tag only if all checks pass

### Usage Examples

Normal release (requires clean tag):

```bash
make release version=v1.0.7
```

Force release (deletes existing tag if present):

```bash
make release version=v1.0.6 force=true
```

## Files to Modify

- `Makefile` (lines 83-97): Update release target with build dependency and tag conflict handling

### To-dos

- [ ] Delete the existing v1.0.6 tag locally and remotely to resolve current conflict
- [ ] Update Makefile release target to include build dependency and tag conflict handling with force option
- [ ] Test the updated release process to ensure it properly validates formatting and build before creating tags