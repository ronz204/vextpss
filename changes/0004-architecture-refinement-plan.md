# 0004 — Refinamiento de Arquitectura: Desacoplamiento, Limpieza y Robustez

## Objetivo

Realizar una segunda pasada arquitectónica sobre el código existente para eliminar acoplamientos implícitos entre capas, centralizar responsabilidades dispersas, corregir un bug funcional en `UpdateSecretUC`, y fortalecer la robustez ante errores. El resultado debe ser un codebase más predecible, más fácil de extender y más resistente a regresiones.

---

## Rationale

El análisis crítico del estado actual revela siete categorías de mejora. Ninguna requiere cambiar la distribución de carpetas (`/cmd`, `/pkg`, `/core`, `/dal`, `/tests`) — la distribución es correcta. Los problemas son más sutiles: violaciones de dirección de dependencia, duplicación de lógica de despacho por tipo, strings mágicos dispersos, y un bug que impide actualizar secretos de tipo `credit`.

### Problemas identificados

| # | Categoría | Problema | Archivo(s) | Severidad |
|---|-----------|---------|------------|-----------|
| P1 | Acoplamiento de capas | `pkg/apps/` importa `dal` y `pkg/tokens` solo para usar sus **interfaces**. La capa de aplicación no debería conocer los nombres de los paquetes de infraestructura. | `retrieve_secret_uc.go`, `store_secret_uc.go`, `update_secret_uc.go`, `export_secrets_uc.go`, `import_secrets_uc.go` | Alta |
| P2 | Bug funcional | `UpdateSecretUC` hardcodea `*secrets.AccountSecret`. Llamar `vext update` sobre un secreto de tipo `credit` hace casting incorrecto y produce comportamiento undefined. | `update_secret_uc.go:72-74` | Alta |
| P3 | Lógica duplicada | `deserialisePayload` existe solo en `retrieve_secret_uc.go`, pero `update_secret_uc.go` lo llama desde ese mismo archivo a través del mismo paquete. Si se agregan nuevos tipos, hay que recordar dos lugares (`retrieve`, el dispatcher en `export`/`import`). | `retrieve_secret_uc.go:63`, `export_secrets_uc.go`, `import_secrets_uc.go` | Media |
| P4 | Strings mágicos | `"account"` y `"credit"` aparecen literales en 6 archivos distintos: collectors, retrieve UC, formatter, export/import UCs. Un typo en cualquiera silencia el bug. | 6 archivos | Media |
| P5 | Null pointer panic | Si el usuario corre `vext add` (o cualquier comando de datos) antes de `vext init`, `storeUC` es `nil`. El handler llama `uc.Execute()` sin nil-check y el proceso explota con panic no controlado en lugar de dar un mensaje útil. | `main.go:47-65`, todos los handlers | Media |
| P6 | Detección de error SQLite frágil | La función `isUniqueConstraintError` usa `strings.Contains(err.Error(), "UNIQUE constraint failed")`. Esto es frágil: un cambio en el mensaje de error del driver SQLite o una actualización de versión puede romper silenciosamente el flujo. | `dal/repos/sqlite_repository.go:139` | Media |
| P7 | Archivo demasiado grande | `cmd/helpers/collectors.go` tiene 152 líneas que mezclan la interfaz `SecretCollector`, la factory, `AccountCollector` y `CreditCollector`. Agregar un tercer tipo volverá esto inmanejable. | `cmd/helpers/collectors.go` | Baja |

---

## Approach

Se mantiene la distribución de carpetas actual. Los cambios refinen **qué vive dentro de cada capa**, no dónde están las carpetas.

### Principio guía: las interfaces pertenecen al consumidor, no al proveedor

En Clean Architecture, las interfaces que define el dominio o la aplicación para sus dependencias externas (repositorio, encryptor) deben vivir en la capa que las **consume**, no en la que las **implementa**. Actualmente `SecretRepository` vive en `dal/` (capa de infraestructura) y `Encryptor` vive en `pkg/tokens/` (infraestructura). Los use cases en `pkg/apps/` importan `dal` y `pkg/tokens` solo para acceder a esas interfaces — un acoplamiento innecesario.

La corrección: mover las interfaces a `core/`, que es la capa de dominio/contrato. La infraestructura (`dal/repos/`, `pkg/tokens/`) implementa contratos del dominio. Los use cases solo importan `core/`.

