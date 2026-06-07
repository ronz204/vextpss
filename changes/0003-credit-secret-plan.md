# 0003 — F-12: Bank Card Credentials (`type: credit`)

## Objective

Implement full support for storing and retrieving bank card credentials (`type: credit`) via the existing `vext add --type card <name>` and `vext get <name>` commands, with excellent terminal UX and clean architectural extensibility.

## Rationale

F-12 was deferred to Phase 2 because the MVP's polymorphic storage model needs no schema changes — only application-layer additions. Now that Phase 2 (F-08, F-09, F-10) is done, the path is clear. The `CreditSecret` model already exists as a skeleton. The main work is:

1. Completing the model
2. Extending the input collection layer (prompts) with a scalable multi-type design
3. Wiring retrieval and display

## Approach

### 1. Introduce `SecretCollector` — scalable input abstraction

The current `AddHandler` hardcodes account-specific prompts inline. With multiple secret types, this doesn't scale. We introduce a **`SecretCollector`** interface:

```go
type SecretCollector interface {
    Collect(p Prompter) (core.SecretPayload, error)
}
```

Each secret type owns its own collector:
- `AccountCollector` — wraps the existing username/password prompt logic
- `CreditCollector` — new; prompts for all card fields with proper hidden/visible grouping

A factory `NewSecretCollector(secretType string, opts CollectorOptions) (SecretCollector, error)` instantiates the right one. `CollectorOptions` carries optional params like `generate`, `genLength`, `genSymbols` (for account only).

**Why this is the right abstraction:** The `Prompter` interface handles I/O mechanics; the `SecretCollector` handles "what to ask for this type". Adding a future `NoteCollector` is a single file addition.

### 2. Refactor `StoreSecretRequest` — accept pre-built `Payload`

The current request carries type-specific fields (`AccountUsername`, `AccountPassword`). This doesn't scale — a `CreditRequest` would need 10+ more fields. Instead:

```go
type StoreSecretRequest struct {
    Name           string
    MasterPassword []byte
    Payload        core.SecretPayload  // built by the collector, validated by the UC
}
```

The UC derives the `Type` string from `Payload.GetType()`. The `buildPayload()` method disappears. The handler builds the payload via the collector before calling the UC.

This is a breaking change: all existing `StoreSecretRequest` usages must be updated.

### 3. Fix `CreditSecret` model — memory safety

The existing skeleton uses `string` for all fields. Sensitive fields must be `[]byte` to allow in-memory zeroing:

| Field | Current | New |
|---|---|---|
| `SecurityCode` | `string` | `[]byte` |
| `Pin` | `string` | `[]byte` |
| `BankPassword` | `string` | `[]byte` |
| `BankVirtualKey` | `string` | `[]byte` |
| All other fields | `string` | `string` (unchanged) |

### 4. Extend retrieval and display

- `deserialisePayload` in `retrieve_secret_uc.go` → add `"credit"` case
- `PrintSecret` in `formatter.go` → add `*secrets.CreditSecret` case with formatted display

## Specific Changes

### New files

| File | Purpose |
|---|---|
| `source/cmd/helpers/collectors.go` | `SecretCollector` interface, `AccountCollector`, `CreditCollector`, factory |
| `source/cmd/helpers/collectors_test.go` | Unit tests for both collectors |
| `source/core/secrets/credit_secret_test.go` | `CreditSecret.Validate()` tests |

### Modified files

| File | Change |
|---|---|
| `source/core/secrets/credit_secret.go` | Fix sensitive fields to `[]byte`; improve `Validate()` |
| `source/pkg/apps/store_secret_uc.go` | Replace type-specific fields with `Payload core.SecretPayload`; drop `buildPayload()` |
| `source/pkg/apps/store_secret_uc_test.go` | Update all `StoreSecretRequest` instantiations; add `credit` success test |
| `source/pkg/apps/retrieve_secret_uc.go` | Add `"credit"` case to `deserialisePayload` |
| `source/pkg/apps/retrieve_secret_uc_test.go` | Add credit retrieval test |
| `source/cmd/handlers/add_handler.go` | Add `--type` flag; dispatch to collector; new `Handle` signature |
| `source/cmd/handlers/add_handler_test.go` | Update call sites; add `--type card` test |
| `source/cmd/helpers/formatter.go` | Add `*secrets.CreditSecret` case to `PrintSecret` |
| `source/cmd/helpers/formatter_test.go` | Add credit display test |
| `context/features.md` | Mark F-12 as implemented |
| `context/commands.md` | Document `--type` flag for `vext add` |

## UX Design

### `vext add --type card visa-debit`

```
Card Number: 4532123456789012
Security Code (CVV): ••••
Expiration Month (1-12): 6
Expiration Year (YYYY): 2028
PIN: ••••
Bank Name (leave blank to skip): Bancolombia
Bank Username (leave blank to skip): user@bank.com
Bank Password (leave blank to skip): ••••
Bank Virtual Key (leave blank to skip): ••••
Cellphone (leave blank to skip): +57 300 123 4567
Country Code (leave blank to skip): CO
Master Password: ••••

[✓] Credential "visa-debit" saved.
```

Required: Card Number, Security Code, Expiration Month, Expiration Year, PIN.
Optional (shown with skip hint): all bank fields.

### `vext get visa-debit`

```
Master Password: ••••

Service:        visa-debit
Card Number:    4532 1234 5678 9012
Security Code:  123
Expires:        06/2028
PIN:            1234
Bank:           Bancolombia
Bank Username:  user@bank.com
Bank Password:  myb4nkp4ss
Virtual Key:    vk123456
Cellphone:      +57 300 123 4567
Country:        CO
```

Optional fields that were left blank are omitted from display.

### `vext add --type card <name>` with `--generate`

`--generate` is silently ignored for card type. The collector controls what is prompted.

## Testing

1. `CreditSecret.Validate()` — required fields, invalid month/year ranges
2. `AccountCollector.Collect()` — happy path, error propagation
3. `CreditCollector.Collect()` — happy path (all fields), optional fields skipped, prompt errors, invalid month
4. `StoreSecretUC.Execute()` — credit success, nil payload guard
5. `deserialisePayload` — credit case
6. `PrintSecret` — credit display, optional fields absent
7. `AddHandler.Handle()` — `--type card` success, unknown type error

## Estimation

~2-3 hours. Mostly mechanical extension of established patterns; the refactor of `StoreSecretRequest` is the most impactful change but touches a contained set of files.
