# Remove Non-Working CLI Commands for API Key Retrieval

**Date:** 2025-11-25  
**Type:** Documentation Update  
**Impact:** Documentation Only

## Summary

Removed references to non-working CLI commands for obtaining API keys from all documentation files. The CLI commands `planton auth login` and `planton auth token` were not functional, so all documentation now only references the Web Console method for obtaining API keys.

## Changes

### Files Modified

1. **README.md**
   - Removed "Option B: From CLI" section
   - Simplified heading from "Option A: From Web Console (Recommended)" to "From Web Console:"
   - Removed CLI commands: `planton auth login` and `planton auth token`

2. **docs/configuration.md**
   - Removed "Option B: From CLI" section from the PLANTON_API_KEY configuration instructions
   - Simplified heading to "From Web Console:"

3. **docs/installation.md**
   - Removed "Option B: From CLI" section from the "Obtain API Key" configuration step
   - Simplified heading to "From Web Console:"

## Rationale

The CLI commands for obtaining API keys were included in the documentation but were not functional. To avoid user confusion and maintain accurate documentation, all references to these commands have been removed. Users should only use the Web Console method to obtain their API keys.

## Migration Guide

No migration needed. This is a documentation-only change that removes incorrect information.

## Testing

- Verified that all three documentation files have been updated
- Confirmed that Web Console instructions remain intact and clear
- Ensured no other references to CLI API key retrieval exist in the codebase
