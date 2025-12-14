# Add Repository Protection Rules and Automation

**Date**: 2025-11-25  
**Type**: Infrastructure / Security Enhancement  
**Impact**: High - Establishes repository governance and security standards

## Problem

The repository lacked standard protection mechanisms that are essential for public open-source projects:

1. **No Branch Protection**: Direct pushes to `main` branch were allowed without review or CI checks
2. **No PR/Issue Templates**: Inconsistent contribution formats and missing information
3. **No Dependency Management**: No automated dependency updates or vulnerability scanning
4. **No Security Scanning**: No automated code security analysis
5. **No Stale Management**: Old issues and PRs accumulating without cleanup
6. **No Pre-commit Hooks**: No local development quality gates
7. **Limited Documentation**: Missing details on contribution requirements and CI checks

This meant that changes could be merged directly to `main` without any review, testing, or validation, which is not acceptable for a public repository.

## Solution

Implemented comprehensive repository protection rules, automation workflows, and developer tooling to match industry-standard practices for public open-source repositories.

## Changes Made

### 1. Branch Protection Setup Instructions

**Created**: `.github/BRANCH_PROTECTION_SETUP.md`

Detailed documentation for configuring GitHub branch protection rules (must be done manually in GitHub UI):

- **Pull Request Requirements**:
  - Require PR before merging
  - Require 1 approval from CODEOWNERS
  - Dismiss stale approvals on new commits
  - Require review from Code Owners

- **Status Check Requirements**:
  - Require all CI checks to pass (`lint-and-test`, `golangci-lint`)
  - Require branches to be up to date before merging

- **Additional Protections**:
  - Require conversation resolution before merging
  - Require signed commits (recommended)
  - Require linear history (no merge commits)
  - Include administrators in enforcement
  - Restrict push access to specific users
  - Disallow force pushes
  - Disallow branch deletion

### 2. GitHub Issue Templates

**Created**:
- `.github/ISSUE_TEMPLATE/bug_report.yml` - Structured bug report form with fields for:
  - Bug description
  - Steps to reproduce
  - Expected vs actual behavior
  - Environment details
  - Logs and configuration
  
- `.github/ISSUE_TEMPLATE/feature_request.yml` - Feature request form with fields for:
  - Problem statement
  - Proposed solution
  - Alternatives considered
  - Use case description
  - Implementation willingness
  
- `.github/ISSUE_TEMPLATE/config.yml` - Template configuration with:
  - Link to GitHub Discussions for questions
  - Link to security advisory page for vulnerabilities
  - Link to documentation

**Benefits**:
- Standardized issue format ensures all necessary information is collected
- Guides contributors to provide complete context
- Separates security issues from public bug reports
- Directs questions to appropriate channels

### 3. Pull Request Template

**Created**: `.github/pull_request_template.md`

Comprehensive PR checklist covering:
- **Description**: Clear explanation of changes
- **Type of Change**: Bug fix, feature, breaking change, etc.
- **Related Issues**: Links to issues being addressed
- **Testing**: Test evidence and manual testing confirmation
- **Documentation**: README, docs, comments, changelog updates
- **Code Quality**: Style compliance, self-review, vet/lint checks
- **Breaking Changes**: Migration guide if applicable
- **Additional Context**: Screenshots, reviewer notes

**Benefits**:
- Ensures PRs are complete before review
- Reminds contributors to update documentation
- Provides reviewers with necessary context
- Enforces quality standards

### 4. Dependabot Configuration

**Created**: `.github/dependabot.yml`

Automated dependency updates for:
- **Go Modules**: Weekly updates every Monday at 9:00 AM
- **GitHub Actions**: Weekly updates every Monday at 9:00 AM
- **Docker Base Images**: Weekly updates every Monday at 9:00 AM

Configuration includes:
- Automatic PR creation with descriptive commit messages
- Proper labeling (`dependencies`, `go`, `github-actions`, `docker`)
- Automatic reviewer and assignee assignment
- Configurable PR limits to avoid overwhelming maintainers

**Benefits**:
- Automated security updates
- Keep dependencies current
- Reduces manual dependency maintenance
- Automatic vulnerability patching

### 5. CodeQL Security Scanning

**Created**: `.github/workflows/codeql.yml`

Automated security scanning workflow:
- **Triggers**:
  - Push to `main` and `develop` branches
  - Pull requests to `main` and `develop`
  - Weekly schedule (Mondays at 6:00 AM UTC)

- **CodeQL Analysis**:
  - Go code security analysis
  - Security and quality queries
  - Automatic vulnerability detection

- **Dependency Review** (PRs only):
  - Checks for vulnerable dependencies
  - Fails on moderate+ severity issues
  - Blocks problematic licenses (GPL-2.0, GPL-3.0)

