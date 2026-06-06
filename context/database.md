# Vext — Database Design

## Engine Choice: SQLite via GORM

Vext uses SQLite as its storage engine. It is a single file on disk, requires zero configuration, has no server process, and is battle-tested in production environments worldwide.

The driver used is `github.com/glebarez/sqlite`, a Pure Go SQLite implementation. This avoids requiring a C compiler (CGO) on the build machine and produces a fully static, portable binary. GORM (`gorm.io/gorm`) is used as the ORM layer: it provides a clean query DSL, handles struct-to-row mapping, and manages migrations via `AutoMigrate`.

**Database location:** `~/.config/vext/vext.db` (Linux/macOS) · `%AppData%\vext\vext.db` (Windows)

---

## Schema Design Decision: Polymorphic vs. Separate Tables

Vext needs to store different types of secrets — accounts have `username` + `password`, bank cards have `card_number` + `cvv` + `expiration`, and so on. These types have incompatible fields.

Two approaches were considered:

**Option A — Separate Tables:** One table per secret type (`credentials`, `bank_cards`, etc.). Clean per-type, but requires a schema migration every time a new type is added, and makes cross-type queries (like `vext list`) complex.

**Option B — Polymorphic Table with JSON Payload:** A single `secrets` table where each row stores a `type` field and an opaque encrypted blob. The blob contains a JSON object whose shape depends on the type.

**Vext uses Option B.** The database schema never needs to change when new secret types are added. The type system lives entirely in Go code, not in SQL columns.

---

## Schema

The schema is defined as a GORM model struct in `dal/schemas.go` and created via `db.AutoMigrate(&SecretRecord{})`:

```go
// SecretRecord is a 1:1 mapping of the `secrets` table.
type SecretRecord struct {
    ID        int64     `gorm:"primaryKey;autoIncrement"`
    Name      string    `gorm:"uniqueIndex;not null"`
    Type      string    `gorm:"not null"`
    Salt      []byte    `gorm:"not null;type:blob"`
    Nonce     []byte    `gorm:"not null;type:blob"`
    Encrypted []byte    `gorm:"not null;type:blob"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (SecretRecord) TableName() string { return "secrets" }
```

The equivalent SQL schema is:

```sql
CREATE TABLE IF NOT EXISTS secrets (
    id         INTEGER  PRIMARY KEY AUTOINCREMENT,
    name       TEXT     UNIQUE NOT NULL,
    type       TEXT     NOT NULL,
    salt       BLOB     NOT NULL,
    nonce      BLOB     NOT NULL,
    encrypted  BLOB     NOT NULL,
    created_at DATETIME,
    updated_at DATETIME
);
```

### Field Reference

| Field | Type | Description |
|---|---|---|
| `id` | INTEGER | Internal auto-incrementing primary key |
| `name` | TEXT | User-facing identifier. e.g. `"github"`, `"visa-debit"`. Must be unique. |
| `type` | TEXT | Secret type tag. e.g. `"account"`, `"card"`. Drives deserialization in Go. |
| `salt` | BLOB | 16 random bytes. Per-record. Used by Argon2id to derive the encryption key. |
| `nonce` | BLOB | 12 random bytes. Per-record. Used by AES-GCM as the initialization vector. |
| `encrypted` | BLOB | The AES-GCM ciphertext. Contains the JSON payload + GCM auth tag. |
| `created_at` | DATETIME | Timestamp of creation. Managed by GORM `autoCreateTime`. |
| `updated_at` | DATETIME | Timestamp of last modification. Managed by GORM `autoUpdateTime`. |

---

## How the Payload Column Works

The `encrypted` column is the heart of the polymorphic design.

### Writing (`vext add`)

```
User inputs (username, password)
        │
        ▼
Assembled into AccountSecret struct
        │
        ▼
Serialized to JSON bytes
  {"username":"bob","password":"<base64-encoded bytes>"}
        │
        ▼
JSON bytes encrypted with AES-256-GCM
  → produces salt + nonce + ciphertext
        │
        ▼
Stored: name, type, salt, nonce, encrypted → secrets table
```

### Reading (`vext get`)

```
Row fetched from secrets table (name, type, salt, nonce, encrypted)
        │
        ▼
Decrypted with AES-256-GCM (master password + salt + nonce)
  → produces JSON bytes
        │
        ▼
JSON deserialized into the correct Go struct (based on `type` field)
        │
        ▼
Fields printed to terminal
```

---

## Payload Schemas (Per Type)

These are the JSON structures that get encrypted and stored in `encrypted`. The database itself never sees these — it only stores the encrypted bytes.

**Note:** `AccountSecret.Password` is `[]byte` in Go, so it serializes as a base64-encoded string in JSON (standard Go behavior for `[]byte` fields).

### type: `"account"`

```json
{
  "username": "bob@example.com",
  "password": "<base64-encoded bytes>"
}
```

### type: `"card"` *(Phase 2)*

```json
{
  "card_number": "4111111111111111",
  "cvv": "123",
  "expiration": "12/27",
  "pin": "1234"
}
```

### type: `"note"` *(Phase 2)*

```json
{
  "content": "Server root password is stored in the safe at office."
}
```

Adding a new type requires only: defining a new Go struct under `core/secrets/`, adding a new `case` to `StoreSecretRequest.buildPayload()` and `RetrieveSecretUC`, and documenting it here. **The SQL schema does not change.**

---

## Migrations

Vext uses GORM's `AutoMigrate` on every `vext init` call. This is safe and idempotent — it only creates missing tables or adds missing columns. It never drops columns or changes existing ones.

```go
func Migrate(db *gorm.DB) error {
    return db.AutoMigrate(&SecretRecord{})
}
```

For future schema changes (e.g. adding a new column), adding the field to `SecretRecord` with the appropriate GORM tag is sufficient — `AutoMigrate` handles the rest.

---

## SQLite Gotchas to Watch

**WAL and journal files:** SQLite may create `-wal` or `-journal` files alongside `vext.db` during writes. These are temporary. Because Vext encrypts before writing, these files will never contain plaintext secrets — only encrypted blobs.

**File permissions:** After `vext init`, the database file is explicitly `chmod`-ed to `0600` (owner read/write only) by `InitStorageUC`. This prevents other users on the same machine from reading the file even if they obtain the path.

**Unique constraint errors:** GORM does not wrap SQLite's unique constraint violation in a typed error. `SQLiteRepository.Save` detects it via string inspection (`"UNIQUE constraint failed"`) and maps it to `core.ErrAlreadyExists`.
