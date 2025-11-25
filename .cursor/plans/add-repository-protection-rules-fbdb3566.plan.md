<!-- fbdb3566-57bf-4fc4-9a3d-d6f20ebc5933 06be2cc6-378a-4959-99fa-deeb9f0f6549 -->
# Add Repository Protection Rules

## Current State Analysis

### What IS Configured ✅

- CI workflow with tests, linting, and formatting checks
- CODEOWNERS file (@sureshattaluri, @swarupdonepudi)
- Contributing guidelines with PR process
- Security policy
- Release automation

### What is MISSING ❌

1. **Branch Protection Rules** (configured in GitHub settings, not files)
2. GitHub Issue/PR templates
3. Dependabot for dependency updates
4. Security scanning (CodeQL)
5. Pre-commit hooks configuration
6. Automated stale issue/PR management

---

## Implementation Plan

### 1. GitHub Branch Protection Settings (Manual Configuration Required)

These settings must be configured in GitHub UI at:

`Settings → Branches → Branch protection rules → Add rule` for `main` branch

**Required Settings:**

```
Branch name pattern: main

☑ Require a pull request before merging
  ☑ Require approvals: 1
  ☑ Dismiss stale pull request approvals when new commits are pushed
  ☑ Require review from Code Owners

☑ Require status checks to pass before merging
  ☑ Require branches to be up to date before merging
  Required checks:
    - lint-and-test
    - golangci-lint

☑ Require conversation resolution before merging

☑ Require signed commits (recommended)

☑ Require linear history

☑ Do not allow bypassing the above settings
  - Include administrators

☑ Restrict who can push to matching branches
  - Only allow: @sureshattaluri, @swarupdonepudi (or use teams)

☑ Allow force pushes: Specify who can force push (empty = nobody)

☑ Allow deletions: ☐ (unchecked)
```

### 2. Create GitHub Issue Templates

**Files to create:**

- `.github/ISSUE_TEMPLATE/bug_report.yml` - Bug report template
- `.github/ISSUE_TEMPLATE/feature_request.yml` - Feature request template
- `.github/ISSUE_TEMPLATE/config.yml` - Template configuration

### 3. Create Pull Request Template

**File to create:**

- `.github/pull_request_template.md` - Standard PR checklist with:
        - Description
        - Type of change
        - Testing checklist
        - Documentation updates
        - Breaking changes notice

### 4. Add Dependabot Configuration

**File to create:**

- `.github/dependabot.yml` - Configure automated dependency updates for:
        - Go modules
        - GitHub Actions
        - Docker base images

### 5. Add CodeQL Security Scanning

**File to create:**

- `.github/workflows/codeql.yml` - Security scanning workflow for:
        - Go code analysis
        - Dependency vulnerability scanning
        - Runs on push/PR and scheduled weekly

### 6. Add Stale Bot Configuration

**File to create:**

- `.github/workflows/stale.yml` - Auto-label and close stale issues/PRs:
        - Mark stale after 60 days
        - Close after 7 days of no activity

### 7. Add Pre-commit Hooks Configuration (Optional)

**File to create:**

- `.pre-commit-config.yaml` - Local development hooks for:
        - go fmt
        - go vet
        - golangci-lint
        - conventional commit message validation

### 8. Update Documentation

**Files to update:**

- `CONTRIBUTING.md` - Add section on branch protection and PR requirements
- `docs/development.md` - Add pre-commit hooks setup instructions
- `README.md` - Add badges for CI status, security scanning

---

## Priority Order

1. **CRITICAL** - Configure branch protection rules in GitHub settings (manual)
2. **HIGH** - Add PR template and issue templates
3. **HIGH** - Add Dependabot and CodeQL workflows
4. **MEDIUM** - Add stale bot workflow
5. **LOW** - Add pre-commit hooks configuration

---

## Expected Outcomes

After implementation:

- ✅ Direct pushes to `main` blocked - all changes must go through PRs
- ✅ PRs require 1 approval from CODEOWNERS before merge
- ✅ CI checks (tests + lint) must pass before merge
- ✅ Standardized issue and PR templates guide contributors
- ✅ Automated security and dependency scanning
- ✅ Automated dependency update PRs from Dependabot
- ✅ Stale issues/PRs automatically managed

### To-dos

- [ ] Configure branch protection rules in GitHub UI for main branch
- [ ] Create PR template and issue templates
- [ ] Add Dependabot configuration for Go modules and GitHub Actions
- [ ] Add CodeQL security scanning workflow
- [ ] Add stale issue/PR automation workflow
- [ ] Add pre-commit hooks configuration file
- [ ] Update CONTRIBUTING.md and development docs with new requirements