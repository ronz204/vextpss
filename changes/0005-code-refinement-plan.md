# 0005 — Refinamiento de Código: Semántica, Calidad y Portabilidad

## Objetivo

Realizar una tercera pasada sobre el codebase — esta vez a nivel de código, no de arquitectura de capas — para corregir semántica confusa en paquetes y archivos, eliminar duplicación, reducir acoplamiento a tecnologías específicas (SQLite), y aplicar principios SOLID donde el beneficio justifica el costo. El resultado debe ser un codebase donde cualquier nombre — carpeta, archivo, función — comunique su propósito sin ambigüedad.

---

## Rationale

El refactoring 0004 resolvió el problema de capas (interfaces en el lugar correcto, codec centralizado, guards de nil). Lo que permanece son problemas de **expresividad**: nombres que no dicen lo que hacen (`tokens`, `helpers`, `shared`, `dal`), duplicación estructural que cresce con cada comando nuevo, y dos piezas de deuda técnica documentadas en el xact de 0004.

### Hallazgos por Categoría

#### Categoría A — Semántica de Paquetes (Alto Impacto)

| # | Paquete Actual | Problema | Severidad |
|---|---|---|---|
| A1 | `pkg/tokens/` | `tokens` implica JWT / tokens de sesión / API keys. El contenido es un encriptador AES-GCM con Argon2id. El nombre activamente engaña. | Alta |
| A2 | `pkg/apps/` | El prefijo `pkg/` es un artefacto histórico de Go (<2015). No agrega valor semántico. En Go moderno el directorio ya comunica visibilidad. Mismo problema en `pkg/configs/` y `pkg/shared/`. | Alta |
| A3 | `dal/` | Acrónimo técnico de tres letras. Un desarrollador nuevo no sabe si es "data", "domain", "deployment" o qué. La carpeta contiene implementaciones de persistencia: `storage/` dice exactamente eso. | Media |
| A4 | `cmd/helpers/` | "helpers" es el nombre más genérico posible — el "misc" del mundo de paquetes. El contenido es exclusivamente UI/presentación: prompts interactivos, colectores de inputs, formatters. `cmd/ui/` comunica ese propósito. | Media |
| A5 | `pkg/shared/` | "shared" no dice nada sobre qué contiene. Tiene utilidades de memoria (`Zero`), tiempo (`Now`), generación de sal criptográfica (`GenerateSalt`) y generación de contraseñas (`GeneratePassword`) — cuatro responsabilidades que merecen mejor agrupación. | Media |
| A6 | `tests/mocks/` | Ubicación inusual. En Go, los helpers de test se ponen en `testutil/` al nivel del módulo (o en `internal/testutil`). La carpeta `tests/` con un solo subdirectorio es redundante. | Baja |

#### Categoría B — Semántica de Archivos (Medio Impacto)

| # | Archivo Actual | Problema |
|---|---|---|
| B1 | `dal/schemas.go` | "schema" es un concepto de base de datos abstracto. El archivo define `SecretRecord`, un modelo GORM. `record.go` es más preciso. |
| B2 | `core/secrets/account_secret.go` | El sufijo `_secret` es redundante: el tipo se llama `AccountSecret`. Por convención Go, el nombre del archivo refleja el tipo principal. `account.go` es suficiente. |
| B3 | `core/secrets/credit_secret.go` | Ídem. → `credit.go`. |
| B4 | `pkg/configs/app_config.go` | El prefijo `app_` es redundante dentro de un paquete ya llamado `configs`. → `config.go`. |
| B5 | `pkg/tokens/aes_gcm_encryptor.go` | El sufijo `_encryptor` es redundante si el paquete se renombra a `crypto`. → `aes_gcm.go`. |
| B6 | `pkg/shared/utils.go` | "utils" no dice nada. El contenido real son operaciones de memoria (`Zero`, `Now`) y una función de sal criptográfica. → `mem.go`. |
| B7 | `dal/repos/sqlite_repository.go` | Si la carpeta padre se renombra a `storage/sqlite/`, el prefijo `sqlite_` en el filename se vuelve redundante (ya está en la carpeta). → `repository.go`. |

