# Vext — Contexto para Claude

Vext es un gestor de contraseñas CLI, local y minimalista escrito en Go.
Sin nube, sin cuenta, solo un binario y una base de datos SQLite cifrada en disco.

Módulo: `vextpss` · Entrypoint: `source/main.go`

## Comandos de desarrollo

```
go build ./...          # verificar compilación
go build -o vext ./source  # compilar el binario
```

No hay suite de tests todavía.

## Arquitectura en capas

```
source/main.go                     # wiring: deps + cobra root
source/funcs/<usecase>/            # casos de uso (lógica de aplicación)
source/secrets/core/               # dominio puro: Secret, Payload, interfaces
source/secrets/moon/               # payloads concretos: accounts, finances
source/shared/cyphers/aesgcm/      # implementación criptográfica
source/shared/storages/            # persistencia: GORM + SQLite
source/shared/terminal/            # I/O de terminal: Prompter, helpers
source/shared/memory/              # limpieza de memoria sensible
source/shared/passgen/             # generador de contraseñas (no conectado aún)
```

**Dirección de dependencia:** `funcs` → `secrets/core` ← `storages` y `cyphers`.
El dominio (`secrets/core`, `secrets/moon`) no importa nada de `shared/storages` ni `shared/cyphers`.

## Casos de uso (`source/funcs/`)

Cada subpaquete tiene:
- `command.go` — definición Cobra, solo wiring de flags y llamada a `run`
- `provider.go` — función `run(ctx, ..., deps)` con la lógica del caso de uso
- `collectors.go` — recolección interactiva de campos + registry map (solo `addsecret` y `updsecret`)

Los casos de uso son: `addsecret`, `updsecret`, `getsecrets`, `listsecrets`, `dropsecret`.

## Flujo de `update` (implementado con decrypt-first)

```
GetByName → ReadSecret("Master password") → Decrypt → collect(type, currentPlaintext, prompter)
  → Encrypt(newPlaintext, master) → Update
```

`collect` en `updsecret` usa `ReadLineOrKeep` / `ReadSecretOrKeep` / `ReadIntegerOrKeep`
para mostrar el valor actual y permitir conservarlo con Enter.
La master password se usa para descifrar y para re-cifrar — no cambia en un update.

> `context/usecases.md` describe el flujo anterior de update (sin decrypt-first). El flujo real es el de arriba.

## Tipos de secreto

Definidos en `source/secrets/core/shapes.go`:
- `core.TypeAccount` (`"account"`) → `moon/accounts.Account{Username, Password}`
- `core.TypeFinance` (`"finance"`) → `moon/finances.Finance{Card, Mobile}`

Para agregar un tipo nuevo:
1. Constante en `core/shapes.go` + entrada en `knownTypes`
2. Subpaquete en `moon/` con aggregate, composites y `sentinels.go`
3. Entrada en `collectors` de `addsecret/collectors.go` y `updsecret/collectors.go`
4. Entrada en el factory de `moon/factory.go`

## Convenciones clave

**`[]byte` vs `string` en payloads:** los campos sensibles (contraseñas, PINs, claves) son `[]byte`
para poder limpiarlos con `memory.Cleaner`. Los identificadores (`Username`, `Number`, `Cellphone`) son `string`.

**Limpieza de memoria:** siempre `defer memory.Cleaner(b)` sobre cualquier `[]byte` sensible en scope
(`plaintext`, `master`, `key` derivada).

**Errores de validación:** cada subpaquete de payload tiene `sentinels.go` con vars `Err...` exportadas.
Usar `errors.Is` para distinguirlos, nunca parsear mensajes.

**Sin comentarios** salvo que el WHY sea no obvio. Sin docstrings.

## Modelo criptográfico

Argon2id (KDF, 64 MB RAM, 3 iteraciones) → clave AES-256-GCM.
Salt y nonce aleatorios por operación, viajan en `core.Encrypted.Metadata` (opaco para el dominio).
`ErrDecryptionFailed` es deliberadamente ambiguo (contraseña incorrecta = datos corruptos).

## Contexto adicional

- `context/overview.md` — filosofía y tipos de secretos
- `context/concepts.md` — Argon2id, AES-GCM, limpieza de memoria (con código real)
- `context/modeling.md` — dominio: `Secret`, `Encrypted`, `Payload`, convención `[]byte`/`string`
- `context/database.md` — SQLite + GORM, `SecretRecord`, por qué una sola tabla
- `context/usecases.md` — flujo de cada comando (nota: `update` está desactualizado)
