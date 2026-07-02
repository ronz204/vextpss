# Vext — Modelado de Dominio

## De qué habla este documento

`concepts.md` explica la *idea* de por qué Vext usa un modelo polimórfico para los secretos. Este documento va un nivel más abajo: cómo está modelado ese dominio en código real, qué decisiones de diseño hay detrás de cada struct, y qué reglas seguir cuando se agrega un tipo de secreto nuevo.

Todo lo que se describe acá vive en el paquete `secrets/` y sus subpaquetes. Este paquete no sabe nada de SQLite, de GORM, ni de cómo se cifra un byte. Solo conoce structs, tipos, y las reglas que los gobiernan. Esa separación es intencional: el dominio no debería tener que cambiar si mañana Vext decide usar otro motor de base de datos o otro algoritmo de cifrado.

---

## La unidad central: `Secret`

```go
type Secret struct {
    ID        int64
    Name      string
    Type      string
    Encrypted Encrypted
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

`Secret` reúne metadatos identificables (`Name`, `Type`, fechas) con el contenido cifrado en un único struct. El campo `Encrypted` es opaco para el dominio — `Secret` no inspecciona su interior, solo lo transporta y persiste.

---

## `Encrypted`: el sobre cifrado

```go
type Encrypted struct {
    Algorithm  string
    Ciphertext []byte
    Metadata   []byte
}
```

`Encrypted` no es un `[]byte` crudo — es un sobre con tres partes:

- **`Algorithm`**: identifica qué implementación de `Encryptor` produjo este payload y cuál debe consumirlo (ej. `"aes-gcm-argon2id"`). El dominio nunca lo interpreta, solo lo transporta para que la capa de cifrado sepa qué hacer.
- **`Ciphertext`**: el blob cifrado.
- **`Metadata`**: bytes opacos que el `Encryptor` sabe interpretar — hoy codifican salt y nonce, pero el dominio nunca los inspecciona directamente.

Este diseño desacopla el dominio de los detalles criptográficos: no hay `Salt []byte` ni `Nonce []byte` sueltos en `Secret`. Esos detalles pertenecen a la capa de cifrado, y `Metadata` es el mecanismo por el que viajan sin que el dominio necesite entenderlos.

---

## El contrato de los payloads: la interfaz `Payload`

```go
type Payload interface {
    Display() string
    Validate() error
}
```

Antes de cifrarse, el contenido de un secreto es un objeto Go concreto que implementa `Payload`. Esta interfaz define el mínimo que cualquier tipo de secreto debe cumplir:

- `Display()` devuelve el identificador de tipo (`"account"`, `"finance"`).
- `Validate()` verifica que el objeto tiene los datos mínimos necesarios.

Los payloads concretos viven en subpaquetes propios — no en el paquete raíz `secrets/`.

---

## Subpaquetes de payload

### `secrets/account`

```go
type Account struct {
    Username string `json:"username"`
    Password []byte `json:"password"`
}
```

Un payload simple: credencial de acceso con usuario y contraseña. `Account` implementa `secrets.Payload`.

### `secrets/finances`

```go
type Finance struct {
    Card   Card   `json:"card"`
    Mobile Mobile `json:"mobile"`
}

type Card struct {
    Pin             []byte `json:"pin"`
    Number          string `json:"number"`
    SecurityCode    []byte `json:"security_code"`
    ExpirationMonth int    `json:"expiration_month"`
    ExpirationYear  int    `json:"expiration_year"`
}

type Mobile struct {
    Username   string `json:"username"`
    Password   []byte `json:"password"`
    VirtualKey []byte `json:"virtual_key"`
    Cellphone  string `json:"cellphone"`
}
```

`Finance` es un agregado compuesto por dos value objects: `Card` (datos de tarjeta) y `Mobile` (acceso a banca móvil). `Finance` implementa `secrets.Payload`; su `Validate()` delega a `Card.Validate()` y `Mobile.Validate()` en secuencia.

### Sentinels de validación

Cada subpaquete tiene su propio archivo `sentinels.go` con errores exportados para cada regla de validación:

```go
// account/sentinels.go
var (
    ErrUsernameRequired = errors.New("username is required")
    ErrPasswordRequired = errors.New("password is required")
)

// finances/sentinels.go
var (
    ErrCardNumberRequired         = errors.New("card number is required")
    ErrCardPinRequired            = errors.New("card pin is required")
    // ...
)
```

Usar sentinels en vez de strings sueltos permite que los llamadores hagan `errors.Is` para distinguir errores concretos sin parsear mensajes.

---

## La convención `[]byte` vs `string`

No todos los campos de texto usan el mismo tipo. `Username`, `Number`, `Cellphone` son `string`. `Password`, `Pin`, `SecurityCode`, `VirtualKey` son `[]byte`.

Esto es una convención deliberada:

- Los campos `[]byte` son **material sensible que debe limpiarse de memoria explícitamente** después de usarse (con `memory.Cleaner`). Un `string` en Go es inmutable — una vez creado, no hay forma de sobreescribir su contenido en memoria. Un `[]byte` se puede poner en cero byte por byte.
- Los campos `string` son datos que, si bien pueden ser privados, no reciben ese tratamiento — típicamente porque son identificadores (un username, un número de tarjeta) más que secretos que por sí solos otorgan acceso a algo.

El tipo del campo es una señal de intención sobre cómo ese dato debe tratarse en memoria durante todo su ciclo de vida.

---

## Validar tipos: el registro de tipos conocidos

```go
const (
    TypeAccount = "account"
    TypeFinance = "finance"
)

var knownTypes = map[string]bool{
    TypeAccount: true,
    TypeFinance: true,
}

func IsKnownType(t string) bool {
    return knownTypes[t]
}
```

El campo `Type` de `Secret` es un `string` — Go no tiene enums reales. `IsKnownType` es la barrera contra strings inválidos: un único punto de verdad que responde "¿este tipo existe en Vext?".

Las constantes `TypeAccount` y `TypeFinance` son las que los subpaquetes usan en sus implementaciones de `Display()`, cerrando el ciclo: el string que identifica un tipo en `Secret.Type` es el mismo que devuelve `Payload.Display()`.

Cuando se agrega un tipo nuevo los pasos son mecánicos y locales:

1. Agregar la constante (`TypeSSHKey = "ssh_key"`).
2. Agregarla a `knownTypes`.
3. Crear el subpaquete con el aggregate, composite objects y sentinels.

---

## Resumen del flujo completo

```
Usuario elige tipo ──► IsKnownType valida ──► se construye el payload concreto
     "account"              (constante conocida)      (account.Account{...})
                                                              │
                                                     Validate() pasa
                                                              │
                                                   serializar a JSON
                                                              │
                                                   cifrar (Encryptor)
                                                              │
                                                              ▼
                                              Secret{ Encrypted{ Algorithm, Ciphertext, Metadata } }
```

`Secret` es el contrato estable que el resto de Vext conoce. Los payloads (`account.Account`, `finances.Finance`, y los que vengan) son el detalle variable completamente encapsulado detrás de ese contrato. `Encrypted` es el puente opaco entre el dominio y la capa criptográfica.
