# 0004 — Refinamiento de Arquitectura [COMPLETED]

## Executive Summary

Se implementaron los 7 puntos de mejora del plan en su totalidad: interfaces movidas a `core/`, codec centralizado, `UpdateSecretUC` rediseñado para soporte polimórfico (incluyendo `CreditSecret`), constantes de tipo, split de collectors, guards de nil en todos los handlers, y tests actualizados. `go build`, `go vet` y `go test ./source/...` pasan sin errores.

---

## What Went Well

- La migración de interfaces a `core/` fue limpia: `SecretRepository` y `Encryptor` se movieron sin circular imports.
- El codec (`pkg/apps/codec.go`) resuelve la duplicación de `deserialisePayload` sin necesidad de un paquete nuevo.
- `UpdateSecretUC` rediseñado correctamente: acepta `NewPayload core.SecretPayload`, verifica contraseña por decryption, y bloquea cambios de tipo accidentales.
- `UpdateHandler` orquesta `retrieveUC` + `updateUC` con naturalidad: master password primero → retrieve verifica → collector recolecta → UC actualiza.
- Los 4 archivos de test de export/import solo necesitaron cambiar el import `dal` → `core` para `FullRecord`. Lógica intacta.
- `guardInit` con patrón `bool` evita el gotcha de Go con interfaces nil.

## What Was Different

- **Tests de export/import con `dal.FullRecord`:** cuatro archivos de tests seguían referenciando `dal.FullRecord` después de la migración. Corregidos actualizando imports a `core.FullRecord`.
- **`encryptedAccountPayload` en tests**: La función auxiliar `encryptedAccountPayload` existía en `update_secret_uc_test.go` y también en `export_secrets_uc_test.go` (con el mismo nombre). Al reescribir `update_secret_uc_test.go` sin esa función (era del diseño anterior), se resolvió la colisión sola.
- **`collectors.go` no puede eliminarse:** Al ser un archivo existente sin mecanismo de borrado, se reemplazó con un comment-header que referencia los nuevos archivos. El paquete Go solo acepta una declaración por símbolo — al mover todo a archivos separados y vaciar `collectors.go`, el compilador queda satisfecho.

## Changes Made

### Archivos nuevos
| Archivo | Contenido |
|---------|-----------|
| `core/encryptor.go` | Interfaz `core.Encryptor` |
| `core/repository.go` | Interfaz `core.SecretRepository` + tipo `core.FullRecord` |
| `core/secrets/types.go` | Constantes `TypeAccount`, `TypeCredit` |
| `pkg/apps/codec.go` | Funciones internas `marshalPayload`, `unmarshalPayload` |
| `cmd/handlers/guard.go` | Función `guardInit(bool) bool` |
| `cmd/helpers/collector.go` | Interfaz `SecretCollector` + `CollectorOptions` + factory |
| `cmd/helpers/account_collector.go` | `AccountCollector` |
| `cmd/helpers/credit_collector.go` | `CreditCollector` |

### Archivos modificados — lógica
| Archivo | Cambio |
|---------|--------|
| `pkg/apps/update_secret_uc.go` | Rediseñado: `UpdateSecretRequest.NewPayload core.SecretPayload`; polimórfico |
| `cmd/handlers/update_handler.go` | Rediseñado: recibe `retrieveUC + updateUC`; master password primero |
| `core/secrets/account_secret.go` | `GetType()` usa `TypeAccount` |
| `core/secrets/credit_secret.go` | `GetType()` usa `TypeCredit` |
| `main.go` | `NewUpdateHandler(retrieveUC, updateUC, prompter)` |

