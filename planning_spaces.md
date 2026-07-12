# Spaces (namespaces) — diseño final

## La idea

Espacios lógicos para agrupar secrets. El usuario siempre está parado en un
space activo — como una branch en git o un context en kubectl. No hay flags
en cada comando: cambiás de space con `vext use` y todos los comandos operan
ahí automáticamente.

```
vext spaces add universidad
vext use universidad
vext add gmail          # va a "universidad"
vext list               # lista solo "universidad"
vext use trabajo
vext add slack          # va a "trabajo"
```

---

## Modelo de datos

Dos entidades nuevas en la DB: `spaces` y `meta`.

### Tabla `spaces`

```go
type SpaceRecord struct {
    ID        int64     `gorm:"primaryKey;autoIncrement"`
    Name      string    `gorm:"uniqueIndex;not null"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (SpaceRecord) TableName() string { return "spaces" }
```

Los spaces son entidades de primera clase. Renombrar un space es un `UPDATE`
de una sola fila — no toca ningún secret.

### Tabla `meta`

```go
type MetaRecord struct {
    Key   string `gorm:"primaryKey"`
    Value string `gorm:"not null"`
}

func (MetaRecord) TableName() string { return "meta" }
```

Tabla genérica de estado. Por ahora guarda una sola clave: `active_space`.
`vext use universidad` → `UPDATE meta SET value = 'universidad' WHERE key = 'active_space'`.

El estado viaja con la DB: si movés el archivo, el space activo se mueve con él.

### Tabla `secrets` — cambio del índice único

`secrets.name` deja de ser único globalmente. La unicidad pasa a `(space_id, name)`.
Podés tener `gmail` en `universidad` y `gmail` en `trabajo` como secrets distintos.

```go
type SecretRecord struct {
    ID        int64     `gorm:"primaryKey;autoIncrement"`
    SpaceID   int64     `gorm:"not null;uniqueIndex:idx_space_name"`
    Name      string    `gorm:"not null;uniqueIndex:idx_space_name"`
    Type      string    `gorm:"not null"`
    Algorithm string    `gorm:"not null"`
    Ciphertext []byte   `gorm:"not null;type:blob"`
    Metadata   []byte   `gorm:"not null;type:blob"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
```

El join `space_id → spaces.name` vive dentro del `GORMRepository`. El dominio
nunca ve un entero — sigue viendo `Space string` en `core.Secret`.

---

## Capa de dominio — `source/secrets/`

### Por qué Space no necesita moon

`moon/` existe por una razón específica: los payloads de secrets son polimórficos
y cifrados. `Account` tiene `{Username, Password}`, `Finance` tiene `{Card, Mobile}`.
Son estructuras distintas que se serializan, deserializan, validan y cuyos campos
sensibles se limpian con `memory.Cleaner`. El factory, los aggregates y composites
existen para manejar esa complejidad.

`Space` no tiene nada de eso:
- No se cifra
- No tiene variantes de tipo
- No tiene campos sensibles
- Su estructura es `{ID, Name, CreatedAt}` — punto

`Meta` tampoco es un concepto de dominio — es estado de aplicación. No pertenece
al dominio de negocio, solo necesita una interfaz que el storage pueda implementar.

**Regla:** `moon/` solo existe para la variedad de payloads cifrados de secrets.
Cualquier entidad sin payloads polimórficos va directo a `core/`.

### Estructura final de `source/secrets/`

```
source/secrets/
  core/
    domain.go       ← sin cambios (Secret, Payload)
    repository.go   ← actualizar firmas (space en GetByName, Rename, Delete, List)
    encryptor.go    ← sin cambios
    shapes.go       ← sin cambios
    space.go        ← Space struct + SpaceRepository + errores de Spaces
    state.go        ← StateRepository (interfaz para active_space)
  moon/             ← sin cambios, sigue siendo solo para payloads de secrets
    accounts/
    finances/
    factory.go
```

### `secrets/core/space.go`

```go
package core

import (
    "context"
    "errors"
    "time"
)

var (
    ErrSpaceNotFound      = errors.New("space not found")
    ErrSpaceAlreadyExists = errors.New("space already exists")
)

type Space struct {
    ID        int64
    Name      string
    CreatedAt time.Time
}

type SpaceRepository interface {
    Create(ctx context.Context, name string) error
    GetByName(ctx context.Context, name string) (Space, error)
    Rename(ctx context.Context, oldName, newName string) error
    Delete(ctx context.Context, name string) error
    List(ctx context.Context) ([]Space, error)
}
```

### `secrets/core/state.go`

```go
package core

import "context"

type StateRepository interface {
    GetActiveSpace(ctx context.Context) (string, error)
    SetActiveSpace(ctx context.Context, name string) error
}
```

---

## Dominio

### `core.Secret`

```go
type Secret struct {
    ID        int64
    Space     string    // nombre del space, resuelto por el repo
    Name      string
    Type      string
    Encrypted Encrypted
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### `core.Space`

```go
type Space struct {
    ID        int64
    Name      string
    CreatedAt time.Time
}
```

### Interfaces del repositorio

```go
type Repository interface {
    Create(ctx context.Context, secret Secret) error
    GetByName(ctx context.Context, space, name string) (Secret, error)
    Update(ctx context.Context, secret Secret) error
    Rename(ctx context.Context, space, oldName, newName string) error
    Delete(ctx context.Context, space, name string) error
    List(ctx context.Context, space string) ([]Secret, error) // "" = todos
}

type SpaceRepository interface {
    Create(ctx context.Context, name string) error
    GetByName(ctx context.Context, name string) (Space, error)
    Rename(ctx context.Context, oldName, newName string) error
    Delete(ctx context.Context, name string) error
    List(ctx context.Context) ([]Space, error)
}

type StateRepository interface {
    GetActiveSpace(ctx context.Context) (string, error) // "" si nunca se usó vext use
    SetActiveSpace(ctx context.Context, name string) error
}
```

Las tres interfaces viven en `secrets/core/`. El `GORMRepository` implementa
las tres (o se separan en structs distintos que comparten el `*gorm.DB`).

---

## Wiring en `main.go`

`loadDeps()` lee el space activo al iniciar y lo pone en `Deps`.
Todos los use cases consumen `deps.ActiveSpace` sin saber nada de estado.

```go
type Deps struct {
    Repo        core.Repository
    SpaceRepo   core.SpaceRepository
    StateRepo   core.StateRepository
    Cryp        core.Encryptor
    Prompter    terminal.Prompter
    ActiveSpace string
}
```

```go
func loadDeps() (funcs.Deps, error) {
    // ...init db...
    stateRepo := storages.NewState(db)
    active, err := stateRepo.GetActiveSpace(context.Background())
    if err != nil {
        return funcs.Deps{}, err
    }
    return funcs.Deps{
        Repo:        storages.New(db),
        SpaceRepo:   storages.NewSpaces(db),
        StateRepo:   stateRepo,
        Cryp:        aesgcm.New(aesgcm.DefaultConfig()),
        Prompter:    terminal.NewPrompter(),
        ActiveSpace: active,
    }, nil
}
```

---

## Comandos

### Comandos existentes — cero flags nuevos

Todos consumen `deps.ActiveSpace` directamente. No cambia la interfaz de ningún
comando desde la perspectiva del usuario.

```
vext add gmail        # agrega en el space activo
vext get gmail        # busca en el space activo
vext list             # lista el space activo
vext drop gmail       # borra en el space activo
vext ren gmail gmail  # renombra en el space activo
vext upd gmail        # actualiza en el space activo
vext rota gmail       # rota en el space activo
```

Escape hatch para ver todo:

```
vext list --all       # ignora el space activo, lista todo
```

### `vext use <space>`

Cambia el space activo. Persiste en la tabla `meta`.

```
vext use universidad    # activa "universidad"
vext use                # muestra el space activo actual
```

### `vext spaces` — subcomandos

```
vext spaces             # lista todos los spaces con conteo de secrets
vext spaces add <name>  # crea un space
vext spaces drop <name> # elimina un space (requiere que esté vacío, o --force)
vext spaces ren <old> <new>  # renombra — un solo UPDATE en spaces, cero en secrets
```

Output de `vext spaces`:

```
  SPACE          SECRETS   ACTIVE
  universidad    5         *
  trabajo        2
  personal       8
```

Output de `vext list` (siempre muestra dónde estás):

```
  space: universidad

  NAME      TYPE       CREATED
  gmail     account    2026-01-01
  moodle    account    2026-01-02
```

---

## Estado sin `vext use`

Si el usuario nunca corrió `vext use`, `active_space = ""`.
Los secrets sin space pertenecen al espacio raíz `""`.
`vext list` muestra los secrets del espacio `""`.
`vext list --all` muestra absolutamente todo.

Esto garantiza compatibilidad total hacia atrás: los secrets existentes siguen
funcionando sin que el usuario tenga que hacer nada.

---

## Organización del código

### `source/shared/storages/`

Un solo paquete, archivos separados por entidad con prefijo. Sin subpaquetes:
los imports quedan simples y `initalizer.go` sigue viendo todos los schemas.

```
shared/storages/
  configs.go
  initalizer.go
  secrets_schema.go        ← rename de schemas.go
  secrets_mapper.go        ← rename de mapper.go
  secrets_repository.go    ← rename de repository.go
  spaces_schema.go         ← nuevo (SpaceRecord + MetaRecord)
  spaces_repository.go     ← nuevo
  state_repository.go      ← nuevo
```

### `source/funcs/`

Agrupado por entidad. Los use cases de secrets se mueven a `funcs/secrets/`;
los de spaces viven en `funcs/spaces/` con un `command.go` raíz que registra
los subcomandos Cobra. `vext use` queda en su propio package al nivel de
`funcs/` porque es un comando top-level, no un subcomando de `spaces`.

```
source/funcs/
  deps.go
  secrets/
    addsecret/
      command.go
      provider.go
      collectors.go
    updsecret/
      command.go
      provider.go
      collectors.go
    getsecrets/
      command.go
      provider.go
    listsecrets/
      command.go
      provider.go
      displayer.go
    dropsecret/
      command.go
      provider.go
    rensecret/
      command.go
      provider.go
    rotasecret/
      command.go
      provider.go
  spaces/
    command.go             ← parent cobra "vext spaces", registra subcomandos
    addspace/
      command.go
      provider.go
    dropspace/
      command.go
      provider.go
    renspace/
      command.go
      provider.go
    listspaces/
      command.go
      provider.go
      displayer.go
  usespace/
    command.go
    provider.go
```

El único costo de este cambio es actualizar los imports en `main.go`:
`funcs/addsecret` → `funcs/secrets/addsecret`, etc. Son 7 líneas.

---

## Alcance del refactor

### Archivos nuevos

| Archivo | Descripción |
|---|---|
| `secrets/core/space.go` | `Space` struct + `SpaceRepository` + `StateRepository` interfaces |
| `shared/storages/spaces_schema.go` | `SpaceRecord` + `MetaRecord` |
| `shared/storages/spaces_repository.go` | implementación GORM de `SpaceRepository` |
| `shared/storages/state_repository.go` | implementación GORM de `StateRepository` |
| `funcs/usespace/command.go` + `provider.go` | `vext use` |
| `funcs/spaces/command.go` | parent command, registra subcomandos |
| `funcs/spaces/addspace/` | `vext spaces add` |
| `funcs/spaces/dropspace/` | `vext spaces drop` |
| `funcs/spaces/renspace/` | `vext spaces ren` |
| `funcs/spaces/listspaces/` | `vext spaces` (list) |

### Archivos renombrados / movidos

| Antes | Después |
|---|---|
| `shared/storages/schemas.go` | `secrets_schema.go` |
| `shared/storages/mapper.go` | `secrets_mapper.go` |
| `shared/storages/repository.go` | `secrets_repository.go` |
| `funcs/addsecret/` | `funcs/secrets/addsecret/` |
| `funcs/updsecret/` | `funcs/secrets/updsecret/` |
| `funcs/getsecrets/` | `funcs/secrets/getsecrets/` |
| `funcs/listsecrets/` | `funcs/secrets/listsecrets/` |
| `funcs/dropsecret/` | `funcs/secrets/dropsecret/` |
| `funcs/rensecret/` | `funcs/secrets/rensecret/` |
| `funcs/rotasecret/` | `funcs/secrets/rotasecret/` |

### Archivos modificados

| Archivo | Cambio |
|---|---|
| `secrets/core/domain.go` | agregar `Space string` a `Secret` |
| `secrets/core/repository.go` | `GetByName`, `Rename`, `Delete`, `List` reciben `space` |
| `secrets_schema.go` | `SpaceID int64` + cambio de unique index a `(space_id, name)` |
| `secrets_repository.go` | queries usan join con tabla `spaces` |
| `shared/storages/initalizer.go` | `AutoMigrate` incluye `SpaceRecord`, `MetaRecord` |
| `funcs/deps.go` | agregar `SpaceRepo`, `StateRepo`, `ActiveSpace` |
| `funcs/secrets/listsecrets/command.go` | agregar flag `--all` |
| `funcs/secrets/listsecrets/provider.go` | pasar `deps.ActiveSpace` (o `""` si `--all`) |
| `funcs/secrets/listsecrets/displayer.go` | mostrar space activo en header |
| `funcs/secrets/addsecret/provider.go` | pasar `deps.ActiveSpace` a `core.Secret{}` |
| `funcs/secrets/getsecrets/provider.go` | pasar `deps.ActiveSpace` a `GetByName` |
| `funcs/secrets/dropsecret/provider.go` | pasar `deps.ActiveSpace` a `Delete` |
| `funcs/secrets/rensecret/provider.go` | pasar `deps.ActiveSpace` a `Rename` |
| `funcs/secrets/updsecret/provider.go` | pasar `deps.ActiveSpace` a `GetByName` + `Update` |
| `funcs/secrets/rotasecret/provider.go` | pasar `deps.ActiveSpace` a `GetByName` + `Update` |
| `main.go` | actualizar imports + registrar `usespace` y `spaces` |

Los cambios a los providers son todos iguales: reemplazar el nombre hardcodeado
por `deps.ActiveSpace`. Mecánico, no hay lógica nueva.

---

## Lo que NO haría (todavía)

**`movesecret`** — mover un secret entre spaces. El flujo de `updsecret` podría
cubrir esto si le permitís cambiar el space. Por ahora, no es urgente.

**Auto-create de spaces** — si corrés `vext add gmail` sin haber hecho `vext use`,
no crear el space automáticamente. Mejor fallar con un mensaje claro:
`no active space — run 'vext spaces add <name>' and 'vext use <name>' first`.
Esto previene secrets huérfanos en el espacio `""` por accidente.
(O revisar: quizás el espacio `""` sí es válido para usuarios que no quieren spaces.)

**Tags / spaces múltiples por secret** — M:N es complejidad que no aporta
suficiente valor para un gestor personal.

---

## Orden de implementación

1. `spaces` y `meta` tables + `SpaceRepository` + `StateRepository` (infraestructura)
2. `vext spaces add/list` + `vext use` (los dos comandos base)
3. Cambio de unique index en `secrets` + join en `GORMRepository`
4. Propagar `deps.ActiveSpace` a todos los providers existentes
5. `vext list --all` + header de space activo en output
6. `vext spaces drop/ren` (gestión completa de spaces)
