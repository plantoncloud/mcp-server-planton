<!-- 97b14612-2fc4-437f-b22c-fcfcec992fe2 c6598426-6f18-493f-a9a0-6d9a7933653e -->
# Add GoFormat Commands to Makefile

## Current State

- CI workflow (`.github/workflows/ci.yml`) already has a GoFormat check that fails builds if code is not formatted
- Makefile has `build`, `test`, and `lint` commands but no format commands
- Developers need to manually run `gofmt` to fix formatting issues

## Proposed Solution

### 1. Add Format Commands to Makefile

Add two new make targets after the existing `lint` target:

**`make fmt`** - Automatically format all Go code:

```makefile
## fmt: Format Go code
fmt:
	@echo "Formatting Go code..."
	@gofmt -w .
	@echo "Code formatted"
```

**`make fmt-check`** - Check if code is formatted (same logic as CI):

```makefile
## fmt-check: Check if Go code is formatted
fmt-check:
	@echo "Checking Go code formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "Go code is not formatted:"; \
		gofmt -l .; \
		echo "Run 'make fmt' to fix formatting"; \
		exit 1; \
	fi
	@echo "All Go code is properly formatted"
```

### 2. Integrate Format Check into Build

Update the `build` target to run format check first:

```makefile
## build: Build the binary for local architecture
build: fmt-check
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@go build -o $(BINARY_PATH) ./cmd/mcp-server-planton
	@echo "Binary built: $(BINARY_PATH)"
```

This ensures that `make build` will fail if code is not formatted, catching issues before pushing.

### 3. Update .PHONY Declaration

Add the new targets to the .PHONY declaration at the top:

```makefile
.PHONY: build install test lint fmt fmt-check docker-build docker-run clean help
```

## Benefits

1. **Developer Experience**: Run `make fmt` to quickly fix all formatting issues
2. **Early Detection**: `make build` catches formatting issues locally before pushing
3. **Consistency**: Same format check logic in Makefile and CI workflow
4. **Simple Workflow**: 

   - Before committing: run `make build` (which checks formatting)
   - If build fails due to formatting: run `make fmt` to fix
   - Re-run `make build` to verify

### 4. Add Release Command

Add a new `release` target that creates and pushes a git tag:

```makefile
## release: Create and push a release tag (usage: make release TAG=v1.0.0)
release:
ifndef TAG
	@echo "Error: TAG is required. Usage: make release TAG=v1.0.0"
	@exit 1
endif
	@echo "Creating release tag $(TAG)..."
	@if ! echo "$(TAG)" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+'; then \
		echo "Error: TAG must follow semantic versioning (e.g., v1.0.0, v2.1.3)"; \
		exit 1; \
	fi
	@git tag -a $(TAG) -m "Release $(TAG)"
	@git push origin $(TAG)
	@echo "Release tag $(TAG) created and pushed"
	@echo "GitHub Actions will now build and publish the release"
```

This command will:

- Require a TAG parameter (e.g., `make release TAG=v1.0.0`)
- Validate that the tag follows semantic versioning pattern (v1.0.0)
- Create an annotated git tag
- Push the tag to origin, which triggers the GitHub release workflow
- The release workflow will then run GoReleaser and build Docker images

**Usage Example:**

```bash
make release TAG=v1.0.0
```

## Files to Modify

- `Makefile` - Add `fmt`, `fmt-check`, and `release` targets, update `build` target, update `.PHONY`

### To-dos

- [ ] Update .PHONY declaration to include fmt and fmt-check targets
- [ ] Add 'make fmt' command to automatically format Go code
- [ ] Add 'make fmt-check' command to verify code formatting
- [ ] Update 'make build' to depend on fmt-check