#### Categoría C — Portabilidad de Base de Datos (Medio Impacto)

| # | Problema | Archivo | Impacto al cambiar de SQLite a PostgreSQL |
|---|---|---|---|
| C1 | `isUniqueConstraintError` usa `strings.Contains(err.Error(), "UNIQUE constraint failed")` — string literal del mensaje de error interno de SQLite. | `dal/repos/sqlite_repository.go` | Roto silenciosamente. PostgreSQL produce `"duplicate key value violates unique constraint"`. |
| C2 | `SecretRecord` usa `gorm:"type:blob"` en tres campos. | `dal/schemas.go` | En PostgreSQL, el tipo equivalente es `bytea`. GORM no mapea `blob` a `bytea` automáticamente. |
| C3 | `dal/db.go` importa `github.com/glebarez/sqlite` directamente. | `dal/db.go` | El driver está hardcodeado en el punto de conexión. Cambiar de DB requiere modificar este archivo. |

La pregunta clave del usuario: "¿si hoy usamos SQLite y luego queremos usar PostgreSQL, el cambio sería fácil?". Respuesta honesta:

**Lo que ya es fácil:**
- La interfaz `core.SecretRepository` aísla completamente las queries GORM. Agregar `storage/postgres/repository.go` no requiere tocar use cases ni handlers.
- GORM abstrae casi todas las diferencias de queries.
- El punto de composición en `main.go` es único — cambiar de `sqlite.NewSQLiteRepository` a `postgres.NewPostgresRepository` es una línea.

**Lo que NO es fácil hoy:**
- C1: El error de unicidad se detecta por string; falla silenciosamente en PostgreSQL.
- C2: Los tags GORM `type:blob` no funcionan en PostgreSQL. Habría que descubrirlo en runtime.
- C3: El driver está hardcodeado en `db.go` — no hay una interfaz de conexión.

#### Categoría D — Calidad de Código (Medio Impacto)

| # | Problema | Archivos Afectados |
|---|---|---|
| D1 | Patrón de error handling repetido en 5+ handlers: `IsNotFound` → mensaje, `IsDomainError` → mensaje, else → "unexpected error". Son 7 líneas idénticas copy-pasteadas. | `add_handler.go`, `get_handler.go`, `rm_handler.go`, `update_handler.go`, `export_handler.go`, `import_handler.go` |
| D2 | Lógica de negocio en handler: "si username vacío, mantener el actual" vive en `update_handler.go` con type assertions sobre `*secrets.AccountSecret`. El handler conoce detalles del dominio. | `cmd/handlers/update_handler.go:84-90` |
| D3 | `MockPrompter` en código de producción: `cmd/helpers/prompter.go` exporta un test double que se compila en el binario final. | `cmd/helpers/prompter.go:57-72` |
| D4 | `GenerateSalt` en `pkg/shared/` es usada únicamente por `pkg/tokens/aes_gcm_encryptor.go`. Es una función criptográfica que pertenece al paquete que la usa. | `pkg/shared/utils.go`, `pkg/tokens/aes_gcm_encryptor.go` |
| D5 | Switch statements en `codec.go` y `collector.go` violan el principio Open-Closed: agregar un nuevo tipo de secreto requiere modificar ambos archivos. | `pkg/apps/codec.go`, `cmd/helpers/collector.go` |

#### Categoría E — Observaciones Menores

| # | Observación |
|---|---|
| E1 | Los handlers reciben tipos concretos (`*apps.StoreSecretUC`) en lugar de interfaces. Esto los acopla a la implementación. Para el tamaño actual del proyecto es aceptable, pero al crecer dificultará los tests de handler sin infraestructura. |
| E2 | `Initialiser.appDir` se recibe en el constructor pero nunca se usa — solo `dbPath`. El campo es deuda del diseño original. |
| E3 | `dal/repository.go` y `pkg/tokens/encryptor.go` quedaron vacíos después del 0004. Son archivos zombies que confunden. |

