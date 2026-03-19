## Linter vs Formatter Classification

### Formatter (auto-fixable)

These are purely stylistic and can be deterministically rewritten without semantic understanding.

---

**FMT-001 — Declaration merging**
Multiple consecutive single `type`, `const`, or `var` declarations MUST be merged into a corresponding declaration block.

**FMT-002 — File declaration order**
Declarations within a file MUST follow the order: `type`, `const`, `var`, `func`. A formatter can reorder top-level declaration groups.

**FMT-003 — Function body spacing**
Functions MUST NOT start or end with empty lines. Logical blocks MUST NOT be separated by more than one blank line within a function body.

**FMT-004 — Interface method spacing**
Interface method signatures MUST be separated by exactly one blank line.

**FMT-005 — Type block spec spacing**
Type specs inside `type (...)` blocks MUST be separated by exactly one blank line.

---

### Linter (requires static analysis, not auto-fixable)

All rules skip generated Go files (files marked with `// Code generated ... DO NOT EDIT.`).

---

**LINT-001 — Log key casing**
String-literal slog key arguments inspected by this rule MUST be in `lowercase_snake_case`. Keys using dot notation (e.g. `error.message`) are permitted. Keys in `PascalCase`, `camelCase`, or containing uppercase letters are flagged.

**LINT-002 — Log message casing**
The first argument (message string) passed to `slog.*` calls MUST start with a lowercase letter.

**LINT-003 — Log key dot notation for grouped keys**
Log keys that semantically belong to a group (e.g. error fields, user fields) MUST use dot notation.

This rule is configuration-driven via `-dot-notation` as comma-separated `key=dotted_key` pairs (for example: `error=error.message,userId=user.id`). Only configured keys are flagged.

**LINT-004 — Context as first parameter**
Any function that accepts a `context.Context` MUST have it as the first parameter. Functions with `context.Context` in any other position MUST be flagged.

**LINT-005 — Excessive function parameters**
Functions with more than 4 parameters MUST be flagged. The message SHOULD suggest introducing a `{Name}Params` struct.

**LINT-006 — Excessive return values**
Functions with more than 2 return values MUST be flagged. The message SHOULD suggest introducing a `{Name}Result` struct.

Exception: functions referenced by `fx.Provide(...)` are excluded from this rule.

**LINT-007 — Enum value prefix**
For any named type backed by a primitive (string, int, etc.) with associated `const` values, each constant MUST be prefixed with the type name. Constants that do not carry the type name as a prefix MUST be flagged.

Configurable exceptions are supported via `package.Type` values. Default exception: `environment.Environment`.

**LINT-008 — Package name underscore**
Package names MUST NOT contain underscores. Any `package` declaration with an underscore in the name MUST be flagged.

Package-name suffixes can be excluded from this check via configuration. Default excluded suffix: `_test`.

**LINT-009 — Package name plural**
Package names that are pluralized MUST NOT be used.

The rule detects plural names generically (for example names ending with `s`), with configurable package-name exceptions.
Default configured exceptions: `types`, `handlertypes`, `params`.

**LINT-010 — Interface location**
Only interfaces suffixed with `Service` or `Store` MUST be declared in a `types` package (i.e. a file whose package is `types`). `Service`/`Store` interface declarations found outside of a `types` package MUST be flagged. Exception: packages under `core/` are allowed to declare infrastructure `Service`/`Store` interfaces outside `types`. Other interfaces are allowed outside `types`.

**LINT-011 — Service interface suffix**
Interfaces whose names do not end with `Service`, `Store`, or `Worker` and are located in a `types` package MUST be evaluated. Specifically, interfaces semantically acting as services MUST be suffixed `Service`, those acting as stores MUST be suffixed `Store`, and worker-style interfaces MAY be suffixed `Worker`. In practice, enforce: all interfaces in `types/` MUST end with `Service`, `Store`, or `Worker`.

**LINT-012 — Store function return types**
In packages whose name contains `store`, methods on receivers suffixed `Store` MUST NOT return types from packages whose import path contains `core/model` (including pointers/slices of those types).

**LINT-013 — Store struct interface assertion**
In packages whose name contains `store`, every exported struct suffixed `Store` MUST have a compile-time assertion in `store.go` whose value side matches `(*{Name}Store)(nil)` (for example: `var _ types.AnyStore = (*UserStore)(nil)`).

**LINT-014 — Service struct interface assertion**
In packages whose name contains `service`, every exported struct suffixed `Service` MUST have a compile-time assertion in `service.go` whose value side matches `(*{Name}Service)(nil)` (for example: `var _ types.AnyService = (*UserService)(nil)`).

**LINT-015 — One exported function per store/service file**
Files in packages whose name contains `store`, `service`, or `handler` (excluding `store.go`, `service.go`, and `fx_module.go`) are checked based on exported methods whose receiver name ends with `Store`, `Service`, or `Handler`.

If a file contains more than one such exported layer method, it is flagged. Exported non-method functions are ignored by this rule.

