# Arquitectura Refactor - VEXT Password Manager

**Documento:** 0001-architecture-refactor.md  
**Fecha:** Junio 2026  
**Estado:** Propuesta  
**Objetivo:** Mejorar escalabilidad, mantenibilidad, testabilidad y seguridad del proyecto

---

## Tabla de Contenidos

1. [Resumen Ejecutivo](#resumen-ejecutivo)
2. [Análisis de la Arquitectura Actual](#análisis-de-la-arquitectura-actual)
3. [Problemas Identificados](#problemas-identificados)
4. [Arquitectura Propuesta](#arquitectura-propuesta)
5. [Estructura de Carpetas Mejorada](#estructura-de-carpetas-mejorada)
6. [Ejemplos de Implementación](#ejemplos-de-implementación)
7. [Plan de Migración](#plan-de-migración)
8. [Beneficios de la Refactorización](#beneficios-de-la-refactorización)
9. [Consideraciones de Seguridad](#consideraciones-de-seguridad)

---

## Resumen Ejecutivo

La arquitectura actual de VEXT mezcla lógica de negocio con interfaz CLI, lo que dificulta:
- ❌ Testing sin I/O
- ❌ Reutilización en otras interfaces (HTTP API, librería)
- ❌ Escalabilidad para nuevos tipos de secretos (Card, Note, SSH key)
- ❌ Mantenimiento y cambios seguros

**Solución propuesta:** Implementar **Clean Architecture con Domain-Driven Design (DDD)** separando claramente:
1. **Domain Layer** (reglas de negocio puras)
2. **Application Layer** (orquestación de use cases)
3. **Infrastructure Layer** (implementaciones técnicas)
4. **Interface Layer** (CLI handlers, presentación)

---

## Análisis de la Arquitectura Actual

### Estructura Actual

```
source/
├── main.go
├── cmd/                    ← Lógica de CLI + dominio mezclados
│   ├── add.go             ← Contiene: I/O + validación + crypto + DB
│   ├── get.go
│   ├── init.go
│   ├── list.go
│   ├── rm.go
│   └── root.go
└── pkg/
    ├── crypto/            ← Infraestructura pura ✓
    ├── database/          ← Infraestructura pura ✓
    └── models/            ← Modelos + payloads (confundidos)
```

### Características Positivas

✅ Separación de responsabilidades en paquetes (crypto, database)  
✅ Uso de GORM para abstracción de BD  
✅ Cobra para CLI robusto  
✅ Encriptación fuerte (AES-256-GCM + Argon2id)  
✅ Manejo de permisos de archivo  
✅ Limpieza de secretos en memoria (crypto.Zero)

---

## Problemas Identificados

### 1. **Manejo de Secretos en Memoria (CRÍTICO - Seguridad)**

**Problema:**
```go
// En cmd/add.go - MALO
payload := models.AccountPayload{
    Username: username,
    Password: string(servicePasswordBytes),  // ❌ Conversión a string
}
plaintextJSON, err := json.Marshal(payload)
```

**Impacto:**
- Las strings en Go son inmutables y no se pueden sobrescribir
- `crypto.Zero([]byte(payload.Password))` crea una copia que no elimina el original
- Secretos persisten en memoria/swap/dumps

**Causa raíz:** `AccountPayload.Password` es `string`, no `[]byte`

---

### 2. **Lógica Mezclada CLI + Dominio (ALTA - Mantenibilidad)**

**Problema:**
- `add.go` hace: lectura de stdin → validación → crypto → DB insert
- No hay forma de testear sin CLI mocking
- Imposible reutilizar desde HTTP API o librería

**Impacto:**
- Testing complicado y frágil
- Código duplicado si se agrega otra interfaz
- Difícil cambiar componentes (ej: cambiar de SQLite a PostgreSQL)

---

### 3. **Reapertura Repetida de Base de Datos (MEDIA - Performance)**

**Problema:**
```go
// Cada comando hace esto:
db, err := database.Open(dbPath)
sqlDB, err := db.DB()
defer sqlDB.Close()
```

**Impacto:**
- Overhead de conexión en cada operación
- Dificulta transacciones, pooling, control centralizado
- Complica testing (no hay control sobre ciclo de vida de DB)

---

### 4. **Errores No Mapeados / Genéricos (MEDIA - UX)**

**Problema:**
```go
// En add.go - infiere duplicados del mensaje de error
if err := database.Insert(db, record); err != nil {
    return fmt.Errorf("[X] Error: a credential named %q already exists...", name)
}
```

**Impacto:**
- No hay tipos específicos: `ErrAlreadyExists`, `ErrNotFound`
- Difícil inspeccionar/testear (¿qué tipo de error es?)
- Cambiar mensaje de DB puede romper lógica

---

### 5. **Falta de Interfaces/Abstracciones (MEDIA - Testabilidad)**

**Problema:**
- Dependencias directas: `database.Open`, `crypto.Encrypt`, `term.ReadPassword`
- No hay forma de inyectar mocks para testing

**Impacto:**
- Tests de integración forzados (requieren BD real, I/O real)
- Impossible unit tests puros

---

### 6. **Ausencia de Tests y CI (ALTA - Confianza)**

**Problema:**
- No hay tests unitarios para crypto, DB, comandos
- No hay linters, vet, staticcheck

**Impacto:**
- Riesgo de regresiones silenciosas
- Cambios inseguros en crypto
- Deuda técnica sin control

---

### 7. **Gestión de Configuración Centralizada (BAJA - Escalabilidad)**

**Problema:**
- Parámetros Argon2 hardcodeados en `pkg/crypto/crypto.go`
- Paths de BD sin abstracción

**Impacto:**
- Difícil ajustar parámetros de seguridad
- No hay forma de usar diferentes configs por ambiente

---

### 8. **Logging y Observabilidad Limitados (BAJA - Operaciones)**

**Problema:**
- Solo `fmt.Printf` para mensajes
- No hay diferenciación entre logs de usuario y debug
- No hay structured logging

**Impacto:**
- Difícil debuggear en producción
- No hay auditoría

---

### 9. **Duplicidad de Código CLI (BAJA - Mantenibilidad)**

**Problema:**
```go
// Patrón repetido en cada comando:
dbPath, err := database.DBPath()
if err != nil { return err }
db, err := database.Open(dbPath)
if err != nil { return err }
sqlDB, err := db.DB()
if err != nil { return err }
defer sqlDB.Close()
```

**Impacto:**
- Código repetitivo
- Cambios requieren editar múltiples archivos

---

## Arquitectura Propuesta

### Clean Architecture con DDD

La solución implementa una arquitectura en capas con dependencias unidireccionales:

```
┌─────────────────────────────────────────────┐
│            INTERFACE LAYER (CLI)            │
│   cmd/handlers/ + cmd/ui/                   │
│   ↓ (depende de)                            │
├─────────────────────────────────────────────┤
│          APPLICATION LAYER                  │
│   Use Cases (orquestación)                  │
│   ↓ (depende de)                            │
├─────────────────────────────────────────────┤
│           DOMAIN LAYER (Core)               │
│   Entidades, Value Objects, Interfaces      │
│   Reglas de negocio PURAS                   │
│   ↓ (depende de)                            │
├─────────────────────────────────────────────┤
│         INFRASTRUCTURE LAYER                │
│   SQLite, AES-GCM, Config Files             │
└─────────────────────────────────────────────┘
```

**Reglas de flujo:**
- 🔴 **Infrastructure** nunca depende de nada (código técnico puro)
- 🟡 **Domain** nunca depende de Infrastructure o Application
- 🟢 **Application** depende de Domain e inyecta Infrastructure
- 🔵 **Interface** depende de Application para ejecutar use cases

---

## Estructura de Carpetas Mejorada

```
source/
├── main.go                          ← Entry point
│
├── cmd/                             ← Interface Layer: CLI
│   ├── root.go                      ← Root command (cobra setup)
│   ├── handlers/
│   │   ├── init_handler.go
│   │   ├── add_handler.go
│   │   ├── get_handler.go
│   │   ├── list_handler.go
│   │   └── rm_handler.go
│   └── ui/                          ← Presentación & I/O
│       ├── prompter.go              ← Interfaz: lectura stdin
│       └── formatter.go             ← Formateo de output
│
├── pkg/
│   ├── domain/                      ← CORE: Reglas de negocio
│   │   ├── secret.go                ← Entity: Secret + Payloads
│   │   ├── account_secret.go        ← Value Object: Account
│   │   ├── card_secret.go           ← Value Object: Card (Phase 2)
│   │   ├── note_secret.go           ← Value Object: Note (Phase 2)
│   │   ├── secret_repository.go     ← Interface: Contrato persistencia
│   │   ├── encryptor.go             ← Interface: Contrato encriptación
│   │   └── errors.go                ← Errores de dominio
│   │
│   ├── application/                 ← Use Cases: Orquestación
│   │   ├── store_secret_uc.go       ← UC: Guardar credencial
│   │   ├── retrieve_secret_uc.go    ← UC: Obtener credencial
│   │   ├── list_secrets_uc.go       ← UC: Listar
│   │   ├── delete_secret_uc.go      ← UC: Eliminar
│   │   └── init_storage_uc.go       ← UC: Inicializar DB
│   │
│   ├── infrastructure/              ← Implementaciones técnicas
│   │   ├── crypto/
│   │   │   ├── aes_gcm_encryptor.go ← Impl: Encryptor
│   │   │   └── key_derivation.go    ← Argon2id helpers
│   │   │
│   │   ├── storage/
│   │   │   ├── sqlite_repository.go ← Impl: SecretRepository
│   │   │   ├── db.go                ← Manejo de conexión GORM
│   │   │   └── migrations.go        ← Schema y migraciones
│   │   │
│   │   └── config/
│   │       └── app_config.go        ← Paths, permisos, flags
│   │
│   ├── models/                      ← Data Transfer Objects (DTOs)
│   │   ├── secret_dto.go            ← DTO para serialización
│   │   └── secret_record.go         ← DB schema (GORM)
│   │
│   └── shared/                      ← Utilidades compartidas
│       ├── crypto_utils.go          ← Zero, GenerateSalt
│       ├── errors.go                ← Tipos de error comunes
│       └── types.go                 ← Constantes, tipos auxiliares
│
├── tests/                           ← Tests (espejo de estructura)
│   ├── domain/
│   │   └── secret_test.go
│   ├── application/
│   │   ├── store_secret_uc_test.go
│   │   └── retrieve_secret_uc_test.go
│   ├── infrastructure/
│   │   ├── crypto/
│   │   │   └── aes_gcm_encryptor_test.go
│   │   └── storage/
│   │       └── sqlite_repository_test.go
│   └── cmd/
│       └── handlers/
│           └── add_handler_test.go
│
└── README.md
```

---

## Ejemplos de Implementación

### 1. Domain Layer: Entidades y Contratos

#### `pkg/domain/secret.go`

```go
package domain

import "time"

// SecretID es un value object que representa la identidad única
type SecretID int64

// Secret es la entidad principal del dominio
// No contiene detalles de DB ni crypto
type Secret struct {
	ID        SecretID
	Name      string
	Type      string          // "account", "card", "note"
	Payload   SecretPayload   // Interface: polimórfico
	Salt      []byte          // Generado por Encryptor
	Nonce     []byte          // Generado por Encryptor
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SecretPayload es una interfaz que todos los tipos de secreto implementan
type SecretPayload interface {
	GetType() string
	Validate() error
}
```

#### `pkg/domain/account_secret.go`

```go
package domain

import "fmt"

// AccountSecret es un payload para credenciales tipo "account"
type AccountSecret struct {
	Username string `json:"username"`
	Password []byte `json:"password"` // ✅ SIEMPRE []byte, nunca string
}

func (a *AccountSecret) GetType() string {
	return "account"
}

func (a *AccountSecret) Validate() error {
	if a.Username == "" {
		return NewDomainError("username is required")
	}
	if len(a.Password) == 0 {
		return NewDomainError("password is required")
	}
	if len(a.Username) > 255 {
		return NewDomainError("username too long")
	}
	return nil
}
```

#### `pkg/domain/secret_repository.go`

```go
package domain

import "context"

// SecretRepository define el contrato para persistencia
// La implementación está en infrastructure/
type SecretRepository interface {
	// Save persiste un secreto con payload encriptado
	Save(ctx context.Context, secret *Secret, encryptedPayload []byte) error
	
	// GetByName recupera un secreto por nombre (encriptado)
	// Retorna: entidad de dominio + payload encriptado
	GetByName(ctx context.Context, name string) (*Secret, []byte, error)
	
	// ListAll devuelve metadatos (SIN desencriptar)
	ListAll(ctx context.Context) ([]Secret, error)
	
	// Delete elimina un secreto
	Delete(ctx context.Context, name string) error
}
```

#### `pkg/domain/encryptor.go`

```go
package domain

import "context"

// Encryptor define el contrato para encriptación/desencriptación
type Encryptor interface {
	// Encrypt cifra plaintext con masterPassword
	// Devuelve salt + nonce + ciphertext (para almacenar en BD)
	Encrypt(ctx context.Context, plaintext, masterPassword []byte) (salt, nonce, ciphertext []byte, err error)
	
	// Decrypt desencripta usando salt+nonce del record y masterPassword
	Decrypt(ctx context.Context, masterPassword, salt, nonce, ciphertext []byte) ([]byte, error)
}
```

#### `pkg/domain/errors.go`

```go
package domain

import "errors"

// DomainError representa un error del negocio (no técnico)
type DomainError struct {
	message string
}

func NewDomainError(msg string) DomainError {
	return DomainError{message: msg}
}

func (e DomainError) Error() string {
	return e.message
}

// Errores de dominio canónicos
var (
	ErrSecretNotFound    = NewDomainError("secret not found")
	ErrAlreadyExists     = NewDomainError("secret already exists")
	ErrInvalidCredential = NewDomainError("invalid credential format")
	ErrDecryptionFailed  = NewDomainError("master password incorrect or data corrupted")
	ErrInvalidInput      = NewDomainError("invalid input")
)

// IsDomainError verifica si es error de dominio
func IsDomainError(err error) bool {
	_, ok := err.(DomainError)
	return ok
}

// AsDecryptionFailed verifica si es error de desencriptación
func AsDecryptionFailed(err error) bool {
	return errors.Is(err, ErrDecryptionFailed)
}
```

---

### 2. Application Layer: Use Cases

#### `pkg/application/store_secret_uc.go`

```go
package application

import (
	"context"
	"encoding/json"
	"fmt"
	"vextpss/source/pkg/domain"
	"vextpss/source/pkg/shared"
)

// StoreSecretUC es el Use Case para almacenar un nuevo secreto
type StoreSecretUC struct {
	repo      domain.SecretRepository
	encryptor domain.Encryptor
}

func NewStoreSecretUC(
	repo domain.SecretRepository,
	enc domain.Encryptor,
) *StoreSecretUC {
	return &StoreSecretUC{repo: repo, encryptor: enc}
}

// StoreSecretRequest es el DTO de entrada (CLI/API)
type StoreSecretRequest struct {
	Name              string
	Type              string // "account", "card", etc.
	MasterPassword    []byte
	AccountUsername   string
	AccountPassword   []byte
	// ... otros campos según tipo
}

// Execute orquesta: validar → construir payload → encriptar → persistir
func (uc *StoreSecretUC) Execute(ctx context.Context, req StoreSecretRequest) error {
	// 1. Validar request
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	// 2. Construir payload tipado
	payload, err := req.BuildPayload()
	if err != nil {
		return err
	}

	// 3. Validar payload del dominio
	if err := payload.Validate(); err != nil {
		return err
	}

	// 4. Serializar a JSON (plaintext temporalmente)
	plaintext, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("serialization failed: %w", err)
	}
	defer shared.Zero(plaintext) // ✅ Limpiar inmediatamente

	// 5. Encriptar con contraseña del usuario
	salt, nonce, ciphertext, err := uc.encryptor.Encrypt(
		ctx,
		plaintext,
		req.MasterPassword,
	)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// 6. Construir entidad
	secret := &domain.Secret{
		Name:      req.Name,
		Type:      req.Type,
		Salt:      salt,
		Nonce:     nonce,
		Payload:   payload,
		CreatedAt: shared.Now(),
		UpdatedAt: shared.Now(),
	}

	// 7. Persistir (nunca plaintext, siempre encriptado)
	if err := uc.repo.Save(ctx, secret, ciphertext); err != nil {
		// Si es error de dominio (ej: duplicado), propagar tal cual
		if domain.IsDomainError(err) {
			return err
		}
		return fmt.Errorf("failed to store secret: %w", err)
	}

	return nil
}

// Validate valida que el request sea completo
func (r *StoreSecretRequest) Validate() error {
	if r.Name == "" {
		return domain.NewDomainError("name is required")
	}
	if r.Type != "account" && r.Type != "card" {
		return domain.NewDomainError("unsupported secret type")
	}
	if len(r.MasterPassword) == 0 {
		return domain.NewDomainError("master password is required")
	}
	return nil
}

// BuildPayload construye el payload específico según tipo
func (r *StoreSecretRequest) BuildPayload() (domain.SecretPayload, error) {
	switch r.Type {
	case "account":
		return &domain.AccountSecret{
			Username: r.AccountUsername,
			Password: r.AccountPassword, // ✅ []byte
		}, nil
	case "card":
		return &domain.CardSecret{
			CardNumber: r.CardNumber,
			CVV:        r.CardCVV,
		}, nil
	default:
		return nil, domain.NewDomainError("unknown secret type")
	}
}
```

#### `pkg/application/retrieve_secret_uc.go`

```go
package application

import (
	"context"
	"encoding/json"
	"fmt"
	"vextpss/source/pkg/domain"
	"vextpss/source/pkg/shared"
)

// RetrieveSecretUC es el Use Case para obtener un secreto
type RetrieveSecretUC struct {
	repo      domain.SecretRepository
	encryptor domain.Encryptor
}

func NewRetrieveSecretUC(
	repo domain.SecretRepository,
	enc domain.Encryptor,
) *RetrieveSecretUC {
	return &RetrieveSecretUC{repo: repo, encryptor: enc}
}

// RetrieveSecretRequest es el input
type RetrieveSecretRequest struct {
	Name           string
	MasterPassword []byte
}

// RetrieveSecretResponse es el output (seguro para serializar)
type RetrieveSecretResponse struct {
	Name    string
	Type    string
	Payload domain.SecretPayload
}

// Execute: obtener → desencriptar → retornar DTO
func (uc *RetrieveSecretUC) Execute(
	ctx context.Context,
	req RetrieveSecretRequest,
) (*RetrieveSecretResponse, error) {
	// 1. Obtener del repo (encriptado)
	secret, ciphertext, err := uc.repo.GetByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	// 2. Desencriptar con contraseña del usuario
	plaintext, err := uc.encryptor.Decrypt(
		ctx,
		req.MasterPassword,
		secret.Salt,
		secret.Nonce,
		ciphertext,
	)
	defer shared.Zero(plaintext)
	if err != nil {
		// No retornar detalles técnicos; usar error genérico
		return nil, domain.ErrDecryptionFailed
	}

	// 3. Deserializar según tipo
	var payload domain.SecretPayload
	switch secret.Type {
	case "account":
		var ap domain.AccountSecret
		if err := json.Unmarshal(plaintext, &ap); err != nil {
			return nil, fmt.Errorf("failed to parse secret: %w", err)
		}
		payload = &ap

	case "card":
		var cp domain.CardSecret
		if err := json.Unmarshal(plaintext, &cp); err != nil {
			return nil, fmt.Errorf("failed to parse secret: %w", err)
		}
		payload = &cp

	default:
		return nil, fmt.Errorf("unknown secret type: %s", secret.Type)
	}

	return &RetrieveSecretResponse{
		Name:    secret.Name,
		Type:    secret.Type,
		Payload: payload,
	}, nil
}
```

---

### 3. Infrastructure Layer: Implementaciones

#### `pkg/infrastructure/storage/sqlite_repository.go`

```go
package storage

import (
	"context"
	"errors"
	"fmt"
	"vextpss/source/pkg/domain"
	"vextpss/source/pkg/models"
	"gorm.io/gorm"
)

// SQLiteRepository implementa domain.SecretRepository
type SQLiteRepository struct {
	db *gorm.DB
}

func NewSQLiteRepository(db *gorm.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// Save persiste un secreto con payload encriptado
func (r *SQLiteRepository) Save(
	ctx context.Context,
	secret *domain.Secret,
	encryptedPayload []byte,
) error {
	record := &models.SecretRecord{
		Name:             secret.Name,
		Type:             secret.Type,
		Salt:             secret.Salt,
		Nonce:            secret.Nonce,
		EncryptedPayload: encryptedPayload,
	}

	result := r.db.WithContext(ctx).Create(record)
	if result.Error != nil {
		// Mapear constraint violation de SQLite a error de dominio
		if isUniqueConstraintError(result.Error) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("insert failed: %w", result.Error)
	}
	return nil
}

// GetByName recupera un secreto (encriptado)
func (r *SQLiteRepository) GetByName(
	ctx context.Context,
	name string,
) (*domain.Secret, []byte, error) {
	var record models.SecretRecord
	result := r.db.WithContext(ctx).Where("name = ?", name).First(&record)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil, domain.ErrSecretNotFound
	}
	if result.Error != nil {
		return nil, nil, fmt.Errorf("query failed: %w", result.Error)
	}

	secret := &domain.Secret{
		ID:        domain.SecretID(record.ID),
		Name:      record.Name,
		Type:      record.Type,
		Salt:      record.Salt,
		Nonce:     record.Nonce,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}

	return secret, record.EncryptedPayload, nil
}

// ListAll devuelve metadatos (sin desencriptar)
func (r *SQLiteRepository) ListAll(ctx context.Context) ([]domain.Secret, error) {
	var records []models.SecretRecord
	result := r.db.WithContext(ctx).
		Select("id, name, type, created_at, updated_at").
		Order("name asc").
		Find(&records)

	if result.Error != nil {
		return nil, fmt.Errorf("list query failed: %w", result.Error)
	}

	secrets := make([]domain.Secret, len(records))
	for i, rec := range records {
		secrets[i] = domain.Secret{
			ID:        domain.SecretID(rec.ID),
			Name:      rec.Name,
			Type:      rec.Type,
			CreatedAt: rec.CreatedAt,
			UpdatedAt: rec.UpdatedAt,
		}
	}
	return secrets, nil
}

// Delete elimina un secreto
func (r *SQLiteRepository) Delete(ctx context.Context, name string) error {
	result := r.db.WithContext(ctx).Where("name = ?", name).Delete(&models.SecretRecord{})

	if result.Error != nil {
		return fmt.Errorf("delete failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrSecretNotFound
	}
	return nil
}

// isUniqueConstraintError detecta si el error es de constraint UNIQUE
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	// SQLite: "UNIQUE constraint failed: secrets.name"
	errMsg := err.Error()
	return strings.Contains(errMsg, "UNIQUE constraint failed")
}
```

#### `pkg/infrastructure/crypto/aes_gcm_encryptor.go`

```go
package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"vextpss/source/pkg/domain"
	"vextpss/source/pkg/shared"
	"golang.org/x/crypto/argon2"
)

// AESGCMEncryptor implementa domain.Encryptor
type AESGCMEncryptor struct {
	// Parámetros Argon2 (centralizados y configurables)
	argonTime    uint32
	argonMemory  uint32
	argonThreads uint8
}

// NewAESGCMEncryptor crea un encryptor con parámetros por defecto
func NewAESGCMEncryptor() *AESGCMEncryptor {
	return &AESGCMEncryptor{
		argonTime:    3,        // 3 iteraciones
		argonMemory:  64 * 1024, // 64 MB
		argonThreads: 2,        // 2 threads
	}
}

// Encrypt cifra plaintext y retorna salt + nonce + ciphertext
func (e *AESGCMEncryptor) Encrypt(
	ctx context.Context,
	plaintext, masterPassword []byte,
) (salt, nonce, ciphertext []byte, err error) {
	// 1. Generar salt aleatorio (16 bytes)
	salt, err = shared.GenerateSalt()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// 2. Derivar key con Argon2id
	key := e.deriveKey(masterPassword, salt)
	defer shared.Zero(key)

	// 3. Generar nonce aleatorio (12 bytes para GCM)
	nonce = make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 4. Crear cipher AES-256-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, err
	}

	// 5. Encriptar
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)

	return salt, nonce, ciphertext, nil
}

// Decrypt desencripta usando salt+nonce del record
func (e *AESGCMEncryptor) Decrypt(
	ctx context.Context,
	masterPassword, salt, nonce, ciphertext []byte,
) ([]byte, error) {
	// 1. Derivar key con mismo salt
	key := e.deriveKey(masterPassword, salt)
	defer shared.Zero(key)

	// 2. Crear cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		// No revelar detalle de error
		return nil, domain.ErrDecryptionFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, domain.ErrDecryptionFailed
	}

	// 3. Desencriptar (falla si password es incorrecta o datos corruptos)
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, domain.ErrDecryptionFailed
	}

	return plaintext, nil
}

// deriveKey deriva una key de 32 bytes usando Argon2id
func (e *AESGCMEncryptor) deriveKey(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, e.argonTime, e.argonMemory, e.argonThreads, 32)
}
```

---

### 4. Interface Layer: CLI Handlers

#### `cmd/handlers/add_handler.go`

```go
package handlers

import (
	"context"
	"fmt"
	"vextpss/source/cmd/ui"
	"vextpss/source/pkg/application"
	"vextpss/source/pkg/domain"
	"github.com/spf13/cobra"
)

// AddHandler encapsula la lógica de CLI para `vext add`
type AddHandler struct {
	storeUC *application.StoreSecretUC
	prompter ui.Prompter
}

func NewAddHandler(
	storeUC *application.StoreSecretUC,
	prompter ui.Prompter,
) *AddHandler {
	return &AddHandler{storeUC: storeUC, prompter: prompter}
}

// CobraCommand devuelve el comando Cobra configurado
func (h *AddHandler) CobraCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name>",
		Short: "Store a new credential",
		Long:  "Interactively stores a new account credential under the given name.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.Handle(context.Background(), args[0])
		},
	}
}

// Handle es la lógica del handler
func (h *AddHandler) Handle(ctx context.Context, name string) error {
	// 1. Leer entrada de usuario usando abstracción de I/O
	username, err := h.prompter.ReadUsername()
	if err != nil {
		return fmt.Errorf("could not read username: %w", err)
	}

	password, err := h.prompter.ReadPassword("Password: ")
	if err != nil {
		return fmt.Errorf("could not read password: %w", err)
	}
	defer h.prompter.Zero(password)

	masterPassword, err := h.prompter.ReadPassword("Master Password: ")
	if err != nil {
		return fmt.Errorf("could not read master password: %w", err)
	}
	defer h.prompter.Zero(masterPassword)

	// 2. Construir request para el use case
	req := application.StoreSecretRequest{
		Name:            name,
		Type:            "account",
		MasterPassword:  masterPassword,
		AccountUsername: username,
		AccountPassword: password,
	}

	// 3. Ejecutar use case (toda la lógica de negocio)
	if err := h.storeUC.Execute(ctx, req); err != nil {
		// Traducir errores de dominio a mensajes amigables
		if domain.IsDomainError(err) {
			fmt.Printf("[X] %s\n", err.Error())
			return nil
		}
		// Errores técnicos (infra): loguear pero no exponer
		fmt.Printf("[X] An error occurred. Please try again.\n")
		return nil
	}

	fmt.Printf("[✓] Credential %q saved.\n", name)
	return nil
}
```

#### `cmd/ui/prompter.go`

```go
package ui

import (
	"fmt"
	"os"
	"golang.org/x/term"
)

// Prompter es la interfaz para I/O (permite mocking en tests)
type Prompter interface {
	ReadUsername() (string, error)
	ReadPassword(prompt string) ([]byte, error)
	Zero(b []byte)
	Confirm(prompt string) (bool, error)
}

// CLIPrompter implementa Prompter para CLI real
type CLIPrompter struct{}

func (p *CLIPrompter) ReadUsername() (string, error) {
	fmt.Print("Username: ")
	var username string
	_, err := fmt.Fscan(os.Stdin, &username)
	return username, err
}

func (p *CLIPrompter) ReadPassword(prompt string) ([]byte, error) {
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return password, err
}

func (p *CLIPrompter) Zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func (p *CLIPrompter) Confirm(prompt string) (bool, error) {
	fmt.Print(prompt)
	var response string
	_, err := fmt.Fscan(os.Stdin, &response)
	if err != nil {
		return false, err
	}
	return response == "y" || response == "yes", nil
}

// MockPrompter para testing (implementa Prompter)
type MockPrompter struct {
	Username string
	Password []byte
	Confirm  bool
}

func (m *MockPrompter) ReadUsername() (string, error) {
	return m.Username, nil
}

func (m *MockPrompter) ReadPassword(prompt string) ([]byte, error) {
	return m.Password, nil
}

func (m *MockPrompter) Zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func (m *MockPrompter) Confirm(prompt string) (bool, error) {
	return m.Confirm, nil
}
```

---

### 5. Main: Inyección de Dependencias

#### `source/main.go`

```go
package main

import (
	"fmt"
	"os"
	"vextpss/source/cmd"
	"vextpss/source/cmd/handlers"
	"vextpss/source/cmd/ui"
	"vextpss/source/pkg/application"
	"vextpss/source/pkg/infrastructure/crypto"
	"vextpss/source/pkg/infrastructure/storage"
	"vextpss/source/pkg/infrastructure/config"
)

func main() {
	// 1. Configuración
	cfg := config.Load()

	// 2. Inicializar DB UNA VEZ (ciclo de vida centralizado)
	db, err := storage.OpenDB(cfg.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not open database: %v\n", err)
		os.Exit(1)
	}
	defer storage.CloseDB(db)

	// 3. Aplicar migraciones
	if err := storage.Migrate(db); err != nil {
		fmt.Fprintf(os.Stderr, "Error: migration failed: %v\n", err)
		os.Exit(1)
	}

	// 4. Inyectar dependencias (inversión de control)
	prompter := &ui.CLIPrompter{}
	repo := storage.NewSQLiteRepository(db)
	encryptor := crypto.NewAESGCMEncryptor()

	// 5. Crear use cases
	storeUC := application.NewStoreSecretUC(repo, encryptor)
	retrieveUC := application.NewRetrieveSecretUC(repo, encryptor)
	listUC := application.NewListSecretsUC(repo)
	deleteUC := application.NewDeleteSecretUC(repo)

	// 6. Crear handlers CLI
	addHandler := handlers.NewAddHandler(storeUC, prompter)
	getHandler := handlers.NewGetHandler(retrieveUC, prompter)
	listHandler := handlers.NewListHandler(listUC)
	rmHandler := handlers.NewRmHandler(deleteUC, prompter)

	// 7. Construir árbol de comandos
	rootCmd := cmd.NewRootCmd()
	rootCmd.AddCommand(addHandler.CobraCommand())
	rootCmd.AddCommand(getHandler.CobraCommand())
	rootCmd.AddCommand(listHandler.CobraCommand())
	rootCmd.AddCommand(rmHandler.CobraCommand())

	// 8. Ejecutar
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

---

## Plan de Migración

### Fase 1: Preparación (0.5 días)
- [ ] Crear estructura de directorios
- [ ] Crear archivos base del domain layer

### Fase 2: Domain Layer (1 día)
- [ ] `pkg/domain/secret.go`, `account_secret.go`, `card_secret.go`
- [ ] `pkg/domain/secret_repository.go`
- [ ] `pkg/domain/encryptor.go`
- [ ] `pkg/domain/errors.go`

### Fase 3: Application Layer (1.5 días)
- [ ] `pkg/application/store_secret_uc.go`
- [ ] `pkg/application/retrieve_secret_uc.go`
- [ ] `pkg/application/list_secrets_uc.go`
- [ ] `pkg/application/delete_secret_uc.go`

### Fase 4: Infrastructure Refactor (1 día)
- [ ] Mover y adaptar `pkg/crypto` → `pkg/infrastructure/crypto`
- [ ] Mover y adaptar `pkg/database` → `pkg/infrastructure/storage`
- [ ] Implementar `SQLiteRepository`
- [ ] Implementar `AESGCMEncryptor`

### Fase 5: Interface Layer (1 día)
- [ ] `cmd/handlers/*_handler.go` (refactor de `cmd/*.go`)
- [ ] `cmd/ui/prompter.go`
- [ ] `cmd/root.go` (actualizar)

### Fase 6: Inyección & Main (0.5 días)
- [ ] Refactor de `source/main.go`
- [ ] Verificar que todo funciona

### Fase 7: Tests (1.5 días)
- [ ] Tests para domain layer
- [ ] Tests para application layer (mocks de repo/encryptor)
- [ ] Tests para handlers (mocks de prompter)
- [ ] CI/CD setup

**Tiempo total:** ~6 días de desarrollo

---

## Beneficios de la Refactorización

### 1. **Testing Mejorado**

**Antes:**
```go
// Imposible testear sin I/O real
func TestAdd(t *testing.T) {
    // ??? Cómo simular stdin/stdout?
}
```

**Después:**
```go
// Tests unitarios puros
func TestStoreSecret(t *testing.T) {
    mockRepo := &MockRepository{}
    mockEnc := &MockEncryptor{}
    uc := application.NewStoreSecretUC(mockRepo, mockEnc)
    
    err := uc.Execute(context.Background(), req)
    assert.NoError(t, err)
}
```

### 2. **Reutilización en Otros Contextos**

**Escenario:** Agregar API HTTP

**Antes:** Reescribir lógica en handlers HTTP  
**Después:** Reutilizar use cases directamente

```go
// En http_handler.go
func (h *AddSecretHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    req := parseRequest(r)
    err := h.storeUC.Execute(r.Context(), req) // ✅ Reutilizar
    respondJSON(w, err)
}
```

### 3. **Fácil Agregar Nuevos Tipos de Secretos**

**Antes:** Editar `add.go`, `get.go`, `models.go`  
**Después:** Solo agregar nuevo valor objeto

```go
// pkg/domain/ssh_key_secret.go (nuevo)
type SSHKeySecret struct {
    PrivateKey []byte
    Passphrase []byte
}

func (s *SSHKeySecret) GetType() string { return "ssh_key" }
func (s *SSHKeySecret) Validate() error { /* ... */ }

// En application, el switch ya lo maneja:
case "ssh_key":
    return &domain.SSHKeySecret{...}, nil
```

### 4. **Seguridad Mejorada**

| Aspecto | Antes | Después |
|---------|-------|---------|
| Passwords en memoria | `string` (no borrables) | `[]byte` (borrable) |
| Copias de secretos | Múltiples conversiones | Minimizadas, todas borrables |
| Errores sensibles | Retornados al usuario | Mapeados a errores genéricos |
| Cambios de algoritmo | Afecta todo el código | Centralizado en Encryptor |

### 5. **Mantenibilidad**

- Cambios localizados: editar un archivo = una responsabilidad
- Clear contracts: Interfaces hacen explícitos los límites
- Logging centralizado: Facilita debugging

### 6. **Escalabilidad**

- Agregar persistencia (PostgreSQL): Solo cambiar `SQLiteRepository`
- Agregar cifrado alternativo: Solo cambiar `AESGCMEncryptor`
- Agregar caché: Envolver `SecretRepository` con proxy

---

## Consideraciones de Seguridad

### Memory Management

✅ **Siempre usar `[]byte` para secretos**
```go
// CORRECTO
type Credential struct {
    Username string  // OK: no es secreto
    Password []byte  // ✓ Borrable
}

// INCORRECTO
type Credential struct {
    Password string  // ✗ No borrable
}
```

✅ **Limpiar siempre después de usar**
```go
password, _ := prompter.ReadPassword("...")
defer prompter.Zero(password) // ✓

plaintext, _ := encryptor.Decrypt(...)
defer shared.Zero(plaintext)   // ✓
```

### Error Handling

✅ **Nunca exponer detalles técnicos**
```go
// CORRECTO: Error genérico
return domain.ErrDecryptionFailed

// INCORRECTO: Expone impl
return fmt.Errorf("AES-GCM auth failed: %v", err)
```

### Database

✅ **Nunca almacenar plaintext**
```go
// CORRECTO
func (uc *StoreUC) Execute(...) error {
    ciphertext, _ := encryptor.Encrypt(plaintext, ...)
    return repo.Save(secret, ciphertext) // Solo encriptado
}
```

✅ **Usar prepared statements (GORM lo hace)**
```go
// GORM previene SQL injection automáticamente
r.db.Where("name = ?", name).First(&record)
```

---

## Próximos Pasos

1. **Crear plan detallado de sprints** con tareas asignables
2. **Iniciar Fase 1** con estructura de directorios
3. **Hacer PR pequeños** para cada fase
4. **Tests mientras se escribe código** (TDD)
5. **Documentar decisiones** en ADR (Architecture Decision Records)

---

## Referencias

- **Clean Architecture:** Robert C. Martin ("Uncle Bob")
- **Domain-Driven Design:** Eric Evans
- **Go Interfaces:** https://golang.org/doc/effective_go#interfaces
- **SOLID Principles:** https://en.wikipedia.org/wiki/SOLID

---

**Versión:** 1.0  
**Último actualizado:** Junio 2026  
**Autor:** Code Review AI Assistant
