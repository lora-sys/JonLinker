```markdown
# JonLinker Development Patterns

> Auto-generated skill from repository analysis

## Overview
This skill provides guidance on the development patterns and coding conventions used in the JonLinker Go codebase. It covers file organization, import/export styles, commit message patterns, and testing practices. This is ideal for contributors looking to maintain consistency and quality in JonLinker projects.

## Coding Conventions

### File Naming
- Use **snake_case** for all file names.
  - Example: `user_profile.go`, `data_parser.go`

### Import Style
- Use **relative imports** within the project.
  - Example:
    ```go
    import "./utils"
    ```

### Export Style
- Use **named exports** for functions, types, and variables.
  - Example:
    ```go
    // In user_profile.go
    package user

    func GetProfile(id int) *Profile {
        // implementation
    }
    ```

### Commit Messages
- Commit messages are **freeform** (no strict prefixes), with an average length of 58 characters.
  - Example:
    ```
    Add support for custom user avatars in profile handler
    ```

## Workflows

_No automated workflows detected in this repository._

## Testing Patterns

- **Testing Framework:** [Playwright](https://playwright.dev/)
- **Test File Naming:** Append `_test.go` to the filename.
  - Example: `user_profile_test.go`
- **Test Structure:**
  - Place test functions in the same package as the code under test.
  - Example:
    ```go
    // user_profile_test.go
    package user

    import "testing"

    func TestGetProfile(t *testing.T) {
        // test implementation
    }
    ```

## Commands
| Command | Purpose |
|---------|---------|
| /test   | Run all Go tests (files ending with `_test.go`) using Playwright |
| /lint   | Check code for style and convention adherence |
| /build  | Build the Go project |
```