---

## Approach

Los cambios se dividen en dos fases:

**Fase Estructural** — renombres de carpetas, paquetes y archivos. Son cambios mecánicos pero tienen amplio impacto en imports. Se hacen todos juntos en un commit para que el repo no quede en estado intermedio.

**Fase de Código** — mejoras de calidad sin cambios de carpetas. Se hacen en orden de dependencia.

---

## Cambios Específicos

### Cambio 1 — Reorganización de Paquetes (Structural)

Nueva estructura propuesta:

```
source/
├── main.go
├── cmd/
│   ├── root.go
│   ├── handlers/             (sin cambio)
│   └── ui/                   ← RENOMBRADO desde cmd/helpers/
│       ├── prompter.go       ← CLIPrompter solamente (sin MockPrompter)
│       ├── collector.go
│       ├── account_collector.go
│       ├── credit_collector.go
│       └── formatter.go
├── core/                     (sin cambio)
│   ├── errors.go
│   ├── secret.go
│   ├── repository.go
│   ├── encryptor.go
│   └── secrets/
│       ├── account.go        ← RENOMBRADO desde account_secret.go
│       ├── credit.go         ← RENOMBRADO desde credit_secret.go
│       └── types.go
├── app/                      ← RENOMBRADO desde pkg/apps/
│   ├── codec.go
│   ├── init_storage_uc.go
│   ├── store_secret_uc.go
│   ├── retrieve_secret_uc.go
│   ├── list_secrets_uc.go
│   ├── delete_secret_uc.go
│   ├── update_secret_uc.go
│   ├── export_secrets_uc.go
│   └── import_secrets_uc.go
├── storage/                  ← RENOMBRADO desde dal/
│   ├── db.go
│   ├── record.go             ← RENOMBRADO desde schemas.go
│   ├── initialiser.go
│   └── sqlite/               ← RENOMBRADO desde dal/repos/
│       └── repository.go     ← RENOMBRADO desde sqlite_repository.go
├── crypto/                   ← RENOMBRADO desde pkg/tokens/
│   └── aes_gcm.go            ← RENOMBRADO desde aes_gcm_encryptor.go
├── config/                   ← RENOMBRADO desde pkg/configs/
│   └── config.go             ← RENOMBRADO desde app_config.go
├── shared/                   ← RENOMBRADO desde pkg/shared/ (fuera de pkg/)
│   ├── mem.go                ← RENOMBRADO desde utils.go (sin GenerateSalt)
│   └── passgen.go            ← RENOMBRADO desde generator.go
└── testutil/                 ← RENOMBRADO desde tests/mocks/
    ├── mock_repo.go
    ├── mock_encryptor.go
    ├── mock_initialiser.go
    └── mock_prompter.go      ← MOVIDO desde cmd/helpers/prompter.go
```

**Archivos eliminados:**
- `dal/repository.go` — vacío desde 0004
- `pkg/tokens/encryptor.go` — vacío desde 0004
- `cmd/helpers/collectors.go` — vacío desde 0004

**Tabla de renombres de import:**

| Import Actual | Import Nuevo |
|---|---|
| `vextpss/source/pkg/apps` | `vextpss/source/app` |
| `vextpss/source/pkg/tokens` | `vextpss/source/crypto` |
| `vextpss/source/pkg/configs` | `vextpss/source/config` |
| `vextpss/source/pkg/shared` | `vextpss/source/shared` |
| `vextpss/source/dal` | `vextpss/source/storage` |
| `vextpss/source/dal/repos` | `vextpss/source/storage/sqlite` |
| `vextpss/source/cmd/helpers` | `vextpss/source/cmd/ui` |
| `vextpss/source/tests/mocks` | `vextpss/source/testutil` |

