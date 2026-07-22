# Phanes DNA Project Rules & Conventions

## 1. Layer Dependencies
<!-- Rules for package imports. Format: `src` cannot/can import `tgt` -->
<!-- No dependencies restricted -->

## 2. Code Conventions (LLM Evaluated)
<!-- Guidelines for code styling and comments -->
- Toda la documentación del código (docstrings) y comentarios de funciones o estructuras deben ser exclusivamente en español.
- Los símbolos del código (funciones, structs, variables) deben escribirse en inglés, pero su documentación explicativa en la línea anterior debe estar en español y seguir buenas prácticas.
- Return clean error contexts instead of panic statements.

## 3. Git Conventions (LLM Evaluated)
- Commit messages must follow Conventional Commits specification.

## 4. Naming & Scaffolding Conventions
- Case Style: `snake_case`
- Patterns:
  - controller: `[name]_controller.go`
  - service: `[name]_service.go`
  - repository: `[name]_repository.go`