**Benefits**:
- Automated security vulnerability detection
- Prevents introduction of vulnerable dependencies
- License compliance enforcement
- Continuous security monitoring

### 6. Stale Issue/PR Management

**Created**: `.github/workflows/stale.yml`

Automated stale content management:
- **Issues**:
  - Marked stale after 60 days of inactivity
  - Closed after 7 days if still stale
  - Exemptions: `pinned`, `security`, `critical`, `roadmap` labels

- **Pull Requests**:
  - Marked stale after 30 days of inactivity
  - Closed after 7 days if still stale
  - Exemptions: `pinned`, `security`, `critical`, `in-progress`, `blocked` labels
  - Draft PRs automatically exempt

- **Features**:
  - Friendly notification messages
  - Automatic stale label removal on activity
  - Configurable operation limits
  - Daily execution schedule

**Benefits**:
- Keeps issue/PR list clean and manageable
- Encourages contributors to update stale work
- Reduces maintainer burden
- Provides clear communication about inactivity

### 7. Pre-commit Hooks Configuration

**Created**: `.pre-commit-config.yaml`

Local development hooks for:
- **File Quality Checks**:
  - Trailing whitespace removal
  - End-of-file fixer
  - YAML validation
  - Large file detection
  - Merge conflict detection

- **Go-Specific Checks**:
  - `go fmt` - Code formatting
  - `go vet` - Static analysis
  - `go test` - Test execution
  - `golangci-lint` - Advanced linting (optional)
  - `go mod tidy` - Dependency cleanup

- **Commit Standards**:
  - Conventional commit message validation
  - Enforces commit message format

**Benefits**:
- Catches issues before commit
- Enforces code quality locally
- Reduces CI failures
- Standardizes commit messages
- Faster feedback loop for developers

### 8. Documentation Updates

**Modified**: `docs/development.md`

Added comprehensive pre-commit hooks documentation:
- Installation instructions
- Hook descriptions
- Manual execution examples
- Skip instructions (for emergency cases)
- Integration with development workflow

**Modified**: `CONTRIBUTING.md`

Added comprehensive contribution requirements:
- **Branch Protection and Requirements** section
  - Detailed explanation of protection rules
  - Signed commits recommendation
  
- **Enhanced Pull Request Process**
  - Expanded from 7 to 12 steps
  - Added pre-commit hooks setup (optional)
  - Added CI check requirements
  - Added review process details
  - Added merge strategy guidance

- **Required CI Checks** section
  - Listed all mandatory checks
  - Explained what each check does
  - Clear guidance on failure handling

**Modified**: `README.md`

Added status badges for:
- CI build status
- CodeQL security scanning status
- Go Report Card score
- License type
- Docker image availability

**Benefits**:
- Clear expectations for contributors
- Better onboarding for new developers
- Transparency about project status
- Professional presentation

## Files Created

### GitHub Configuration (10 files)
1. `.github/BRANCH_PROTECTION_SETUP.md`
2. `.github/ISSUE_TEMPLATE/bug_report.yml`
3. `.github/ISSUE_TEMPLATE/feature_request.yml`
4. `.github/ISSUE_TEMPLATE/config.yml`
5. `.github/pull_request_template.md`
6. `.github/dependabot.yml`
7. `.github/workflows/codeql.yml`
8. `.github/workflows/stale.yml`

### Development Tools (1 file)
9. `.pre-commit-config.yaml`

### Changelog (1 file)
10. `_changelog/2025-11/2025-11-25-131500-add-repository-protection-rules.md`

## Files Modified

### Documentation (3 files)
1. `README.md` - Added status badges
2. `CONTRIBUTING.md` - Added branch protection and PR requirements
3. `docs/development.md` - Added pre-commit hooks documentation

## Technical Details

### CI/CD Integration

The new workflows integrate with existing CI:
- **Existing**: `ci.yml` (tests, linting, formatting)
- **New**: `codeql.yml` (security scanning)
- **New**: `stale.yml` (maintenance automation)
- **Existing**: `release.yml` (release automation) - unchanged

### Automated Checks Flow

```
┌─────────────────┐
│   Developer     │
│  Makes Changes  │
└────────┬────────┘
         │
         ├─── (Optional) Pre-commit hooks run locally
         │
         ▼
┌─────────────────┐
│  Push to Fork   │
│  Create PR      │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────┐
│     GitHub Automated Checks         │
├─────────────────────────────────────┤
│ 1. lint-and-test (ci.yml)           │
│ 2. golangci-lint (ci.yml)           │
│ 3. CodeQL (codeql.yml)              │
│ 4. Dependency Review (codeql.yml)   │
└────────┬────────────────────────────┘
         │
         ├─── All checks must pass
         │
         ▼
┌─────────────────┐
│  Code Review    │
│  by CODEOWNERS  │
└────────┬────────┘
         │
         ├─── Approval required
         │
         ▼
┌─────────────────┐
│  Merge to Main  │
└─────────────────┘
```