```
ANTES:
  pkg/apps/ → imports dal         (para SecretRepository)
  pkg/apps/ → imports pkg/tokens  (para Encryptor)
  pkg/apps/ → imports core

DESPUÉS:
  pkg/apps/ → imports core        (SecretRepository, Encryptor, entidades, errores)
  pkg/apps/ → imports pkg/shared  (utilidades)

  dal/repos/ → imports core       (implementa core.SecretRepository)
  pkg/tokens/ → imports core      (implementa core.Encryptor)
```

---

## Cambios Específicos

### Cambio 1 — Mover `SecretRepository` a `core/`

**Acción:** Crear `core/repository.go`. Mover la interfaz `SecretRepository` y el tipo `FullRecord` desde `dal/repository.go` a `core/repository.go`. Vaciar `dal/repository.go` (o eliminarlo si `dal/` no necesita exportar más contratos).

**`core/repository.go` (nuevo):**
```go
package core

import "context"

// FullRecord pairs a Secret with its encrypted payload for bulk operations.
type FullRecord struct {
    Secret    Secret
    Encrypted []byte
}

// SecretRepository is the persistence contract consumed by use cases.
type SecretRepository interface {
    Save(ctx context.Context, secret *Secret, encrypted []byte) error
    GetByName(ctx context.Context, name string) (*Secret, []byte, error)
    ListAll(ctx context.Context) ([]Secret, error)
    GetAll(ctx context.Context) ([]FullRecord, error)
    Delete(ctx context.Context, name string) error
    Update(ctx context.Context, secret *Secret, encrypted []byte) error
}
```

**Impacto en imports:**
- `dal/repos/sqlite_repository.go`: cambia `dal.SecretRepository` → `core.SecretRepository`, `dal.FullRecord` → `core.FullRecord`
- Todos los use cases en `pkg/apps/`: eliminar `import "vextpss/source/dal"`
- `dal/repository.go`: puede quedar vacío o eliminarse

---

### Cambio 2 — Mover `Encryptor` a `core/`

**Acción:** Crear `core/encryptor.go`. Mover la interfaz `Encryptor` desde `pkg/tokens/encryptor.go` a `core/encryptor.go`. El archivo `pkg/tokens/encryptor.go` puede eliminarse o quedar vacío.

**`core/encryptor.go` (nuevo):**
```go
package core

import "context"

// Encryptor is the cryptographic contract consumed by use cases.
type Encryptor interface {
    Encrypt(ctx context.Context, plaintext, masterPassword []byte) (salt, nonce, ciphertext []byte, err error)
    Decrypt(ctx context.Context, masterPassword, salt, nonce, ciphertext []byte) ([]byte, error)
}
```

**Impacto en imports:**
- `pkg/tokens/aes_gcm_encryptor.go`: cambia `tokens.Encryptor` → `core.Encryptor`
- Todos los use cases en `pkg/apps/`: eliminar `import "vextpss/source/pkg/tokens"`
- `tests/mocks/mock_encryptor.go`: actualizar tipo a `core.Encryptor`

---

### Cambio 3 — Crear constantes de tipo en `core/secrets/types.go`

**Acción:** Crear `core/secrets/types.go` con las constantes de tipo string. Reemplazar todos los literales `"account"` y `"credit"` en el código por estas constantes.

**`core/secrets/types.go` (nuevo):**
```go
package secrets

const (
    TypeAccount = "account"
    TypeCredit  = "credit"
)
```

**Archivos a actualizar:**
- `cmd/helpers/collectors.go` — factory `NewSecretCollector`
- `cmd/helpers/formatter.go` — `PrintSecret` type switch
- `pkg/apps/retrieve_secret_uc.go` — `deserialisePayload`
- `pkg/apps/update_secret_uc.go` — casting hardcodeado
- `pkg/apps/export_secrets_uc.go` — si usa literales
- `pkg/apps/import_secrets_uc.go` — si usa literales
- `cmd/handlers/add_handler.go` — flag default `--type`

---

### Cambio 4 — Centralizar el codec de payload en `pkg/apps/codec.go`

