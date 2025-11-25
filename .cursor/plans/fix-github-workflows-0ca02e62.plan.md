<!-- 0ca02e62-65ca-43fb-841e-3f8cd0487f11 303e5b10-fb86-47a6-b60e-a84d53d2b8d6 -->
# Fix GitHub Workflows and Repository Security

## Part 1: Branch and Tag Setup

### 1. Rename Branch from master to main

- Update local branch name
- Push to remote and set as default
- Update branch references in documentation

**Note**: The CI workflow already uses `main`, so no changes needed there!

### 2. Create v1.0.0 Tag

- Tag the current commit
- Push tag to trigger release workflow
- This will automatically build binaries and Docker images

## Part 2: Repository Security Review

### Current Security Status

✅ **Good practices in place**:

- CI workflow runs tests and linting on all PRs
- Workflow permissions properly scoped (contents: write, packages: write)
- CONTRIBUTING.md with clear guidelines
- Fork-based contribution workflow documented
- Conventional commit format guidelines

❌ **Missing security features**:

- No CODEOWNERS file (defines who reviews PRs)
- No SECURITY.md (security policy for vulnerability reporting)
- No issue/PR templates

### 3. Add CODEOWNERS File

Create `.github/CODEOWNERS`:

```
# Global owners - review all changes
* @sureshattaluri @swarupdonepudi

# Specific ownership
/internal/grpc/ @sureshattaluri @swarupdonepudi
/internal/mcp/ @sureshattaluri @swarupdonepudi
/.github/ @sureshattaluri
```

This ensures you and Swaroop are automatically assigned as reviewers on all PRs.

### 4. Add SECURITY.md

Create a security policy for vulnerability reporting:

- Define supported versions
- Provide contact information for security issues
- Set expectations for response times

### 5. GitHub Repository Settings (Manual Steps)

You'll need to configure these in GitHub Settings:

**Branch Protection Rules** (Settings → Branches → Add rule):

- Require pull request reviews before merging (at least 1 approval)
- Require status checks to pass (CI workflow)
- Require conversation resolution before merging
- Include administrators (optional - you decide)

**Collaborator Permissions** (Settings → Collaborators):

- You and Swaroop as admins (you mentioned doing this manually)
- External contributors fork and submit PRs

**Other Settings** (Settings → General):

- Disable "Allow merge commits" (optional - keep history clean)
- Enable "Automatically delete head branches"
- Enable "Allow squash merging" for cleaner history

## Implementation Order

1. Add security files (CODEOWNERS, SECURITY.md)
2. Rename branch master → main
3. Push changes
4. Create and push v1.0.0 tag
5. Manually configure GitHub settings (branch protection, etc.)
6. Verify CI runs on the new main branch
7. Verify release workflow triggers from tag

## What's Already Secure

Your workflows are well-configured:

- Using official GitHub actions
- Proper secret handling with `${{ secrets.GITHUB_TOKEN }}`
- No hardcoded credentials
- Appropriate permission scopes
- Multi-platform builds with checksums