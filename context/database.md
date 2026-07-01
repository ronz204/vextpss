# Vext — Persistencia

## De qué habla este documento

`modeling.md` describe el dominio de secretos (`Secret`, `Credential`, los payloads por tipo) tal como vive en el paquete `secrets/` — sin ninguna noción de bases de datos. Este documento cubre la capa que sí sabe de eso: cómo esos structs terminan guardados en disco.

La regla de dependencia es unidireccional: la capa de persistencia conoce el dominio (necesita construir un `Secret` para devolverlo), pero el dominio **no conoce la persistencia**. `secrets/` no importa nada relacionado a SQLite ni a GORM. Esto es lo que permite, en teoría, cambiar el motor de almacenamiento el día de mañana sin tocar una sola línea del modelo de dominio.

---

## Motor elegido: SQLite vía GORM

Vext guarda todo en un único archivo SQLite. No hay servidor que levantar, no hay configuración de conexión, no hay credenciales de base de datos — el "motor" es, literalmente, un archivo en el filesystem del usuario. Para una herramienta local-first como Vext, es la elección obvia: cualquier alternativa cliente-servidor introduciría una dependencia externa que va en contra de la filosofía del proyecto.

Sobre SQLite se usa GORM como ORM. Su trabajo es abstraer el SQL, mapear structs Go a filas de tabla, y manejar migraciones de forma declarativa con `AutoMigrate` en vez de escribir y versionar archivos `.sql` a mano.

**Ubicación del archivo**, típicamente bajo el directorio de configuración del usuario:

```
~/.config/vext/vext.db
```

---

## El Struct de Persistencia: un mapeo, no el dominio mismo

Acá aparece una distinción importante que ya se mencionó arriba: el struct que GORM usa para mapear filas de la tabla **no es** `secrets.Secret`. Es un struct propio de la capa de storage, con sus propios tags de GORM, que se construye *a partir* de la información de `Secret` y `Credential` — pero vive en un paquete distinto.

```go
// Ejemplo ilustrativo de cómo podría verse — el struct final
// puede tener ajustes menores respecto a esto.
type SecretRecord struct {
    ID        int64     `gorm:"primaryKey;autoIncrement"`
    Name      string    `gorm:"uniqueIndex;not null"`
    Type      string    `gorm:"not null"`
    Salt      []byte    `gorm:"not null;type:blob"`
    Nonce     []byte    `gorm:"not null;type:blob"`
    Encrypted []byte    `gorm:"not null;type:blob"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (SecretRecord) TableName() string { return "secrets" }