**Acción:** Crear `pkg/apps/codec.go` con dos funciones internas al paquete `apps`: `marshalPayload` y `unmarshalPayload`. Eliminar la función `deserialisePayload` de `retrieve_secret_uc.go` y reemplazar todos sus usos con `unmarshalPayload`.

**`pkg/apps/codec.go` (nuevo):**
```go
package apps

import (
    "encoding/json"
    "fmt"

    "vextpss/source/core"
    "vextpss/source/core/secrets"
)

func marshalPayload(payload core.SecretPayload) ([]byte, error) {
    return json.Marshal(payload)
}

func unmarshalPayload(secretType string, data []byte) (core.SecretPayload, error) {
    switch secretType {
    case secrets.TypeAccount:
        var p secrets.AccountSecret
        if err := json.Unmarshal(data, &p); err != nil {
            return nil, fmt.Errorf("failed to parse account secret: %w", err)
        }
        return &p, nil
    case secrets.TypeCredit:
        var p secrets.CreditSecret
        if err := json.Unmarshal(data, &p); err != nil {
            return nil, fmt.Errorf("failed to parse credit secret: %w", err)
        }
        return &p, nil
    default:
        return nil, fmt.Errorf("unknown secret type: %s", secretType)
    }
}
```

**Archivos a actualizar:**
- Eliminar `deserialisePayload` de `retrieve_secret_uc.go`; llamar `unmarshalPayload` en su lugar
- `export_secrets_uc.go` y `import_secrets_uc.go`: reemplazar `json.Unmarshal` directo por `unmarshalPayload`
- `store_secret_uc.go`: usar `marshalPayload` en lugar de `json.Marshal` directo
- `update_secret_uc.go`: usar ambas (ver Cambio 5)

**Nota:** `pkg/apps/codec.go` puede importar `core/secrets/` sin ciclos porque `core/secrets/` importa `core`, y `pkg/apps/` importa `core/secrets/`. No hay ciclo.

---

### Cambio 5 — Corregir `UpdateSecretUC` para soporte polimórfico

**Problema:** `UpdateSecretUC` hardcodea `*secrets.AccountSecret` en la línea 72-81. No puede actualizar secretos de tipo `credit`.

**Solución:** Cambiar el contrato del UC para aceptar un `NewPayload core.SecretPayload` ya construido por el llamador. El UC solo verifica la contraseña mediante decriptado, valida el tipo coincidente y re-encripta.

**`UpdateSecretRequest` rediseñado:**
```go
type UpdateSecretRequest struct {
    Name           string
    MasterPassword []byte
    NewPayload     core.SecretPayload  // Pre-built by handler; must match current type
}
```

**Lógica del UC actualizada:**
```
1. Validar Name, MasterPassword, NewPayload no nulos
2. Fetch record via repo.GetByName()
3. Decrypt con master password (verifica que la contraseña es correcta)
4. defer shared.Zero(plaintext)
5. Verificar que current.Type == req.NewPayload.GetType() → error si mismatch
6. Llamar req.NewPayload.Validate()
7. marshalPayload(req.NewPayload)
8. Encrypt con fresh salt/nonce
9. repo.Update()
```

**`UpdateHandler` actualizado:**
El handler necesita conocer el tipo actual antes de colectar el nuevo payload. Opciones:
- Opción A: El handler usa `RetrieveSecretUC` para obtener el tipo actual (y de paso verifica la contraseña antes de continuar)
- Opción B: Agregar flag `--type` al comando `vext update` (más simple, pero peor UX)

**Recomendación: Opción A.** El flujo queda:
1. Prompt: master password
2. Llamar `RetrieveSecretUC` (verifica password y obtiene tipo actual)
3. Usar `NewSecretCollector(currentType, opts)` en modo update (pre-filled con valores actuales donde aplique)
4. Llamar `UpdateSecretUC` con el nuevo payload

Para esto, `UpdateHandler` recibirá tanto `retrieveUC` como `updateUC`. Esto es aceptable — el handler orquesta la interacción completa.

**Nota de UX:** El collector en modo update para `AccountSecret` pedirá nuevo username (vacío = mantener actual) y nueva contraseña. Para `CreditSecret`, pedirá todos los campos de nuevo (o solo los cambiables). Definir en detalle durante la implementación.

---

### Cambio 6 — Robustecer detección de error SQLite

**Problema:** `isUniqueConstraintError` usa `strings.Contains(err.Error(), "UNIQUE constraint failed")`. Frágil.

