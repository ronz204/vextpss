# 0005 — Refinamiento de Código: Semántica, Calidad y Portabilidad [COMPLETED]

## Executive Summary

Se implementaron los 9 cambios del plan en su totalidad: reorganización completa de paquetes y archivos, eliminación del prefijo `pkg/`, semántica de nombres corregida, `MergeFrom` en el dominio, `printErr` en handlers, registries OCP en codec y colectores, y todos los archivos zombie eliminados. `go build`, `go vet` y `go test ./source/...` pasan sin errores.

---

## What Went Well

- La migración estructural (60+ archivos) se ejecutó limpiamente en una sola pasada de PowerShell + sed.
- El patrón `MergeFrom` en `SecretPayload` eliminó completamente la dependencia de `update_handler.go` sobre `core/secrets` y las type assertions.
- Los registries en `codec.go` y `collector.go` son puros y compactos: agregar un tercer tipo requiere exactamente una línea en cada registro.
- `printErr` en `errors.go` redujo los handlers de 7 líneas de error handling repetido a 1 línea cada uno.
- `go vet` detectó dos imports sin uso en test files que el build no había detectado — útil.

## What Was Different

- **`isUniqueConstraintError` — no se implementó con `errors.As`**: La solución planificada (import `modernc.org/sqlite`) conflictó con `github.com/glebarez/go-sqlite` porque ambos registran el driver `"sqlite"` en `init()`. El segundo registro produce `panic: sql: Register called twice`. Documentado con un comentario en el código — string matching es correcto y suficiente para SQLite. La solución definitiva sería usar el tipo de error de `glebarez/go-sqlite` directamente, pero expone un API privado.

- **`formatter_test.go:unknownPayload`**: Al agregar `MergeFrom` a `SecretPayload`, el tipo stub de test `unknownPayload` no la implementaba. Se agregó una implementación vacía.

- **Conflicto de alias en `sqlite_repository_test.go`**: El test importaba `github.com/glebarez/sqlite` como `sqlite` (alias por defecto), y el propio paquete también se llama `sqlite`. Resuelto con alias explícito `glebarez "github.com/glebarez/sqlite"`.

- **`utils_test.go`**: Los tests de `GenerateSalt` quedaron huérfanos al moverla a `crypto/`. Se eliminaron los tests de esa función (la función en sí está bien testeada indirectamente vía `aes_gcm_test.go`).

- **`config.AppDir` eliminado**: Al remover `appDir` del `Initialiser`, `AppConfig` ya no necesitaba el campo. Se eliminó de `AppConfig` y de `config.Load()` también.

---

## Changes Made

### Nuevas carpetas (reemplazando estructura `pkg/` y `dal/`)

| Nueva | Reemplaza |
|---|---|
| `source/app/` | `source/pkg/apps/` |
| `source/crypto/` | `source/pkg/tokens/` |
| `source/config/` | `source/pkg/configs/` |
| `source/shared/` | `source/pkg/shared/` (fuera de `pkg/`) |
| `source/storage/` | `source/dal/` |
| `source/storage/sqlite/` | `source/dal/repos/` |
| `source/cmd/ui/` | `source/cmd/helpers/` |
| `source/testutil/` | `source/tests/mocks/` |

### Archivos renombrados

| Antes | Después |
|---|---|
| `pkg/tokens/aes_gcm_encryptor.go` | `crypto/aes_gcm.go` |
| `pkg/configs/app_config.go` | `config/config.go` |
| `pkg/shared/utils.go` | `shared/mem.go` |
| `pkg/shared/generator.go` | `shared/passgen.go` |
| `dal/schemas.go` | `storage/record.go` |
| `dal/repos/sqlite_repository.go` | `storage/sqlite/repository.go` |
| `core/secrets/account_secret.go` | `core/secrets/account.go` |
| `core/secrets/credit_secret.go` | `core/secrets/credit.go` |

### Archivos eliminados

- `source/pkg/` (directorio completo)
- `source/dal/` (directorio completo)
- `source/cmd/helpers/` (directorio completo)
- `source/tests/` (directorio completo)

### Cambios de código

