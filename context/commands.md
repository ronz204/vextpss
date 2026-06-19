# Vext — Commands Reference

## Global Rules

- **Secrets are never passed as flags.** All sensitive inputs (service passwords, master password) are collected via hidden interactive prompts. This prevents them from appearing in shell history.
- **The master password is never stored.** It is requested on-demand and discarded from memory immediately after use via `memory.Cleaner`.
- **Commands are idempotent where stated.** `vext init` can be run multiple times safely.

---

## `vext init`

Initializes the Vext environment on first use.

```
vext init
```

**What it does:**
- Creates the config directory at `~/.config/vext/` if it doesn't exist.
- Creates `vext.db` inside that directory.
- Runs the `AutoMigrate` migration to set up the schema (`CREATE TABLE IF NOT EXISTS`).
- Sets file permissions to `0600` on the database file.

**Requires master password:** No

**Output:**
```
[✓] Vext initialized at ~/.config/vext/vext.db
```

**Notes:**
- Safe to run more than once. AutoMigrate is idempotent — it never drops or overwrites existing data.

---

## `vext add <name>`

Stores a new secret under the given name.

```
vext add github
vext add visa-debit --type finance
```

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--type`, `-t` | `account` | Secret type: `account` or `finance` |

**For `--type account`:**
1. Prompts for username (visible).
2. Prompts for password (hidden).
3. Prompts for master password (hidden).

**For `--type finance`:**
1. Prompts for card number (visible), CVV (hidden), card PIN (hidden), expiry month (visible), expiry year (visible).
2. Prompts for optional bank fields: bank username (visible), bank password (hidden), virtual key (hidden), cellphone (visible). Leave blank to skip.
3. Prompts for master password (hidden).

**Requires master password:** Yes

**Output (success):**
```
[✓] Credential "github" saved.
```

**Output (duplicate name):**
```
[X] a credential named "github" already exists. Use `vext upd` to modify it.
```

**Notes:**
- `<name>` is case-sensitive. `github` and `GitHub` are two different records.
- Secret data is never passed as a flag.

---

## `vext get <name>`

Retrieves and displays a stored secret in plaintext.

```
vext get github
```

**What it does:**
1. Prompts for the master password (hidden input).
2. Opens the database and looks up the record by `<name>`.
3. Derives the encryption key using Argon2id + the stored Salt for that record.
4. Decrypts the payload using AES-256-GCM.
5. Dispatches on type to deserialize and print the fields.
6. Zeros all sensitive values in memory.

**Requires master password:** Yes

**Output (account):**
```
Service:  github
Username: bob@example.com
Password: hunter2
```

**Output (finance):**
```
Service:         visa-debit
Card Number:     4111111111111111
CVV:             ***
Card PIN:        ***
Expiry:          12/2027
Bank Username:   bob
Bank Password:   ***
Bank Virtual Key:***
Bank Cellphone:  +1234567890
```

**Output (not found):**
```
[X] no secret named "github" found
```

**Output (wrong master password or tampered data):**
```
[X] wrong master password or data corrupted
```

---

## `vext list`

Lists all stored secret names and their types.

```
vext list
```

**What it does:**
- Queries SQLite for `name`, `type`, and timestamps only (no encrypted data is touched).
- Formats and prints a table sorted alphabetically by name.

**Requires master password:** No

**Output:**
```
NAME             TYPE        CREATED
────────────────────────────────────────
github           account     2025-06-01
netflix          account     2025-06-02
visa-debit       finance     2025-06-03
────────────────────────────────────────
Total: 3 secrets.
```

---

## `vext upd <name>`

Updates the payload for an existing secret.

```
vext upd github
```

**What it does:**
1. Opens the database and looks up the existing record by `<name>` to determine its type.
2. Re-prompts for all fields of that type (same flow as `vext add` for that type).
3. Prompts for master password.
4. Re-encrypts the new payload with a fresh salt and nonce.
5. Replaces the stored record.

**Requires master password:** Yes

**Output (success):**
```
[✓] Secret "github" updated.
```

**Output (not found):**
```
[X] no secret named "github" found
```

**Notes:**
- The `--type` flag is not needed — the type is resolved from the existing record.
- A fresh salt and nonce are generated on every update. The previous ciphertext is fully replaced.

---

## `vext rm <name>`

Permanently deletes a stored secret.

```
vext rm github
```

**What it does:**
1. Prompts for confirmation: `Delete "github"? This cannot be undone (y/N)`.
2. On confirmation, opens the database and executes `DELETE FROM secrets WHERE name = ?`.
3. Prints a success message.

**Requires master password:** No

**Output (confirmed):**
```
[✓] Secret "github" deleted.
```

**Output (cancelled):**
```
Aborted.
```

**Output (not found):**
```
[X] no secret named "github" found
```

**Notes:**
- Deletion is permanent and irreversible.
- The confirmation prompt defaults to `N` (no). The user must explicitly type `y` to proceed.
- No master password required since no secret data is read.

---

## `vext export`

Exports an encrypted backup of all secrets to a portable file.

```
vext export
vext export --out ~/vext-backup.vxt
```

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--out`, `-o` | `vext-export-YYYYMMDD-HHMMSS.vxt` | Output file path |

**What it does:**
1. Prompts for master password.
2. Reads all secrets (with encrypted blobs) from the database.
3. Bundles them into a JSON structure and re-encrypts the entire bundle with the master password.
4. Writes the encrypted export file to disk with `0600` permissions.

**Requires master password:** Yes

**Output:**
```
[✓] Exported to vext-export-20260618-143022.vxt
```

**Notes:**
- The export file is itself encrypted with the master password and is safe to store in cloud storage or transfer over any channel.

---

## `vext import <file>`

Imports secrets from an encrypted export file.

```
vext import ~/vext-backup.vxt
```

**What it does:**
1. Prompts for master password.
2. Reads and decrypts the export file.
3. Inserts each record into the current database.
4. Skips records whose name already exists (no overwrite).

**Requires master password:** Yes

**Output:**
```
[✓] Imported 3 secret(s), skipped 1 duplicate(s).
```

**Output (wrong master password or corrupted file):**
```
[X] wrong master password or corrupted export file
```

---

## Phase 2 Commands (Planned)

### `vext gen`

Generates a cryptographically secure random password.

```
vext gen --length 24 --no-symbols
```

The generator implementation already exists in `shared/passgen/generator.go`. This feature only requires wiring a Cobra command to it.

---

### Shell Autocompletion

Cobra has native support for generating completion scripts:

```
vext completion bash
vext completion zsh
vext completion fish
vext completion powershell
```