**Solución:** Usar `errors.As` para extraer el tipo de error nativo del driver SQLite (`modernc.org/sqlite`).

El driver `glebarez/sqlite` usa `modernc.org/sqlite` internamente. Sus errores satisfacen la interfaz `interface { Code() sqlite3.ErrorCode }`. El código de error para constraint unique es `sqlite3.SQLITE_CONSTRAINT_UNIQUE` (valor 2067).

**Implementación:**
```go
import "modernc.org/sqlite"

func isUniqueConstraintError(err error) bool {
    var sqliteErr *sqlite.Error
    if errors.As(err, &sqliteErr) {
        return sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE
    }
    // Fallback conservador para no romper si el tipo cambia.
    return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
```

**Nota:** Verificar en `go.mod` si `modernc.org/sqlite` ya está como dependencia transitiva (es el backend de `glebarez/sqlite`). Si no está listado directamente, agregar con `go get`.

---

### Cambio 7 — Guardia nil en handlers para DB no inicializada

**Problema:** Si el usuario corre `vext add` (o cualquier comando de datos) antes de `vext init`, el use case correspondiente es `nil` y el proceso entra en panic.

**Solución A (recomendada): Guard en cada handler que requiere DB.**

Cada handler que depende de un UC de datos verifica si el UC es nil antes de ejecutar:

```go
func (h *AddHandler) Handle(ctx context.Context, cmd *cobra.Command, args []string) {
    if h.uc == nil {
        fmt.Fprintln(os.Stderr, "Error: vault not initialised. Run 'vext init' first.")
        return
    }
    // ... resto del flujo
}
```

Esto aplica a: `AddHandler`, `GetHandler`, `ListHandler`, `RmHandler`, `UpdateHandler`, `ExportHandler`, `ImportHandler`.

**Solución B (alternativa): Wrapper NilUC que implementa la interfaz y devuelve el error.**

Crear un use case vacío que responde con el error en cada operación. Más elegante pero agrega indirection. No recomendado para este caso.

**Implementación:** Opción A. Un helper compartido en `cmd/handlers/` puede centralizar el check:

```go
// cmd/handlers/guard.go
func requiresInit(name string, uc any) bool {
    if uc == nil {
        fmt.Fprintf(os.Stderr, "Error: vault not initialised. Run 'vext init' first.\n")
        return false
    }
    return true
}
```

---

### Cambio 8 — Dividir `collectors.go` en tres archivos

**Problema:** 152 líneas mezclando interfaz, factory y dos implementaciones. Agregar `NoteCollector` llevaría el archivo a ~230 líneas.

**Acción:** Dividir en:

```
cmd/helpers/
├── collector.go          ← Interfaz SecretCollector + CollectorOptions + factory NewSecretCollector
├── account_collector.go  ← AccountCollector (toda su lógica)
└── credit_collector.go   ← CreditCollector (toda su lógica)
```

El archivo original `collectors.go` desaparece. Los tests en `collectors_test.go` mantienen el mismo nombre o se dividen si se prefiere.

---

## Estructura Final Post-Cambios

