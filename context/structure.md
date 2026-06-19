# Vext — Structure, Architecture & Best Practices

## Project Layout

```
source/
├── main.go                              ← Entry point: calls cmd.Execute()
├── cmd/
│   ├── root.go                          ← Builds AppDeps, wires all Cobra commands
│   ├── adapters/                        ← One file per CLI command
│   │   ├── init_adapter.go
│   │   ├── add_adapter.go
│   │   ├── get_adapter.go
│   │   ├── list_adapter.go
│   │   ├── upd_adapter.go
│   │   ├── rm_adapter.go
│   │   ├── export_adapter.go
│   │   └── import_adapter.go
│   ├── collectors/                      ← Terminal input: prompts the user for data
│   │   ├── prompter.go                  ← Prompter struct + NewPrompter
│   │   ├── collector.go                 ← Collector struct (Payload, Master, Confirm)
│   │   ├── account.go                   ← collectAccount: prompts for account fields
│   │   └── finance.go                   ← collectFinance: prompts for finance fields
│   └── formatters/                      ← Terminal output: prints secrets and messages
│       ├── helpers.go                   ← Success, Error, Info printers
│       ├── tabtable.go                  ← PrintTabTable for vext list
│       ├── account.go                   ← PrintAccount
│       └── finance.go                   ← PrintFinance
├── funcs/                               ← Use cases: one file per operation
│   ├── create_secrets_func.go
│   ├── obtain_secret_func.go
│   ├── retrieve_secrets_func.go
│   ├── update_secret_func.go
│   ├── delete_secret_func.go
│   ├── export_secrets_func.go
│   └── import_secrets_func.go
├── secrets/                             ← Domain models and type constants
│   ├── secret.go                        ← Secret and Credential structs
│   ├── types.go                         ← TypeAccount, TypeFinance constants + IsKnownType
│   ├── account.go                       ← AccountSecret struct
│   ├── finance.go                       ← FinanceSecret struct
└── shared/
    ├── config.go                        ← AppName, DBPath()
    ├── deps.go                          ← AppDeps struct (shared across all adapters)
    ├── cryptors/
    │   ├── aes_gcm.go                   ← AESGCMEncryptor + EncryptInDto/OutDto/DecryptInDto
    │   └── aes_config.go                ← AESGCMConfig, Argon2Config, DefaultConfig
    ├── memory/
    │   └── cleaner.go                   ← Cleaner: zeros sensitive byte slices
    ├── passgen/
    │   └── generator.go                 ← Generate: cryptographically secure password gen
    ├── sentinel/
    │   └── domain.go                    ← Canonical error variables (ErrNotFound, etc.)
    └── storage/
        ├── database.go                  ← Open/Close gorm.DB
        ├── initialiser.go               ← Initialiser: creates dir + DB + schema
        ├── schemas.go                   ← SecretRecord GORM model
        ├── repository.go                ← SecretRepository: all CRUD operations
        └── session.go                   ← WithRepo: open DB → run fn → close DB
```

---

## Architectural Layers

Vext is organized in four layers. Each layer communicates only downward.

```
┌──────────────────────────────────────────────────────┐
│          Interface Layer  (cmd/adapters/)             │  ← CLI: parse args, format output
│          Input Layer      (cmd/collectors/)           │  ← Terminal: collect user input
│          Output Layer     (cmd/formatters/)           │  ← Terminal: print results
├──────────────────────────────────────────────────────┤
│          Use Case Layer   (funcs/)                    │  ← Orchestrate: validate + coordinate
├──────────────────────────────────────────────────────┤
│          Domain Layer     (secrets/)                  │  ← Structs, type constants
├──────────────────────────────────────────────────────┤
│          Infrastructure   (shared/storage/)           │  ← SQLite via GORM
│                           (shared/cryptors/)          │  ← AES-256-GCM + Argon2id
└──────────────────────────────────────────────────────┘
```

**Interface Layer (`cmd/adapters/`)** receives the Cobra command dispatch. Each adapter validates arguments, calls the appropriate use case func, and delegates formatting to `cmd/formatters/`. It does not contain business logic.

**Input Layer (`cmd/collectors/`)** handles all terminal I/O for data collection. `Prompter` reads visible and hidden lines. `Collector` assembles typed payloads from those prompts.

