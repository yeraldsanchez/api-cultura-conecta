# Guía de despliegue

Este documento cubre la configuración completa del flujo CI/CD: desde cero hasta tener la API desplegada automáticamente en Render cada vez que se hace push a `master`.

## Visión general del flujo

```
push a master
      │
      ▼
[GitHub Actions] Correr tests
      │ falla → pipeline se detiene
      ▼
[GitHub Actions] Build imagen Docker → push a GHCR
      │
      ▼
[GitHub Actions] Llamar deploy hook de Render
      │
      ▼
[Render] Pre-deploy: correr migraciones
      │
      ▼
[Render] Nueva versión live
```

---

## Requisitos previos

- Cuenta en [GitHub](https://github.com)
- Cuenta en [Render](https://render.com) (free, no requiere tarjeta)
- Git instalado localmente
- El código debe estar en un repositorio de GitHub

---

## Paso 1 — Subir el código a GitHub

Si aún no tienes el repositorio en GitHub:

1. Ve a [github.com/new](https://github.com/new) y crea un repositorio vacío llamado `api-cultura-conecta`
2. En la terminal, desde la raíz del proyecto:

```bash
git init
git add .
git commit -m "first commit"
git branch -M master
git remote add origin https://github.com/TU_USUARIO/api-cultura-conecta.git
git push -u origin master
```

Reemplaza `TU_USUARIO` con tu nombre de usuario de GitHub.

---

## Paso 2 — Habilitar GitHub Actions

GitHub Actions está habilitado por defecto en todos los repositorios. El archivo del pipeline ya existe en el proyecto:

```
.github/
└── workflows/
    └── deploy.yml
```

Este archivo define dos jobs que corren en secuencia cuando hay un push a `master`:

- **test**: descarga el código, instala Go y corre `go test ./...`
- **deploy**: si los tests pasan, construye la imagen Docker, la sube a GHCR y llama al deploy hook de Render

Puedes ver el contenido completo del workflow en [`.github/workflows/deploy.yml`](.github/workflows/deploy.yml).

### Verificar que Actions está activo

1. Ve a tu repositorio en GitHub
2. Click en la pestaña **Actions**
3. Si ves el mensaje *"Get started with GitHub Actions"*, click en **I understand my workflows, go ahead and enable them**

Con el primer push que hagas, el workflow aparecerá automáticamente en esa pestaña.

---

## Paso 3 — Crear la base de datos en Render

1. Inicia sesión en [render.com](https://render.com)
2. Click en **New** → **PostgreSQL**
3. Completa el formulario:
   - **Name**: `cultura-conecta-db`
   - **Region**: Oregon (US West) — recomendado para free tier
   - **Plan**: Free
4. Click en **Create Database**
5. Espera a que la BD termine de crearse (tarda ~1 minuto)
6. Ve a la sección **Connections** y copia el valor de **Internal Database URL**

> La base de datos gratuita de Render expira a los 90 días. Para renovarla, elimínala y crea una nueva; las migraciones se correrán automáticamente en el siguiente despliegue.

---

## Paso 4 — Crear el web service en Render

1. Click en **New** → **Web Service**
2. Selecciona **Deploy an existing image registry**
3. En **Image URL** escribe:
   ```
   ghcr.io/TU_USUARIO/api-cultura-conecta:latest
   ```
   Reemplaza `TU_USUARIO` con tu nombre de usuario de GitHub en minúsculas.
4. Click en **Connect**

### Credenciales del registry

La imagen en GHCR es privada por defecto. Render necesita un token de GitHub para descargarla.

Genera el token:
1. En GitHub, ve a tu foto de perfil → **Settings** → **Developer settings** → **Personal access tokens** → **Tokens (classic)**
2. Click en **Generate new token (classic)**
3. En **Note** escribe `render-ghcr-read`
4. En **Expiration** elige `No expiration`
5. Marca únicamente el scope `read:packages`
6. Click en **Generate token** y copia el valor (solo se muestra una vez)

Configura las credenciales en Render:
- **Registry**: `ghcr.io`
- **Username**: tu usuario de GitHub
- **Password**: el token que acabas de generar

### Configuración del servicio

- **Name**: `api-cultura-conecta`
- **Region**: Oregon (US West)
- **Plan**: Free

### Variables de entorno

En la sección **Environment Variables**, agrega estas tres variables:

| Key | Value |
|---|---|
| `DATABASE_URL` | El Internal Database URL copiado en el paso 3 |
| `JWT_SECRET` | Una cadena aleatoria segura (mínimo 32 caracteres) |
| `PORT` | `8080` |

Para generar el `JWT_SECRET` puedes correr en la terminal:
```bash
openssl rand -hex 32
```

### Pre-deploy command (migraciones)

En la sección **Deploy** → **Pre-deploy Command** escribe:
```
/migrate up
```

Esto corre las migraciones automáticamente antes de que la nueva versión reciba tráfico. Si las migraciones fallan, Render cancela el despliegue y la versión anterior sigue activa.

5. Click en **Create Web Service**

### Obtener el deploy hook

Una vez creado el servicio:
1. Ve a **Settings** (dentro del servicio) → busca la sección **Deploy Hook**
2. Copia la URL completa, tiene esta forma:
   ```
   https://api.render.com/deploy/srv-xxxxxxxxxxxx?key=yyyyyyyyyyyyyyy
   ```

---

## Paso 5 — Configurar el secret en GitHub

El workflow necesita la URL del deploy hook para poder llamar a Render al final del pipeline.

1. Ve a tu repositorio en GitHub
2. Click en **Settings** → **Secrets and variables** → **Actions**
3. Click en **New repository secret**
4. Agrega:
   - **Name**: `RENDER_DEPLOY_HOOK_URL`
   - **Secret**: la URL del deploy hook copiada en el paso 4
5. Click en **Add secret**

> El token para publicar la imagen en GHCR (`GITHUB_TOKEN`) es generado automáticamente por GitHub en cada ejecución del workflow; no necesitas configurarlo.

---

## Paso 6 — Hacer la imagen pública en GHCR (recomendado)

El primer push creará el paquete en GHCR. Para que Render pueda descargarlo sin credenciales adicionales, conviene hacerlo público:

1. Después del primer push exitoso, ve a tu perfil de GitHub → **Packages**
2. Busca el paquete `api-cultura-conecta`
3. Click en **Package settings** → **Change visibility** → **Public** → confirma

Con la imagen pública puedes eliminar las credenciales del registry que configuraste en Render (ya no son necesarias).

---

## Paso 7 — Primer despliegue

Con todo configurado, activa el pipeline haciendo un push a `master`:

```bash
git add .
git commit -m "configurar CI/CD"
git push origin master
```

### Seguir el progreso

**En GitHub Actions:**
1. Ve a la pestaña **Actions** del repositorio
2. Verás el workflow `Deploy` en ejecución
3. Click en él para ver los logs de cada step en tiempo real
4. El job `test` debe terminar en verde antes de que `deploy` arranque

**En Render:**
1. Ve al web service en el dashboard de Render
2. Click en **Events** para ver el estado del despliegue
3. Verás el pre-deploy command corriendo las migraciones, luego el contenedor levantando

La API quedará disponible en:
```
https://api-cultura-conecta.onrender.com
```
(el nombre exacto aparece en el dashboard de Render)

---

## Qué pasa en cada push a master

1. GitHub Actions detecta el push y lanza el job `test`
2. Si algún test falla, el pipeline se detiene ahí — Render no recibe ninguna llamada y la versión anterior sigue activa
3. Si todos los tests pasan, se construye la imagen Docker y se sube a GHCR con dos tags: `latest` y el SHA del commit
4. Se llama al deploy hook de Render
5. Render descarga la imagen `:latest`, corre el pre-deploy command (`/migrate up`) y levanta el nuevo contenedor
6. Si el pre-deploy falla (por ejemplo una migración con error), Render cancela y mantiene la versión anterior

---

## Comportamiento del free tier de Render

El servicio gratuito **se duerme tras 15 minutos de inactividad**. La primera petición después de ese período tarda ~30 segundos mientras el contenedor se reactiva. Las peticiones siguientes responden con normalidad.

---

## Desarrollo local

Para correr la API localmente con Docker Compose:

```bash
cp .env.example .env   # completa tus variables locales
docker compose up --build
```

Para correr solo las migraciones localmente:

```bash
go run ./cmd/migrate up
```

Para correr los tests:

```bash
go test ./...
```

---

## Referencia rápida

| Qué | Dónde |
|---|---|
| Pipeline CI/CD | `.github/workflows/deploy.yml` |
| Imagen Docker | `Dockerfile` |
| Migraciones | `migrations/` |
| Variables de entorno | Render → Web Service → Environment |
| Deploy hook | Render → Web Service → Settings |
| Secret del pipeline | GitHub → Settings → Secrets → `RENDER_DEPLOY_HOOK_URL` |
| Imagen publicada | `ghcr.io/TU_USUARIO/api-cultura-conecta` |
