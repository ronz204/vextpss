# CLAUDE.md — Context and Collaboration Guide

## Purpose of This Document

This document is a compass for you, Claude. It explains how knowledge about the **Vext** project is organized, what each folder means, how specifications should be interpreted, and how we work together to iterate on code.

Read this once at the beginning of each session. Then, whenever the user mentions "context" or "changes", you'll know exactly where to look and how to proceed.

---

## Project Structure Overview

```
vext/
├── .claude/
│   └── CLAUDE.md           ← You are here
├── context/                ← Complete specifications (rarely change)
│   ├── overview.md
│   ├── structure.md
│   ├── database.md
│   ├── encryption.md
│   ├── modeling.md
│   ├── features.md
│   ├── workflows.md
│   └── commands.md
├── changes/                ← Work iterations (plan → xact)
│   └── 0001-architecture-refactor-plan.md    (what will change)
│   └── 0001-architecture-refactor-xact.md    (what changed)
├── source/                 ← The actual Go code
│   ├── cmd/
│   ├── pkg/
│   ├── go.mod
│   └── go.sum
├── .gitignore
└── README.md
```

---

## The `context/` Folder

### What It Is

`context/` contains **the complete project specification**: architecture, cryptography, data models, planned features, user workflows, and command reference.

These documents are **the source of truth** for the project. They are not subject to frequent change. They are descriptions of how the system should behave.

### How to Read Them

1. **overview.md** — Start here. Understand what Vext is, why it exists, and what problem it solves.
2. **structure.md** — Code architecture: layers, design patterns, how the project is organized.
3. **database.md** — How SQLite is modeled, the schema, and the decision to use a polymorphic table.
4. **encryption.md** — The two-stage security model (Argon2id + AES-256-GCM).
5. **modeling.md** — The Go structs that represent each secret type (Account, Card, Note).
6. **features.md** — What gets built in Phase 1 (MVP) versus Phase 2 (expansion).
7. **workflows.md** — Complete user flows: how data travels through the system.
8. **commands.md** — CLI reference: what parameters each command accepts, what it prints, what errors it returns.

### When to Consult Context

- When you need to understand **why** a design decision was made.
- When you need to verify **what a command should do**.
- When you need to remember **how a feature works internally**.
- **Before writing code**, to ensure you understand the requirements.

---

## The `changes/` Folder

### What It Is

`changes/` contains the **iterative history** of the project: work plans, refactors, and the results of implementing them.

Each change has two documents:

1. **`NNNN-description-plan.md`** — The plan: what will change, why, and how.
2. **`NNNN-description-xact.md`** — The outcome: what actually changed, what worked, what didn't, and why.

### Naming Convention

- `NNNN` is a sequential number (0001, 0002, ...).
- `description` is a friendly slug name (e.g., `architecture-refactor`, `add-clipboard-support`).
- `plan` means the **pre-implementation planning** document.
- `xact` means the **actual execution** document (transaction, as in finance).

### Workflow

```
1. Identify a necessary change
        │
        ▼
2. Write NNNN-description-plan.md
   (What will be done, why, estimated steps)
        │
        ▼
3. Execute the plan
   (Write code, make changes)
        │
        ▼
4. Write NNNN-description-xact.md
   (What happened, what went well, what was different, lessons)
        │
        ▼
5. Update context/ if necessary
   (If behavior changed, reflect in overview/features/etc)
```

### What to Include in `plan.md`

```markdown
# NNNN — Description

## Objective

One or two clear sentences about what we're trying to achieve.

## Rationale

Why this change is necessary. What problem does it solve.

## Approach

How it will be done. What changes architecturally.

## Specific Changes

- File X: specific changes
- File Y: specific changes
- Context Z: update if behavior changed

## Testing

How we'll verify it works.

## Estimation

If relevant, how long we expect this to take.
```

### What to Include in `xact.md`

