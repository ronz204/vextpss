# Vext — Local-First Password Manager

Vext es un gestor de contraseñas local escrito en Go. No hay servidores, no hay sincronización en la nube — los secretos viven cifrados en tu máquina y solo tú tienes la llave.

---

## Comandos

### Secretos

```
vext add <name> [-t type]     Agrega un secreto nuevo (tipo por defecto: account)
vext get <name>               Muestra un secreto descifrado
vext list                     Lista los secretos del space activo
vext upd <name>               Actualiza un secreto (muestra valores actuales, Enter para conservar)
vext drop <name>              Elimina un secreto
vext ren <old> <new>          Renombra un secreto
vext rota <name>              Rota la master password de un secreto
```

### Generador de contraseñas

```
vext gen [-l length] [-s]     Genera una contraseña criptográficamente segura
```

Flags:

| Flag | Por defecto | Descripción |
|---|---|---|
| `-l`, `--length` | 20 | Longitud de la contraseña |
| `-s`, `--symbols` | false | Incluir símbolos (`!@#$%^&*...`) |

### Spaces _(próximamente)_

Los spaces son namespaces lógicos para agrupar secretos. Funcionan como branches en git: siempre hay un space activo y todos los comandos operan sobre él sin flags adicionales.

```
vext use <space>              Cambia el space activo
vext spaces                   Lista todos los spaces
vext spaces add <name>        Crea un space
vext spaces drop <name>       Elimina un space
vext spaces ren <old> <new>   Renombra un space
```

La base de datos ya incluye las tablas `spaces` y `meta`, y crea un space `default` en el primer arranque. Los comandos estarán disponibles en la siguiente iteración.

---

## Tipos de secreto

Vext usa un modelo polimórfico: todos los secretos comparten el mismo contenedor externo (`Secret`) y es solo dentro del blob cifrado donde la forma cambia según el tipo.

### `account`
Credencial de acceso estándar.

| Campo | Tipo | Notas |
|---|---|---|
| `username` | `string` | |
| `password` | `[]byte` | se limpia de memoria tras su uso |

### `finance`
Datos financieros compuestos por dos partes.

**Tarjeta (`card`)**

| Campo | Tipo | Notas |
|---|---|---|
| `number` | `string` | |
| `pin` | `[]byte` | se limpia de memoria tras su uso |
| `security_code` | `[]byte` | se limpia de memoria tras su uso |
| `expiration_month` | `int` | 1–12 |
| `expiration_year` | `int` | |

**Banca móvil (`mobile`)**

| Campo | Tipo | Notas |
|---|---|---|
| `username` | `string` | |
| `password` | `[]byte` | se limpia de memoria tras su uso |
| `virtual_key` | `[]byte` | se limpia de memoria tras su uso |
| `cellphone` | `string` | |

---

## Cifrado

El cifrado de cada secreto sigue un pipeline de dos etapas:

**1. Derivación de clave — Argon2id**

La contraseña maestra no se usa directamente como clave criptográfica. Primero pasa por Argon2id, una KDF diseñada para ser costosa tanto en tiempo como en memoria, lo que hace inviable la fuerza bruta con GPUs. Cada cifrado genera un salt aleatorio, por lo que la misma contraseña produce claves distintas para cada secreto.

```
contraseña maestra + salt aleatorio → Argon2id → clave de 256 bits
```

Configuración por defecto: 3 iteraciones, 64 MB de RAM, 2 threads, clave de 32 bytes.

**2. Cifrado autenticado — AES-256-GCM**

La clave derivada cifra el payload con AES-256-GCM, que además de confidencialidad garantiza integridad: si un solo byte del ciphertext es alterado, el descifrado falla por completo. Un nonce aleatorio se genera por operación para garantizar que la misma clave nunca cifre dos bloques de la misma manera.

```
payload (JSON) + nonce aleatorio + clave → AES-256-GCM → ciphertext + tag
```

**Resultado**

Todo lo que produce un cifrado queda empaquetado en un struct `Encrypted`:

```go
type Encrypted struct {
    Algorithm  string // "aes-gcm-argon2id"
    Ciphertext []byte // datos cifrados + tag de autenticación
    Metadata   []byte // salt y nonce serializados
}
```

`Algorithm` permite que en el futuro convivan múltiples implementaciones sin romper datos existentes. `Metadata` es opaco para el dominio — solo el encryptor sabe cómo interpretarlo.

**Seguridad en memoria**

La clave derivada se sobreescribe con ceros en cuanto termina su uso (`defer memory.Cleaner(key)`), sin esperar al garbage collector. Los campos sensibles de los payloads (`Password`, `Pin`, `VirtualKey`, etc.) son `[]byte` por esta misma razón — a diferencia de `string`, pueden limpiarse explícitamente.

---

## Dominio

El paquete `secrets/core/` es el núcleo del dominio — no importa nada de SQLite, GORM, ni detalles de cifrado. Solo define contratos:

- **`Secret`** — metadatos (nombre, tipo, space, fechas) + `Encrypted` opaco.
- **`Payload`** — interfaz que implementa cada tipo de secreto (`Display()`, `Validate()`).
- **`Encryptor`** — interfaz de cifrado; la implementación concreta (`aesgcm`) vive en `shared/`.
- **`SecretRepository`** — interfaz de persistencia para secretos.
- **`SpaceRepository`** — interfaz de persistencia para spaces.
- **`StateRepository`** — interfaz para leer y escribir el space activo.

---

## Estructura

```
source/
├── main.go                        # wiring: deps + cobra root
├── funcs/
│   ├── deps.go                    # Deps: repos, cypher, prompter, ActiveSpace
│   └── secrets/
│       ├── addsecret/             # vext add
│       ├── updsecret/             # vext upd
│       ├── getsecrets/            # vext get
│       ├── listsecrets/           # vext list
│       ├── dropsecret/            # vext drop
│       ├── rensecret/             # vext ren
│       ├── rotasecret/            # vext rota
│       └── gensecret/             # vext gen
├── secrets/
│   ├── core/
│   │   ├── secrets.go             # Secret, Payload, SecretRepository
│   │   ├── encryptor.go           # Encrypted, Encryptor
│   │   ├── space.go               # Space, SpaceRepository
│   │   ├── state.go               # StateRepository
│   │   └── shapes.go              # constantes de tipo e IsKnownType
│   └── moon/                      # payloads concretos
│       ├── accounts/              # Account{Username, Password}
│       ├── finances/              # Finance{Card, Mobile}
│       └── factory.go
└── shared/
    ├── cyphers/aesgcm/            # AES-256-GCM + Argon2id
    ├── storages/                  # GORM + SQLite
    │   ├── secrets_schema.go
    │   ├── secrets_mapper.go
    │   ├── secrets_repository.go
    │   ├── spaces_schema.go       # SpaceRecord + MetaRecord
    │   ├── spaces_mapper.go
    │   ├── spaces_repository.go
    │   ├── state_schema.go
    │   ├── state_mapper.go
    │   ├── state_repository.go
    │   ├── initalizer.go
    │   └── configs.go
    ├── terminal/                  # Prompter, helpers (Success, Info, Error)
    ├── memory/                    # Cleaner — limpieza segura de slices sensibles
    └── passgen/                   # generador criptográfico de contraseñas

context/
├── concepts.md                    # criptografía y decisiones de diseño
├── modeling.md                    # modelado del dominio en código
├── database.md                    # SQLite + GORM, SecretRecord
└── usecases.md                    # flujo de cada comando
```

---

## Requisitos

- Go 1.24+
- `golang.org/x/crypto` (Argon2id)
