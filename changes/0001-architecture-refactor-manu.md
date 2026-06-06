# 0001 — Architecture Refactor [MANU]

**Fecha:** Junio 2026  
**Tipo:** Ajuste manual post-ejecución  
**Base:** `0001-architecture-refactor-xact.md`  
**Estado:** Completado — build OK, 8 suites de tests pasan

---

## Qué es un MANU

Un archivo `manu` documenta los cambios manuales realizados **después** de que la ejecución automatizada (`xact`) completó su sesión. Ocurre cuando el desarrollador itera sobre el resultado, ajustando nombres, estructuras o dependencias antes de consolidar el código. El `manu` no reemplaza al `xact` — lo complementa registrando la brecha entre el estado al terminar la sesión y el estado final commiteado.

---

## Executive Summary

Tras la sesión del `xact`, se realizó una ronda de ajustes manuales que reorganizó los nombres de paquetes, relocalizó interfaces, adoptó GORM como ORM y amplió la cobertura de tests de 3 a 8 suites. El código final compila, `go vet` pasa limpio, y todas las suites pasan.

---

## Cambios Respecto al XACT

### 1. Renaming de paquetes

La arquitectura Clean Architecture se mantuvo, pero los nombres de paquetes cambiaron para ser más idiomáticos en Go:

| Nombre en XACT | Nombre final | Observación |
|---|---|---|
| `source/pkg/domain/` | `source/core/` | Promovido a top-level; más corto e idiomático |
| `source/pkg/application/` | `source/pkg/apps/` | Abreviado |
| `source/pkg/infrastructure/storage/` | `source/dal/` | Data Access Layer, promovido a top-level |
| `source/pkg/infrastructure/crypto/` | `source/pkg/tokens/` | Renombrado para evitar conflicto con `crypto/` stdlib |
| `source/pkg/infrastructure/config/` | `source/pkg/configs/` | Renombrado a plural para consistencia |
| `source/cmd/ui/` | `source/cmd/helpers/` | Renombrado; refleja mejor que no solo hace UI |

El paquete `core` quedó particionado en:
- `source/core/` — entidad `Secret`, `SecretID`, `SecretPayload`, errores de dominio
- `source/core/secrets/` — value objects concretos (`AccountSecret`)

El DAL quedó particionado en:
- `source/dal/` — interfaces (`SecretRepository`), modelo GORM (`SecretRecord`), `db.go`, `initialiser.go`
- `source/dal/repos/` — implementación `SQLiteRepository`

---

### 2. Relocalización de interfaces

El `xact` colocó `SecretRepository` y `Encryptor` en `pkg/domain/`. En el manu estas interfaces se movieron a los paquetes que las definen e implementan:

| Interfaz | En XACT | En MANU | Razón |
|---|---|---|---|
| `SecretRepository` | `core` (domain) | `dal` | Vive con su contrato de persistencia; importada por `apps` |
| `Encryptor` | `core` (domain) | `pkg/tokens` | Vive con su implementación; importada por `apps` |

Esto modifica la dirección de dependencias: `apps` importa de `dal` y `tokens`, en lugar de que `dal` y `tokens` importen de `core`. La interfaz `StorageInitialiser` quedó correctamente en `apps` (la capa que la consume), sin cambios respecto al xact.

---

### 3. Adopción de GORM

El `xact` planeó usar `modernc.org/sqlite` con SQL raw. El manu adoptó **GORM** (`gorm.io/gorm`) con el driver `github.com/glebarez/sqlite` (Pure Go, sin CGO).

**Lo que cambió:**

| Aspecto | XACT (planeado) | MANU (real) |
|---|---|---|
| Driver | `modernc.org/sqlite` | `github.com/glebarez/sqlite` via GORM |
| Queries | SQL strings manuales | GORM DSL (`db.Create`, `db.Where(...).First`, etc.) |
| Modelo | Struct con `database/sql` scanners | `SecretRecord` con struct tags GORM |
| Migraciones | `CREATE TABLE IF NOT EXISTS` manual | `db.AutoMigrate(&SecretRecord{})` |
| Nombre columna payload | `encrypted_payload` | `encrypted` (nombre de campo GORM) |

