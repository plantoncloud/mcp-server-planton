<!-- 53c8966d-b58a-4b17-a39a-e61e1ee69884 268cf447-c9f1-4472-8e12-9aacf45289be -->
# Implement GitHub Branch Protection Rules

## Overview

Remove the documentation file and implement actual branch protection rules using GitHub CLI/API commands.

## Current State

- Repository: `plantoncloud/mcp-server-planton`
- Branch: `main` (currently unprotected)
- Documentation file exists at `.github/BRANCH_PROTECTION_SETUP.md`

## Implementation Steps

### 1. Delete Documentation File

Remove `.github/BRANCH_PROTECTION_SETUP.md` since we'll implement the actual protection rules.

### 2. Apply Branch Protection Rules

Use the GitHub API via `gh api` to configure branch protection with these settings:

**Pull Request Requirements:**

- Require PR before merging
- Require 1 approval
- Dismiss stale reviews when new commits pushed
- Require review from Code Owners

**Status Checks:**

- Require status checks to pass
- Require branches to be up to date
- Required checks: `lint-and-test`, `golangci-lint`

**Other Protections:**

- Require conversation resolution
- Require linear history
- **Allow administrators to bypass** (admins can override and merge without approvals)
- Block force pushes
- Block branch deletion

**Fork-based Contributions:**

- External contributors automatically fork (no write access to main repo)
- This is standard GitHub behavior for public repos

### 3. Implementation Command

Will use `gh api` with a PUT request to `/repos/plantoncloud/mcp-server-planton/branches/main/protection` with a JSON payload containing all the protection rules.

## Notes

- Status checks (`lint-and-test`, `golangci-lint`) must have run at least once to be enforced
- Administrators will still be bound by these rules (include_administrators: true)
- Direct pushes to main will be blocked - all changes through PRs only