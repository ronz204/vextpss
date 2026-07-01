# Vext — Vista General

## ¿Qué es Vext?

Vext es un gestor de contraseñas CLI, local y minimalista, escrito en Go.

Nació de una frustración concreta: usar herramientas como Notion para guardar credenciales es incómodo, inseguro, y sencillamente no es para lo que fueron diseñadas. Vext es la alternativa directa a eso — una herramienta rápida, simple y cifrada, que vive completamente en tu máquina.

Sin nube. Sin sincronización. Sin cuentas. Solo un binario y una base de datos local cifrada.

---

## Filosofía

**Local primero.** Tus datos nunca salen de tu máquina. No hay servidores, no hay APIs, no hay dependencias externas en runtime. Lo que guardás en Vext se queda en Vext.

**Simple por diseño.** Vext hace una sola cosa: guardar y recuperar secretos de forma segura. Cada decisión de diseño favorece la claridad y la corrección por encima de la proliferación de features. Si una funcionalidad no sirve directamente a ese propósito, no entra.

**Seguridad sin compromisos.** Que la herramienta sea simple no significa que la base criptográfica lo sea. Vext usa el mismo modelo de cifrado que gestores comerciales como Bitwarden o 1Password: Argon2id para derivación de clave y AES-256-GCM para cifrado autenticado. Simplicidad en la superficie, rigor debajo.

**Filosofía Unix.** Cada comando hace exactamente una cosa, la hace bien, y no intenta adivinar qué querés hacer. La interfaz es predecible y componible — como debería ser cualquier herramienta de terminal.

---

## El Problema que Resuelve

| Punto de dolor | Cómo lo resuelve Vext |
|---|---|
| Notion / hojas de cálculo usadas como gestor de contraseñas | Herramienta dedicada, construida específicamente para ese propósito |
| Credenciales guardadas en texto plano | Cifrado AES-GCM en reposo, siempre |
| Herramientas dependientes de la nube | Completamente offline, sin cuenta necesaria |
| UX compleja para tareas simples | Binario único, un comando por acción |
| Modelos de datos rígidos | Payload polimórfico según el tipo de secreto |

---

## Usuario Objetivo

Vext está pensado para desarrolladores y usuarios técnicos que:

- Se sienten cómodos trabajando en la terminal.
- Quieren control total sobre dónde viven sus datos.
- No quieren depender de un SaaS para algo tan personal como sus credenciales.
- Valoran entender el modelo de seguridad de las herramientas que usan, en vez de simplemente confiar a ciegas.

---

## Tipos de Secretos

Vext usa un modelo polimórfico: todos los secretos se guardan en la misma tabla, y el *tipo* de secreto es lo que determina qué campos se recolectan y cómo se presentan.

### `account` (por defecto)

Pensado para cuentas de servicios online: usuario o email + contraseña.

```
$ vext add github
Username: bob@example.com
Password: ••••••••
Master password: ••••••••
✓ Credential "github" saved.
```

### `finance`

Pensado para tarjetas bancarias y credenciales financieras: número de tarjeta, PIN, código de seguridad, vencimiento, usuario/clave del banco, clave virtual, celular asociado.

```
$ vext add visa-debito -t finance
Card number: 4111111111111111
Card PIN: ••••
Security code: •••
...
Master password: ••••••••
✓ Credential "visa-debito" saved.
```