```
source/
├── main.go
├── cmd/
│   ├── root.go
│   ├── handlers/
│   │   ├── guard.go              ← NUEVO: helper nil-check
│   │   ├── add_handler.go
│   │   ├── get_handler.go
│   │   ├── list_handler.go
│   │   ├── rm_handler.go
│   │   ├── gen_handler.go
│   │   ├── update_handler.go     ← MODIFICADO: recibe retrieveUC + updateUC
│   │   ├── export_handler.go
│   │   ├── import_handler.go
│   │   └── [test files]
│   └── helpers/
│       ├── prompter.go
│       ├── formatter.go          ← MODIFICADO: usa constantes de tipo
│       ├── collector.go          ← NUEVO (extraído de collectors.go)
│       ├── account_collector.go  ← NUEVO (extraído de collectors.go)
│       ├── credit_collector.go   ← NUEVO (extraído de collectors.go)
│       └── [test files]
├── core/
│   ├── errors.go
│   ├── secret.go
│   ├── repository.go             ← NUEVO: SecretRepository + FullRecord
│   ├── encryptor.go              ← NUEVO: Encryptor interface
│   └── secrets/
│       ├── types.go              ← NUEVO: TypeAccount, TypeCredit
│       ├── account_secret.go
│       ├── credit_secret.go
│       └── [test files]
├── dal/
│   ├── db.go
│   ├── schemas.go
│   ├── initialiser.go
│   ├── repository.go             ← VACIADO: interfaz movida a core/
│   └── repos/
│       ├── sqlite_repository.go  ← MODIFICADO: implementa core.SecretRepository
│       └── [test files]
├── pkg/
│   ├── apps/
│   │   ├── codec.go              ← NUEVO: marshalPayload + unmarshalPayload
│   │   ├── store_secret_uc.go    ← MODIFICADO: usa codec, core interfaces
│   │   ├── retrieve_secret_uc.go ← MODIFICADO: usa codec, elimina deserialisePayload
│   │   ├── list_secrets_uc.go
│   │   ├── delete_secret_uc.go
│   │   ├── update_secret_uc.go   ← MODIFICADO: acepta NewPayload, tipo-agnóstico
│   │   ├── export_secrets_uc.go  ← MODIFICADO: usa codec, core interfaces
│   │   ├── import_secrets_uc.go  ← MODIFICADO: usa codec, core interfaces
│   │   ├── init_storage_uc.go
│   │   └── [test files]
│   ├── configs/
│   │   └── app_config.go
│   ├── shared/
│   │   ├── utils.go
│   │   └── generator.go
│   └── tokens/
│       ├── encryptor.go          ← VACIADO: interfaz movida a core/
│       └── aes_gcm_encryptor.go  ← MODIFICADO: implementa core.Encryptor
├── tests/
│   └── mocks/
│       ├── mock_repo.go          ← MODIFICADO: tipo core.SecretRepository
│       ├── mock_encryptor.go     ← MODIFICADO: tipo core.Encryptor
│       └── mock_initialiser.go
└── README.md
```

---

## Diagrama de Dependencias (Post-Cambios)

```
                        ┌─────────────────────────────────┐
                        │              core/               │
                        │  Secret, SecretPayload           │
                        │  SecretRepository (interface)    │
                        │  Encryptor (interface)           │
                        │  DomainError, errores canónicos  │
                        │  secrets/TypeAccount, TypeCredit │
                        └──────────────┬──────────────────┘
                                       │ ▲ importado por
               ┌───────────────────────┼────────────────┐
               │                       │                │
       ┌───────┴────────┐    ┌──────────┴──────────┐   │
       │   pkg/apps/    │    │       dal/repos/     │   │
       │  (use cases)   │    │  SQLiteRepository    │   │
       │  codec.go      │    │  implements          │   │
       └───────┬────────┘    │  core.SecretRepository│  │
               │             └─────────────────────┘  │
               │                                        │
       ┌───────┴────────┐    ┌────────────────────────┐ │
       │  cmd/handlers/ │    │      pkg/tokens/        │ │
       │   +guard.go    │    │  AESGCMEncryptor         │ │
       └───────┬────────┘    │  implements             │ │
               │             │  core.Encryptor          │ │
       ┌───────┴────────┐    └────────────────────────┘ │
       │  cmd/helpers/  │                                │
       │  collector.go  │                                │
       │  account_coll. │                                │
       │  credit_coll.  │                                │
       │  formatter.go  │◄───────────────────────────────┘
       └────────────────┘
```

---

## Lo que NO cambia

- Distribución de carpetas `/cmd`, `/pkg`, `/core`, `/dal`, `/tests`
- Modelo de seguridad (AES-256-GCM + Argon2id, parámetros hardcodeados)
- Schema de base de datos (ninguna migración requerida)
- Formato de export/import
- Interfaz CLI (comandos, flags, outputs)
- Todos los tests existentes (se actualizan imports, no lógica)

---

## Orden de Implementación

El orden importa para no romper el build en pasos intermedios:

