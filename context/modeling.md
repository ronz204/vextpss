# Vext — Modelado de Dominio

## De qué habla este documento

`concepts.md` explica la *idea* de por qué Vext usa un modelo polimórfico para los secretos. Este documento va un nivel más abajo: cómo está modelado ese dominio en código real, qué decisiones de diseño hay detrás de cada struct, y qué reglas seguir cuando se agrega un tipo de secreto nuevo.

Todo lo que se describe acá vive en el paquete `secrets/` — el dominio puro de Vext. Este paquete no sabe nada de SQLite, de GORM, ni de cómo se cifra un byte. Solo conoce structs, tipos, y las reglas que los gobiernan. Esa separación es intencional: el dominio no debería tener que cambiar si mañana Vext decide usar otro motor de base de datos o otro algoritmo de cifrado.

---

## Dos preguntas, dos structs: `Secret` y `Credential`

Un secreto guardado en Vext tiene dos clases de información con necesidades muy distintas:

- **Metadatos**: nombre, tipo, cuándo se creó, cuándo se modificó. Esto es información que tiene sentido mostrar tal cual — por ejemplo, en la salida de `vext list`. No es sensible por sí sola.
- **Contenido cifrado**: el blob de bytes que representa el secreto real. Esto es lo que hay que proteger, y no tiene sentido cargarlo en memoria si lo único que se necesita es listar nombres.

Vext modela esa distinción con dos structs separados en vez de uno solo con todo mezclado:

```go
type Secret struct {
    ID        int64
    Name      string
    Type      string
    Salt      []byte
    Nonce     []byte
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Credential struct {
    Secret    Secret
    Encrypted []byte
}
```

`Secret` es el conjunto de metadatos — incluye `Salt` y `Nonce` porque, aunque están directamente relacionados con el cifrado, **no son secretos en sí mismos** (como se explica en `concepts.md`, ambos pueden guardarse en texto plano sin comprometer la seguridad). `Credential` envuelve a `Secret` y le agrega lo único que realmente hace falta proteger: el blob `Encrypted`.

### ¿Por qué separarlos en vez de un solo struct?

Porque las operaciones que necesita Vext tienen necesidades distintas:

- `vext list` solo necesita nombres, tipos y fechas — nunca necesita tocar `Encrypted`. Si `Secret` y `Credential` fueran el mismo struct, cada `list` estaría cargando (y potencialmente exponiendo en memoria) contenido cifrado que ni siquiera va a usar.
- `vext get`, en cambio, sí necesita el contenido — ahí es donde entra `Credential`, que agrega el blob sobre la base de `Secret`.

Modelar esto como dos structs distintos hace que la intención de cada operación sea explícita en su firma: una función que devuelve `[]Secret` está declarando, por su tipo de retorno, que no toca contenido cifrado. Una función que devuelve `Credential` está declarando lo contrario. El compilador ayuda a que ese contrato se respete.

---

## El Payload: la parte que cambia según el tipo

`Credential.Encrypted` es un blob de bytes — desde la perspectiva de `Secret`/`Credential`, es opaco. Pero antes de cifrarse, ese blob fue un objeto Go concreto: un *payload*, con una forma específica según el tipo de secreto.

```go
type AccountSecret struct {
    Username string `json:"username"`
    Password []byte `json:"password"`
}

type FinanceSecret struct {
    CardPin         []byte `json:"card_pin"`
    CardNumber      string `json:"card_number"`
    SecurityCode    []byte `json:"security_code"`
    ExpirationMonth int    `json:"expiration_month"`
    ExpirationYear  int    `json:"expiration_year"`
    BankUsername    string `json:"bank_username"`
    BankPassword    []byte `json:"bank_password"`
    BankVirtualKey  []byte `json:"bank_virtual_key"`
    BankCellphone   string `json:"bank_cellphone"`
}
```

Cada payload es un struct Go normal, serializable a JSON. El flujo conceptual es: payload → serializar a JSON → cifrar el JSON completo como un blob único → ese blob es lo que termina en `Credential.Encrypted`. Ni `Secret` ni `Credential` necesitan saber que `AccountSecret` o `FinanceSecret` existen — esa capa de conocimiento vive un nivel más arriba, en el código que decide "si el tipo es `account`, deserializar como `AccountSecret`".

### La convención `[]byte` vs `string`

Fijate que no todos los campos usan el mismo tipo para texto. `Username`, `CardNumber`, `BankUsername`, `BankCellphone` son `string`. Pero `Password`, `CardPin`, `SecurityCode`, `BankPassword`, `BankVirtualKey` son `[]byte`.

Esto no es arbitrario — es una convención deliberada:

- Los campos `[]byte` son los que se consideran **material sensible que debe limpiarse de memoria explícitamente** después de usarse (con el patrón `memory.Cleaner` descrito en `concepts.md`). Un `string` en Go es inmutable — una vez creado, no hay forma de sobreescribir su contenido en memoria sin recurrir a trucos de bajo nivel. Un `[]byte`, en cambio, se puede poner en cero byte por byte.
- Los campos `string` son datos que, si bien pueden ser privados, no reciben ese tratamiento de limpieza activa — típicamente porque son identificadores (un username, un número de tarjeta) más que secretos que por sí solos otorgan acceso a algo.

En otras palabras: el tipo del campo no es solo una decisión de serialización — es una señal de intención sobre cómo ese dato debe tratarse en memoria durante todo su ciclo de vida.

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

El campo `Type` de `Secret` es un `string` — no un tipo enumerado nativo de Go (Go no tiene enums reales). Eso significa que, en teoría, nada impide que alguien intente crear un secreto con `Type: "cuenta"` en vez de `"account"`, o con un typo. `IsKnownType` es la barrera contra eso: un único punto de verdad que responde "¿este string es un tipo que Vext realmente entiende?".

Este patrón — constantes + mapa de validación + función de consulta — es intencionalmente simple. No hace falta reflection, ni un sistema de registro dinámico de tipos, ni una interfaz `SecretType`. Cuando se agrega un tipo nuevo, los pasos son mecánicos y locales a este archivo:

1. Agregar la constante (`TypeSSHKey = "ssh_key"`, por ejemplo).
2. Agregarla a `knownTypes`.
3. Definir el struct de payload correspondiente (`SSHKeySecret`).

Nada de esto obliga a tocar `Secret`, `Credential`, ni ninguna capa de persistencia — que es exactamente la promesa central del modelo polimórfico.

---

## Resumen del flujo completo

Uniendo todo lo anterior, así se ve el camino completo de un secreto dentro del dominio:

```
Usuario elige tipo ──► IsKnownType valida ──► se arma el payload correspondiente
     "account"              (constante conocida)      (AccountSecret{...})
                                                              │
                                                     serializar a JSON
                                                              │
                                                   cifrar (ver encryption.md)
                                                              │
                                                              ▼
                                              Credential{ Secret{...}, Encrypted }
```

`Secret` y `Credential` son el contrato estable que el resto de Vext conoce. Los payloads (`AccountSecret`, `FinanceSecret`, y los que vengan) son el detalle variable que vive completamente encapsulado detrás de ese contrato.