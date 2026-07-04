# Vext — Casos de Uso

## De qué habla este documento

Cada comando de Vext es un caso de uso concreto: una acción que el usuario quiere realizar sobre el vault. Este documento describe qué hace cada comando, cuál es su flujo interno paso a paso, qué errores puede encontrar, y las decisiones de diseño que no son obvias desde la firma del comando.

El árbol de comandos completo:

```
vext
├── add <name> [-t type]   Guardar un nuevo secreto
├── get <name>             Recuperar y mostrar un secreto
├── list                   Listar todos los secretos almacenados
├── update <name>          Actualizar las credenciales de un secreto existente
└── rm <name>              Eliminar un secreto del vault
```

---

## `add` — Guardar un nuevo secreto

```
vext add <name> [-t type]
```

### Uso

```
$ vext add github
Username: bob@example.com
Password: ••••••••
Master password: ••••••••
[✓] Secret "github" saved.

$ vext add visa-debito -t finance
Card number: 4111111111111111
Card PIN: ••••
Security code: •••
Expiration month (1-12): 12
Expiration year: 2027
Mobile banking username: bob
Mobile banking password: ••••••••
Virtual key: ••••••••
Cellphone: +1234567890
Master password: ••••••••
[✓] Secret "visa-debito" saved.
```

### Flujo interno

```
IsKnownType(type)
  → recolectar campos del payload (según tipo)
  → ReadSecret("Master password")
  → json.Marshal(payload) → plaintext
  → Encrypt(plaintext, master)
  → repo.Create(Secret{ Name, Type, Encrypted })
  → limpiar plaintext y master de memoria
```

1. **Validación de tipo** — antes de pedirle nada al usuario, se verifica que el tipo pedido (`account` por defecto, o el que se pase con `-t`) es un tipo conocido. Si no lo es, el comando falla inmediatamente sin prompts.

2. **Recolección del payload** — el registro de collectors (`registry.go`) mapea cada tipo a su función de recolección. Para `account`: dos prompts (usuario y contraseña). Para `finance`: nueve prompts agrupados en tarjeta y banca móvil. La función de recolección construye y valida el payload concreto antes de continuar.

3. **Contraseña maestra** — se pide *después* de todos los campos del payload. Si el usuario cancela en este punto, los campos ya ingresados se descartan sin haber tocado la base de datos.

4. **Serialización y cifrado** — el payload se serializa a JSON y ese `[]byte` es el que se cifra. El bloque cifrado resultante (`Encrypted`) viaja junto al nombre y el tipo en un `Secret`.

5. **Limpieza de memoria** — `plaintext` y `master` se ponen en cero con `defer memory.Cleaner` sin importar cómo termine la función.

### Errores posibles

| Error | Causa |
|---|---|
| `unknown secret type "..."` | Se pasó un `-t` con un tipo no registrado en `core.IsKnownType` |
| sentinels del payload | Campos vacíos o inválidos (ej. `ErrUsernameRequired`) |
| `secret already exists` | Ya existe un secreto con ese nombre en el vault |

---

## `get` — Recuperar y mostrar un secreto

```
vext get <name>
```

### Uso

```
$ vext get github
Master password: ••••••••
  Username         bob@example.com
  Password         ••••••••

$ vext get visa-debito
Master password: ••••••••
  Card
    Number           4111111111111111
    PIN              ••••
    Security Code    •••
    Expiration       12/2027

  Banking
    Username         bob
    Password         ••••••••
    Virtual Key      ••••••••
    Cellphone        +1234567890
```

### Flujo interno

```
repo.GetByName(name)
  → ReadSecret("Master password")
  → Decrypt(secret.Encrypted, master)
  → moon.Decode(secret.Type, plaintext)
  → payload.Display()
  → limpiar plaintext y master de memoria
```

1. **Búsqueda** — se consulta el repositorio por nombre. Si no existe, se devuelve `ErrNotFound` antes de pedirle la contraseña al usuario.

2. **Contraseña maestra** — se pide una sola vez, para ese secreto. Vext no tiene sesión global ni contraseña maestra "de la app" — cada operación que accede a datos cifrados requiere la contraseña en ese momento.

3. **Descifrado** — si la contraseña es incorrecta, `Decrypt` devuelve `ErrDecryptionFailed` sin distinguir entre contraseña mal escrita y datos corruptos. El error es deliberadamente ambiguo.

4. **Deserialización polimórfica** — `moon.Decode(type, plaintext)` consulta el registry en `moon/factory.go`, instancia el tipo concreto correcto, y desmarshala el JSON en él. El resultado es un `core.Payload` que sabe imprimirse a sí mismo.

5. **Display** — se llama `payload.Display()` directamente. Cada tipo controla cómo se presentan sus campos — esa lógica vive en el aggregate, no en la capa de presentación.

### Errores posibles

| Error | Causa |
|---|---|
| `secret not found` | No existe un secreto con ese nombre |
| `wrong password or corrupted data` | Contraseña maestra incorrecta o datos corruptos en disco |
| `payload is corrupt or truncated` | El JSON descifrado no es un payload válido para ese tipo |
| `unknown secret type "..."` | El tipo almacenado no tiene factory registrado en `moon` |

---

## `list` — Listar todos los secretos

