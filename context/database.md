# Vext — Database Design

## Engine Choice: SQLite

Vext uses SQLite as its storage engine. It is a single file on disk, requires zero configuration, has no server process, and is battle-tested in production environments worldwide. For a local-first CLI tool, it is the correct choice.

The driver used is `modernc.org/sqlite`, a Pure Go implementation. This avoids requiring a C compiler (CGO) on the build machine and produces a fully static, portable binary.

**Database location:** `~/.config/vext/vext.db`

---

## Schema Design Decision: Polymorphic vs. Separate Tables

Vext needs to store different types of secrets — accounts have `username` + `password`, bank cards have `card_number` + `cvv` + `expiration`, and so on. These types have incompatible fields.

Two approaches were considered:

**Option A — Separate Tables:** One table per secret type (`credentials`, `bank_cards`, etc.). Clean per-type, but requires a schema migration every time a new type is added, and makes cross-type queries (like `vext list`) complex.

**Option B — Polymorphic Table with JSON Payload:** A single `secrets` table where each row stores a `type` field and an opaque encrypted blob. The blob contains a JSON object whose shape depends on the type.

**Vext uses Option B.** The database schema never needs to change when new secret types are added. The type system lives entirely in Go code, not in SQL columns.

---

## Schema

```sql
CREATE TABLE IF NOT EXISTS secrets (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    name             TEXT    UNIQUE NOT NULL,
    type             TEXT    NOT NULL,
    salt             BLOB    NOT NULL,
    nonce            BLOB    NOT NULL,
    encrypted_payload BLOB   NOT NULL,
    created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP
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
| `encrypted_payload` | BLOB | The AES-GCM ciphertext. Contains the JSON payload + GCM auth tag. |
| `created_at` | DATETIME | Timestamp of creation. Informational only. |
| `updated_at` | DATETIME | Timestamp of last modification. Informational only. |

---

## How the Payload Column Works

The `encrypted_payload` column is the heart of the polymorphic design. Here's the lifecycle of data flowing into and out of it:

### Writing (add)

```
User inputs (username, password)
        │
        ▼
Assembled into a typed Go struct (AccountPayload)
        │
        ▼
Serialized to JSON string
  {"username":"bob","password":"hunter2"}
        │
        ▼
JSON bytes encrypted with AES-256-GCM
  → produces opaque ciphertext + auth tag
        │
        ▼
Stored in encrypted_payload column (BLOB)
```

### Reading (get)

```
encrypted_payload BLOB fetched from SQLite
        │
        ▼
Decrypted with AES-256-GCM using master password + stored nonce + salt
  → produces JSON bytes
        │
        ▼
JSON deserialized into the correct Go struct (based on `type` column)
        │
        ▼
Fields printed to terminal
```

---

## Payload Schemas (Per Type)

These are the JSON structures that get encrypted and stored in `encrypted_payload`. The database itself never sees these — it only stores the encrypted bytes.

### type: `"account"`

```json
{
  "username": "bob@example.com",
  "password": "hunter2"
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

### type: `"note"` *(Phase 2, example)*

```json
{
  "content": "Server root password is stored in the safe at office."
}
```

Adding a new type requires only: defining a new Go struct, adding a new `case` to the switch that handles serialization/deserialization, and documenting it here. **The SQL schema does not change.**

---

## Migrations

Vext runs `CREATE TABLE IF NOT EXISTS` on every startup. This is safe and idempotent — it only creates the table if it doesn't already exist.

For future schema changes (e.g. adding a new column), Vext will use a simple integer versioning approach stored in SQLite's built-in `PRAGMA user_version`. On startup, the current version is read and any pending migrations are applied in order.

---

## SQLite Gotchas to Watch

**WAL and journal files:** SQLite may create `-wal` or `-journal` files alongside `vext.db` during writes. These are temporary, but they could momentarily contain data in an intermediate state. Because Vext encrypts before writing, these files will never contain plaintext secrets — only encrypted blobs.

**File permissions:** On creation, `vext.db` should be created with permissions `0600` (owner read/write only). This prevents other users on the same machine from reading the file.