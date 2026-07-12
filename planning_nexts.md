# Vext — Ideas para la siguiente etapa

## Estado actual

El vault base está completo y bien construido:
`add` · `get` · `list` · `update` (con decrypt-first) · `rm`
Modelo criptográfico sólido. Dominio limpio. Dependencias controladas.

Lo que existe sin conectar: `shared/passgen/generator.go` — generador de contraseñas
listo para usar con `crypto/rand`, sin ningún comando que lo llame.

---

## Ideas por categoría

### 1. Comandos de alta prioridad (bajo costo, valor inmediato)

#### `gen` — generador de contraseñas

Infraestructura ya existe. Solo falta el comando.

```
vext gen [--length 20] [--symbols]
```

Genera una contraseña criptográficamente segura y la imprime.
Extensión natural: `--copy` para mandarla directo al clipboard.

**Complejidad:** mínima — un `command.go` + `provider.go` que llama `passgen.Generate`.

---

#### ~~`rename <old> <new>` — renombrar un secreto~~ ✓ implementado

---

#### Shell completion

Cobra tiene soporte nativo. Un solo comando expone completions para bash, zsh, fish, PowerShell:

```
vext completion bash > /etc/bash_completion.d/vext
```

Para que el completion de `get`/`update`/`rm` sugiera nombres reales de secretos hay que
registrar un `ValidArgsFunction` que haga `repo.List` — eso sí requiere instanciar deps en el completion.
Como mínimo, el completion de subcomandos ya es útil sin eso.

**Complejidad:** baja — `root.AddCommand(root.GenBashCompletionCmd())` o equivalente.

---

### 2. UX del vault (calidad de vida)

#### `clip <name> [--field password]` — copiar al portapapeles

En vez de mostrar una contraseña en pantalla (que queda en scroll history),
copiarla al clipboard y limpiarla después de N segundos.

```
vext clip github
[✓] Password copied. Clears in 30s.
```

Necesita una librería de clipboard (`atotto/clipboard` — pura Go, cross-platform).
El timeout de limpieza requiere un goroutine con `time.AfterFunc` que sobrescriba el clipboard.

**Valor:** muy alto para uso diario. Es la UX que diferencia un gestor real de un script.
**Complejidad:** media — nueva dependencia, manejo de tiempo, plataforma.

---

#### `search <query>` o filtro en `list`

```
vext list --filter visa
vext search visa
```

`WHERE name LIKE '%query%'` en GORM. Nada más.
Alternativa más simple: filtrar en memoria sobre el resultado de `List`.

**Complejidad:** mínima.

---

#### Indicador de antigüedad en `list`

Agregar una columna `AGE` o un marcador visual para secretos no actualizados en mucho tiempo:

```
  NAME          TYPE      CREATED     UPDATED
  github        account   2025-01-10  2025-01-10  ← 6 meses sin tocar
  visa-debito   finance   2025-06-10  2025-07-01
```

O una señal discreta: `[!]` para secretos con más de N días sin `update`.
El dato ya está: `UpdatedAt` vive en `SecretRecord`.

**Complejidad:** mínima — solo visual, dentro de `listsecrets/displayer.go`.

---

#### `-o json` en `list`

Útil para scripting, grep, jq. El campo `Encrypted` **no** se incluye — solo metadatos.

```
vext list -o json
[{"name":"github","type":"account","created":"2025-01-10","updated":"2025-01-10"}]
```

**Complejidad:** mínima — flag en el comando, `json.Marshal` sobre los metadatos.

---

### 3. Gestión de la master password

#### ~~`rotate <name>` — cambiar la master password de un secreto~~ ✓ implementado

---

#### `rekey` — rotar la master password de todos los secretos a la vez

Para el escenario "mi master password se comprometió, quiero cambiarla en todo el vault".

```
vext rekey
Current master: ••••••••
New master:     ••••••••
Confirm new:    ••••••••
Re-encrypting 12 secrets... [✓]
```

Requiere procesar todos los secretos en una transacción. Si falla a mitad, el vault no
debe quedar en estado inconsistente → se necesita soporte de transacciones en `Repository`
(actualmente no existe).

**Complejidad:** media — requiere extender la interfaz `Repository` con transacciones.

---

### 4. Tipos de secreto nuevos

#### `ssh` — pares de claves SSH

