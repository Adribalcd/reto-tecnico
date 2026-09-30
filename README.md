# Reto técnico: factorización QR y estadísticas

Dos API RESTful que se comunican por HTTP. La primera, escrita en **Go** con
**Fiber**, recibe una matriz rectangular y devuelve su factorización QR. La
segunda, escrita en **Node.js** con **Express**, recibe las matrices resultantes
y calcula estadísticas sobre sus valores. Un frontend estático consume ambas.

```
┌────────────┐        POST /go/api/v1/qr             ┌──────────────┐
│  Frontend  │ ──────────────────────────────────────▶│   go-api     │
│  (nginx)   │◀──── Q, R y estadísticas ───────────── │  :8080       │
└─────┬──────┘                                        └──────┬───────┘
      │  POST /node/auth/login                              │ POST /api/v1/statistics
      │  (obtiene el JWT)                                    │ (reenvía el mismo JWT)
      ▼                                                      ▼
┌────────────────────────────────────────────────────────────────────┐
│                             node-api  :3000                        │
└────────────────────────────────────────────────────────────────────┘
```

## Arquitectura

- **go-api** valida la matriz, calcula su factorización QR con reflexiones de
  Householder y envía `Q` y `R` a `node-api` por HTTP. La respuesta al frontend
  incluye la factorización y las estadísticas en una sola llamada.
- **node-api** emite los JWT (`POST /auth/login`) y calcula las estadísticas de
  las matrices que recibe (`POST /api/v1/statistics`).
- **frontend** es un sitio estático servido por nginx que además hace de proxy
  hacia las dos APIs, de modo que el navegador ve un único origen.

Las tres piezas se orquestan con Docker Compose:

```
go-api/      API en Go (Fiber); el QR vive en internal/matrix
node-api/    API en Node (Express); login y estadísticas en src
frontend/    HTML, CSS y JS sin dependencias, servidos por nginx
```

## Cómo ejecutarlo

Copia el ejemplo de entorno y levanta el stack:

```
cp .env.example .env       # en PowerShell: Copy-Item .env.example .env
docker compose up --build
```

El frontend queda en <http://localhost:8081> (usuario `admin`, clave
`admin123`), la API de Go en <http://localhost:8080> y la de Node en
<http://localhost:3000>.

### Prueba rápida con curl

```bash
# 1. Obtener el token
TOKEN=$(curl -s -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r .token)

# 2. Factorizar una matriz (go-api llama internamente a node-api)
curl -s -X POST http://localhost:8080/api/v1/qr \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"matrix":[[12,-51,4],[6,167,-68],[-4,24,-41]]}' | jq
```

### Sin Docker

Con Go 1.24+ y Node 20+, en dos terminales:

```bash
# Terminal 1
cd node-api
npm install
JWT_SECRET=dev-secret AUTH_USERS=admin:admin123 npm start

# Terminal 2
cd go-api
JWT_SECRET=dev-secret NODE_API_URL=http://localhost:3000 go run ./cmd/api
```

## Endpoints

Todas las respuestas de error comparten el mismo formato:

```json
{ "error": { "code": "invalid_matrix", "message": "la fila 1 tiene 1 valores, se esperaban 2" } }
```

- `POST /auth/login` (node-api). Recibe `{"username": "...", "password": "..."}`
  y responde `{"token": "<jwt>", "tokenType": "Bearer", "expiresIn": "1h"}`.
- `POST /api/v1/qr` (go-api, requiere `Bearer`). Recibe `{"matrix": [[...]]}` y
  devuelve `matrix`, `q`, `r` y `statistics`.
- `POST /api/v1/statistics` (node-api, requiere `Bearer`). Acepta un objeto de
  matrices con nombre, una lista de matrices o una sola matriz; la llamada
  interna de go-api usa la primera forma.
- `GET /health` responde sin token en ambos servicios.

```json
// POST /api/v1/qr  →  respuesta (Q y R redondeados para el ejemplo)
{
  "matrix": [[12, -51, 4], [6, 167, -68], [-4, 24, -41]],
  "q": [[-0.857143, 0.394286, -0.331429], ...],
  "r": [[-14, -21, 14], [0, -175, 70], [0, 0, -35]],
  "statistics": {
    "count": 18,
    "sum": -93.92,
    "average": -5.2177,
    "min": -175,
    "max": 70,
    "diagonal": { "any": false, "matrices": { "q": false, "r": false } }
  }
}
```

Códigos de error de go-api: `invalid_body`, `invalid_matrix`, `matrix_too_large`,
`unauthorized`, `statistics_unavailable`, `internal_error`.

## Pruebas

```bash
make test          # ambas suites
make test-go       # go test ./...
make test-node     # node --test
```

En Go se prueba la factorización (que `Q` sea ortogonal, `R` triangular y que
`Q*R` reconstruya la matriz, incluidos los casos rectangulares y de rango
deficiente), la validación de matrices, el JWT, el cliente HTTP y los handlers.
En Node se cubren las formas de entrada, el cálculo de estadísticas, la detección
de diagonales y los endpoints con `node:test` y `supertest`.

## Decisiones y notas

El enunciado menciona "rotación de la matriz" al describir la arquitectura y
"factorización QR" en la funcionalidad. Me quedé con la QR: es la operación que
se pide de forma explícita y la que le da sentido a las dos matrices que consume
la segunda API. Devuelvo la forma reducida, con `Q` de `m x k` y `R` de `k x n`
donde `k = min(m, n)`.

Para el cálculo elegí reflexiones de Householder en lugar de Gram-Schmidt.
Cuestan el mismo orden, pero se comportan mejor numéricamente. Trabajo con
`float64` y limpio el ruido de redondeo por debajo de `1e-12` para que la salida
no quede llena de colas decimales.

La segunda API acepta varias formas de matrices para poder probarla de forma
aislada sin acoplarla al formato exacto de la primera. Las dos usan el mismo
envoltorio de error, `{"error": {"code", "message"}}`, que además de simplificar
el cliente permite que go-api reenvíe con contexto los errores de node-api.

node-api emite los tokens y ambas APIs los validan contra el mismo secreto. Fijo
el algoritmo en `HS256` y valido emisor y expiración; go-api reenvía el
`Authorization` que recibe, así node-api ve la misma identidad. Todas las rutas
bajo `/api/v1` exigen token y solo quedan abiertos `/health` y el login.

## Limitaciones

- Las credenciales de demostración viven en `AUTH_USERS` y se comparan en texto
  plano. En producción iría un almacén de usuarios con claves hasheadas.
- No hay persistencia: todo se calcula en memoria, sin base de datos.
- El frontend guarda el token en memoria, así que se pierde al recargar.
- No hay CI.

## Despliegue

El stack se levanta en una VM con Docker. `docker-compose.prod.yml` expone solo
el frontend en el puerto 80 y deja las APIs en la red interna:

```bash
git clone https://github.com/Adribalcd/reto-tecnico.git
cd reto-tecnico
cp .env.example .env          # edita JWT_SECRET
docker compose -f docker-compose.prod.yml up -d --build
```

Sirve cualquier VM con Docker y capa gratuita (una e2-micro de GCP, por ejemplo);
si le pones un dominio propio, conviene meter HTTPS por delante. Para un PaaS que
separa los servicios, hay que apuntar el frontend a las URLs de cada API y
agregar CORS en la de Go.

