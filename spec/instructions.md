## Important Files to Reference

When working on this project, always reference these key files for context and requirements:

### Technical Requirements and Standards
  - **Key information**: Architecture constraints, dependency restrictions, testing requirements
- `/spec/spec.md` - Detailed specifications and use cases
  - **When to reference**: Before detailed implementation, when understanding use cases, when designing interfaces
  - **Key information**: User scenarios, input/output formats, expected behaviors

### Reference Implementation
- `/tmp/atat-original/` - Rust reference implementation
  - **When to reference**: When implementing each module, when writing tests
  - **Key information**: Original logic, test cases, expected behavior

### Project Status

- `/TODO.md` - Current development status, issues, and progress tracking
  - **When to reference**:
    - Before starting work (checking current status)
    - During work (recording progress)
    - When completing work (marking completed items)
    - When discovering issues (recording problems)
  - **What to check/update**: Current work context, error and bug tracking, implementation status

example:
```markdown
- [ ] Implement feature C
- [ ] Fix bug in module D
- [ ] Update documentation
```

### How to Check Progress

Check both `/TODO.md` and `git log --oneline` to verify:
- TODO.md: Current status and remaining tasks
- git log: Recent commits and actual implementation progress

## Rust to Go Migration Guidelines

### Code Equivalence Principles
- **Line-by-line matching is NOT required**: Rust and Go have different idioms, libraries, and architectures
- **Behavioral equivalence is REQUIRED**: All test cases from the reference implementation must pass
- **Test-driven approach**: Port test cases first, then implement functionality to pass them

### Implementation Rules
1. **One module per PR**: Implement and complete one module at a time
2. **No parallel module implementation**: If 2+ modules are implemented simultaneously, all will be rejected
3. **Complete through GitHub Actions**: All implementations must pass CI/CD before merging

### Module Implementation Process
1. Create a feature branch: `git switch -c <branch-name>`
2. Implement the module
3. Port ALL test cases from reference implementation
4. **Add module to coverage tracking**: When adding small tests, add the module path to `/workspaces/gh-atat/scripts/coverage.txt` (e.g., `./internal/<module>/...`)
5. Ensure all tests pass: `go test -v ./internal/<module>/`
6. Create PR and verify through GitHub Actions
7. After merge, proceed to next module
