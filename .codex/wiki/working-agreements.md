# Working Agreements

## Before Editing

- Read the relevant wiki page and OpenSpec artifacts.
- Inspect the existing repository structure before choosing an implementation
  pattern.
- Keep changes scoped to the request.

## Documentation

- Update this wiki when stable project facts change.
- Update OpenSpec when requirements or behavior change.
- Avoid recording guesses as facts.

## Code And Tests

- No build or test command is known yet.
- When a tech stack is introduced, record setup and verification commands here.
- Prefer focused tests for narrow behavior and broader tests for shared
  contracts or user-facing workflows.
- Any API contract change MUST update the backend implementation, frontend API
  client or callers, and related API documentation together.
- Any new logic that stores, reads, or changes persistent data MUST integrate
  with the real database layer. Do not satisfy persistent-data behavior with mock
  data, static fixtures, or frontend-only placeholders outside tests.

## Git Hygiene

- Do not revert unrelated work.
- Treat untracked or modified files as user-owned unless the current task
  explicitly created them.
