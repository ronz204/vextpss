# Vext — Local-First Password Manager

Vext es un gestor de contraseñas local escrito en Go. No hay servidores, no hay sincronización en la nube — los secretos viven cifrados en tu máquina y solo tú tienes la llave.

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

## Tipos de secreto

Vext usa un modelo polimórfico: todos los secretos comparten el mismo contenedor externo (`Secret`) y es solo dentro del blob cifrado donde la forma cambia según el tipo. Agregar un tipo nuevo no requiere tocar el modelo de almacenamiento.

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

## Dominio

El paquete `secrets/` es el núcleo del dominio — no importa nada de SQLite, GORM, ni detalles de cifrado. Solo define contratos:

- **`Secret`** — metadatos (nombre, tipo, fechas) + `Encrypted` opaco.
- **`Payload`** — interfaz que implementa cada tipo de secreto (`Display()`, `Validate()`).
- **`Encryptor`** — interfaz de cifrado; la implementación concreta (`aesgcm`) vive en `shared/`.
- **`Repository`** — interfaz de persistencia; la implementación concreta vive en `infra/`.

---

## Estructura

```
source/
├── secrets/              # dominio puro
│   ├── account/          # agregado: Account + sentinels
│   ├── finances/         # agregado: Finance (Card + Mobile) + sentinels
│   ├── domain.go         # Secret, Payload
│   ├── encryptor.go      # Encrypted, Encryptor, errores de cifrado
│   ├── repository.go     # Repository
│   └── types.go          # constantes de tipo e IsKnownType
└── shared/
    ├── crypto/aesgcm/    # implementación concreta: AES-256-GCM + Argon2id
    └── memory/           # Cleaner — limpieza segura de slices sensibles

context/
├── concepts.md           # criptografía y decisiones de diseño (leer primero)
└── modeling.md           # modelado del dominio en código
```

---

## Requisitos

- Go 1.24+
- `golang.org/x/crypto` (Argon2id)
