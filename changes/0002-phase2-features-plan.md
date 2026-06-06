# 0002 — Phase 2 Features: F-08, F-09, F-10

## Objective

Implement three Phase 2 features: password generator (`vext gen`), credential update (`vext update`), and encrypted export/import (`vext export` / `vext import`). All features follow the existing four-layer architecture without schema changes.

## Rationale

F-08 eliminates weak human-chosen passwords at creation time. F-09 closes the usability gap of the current `rm` + `add` rotation flow. F-10 provides disaster recovery without cloud dependency — a known gap flagged since MVP.

## Approach

- F-08: Pure utility function in `pkg/shared`. No repo or encryptor needed. GenHandler calls it directly. AddHandler gains `--generate` flag to use it during `vext add`.
- F-09: New `Update` method added to `dal.SecretRepository`. UpdateSecretUC decrypts current state to resolve an optional username, re-encrypts with new salt/nonce, and delegates to the repo.
- F-10: New `GetAll` method on `dal.SecretRepository` returns all records with ciphertext. ExportSecretsUC decrypts each record then re-encrypts the full bundle. ImportSecretsUC reverses the process, skipping already-existing names.

## Specific Changes

### New files
- `source/pkg/shared/generator.go` — `GeneratePassword(length int, useSymbols bool) (string, error)` via `crypto/rand`
- `source/pkg/apps/update_secret_uc.go` — `UpdateSecretUC`, `UpdateSecretRequest`
- `source/pkg/apps/export_secrets_uc.go` — `ExportSecretsUC`, `ExportSecretsRequest`, archive types
- `source/pkg/apps/import_secrets_uc.go` — `ImportSecretsUC`, `ImportSecretsRequest`, `ImportSecretsResponse`
- `source/cmd/handlers/gen_handler.go` — `GenHandler` (flags: `--length`, `--no-symbols`)
- `source/cmd/handlers/update_handler.go` — `UpdateHandler`
- `source/cmd/handlers/export_handler.go` — `ExportHandler` (flag: `--out`)
- `source/cmd/handlers/import_handler.go` — `ImportHandler`

### Modified files
- `source/dal/repository.go` — add `FullRecord` type; add `Update`, `GetAll` to `SecretRepository`
- `source/dal/repos/sqlite_repository.go` — implement `Update`, `GetAll`
- `source/tests/mocks/mock_repo.go` — add `UpdateFn`, `GetAllFn` mock fields
- `source/cmd/handlers/add_handler.go` — add `--generate`, `--gen-length`, `--gen-no-symbols` flags
- `source/main.go` — wire up new use cases and handlers

### Export file format (`.vext`)
JSON envelope on disk:
```json
{ "version": 1, "salt": "<base64>", "nonce": "<base64>", "payload": "<base64 ciphertext>" }
```
The ciphertext decrypts to an `exportBundle` containing name, type, and plaintext payload for each secret.

## Testing

- Unit tests for `GeneratePassword` (length, charset correctness, randomness sanity)
- Unit tests for each new UC with mock repo + mock encryptor
- Unit tests for each new handler with mock prompter

## Estimation

Single session; no external dependencies required (all crypto from stdlib).
