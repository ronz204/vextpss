# Vext — Base de Datos

## Motor elegido: SQLite vía GORM

Vext usa SQLite como motor de almacenamiento. Es un archivo único en disco, no requiere servidor, no necesita configuración y está probado en producción a escala mundial.

El driver utilizado es `github.com/glebarez/sqlite`, una implementación en Go puro (sin CGO). Esto evita depender de un compilador C en la máquina de build y produce un binario estático completamente portable. GORM (`gorm.io/gorm`) actúa como ORM: abstrae el SQL, mapea structs a filas y maneja migraciones con `AutoMigrate`.

**Ubicación del archivo:**
```
~/.config/vext/vext.db   (Linux/macOS/Windows — basado en os.UserHomeDir)
```

---

## Diseño del Schema: Tabla Polimórfica

Vext guarda distintos tipos de secretos: cuentas tienen `username` + `password`, tarjetas financieras tienen `card_number` + `cvv` + `expiration`, etc.

Dos enfoques fueron considerados:

**Opción A — Tablas separadas:** Una tabla por tipo (`accounts`, `finance_cards`, etc.). Limpio por tipo, pero requiere una migración de schema cada vez que se agrega un nuevo tipo, y complica las consultas cruzadas como `vext list`.

**Opción B — Tabla polimórfica con blob cifrado:** Una sola tabla `secrets` donde cada fila guarda un campo `type` y un blob opaco cifrado. El blob contiene un objeto JSON cuya forma depende del tipo.

**Vext usa la Opción B.** El schema de base de datos nunca cambia cuando se agregan nuevos tipos de secreto. El sistema de tipos vive enteramente en código Go, no en columnas SQL.

---

## Schema

Definido como un struct GORM en `shared/storage/schemas.go`:

```go
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

El SQL equivalente:

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

### Referencia de campos

| Campo | Tipo | Descripción |
|---|---|---|
| `id` | INTEGER | Clave primaria auto-incremental. Uso interno. |
| `name` | TEXT | Identificador del usuario. Ej: `"github"`, `"visa-debito"`. Debe ser único. |
| `type` | TEXT | Tag del tipo de secreto. Ej: `"account"`, `"finance"`. Controla la deserialización en Go. |
| `salt` | BLOB | 16 bytes aleatorios por registro. Usado por Argon2id para derivar la clave de cifrado. |
| `nonce` | BLOB | 12 bytes aleatorios por registro. Usado por AES-GCM como vector de inicialización. |
| `encrypted` | BLOB | El ciphertext de AES-GCM. Contiene el JSON del secreto + el tag de autenticación GCM. |
| `created_at` | DATETIME | Timestamp de creación. Gestionado automáticamente por GORM. |
| `updated_at` | DATETIME | Timestamp de última modificación. Gestionado automáticamente por GORM. |

---

## Repositorio: Operaciones CRUD

El acceso a datos vive en `shared/storage/repository.go`. La interfaz que expone está definida en `secrets/bases.go`:

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

Detalles de cada operación:

- **`Create`** — Inserta un nuevo secreto con su material criptográfico. Retorna `ErrAlreadyExists` si el nombre ya existe (detectado por string inspection del error de SQLite: `"UNIQUE constraint failed"`).
- **`GetByName`** — Recupera un secreto con su salt, nonce y blob cifrado. Se usa en `vext get` y `vext upd`.
- **`Update`** — Reemplaza el material criptográfico completo (salt, nonce y encrypted) por nombre. Retorna `ErrSecretNotFound` si el nombre no existe.
- **`Delete`** — Elimina el registro por nombre. Retorna `ErrSecretNotFound` si no existe.
- **`ListAll`** — Retorna solo los metadatos (`id`, `name`, `type`, `created_at`, `updated_at`) sin columnas blob. Se usa en `vext list`. La query es: `SELECT id, name, type, created_at, updated_at FROM secrets ORDER BY name asc`.
- **`GetAll`** — Retorna todos los secretos con su payload cifrado. Se usa en `vext export`.

Todas las operaciones reciben un `context.Context` para soporte de cancelación.

---

## Ciclo de Vida de la Conexión

La conexión a la base de datos se abre una vez al inicio de la aplicación y se cierra cuando el proceso termina. Esto ocurre en `cmd/adapters/bootstrap.go`:

```go
func Build() (App, func(), error) {
    dbPath := storage.DBPath()

    db, err := storage.Open(dbPath)
    if err != nil {
        return App{}, nil, fmt.Errorf("open database: %w", err)
    }

    return App{
        DBPath:     dbPath,
        Repository: storage.NewSecretRepository(db),
        Encryptor:  cryptors.NewAESGCMEncryptor(cryptors.DefaultConfig()),
        Collector:  collectors.NewCollector(collectors.NewPrompter()),
    },
    func() { _ = storage.Close(db) }, // cleanup: cierra la BD al salir
    nil
}
```

`Build()` retorna una función de cleanup que `cmd/root.go` registra con `defer cleanup()`. Todos los subcomandos reciben el mismo `App` con el repositorio ya construido.

---

## Inicialización: `vext init`

El comando `vext init` usa un `Initialiser` (`shared/storage/initialiser.go`) que:

1. Crea el directorio de configuración con permisos `0700` (solo el propietario puede leer/escribir).
2. Abre la base de datos (la crea si no existe).
3. Ejecuta `AutoMigrate` para crear las tablas faltantes.

```go
func (i *Initialiser) Setup(ctx context.Context) (*gorm.DB, error) {
    if err := os.MkdirAll(filepath.Dir(i.dbPath), 0700); err != nil {
        return nil, fmt.Errorf("could not create config directory: %w", err)
    }
    db, err := Open(i.dbPath)
    if err != nil {
        return nil, err
    }
    if err := Migrate(db); err != nil {
        Close(db)
        return nil, err
    }
    return db, nil
}
```

---

## Migraciones

Vext usa `AutoMigrate` de GORM, ejecutado en cada `vext init`. Es seguro e idempotente: solo crea tablas faltantes o agrega columnas nuevas. Nunca elimina columnas ni cambia las existentes.

```go
func Migrate(db *gorm.DB) error {
    return db.AutoMigrate(&SecretRecord{})
}
```

Para agregar una nueva columna en el futuro basta con agregarla al struct `SecretRecord` con el tag GORM apropiado.

---

## Payloads por Tipo

Estas son las estructuras JSON que se cifran y se guardan en `encrypted`. La base de datos nunca ve estos campos directamente — solo almacena los bytes cifrados.

### type: `"account"`

```json
{
  "username": "bob@example.com",
  "password": "<bytes en base64>"
}
```

### type: `"finance"`

```json
{
  "card_number": "4111111111111111",
  "security_code": "<bytes en base64>",
  "card_pin": "<bytes en base64>",
  "expiration_month": 12,
  "expiration_year": 2027,
  "bank_username": "bob",
  "bank_password": "<bytes en base64>",
  "bank_virtual_key": "<bytes en base64>",
  "bank_cellphone": "+1234567890"
}
```

Agregar un nuevo tipo solo requiere: definir un nuevo struct Go en `secrets/`, agregar un nuevo `case` al `Collector.Payload()` y al dispatch de `vext get`, y documentarlo aquí. **El schema SQL no cambia.**

---

## Cosas a Tener en Cuenta

**Archivos WAL y journal:** SQLite puede crear archivos `-wal` o `-journal` junto a `vext.db` durante escrituras. Son temporales. Dado que Vext cifra antes de escribir, estos archivos nunca contendrán secretos en texto plano — solo blobs cifrados.

**Permisos del directorio:** El directorio `~/.config/vext/` se crea con permisos `0700`. Esto previene que otros usuarios de la misma máquina lean el archivo de base de datos.

**Errores de constraint único:** GORM no envuelve la violación de constraint único de SQLite en un error tipado. `SecretRepository.Create` lo detecta via inspección de string (`"UNIQUE constraint failed"`) y lo mapea a `secrets.ErrAlreadyExists`.
