# 0002 — Phase 2 Features: F-08, F-09, F-10 [COMPLETED]

## Executive Summary

Implemented three Phase 2 features in a single session: password generator (`vext gen`), credential update (`vext update`), and encrypted export/import (`vext export` / `vext import`). All four layers respected; no schema changes required.

## What Went Well

- Architecture fit naturally: `Update` + `GetAll` on the repo interface required minimal changes to the DAL.
- The polymorphic `deserialisePayload` function in `retrieve_secret_uc.go` was already accessible from `update_secret_uc.go` (same package), so no refactor was needed.
- The export/import roundtrip test validated the full cycle using the real UC pair with MockEncryptor.

## What Was Different

- **MockEncryptor returned same slice instead of a copy.** Both `Encrypt` and `Decrypt` were returning the input slice directly. Callers that `defer shared.Zero(plaintext)` would corrupt the "ciphertext" in the same allocation. Fixed: the mock now returns a fresh copy, matching real AES-GCM behavior.
- **Export UC zeroed RawMessage before serialization.** `json.RawMessage(plaintext)` is a type cast, not a copy. `shared.Zero(plaintext)` wiped the bundle payload before `json.Marshal`. Fixed: copy the decrypted bytes before zeroing, then wrap the copy in `json.RawMessage`.

## Changes Made

### New files
| File | Purpose |
|---|---|
| `source/pkg/shared/generator.go` | `GeneratePassword(length, useSymbols)` via `crypto/rand` |
| `source/pkg/apps/update_secret_uc.go` | `UpdateSecretUC` — decrypt current, resolve username, re-encrypt |
| `source/pkg/apps/export_secrets_uc.go` | `ExportSecretsUC` + archive types (exportBundle, exportFile) |
| `source/pkg/apps/import_secrets_uc.go` | `ImportSecretsUC` — skips existing names |
| `source/cmd/handlers/gen_handler.go` | `GenHandler` (flags: `--length`, `--no-symbols`) |
| `source/cmd/handlers/update_handler.go` | `UpdateHandler` |
| `source/cmd/handlers/export_handler.go` | `ExportHandler` (flag: `--out`, defaults to timestamped filename) |
| `source/cmd/handlers/import_handler.go` | `ImportHandler` |
| Test files for all of the above | Unit tests following existing patterns |

### Modified files
| File | Change |
|---|---|
| `source/dal/repository.go` | Added `FullRecord` type; added `GetAll`, `Update` to `SecretRepository` |
| `source/dal/repos/sqlite_repository.go` | Implemented `GetAll`, `Update` |
| `source/tests/mocks/mock_encryptor.go` | Fixed: Encrypt + Decrypt now return copies (not same slices) |
| `source/tests/mocks/mock_repo.go` | Added `GetAllFn`, `UpdateFn` fields |
| `source/cmd/handlers/add_handler.go` | Added `--generate`, `--gen-length`, `--gen-no-symbols` flags; Handle signature extended |
| `source/cmd/handlers/add_handler_test.go` | Updated call sites for new Handle signature; added `--generate` test |
| `source/main.go` | Wired up `updateUC`, `exportUC`, `importUC` and their handlers |

## Lessons Learned

- `json.RawMessage(slice)` is a type alias — it does not allocate. Always copy sensitive byte slices before zeroing when the copy is still needed downstream.
- Mock encryptors should return independent allocations to simulate real cipher behavior (real AES-GCM always returns a new allocation).

## Next Steps

- F-11 (Shell Autocompletion) — trivially enabled via Cobra's `completion` subcommand.
- F-12 / F-13 (Card + Note types) — need additions to the polymorphic payload switch in `store_secret_uc.go`, `retrieve_secret_uc.go`, and `update_secret_uc.go`.
- F-14 (Clipboard) — `vext get --copy` flag with optional auto-clear.
