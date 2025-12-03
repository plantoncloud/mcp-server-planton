## Summary

Initial release of the Planton Cloud MCP server as a standalone public repository. Extracted from the planton-cloud monorepo and restructured as a proper Python package with comprehensive documentation, public dependencies, and CI/CD infrastructure for PyPI publishing.

## Context

This MCP server was initially developed in the planton-cloud monorepo but has been extracted to follow industry patterns established by AWS and GitHub MCP servers. The extraction enables:

- **Public distribution**: Standard PyPI installation via `pip install mcp-server-planton`
- **Independent versioning**: Semantic versioning independent of monorepo releases
- **Community contributions**: External developers can contribute without monorepo access
- **Ecosystem alignment**: Discoverable in MCP server directories and marketplaces

This initial commit represents the complete extraction and restructuring work.

## Changes

**Package Structure:**
- Created `src/mcp_server_planton/` proper Python package structure
- Migrated all source code from monorepo with updated imports
- Added package initialization files with version metadata

**Dependencies:**
- Updated from monorepo path dependencies to public buf.build registry
- Using `blintora-apis-protocolbuffers-python` for protobuf stubs
- All dependencies publicly available (no private packages)

**Documentation:**
- `README.md`: Comprehensive external-facing documentation
- `docs/installation.md`: Complete installation guide for multiple methods
- `docs/configuration.md`: Detailed configuration reference with security best practices
- `docs/development.md`: Contributing and development guide
- `CONTRIBUTING.md`: External contribution guidelines
- `NEXT_STEPS.md`: Publishing instructions for maintainers

**Infrastructure:**
- `.github/workflows/ci.yml`: Continuous integration (linting, type checking)
- `.github/workflows/publish.yml`: Automated PyPI publishing on tag push
- `pyproject.toml`: Complete PyPI packaging metadata
- `LICENSE`: Apache-2.0 license
- `.gitignore`: Python standard ignores

**Code Quality:**
- Fixed type safety issues (Optional type hints)
- Passes ruff linting with no errors
- Passes mypy type checking across 9 source files
- All 43 dependencies install successfully

## Implementation notes

**Security Model Preserved:**
- User JWT authentication (not machine accounts)
- Fine-Grained Authorization (FGA) enforcement
- No JWT persistence in long-lived storage
- Same security architecture as monorepo implementation

**Available Tools:**
- `list_environments_for_org`: Query environments by organization with user permissions

**CLI Entry Point:**
- Installed as `mcp-server-planton` command
- Standard invocation pattern matching other MCP servers

**Testing:**
- Verified build with Poetry
- Quality gates: ruff (0 errors), mypy (0 errors)
- Ready for PyPI publication as v0.1.0

## Breaking changes

None. This is the initial public release with no prior versions to maintain compatibility with.

## Test plan

- ✅ Package structure verified with Poetry install
- ✅ Ruff linting passes (0 errors, 0 warnings)
- ✅ MyPy type checking passes (9 source files)
- ✅ All imports resolve correctly with buf.build stubs
- ✅ Documentation reviewed and links verified
- ✅ CI workflows syntax validated
- ✅ CLI entry point tested

## Risks

**Low risk** - This is a new public repository with no production dependencies yet. The code is battle-tested from the monorepo implementation.

**Publishing readiness**: Repository is ready for PyPI publication once PYPI_TOKEN is configured in GitHub secrets.

**Integration**: When graph-fleet or other consumers are ready to integrate, they can install via:
```bash
pip install mcp-server-planton
```

Or for pre-release testing:
```toml
mcp-server-planton = { git = "https://github.com/plantoncloud-inc/mcp-server-planton.git", branch = "master" }
```

## Checklist

- [x] Docs updated (comprehensive documentation created)
- [x] Tests added/updated (quality checks configured and passing)
- [x] Backward compatible (N/A - initial release)
- [x] License added (Apache-2.0)
- [x] Contributing guidelines added
- [x] CI/CD pipelines configured
- [x] PyPI packaging metadata complete
- [x] README covers installation, configuration, usage



















