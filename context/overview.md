# Vext — Overview

## What is Vext?

Vext is a local-first CLI password manager built in Go. It was born out of a simple frustration: managing credentials in tools like Notion is clunky and not what they were designed for. Vext gives you a fast, minimal, and secure alternative that lives entirely on your machine.

No cloud. No sync. No accounts. Just a binary and an encrypted local database.

---

## Core Philosophy

**Local-first.** Your data never leaves your machine. There are no servers, no APIs, no third-party dependencies at runtime.

**Simple by design.** Vext does one thing: store and retrieve secrets securely. Every design decision favors clarity and correctness over feature bloat.

**Security without compromise.** Even though the tool is simple, the cryptographic foundation is production-grade. The encryption model is the same class of approach used by commercial password managers like Bitwarden and 1Password.

**Unix philosophy.** Each command does exactly one thing. The interface is predictable and composable.

---

## The Problem It Solves

| Pain Point | How Vext Addresses It |
|---|---|
| Notion/spreadsheets for passwords | Dedicated, purpose-built tool |
| Passwords stored in plain text | AES-GCM encryption at rest |
| Cloud-dependent tools | Fully offline, no account needed |
| Complex UX for simple tasks | Single binary, one command per action |
| Rigid data models | Polymorphic JSON payload per secret type |

---

## Target User

Vext is designed for developers and technical users who:
- Are comfortable with the terminal
- Want full control over where their data lives
- Don't want to depend on a SaaS product for something this personal
- Value understanding the security model of the tools they use

---

## Technology Stack

| Layer | Technology | Reason |
|---|---|---|
| Language | Go | Fast, single binary output, excellent crypto stdlib |
| CLI Framework | Cobra | Industry standard for Go CLIs, clean subcommand support |
| Database | SQLite (Pure Go) | Zero dependencies, local file, reliable |
| KDF | Argon2id | Winner of Password Hashing Competition, GPU-resistant |
| Cipher | AES-256-GCM | Authenticated encryption, detects tampering |

---

## Project Status

**Phase 1 (MVP):** Account credentials — email/username + password per service.

**Phase 2 (Planned):** Expanded secret types (bank cards, secure notes, SSH keys) using the same polymorphic storage model built into the foundation from day one.