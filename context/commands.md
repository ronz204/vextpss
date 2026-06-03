# Vext — Commands Reference

## Global Rules

- **Secrets are never passed as flags.** All sensitive inputs (service passwords, master password) are collected via hidden interactive prompts. This prevents them from appearing in shell history.
- **The master password is never stored.** It is requested on-demand and discarded from memory immediately after use.
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
- Runs the `CREATE TABLE IF NOT EXISTS` migration to set up the schema.
- Sets file permissions to `0600` on the database file.

**Requires master password:** No

**Output:**
```
[✓] Vext initialized at ~/.config/vext/vext.db
```

**Notes:**
- Safe to run more than once. If the database already exists, it does nothing.
- Should be the first command documented in the README under "Getting Started".

---

## `vext add <name>`

Stores a new secret under the given name.

```
vext add github
vext add protonmail
```

**What it does:**
1. Accepts `<name>` as the service identifier (must be unique).
2. Prompts for the username (visible input).
3. Prompts for the service password (hidden input).
4. Prompts for the master password (hidden input).
5. Derives an encryption key from the master password using Argon2id + a freshly generated Salt.
6. Encrypts the payload JSON using AES-256-GCM + a freshly generated Nonce.
7. Persists the record to SQLite.
8. Zeros out sensitive values in memory.

**Requires master password:** Yes

**Output (success):**
```
[✓] Credential "github" saved.
```

**Output (duplicate name):**
```
[X] Error: a credential named "github" already exists. Use `vext update` to modify it.
```

**Notes:**
- `<name>` is case-sensitive. `github` and `GitHub` would be two different records.
- The username prompt is visible (not hidden) because usernames are not sensitive.

---

## `vext get <name>`

Retrieves and displays a stored secret in plaintext.

```
vext get github
```

**What it does:**
1. Looks up the record by `<name>` in SQLite.
2. If not found, exits with an error.
3. Prompts for the master password (hidden input).
4. Derives the encryption key using Argon2id + the stored Salt for that record.
5. Attempts to decrypt the payload using AES-256-GCM.
6. If decryption succeeds, prints the fields to the terminal.
7. If decryption fails (wrong password or tampered data), prints a generic error.
8. Zeros out sensitive values in memory.

**Requires master password:** Yes

**Output (success):**
```
Service:  github
Username: bob@example.com
Password: hunter2
```

**Output (not found):**
```
[X] Error: no credential named "github" found.
```

**Output (wrong master password or tampered data):**
```
[X] Error: master password incorrect or data corrupted.
```

**Notes:**
- The password is displayed in plaintext. Physical screen security is the user's responsibility.
- A future Phase 2 version may add a `--copy` flag to send the password to the clipboard instead.

---

## `vext list`

Lists all stored secret names and their types.

```
vext list
```

**What it does:**
- Queries SQLite for `name` and `type` columns only (no encrypted data is touched).
- Formats and prints a table sorted alphabetically by name.

**Requires master password:** No

**Output:**
```
Your stored secrets:
──────────────────────────────
NAME             TYPE
──────────────────────────────
github           account
netflix          account
protonmail       account
──────────────────────────────
Total: 3 secrets.
```

**Notes:**
- Usernames and passwords are never shown. Only names and types.
- No master password needed — this operates entirely on non-sensitive metadata.

---

## `vext rm <name>`

Permanently deletes a stored secret.

```
vext rm github
```

**What it does:**
1. Verifies the record exists.
2. Prompts for confirmation: `Are you sure you want to delete "github"? (y/N)`
3. On confirmation, executes `DELETE FROM secrets WHERE name = ?`.
4. Prints a success message.

**Requires master password:** No

**Output (confirmed):**
```
[✓] Credential "github" deleted.
```

**Output (cancelled):**
```
Aborted.
```

**Output (not found):**
```
[X] Error: no credential named "github" found.
```

**Notes:**
- Deletion is permanent and irreversible. There is no recycle bin or undo.
- The master password is not required to delete because no secret data is being read. This is a known tradeoff. A future version may require it for added protection.
- The confirmation prompt defaults to `N` (no). The user must explicitly type `y` to proceed.

---

## Phase 2 Commands (Planned)

These commands are planned for a future release and are documented here for design reference.

### `vext gen`

Generates a cryptographically secure random password.

```
vext gen --length 24 --no-symbols
vext add twitter --generate
```

Uses `crypto/rand` (never `math/rand`). Can be piped directly into `vext add`.

---

### `vext update <name>`

Updates the password for an existing credential. Requires master password.

```
vext update github
```

---

### `vext export`

Exports an encrypted backup of the entire database to a portable file.

```
vext export --out ~/backup.vext
```

The output file is encrypted with the master password and can be imported on another machine.

---

### `vext import`

Imports a backup file created by `vext export`.

```
vext import ~/backup.vext
```