### Security Enhancements

1. **Vulnerability Detection**: CodeQL scans for security issues
2. **Dependency Security**: Dependabot updates vulnerable dependencies
3. **License Compliance**: Blocks GPL licenses in dependencies
4. **Code Review**: Mandatory review from CODEOWNERS
5. **Audit Trail**: All changes tracked through PRs

### Developer Experience Improvements

1. **Clear Guidelines**: Templates guide proper contribution format
2. **Fast Feedback**: Pre-commit hooks catch issues early
3. **Automated Maintenance**: Dependabot keeps dependencies current
4. **Quality Assurance**: Multiple layers of automated checks
5. **Documentation**: Clear instructions for all processes

## Manual Steps Required

### Critical: Branch Protection Configuration

The branch protection rules must be configured manually in GitHub UI:

1. Navigate to: `https://github.com/plantoncloud/mcp-server-planton/settings/branches`
2. Click **"Add rule"** or **"Add branch protection rule"**
3. Follow all steps in `.github/BRANCH_PROTECTION_SETUP.md`

**Important Notes**:
- Status checks (`lint-and-test`, `golangci-lint`) will only appear after they run at least once
- You may need to update the protection rules after the first CI run
- Until this is configured, direct pushes to `main` are still possible

### Optional: Pre-commit Hooks Setup

Contributors can optionally install pre-commit hooks locally:

```bash
pip install pre-commit
pre-commit install
pre-commit install --hook-type commit-msg
```

## Impact

### Security
✅ Automated security scanning (CodeQL)  
✅ Automated vulnerability detection in dependencies  
✅ License compliance enforcement  
✅ Reduced attack surface through code review  

### Code Quality
✅ Mandatory CI checks before merge  
✅ Consistent code formatting  
✅ Pre-commit hooks for local validation  
✅ Linting enforcement  

### Process
✅ Branch protection prevents direct pushes to main  
✅ Required PR reviews from CODEOWNERS  
✅ Standardized issue and PR formats  
✅ Automated stale content management  

### Maintenance
✅ Automated dependency updates  
✅ Reduced manual dependency tracking  
✅ Clear contribution guidelines  
✅ Self-service documentation  

### Visibility
✅ Status badges show project health  
✅ Clear CI/CD status  
✅ Transparent security scanning results  

## Verification

After branch protection is configured, verify:

1. **Branch Protection**:
   ```bash
   # This should fail
   git push origin main
   # Error: Protected branch update failed
   ```

2. **PR Requirements**:
   - Create test PR
   - Verify CI checks run automatically
   - Verify approval is required before merge
   - Verify "Merge" button is disabled until checks pass

3. **Issue Templates**:
   - Create new issue
   - Verify templates appear in dropdown
   - Verify form validation works

4. **Dependabot**:
   - Wait for weekly run (or manually trigger)
   - Verify dependency update PRs are created

5. **CodeQL**:
   - Check "Security" tab in GitHub
   - Verify CodeQL analysis runs on schedule

## Best Practices Implemented

1. **Defense in Depth**: Multiple layers of protection
2. **Shift Left**: Pre-commit hooks catch issues early
3. **Automation**: Reduce manual toil and human error
4. **Transparency**: Clear documentation and visible status
5. **Developer Experience**: Smooth contribution process
6. **Security First**: Automated security scanning and updates
7. **Maintainability**: Automated cleanup and dependency updates
8. **Standards Compliance**: Industry-standard practices

## References

- [GitHub Branch Protection Documentation](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches)
- [Dependabot Documentation](https://docs.github.com/en/code-security/dependabot)
- [CodeQL Documentation](https://codeql.github.com/docs/)
- [Pre-commit Framework](https://pre-commit.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)

## Notes

- All automation is configured but branch protection requires manual setup in GitHub UI
- Pre-commit hooks are optional but recommended for contributors
- Status badges will show actual status once CI runs complete
- Dependabot PRs will start appearing weekly after configuration
- CodeQL results will appear in the "Security" tab

## Success Criteria

✅ Issue and PR templates created  
✅ Dependabot configuration in place  
✅ CodeQL security scanning configured  
✅ Stale bot automation configured  
✅ Pre-commit hooks configuration added  
✅ Documentation updated with new requirements  
✅ Branch protection instructions provided  
⏳ Manual branch protection configuration (pending)

Once branch protection is configured, the repository will meet industry standards for public open-source projects.