**Use Case Layer (`funcs/`)** is where business logic lives. Each func validates its DTO, coordinates the encryptor and repository, and returns a result or error. It knows nothing about Cobra or terminal I/O.

**Domain Layer (`secrets/`)** defines the data shapes: the `Secret` and `Credential` structs used at the persistence boundary, the typed payload structs (`AccountSecret`, `FinanceSecret`), and the `TypeAccount`/`TypeFinance` constants.

**Infrastructure (`shared/storage/`, `shared/cryptors/`)** provides the concrete implementations: SQLite persistence via GORM and AES-256-GCM encryption via Argon2id.

---

## Dependency Graph

```
main.go
  └──→ cmd/root.go
         ├──→ cmd/adapters      ──→ funcs
         │                      ──→ shared/storage
         │                      ──→ cmd/formatters
         │                      ──→ secrets
         │                      ──→ shared/sentinel
         ├──→ cmd/collectors    ──→ secrets
         │                      ──→ shared/memory
         │                      ──→ shared/sentinel
         ├──→ shared            ──→ cmd/collectors   [⚠ see audit.md]
         │                      ──→ shared/cryptors
         ├──→ shared/cryptors   ──→ shared/memory
         │                      ──→ shared/sentinel
         └──→ shared/storage    ──→ secrets
                                ──→ shared/sentinel

funcs ──→ secrets
      ──→ shared/cryptors
      ──→ shared/memory
      ──→ shared/sentinel
      ──→ shared/storage
```

**No abstract interfaces exist** between layers. All types flowing through `AppDeps` and func constructors are concrete:
- `*collectors.Collector`
- `*cryptors.AESGCMEncryptor`
- `*storage.SecretRepository`

---

## Key Design Patterns

### Adapter per Command
Each Cobra command is defined in a dedicated file under `cmd/adapters/`. The adapter function (e.g., `AddCmd`) receives `shared.AppDeps` and returns a `*cobra.Command`. The actual logic lives in a private `runXxx` function. This keeps the command wiring separate from the execution logic.

### WithRepo Session Pattern
Database access is always wrapped in `storage.WithRepo(dbPath, fn)`. This function opens the DB, creates a `*SecretRepository`, passes it to `fn`, then closes the DB — regardless of error. No adapter or func manages DB lifecycle directly.

### Encrypt Before Persist
The persistence layer (`shared/storage/`) **never receives plaintext**. Encryption always happens in `funcs/` before any data crosses the `storage` boundary. This is a hard architectural rule.

### Deferred Memory Zeroing
Every `[]byte` holding sensitive data is zeroed via `defer memory.Cleaner(b)` immediately after it is assigned — before any error path can return early and leave sensitive data in memory.

### Fail Loudly on Crypto Errors
When AES-GCM authentication fails, `AESGCMEncryptor` returns `sentinel.ErrDecryptionFailed`. The raw `crypto/cipher` error is never surfaced to the user.

### No Flags for Secrets
Passwords and master passwords are **never** accepted as CLI flags. They are always collected via `Collector.Master()` or `Collector.Payload()`, which use `Prompter.ReadSecret()` → `term.ReadPassword` internally.

---

## Coding Conventions

### Memory Hygiene
Any `[]byte` holding sensitive data must be zeroed via `memory.Cleaner(b)` as a `defer` immediately after assignment. This applies to master passwords, plaintext payloads, and derived keys.

### Error Messages
User-facing security errors are intentionally vague (`"wrong master password or data corrupted"`). Internal errors are wrapped with `fmt.Errorf("context: %w", err)` but never printed raw.

### Crypto Values Are Per-Record
Every stored secret gets its own randomly generated Salt (16 bytes) and Nonce (12 bytes). Reusing these values would allow an attacker with the database to correlate records.

### Database Path
The database file is stored via `os.UserConfigDir()`:
- Linux/macOS: `~/.config/vext/vext.db`
- Windows: `%AppData%\vext\vext.db`

---

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI framework |
| `golang.org/x/term` | Secure terminal input (hidden password prompts) |
| `golang.org/x/crypto/argon2` | Argon2id key derivation |
| `gorm.io/gorm` | ORM — query DSL, AutoMigrate, error handling |
| `github.com/glebarez/sqlite` | Pure Go SQLite driver (no CGO required) |

All cryptographic primitives (`crypto/aes`, `crypto/cipher`, `crypto/rand`) come from Go's standard library.