```go
type SSHKey struct {
    PrivateKey []byte `json:"private_key"`
    PublicKey  string `json:"public_key"`
    Comment    string `json:"comment"`
}
```

Pasos: constante en `core/shapes.go`, subpaquete `moon/sshkeys/`, collectors en `addsecret`/`updsecret`.
El collector para `add` puede leer el archivo de la clave privada desde disco con `p.ReadLine("Key path")`.

**Valor:** muy común en perfiles técnicos (el usuario objetivo de Vext).
**Complejidad:** baja — sigue exactamente el patrón existente.

---

#### `totp` — códigos 2FA (TOTP/HOTP)

Guardar el secreto TOTP (la seed) cifrado, y que `vext get` genere el código de 6 dígitos actual.

```
$ vext get github-2fa
Master password: ••••••••
  TOTP code   847 291   (válido por 23s)
```

Necesita: `github.com/pquerna/otp` (librería TOTP pura Go).
El payload almacena el secreto base32, el emisor, y el usuario. El código se genera en runtime.

**Valor:** alto — centraliza 2FA con el resto de credenciales.
**Complejidad:** media — nueva dependencia + lógica de tiempo en `Display()`.

---

### 5. Backup y portabilidad

#### `export` — backup cifrado del vault

```
vext export --out vault-backup.json.enc
Master password: ••••••••
[✓] 12 secrets exported to vault-backup.json.enc
```

Dos modalidades posibles:
- **Cifrado** (recomendada): exporta un JSON con todos los `Secret` (incluyendo `Encrypted` intacto). El archivo resultante es portátil — no necesita re-descifrar nada, el contenido ya está cifrado por secreto.
- **Plano** (riesgoso): descifra todo y exporta JSON legible. Solo útil para migración a otra herramienta.

La modalidad cifrada no requiere descifrar nada — solo serializar los `SecretRecord` tal cual.
Esto es una ventaja enorme en términos de seguridad del proceso de export.

**Complejidad:** baja para export cifrado.

---

#### `import` — restaurar desde backup

Contraparte de `export`. Lee el archivo y llama `repo.Create` para cada secreto.
Conflictos (nombre ya existe): flag `--skip` o `--overwrite`.

**Complejidad:** baja — si el formato de export está definido, import es directo.

---

### 6. Deuda técnica puntual

Cosas que no son features pero vale la pena registrar:

**`dropsecret` no verifica existencia antes de pedir confirmación.**
Actualmente pregunta "Delete secret X?" y solo después descubre que X no existe.
Fix: `GetByName` antes del `Confirm`.

**`passgen.Generate` devuelve `string`.**
La contraseña generada vive como string Go (inmutable, no se puede limpiar).
Debería devolver `[]byte` para poder pasarlo a `memory.Cleaner`.

**`moon/factory.go` y `core/shapes.go` son registros paralelos.**
Agregar un tipo nuevo exige tocar ambos. Podrían unificarse, aunque el costo de mantenerlos
separados es bajo mientras el catálogo sea pequeño.

**Errores de `cobra` sin formato.**
Cuando `run` devuelve un error, Cobra lo imprime directamente sin usar `terminal.Error`.
`SilenceErrors: true` + manejo explícito en `main` daría control sobre el formato de salida.

---

## Resumen priorizado

| Feature              | Valor | Costo | Infraestructura previa |
|----------------------|-------|-------|------------------------|
| `gen` command        | alto  | mínimo | `passgen` ya existe    |
| ~~`rename`~~         | alto  | mínimo | ✓                      |
| Shell completion     | medio | bajo   | Cobra built-in         |
| `search` / filtro    | medio | mínimo | —                      |
| `clip` (clipboard)   | alto  | medio  | nueva dep              |
| Antigüedad en `list` | bajo  | mínimo | `UpdatedAt` ya existe  |
| `-o json` en `list`  | medio | mínimo | —                      |
| ~~`rotate`~~         | alto  | bajo   | ✓                      |
| Tipo `ssh`           | alto  | bajo   | patrón claro           |
| Tipo `totp`          | alto  | medio  | nueva dep              |
| `export` cifrado     | alto  | bajo   | —                      |
| `import`             | medio | bajo   | depende de export      |
| `rekey`              | medio | medio  | requiere transacciones |

**Próximos tres más naturales:** `gen` (código ya existe), `rename` (operación obvia que falta), tipo `ssh` (usuario objetivo lo necesita).
