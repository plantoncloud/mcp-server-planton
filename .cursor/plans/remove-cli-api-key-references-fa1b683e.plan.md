<!-- fa1b683e-9bcd-47f1-b889-d7645d4e0c56 4fb5bab2-4317-482e-a6ad-82d358b64a60 -->
# Remove Non-Working CLI Commands for API Key Retrieval

## Overview

The documentation currently includes instructions for obtaining API keys via CLI commands (`planton auth login` and `planton auth token`), but these commands don't work. We need to remove these references and keep only the Web Console method.

## Files to Modify

### 1. `README.md` (lines 174-181)

Remove the "Option B: From CLI" section that contains:

```bash
planton auth login
planton auth token
```

### 2. `docs/configuration.md` (lines 28-32)

Remove the same "Option B: From CLI" section from the "How to obtain" subsection.

### 3. `docs/installation.md` (lines 119-123)

Remove the same "Option B: From CLI" section from the "Obtain API Key" section.

## Changes

- Remove all CLI command references for API key retrieval
- Keep the Web Console method as the recommended (and only) approach
- Maintain all other content and formatting
- The Web Console method remains as "Option A (Recommended)", but since there's no Option B anymore, we can simplify the heading