```
vext list
```

### Uso

```
$ vext list
  NAME            TYPE       CREATED
  github          account    2025-06-01
  visa-debito     finance    2025-06-10
  api-key-prod    account    2025-06-15
```

### Flujo interno

```
repo.List()
  → terminal.PrintSecretsTable(secrets)
```

`list` es el único comando que no requiere contraseña maestra. Lo que muestra son los **metadatos** del secreto — nombre, tipo y fecha de creación — que se guardan en texto plano en la base de datos. El contenido cifrado nunca se toca.

Este es un tradeoff explícito: la lista de *qué* secretos existen no está cifrada, lo que permite listarlos sin autenticación. El *contenido* de cada secreto — las credenciales reales — sí está cifrado y nunca aparece en esta vista.

La consulta usa una proyección selectiva: `SELECT id, name, type, created_at, updated_at`. El campo `ciphertext` y el `metadata` nunca se leen en este camino.

---

## `update` — Actualizar un secreto existente

```
vext update <name>
```

### Uso

```
$ vext update github
Username: bob@newdomain.com
Password: ••••••••
Master password: ••••••••
[✓] Secret "github" updated.
```

### Flujo interno

```
repo.GetByName(name)
  → recolectar campos del payload (según tipo del secreto existente)
  → ReadSecret("Master password")
  → json.Marshal(payload) → plaintext
  → Encrypt(plaintext, master)
  → repo.Update(Secret{ Name, Type, Encrypted })
  → limpiar plaintext y master de memoria
```

1. **Lookup previo** — se busca el secreto existente para obtener su tipo. El tipo no puede cambiarse en un update; si querés cambiar de `account` a `finance`, tenés que borrar y crear de nuevo.

2. **Nueva contraseña maestra** — update no descifra el secreto actual para mostrarlo: pide todos los campos desde cero con los valores nuevos. La contraseña maestra que se pide es la *nueva* — puede ser distinta de la que se usó para cifrarlo originalmente. Esto significa que update es también un cambio de contraseña maestra implícito si así se quiere.

3. **Reencifrado completo** — se genera un salt y nonce nuevos para el nuevo ciphertext. El registro anterior en la base de datos se reemplaza completamente.

> La limitación actual de update — que pide todos los campos desde cero en lugar de mostrar los valores actuales para editar solo los que cambian — está identificada. El plan para resolverlo es el paquete `shared/collectors/` con firmas `CollectAccount(p, defaults *Account)` y los métodos `ReadLineOrKeep`/`ReadSecretOrKeep`/`ReadIntegerOrKeep` del Prompter (ya implementados). Ese refactor está documentado en `planning_collect.md`.

### Errores posibles

| Error | Causa |
|---|---|
| `secret not found` | No existe un secreto con ese nombre |
| sentinels del payload | Campos vacíos o inválidos en la nueva entrada |

---

## `rm` — Eliminar un secreto

```
vext rm <name>
```

### Uso

```
$ vext rm github
Delete secret "github" [y/N]: y
[✓] Secret "github" deleted.

$ vext rm github
Delete secret "github" [y/N]: n
$
```

### Flujo interno

```
Confirm("Delete secret \"name\"")
  → repo.Delete(name)
  → terminal.Success(...)
```

1. **Confirmación explícita** — antes de tocar la base de datos, se pide confirmación interactiva. La respuesta por defecto es `N` — un Enter vacío cancela sin efecto. Esto previene eliminaciones accidentales, especialmente relevante porque `rm` es **irreversible**: no hay papelera, no hay undo.

2. **Cancelación silenciosa** — si el usuario responde `n` (o presiona Enter), el comando termina con exit code 0 sin mensaje. No hay "operación cancelada" porque no ocurrió ningún error.

3. **No requiere contraseña maestra** — `rm` elimina el registro de la base de datos sin descifrar. La contraseña maestra no participa porque el contenido del secreto nunca se lee.

### Errores posibles

| Error | Causa |
|---|---|
| `secret not found` | No existe un secreto con ese nombre (se evalúa dentro de `Delete`) |

---

## El modelo de autenticación: sin sesión global

Una decisión que atraviesa todos los comandos con cifrado: **no existe una sesión global ni una contraseña maestra "de la app"**.

Cada operación que necesita acceder al contenido cifrado (`get`, `update`) pide la contraseña maestra en ese momento, para ese secreto. Esto tiene consecuencias concretas:

- **No hay estado entre comandos.** Cerrar y abrir la terminal no cambia nada — no hay token de sesión que expirar ni cachear.
- **Contraseña por secreto, no por vault.** Dos secretos distintos pueden estar cifrados con contraseñas maestras distintas. Vext no sabe ni necesita saber si son iguales — cada `Encrypted` es autosuficiente.
- **El costo es la UX de `update`.** Sin acceso al contenido actual, update no puede mostrar los valores existentes para editar solo los que cambian. El tradeoff fue consciente: la implementación correcta del modo update (con `defaults`) está pendiente.

Esta arquitectura es la que hace que la seguridad de Vext sea composable: comprometer un secreto no compromete ningún otro, ni siquiera si comparten la misma contraseña maestra — el salt único por operación garantiza claves derivadas distintas para cada ciphertext.
