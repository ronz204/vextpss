# Vext — Workflows & Use Cases

## How to Read This Document

Each workflow describes a complete user action from command input to terminal output, including what happens internally at each step. The goal is to understand the full data journey, not just the surface behavior.

---

## Workflow 1: First-Time Setup

**Trigger:** User installs Vext for the first time.

```
vext init
```

```
User runs `vext init`
        │
        ▼
Check if ~/.config/vext/ exists
        │
        ├── No  → Create directory with permissions 0700
        │
        └── Yes → Continue
        │
        ▼
Check if vext.db exists
        │
        ├── No  → Create vext.db, set permissions 0600
        │          Run CREATE TABLE IF NOT EXISTS migration
        │
        └── Yes → Run CREATE TABLE IF NOT EXISTS (safe, idempotent)
        │
        ▼
Print: [✓] Vext initialized at ~/.config/vext/vext.db
```

**Key point:** This command is safe to run multiple times. It will never overwrite existing data.

---

## Workflow 2: Saving a New Credential

**Trigger:** User wants to store credentials for a service.

```
vext add github
```

```
Receive argument: name = "github"
        │
        ▼
Prompt: Username (visible input)
  → user types: bob@example.com
        │
        ▼
Prompt: Password (hidden input — term.ReadPassword)
  → user types: hunter2
        │
        ▼
Prompt: Master Password (hidden input — term.ReadPassword)
  → user types: MyMasterKey!
        │
        ▼
Generate 16 random bytes → Salt
Generate 12 random bytes → Nonce
        │
        ▼
Argon2id(MyMasterKey!, Salt) → 32-byte Encryption Key
        │
        ▼
Build JSON payload:
  {"username":"bob@example.com","password":"hunter2"}
        │
        ▼
AES-256-GCM Encrypt(JSON payload, Key, Nonce)
  → EncryptedPayload (opaque bytes)
        │
        ▼
Zero out: master password bytes, plaintext password bytes, key bytes
        │
        ▼
INSERT INTO secrets (name, type, salt, nonce, encrypted_payload)
  VALUES ("github", "account", <salt>, <nonce>, <encrypted_payload>)
        │
        ▼
Print: [✓] Credential "github" saved.
```

---

## Workflow 3: Retrieving a Credential

**Trigger:** User needs to log into a service and wants to look up their credentials.

```
vext get github
```

```
Receive argument: name = "github"
        │
        ▼
SELECT salt, nonce, encrypted_payload, type FROM secrets WHERE name = "github"
        │
        ├── No rows → Print error: no credential named "github" found. Exit.
        │
        └── Found → Load salt, nonce, encrypted_payload into memory
        │
        ▼
Prompt: Master Password (hidden input)
  → user types: MyMasterKey!
        │
        ▼
Argon2id(MyMasterKey!, stored_salt) → 32-byte Key
        │
        ▼
AES-256-GCM Decrypt(encrypted_payload, Key, stored_nonce)
        │
        ├── Auth tag FAILS (wrong password or tampered data)
        │     → Zero out all sensitive bytes
        │     → Print: [X] Error: master password incorrect or data corrupted.
        │     → Exit
        │
        └── Auth tag PASSES → JSON payload bytes
        │
        ▼
switch type {
  case "account": unmarshal into AccountPayload
}
        │
        ▼
Zero out: master password bytes, key bytes, JSON bytes
        │
        ▼
Print:
  Service:  github
  Username: bob@example.com
  Password: hunter2
```

**Key point:** The error message on wrong password is identical to the error on data tampering. An attacker learns nothing about which case occurred.

---

## Workflow 4: Browsing Stored Secrets

**Trigger:** User doesn't remember the exact name they used for a service.

```
vext list
```

```
SELECT name, type FROM secrets ORDER BY name ASC
        │
        ▼
Format into table:
  ──────────────────────────
  NAME             TYPE
  ──────────────────────────
  github           account
  netflix          account
  protonmail       account
  ──────────────────────────
  Total: 3 secrets.
```

**Key point:** No master password. No decryption. The encrypted payload is never touched. This is a read of metadata only.

---

## Workflow 5: Deleting a Credential

**Trigger:** User no longer uses a service and wants to clean up.

```
vext rm github
```

```
Receive argument: name = "github"
        │
        ▼
SELECT id FROM secrets WHERE name = "github"
        │
        ├── No rows → Print: [X] Error: no credential named "github" found. Exit.
        │
        └── Found → Continue
        │
        ▼
Prompt: Are you sure you want to delete "github"? (y/N)
        │
        ├── User types N or presses Enter → Print: Aborted. Exit.
        │
        └── User types y → Continue
        │
        ▼
DELETE FROM secrets WHERE name = "github"
        │
        ▼
Print: [✓] Credential "github" deleted.
```

**Key point:** No decryption happens. The encrypted payload is deleted without ever being read.

---

## Use Case: Rotating a Password

When a service forces a password change, the current MVP flow is:

```
vext rm github         # Delete the old record
vext add github        # Re-add with the new password
```

This two-step process is intentional for the MVP. A dedicated `vext update <name>` command is planned for Phase 2 to make this atomic and cleaner.

---

## Use Case: Moving to a New Machine (Phase 2)

With the `vext export` / `vext import` commands planned for Phase 2:

```
# On old machine:
vext export --out ~/vext-backup.enc

# Transfer the file (USB, secure channel, etc.)

# On new machine:
vext init
vext import ~/vext-backup.enc
```

The export file is itself encrypted with the master password, so it can be transported without risk even over an insecure channel.