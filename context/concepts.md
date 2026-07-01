# Vext — Conceptos Fundamentales

## ¿Para qué sirve este documento?

Vext se apoya en un puñado de ideas de criptografía y diseño de software que, si no se entienden, hacen que el resto del proyecto parezca magia (o peor, arbitrario). Este documento explica esas ideas desde cero: qué problema resuelven, por qué se eligió esa solución y no otra, y cómo encajan entre sí.

Es teoría, pero con ejemplos concretos. Los bloques de código que vas a ver son **ilustrativos** — muestran la idea, no necesariamente la implementación final de Vext (que puede cambiar en el refactor). Si en algún momento el código real diverge de estos ejemplos, gana el código real.

---

## Primero lo primero: ¿Simétrico o Asimétrico?

Antes de hablar de Argon2id o AES, vale la pena resolver esta pregunta, porque es la base de todo lo demás.

Hay dos grandes familias de cifrado:

### Cifrado Simétrico

Una sola clave. Esa misma clave se usa tanto para cifrar como para descifrar. Es rápido, simple conceptualmente, y perfecto cuando **la misma persona (o sistema) que cifra es la que después va a descifrar**.

```
   Texto plano ──[ Clave K ]──► Ciphertext ──[ Clave K ]──► Texto plano
   (cifrar)                                   (descifrar)
```

El problema clásico del cifrado simétrico aparece cuando dos partes distintas necesitan comunicarse: ¿cómo comparten la clave sin que alguien la intercepte en el camino? Ese problema se llama "distribución de claves", y es la razón por la que existe el cifrado asimétrico.

### Cifrado Asimétrico

Un par de claves matemáticamente relacionadas: una **pública** (se puede compartir libremente) y una **privada** (se guarda en secreto). Lo que se cifra con la pública solo se puede descifrar con la privada correspondiente.

```
   Texto plano ──[ Clave Pública ]──► Ciphertext ──[ Clave Privada ]──► Texto plano
   (cualquiera puede cifrar)                        (solo el dueño puede descifrar)
```

Esto resuelve el problema de distribución de claves — no hace falta compartir ningún secreto de antemano — pero tiene un costo: es órdenes de magnitud más lento computacionalmente que el cifrado simétrico. Por eso, en la práctica, el cifrado asimétrico casi nunca se usa para cifrar datos grandes directamente; se usa para intercambiar de forma segura una clave simétrica, y de ahí en adelante se usa esa clave simétrica para todo el trabajo pesado (así funciona, por ejemplo, HTTPS).

### ¿Y Vext?

Vext usa **cifrado simétrico**, exclusivamente. Y tiene sentido: no hay dos partes distintas que necesiten comunicarse. Hay una sola persona, cifrando datos en su propia máquina, para leerlos ella misma después. No existe el problema de "¿cómo le mando la clave a la otra parte?" — porque no hay otra parte. La contraseña maestra cumple el rol de esa única clave (aunque, como vas a ver, no se usa directamente — primero pasa por un proceso de transformación).

Usar cifrado asimétrico acá sería resolver un problema que Vext no tiene, pagando un costo de performance que no aporta ningún beneficio real.

---

## Etapa 1: Derivación de Clave — de contraseña humana a clave criptográfica

### El problema

Una contraseña maestra como `MiGato2019!` es una cadena pensada para que un humano la recuerde. No es una clave criptográfica. AES-256, el algoritmo simétrico que usa Vext, espera una clave de exactamente 256 bits de alta entropía (es decir: tan impredecible como sea posible, sin patrones).

Si tomáramos la contraseña tal cual y la metiéramos directo en el cifrador, tendríamos dos problemas:

1. La mayoría de las contraseñas humanas tienen mucha menos entropía real de la que aparentan.
2. Nada haría lento el proceso de "probar contraseñas" — un atacante podría intentar millones por segundo.

### La solución: una KDF

Una **Función de Derivación de Clave** (KDF) transforma una contraseña arbitraria en una clave de longitud fija y alta entropía. Es determinística — la misma contraseña siempre produce la misma clave — pero deliberadamente costosa de calcular.

```
"MiGato2019!"  ──►  [ Argon2id ]  ──►  a3f9...e21c  (256 bits, indistinguible de aleatorio)
                          ▲
                       + Salt
```

### ¿Por qué Argon2id y no otra cosa?

KDFs más viejas como PBKDF2 solo hacen costoso el *tiempo* de cómputo. Eso no alcanza cuando el atacante tiene GPUs o hardware especializado, que paralelizan ese cómputo brutalmente — miles de núcleos calculando en simultáneo.

