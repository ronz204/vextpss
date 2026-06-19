# Vext — Data Modeling

## Overview

Vext's data model has two distinct layers: the **persistence layer** (what SQLite stores) and the **domain layer** (what Go works with). These are intentionally separated. The database is type-agnostic; Go owns the type system.

---

## Persistence Layer: Secret and Credential

The persistence boundary uses two structs defined in `secrets/secret.go`:

```go
// Secret holds metadata and crypto material for a stored record.
// It is what the repository reads and writes (minus the encrypted blob).
type Secret struct {
    ID        SecretID
    Name      string
    Type      string
    Salt      []byte
    Nonce     []byte
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Credential pairs a Secret with its encrypted payload.
// Used when the full blob is needed (export, get).
type Credential struct {
    Secret
    Encrypted []byte
}
```

`Secret` and `Credential` carry zero plaintext secret data — `Encrypted` is always an opaque blob at this level.

---

## Domain Layer: Typed Payload Structs

The domain layer defines one struct per secret type under `secrets/`. These structs are serialized to JSON and then encrypted before storage.

### AccountSecret

Used when `type = "account"`. Represents a typical login credential.
Defined in `secrets/account.go`.

```go
type AccountSecret struct {
    Username string `json:"username"`
    Password []byte `json:"password"`
}
```

**Note:** `Password` is `[]byte` so it serializes as a base64-encoded string in JSON (standard Go behavior). This makes it straightforward to zero out in memory after use.

### FinanceSecret

Used when `type = "finance"`. Represents a payment card with optional bank portal fields.
Defined in `secrets/finance.go`.

```go
type FinanceSecret struct {
    CardNumber      string `json:"card_number"`
    SecurityCode    []byte `json:"security_code"`
    CardPin         []byte `json:"card_pin"`
    ExpirationMonth int    `json:"expiration_month"`
    ExpirationYear  int    `json:"expiration_year"`
    BankUsername    string `json:"bank_username"`
    BankPassword    []byte `json:"bank_password"`
    BankVirtualKey  []byte `json:"bank_virtual_key"`
    BankCellphone   string `json:"bank_cellphone"`
}
```

Required fields: `CardNumber`, `SecurityCode`, `CardPin`, `ExpirationMonth`, `ExpirationYear`.
Optional fields: all `Bank*` fields — left as empty/zero if not provided.

### NoteSecret *(Phase 2)*

```go
type NoteSecret struct {
    Content string `json:"content"`
}
```

---

## Type Constants

Secret types are defined as constants in `secrets/types.go`:

```go
const (
    TypeAccount = "account"
    TypeFinance = "finance"
)

func IsKnownType(t string) bool {
    switch t {
    case TypeAccount, TypeFinance:
        return true
    }
    return false
}
```

`IsKnownType` is used by use case funcs to validate DTOs before attempting encryption.

---

## The Type Dispatch Pattern

When reading a secret, the adapter decrypts the payload first (getting raw JSON bytes), then uses the `Type` field from the database record to decide how to deserialize. This switch lives in `cmd/adapters/get_adapter.go`:

```go
switch result.Type {
case secrets.TypeAccount:
    var s secrets.AccountSecret
    json.Unmarshal(result.Payload, &s)
    formatters.PrintAccount(result.Name, s)
case secrets.TypeFinance:
    var s secrets.FinanceSecret
    json.Unmarshal(result.Payload, &s)
    formatters.PrintFinance(result.Name, s)
default:
    return fmt.Errorf("unknown secret type %q", result.Type)
}
```

The crypto and storage layers are completely unaware of the type distinction — they only see bytes.

---

## Adding a New Secret Type

The polymorphic model is designed so adding a new type is contained and predictable:

1. Define a new struct in `secrets/xxx.go`.
2. Add the type constant to `secrets/types.go` and update `IsKnownType`.
3. Add a `collectXxx(p *Prompter)` function in `cmd/collectors/xxx.go`.
4. Add a `case secrets.TypeXxx` to `Collector.Payload()` in `cmd/collectors/collector.go`.
5. Add a `case secrets.TypeXxx` to the dispatch switch in `cmd/adapters/get_adapter.go`.
6. Add a `formatters.PrintXxx` function if a custom display is needed.
7. Update `database.md`, `features.md`, and `commands.md`.

**The SQL schema does not change. The crypto layer does not change.**

---

## What the Database Sees vs. What Go Sees

| Concern | Database Sees | Go Sees |
|---|---|---|
| Secret name | `"github"` (plaintext) | `string` |
| Secret type | `"account"` (plaintext) | `string`, drives dispatch |
| Secret data | Encrypted BLOB | `AccountSecret{Username: "...", Password: [...]}` |
| Salt | Raw bytes | `[]byte` passed to Argon2id |
| Nonce | Raw bytes | `[]byte` passed to AES-GCM |

Name and type are the only fields that live unencrypted. They are needed for lookup and dispatch but are not sensitive. The actual secret content is always encrypted.

---

## Design Constraints

**No nullable payload fields.** All fields within a payload struct should be treated as required. If a field is optional for a given type, use an empty value rather than a pointer, to keep serialization predictable.

**No cross-type queries on payload content.** Because the payload is an opaque encrypted blob, SQLite cannot query inside it. Lookups are always by `name` or `type`.

**Payload structs are append-only.** Once a payload type is in production use, do not rename or remove fields — this would break deserialization of existing records. New fields can be added with zero-value defaults.
