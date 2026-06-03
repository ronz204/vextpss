# Vext — Data Modeling

## Overview

Vext's data model has two distinct layers: the **database layer** (what SQLite stores) and the **application layer** (what Go works with). These are intentionally separated. The database is type-agnostic; Go owns the type system.

---

## Database Layer: The Secret Record

Every item stored in Vext, regardless of type, maps to this single struct in Go:

```go
type SecretRecord struct {
    ID               int64
    Name             string
    Type             string
    Salt             []byte
    Nonce            []byte
    EncryptedPayload []byte
    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

This is a 1:1 mapping of the `secrets` table. It is what the database layer reads and writes. It contains zero plaintext secret data — `EncryptedPayload` is always opaque bytes at this level.

---

## Application Layer: Typed Payloads

The application layer defines one struct per secret type. These structs are what get serialized to JSON and then encrypted before storage.

### AccountPayload

Used when `type = "account"`. Represents a typical login credential.

```go
type AccountPayload struct {
    Username string `json:"username"`
    Password string `json:"password"`
}
```

### CardPayload *(Phase 2)*

Used when `type = "card"`. Represents a payment card.

```go
type CardPayload struct {
    CardNumber string `json:"card_number"`
    CVV        string `json:"cvv"`
    Expiration string `json:"expiration"`
    PIN        string `json:"pin"`
}
```

### NotePayload *(Phase 2, example)*

Used when `type = "note"`. Represents a free-form secure note.

```go
type NotePayload struct {
    Content string `json:"content"`
}
```

---

## The Type Dispatch Pattern

When reading a secret, Vext decrypts the payload first (getting raw JSON bytes), then uses the `Type` field from the database record to decide how to deserialize it:

```go
switch record.Type {
case "account":
    // unmarshal into AccountPayload, display username + password
case "card":
    // unmarshal into CardPayload, display card fields
case "note":
    // unmarshal into NotePayload, display content
default:
    // return an error: unknown type
}
```

This switch lives in the CLI layer, not in the crypto or database layer. Each layer only does what it's responsible for.

---

## Adding a New Secret Type

The polymorphic model is designed so that adding a new type is contained and predictable. The checklist is:

1. Define a new `XxxPayload` struct in `pkg/models/`.
2. Add a corresponding `case "xxx"` to the type dispatch switch.
3. Create a new Cobra command (or subcommand) that collects the right fields from the user and populates the struct.
4. Update `database.md` and `features.md` documentation.

The SQL schema does not change. The crypto layer does not change.

---

## What the Database Sees vs. What Go Sees

This table illustrates the separation between what is stored and what the application works with:

| Concern | Database Sees | Go Sees |
|---|---|---|
| Secret name | `"github"` (plaintext) | `"github"` (string) |
| Secret type | `"account"` (plaintext) | `"account"` (string, drives dispatch) |
| Secret data | Encrypted BLOB | `AccountPayload{Username: "...", Password: "..."}` |
| Salt | Raw bytes | `[]byte` passed to Argon2id |
| Nonce | Raw bytes | `[]byte` passed to AES-GCM |

The name and type are the only fields that live unencrypted in the database. This is intentional: they are needed for lookup and dispatch, but they are not sensitive. The actual secret content is always encrypted.

---

## Design Constraints

**No nullable payload fields.** All fields within a payload struct should be treated as required. If a field is optional for a given type, use an empty string rather than a pointer, to keep serialization predictable.

**No cross-type queries on payload content.** Because the payload is an opaque encrypted blob, SQLite cannot query inside it (e.g., `WHERE username = ?`). Lookups are always by `name` or `type`. This is an acceptable limitation for a local personal tool.

**Payload structs are append-only.** Once a payload type is in production use, do not rename or remove fields. Doing so would break deserialization of existing records. New fields can be added with zero-value defaults.