Argon2id ataca esto siendo costoso en **dos dimensiones**: tiempo *y* memoria. Cada intento de "adivinar" la contraseña requiere no solo ciclos de CPU, sino una cantidad significativa de RAM dedicada. Eso rompe la ventaja económica de las GPUs, que tienen muchísimos núcleos pero poca memoria por núcleo — no pueden simplemente "correr en paralelo" miles de intentos si cada uno necesita 64 MB propios.

Argon2id ganó la Password Hashing Competition en 2015 y hoy es la recomendación estándar de la industria. El nombre viene de combinar dos variantes:

- **Argon2i** — optimizada contra ataques de canal lateral (inferir la contraseña observando patrones de acceso a memoria).
- **Argon2d** — optimizada contra fuerza bruta con hardware paralelo.
- **Argon2id** — un híbrido de ambas. La opción más segura para un caso de uso general.

Con una configuración razonable, romper una contraseña maestra bien elegida deja de ser un problema de horas — pasa a ser un problema de siglos.

**Ejemplo ilustrativo de configuración:**

```go
// Ejemplo ilustrativo — los valores reales pueden cambiar
type Argon2Params struct {
    Time    uint32 // cuántas "pasadas" hace el algoritmo (más = más lento)
    Memory  uint32 // KB de RAM requeridos por intento (más = más caro paralelizar)
    Threads uint8  // paralelismo interno permitido
    KeyLen  uint32 // largo de la clave resultante, en bytes (32 = 256 bits)
}

func DeriveKey(password, salt []byte, p Argon2Params) []byte {
    return argon2.IDKey(password, salt, p.Time, p.Memory, p.Threads, p.KeyLen)
}
```

### El Salt: por qué la misma contraseña no siempre da la misma clave

Cada vez que se deriva una clave, se usa junto a un **salt**: un valor aleatorio generado para esa operación puntual. El salt no es secreto — se puede guardar en texto plano sin ningún problema — pero cumple un rol clave:

- La misma contraseña maestra + salts distintos = claves derivadas completamente distintas.
- Esto significa que si guardás dos secretos protegidos por la misma contraseña maestra, terminan cifrados con claves diferentes entre sí.
- Y, sobre todo, inutiliza las *rainbow tables* — tablas precomputadas de "contraseña → clave" para contraseñas comunes. Sin saber el salt de antemano, esas tablas no sirven para nada.

---

## Etapa 2: Cifrado Autenticado — no alcanza con solo "esconder" los datos

### Confidencialidad vs. Integridad

Hay dos propiedades distintas que la gente suele confundir:

| Propiedad | Pregunta que responde |
|---|---|
| **Confidencialidad** | ¿Alguien sin la clave puede leer el contenido? |
| **Integridad / Autenticidad** | ¿Alguien puede modificar el contenido cifrado sin que se note? |

El cifrado simétrico "clásico" (como AES en modo CBC) solo da confidencialidad. Si un atacante altera bytes del ciphertext, muchas veces el descifrado simplemente produce basura silenciosa — sin ninguna alarma. En ciertos escenarios eso es explotable: un atacante puede aprender cosas sobre el contenido, o corromper datos de forma controlada, sin nunca haber tenido la clave.

### AEAD: cifrar y verificar en un solo paso

Vext usa **AES-256-GCM**, que pertenece a la familia AEAD (*Authenticated Encryption with Associated Data*). La idea: además de cifrar, el algoritmo genera un **tag de autenticación** — una huella digital corta del ciphertext.

```
                     ┌─────────────────────────┐
Texto plano ────────►│                         │───► Ciphertext
                     │      AES-256-GCM        │
Clave (256 bits) ───►│      (cifrado)          │───► Tag de autenticación
Nonce ──────────────►│                         │
                     └─────────────────────────┘
```

Al descifrar, el tag se recalcula y se compara contra el guardado:

```
                     ┌─────────────────────────┐
Ciphertext ─────────►│                         │
Tag guardado ───────►│      AES-256-GCM        │───► ¿Tag coincide?
Clave ──────────────►│      (descifrado)       │       │
Nonce ──────────────►│                         │       ├── Sí → Texto plano
                     └─────────────────────────┘       └── No → Error, nada se entrega
```

Si un solo byte del ciphertext (o del tag) fue alterado — por un atacante o por corrupción accidental del disco — la verificación falla y el descifrado se **rechaza por completo**. No existe un "descifrado parcial" ni un resultado ambiguo.

**Ejemplo ilustrativo:**

```go
// Ejemplo ilustrativo — cifrar
block, _ := aes.NewCipher(key)          // key: 32 bytes (256 bits)
gcm, _ := cipher.NewGCM(block)
ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
// ciphertext ya incluye el tag de autenticación al final

// Ejemplo ilustrativo — descifrar
plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
if err != nil {
    // el tag no coincidió: clave incorrecta o datos manipulados
    return ErrDecryptionFailed
}
```

### El Nonce: usar la clave una sola vez, cada vez

