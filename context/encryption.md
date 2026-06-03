# Vext — Encryption & Security Model

## Guiding Principle

Vext is designed so that **even if someone steals your database file, they gain nothing useful**. The encrypted blobs in SQLite are meaningless without the master password, and brute-forcing the master password is computationally infeasible by design.

---

## The Two-Stage Security Model

Every secret stored in Vext goes through two distinct cryptographic operations: **key derivation** and **authenticated encryption**. These are separate concerns solved by separate algorithms.

```
Master Password  ──►  [ Argon2id KDF ]  ──►  32-byte Encryption Key
                             ▲
                           Salt (random, stored in DB)

Plaintext Secret ──►  [ AES-256-GCM ]   ──►  Encrypted Blob
                             ▲
                           Key (from above) + Nonce (random, stored in DB)
```

---

## Stage 1: Key Derivation — Argon2id

### Why Not Use the Password Directly?

A master password like `MyCat2019!` is a human-friendly string. It has far less entropy than what encryption algorithms expect. Feeding it directly into AES would be insecure.

A Key Derivation Function (KDF) solves this by transforming any password into a key of exact length and high entropy.

### Why Argon2id?

Argon2id won the Password Hashing Competition in 2015 and is the current industry recommendation for password hashing and KDFs. Its key advantage is that it is deliberately **expensive in both time and memory**, which directly attacks the economics of brute-force attempts.

- A GPU farm that could crack a naive hash in hours would take **centuries** against a properly tuned Argon2id configuration.
- The `id` variant is a hybrid of Argon2i (side-channel resistant) and Argon2d (GPU-resistant), making it the safest general-purpose choice.

### The Salt

Each record in the database has its own randomly generated 16-byte Salt. The Salt is not secret — it is stored in plaintext in the database. Its purpose is to ensure that:

1. The same master password + different salt = a completely different derived key.
2. Two users with the same master password will produce different keys.
3. Precomputed attack tables (rainbow tables) are useless.

---

## Stage 2: Authenticated Encryption — AES-256-GCM

### Why Authenticated Encryption?

Standard encryption (like AES in CBC mode) only provides **confidentiality** — it hides the content. But it doesn't detect if someone has tampered with the encrypted bytes.

AES-GCM (Galois/Counter Mode) provides **Authenticated Encryption with Associated Data (AEAD)**, which means:
- It encrypts the data (confidentiality).
- It generates a short authentication tag that acts as a fingerprint of the ciphertext.
- On decryption, if a single byte has been altered — either by an attacker or by data corruption — the authentication tag check fails and decryption is refused entirely.

This means the database is tamper-evident. Any modification is detected.

### The Nonce

Each record also has its own randomly generated 12-byte Nonce (Number Used Once). Like the Salt, the Nonce is not secret and is stored in the database.

The critical rule: **a Nonce must never be reused with the same key**. Nonce reuse in GCM can catastrophically break confidentiality. By generating a fresh random Nonce per record, this risk is eliminated.

### What Gets Encrypted?

The entire secret payload is serialized to JSON first, then the complete JSON string is encrypted as a single blob. This means the encryption layer is agnostic to what kind of data it's protecting — it just sees bytes.

---

## What Happens with a Wrong Master Password?

When `vext get` is called with the wrong master password:

1. Argon2id derives a **different** 32-byte key (because a different password was used as input).
2. AES-GCM attempts to decrypt using this wrong key.
3. The authentication tag check fails immediately.
4. Vext returns: `[X] Error: master password incorrect or data corrupted.`

The user learns nothing about whether the password was close or what the actual data looks like. This is intentional.

---

## Memory Safety

Cryptographic keys and plaintext secrets exist in RAM only for the duration of an operation. Immediately after encryption or decryption completes, every `[]byte` holding sensitive data is overwritten with zeros before being released.

Go's garbage collector does not guarantee when memory is reclaimed or whether it can be read by another process in the interim. Zeroing manually is the only reliable mitigation.

---

## What Vext Does NOT Do (Intentional Scope Limits)

| Feature | Status | Reason |
|---|---|---|
| Cloud sync | Never | Defeats the local-first model |
| Biometric unlock | Not in scope | OS-level complexity, out of scope for MVP |
| Master password recovery | Never | No recovery = no backdoor. Lose the password, lose the data. |
| Clipboard integration | Phase 2 | Intentionally deferred for simplicity |
| Keyring/OS integration | Phase 2 | Feasible but adds complexity |

---

## Threat Model

| Threat | Mitigated? | How |
|---|---|---|
| Someone reads your `vext.db` file | ✅ Yes | All secrets are AES-GCM encrypted |
| Someone modifies your `vext.db` | ✅ Yes | GCM authentication tag detects tampering |
| Brute force against the database | ✅ Yes | Argon2id makes each attempt slow and memory-intensive |
| Shell history captures your passwords | ✅ Yes | Secrets are never passed as flags or arguments |
| Someone watches your screen during `vext get` | ❌ No | Plaintext is displayed; physical security is user's responsibility |
| Someone with terminal access runs `vext rm` | ⚠️ Partial | Deletion prompts for confirmation but does not require master password |