# 0001 — Architecture Refactor [COMPLETED]

**Fecha:** Junio 2026  
**Duración:** 1 sesión  
**Estado:** Completado — build OK, vet OK, 3 suites de tests pasan

---

## Executive Summary

Se implementó por completo la arquitectura Clean Architecture + DDD propuesta en el plan. El proyecto pasó de tener lógica mezclada en 5 archivos CLI a una separación clara en 4 capas con 25 archivos de código y 3 suites de tests unitarios. El binario compila y `go vet` pasa sin advertencias.

---

## What Went Well

- **Implementación limpia de todas las fases** sin necesidad de backtracking en diseño.
- **Interfaces del dominio bien definidas** — `SecretRepository` y `Encryptor` son contratos claros que el application layer consume sin saber nada de SQLite ni AES.
- **Tests unitarios sin I/O ni DB real** — `store_secret_uc_test.go` usa mocks puros de repo y encryptor, confirmando que el diseño es efectivamente testeable.
- **Seguridad mejorada** — `AccountSecret.Password` pasó de `string` a `[]byte`, eliminando el problema de copias inmutables en memoria.
- **`go vet` limpio** desde el primer build.
- **Eliminación total de código obsoleto** sin dejar importaciones huérfanas.

---

## What Was Different

### Conflicto de package names detectado inmediatamente

Al crear el nuevo `cmd/root.go` con `package cmd` pero antes de eliminar los viejos `cmd/add.go`, `cmd/get.go`, etc. (que usaban `package commands`), el compilador detectó un conflicto de nombres de paquete dentro del mismo directorio. Fue detectado en tiempo real gracias a los diagnósticos del IDE y resuelto eliminando los archivos obsoletos en ese momento (Fase 8 se adelantó).

**Lección:** en refactors que cambian el nombre de un paquete, conviene eliminar los archivos viejos antes de escribir los nuevos en el mismo directorio.

### Manejo de DB no inicializada en `main.go`

El plan original asumía que la DB siempre existe al arrancar. Se agregó lógica defensiva: si `storage.Open` falla y el archivo no existe, se arranca con `db = nil` y los use cases de data quedan como nil. Esto permite ejecutar `vext init` en un entorno fresco sin error de arranque.

### `StorageInitialiser` como interfaz en application layer

El plan mencionaba `init_storage_uc.go` pero no especificaba cómo evitar que el use case dependiera directamente de `storage.Open`. Se resolvió definiendo la interfaz `StorageInitialiser` en el paquete `application` e implementándola con `storage.Initialiser`. Esto mantiene la regla de dependencias unidireccionales.

### `pkg/models/secret_record.go` se mantuvo

El plan proponía moverlo a `pkg/models/secret_record.go` (ya estaba ahí). Solo se eliminó `payloads.go` ya que los payload types ahora viven en `pkg/domain/`.

---

## Changes Made

### Archivos creados

| Archivo | Capa | Descripción |
|---|---|---|
| `source/pkg/domain/errors.go` | Domain | Tipos de error canónicos (`ErrSecretNotFound`, `ErrAlreadyExists`, etc.) |
| `source/pkg/domain/secret.go` | Domain | Entidad `Secret` + interfaz `SecretPayload` |
| `source/pkg/domain/account_secret.go` | Domain | Value object `AccountSecret` con `Password []byte` |
| `source/pkg/domain/secret_repository.go` | Domain | Contrato `SecretRepository` |
| `source/pkg/domain/encryptor.go` | Domain | Contrato `Encryptor` |
| `source/pkg/shared/utils.go` | Shared | `Zero`, `GenerateSalt`, `Now` |
| `source/pkg/application/store_secret_uc.go` | Application | Use case: guardar secreto |
| `source/pkg/application/retrieve_secret_uc.go` | Application | Use case: obtener secreto |
| `source/pkg/application/list_secrets_uc.go` | Application | Use case: listar secretos |
| `source/pkg/application/delete_secret_uc.go` | Application | Use case: eliminar secreto |
| `source/pkg/application/init_storage_uc.go` | Application | Use case: inicializar storage + interfaz `StorageInitialiser` |
| `source/pkg/infrastructure/crypto/aes_gcm_encryptor.go` | Infrastructure | Implementación `AESGCMEncryptor` |
| `source/pkg/infrastructure/storage/db.go` | Infrastructure | `Open`, `Close`, `Migrate` |
| `source/pkg/infrastructure/storage/sqlite_repository.go` | Infrastructure | `SQLiteRepository` + `Initialiser` |
| `source/pkg/infrastructure/config/app_config.go` | Infrastructure | `AppConfig` con paths OS-aware |
| `source/cmd/ui/prompter.go` | Interface | `Prompter` interfaz + `CLIPrompter` + `MockPrompter` |
| `source/cmd/ui/formatter.go` | Interface | `PrintSecret`, `PrintSecretList` |
| `source/cmd/handlers/init_handler.go` | Interface | Handler para `vext init` |
| `source/cmd/handlers/add_handler.go` | Interface | Handler para `vext add` |
| `source/cmd/handlers/get_handler.go` | Interface | Handler para `vext get` |
| `source/cmd/handlers/list_handler.go` | Interface | Handler para `vext list` |
| `source/cmd/handlers/rm_handler.go` | Interface | Handler para `vext rm` |