**Nota de paquetes Go:** Cuando se renombra una carpeta, el `package` declaration del archivo debe coincidir con el nuevo nombre. Por ejemplo, `package apps` → `package app`, `package tokens` → `package crypto`, `package helpers` → `package ui`, `package dal` → `package storage`, `package repos` → `package sqlite`, `package configs` → `package config`, `package shared` permanece igual.

---

### Cambio 2 — Fix `isUniqueConstraintError` (C1)

**Problema:** String matching frágil y SQLite-específico.

**Solución:** `modernc.org/sqlite` ya está en `go.mod` como dependencia transitiva. Usar `errors.As` para extraer el tipo nativo y comparar códigos numéricos:

```go
import (
    "errors"
    "strings"
    
    sqlite3 "modernc.org/sqlite/lib"
    sqliteErr "modernc.org/sqlite"
)

func isUniqueConstraintError(err error) bool {
    var sqlErr *sqliteErr.Error
    if errors.As(err, &sqlErr) {
        return sqlErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE
    }
    // Fallback defensivo por si el tipo cambia en versiones futuras.
    return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
```

**Beneficio inmediato:** Robusto ante cambios de mensaje en el driver.
**Beneficio futuro:** Al agregar `storage/postgres/repository.go`, se puede implementar su propia `isUniqueConstraintError` con el código de error de `pq` o `pgx` sin afectar el SQLite repo.

---

### Cambio 3 — Documentar Variabilidad de GORM Tags (C2, C3)

No se cambia el código ahora (cambiar los tipos GORM sin una migración funcional sería riesgoso), pero se documenta el punto de variabilidad con comentarios precisos:

**En `storage/record.go`:**
```go
// SQLite uses blob; change to "type:bytea" when targeting PostgreSQL.
Salt      []byte `gorm:"not null;type:blob"`
Nonce     []byte `gorm:"not null;type:blob"`
Encrypted []byte `gorm:"not null;type:blob"`
```

**En `storage/db.go`:**
```go
// Open opens a GORM connection to the SQLite database at path.
// To support a different driver, replace sqlite.Open() with the target driver
// and update SecretRecord's GORM type tags accordingly.
func Open(path string) (*gorm.DB, error) {
```

Esto es suficiente para que un desarrollador sepa exactamente qué cambiar. Crear una interfaz de conexión abstracta ahora sería over-engineering para un CLI local-first.

---

### Cambio 4 — Extraer `handleErr` en handlers (D1)

**Problema:** 7 líneas de manejo de errores están copy-pasteadas en 6 handlers:
```go
if core.IsNotFound(err) { ... }
if core.IsDomainError(err) { ... }
fmt.Println("[X] An unexpected error occurred...")
```

**Solución:** Función auxiliar en `cmd/handlers/errors.go` (o extender `guard.go`):

```go
// printErr prints a user-facing error message and returns true if err != nil.
// notFoundMsg is printed when err is ErrSecretNotFound; "" falls back to the domain message.
func printErr(err error, notFoundMsg string) bool {
    if err == nil {
        return false
    }
    if notFoundMsg != "" && core.IsNotFound(err) {
        fmt.Fprintf(os.Stderr, "[X] %s\n", notFoundMsg)
        return true
    }
    if core.IsDomainError(err) {
        fmt.Fprintf(os.Stderr, "[X] %s\n", err)
        return true
    }
    fmt.Fprintln(os.Stderr, "[X] An unexpected error occurred. Please try again.")
    return true
}
```

**Uso en handlers:**
```go
// Antes (7 líneas):
if core.IsNotFound(err) {
    fmt.Printf("[X] Error: no credential named %q found.\n", name)
    return nil
}
if core.IsDomainError(err) {
    fmt.Printf("[X] %s\n", err)
    return nil
}
fmt.Println("[X] An unexpected error occurred. Please try again.")
return nil

// Después (1 línea):
if printErr(err, fmt.Sprintf("no credential named %q found", name)) {
    return nil
}
```

