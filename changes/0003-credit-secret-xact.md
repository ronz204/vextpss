# 0003 — F-12: Bank Card Credentials [COMPLETED]

## Executive Summary

Implemented full support for `type: credit` secrets. Users can now store and retrieve bank card credentials via `vext add --type credit <name>` and `vext get <name>`. The change introduced a `SecretCollector` abstraction that makes adding future secret types a single file addition.

## What Went Well

- All tests passed on first run after implementation. No fixes needed.
- The `SecretCollector` interface cleanly separated "what to prompt for" from "how to prompt" — the existing `Prompter` interface required no changes.
- The `StoreSecretRequest` refactor (removing type-specific fields in favor of a pre-built `Payload`) made the UC simpler and eliminated the type-specific branching that would have grown unbounded.
- The `MockEncryptor` transparent pass-through design (returns input as output) made collector and retrieve tests straightforward.

## What Was Different

- The plan mentioned `"card"` as the type string in one place, but the existing model already used `"credit"` via `GetType()`. Used `"credit"` consistently throughout.
- The `CreditCollector` only prompts for `BankPassword` if `BankUsername` was provided (UX decision not in the original plan). This avoids a confusing situation where a bank password prompt appears with no associated username.

## Changes Made

### New files

| File | Purpose |
|---|---|
| `source/cmd/helpers/collectors.go` | `SecretCollector` interface, `AccountCollector`, `CreditCollector`, `NewSecretCollector` factory |
| `source/cmd/helpers/collectors_test.go` | 10 tests covering both collectors and factory |
| `source/core/secrets/credit_secret_test.go` | 8 tests for `CreditSecret.Validate()` |

### Modified files

| File | Change |
|---|---|
| `source/core/secrets/credit_secret.go` | `SecurityCode`, `Pin`, `BankPassword`, `BankVirtualKey` changed to `[]byte`; `Validate()` added month/year range checks; added `fmt` import |
| `source/pkg/apps/store_secret_uc.go` | `StoreSecretRequest` refactored: removed `Type`, `AccountUsername`, `AccountPassword`; added `Payload core.SecretPayload`; removed `buildPayload()`; type derived from `Payload.GetType()` |
| `source/pkg/apps/store_secret_uc_test.go` | Rewrote all tests to use new request format; added credit success and type-check tests |
| `source/pkg/apps/retrieve_secret_uc.go` | `deserialisePayload` extended with `"credit"` case |
| `source/pkg/apps/retrieve_secret_uc_test.go` | Added `TestRetrieveSecretUC_Execute_CreditSuccess`; updated unknown type test to use `"note"` (since `"credit"` is now valid) |
| `source/cmd/handlers/add_handler.go` | Added `--type` flag; dispatches to `NewSecretCollector`; `Handle` signature extended with `secretType string` |
| `source/cmd/handlers/add_handler_test.go` | Updated all `Handle` call sites; added `TestAddHandler_Handle_CreditSuccess` and `TestAddHandler_Handle_UnknownType` |
| `source/cmd/helpers/formatter.go` | Added `*secrets.CreditSecret` case to `PrintSecret`; `formatCardNumber` helper for 4-digit grouping; optional fields omitted when empty |
| `source/cmd/helpers/formatter_test.go` | Added 3 credit display tests |
| `context/features.md` | F-12 marked implemented; type string corrected to `credit` |
| `context/commands.md` | `vext add` documented with `--type`, credit field flow |

## Architecture After This Change

```
AddHandler
  └── NewSecretCollector(type, opts) → SecretCollector
        ├── AccountCollector.Collect(Prompter) → *AccountSecret
        └── CreditCollector.Collect(Prompter)  → *CreditSecret
  └── StoreSecretUC.Execute(req{Name, MasterPassword, Payload})
        └── Payload.GetType() → stored as secret.Type
```

Adding a future `NoteCollector`:
1. Add `NoteCollector` in `collectors.go`
2. Add `"note"` case to `NewSecretCollector` factory
3. Add `"note"` case to `deserialisePayload` in `retrieve_secret_uc.go`
4. Add `*NoteSecret` case to `PrintSecret` in `formatter.go`

## Lessons Learned

- Type-specific fields in request structs don't scale. A pre-built `Payload` interface in the request is the right pattern for polymorphic domain objects.
- Conditional prompts (bank password only if bank username is set) improve UX without adding complexity — the collector is the right place for that logic.

## Post-Session Adjustment

**`BankName` removed from `CreditSecret`:** After initial implementation, the field was deemed unnecessary. Removed from model, collector, formatter, and all tests. The `BankUsername` field implicitly identifies the bank portal.

## Next Steps

- F-09 `vext update` for `credit` type — currently update only handles accounts. The `UpdateSecretUC` would need type-aware resolution similar to how `AccountCollector` works.
- F-13 Secure Notes — follow the same collector pattern: `NoteCollector` + `NoteSecret` model.
- F-14 Clipboard (`vext get --copy`) — handler-level flag, not type-specific.
