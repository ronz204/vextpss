# Vext — Features

## Phase 1: MVP (Implemented)

---

### F-01 · Local Encrypted Storage

**What it is:** All secrets are stored in an encrypted SQLite database on the local machine. No cloud, no sync, no network calls at any point.

**Why it matters:** The database file is meaningless to anyone who obtains it without the master password. AES-256-GCM encryption + Argon2id key derivation ensure that even a stolen database file reveals nothing.

**Status:** ✅ Implemented

---

### F-02 · Master Password Model

**What it is:** A single master password known only to the user is required to encrypt and decrypt all secrets. It is never stored anywhere — not in the database, not in a config file, not in memory beyond the duration of a single operation.

**Why it matters:** There is no backdoor. If you forget the master password, the data is permanently inaccessible. This is the correct security tradeoff for a local-first tool.

**Status:** ✅ Implemented

---

### F-03 · Add Credential (`vext add`)

**What it is:** Interactively stores a new secret (account or finance) under a unique name. All sensitive inputs are collected via hidden prompts. The `--type` flag selects the secret type.

**Status:** ✅ Implemented

---

### F-04 · Retrieve Credential (`vext get`)

**What it is:** Looks up a stored secret by name, prompts for the master password, decrypts, and displays the fields in the terminal.

**Status:** ✅ Implemented

---

### F-05 · List Secrets (`vext list`)

**What it is:** Displays a formatted table of all stored secret names and their types. Does not require the master password. No encrypted data is touched.

**Status:** ✅ Implemented

---

### F-06 · Delete Credential (`vext rm`)

**What it is:** Permanently removes a stored secret by name after a confirmation prompt. No master password required since no secret data is being read.

**Status:** ✅ Implemented

---

### F-07 · Initialization (`vext init`)

**What it is:** Sets up the local environment (config directory + database file + schema) on first use. Safe to run multiple times.

**Status:** ✅ Implemented

---

### F-08 · Update Credential (`vext upd`)

**What it is:** Updates the payload for an existing secret. Looks up the existing record to determine its type, re-prompts for all fields, re-encrypts, and persists the replacement. Requires the master password.

**Status:** ✅ Implemented

---

### F-09 · Encrypted Export/Import (`vext export` / `vext import`)

**What it is:** Exports the entire secrets database to a single encrypted `.vxt` file. The export file is encrypted with the master password and can be safely transferred to another machine. `vext import` reads the file, decrypts the bundle, and inserts records (skipping duplicates).

**Why it matters:** Disaster recovery without a cloud dependency. A backup can be stored on a USB drive or cloud storage — it's safe because it's encrypted.

**Status:** ✅ Implemented

---

### F-10 · Finance Credentials (`type: finance`)

**What it is:** A secret type for storing payment card data (card number, CVV, expiration, PIN) plus optional bank portal fields (username, password, virtual key, cellphone). Uses the same polymorphic storage model — no schema changes required.

**Status:** ✅ Implemented

---

## Phase 2: Expansion (Planned)

---

### F-11 · Password Generator (`vext gen`)

**What it is:** A command that generates cryptographically secure random passwords using `crypto/rand`. Supports configurable length and character set options (symbols, numbers, uppercase).

**Note:** The generator implementation (`shared/passgen/generator.go`) is already written. This feature only requires wiring a `vext gen` Cobra command to it.

**Integration:** Can be used standalone (`vext gen --length 20`) or piped into `vext add` via a `--generate` flag.

**Status:** 🔄 Phase 2 — generator implemented, command not wired

---

### F-12 · Shell Autocompletion

**What it is:** Cobra has native support for generating shell completion scripts for bash, zsh, fish, and PowerShell. This allows `vext get git<TAB>` to autocomplete to `vext get github`.

**Status:** 📋 Phase 2

---

### F-13 · Secure Notes (`type: note`)

**What it is:** A free-form encrypted text note. Useful for storing recovery codes, server credentials, or any secret that doesn't fit neatly into a username/password model.

**Status:** 📋 Phase 2

---

### F-14 · Clipboard Integration (`vext get --copy`)

**What it is:** Instead of printing the password to the terminal (visible to anyone nearby), copies it directly to the clipboard. Optionally clears the clipboard automatically after 30 seconds.

**Status:** 📋 Phase 2

---

## Feature Summary Table

| ID | Feature | Status | Master Password |
|---|---|---|---|
| F-01 | Local Encrypted Storage | ✅ Phase 1 | — (architectural) |
| F-02 | Master Password Model | ✅ Phase 1 | — (architectural) |
| F-03 | Add Credential (`vext add`) | ✅ Phase 1 | Yes |
| F-04 | Retrieve Credential (`vext get`) | ✅ Phase 1 | Yes |
| F-05 | List Secrets (`vext list`) | ✅ Phase 1 | No |
| F-06 | Delete Credential (`vext rm`) | ✅ Phase 1 | No |
| F-07 | Initialization (`vext init`) | ✅ Phase 1 | No |
| F-08 | Update Credential (`vext upd`) | ✅ Phase 1 | Yes |
| F-09 | Encrypted Export/Import | ✅ Phase 1 | Yes |
| F-10 | Finance Credentials | ✅ Phase 1 | Yes |
| F-11 | Password Generator (`vext gen`) | 🔄 Phase 2 | No |
| F-12 | Shell Autocompletion | 📋 Phase 2 | No |
| F-13 | Secure Notes | 📋 Phase 2 | Yes |
| F-14 | Clipboard Integration | 📋 Phase 2 | Yes |