### Archivos modificados — refactor de imports
| Archivo | Cambio |
|---------|--------|
| `pkg/apps/retrieve_secret_uc.go` | Elimina `dal`, `pkg/tokens`; usa `core.*` |
| `pkg/apps/store_secret_uc.go` | Idem + usa `marshalPayload` del codec |
| `pkg/apps/list_secrets_uc.go` | Elimina `dal` |
| `pkg/apps/delete_secret_uc.go` | Elimina `dal` |
| `pkg/apps/export_secrets_uc.go` | Elimina `dal`, `pkg/tokens` |
| `pkg/apps/import_secrets_uc.go` | Elimina `dal`, `pkg/tokens` |
| `dal/repos/sqlite_repository.go` | `dal.FullRecord` → `core.FullRecord` |
| `tests/mocks/mock_repo.go` | Elimina `dal`; usa `core.SecretRepository`, `core.FullRecord` |

### Archivos vaciados
| Archivo | Razón |
|---------|-------|
| `dal/repository.go` | Interfaz movida a `core/repository.go` |
| `pkg/tokens/encryptor.go` | Interfaz movida a `core/encryptor.go` |
| `cmd/helpers/collectors.go` | Contenido movido a 3 archivos separados |

### Handlers con guardInit
`add_handler.go`, `get_handler.go`, `list_handler.go`, `rm_handler.go`, `export_handler.go`, `import_handler.go` — todos tienen `if !guardInit(h.uc != nil) { return nil }` al inicio de `Handle`.

### Tests actualizados
| Archivo | Cambio |
|---------|--------|
| `pkg/apps/update_secret_uc_test.go` | Reescrito: usa `NewPayload`; agrega tests `CreditSecret` y `TypeMismatch` |
| `cmd/handlers/update_handler_test.go` | Reescrito: `newUpdateHandler` con `retrieveUC+updateUC`; orden de prompts actualizado |
| `pkg/apps/export_secrets_uc_test.go` | `dal.FullRecord` → `core.FullRecord` |
| `pkg/apps/import_secrets_uc_test.go` | `dal.FullRecord` → `core.FullRecord` |
| `cmd/handlers/export_handler_test.go` | `dal.FullRecord` → `core.FullRecord` |
| `cmd/handlers/import_handler_test.go` | `dal.FullRecord` → `core.FullRecord` |

---

## Verification Results

```
go build ./source/...   → OK (sin output)
go vet ./source/...     → OK (sin output)
go test ./source/...    → todos los paquetes: OK
  cmd/handlers          OK   1.143s
  cmd/helpers           OK   (cached)
  core                  OK   (cached)
  core/secrets          OK   (cached)
  dal/repos             OK   (cached)
  pkg/apps              OK   1.135s
  pkg/shared            OK   (cached)
  pkg/tokens            OK   (cached)
```

---

## Lessons Learned

1. **Gotcha de nil en interfaces Go:** `requiresInit(uc any)` no funciona si `uc` es un puntero tipado nil — la interfaz `any` tiene type-info y no es nil. Solución: pasar el resultado del operador `!=` como `bool`. Documentado en `guard.go`.

2. **`FullRecord` en tests:** Cuando se mueve un tipo de paquete, los archivos de tests que referencian ese tipo (y que no fueron explícitamente identificados en el plan como "a modificar") también requieren actualización. Lección: al migrar un tipo, grep todos sus usos incluyendo `_test.go`.

3. **Import circular imposible para el codec:** El codec en `core/` violaría el ciclo `core → core/secrets → core`. La decisión de ponerlo en `pkg/apps/` como archivo interno es correcta y permanece.

4. **Orden de prompts en UpdateHandler:** Mover "Master Password" al inicio (antes que los campos nuevos) mejora la UX: el usuario verifica su acceso antes de ingresar todos los valores nuevos. Un error de contraseña no desperdicia el resto del flujo.

---

## Next Steps

- Considerar agregar `NoteSecret` (notas de texto cifradas) como tercer tipo — solo requiere: `core/secrets/note_secret.go`, un case en `codec.go`, un `NoteCollector`, y un case en `PrintSecret`.
- `isUniqueConstraintError` en `sqlite_repository.go` sigue usando string matching (no se implementó el cambio a `errors.As` por bajo ROI vs el riesgo de cambiar una dependencia transitiva). Aceptado como deuda técnica controlada.
