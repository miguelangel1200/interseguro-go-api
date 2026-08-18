# interseguro-go-api

API HTTP escrita en **Go** con el framework **Fiber**.

Recibe una matriz, la rota 90° en sentido horario, calcula su factorización
QR (A = Q·R) y delega el cálculo de estadísticas a un servicio externo
(`interseguro-node-api`), propagando el token JWT del usuario.

Sigue arquitectura hexagonal (dominio → aplicación → infraestructura) para
mantener la lógica de negocio independiente de los frameworks.

## Endpoints

| Método | Ruta      | Descripción                                                       | Auth |
|--------|-----------|-------------------------------------------------------------------|------|
| GET    | `/health` | Health check (usado por Cloud Run).                               | No   |
| POST   | `/process`| Recibe `{ "matrix": [...] }` → rotación + QR + estadísticas.      | JWT  |

### POST /process

- `matrix`: matriz rectangular de números de punto flotante (máx. 100×100).
- Respuesta: `message`, `result` con las matrices `original`, `rotated`,
  `Q` y `R`, y `statistics` cuando el servicio de estadísticas responde.
  Si el servicio externo no está disponible, devuelve el resultado local
  con `message: "processed locally; node api unreachable"`.

## Variables de entorno

| Variable        | Descripción                                   | Default                |
|-----------------|-----------------------------------------------|------------------------|
| `PORT`          | Puerto HTTP (Cloud Run lo inyecta).           | `8080`                 |
| `NODE_API_URL`  | URL base de la API Node.                      | `http://localhost:3000`|
| `CORS_ORIGIN`   | Origen CORS permitido (p. ej. el frontend).   | `*`                    |
| `JWT_SECRET`    | Secreto HS256 compartido con node-api.        | `interseguro-dev-secret`|
| `JWT_ISSUER`    | Issuer del JWT.                               | `interseguro`          |
| `JWT_AUDIENCE`  | Audience del JWT.                             | `interseguro-api`      |

## Ejecución local

```bash
go run ./cmd
# o junto a las demás APIs: docker compose up (ver interseguro-infra)
```

## Despliegue

- **CI (GitHub Actions):** en cada push a `main` se ejecutan los tests y se
  construye/publica la imagen en Artifact Registry (`.github/workflows/build.yml`).
- **Manual:** `./deploy.sh [PROJECT_ID] [REGION]` construye y sube la imagen
  con Cloud Build.

## Tests

```bash
make test        # go test ./...
make test-cover  # tests + cobertura agregada (objetivo >85%)
make cover-html  # reporte HTML de cobertura
make vuln        # govulncheck (escaneo de vulnerabilidades)
```