### Archivos modificados

| Archivo | Cambio |
|---|---|
| `source/main.go` | Reescrito: wiring DI completo, DB lifecycle centralizado |
| `source/cmd/root.go` | Reescrito: `package cmd`, solo construye el root command |

### Archivos eliminados

| Archivo | Reemplazado por |
|---|---|
| `source/cmd/add.go` | `cmd/handlers/add_handler.go` + `application/store_secret_uc.go` |
| `source/cmd/get.go` | `cmd/handlers/get_handler.go` + `application/retrieve_secret_uc.go` |
| `source/cmd/init.go` | `cmd/handlers/init_handler.go` + `application/init_storage_uc.go` |
| `source/cmd/list.go` | `cmd/handlers/list_handler.go` + `application/list_secrets_uc.go` |
| `source/cmd/rm.go` | `cmd/handlers/rm_handler.go` + `application/delete_secret_uc.go` |
| `source/pkg/crypto/crypto.go` | `infrastructure/crypto/aes_gcm_encryptor.go` + `shared/utils.go` |
| `source/pkg/database/db.go` | `infrastructure/storage/db.go` |
| `source/pkg/database/path.go` | `infrastructure/config/app_config.go` |
| `source/pkg/database/queries.go` | `infrastructure/storage/sqlite_repository.go` |
| `source/pkg/models/payloads.go` | `domain/account_secret.go` |

### Tests creados

| Archivo | Qué cubre |
|---|---|
| `tests/domain/secret_test.go` | `AccountSecret.Validate`, errores de dominio, `GetType` |
| `tests/infrastructure/crypto/aes_gcm_encryptor_test.go` | Encrypt/Decrypt roundtrip, contraseña incorrecta, nonces únicos |
| `tests/application/store_secret_uc_test.go` | StoreSecretUC, DeleteSecretUC, ListSecretsUC con mocks puros |

---

## Resultados de verificación

```
$ go build ./source/...   → OK
$ go vet ./source/...     → OK
$ go test ./tests/...     → ok  vextpss/tests/application   (1.7s)
                             ok  vextpss/tests/domain        (1.6s)
                             ok  vextpss/tests/infrastructure/crypto (1.9s)
```

---

## Lessons Learned

1. **Eliminar archivos obsoletos antes de crear los nuevos en el mismo directorio** cuando se cambia el nombre de paquete — evita conflictos de package name detectados a mitad del trabajo.

2. **Las interfaces definidas en la capa que las consume** (no en la que las implementa) son la clave para respetar la regla de dependencias unidireccionales. `StorageInitialiser` en `application/` es el ejemplo.

3. **`[]byte` para todos los secretos en memoria** es una restricción de diseño que debe establecerse desde el dominio y propagarse hacia arriba — cambiarla después es costoso porque afecta serialización JSON (base64 en lugar de string plano).

4. **Los use cases con nil-safety en main** (db puede ser nil si `vext init` no se ha corrido) es un patrón que se repite en CLIs locales — vale la pena documentarlo.

---

## Next Steps

- [ ] Agregar test de integración para `SQLiteRepository` usando SQLite en memoria.
- [ ] Test del `add_handler` usando `MockPrompter`.
- [ ] Agregar `go test ./...` a un CI pipeline (GitHub Actions).
- [ ] Phase 2: implementar `CardSecret` y `NoteSecret` siguiendo el patrón establecido.
- [ ] Considerar logging estructurado (`slog`) reemplazando `fmt.Printf` en handlers.
- [ ] Advertir al usuario cuando corre `vext add/get/list/rm` sin haber corrido `vext init` (cuando `db == nil`).
