# Vext — Workflows & Use Cases

## How to Read This Document

Each workflow describes a complete user action from command input to terminal output, including what happens internally at each step.

---

## Workflow 1: First-Time Setup

```
vext init
```

```
User runs `vext init`
        │
        ▼
Initialiser checks if ~/.config/vext/ exists
        ├── No  → Create directory with permissions 0700
        └── Yes → Continue
        │
        ▼
Check if vext.db exists
        ├── No  → Create vext.db, set permissions 0600
        │          Run AutoMigrate (CREATE TABLE IF NOT EXISTS)
        └── Yes → Run AutoMigrate (idempotent, no-op)
        │
        ▼
Print: [✓] Vext initialized at ~/.config/vext/vext.db
```

**Key point:** Safe to run multiple times. It will never overwrite existing data.

---

## Workflow 2: Saving a New Account Credential

```
vext add github
```

```
Receive argument: name = "github"
        │
        ▼
Collector.Payload("account")
  → Prompt: Username (visible)    → bob@example.com
  → Prompt: Password (hidden)     → hunter2
  → Marshal into AccountSecret JSON → {"username":"...","password":"<b64>"}
        │
        ▼
Collector.Master()
  → Prompt: Master password (hidden) → MyMasterKey!
        │
        ▼
storage.WithRepo opens vext.db
        │
        ▼
CreateSecretFunc.Run(dto)
  → Validate DTO (name, type, plaintext, master password all present)
  → AESGCMEncryptor.Encrypt(plaintext, masterPassword)
      → Generate 16 random bytes → Salt
      → Argon2id(masterPassword, Salt) → 32-byte Key
      → Generate 12 random bytes → Nonce
      → AES-256-GCM Encrypt(JSON, Key, Nonce) → Ciphertext
      → defer memory.Cleaner(key)
  → SecretRepository.Create(secret, ciphertext)
      → INSERT INTO secrets (name, type, salt, nonce, encrypted)
        VALUES ("github", "account", <salt>, <nonce>, <ciphertext>)
  → defer memory.Cleaner(masterPassword, plaintext)
        │
        ▼
storage.WithRepo closes vext.db
        │
        ▼
Print: [✓] Credential "github" saved.
```

---

## Workflow 3: Retrieving a Credential

```
vext get github
```

```
Receive argument: name = "github"
        │
        ▼
Collector.Master()
  → Prompt: Master password (hidden) → MyMasterKey!
        │
        ▼
storage.WithRepo opens vext.db
        │
        ▼
ObtainSecretFunc.Run(dto)
  → Validate DTO
  → SecretRepository.GetByName("github")
      → SELECT * FROM secrets WHERE name = "github"
      ├── Not found → return sentinel.ErrSecretNotFound
      └── Found → secret (metadata) + encrypted blob
  → AESGCMEncryptor.Decrypt(masterPassword, salt, nonce, ciphertext)
      → Argon2id(masterPassword, salt) → 32-byte Key
      → AES-256-GCM Decrypt(ciphertext, Key, nonce)
      ├── Auth tag FAILS → return sentinel.ErrDecryptionFailed
      └── Auth tag PASSES → JSON bytes
  → defer memory.Cleaner(masterPassword, key)
  → return ObtainSecretResult{Name, Type, Payload}
        │
        ▼
storage.WithRepo closes vext.db
        │
        ▼
Adapter dispatches on result.Type
  case "account": json.Unmarshal → AccountSecret
  → formatters.PrintAccount(name, secret)
  → defer memory.Cleaner(result.Payload)
        │
        ▼
Print:
  Service:  github
  Username: bob@example.com
  Password: hunter2
```

**Key point:** The error on wrong password is identical to the error on data tampering. An attacker learns nothing about which case occurred.

---

## Workflow 4: Browsing Stored Secrets

```
vext list
```

```
storage.WithRepo opens vext.db
        │
        ▼
RetrieveSecretsFunc.Run()
  → SecretRepository.ListAll()
      → SELECT id, name, type, created_at, updated_at FROM secrets ORDER BY name ASC
      (salt, nonce, encrypted are excluded from this query)
        │
        ▼
storage.WithRepo closes vext.db
        │
        ▼
formatters.PrintTabTable(secrets)

  NAME             TYPE        CREATED
  ──────────────────────────────────────
  github           account     2025-06-01
  netflix          account     2025-06-02
  visa-debit       finance     2025-06-03
  ──────────────────────────────────────
  Total: 3 secrets.
```

**Key point:** No master password. No decryption. The encrypted payload is never touched.

---

## Workflow 5: Updating a Credential

```
vext upd github
```

