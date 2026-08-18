# interseguro-go-api

API en Go (Fiber) del Reto Interseguro. Recibe una matriz, la rota 90° en
sentido horario, calcula su factorización QR (A = Q·R) y propaga el resultado
a `interseguro-node-api` para obtener las estadísticas.

## Endpoints

| Método | Ruta      | Descripción                                   | Auth |
|--------|-----------|-----------------------------------------------|------|
| GET    | `/health` | Health check para Cloud Run.                  | No   |
| POST   | `/process`| Procesa `{ "matrix": [...] }` y rota/QR.      | JWT  |

## Variables de entorno

| Variable        | Descripción                                   | Default               |
|-----------------|-----------------------------------------------|-----------------------|
| `PORT`          | Puerto HTTP (Cloud Run lo inyecta).           | `8080`                |
| `NODE_API_URL`  | URL base de la API Node.                      | `http://localhost:3000`|
| `CORS_ORIGIN`   | Origen CORS permitido (p. ej. el frontend).   | `*`                   |
| `JWT_SECRET`    | Secreto HS256 compartido con node-api.        | `interseguro-dev-secret`|
| `JWT_ISSUER`    | Issuer del JWT.                               | `interseguro`         |
| `JWT_AUDIENCE`  | Audience del JWT.                             | `interseguro-api`     |

Seguridad: matrices limitadas a 100x100 (`MATRIX_TOO_LARGE`), body limit 1 MiB,
cabeceras de seguridad básicas y error del servicio Node enmascarado.

## Despliegue

```bash
./deploy.sh                 # usa el proyecto de gcloud activo y us-central1
./deploy.sh MI-PROYECTO     # proyecto explícito
```

El script construye la imagen con Cloud Build y la sube a Artifact Registry:

```
{region}-docker.pkg.dev/{project}/interseguro-go-api/interseguro-go-api:latest
```

## Desarrollo local

```bash
go run ./cmd
# o junto con node-api: docker compose up (desde interseguro-infra)
```

## Tests y cobertura

```bash
make test        # go test ./...
make test-cover  # tests + cobertura agregada (>/internal/...); umbral objetivo 85%+
make cover-html  # reporte HTML de cobertura
```

Organización de los tests por capa:

| Paquete                       | Archivos de test                                     |
|-------------------------------|------------------------------------------------------|
| `internal/domain`             | `matrix_ops_test.go`, `domain_test.go`               |
| `internal/application/services`| `matrix_service_test.go` (mock del puerto de salida) |
| `internal/infrastructure/http`| `handler_test.go`, `process_test.go`, `auth_test.go`, `error_test.go` |
| `internal/infrastructure/remote`| `node_client_test.go` (con `httptest.Server`)      |

La cobertura se mide sobre `./internal/...`; `cmd/main.go` queda fuera por ser
únicamente el punto de entrada de wiring (estándar en proyectos Go).
