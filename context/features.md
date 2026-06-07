# Vext — Features

## Phase 1: MVP

The MVP is focused entirely on one thing: storing and retrieving account credentials (username + password) securely from the command line. Every feature in this phase serves that core loop.

---

### F-01 · Local Encrypted Storage

**What it is:** All secrets are stored in an encrypted SQLite database on the local machine. No cloud, no sync, no network calls at any point.

**Why it matters:** The database file is meaningless to anyone who obtains it without the master password. AES-256-GCM encryption + Argon2id key derivation ensure that even a stolen database file reveals nothing.

**Status:** Phase 1

---

### F-02 · Master Password Model

**What it is:** A single master password known only to the user is required to encrypt and decrypt all secrets. It is never stored anywhere — not in the database, not in a config file, not in memory beyond the duration of a single operation.

**Why it matters:** There is no backdoor. If you forget the master password, the data is permanently inaccessible. This is the correct security tradeoff for a local-first tool.

**Status:** Phase 1

---

### F-03 · Add Credential (`vext add`)

**What it is:** Interactively stores a new account credential (service name, username, password) under a unique name. All secret inputs are collected via hidden prompts.

**Status:** Phase 1

---

### F-04 · Retrieve Credential (`vext get`)

**What it is:** Looks up a stored credential by name, prompts for the master password, decrypts, and displays the username and password in the terminal.

**Status:** Phase 1

---

### F-05 · List Secrets (`vext list`)

**What it is:** Displays a formatted table of all stored secret names and their types. Does not require the master password. No encrypted data is touched.

**Status:** Phase 1

---

### F-06 · Delete Credential (`vext rm`)

**What it is:** Permanently removes a stored credential by name after a confirmation prompt. No master password required since no secret data is being read.

**Status:** Phase 1

---

### F-07 · Initialization (`vext init`)

**What it is:** Sets up the local environment (config directory + database file + schema) on first use. Safe to run multiple times.

**Status:** Phase 1

---

## Phase 2: Expansion

Phase 2 extends Vext beyond basic account credentials. The polymorphic database design built in Phase 1 means none of these additions require schema changes.

---

### F-08 · Password Generator (`vext gen`)

**What it is:** A command that generates cryptographically secure random passwords using `crypto/rand`. Supports configurable length and character set options (symbols, numbers, uppercase).

**Integration:** Can be used standalone (`vext gen --length 20`) or piped into `vext add` via a `--generate` flag.

**Why it matters:** Removes the human from the password creation step entirely. The generator will never produce weak passwords.

**Status:** Phase 2

---

### F-09 · Update Credential (`vext update`)

**What it is:** Updates the password (or any field) for an existing record. Requires the master password. More ergonomic than the current `rm` + `add` workaround.

**Status:** Phase 2

---

### F-10 · Encrypted Export/Import (`vext export` / `vext import`)

**What it is:** Exports the entire secrets database to a single encrypted file. The export file is encrypted with the master password and can be safely stored in cloud storage or transferred to another machine.

**Why it matters:** The current MVP has no disaster recovery. A corrupted or lost disk means all secrets are gone. Export/import solves this without introducing a cloud dependency.

**Status:** Phase 2

---

### F-11 · Shell Autocompletion

**What it is:** Cobra has native support for generating shell completion scripts for bash, zsh, fish, and PowerShell. This allows `vext get git<TAB>` to autocomplete to `vext get github`.

**Why it matters:** Dramatically reduces friction for day-to-day use. A tool you actually reach for is a tool that works.

**Status:** Phase 2

---

### F-12 · Bank Card Credentials (`type: credit`)

**What it is:** A new secret type for storing payment card data (card number, CVV, expiration date, PIN) plus optional bank portal fields. Uses the same polymorphic storage model — no schema changes required.

**Commands:**
- `vext add --type credit visa-debit`
- `vext get visa-debit`

**Required fields:** card number, security code (CVV), expiration month/year, PIN.
**Optional fields:** bank name, bank username, bank password, bank virtual key, cellphone, country code.

**Status:** Phase 2 ✓ (implemented in 0003)

---

### F-13 · Secure Notes (`type: note`)

**What it is:** A free-form encrypted text note. Useful for storing things like server credentials, recovery codes, or any secret that doesn't fit neatly into a username/password model.

**Status:** Phase 2

---

### F-14 · Clipboard Integration (`vext get --copy`)

**What it is:** Instead of printing the password to the terminal (visible to anyone nearby), copies it directly to the clipboard. Optionally clears the clipboard automatically after 30 seconds.

**Why it matters:** Improves day-to-day UX significantly while reducing the risk of shoulder surfing.

**Status:** Phase 2

---

## Feature Summary Table

| ID | Feature | Phase | Master Password Required |
|---|---|---|---|
| F-01 | Local Encrypted Storage | 1 | — (architectural) |
| F-02 | Master Password Model | 1 | — (architectural) |
| F-03 | Add Credential | 1 | Yes |
| F-04 | Retrieve Credential | 1 | Yes |
| F-05 | List Secrets | 1 | No |
| F-06 | Delete Credential | 1 | No |
| F-07 | Initialization | 1 | No |
| F-08 | Password Generator | 2 | No |
| F-09 | Update Credential | 2 | Yes |
| F-10 | Encrypted Export/Import | 2 | Yes |
| F-11 | Shell Autocompletion | 2 | No |
| F-12 | Bank Card Credentials | 2 | Yes |
| F-13 | Secure Notes | 2 | Yes |
| F-14 | Clipboard Integration | 2 | Yes |