GCM depende de un **nonce** ("number used once"): un valor que debe ser único para cada operación de cifrado hecha con una clave dada.

La regla no tiene excepciones: **un nonce nunca se repite con la misma clave**. Si eso pasa, las garantías de seguridad de GCM se rompen de forma catastrófica — un atacante que observa dos ciphertexts cifrados con el mismo par clave/nonce puede, en el peor caso, recuperar información sobre el contenido sin conocer la clave.

La mitigación es simple: generar un nonce nuevo y aleatorio en cada cifrado. Con un espacio de valores suficientemente grande (12 bytes = 2⁹⁶ posibilidades), la probabilidad de una colisión accidental es despreciable.

Igual que el salt, el nonce no es secreto — se puede guardar junto al dato cifrado sin problema.

---

## Las Dos Etapas, juntas

```
Contraseña maestra ──► [ Argon2id + Salt ] ──► Clave (256 bits)
                                                      │
Secreto en texto plano ──► [ AES-256-GCM + Nonce + Clave ] ──► Ciphertext + Tag
```

Son dos algoritmos con responsabilidades distintas — uno convierte una contraseña humana en una clave robusta; el otro usa esa clave para proteger datos de forma verificable — encadenados en serie.

### ¿Qué pasa con una contraseña incorrecta?

Un detalle de diseño interesante: si alguien escribe la contraseña maestra incorrecta, Argon2id **igual deriva una clave** — simplemente es la clave equivocada, porque partió de una contraseña distinta. Esa clave incorrecta hace que la verificación del tag en GCM falle, y el sistema devuelve un error genérico.

```go
// Ejemplo ilustrativo
if err != nil {
    return errors.New("wrong master password or data corrupted")
    // deliberadamente ambiguo: no distingue entre ambos casos
}
```

Esto es intencional. El error no le dice al usuario si la contraseña estuvo "cerca" o muy lejos. Dar más detalle que eso le regalaría información útil a un atacante intentando adivinar por aproximación.

---

## Limpieza de Memoria: proteger los datos también *después* de usarlos

Cifrar no termina el trabajo. Mientras el proceso ocurre, hay valores extremadamente sensibles flotando en RAM: la contraseña maestra, la clave derivada, el secreto en texto plano.

El problema es que un lenguaje con garbage collector, como Go, no garantiza *cuándo* esa memoria se libera ni si sigue siendo legible mientras tanto. Si alguien lograra volcar la memoria del proceso en el momento equivocado, esos valores podrían seguir ahí, sin cifrar.

La mitigación: apenas un valor sensible deja de ser necesario, se sobreescribe explícitamente con ceros — sin esperar al garbage collector.

```go
// Ejemplo ilustrativo
func Cleaner(b []byte) {
    for i := range b {
        b[i] = 0
    }
}

func DeriveAndUse(password, salt []byte) {
    key := DeriveKey(password, salt)
    defer Cleaner(key) // se limpia sin importar cómo termine la función
    // ... usar key ...
}
```

No es una protección perfecta — Go no da control total sobre la memoria — pero reduce mucho la ventana de tiempo en la que un secreto vive expuesto.

---

## Secretos Polimórficos: un solo modelo, muchas formas

Vext necesita guardar tipos de secretos muy distintos: una cuenta online tiene usuario y contraseña; una tarjeta bancaria tiene número, PIN, código de seguridad, vencimiento, y más. Con el tiempo van a aparecer más tipos — notas seguras, claves SSH, lo que sea.

La pregunta de diseño: ¿cómo modelás datos con formas tan distintas sin que cada tipo nuevo obligue a rediseñar todo?

La respuesta es **polimorfismo de datos**: todos los secretos comparten el mismo contenedor externo (un nombre, una etiqueta de tipo, y un bloque cifrado), y es solo *adentro* de ese bloque donde la forma cambia según el tipo.

Así se ve en el código real de Vext. Primero, el catálogo de tipos conocidos:

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

Y después, un struct de payload por cada tipo — cada uno con exactamente los campos que ese tipo necesita, ni uno más:

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

Desde afuera del sistema, no importa si estás manejando una cuenta o una tarjeta — solo manejás "secretos". La responsabilidad de saber qué campos corresponden a cada tipo vive en un solo lugar: el struct de payload correspondiente, seleccionado según el valor de `Type`. Agregar un tipo nuevo es agregar un nuevo struct y una nueva constante — no inventar un sistema paralelo.

Esta idea es la que le permite a Vext prometer algo fuerte: el modelo de almacenamiento no cambia aunque el catálogo de tipos de secretos siga creciendo.

> Si querés el detalle completo de cómo está modelado el dominio de secretos — la separación entre metadatos y contenido cifrado, por qué algunos campos son `[]byte` y otros `string`, y cómo se valida un tipo — eso está en `modeling.md`.