**`SecretRecord` (en `dal/schemas.go`):**

```go
type SecretRecord struct {
    ID        int64     `gorm:"primaryKey;autoIncrement"`
    Name      string    `gorm:"uniqueIndex;not null"`
    Type      string    `gorm:"not null"`
    Salt      []byte    `gorm:"not null;type:blob"`
    Nonce     []byte    `gorm:"not null;type:blob"`
    Encrypted []byte    `gorm:"not null;type:blob"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
```

La función `TableName()` fuerza el nombre de tabla a `"secrets"` (en lugar del default `"secret_records"` de GORM).

**Motivo del cambio:** GORM reduce el boilerplate de scan/map y maneja el unique constraint error de forma más limpia. `isUniqueConstraintError` sigue siendo necesaria ya que GORM no wrappea ese error en un tipo propio.

---

### 4. Expansión de tests: de 3 a 8 suites

El `xact` entregó 3 suites. El manu agregó 5 más:

| Suite | Estado en XACT | Estado en MANU |
|---|---|---|
| `core` (errores, tipos) | ✓ | ✓ |
| `core/secrets` (AccountSecret) | — | ✓ nuevo |
| `pkg/apps` (use cases) | ✓ | ✓ |
| `pkg/tokens` (AES-GCM) | ✓ | ✓ |
| `pkg/shared` (utils) | — | ✓ nuevo |
| `cmd/handlers` (handlers con prompter mock) | — | ✓ nuevo |
| `cmd/helpers` (formatter) | — | ✓ nuevo |
| `dal/repos` (integración SQLite) | — | ✓ nuevo |

**Highlights de los tests nuevos:**

- `cmd/handlers/*_test.go` — usa `sequentialPrompter` (definido en `testhelpers_test.go`) para simular múltiples prompts en orden; testea el golden path y errores de dominio en cada handler.
- `dal/repos/sqlite_repository_test.go` — test de integración real con SQLite en memoria vía GORM; cubre Save/GetByName/ListAll/Delete y el manejo de `ErrAlreadyExists`.
- `MockPrompter` vive en `cmd/helpers/prompter.go` (paquete de producción) para estar disponible sin imports de test desde los handler tests.

---

### 5. Mocks centralizados en `source/tests/mocks/`

El `xact` mencionaba mocks inline. En el manu se creó el paquete `source/tests/mocks/` con tres mocks reutilizables:

| Mock | Interfaz que implementa |
|---|---|
| `MockRepository` | `dal.SecretRepository` |
| `MockEncryptor` | `tokens.Encryptor` |
| `MockInitialiser` | `apps.StorageInitialiser` |

Cada mock usa campos `*Fn` opcionales (e.g. `SaveFn func(...) error`) para configurar comportamiento por test sin necesidad de sub-types.

---

## Estructura Final de Paquetes

```
source/
├── main.go                         ← DI wiring completo
├── cmd/
│   ├── root.go                     ← package cmd, root Cobra command
│   ├── handlers/
│   │   ├── add_handler.go
│   │   ├── get_handler.go
│   │   ├── init_handler.go
│   │   ├── list_handler.go
│   │   ├── rm_handler.go
│   │   ├── testhelpers_test.go     ← sequentialPrompter
│   │   └── *_test.go
│   └── helpers/
│       ├── prompter.go             ← Prompter interface + CLIPrompter + MockPrompter
│       └── formatter.go
├── core/
│   ├── errors.go                   ← DomainError, ErrSecretNotFound, etc.
│   ├── secret.go                   ← Secret entity, SecretPayload interface
│   └── secrets/
│       └── account_secret.go       ← AccountSecret value object
├── dal/
│   ├── db.go                       ← Open/Close/Migrate via GORM
│   ├── initialiser.go              ← Initialiser struct
│   ├── repository.go               ← SecretRepository interface
│   ├── schemas.go                  ← SecretRecord GORM model
│   └── repos/
│       └── sqlite_repository.go    ← SQLiteRepository implementation
├── pkg/
│   ├── apps/
│   │   ├── init_storage_uc.go      ← StorageInitialiser interface + InitStorageUC
│   │   ├── store_secret_uc.go
│   │   ├── retrieve_secret_uc.go
│   │   ├── list_secrets_uc.go
│   │   ├── delete_secret_uc.go
│   │   └── *_test.go
│   ├── configs/
│   │   └── app_config.go
│   ├── shared/
│   │   └── utils.go                ← Zero, GenerateSalt, Now
│   └── tokens/
│       ├── encryptor.go            ← Encryptor interface
│       └── aes_gcm_encryptor.go
└── tests/
    └── mocks/
        ├── mock_repo.go
        ├── mock_encryptor.go
        └── mock_initialiser.go
```

---

## Diagrama de Dependencias

```
cmd/handlers ──→ pkg/apps ──→ core
     │               │──────→ dal (SecretRepository interface)
     │               └──────→ pkg/tokens (Encryptor interface)
     │
cmd/helpers (Prompter)

dal/repos ────→ dal (SecretRecord, SecretRepository)
          ────→ core (Secret, ErrSecretNotFound, etc.)
          ────→ gorm.io/gorm

pkg/tokens ───→ core (ErrDecryptionFailed)
           ───→ pkg/shared

main.go ──────→ todos los paquetes anteriores (único punto de wiring)
```

**Regla de dependencias:** `core` no importa nada del proyecto. `apps` no importa de `cmd`. `dal/repos` no importa de `apps`. El flujo es unidireccional.

---

## Dependencias Finales

| Paquete | Propósito | Cambio vs XACT |
|---|---|---|
| `github.com/spf13/cobra` | CLI framework | Sin cambio |
| `golang.org/x/term` | Input seguro (prompts) | Sin cambio |
| `golang.org/x/crypto/argon2` | Argon2id KDF | Sin cambio |
| `gorm.io/gorm` | ORM | **Nuevo** (reemplaza modernc raw) |
| `github.com/glebarez/sqlite` | Driver SQLite Pure Go para GORM | **Nuevo** (reemplaza modernc.org/sqlite) |

---

## Resultados de Verificación

```
$ go build ./source/...   → OK
$ go vet ./source/...     → OK
$ go test ./source/...
ok  vextpss/source/cmd/handlers        3.659s
ok  vextpss/source/cmd/helpers         2.638s
ok  vextpss/source/core                2.435s
ok  vextpss/source/core/secrets        2.427s
ok  vextpss/source/dal/repos           0.403s
ok  vextpss/source/pkg/apps            3.618s
ok  vextpss/source/pkg/shared          2.393s
ok  vextpss/source/pkg/tokens          3.061s
```

---

## Lessons Learned

1. **Los nombres de paquetes importan más de lo que parece.** `infrastructure/crypto` colisiona visualmente con la stdlib `crypto/`. Usar nombres de negocio (`tokens`) evita esa confusión.

2. **Las interfaces viven mejor en la capa que las define como contrato**, no necesariamente en la que representa el dominio puro. `SecretRepository` en `dal` es más legible que en `core`, porque `core` no sabe de persistencia.

3. **GORM reduce el boilerplate de integración significativamente** — el repositorio pasó de necesitar scan manual a ser lectura casi directa. El costo es la dependencia, justificada aquí porque el driver es Pure Go y no rompe portabilidad.

4. **Los handler tests con `sequentialPrompter` son el test más valioso del sistema**: testean el path completo desde la entrada del usuario hasta el resultado visible, sin tocar I/O ni base de datos real.

---

## Next Steps

- [ ] CI pipeline con GitHub Actions (`go test ./source/...` + `go vet`).
- [ ] Phase 2: `CardSecret` y `NoteSecret` — el patrón está establecido en `core/secrets/`.
- [ ] `vext update` command (F-09): update de campo sin rm + add.
- [ ] Advertencia explícita cuando se corre `vext add/get/list/rm` sin haber inicializado (`db == nil`).
- [ ] Logging estructurado (`slog`) reemplazando `fmt.Printf` en handlers.