---

### Cambio 5 — Mover Lógica de Merge de Username al Dominio (D2)

**Problema:** `update_handler.go:84-90` contiene una regla de negocio ("si el usuario deja username en blanco, conservar el valor actual") con type assertions directas sobre `*secrets.AccountSecret`. El handler debe ser agnóstico de los tipos concretos de payload.

**Solución:** Agregar método `MergeFrom(current core.SecretPayload)` a la interfaz `SecretPayload`:

```go
// core/secret.go
type SecretPayload interface {
    GetType() string
    Validate() error
    // MergeFrom fills any zero-value fields from current. Used during update.
    MergeFrom(current SecretPayload)
}
```

**Implementación en `core/secrets/account.go`:**
```go
func (a *AccountSecret) MergeFrom(current core.SecretPayload) {
    if cur, ok := current.(*AccountSecret); ok && a.Username == "" {
        a.Username = cur.Username
    }
}
```

**Implementación en `core/secrets/credit.go`:**
```go
func (c *CreditSecret) MergeFrom(current core.SecretPayload) {
    // Credit cards replace all fields on update; no merge needed.
}
```

**En `update_handler.go`**, reemplazar el bloque de type assertion por:
```go
newPayload.MergeFrom(retrieveResp.Payload)
```

**Beneficios:**
- Handler ya no importa `core/secrets` para type assertions.
- La regla de merge vive en el tipo de dominio que la define.
- Agregar `NoteSecret` en Phase 2 solo requiere implementar `MergeFrom` en `note.go`.

---

### Cambio 6 — Mover `MockPrompter` a `testutil/` (D3)

**Problema:** `cmd/helpers/prompter.go` exporta `MockPrompter`, un test double que se incluye en el binario de producción.

**Solución:** Mover `MockPrompter` a `testutil/mock_prompter.go`. Los handler tests actualizan su import de `helpers.MockPrompter` a `testutil.MockPrompter`.

```go
// testutil/mock_prompter.go
package testutil

type MockPrompter struct {
    LineResponse    string
    PasswordBytes   []byte
    ConfirmResponse bool
    Err             error
}
// ... mismos métodos
```

**Nota:** `cmd/ui/prompter.go` queda con solo `Prompter` (interface) y `CLIPrompter` (production impl) — cohesión perfecta.

---

### Cambio 7 — Mover `GenerateSalt` a `crypto/` (D4)

**Problema:** `GenerateSalt` en `shared/utils.go` solo es usada por `crypto/aes_gcm.go`. Poner una función criptográfica en "shared" la desvincula de su contexto.

**Solución:** Moverla como función privada en `crypto/aes_gcm.go`:

```go
// crypto/aes_gcm.go
func generateSalt() ([]byte, error) {
    b := make([]byte, saltLen)
    if _, err := io.ReadFull(rand.Reader, b); err != nil {
        return nil, err
    }
    return b, nil
}
```

`shared/mem.go` queda con `Zero` y `Now` — dos funciones de utilería general que sí merecen estar "shared".

---

### Cambio 8 — Eliminar Archivos Zombies (E2, E3)

- Eliminar `dal/repository.go` — vacío desde 0004
- Eliminar `pkg/tokens/encryptor.go` — vacío desde 0004
- Eliminar `cmd/helpers/collectors.go` — vacío desde 0004
- Limpiar `Initialiser.appDir` en `dal/initialiser.go`: el campo se recibe pero nunca se usa. Remover del constructor y struct.

---

### Cambio 9 — OCP Registry para Codec y Collector (D5) [OPCIONAL]

**Problema:** Tanto `codec.go` como `collector.go` tienen switch statements que deben modificarse al agregar un nuevo tipo de secreto.

**Solución propuesta:** Registry centralizado en `core/secrets/registry.go`:

```go
// core/secrets/registry.go — para definir qué tipos existen
// app/codec.go — registro de decoders (unmarshal)
// cmd/ui/collector.go — registro de collectors

// En app/codec.go:
var decoders = map[string]func([]byte) (core.SecretPayload, error){
    secrets.TypeAccount: func(b []byte) (core.SecretPayload, error) {
        var p secrets.AccountSecret
        return &p, json.Unmarshal(b, &p)
    },
    secrets.TypeCredit: func(b []byte) (core.SecretPayload, error) {
        var p secrets.CreditSecret
        return &p, json.Unmarshal(b, &p)
    },
}

func unmarshalPayload(t string, b []byte) (core.SecretPayload, error) {
    dec, ok := decoders[t]
    if !ok {
        return nil, fmt.Errorf("unknown secret type: %s", t)
    }
    return dec(b)
}
```

**Trade-off:** El registry pattern reduce el boilerplate al agregar tipos pero introduce una indirección que puede confundir en lecturas de código. Para Phase 1 con solo 2 tipos, el switch es perfectamente legible. **Recomendación: implementar cuando se agregue el tercer tipo** (`NoteSecret`), no antes.

---

## Estructura Final

```
vextpss/
├── go.mod
├── changes/
├── context/
└── source/
    ├── main.go
    ├── cmd/
    │   ├── root.go
    │   ├── handlers/
    │   │   ├── guard.go
    │   │   ├── errors.go          ← NUEVO: printErr helper
    │   │   ├── add_handler.go
    │   │   ├── get_handler.go
    │   │   ├── list_handler.go
    │   │   ├── rm_handler.go
    │   │   ├── gen_handler.go
    │   │   ├── update_handler.go  ← MODIFICADO: usa MergeFrom, printErr
    │   │   ├── export_handler.go
    │   │   └── import_handler.go
    │   └── ui/                    ← RENOMBRADO desde cmd/helpers/
    │       ├── prompter.go        ← solo CLIPrompter + interfaz
    │       ├── collector.go
    │       ├── account_collector.go
    │       ├── credit_collector.go
    │       └── formatter.go
    ├── core/
    │   ├── errors.go
    │   ├── secret.go              ← MODIFICADO: MergeFrom en SecretPayload
    │   ├── repository.go
    │   ├── encryptor.go
    │   └── secrets/
    │       ├── account.go         ← RENOMBRADO + MergeFrom impl
    │       ├── credit.go          ← RENOMBRADO + MergeFrom impl
    │       └── types.go
    ├── app/                       ← RENOMBRADO desde pkg/apps/
    │   ├── codec.go
    │   ├── init_storage_uc.go
    │   ├── store_secret_uc.go
    │   ├── retrieve_secret_uc.go
    │   ├── list_secrets_uc.go
    │   ├── delete_secret_uc.go
    │   ├── update_secret_uc.go
    │   ├── export_secrets_uc.go
    │   └── import_secrets_uc.go
    ├── storage/                   ← RENOMBRADO desde dal/
    │   ├── db.go                  ← comentario de variabilidad de driver
    │   ├── record.go              ← RENOMBRADO + comentarios de tipo GORM
    │   ├── initialiser.go         ← MODIFICADO: remove campo appDir
    │   └── sqlite/                ← RENOMBRADO desde dal/repos/
    │       └── repository.go      ← RENOMBRADO + fix isUniqueConstraintError
    ├── crypto/                    ← RENOMBRADO desde pkg/tokens/
    │   └── aes_gcm.go             ← RENOMBRADO + GenerateSalt inline
    ├── config/                    ← RENOMBRADO desde pkg/configs/
    │   └── config.go              ← RENOMBRADO
    ├── shared/                    ← RENOMBRADO desde pkg/shared/ (fuera de pkg/)
    │   ├── mem.go                 ← RENOMBRADO desde utils.go (sin GenerateSalt)
    │   └── passgen.go             ← RENOMBRADO desde generator.go
    └── testutil/                  ← RENOMBRADO desde tests/mocks/
        ├── mock_repo.go
        ├── mock_encryptor.go
        ├── mock_initialiser.go
        └── mock_prompter.go       ← MOVIDO desde cmd/helpers/prompter.go
```