**LINT-016 — Middleware naming: Inject***
In packages whose names end with `handler`, any function named `Inject{Name}` or `inject{Name}` (with non-empty `{Name}`) MUST be declared in `inject_{name}.go`. Violations are flagged.

**LINT-017 — Middleware naming: Require***
In packages whose names end with `handler`, any function named `Require{Name}` or `require{Name}` (with non-empty `{Name}`) MUST be declared in `require_{name}.go`. Violations are flagged.

**LINT-018 — Middleware naming outside handler**
Outside packages whose names end with `handler`, exported functions with middleware signature `func(http.Handler) http.Handler` MUST be named `Middleware`. Non-matching names are flagged.

**LINT-019 — FxModule file location**
In packages whose names end with `store`, `service`, or `handler`, if a top-level variable named `FxModule` is declared, it MUST be located in:
- `store.go` for `*store` packages,
- `service.go` for `*service` packages,
- `handler.go` for `*handler` packages.

**LINT-020 — Error variable location (types package)**
In `types` packages only, error variables prefixed with `Err` MUST be declared in `errors.go`. `Err*` variables declared in other files within `types` MUST be flagged. Non-`types` packages are excluded from this rule.

**LINT-021 — RecordNotFound as typed error**
In packages whose name contains `store`, direct `return` expressions of these known not-found sentinels are flagged:
- `sql.ErrNoRows`
- `pgx.ErrNoRows`
- `gorm.ErrRecordNotFound`

**LINT-022 — Handler route file naming**
In module-scoped handler packages (names ending with `handler`, excluding package `handler`), exported methods on receivers `*{Name}Handler` MUST be located in `{route}.go` files, where `{route}` is the method name in snake_case after de-duplicating `{name}` when it is already present in `{route}` (including simple plural forms).

Examples:
- `assethandler`: `AssetHandler.ListAssets` -> `list.go`
- `assethandler`: `AssetHandler.GetAsset` -> `get.go`
- `authhandler`: `AuthHandler.Login` -> `login.go`

**LINT-023 — Route Input/Output type location**
In module-scoped handler packages (names ending with `handler`, excluding package `handler`), route wrapper types suffixed `Input` or `Output` MUST be declared in the route file determined by LINT-022 (`{route}.go` after de-duplication).

Shared payload structs SHOULD be declared in package `handlertypes` and SHOULD use descriptive names such as `Tenant`, `User`, or `InvitationToken` rather than `*BodyOutput`.

**LINT-024 — Shared body type naming**
In packages whose names end with `handler`, for files that are not route files, explicit body helper type names containing `Body` MUST match `{Name}BodyInput` or `{Name}BodyOutput`. Non-matching names are flagged.

**LINT-025 — Handler struct file location**
In module-scoped handler packages (names ending with `handler`, excluding package `handler`), struct types suffixed `Handler` MUST be declared in `handler.go`.

**LINT-026 — Body-only helper struct naming**
In packages whose names end with `handler`, struct types that are referenced only by body structs (`*BodyInput`/`*BodyOutput`) MUST:
- start with the parent body prefix (parent name without `Input`/`Output`), and
- end with the corresponding parent suffix (`Input` or `Output`).

**LINT-027 — No json tags in model structs**
In `model` packages, struct fields MUST NOT declare `json` tags. Fields with `json` tags are flagged.

**LINT-028 — Exported model fields require gorm tag**
In `model` packages, exported struct fields MUST declare a `gorm` tag attribute.

**LINT-029 — Relation field pointer shape**
In `model` packages, relation fields identified by `gorm` tag attributes `foreignKey`, `many2many`, or `polymorphicType` MUST be either:
- a pointer (`*Type`), or
- a slice of pointers (`[]*Type`).

**LINT-030 — Protected roots must not import sibling roots**
Packages under protected module roots (default: `core`) MUST NOT import packages from sibling roots in the same module.

The protected roots are configurable via `-lint030.roots` as a comma-separated list.

Example in module `daiteo.io`:
- `daiteo.io/core/pagination` importing `daiteo.io/core/model` is allowed.
- `daiteo.io/core/pagination` importing `daiteo.io/smarthubserver/types` is flagged.

**LINT-031 — httpapi path params lowerCamelCase**
For `httpapi` route registrations using `WithPattern("METHOD /path")` with a string-literal pattern, path parameters inside `{...}` MUST be `lowerCamelCase`.

Struct field tags using `path:"..."` MUST also use `lowerCamelCase`.

Examples:
- `/api/objects/{objectId}/definition` is valid.
- `/api/objects/{object_id}/definition` is flagged.
- `` `path:"invitationToken"` `` is valid.
- `` `path:"invitation_token"` `` is flagged.

**LINT-032 — Layer constructor naming and uniqueness**
In packages whose names end with `store`, `service`, or `<module>handler` (excluding package `handler`), exported top-level constructor functions prefixed with `New` MUST follow these rules:
- the constructor name MUST be exactly `New` (for example `NewUserService` is flagged),
- at most one exported `New*` constructor may be declared in the package.