```
Receive argument: name = "github"
        │
        ▼
storage.WithRepo opens vext.db
        │
        ▼
SecretRepository.GetByName("github") → existing secret (to get its type)
        │
        ▼
Collector.Payload(existing.Type)
  → Re-prompts all fields for that type (hidden where sensitive)
        │
        ▼
Collector.Master()
  → Prompt: Master password (hidden)
        │
        ▼
UpdateSecretFunc.Run(dto)
  → Validate DTO
  → AESGCMEncryptor.Encrypt(plaintext, masterPassword) → new salt + nonce + ciphertext
  → SecretRepository.Update(secret, ciphertext)
      → UPDATE secrets SET salt=?, nonce=?, encrypted=? WHERE name = "github"
      ├── No rows affected → return sentinel.ErrSecretNotFound
      └── Updated
  → defer memory.Cleaner(masterPassword, plaintext)
        │
        ▼
storage.WithRepo closes vext.db
        │
        ▼
Print: [✓] Secret "github" updated.
```

**Key point:** The type is resolved from the existing record — the user does not need to specify `--type` when updating.

---

## Workflow 6: Deleting a Credential

```
vext rm github
```

```
Receive argument: name = "github"
        │
        ▼
Collector.Confirm("Delete \"github\"? This cannot be undone")
  → Prompt: Delete "github"? This cannot be undone (y/N)
  ├── N or Enter → Print: Aborted. Exit.
  └── y → Continue
        │
        ▼
storage.WithRepo opens vext.db
        │
        ▼
DeleteSecretFunc.Run(dto)
  → SecretRepository.Delete("github")
      → DELETE FROM secrets WHERE name = "github"
      ├── No rows affected → return sentinel.ErrSecretNotFound
      └── Deleted
        │
        ▼
storage.WithRepo closes vext.db
        │
        ▼
Print: [✓] Secret "github" deleted.
```

**Key point:** No decryption happens. No master password required.

---

## Workflow 7: Exporting a Backup

```
vext export --out ~/vext-backup.vxt
```

```
Collector.Master()
  → Prompt: Master password (hidden)
        │
        ▼
storage.WithRepo opens vext.db
        │
        ▼
ExportSecretsFunc.Run(dto)
  → SecretRepository.GetAll()
      → SELECT * FROM secrets ORDER BY name ASC
      → Returns []Credential (each with encrypted blob already stored)
  → Bundle all records into exportBundle{Version, ExportedAt, Records}
  → json.Marshal(bundle) → bundleBytes
  → AESGCMEncryptor.Encrypt(bundleBytes, masterPassword)
      → New salt + nonce + ciphertext (the bundle is re-encrypted end-to-end)
  → json.Marshal(exportFile{Salt, Nonce, Data}) → fileBytes
  → os.WriteFile("~/vext-backup.vxt", fileBytes, 0600)
  → defer memory.Cleaner(masterPassword, bundleBytes)
        │
        ▼
storage.WithRepo closes vext.db
        │
        ▼
Print: [✓] Exported to ~/vext-backup.vxt
```

**Key point:** The export file is a double-encrypted structure. Each record's payload is already encrypted in the database, and the entire bundle is re-encrypted with the master password for transport.

---

## Workflow 8: Importing a Backup

```
vext import ~/vext-backup.vxt
```

```
Collector.Master()
  → Prompt: Master password (hidden)
        │
        ▼
storage.WithRepo opens vext.db
        │
        ▼
ImportSecretsFunc.Run(dto)
  → os.ReadFile("~/vext-backup.vxt") → raw bytes
  → json.Unmarshal → exportFile{Salt, Nonce, Data}
  → AESGCMEncryptor.Decrypt(masterPassword, salt, nonce, data)
      ├── Fails → return sentinel.ErrDecryptionFailed
      └── plaintext → bundleBytes
  → json.Unmarshal(bundleBytes) → exportBundle{Records}
  → For each record:
      → SecretRepository.Create(secret, record.Encrypted)
      ├── ErrAlreadyExists → Skipped++
      └── Success → Imported++
  → defer memory.Cleaner(masterPassword, plaintext)
        │
        ▼
storage.WithRepo closes vext.db
        │
        ▼
Print: [✓] Imported 3 secret(s), skipped 1 duplicate(s).
```

---

## Use Case: Moving to a New Machine

```
# On old machine:
vext export --out ~/vext-backup.vxt

# Transfer the file (USB, secure channel, etc.)

# On new machine:
vext init
vext import ~/vext-backup.vxt
```

The export file is encrypted with the master password, so it can be transported without risk even over an insecure channel.