```
Fase 1 — Sin romper nada (solo agregar)
  1.1  Crear core/secrets/types.go con constantes
  1.2  Crear core/repository.go (copia de dal/repository.go)
  1.3  Crear core/encryptor.go  (copia de pkg/tokens/encryptor.go)

Fase 2 — Actualizar implementaciones
  2.1  dal/repos/sqlite_repository.go → implementa core.SecretRepository
  2.2  pkg/tokens/aes_gcm_encryptor.go → implementa core.Encryptor
  2.3  tests/mocks/ → actualizar tipos a core.*

Fase 3 — Actualizar use cases
  3.1  Crear pkg/apps/codec.go
  3.2  Actualizar todos los UCs: eliminar imports de dal y pkg/tokens,
       usar core.*, usar codec

Fase 4 — Corregir UpdateSecretUC (cambio de contrato)
  4.1  Rediseñar UpdateSecretRequest y lógica del UC
  4.2  Actualizar UpdateHandler para recibir retrieveUC + updateUC
  4.3  Actualizar main.go: pasar retrieveUC a NewUpdateHandler
  4.4  Actualizar tests de UpdateSecretUC y UpdateHandler

Fase 5 — Robustez y limpieza
  5.1  Agregar cmd/handlers/guard.go y nil-checks en todos los handlers
  5.2  Mejorar isUniqueConstraintError en sqlite_repository.go
  5.3  Dividir collectors.go → collector.go, account_collector.go, credit_collector.go
  5.4  Reemplazar todos los string literals de tipo por constantes

Fase 6 — Verificación
  6.1  go build ./...
  6.2  go vet ./...
  6.3  go test ./...
  6.4  Smoke test manual: vext init → add → get → update → list → rm
```

---

## Testing

### Tests unitarios (sin cambios de lógica)

Todos los tests existentes deben pasar sin modificar su lógica. Solo se actualizan imports.

Verificaciones adicionales:
- `UpdateSecretUC` test: verificar que un `CreditSecret` se puede actualizar correctamente
- `UpdateHandler` test: verificar flujo con `retrieveUC` mock + `updateUC` mock
- `guard.go` test: verificar que el nil-check imprime el mensaje correcto
- `isUniqueConstraintError` test: verificar con un error de tipo `*sqlite.Error`

### Smoke test manual

```bash
vext init
vext add github --type account
vext get github
vext update github
vext add amex --type credit
vext update amex         # NUEVO: debe funcionar post-cambio 5
vext list
vext export --out /tmp/backup.vext
vext import /tmp/backup.vext
vext rm amex
```

### Verificación de build limpio

```bash
go build ./...
go vet ./...
go test ./... -v
```

---

## Estimación

| Fase | Complejidad | Estimado |
|------|-------------|----------|
| Fase 1 — Agregar interfaces a core/ | Baja | 15 min |
| Fase 2 — Actualizar implementaciones | Baja | 20 min |
| Fase 3 — Actualizar use cases | Media | 30 min |
| Fase 4 — Fix UpdateSecretUC | Alta | 45 min |
| Fase 5 — Robustez y limpieza | Media | 30 min |
| Fase 6 — Verificación | Baja | 15 min |
| **Total** | | **~2.5 horas** |

---

## Decisiones de Diseño Documentadas

### ¿Por qué mover interfaces a `core/` y no a un paquete `ports/`?

Un paquete `ports/` separado introduce otra capa de indirección sin beneficio tangible en este tamaño de proyecto. Go favorece paquetes pequeños con cohesión alta. `core/` ya es el contrato del dominio — tener `SecretRepository` y `Encryptor` ahí es idiomático.

### ¿Por qué el codec en `pkg/apps/` y no en `core/`?

El codec necesita importar `core/secrets/` (para hacer `json.Unmarshal` en los tipos concretos). `core/secrets/` importa `core` (para la interfaz `SecretPayload`). Si el codec viviera en `core/`, el ciclo sería: `core/ → core/secrets/ → core/`. Go no permite ciclos de importación aunque sean entre paquetes del mismo módulo.

### ¿Por qué `UpdateHandler` recibe dos UCs en lugar de uno?

La acción de actualizar tiene dos responsabilidades distintas: leer el estado actual (para conocer el tipo y pre-rellenar) y escribir el nuevo estado. Mezclarlas en un único UC violaría el principio de responsabilidad única. El handler es el lugar correcto para orquestar dos UCs cuando la interacción lo requiere.

### ¿Por qué no eliminar `dal/repository.go` y `pkg/tokens/encryptor.go` completamente?

Para no romper imports en código que el compilador no haya procesado aún durante la transición. Una vez migrado todo, se puede eliminar o dejar vacío con un comentario que explique la migración.
