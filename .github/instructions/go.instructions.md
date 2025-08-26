---
applyTo: "app/models/**/*.go"
---

- Do not add comments inside the function unless explicitly requested.
- Do not add any new dependencies or external packages without approval.
- Follow the existing code style and conventions.
- Follow Go coding conventions (https://go.dev/doc/effective_go)
- Follow best practices for error handling and logging.
- Ensure proper context usage and cancellation handling.
- Before making any changes:
  - review the existing code and understand its functionality.
  - consider potential edge cases and how to handle them.
  - wait for approval before making any changes.
- error strings should not be capitalized
- follow go-staticcheck rules