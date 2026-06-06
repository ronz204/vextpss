# Vext — Structure, Architecture & Best Practices

## Project Layout

```
source/
├── main.go                         ← DI wiring: builds all deps, assembles command tree
├── cmd/
│   ├── root.go                     ← package cmd; Cobra root command only
│   ├── handlers/                   ← One handler per CLI command
│   │   ├── add_handler.go
│   │   ├── get_handler.go
│   │   ├── init_handler.go
│   │   ├── list_handler.go
│   │   └── rm_handler.go
│   └── helpers/
│       ├── prompter.go             ← Prompter interface + CLIPrompter + MockPrompter
│       └── formatter.go            ← PrintSecret, PrintSecretList
├── core/                           ← Domain layer: entities, value objects, errors
│   ├── errors.go                   ← DomainError type + canonical error vars
│   ├── secret.go                   ← Secret entity, SecretID, SecretPayload interface
│   └── secrets/
│       └── account_secret.go       ← AccountSecret value object (Password []byte)
├── dal/                            ← Data Access Layer contracts + GORM model
│   ├── db.go                       ← Open/Close/Migrate via GORM
│   ├── initialiser.go              ← Initialiser struct (creates dir + DB + schema)
│   ├── repository.go               ← SecretRepository interface
│   ├── schemas.go                  ← SecretRecord GORM model (maps to "secrets" table)
│   └── repos/
│       └── sqlite_repository.go    ← SQLiteRepository: GORM implementation
├── pkg/
│   ├── apps/                       ← Application use cases (business logic)
│   │   ├── init_storage_uc.go      ← StorageInitialiser interface + InitStorageUC
│   │   ├── store_secret_uc.go      ← StoreSecretUC
│   │   ├── retrieve_secret_uc.go   ← RetrieveSecretUC
│   │   ├── list_secrets_uc.go      ← ListSecretsUC
│   │   └── delete_secret_uc.go     ← DeleteSecretUC
│   ├── configs/
│   │   └── app_config.go           ← AppConfig, Load() — OS-aware paths
│   ├── shared/
│   │   └── utils.go                ← Zero, GenerateSalt, Now
│   └── tokens/                     ← Encryption contracts + implementation
│       ├── encryptor.go            ← Encryptor interface
│       └── aes_gcm_encryptor.go    ← AESGCMEncryptor (Argon2id + AES-256-GCM)
└── tests/
    └── mocks/                      ← Reusable test doubles
        ├── mock_repo.go            ← MockRepository (dal.SecretRepository)
        ├── mock_encryptor.go       ← MockEncryptor (tokens.Encryptor)
        └── mock_initialiser.go     ← MockInitialiser (apps.StorageInitialiser)
```

---

## Architectural Layers

Vext is structured in four clean, decoupled layers. Each layer has a single responsibility and communicates only with the layer directly below it.

```
┌──────────────────────────────────────────────┐
│         Interface Layer (cmd/)               │  ← CLI: handlers, prompter, formatter
├──────────────────────────────────────────────┤
│      Application Layer (pkg/apps/)           │  ← Use cases: orchestrates domain + infra
├──────────────────────────────────────────────┤
│           Domain Layer (core/)               │  ← Entities, value objects, domain errors
├──────────────────────────────────────────────┤
│  Infrastructure Layer (dal/ + pkg/tokens/)   │  ← SQLite (GORM), AES-GCM encryptor
└──────────────────────────────────────────────┘
```

**Interface Layer (`cmd/`)** handles user interaction. It collects input, calls use cases, and formats output. It knows nothing about SQL, crypto, or domain rules.

**Application Layer (`pkg/apps/`)** orchestrates the operation: validates requests, calls the domain, invokes the encryptor, delegates to the repository. It contains no SQL or crypto implementations — only interfaces.

**Domain Layer (`core/`)** defines entities (`Secret`), value objects (`AccountSecret`), the `SecretPayload` polymorphic interface, and canonical domain errors. It imports nothing from the project — zero dependencies on infra or framework.

**Infrastructure Layer (`dal/`, `pkg/tokens/`)** provides concrete implementations of the contracts defined in the layers above:
- `dal/repos/SQLiteRepository` implements `dal.SecretRepository` using GORM.
- `pkg/tokens/AESGCMEncryptor` implements `tokens.Encryptor` using Argon2id + AES-256-GCM.

