# Vext — Vista General

## ¿Qué es Vext?

Vext es un gestor de contraseñas CLI, local y minimalista, escrito en Go. Nació de una frustración concreta: administrar credenciales en herramientas como Notion es incómodo, inseguro y no es para lo que fueron diseñadas. Vext es la alternativa: rápida, simple y cifrada, que vive completamente en tu máquina.

Sin nube. Sin sincronización. Sin cuentas. Solo un binario y una base de datos local cifrada.

---

## Filosofía

**Local primero.** Tus datos nunca salen de tu máquina. No hay servidores, no hay APIs, no hay dependencias externas en runtime.

**Simple por diseño.** Vext hace una sola cosa: guardar y recuperar secretos de forma segura. Cada decisión de diseño favorece la claridad y la corrección por encima de la proliferación de features.

**Seguridad sin compromisos.** Aunque la herramienta es simple, la base criptográfica es de nivel producción. El modelo de cifrado es el mismo que usan gestores comerciales como Bitwarden y 1Password: Argon2id para derivación de clave + AES-256-GCM para cifrado autenticado.

**Filosofía Unix.** Cada comando hace exactamente una cosa. La interfaz es predecible y componible.

---

## El Problema que Resuelve

| Punto de dolor | Cómo lo resuelve Vext |
|---|---|
| Notion / hojas de cálculo para contraseñas | Herramienta dedicada y de propósito específico |
| Credenciales guardadas en texto plano | Cifrado AES-GCM en reposo |
| Herramientas dependientes de la nube | Completamente offline, sin cuenta necesaria |
| UX compleja para tareas simples | Binario único, un comando por acción |
| Modelos de datos rígidos | Payload JSON polimórfico por tipo de secreto |

---

## Usuario Objetivo

Vext está diseñado para desarrolladores y usuarios técnicos que:
- Se sienten cómodos con la terminal.
- Quieren control total sobre dónde viven sus datos.
- No quieren depender de un SaaS para algo tan personal.
- Valoran entender el modelo de seguridad de las herramientas que usan.

---

## Comandos Disponibles

Estos son los cinco comandos implementados actualmente:

| Comando | Descripción |
|---|---|
| `vext init` | Crea el directorio de configuración y la base de datos. Seguro para ejecutar múltiples veces. |
| `vext add <name>` | Guarda un nuevo secreto. Flag `-t` para elegir tipo (`account` o `finance`). Default: `account`. |
| `vext get <name>` | Recupera y muestra un secreto. Pide la contraseña maestra. |
| `vext list` | Lista todos los secretos guardados (solo metadatos, nunca el contenido). |
| `vext rm <name>` | Elimina un secreto permanentemente. Pide confirmación antes de proceder. |

---

## Tipos de Secretos

Vext usa un modelo polimórfico: todos los secretos se guardan en la misma tabla, y el tipo controla qué campos se recolectan y cómo se muestran.

### `account` (por defecto)

Para cuentas de servicios online: email / usuario + contraseña.

```
$ vext add github
Username: bob@example.com
Password: ••••••••
Master password: ••••••••
✓ Credential "github" saved.
```

### `finance`

Para tarjetas bancarias y credenciales financieras: número de tarjeta, PIN, CVV, vencimiento, usuario/clave del banco, clave virtual, celular.

```
$ vext add visa-debito -t finance
Card number: 4111111111111111
Card PIN: ••••
Security code: •••
...
Master password: ••••••••
✓ Credential "visa-debito" saved.
```

---

## Stack Tecnológico

| Capa | Tecnología | Razón |
|---|---|---|
| Lenguaje | Go | Binario único, excelente stdlib de crypto, compilación estática |
| CLI Framework | Cobra | Estándar de la industria para CLIs en Go, soporte limpio de subcomandos |
| Base de datos | SQLite (Pure Go) | Cero dependencias externas, archivo local, confiable en producción |
| KDF | Argon2id | Ganador del Password Hashing Competition, resistente a GPUs |
| Cipher | AES-256-GCM | Cifrado autenticado, detecta manipulaciones |
| Generador de claves | `crypto/rand` | Aleatoriedad criptográficamente segura del sistema operativo |

---

## Arquitectura en Capas

```
main.go
    └── cmd/              ← Comandos Cobra + adaptadores de entrada/salida
         ├── adapters/    ← Puente entre CLI y lógica de negocio
         ├── collectors/  ← Recolección de inputs del usuario (prompts)
         └── formatters/  ← Presentación de outputs en terminal

funcs/                    ← Casos de uso (lógica de negocio pura)
                            CreateSecretFunc, ObtainSecretFunc, etc.

secrets/                  ← Dominio: structs, interfaces, errores, tipos
shared/
    ├── cryptors/         ← AESGCMEncryptor (Argon2id + AES-256-GCM)
    ├── storage/          ← SQLite: Open, Close, Migrate, SecretRepository
    ├── memory/           ← Limpieza de bytes sensibles en RAM
    └── passgen/          ← Generador de contraseñas criptográficamente seguro
```

Las capas superiores dependen de las inferiores. `funcs/` solo conoce interfaces de `secrets/` — nunca importa `storage/` ni `cryptors/` directamente. Eso está conectado en `cmd/adapters/bootstrap.go`.

---

## Estado del Proyecto

**Fase 1 — MVP (implementado):**
- Tipos de secretos: `account` y `finance`.
- Comandos: `init`, `add`, `get`, `list`, `rm`.
- Lógica de exportación / importación implementada a nivel de `funcs/` (pendiente de exponer como comandos CLI).
- Generador de contraseñas seguras implementado en `shared/passgen/` (pendiente de exponer como comando CLI).

**Fase 2 — Planeado:**
- Comandos `export` e `import` para backup cifrado portátil.
- Comando `gen` para generar contraseñas desde la terminal.
- Nuevos tipos de secretos (notas seguras, claves SSH) usando el mismo modelo polimórfico ya construido.
- Integración con clipboard.