```markdown
# NNNN — Description [COMPLETED / PARTIAL / ABANDONED]

## Executive Summary

What was achieved. One or two sentences.

## What Went Well

- Aspect 1
- Aspect 2

## What Was Different

What didn't go exactly as planned and why.

## Changes Made

Exact list of what was modified.

## Lessons Learned

What we learned for next time.

## Next Steps

What remains pending, if anything.
```

### When to Consult Changes

- When the user says "check the plan" or "according to the execution".
- When you need to understand **what recently changed** in the project.
- When you need context on **why a decision was made**, not just in design but in execution.
- To see the **record of lessons learned** from previous iterations.

---

## How to Work With You (The User) Using This Structure

### Scenario 1: User Says "I Need to Implement X"

1. I read the relevant `context/` (e.g., `features.md` and `commands.md`).
2. I say: "I understand. X should behave like Y. Do you want a plan first?"
3. You say "yes" → I write `NNNN-X-plan.md`.
4. You review and say "proceed" or "change this".
5. I implement (write or modify code).
6. I write `NNNN-X-xact.md` with what happened.
7. If behavior changed the spec, I update `context/`.

### Scenario 2: User Says "Why is it Like That?"

1. I consult `context/` to find the design rationale.
2. I cite the relevant document (e.g., "According to `encryption.md`, ...").
3. If it's a recent change, I also check `changes/` for lessons learned.

### Scenario 3: User Says "Check the Plan"

1. I open `NNNN-description-plan.md`.
2. I understand what was intended.
3. I do what the plan says (or ask questions if unclear).

### Scenario 4: User Says "According to the Execution"

1. I open `NNNN-description-xact.md`.
2. I see what was actually achieved, what wasn't, and why.
3. I adjust my recommendations based on that.

---

## Important: Truth Lives in Context

If there's a conflict between what `context/` says and what happened in `changes/`, **context wins**.

`context/` is the "official" specification. If execution was different (because we learned something or something didn't work), that's documented in `xact`, but `context` gets updated to reflect reality.

Example:
- `features.md` says: "Phase 2 includes Bank Cards".
- `0001-xact.md` says: "We decided not to do it because..."
- I update `features.md` to reflect the actual decision.

---

## Your Role as User

1. **Keep `context/` updated** when the specification changes.
2. **Write `plan.md` before requesting big changes** — it forces clear thinking.
3. **Write `xact.md` afterward** — it creates a learning record.
4. **Ask questions** if you see me doing something that contradicts `context/`.
5. **Be specific** — if you say "implement X", ideally cite which `context/` document defines X.

---

## My Role as Claude

1. **I assume `context/` is truth** until you tell me otherwise.
2. **I consult `context/` before writing code** to ensure I understand requirements.
3. **I document each change in `changes/`** to create an auditable record.
4. **I ask questions** if there's ambiguity between documents.
5. **I suggest updates to `context/`** if I see something has changed.

---

## Key Points Summary

| Aspect | Context | Changes |
|---|---|---|
| **Contains** | Complete specification | Iteration history |
| **Change frequency** | Infrequent (only if spec changes) | Frequent (per work item) |
| **Purpose** | Source of truth | Learning record |
| **When to consult** | Always when starting something | When you need lessons |
| **Authority** | High — defines the system | Medium — explains decisions |

---

## Writing Convention

All documents in this project (including `plan.md` and `xact.md`) follow the same style:

- **Clean and structured Markdown**.
- **Clear hierarchical headings** (# level 1, ## level 2).
- **Detailed but concise explanations** — not verbose.
- **Code examples when applicable**, with properly formatted code blocks.
- **Tables for comparisons** or option lists.
- **ASCII diagrams** for complex flows.
- **Professional but accessible tone** — technical but not pedantic.

---

## Final Point

This document (`CLAUDE.md`) is **your compass**. If at any point you don't know where to look or how to proceed, reread it. It's designed to be your guide when there's ambiguity.

Welcome to Vext.