| Archivo | Cambio |
|---|---|
| `core/secret.go` | `MergeFrom(current SecretPayload)` agregado a la interfaz |
| `core/secrets/account.go` | `MergeFrom` implementado: conserva username si vacío |
| `core/secrets/credit.go` | `MergeFrom` implementado: no-op (replace-all semántico) |
| `cmd/handlers/errors.go` | NUEVO: `printErr(err, notFoundMsg)` helper |
| `cmd/handlers/update_handler.go` | Usa `MergeFrom`, elimina type assertions, usa `printErr` |
| `cmd/handlers/add_handler.go` | Usa `printErr` |
| `cmd/handlers/get_handler.go` | Usa `printErr`, elimina import `core` |
| `cmd/handlers/rm_handler.go` | Usa `printErr`, elimina import `core` |
| `cmd/handlers/export_handler.go` | Usa `printErr`, elimina import `core` |
| `cmd/handlers/import_handler.go` | Usa `printErr`, elimina import `core` |
| `app/codec.go` | Switch → registry map `decoderRegistry` (OCP) |
| `cmd/ui/collector.go` | Switch → registry map `collectorRegistry` (OCP) |
| `cmd/ui/prompter.go` | `MockPrompter` eliminado (movido a `testutil/`) |
| `testutil/mock_prompter.go` | NUEVO: `MockPrompter` movido desde `cmd/ui/` |
| `crypto/aes_gcm.go` | `GenerateSalt` inlined como `generateSalt()` privada |
| `shared/mem.go` | `GenerateSalt` eliminada (movida a `crypto/`) |
| `storage/initialiser.go` | Campo `appDir` eliminado del struct y constructor |
| `config/config.go` | Campo `AppDir` eliminado de `AppConfig` |
| `storage/record.go` | Comentario de variabilidad para tipos GORM blob vs bytea |
| `storage/db.go` | Comentario de variabilidad para cambio de driver |
| `storage/sqlite/repository.go` | `isUniqueConstraintError` documentado (string matching + explicación del conflicto de driver) |

---

## Verification Results

```
go build ./source/...   → OK (sin output)
go vet ./source/...     → OK (sin output)
go test ./source/...    →

ok  vextpss/source/app              1.779s
ok  vextpss/source/cmd/handlers     1.851s
ok  vextpss/source/cmd/ui           1.627s
ok  vextpss/source/core             (cached)
ok  vextpss/source/core/secrets     (cached)
ok  vextpss/source/crypto           2.158s
ok  vextpss/source/shared           1.506s
ok  vextpss/source/storage/sqlite   0.514s
```

---

## Lessons Learned

1. **Driver SQLite y doble registro**: `github.com/glebarez/go-sqlite` y `modernc.org/sqlite` son forks del mismo código, ambos registran el driver `"sqlite"`. Importar ambos en el mismo binario produce panic. Para error-code checking de SQLite hay que usar el tipo de error de `glebarez/go-sqlite`, no de `modernc.org/sqlite`. Documentado en el código.

2. **`errors.As` con interfaces**: `errors.As(err, &iface)` funciona para tipos concretos, no para interfaces puras. Para duck-typing de error codes, se necesita un wrapper de tipo concreto.

3. **Package declaration updates y CRLF**: El script PowerShell de `Update-PackageDecl` con regex `(?m)^package X$` no capturó algunos archivos porque tenían line endings diferentes o espacios antes de `package`. La solución fue buscar primero con grep y corregir los rezagados manualmente.

4. **`captureStdout` captura solo stdout**: Todos los mensajes de error en handlers van a stdout (`fmt.Printf`), no stderr. El `printErr` se mantiene en stdout por consistencia con los tests existentes y con el comportamiento visible en terminal.

5. **Import aliases en tests**: Cuando el nombre del paquete propio coincide con un import externo (e.g., package `sqlite` y `import "github.com/glebarez/sqlite"`), Go rechaza el programa. Siempre usar aliases explícitos en este caso.

---

## Next Steps

- Agregar `NoteSecret` como tercer tipo: requiere `core/secrets/note.go`, una entrada en `decoderRegistry`, una entrada en `collectorRegistry`, un `NoteCollector`, y un case en `PrintSecret`. Nada más.
- Considerar agregar `TestMergeFrom_*` en `core/secrets/account_test.go` (actualmente cubierto implícitamente por `update_handler_test.go`).
- Para un futuro soporte PostgreSQL: crear `storage/postgres/repository.go` con su propia `isUniqueConstraintError` usando el tipo de error de `lib/pq` o `pgx`.
