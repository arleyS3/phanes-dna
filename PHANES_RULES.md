# Phanes DNA Project Rules & Conventions

## 1. Layer Dependencies
<!-- Rules for package imports. Format: `src` cannot/can import `tgt` -->
<!-- No dependencies restricted -->

## 2. Code Conventions (LLM Evaluated)
<!-- Guidelines for code styling and comments -->
- Inline comments can be in Spanish, but code symbols (classes, functions) must be in English.
- Return clean error contexts instead of panic statements.

## 3. Git Conventions (LLM Evaluated)
- Commit messages must follow Conventional Commits specification.

## 4. Naming & Scaffolding Conventions
- Case Style: `snake_case`
- Patterns:
  - controller: `[name]_controller.go`
  - service: `[name]_service.go`
  - repository: `[name]_repository.go`