---

## Diagrama de Dependencias (Post-Cambios)

```
                    ┌────────────────────────────────────────┐
                    │               core/                     │
                    │  Secret, SecretPayload (+MergeFrom)     │
                    │  SecretRepository, Encryptor            │
                    │  DomainError, errores canónicos         │
                    │  secrets/: Account, Credit, types       │
                    └─────────────────┬──────────────────────┘
                                      │ importado por ▲
          ┌───────────────────────────┼──────────────────────┐
          │                           │                       │
  ┌───────┴────────┐       ┌──────────┴──────────┐           │
  │     app/       │       │    storage/sqlite/   │           │
  │  (use cases)   │       │   SQLiteRepository   │           │
  │   codec.go     │       │  implements core.*   │           │
  └───────┬────────┘       └──────────────────────┘           │
          │                                                    │
  ┌───────┴────────┐       ┌────────────────────────┐         │
  │ cmd/handlers/  │       │       crypto/           │         │
  │  guard.go      │       │  AESGCMEncryptor         │         │
  │  errors.go     │       │  implements core.*       │         │
  └───────┬────────┘       └────────────────────────┘         │
          │                                                    │
  ┌───────┴────────┐                                          │
  │    cmd/ui/     │◄─────────────────────────────────────────┘
  │  prompter.go   │
  │  collector.go  │
  │  formatter.go  │
  └────────────────┘

shared/   → usado por: app/, crypto/, cmd/ui/, cmd/handlers/
config/   → usado por: main.go únicamente
testutil/ → usado por: *_test.go únicamente
```

---

## Lo que NO Cambia

- **Arquitectura de capas** — las cuatro capas (Interface, Application, Domain, Infrastructure) se mantienen intactas.
- **Schema de base de datos** — ninguna migración requerida.
- **Formato de export/import** — los archivos `.vext` son compatibles.
- **Interfaz CLI** — comandos, flags, outputs, mensajes al usuario.
- **Modelo de seguridad** — AES-256-GCM, Argon2id, parámetros sin cambio.
- **Lógica de use cases** — solo cambia el paquete de residencia.
- **Tests existentes** — se actualizan imports pero no la lógica.

---

## Orden de Implementación

```
Fase 1 — Preparar antes del rename masivo
  1.1  Implementar MergeFrom en SecretPayload, AccountSecret, CreditSecret
  1.2  Implementar printErr en cmd/handlers/errors.go
  1.3  Crear testutil/mock_prompter.go (copiar MockPrompter)
  1.4  Verificar build: go build ./source/...

Fase 2 — Rename estructural (un solo commit)
  2.1  Crear nuevas carpetas: app/, storage/, storage/sqlite/, crypto/, config/, shared/, testutil/, cmd/ui/
  2.2  Copiar + renombrar archivos a nuevas ubicaciones con package declarations actualizadas
  2.3  Actualizar todos los imports en todos los archivos Go
  2.4  Eliminar carpetas/archivos vacíos o renombrados: pkg/, dal/, tests/, cmd/helpers/
  2.5  go build ./source/...  → debe compilar
  2.6  go test ./source/...   → todos los tests deben pasar

Fase 3 — Cambios de código
  3.1  Fix isUniqueConstraintError → errors.As con modernc.org/sqlite
  3.2  Actualizar update_handler.go → newPayload.MergeFrom(retrieveResp.Payload)
  3.3  Refactorizar handlers para usar printErr
  3.4  Mover GenerateSalt a crypto/aes_gcm.go (privada)
  3.5  Eliminar GenerateSalt de shared/mem.go
  3.6  Remover campo appDir de storage/Initialiser
  3.7  Agregar comentarios de variabilidad en storage/db.go y storage/record.go
  3.8  Eliminar MockPrompter de cmd/ui/prompter.go (ya está en testutil/)

Fase 4 — Verificación
  4.1  go build ./source/...
  4.2  go vet ./source/...
  4.3  go test ./source/... -v
  4.4  Smoke test: vext init → add → get → update → list → rm → export → import
```

