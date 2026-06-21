# Vext — Cifrado y Modelo de Seguridad

## Principio Guía

Vext está diseñado para que **incluso si alguien roba tu archivo de base de datos, no obtenga nada útil**. Los blobs cifrados en SQLite no tienen sentido sin la contraseña maestra, y forzarla por fuerza bruta es computacionalmente inviable por diseño.

---

## El Modelo de Dos Etapas

Cada secreto guardado en Vext pasa por dos operaciones criptográficas distintas: **derivación de clave** y **cifrado autenticado**. Son responsabilidades separadas resueltas por algoritmos separados.

```
Contraseña maestra  ──►  [ Argon2id KDF ]  ──►  Clave de 32 bytes
                                ▲
                          Salt (aleatorio, guardado en BD)

Secreto en texto plano  ──►  [ AES-256-GCM ]  ──►  Blob cifrado
                                    ▲
                         Clave (de arriba) + Nonce (aleatorio, guardado en BD)
```

Ambas etapas están implementadas en `shared/cryptors/aes_gcm.go` a través de `AESGCMEncryptor`.

---

## Etapa 1: Derivación de Clave — Argon2id

### ¿Por qué no usar la contraseña directamente?

Una contraseña maestra como `MiGato2019!` es una cadena amigable para humanos con mucho menos entropía de lo que los algoritmos de cifrado esperan. Pasarla directamente a AES sería inseguro.

Una Función de Derivación de Clave (KDF) resuelve esto transformando cualquier contraseña en una clave de longitud exacta y alta entropía.

### ¿Por qué Argon2id?

Argon2id ganó el Password Hashing Competition en 2015 y es la recomendación actual de la industria para hashing de contraseñas y KDFs. Su ventaja principal es que es deliberadamente **costoso tanto en tiempo como en memoria**, lo que ataca directamente la economía de los intentos de fuerza bruta.

- Una granja de GPUs que podría romper un hash simple en horas tardaría **siglos** con una configuración bien ajustada de Argon2id.
- La variante `id` es un híbrido de Argon2i (resistente a ataques de canal lateral) y Argon2d (resistente a GPUs), lo que la hace la elección más segura para propósito general.

### Configuración (`shared/cryptors/aes_conf.go`)

```go
func DefaultConfig() AESGCMConfig {
    return AESGCMConfig{
        Argon: Argon2Config{
            Time:    3,          // 3 pasadas
            Memory:  64 * 1024,  // 64 MB de RAM requeridos por intento
            Threads: 2,
            KeyLen:  32,         // Clave de salida de 256 bits
        },
        SaltLen:  16, // 16 bytes
        NonceLen: 12, // 12 bytes (estándar para GCM)
    }
}
```

### El Salt

Cada registro en la base de datos tiene su propio Salt de 16 bytes generado aleatoriamente. El Salt **no es secreto** — se guarda en texto plano en la BD. Su función es:

1. La misma contraseña maestra + diferente salt = clave derivada completamente diferente.
2. Dos registros cifrados con la misma contraseña maestra producen claves distintas.
3. Las tablas precomputadas de ataque (rainbow tables) son inútiles.

---

## Etapa 2: Cifrado Autenticado — AES-256-GCM

### ¿Por qué cifrado autenticado?

El cifrado estándar (como AES en modo CBC) solo provee **confidencialidad** — oculta el contenido. Pero no detecta si alguien manipuló los bytes cifrados.

AES-GCM (Galois/Counter Mode) provee **Cifrado Autenticado con Datos Asociados (AEAD)**:
- Cifra los datos (confidencialidad).
- Genera un tag de autenticación corto que actúa como huella digital del ciphertext.
- Al descifrar, si un solo byte fue alterado — por un atacante o por corrupción — la verificación del tag falla y el descifrado se rechaza por completo.

Esto hace que la base de datos sea **a prueba de manipulaciones**: cualquier modificación es detectada.

### El Nonce

Cada registro tiene su propio Nonce de 12 bytes generado aleatoriamente. Como el Salt, no es secreto y se guarda en la BD.

La regla crítica: **un Nonce nunca debe reutilizarse con la misma clave**. La reutilización del Nonce en GCM puede romper catastróficamente la confidencialidad. Al generar un Nonce aleatorio fresco por registro, este riesgo se elimina.

### ¿Qué se cifra?

El payload completo del secreto se serializa a JSON primero, luego el string JSON completo se cifra como un blob único. La capa de cifrado es agnóstica a qué tipo de datos está protegiendo — solo ve bytes.

---

## Implementación (`shared/cryptors/aes_gcm.go`)

### Cifrado

```go
func (e *AESGCMEncryptor) Encrypt(_ context.Context, plaintext, password []byte) (salt, nonce, ciphertext []byte, err error) {
    // 1. Generar salt aleatorio de 16 bytes
    salt, err = e.randomBytes(e.config.SaltLen)

    // 2. Derivar clave de 256 bits con Argon2id
    key := e.deriveKey(password, salt)
    defer memory.Cleaner(key) // Limpiar clave después de usar

    // 3. Generar nonce aleatorio de 12 bytes
    nonce, err = e.randomBytes(e.config.NonceLen)

    // 4. Crear cipher AES-256-GCM y cifrar
    gcm, err := e.newGCM(key)
    ciphertext = gcm.Seal(nil, nonce, plaintext, nil)

    return salt, nonce, ciphertext, nil
}
```

### Descifrado

```go
func (e *AESGCMEncryptor) Decrypt(_ context.Context, password, salt, nonce, ciphertext []byte) ([]byte, error) {
    // 1. Validar longitudes de salt y nonce (previene manipulación)
    if len(salt) != e.config.SaltLen || len(nonce) != e.config.NonceLen {
        return nil, secrets.ErrDecryptionFailed
    }

    // 2. Derivar la misma clave con el salt guardado
    key := e.deriveKey(password, salt)
    defer memory.Cleaner(key) // Limpiar clave después de usar

    // 3. Descifrar (GCM verifica el tag de autenticación automáticamente)
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, secrets.ErrDecryptionFailed // Error genérico intencional
    }

    return plaintext, nil
}
```

---

## Limpieza de Memoria

Las claves criptográficas y los secretos en texto plano existen en RAM solo durante la duración de una operación. Inmediatamente después de que el cifrado o descifrado termina, cada `[]byte` que contiene datos sensibles se sobrescribe con ceros llamando a `memory.Cleaner(b)` (`shared/memory/cleaner.go`).

El garbage collector de Go no garantiza cuándo la memoria es recuperada ni si puede ser leída por otro proceso en el ínterin. Limpiar manualmente es la única mitigación confiable.

El patrón usado en todo el código:

```go
key := e.deriveKey(password, salt)
defer memory.Cleaner(key) // Cero en el retorno de la función, sin importar el path de error
```

Esto aplica a: contraseñas maestras, claves derivadas, payloads en texto plano, y buffers temporales con datos sensibles.


## ¿Qué pasa con una Contraseña Maestra Incorrecta?

Cuando `vext get` se llama con la contraseña incorrecta:

1. Argon2id deriva una clave de 32 bytes **diferente** (porque se usó una contraseña diferente).
2. AES-GCM intenta descifrar con esta clave incorrecta.
3. La verificación del tag de autenticación falla inmediatamente.
4. Vext retorna `secrets.ErrDecryptionFailed`, que el adaptador mapea a: `"wrong master password or data corrupted"`.

El usuario no obtiene información sobre si la contraseña estuvo cerca o cómo se ven los datos reales. Esto es intencional.