```

Fijate el paralelismo con `Secret` + `Credential` de `modeling.md`: `SecretRecord` es esencialmente esos dos structs aplanados en uno solo, más las anotaciones de GORM que le dicen a SQLite cómo tratar cada columna. Esto no es duplicación por descuido — es la traducción explícita entre "cómo piensa el dominio" (dos structs separados por responsabilidad) y "cómo se guarda en una sola fila de tabla" (todo junto, porque eso es lo que una fila es).

Mantener estos structs separados — uno de dominio, uno de persistencia — significa que agregar un tag de GORM, cambiar un índice, o eventualmente migrar de SQLite a otra cosa, nunca obliga a tocar `secrets.Secret` ni ningún código que dependa de él.

El SQL equivalente al struct de arriba:

```sql
CREATE TABLE IF NOT EXISTS secrets (
    id         INTEGER  PRIMARY KEY AUTOINCREMENT,
    name       TEXT     UNIQUE NOT NULL,
    type       TEXT     NOT NULL,
    salt       BLOB     NOT NULL,
    nonce      BLOB     NOT NULL,
    encrypted  BLOB     NOT NULL,
    created_at DATETIME,
    updated_at DATETIME
);
```

### Por qué una sola tabla, no una por tipo

Con el modelo polimórfico ya explicado en `concepts.md` y `modeling.md`, la consecuencia natural del lado de persistencia es: **una sola tabla para todos los tipos de secreto**, sin importar si es `account`, `finance`, o cualquier tipo futuro.

La alternativa — una tabla `accounts`, otra `finance_cards`, etc. — se descartó porque rompería justamente la promesa central del modelo polimórfico: cada tipo nuevo exigiría una migración de schema, y una consulta como `vext list` (que necesita listar *todos* los secretos sin importar su tipo) se volvería una unión de N tablas en vez de una consulta simple.

Con una sola tabla, el schema SQL **nunca cambia** cuando se agrega un tipo de secreto nuevo. Toda la variación vive en el campo `Type` (texto libre, validado por `IsKnownType` del lado del dominio) y en el contenido — opaco para SQLite — de `Encrypted`.

---

## Qué guarda cada campo, y qué no

| Campo | ¿Qué es? | ¿Es sensible? |
|---|---|---|
| `id` | Clave primaria autoincremental, uso interno | No |
| `name` | Identificador elegido por el usuario (`"github"`, `"visa-debito"`) | No, pero debe ser único |
| `type` | Etiqueta de tipo (`"account"`, `"finance"`) | No |
| `salt` | Usado por Argon2id para derivar la clave de cifrado | No — puede estar en texto plano |
| `nonce` | Usado por AES-GCM como vector de inicialización | No — puede estar en texto plano |
| `encrypted` | El JSON del payload, cifrado (ciphertext + tag de GCM) | **Sí — esto es lo único que realmente hay que proteger** |
| `created_at` / `updated_at` | Timestamps, gestionados por GORM | No |

Vale la pena remarcar algo que ya se explicó en `encryption.md`/`concepts.md`: que `salt` y `nonce` vivan sin cifrar en la misma fila que el secreto **no es una debilidad** — es parte del diseño. Ninguno de los dos otorga acceso a nada por sí solo; son insumos públicos del proceso criptográfico, no secretos.

---

## El Repositorio: la frontera entre dominio y storage

La capa de storage no expone GORM ni SQL hacia arriba. Expone una interfaz en términos del dominio — recibe y devuelve `Secret` / `Credential`, nunca `SecretRecord`:

```go
type Repository interface {
    Create(ctx context.Context, secret *Secret, encrypted []byte) error
    GetByName(ctx context.Context, name string) (*Secret, []byte, error)
    Update(ctx context.Context, secret *Secret, encrypted []byte) error
    Delete(ctx context.Context, name string) error
    ListAll(ctx context.Context) ([]Secret, error)
    GetAll(ctx context.Context) ([]Credential, error)
}
```

Esta interfaz es la traducción práctica de la separación `Secret`/`Credential` que vimos en `modeling.md`:

- **`ListAll`** devuelve `[]Secret` — solo metadatos, sin tocar la columna `encrypted`. Es lo que usa `vext list`: rápido, y sin necesidad de mover contenido cifrado por memoria si no hace falta.
- **`GetAll`** devuelve `[]Credential` — metadatos *más* el blob cifrado. Pensado para operaciones que sí necesitan el contenido completo, como una futura exportación de backup.
- **`GetByName`** devuelve el secreto puntual con su blob, para el flujo de `vext get`.

El resto del sistema (los casos de uso, los comandos) programa contra esta interfaz, no contra GORM directamente. Esto es lo que hace posible, en teoría, reemplazar SQLite por otro motor sin que nada fuera de la capa de storage se entere.

---

## Migraciones

GORM ofrece `AutoMigrate`, que crea tablas faltantes y agrega columnas nuevas de forma idempotente — se puede correr múltiples veces sin efectos destructivos. No elimina columnas ni modifica las existentes, así que agregar un campo nuevo al struct de persistencia es, en general, seguro de aplicar directamente en producción sin un sistema de migraciones versionado aparte.

Esto encaja bien con la filosofía de Vext: para una base de datos local de un solo usuario, un sistema de migraciones más sofisticado (versionado, rollback, etc.) sería complejidad que no se traduce en ningún beneficio real.

---

## Cosas a tener en cuenta

**Archivos WAL y journal.** SQLite puede crear archivos temporales (`-wal`, `-journal`) junto al archivo principal durante escrituras. No son un problema de seguridad: como Vext cifra *antes* de escribir, esos archivos temporales nunca van a contener texto plano — como mucho, contienen los mismos blobs cifrados a medio persistir.

**Permisos del directorio.** El directorio de configuración donde vive `vext.db` debería crearse con permisos restrictivos (por ejemplo, `0700` en sistemas Unix), de forma que otros usuarios de la misma máquina no puedan siquiera leer el archivo de base de datos.

**Errores de constraint único.** Cuando se intenta crear un secreto con un `name` que ya existe, SQLite rechaza el insert por la restricción `UNIQUE`. GORM no siempre envuelve esto en un error tipado y reconocible directamente — suele hacer falta inspeccionar el mensaje de error para mapearlo a un error de dominio propio (algo como `ErrAlreadyExists`) antes de devolverlo hacia arriba.