---

## Testing

### Tests sin cambio de lógica

Todos los tests existentes pasan actualizando solo los imports. La lógica no cambia.

### Tests nuevos requeridos

| Test | Qué verifica |
|------|-------------|
| `core/secrets/account_test.go` | `MergeFrom` conserva username cuando el nuevo es vacío; no lo modifica cuando tiene valor |
| `core/secrets/credit_test.go` | `MergeFrom` es no-op para CreditSecret |
| `cmd/handlers/errors_test.go` | `printErr` retorna `false` con nil; `true` + mensaje correcto según tipo de error |
| `storage/sqlite/repository_test.go` | `isUniqueConstraintError` identifica correctamente un error de UNIQUE real (test de integración existente ampliado) |

### Smoke test manual

```bash
vext init
vext add github --type account
vext get github
vext update github       # username en blanco debe conservar el actual
vext add amex --type credit
vext update amex
vext list
vext export --out /tmp/backup.vext
vext rm github
vext import /tmp/backup.vext   # debe importar github, saltar amex (ya existe)
vext rm amex
vext rm github
```

---

## Estimación

| Fase | Complejidad | Estimado |
|------|-------------|----------|
| Fase 1 — Preparación | Baja | 20 min |
| Fase 2 — Rename estructural | Media (mecánico pero amplio) | 45 min |
| Fase 3 — Cambios de código | Media | 40 min |
| Fase 4 — Verificación + smoke test | Baja | 20 min |
| **Total** | | **~2 horas** |

---

## Decisiones de Diseño Documentadas

### ¿Por qué `crypto/` y no `infra/crypto/`?

Un wrapper `infra/` añadiría otro nivel de anidamiento (`infra/crypto/`, `infra/storage/`) que es más común en proyectos grandes con múltiples subsistemas de infraestructura. Para Vext, que tiene exactamente un proveedor de crypto y un proveedor de storage, la nesting adicional no aporta claridad — la agrega. Directorios planos a nivel de `source/` son suficientemente descriptivos.

### ¿Por qué `MergeFrom` en la interfaz `SecretPayload` y no en el use case?

Opciones consideradas:
1. **En el handler** (estado actual): el handler hace type assertions, importa `core/secrets` — acoplamiento de capas.
2. **En el use case**: el UC recibe current + new payload y hace el merge — lógica de negocio en el lugar correcto, pero requiere cambiar la firma del request.
3. **En `SecretPayload` como método** (propuesta): cada tipo sabe cómo mergearse consigo mismo. Es polimorfismo puro. El handler solo llama `newPayload.MergeFrom(current)` sin saber el tipo concreto.

La opción 3 es la más idiomática en Go: comportamiento polimórfico expresado como método en la interfaz. El costo (agregar un método a la interfaz) es bajo porque todos los tipos de secreto ya implementan `GetType()` y `Validate()`.

### ¿Por qué no implementar la interfaz de conexión abstracta ahora (C3)?

El valor de una interfaz de conexión (`type DBOpener interface { Open(string) (*gorm.DB, error) }`) solo se materializa cuando hay dos implementaciones. Crear la interfaz hoy sería YAGNI puro. El punto de variabilidad está documentado en el código — eso es suficiente.

### ¿Por qué el Cambio 9 (OCP registry) es opcional?

El principio de Open-Closed aplica cuando el eje de extensión es frecuente. Con 2 tipos de secreto y un tercer tipo previsto (notas), el switch es legible y el costo de modificarlo es bajo. El registry añade indirección que oscurece el flujo cuando se lee el código por primera vez. La recomendación es aplicarlo cuando se alcancen 3+ tipos — el beneficio supera el costo exactamente en ese punto.
