# Vext — Structure, Architecture & Best Practices

## Project Layout

```
vext/
├── cmd/                  # Cobra command definitions (one file per command)
│   ├── root.go
│   ├── init.go
│   ├── add.go
│   ├── get.go
│   ├── list.go
│   └── rm.go
├── pkg/
│   ├── crypto/           # All cryptographic logic (KDF + cipher)
│   ├── database/         # SQLite connection, migrations, queries
│   └── models/           # Go structs for each secret type
├── go.mod
├── go.sum
└── README.md
```

---

## Architectural Layers

Vext is structured in three clean, decoupled layers. Each layer has a single responsibility and communicates only with the layer directly below it.

```
┌─────────────────────────────────┐
│       CLI Layer (cmd/)          │  ← User interaction, input collection
├─────────────────────────────────┤
│     Domain Layer (pkg/)         │  ← Crypto logic + data models
├─────────────────────────────────┤
│   Persistence Layer (database/) │  ← SQLite read/write, no business logic
└─────────────────────────────────┘
```

**CLI Layer** orchestrates: it reads input from the user, calls crypto functions, and passes the result to the database. It knows about both layers below but contains zero crypto or SQL logic itself.

**Domain Layer** is pure logic: it doesn't know about Cobra, terminals, or SQLite. It takes bytes in, returns bytes out.

**Persistence Layer** only speaks SQL. It receives already-encrypted blobs and stores them. It never handles plaintext secrets.

---

## Design Patterns

### One Command, One File
Each Cobra command lives in its own file under `cmd/`. This keeps the logic isolated and makes the codebase easy to navigate. Adding a new command never touches existing files.

### Encrypt Before Persist
The database layer **never receives plaintext**. Encryption always happens in the `pkg/crypto` layer before any data is passed down. This is a hard architectural rule, not a convention.

### Fail Loudly on Crypto Errors
When AES-GCM authentication fails (wrong master password or tampered data), Vext returns a single generic error to the user: `master password incorrect or data corrupted`. The internal error is never exposed. This prevents leaking information about the failure mode.

### No Flags for Secrets
Passwords and the master password are **never** accepted as CLI flags or arguments. They are always collected via interactive prompts (`term.ReadPassword`). Flags end up in shell history; prompts do not.

---

## Coding Style & Best Practices

### Memory Hygiene
Any `[]byte` slice that holds a plaintext secret must be zeroed out immediately after use. Go's garbage collector does not guarantee when memory is reclaimed, so relying on it for secret data is unsafe.

```
// Immediately after encryption or decryption is complete:
zero out the master password byte slice
zero out the plaintext credential byte slice
```

### Error Messages
Errors shown to users are intentionally vague when they involve security (wrong password, bad decryption). Internal errors are logged for debugging but never surfaced raw to the terminal.

### Crypto Values Are Per-Record
Every stored secret gets its own randomly generated Salt and Nonce. This is non-negotiable. Reusing these values across records would allow an attacker with the database to compare ciphertexts and infer information.

### Database Path
The database file is stored using `os.UserConfigDir()`, which resolves to the OS-appropriate config directory:
- Linux/macOS: `~/.config/vext/vext.db`
- Windows: `%AppData%\vext\vext.db`

Never store the database in the current working directory or next to the binary.

---

## Dependencies

Vext keeps its dependency surface minimal and intentional.

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI framework |
| `golang.org/x/term` | Secure terminal input (hidden password prompts) |
| `golang.org/x/crypto/argon2` | Argon2id KDF |
| `modernc.org/sqlite` | Pure Go SQLite driver (no CGO required) |

All cryptographic primitives (`crypto/aes`, `crypto/cipher`, `crypto/rand`) come from Go's standard library. No third-party crypto packages are used.