---

## Dependency Graph

```
cmd/handlers ──→ pkg/apps ──→ core
     │               │──────→ dal            (SecretRepository interface)
     │               └──────→ pkg/tokens     (Encryptor interface)
     │
cmd/helpers  (Prompter interface — used by handlers)

dal/repos ────→ dal          (SecretRecord, SecretRepository)
          ────→ core         (Secret, ErrSecretNotFound, etc.)
          ────→ gorm.io/gorm

pkg/tokens ───→ core         (ErrDecryptionFailed)
           ───→ pkg/shared

main.go ──────→ all packages above (single composition root)
```

**Key rule:** `core` imports nothing from the project. `pkg/apps` does not import from `cmd/`. `dal/repos` does not import from `pkg/apps`. Dependencies flow inward — never outward.

---

## Interface Placement

Interfaces are defined in the package that **uses** them (the consumer), not the one that implements them. This enforces the Dependency Inversion Principle and keeps layers independent:

| Interface | Package | Implemented by |
|---|---|---|
| `SecretRepository` | `dal` | `dal/repos.SQLiteRepository` |
| `Encryptor` | `pkg/tokens` | `pkg/tokens.AESGCMEncryptor` |
| `StorageInitialiser` | `pkg/apps` | `dal.Initialiser` |
| `Prompter` | `cmd/helpers` | `cmd/helpers.CLIPrompter` |

---

## Design Patterns

### Handler per Command
Each Cobra command is wrapped in a dedicated handler struct under `cmd/handlers/`. The handler's `CobraCommand()` method returns the `*cobra.Command`, and `Handle(ctx, ...)` contains the actual logic. This makes handlers testable without Cobra.

### Encrypt Before Persist
The persistence layer (`dal`) **never receives plaintext**. Encryption always happens in `pkg/tokens` before any data crosses the `dal` boundary. This is a hard architectural rule.

### Fail Loudly on Crypto Errors
When AES-GCM authentication fails, `AESGCMEncryptor` returns `core.ErrDecryptionFailed` — a domain error with the message `"master password incorrect or data corrupted"`. The raw crypto error is never surfaced.

### No Flags for Secrets
Passwords and the master password are **never** accepted as CLI flags. They are always collected via `Prompter.ReadPassword()`, which uses `term.ReadPassword` internally and leaves no shell history trail.

### Nil-safe Use Cases in main
When the database file does not exist yet (first run before `vext init`), `db` is nil and the data use cases (`storeUC`, `retrieveUC`, etc.) are initialized as nil pointers. Handlers must check for nil before use, which allows `vext init` to run cleanly on a fresh environment without an early crash.

---

## Coding Style & Best Practices

### Memory Hygiene
Any `[]byte` holding a plaintext secret must be zeroed immediately after use via `shared.Zero()`. This applies to master passwords, account passwords, and derived keys. Go's GC does not guarantee prompt reclamation of sensitive memory.

### Error Messages
User-facing errors involving security use intentionally vague messages (e.g., `"master password incorrect or data corrupted"`). Internal errors are wrapped with context for debugging but never printed raw to the terminal.

### Crypto Values Are Per-Record
Every stored secret gets its own randomly generated Salt (16 bytes) and Nonce (12 bytes). Reusing these values would allow an attacker with the database to correlate records.

### Database Path
The database file is stored via `os.UserConfigDir()`:
- Linux/macOS: `~/.config/vext/vext.db`
- Windows: `%AppData%\vext\vext.db`

Never store the database in the current working directory or next to the binary.

---

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI framework |
| `golang.org/x/term` | Secure terminal input (hidden password prompts) |
| `golang.org/x/crypto/argon2` | Argon2id KDF |
| `gorm.io/gorm` | ORM — query DSL, AutoMigrate, error handling |
| `github.com/glebarez/sqlite` | Pure Go SQLite driver for GORM (no CGO required) |

All cryptographic primitives (`crypto/aes`, `crypto/cipher`, `crypto/rand`) come from Go's standard library. No third-party crypto packages are used.
