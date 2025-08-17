# Go Weather Service (NWS)

A small SRP-compliant Go service that calls the **National Weather Service (NWS) API** and returns:

* **Today’s short forecast** (e.g., “Partly Cloudy”)
* A **temperature category**: `"hot"`, `"cold"`, or `"moderate"` based on configurable thresholds

It exposes one HTTP GET endpoint: `GET /weather?lat=<lat>&lon=<lon>`.

---

## How it works (functional overview)

1. **HTTP handler** (`internal/httpapi`): parses `lat`/`lon`, orchestrates services, returns JSON.
2. **NWS client** (`internal/weather`):

    * Calls `GET /points/{lat},{lon}` to resolve the correct gridpoint forecast URL.
    * Calls that **forecast** URL to get forecast periods (includes `shortForecast`, `temperature`, `startTime`).
3. **Short forecast service**: selects the first forecast period that **starts today** (local time) and returns its `shortForecast`.
4. **Temperature classifier**: converts a Fahrenheit temperature into `"hot"`, `"cold"`, or `"moderate"` using thresholds.

Everything adheres to **Single Responsibility Principle**:

* HTTP concerns in one place,
* external API fetching in one place,
* business rules in small, focused services.

---

## API

### `GET /weather?lat=<float>&lon=<float>`

**Query params**

* `lat` – latitude (float)
* `lon` – longitude (float)

**Response**

```json
{
  "shortForecast": "Partly Cloudy",
  "tempCategory": "moderate"
}
```

**Errors**

* `400` invalid/missing `lat` or `lon`
* `502` upstream or processing error (NWS unreachable/invalid response, etc.)

---

## Build & Run

### Prereqs

* Go ≥ 1.20
* Internet access (calls `https://api.weather.gov`)
* **Required** env var: `NWS_USER_AGENT` (NWS requires an identifying User-Agent with contact info)

### Run locally (no container)

```bash
export NWS_USER_AGENT="go-weather-service/1.0 (your.email@example.com)"
# optional thresholds
export COLD_LT_F=50
export HOT_GE_F=80

go run ./cmd/server
# logs: listening on :8080

curl "http://localhost:8080/weather?lat=38.8894&lon=-77.0352"
```

### Build binary

```bash
go build -o weather-service ./cmd/server
./weather-service
```

### Run with Docker

1. Build:

```bash
docker build -t go-weather-service:dev .
```

2. Run:

```bash
docker run --rm -p 8080:8080 \
  -e NWS_USER_AGENT="go-weather-service/1.0 (your.email@example.com)" \
  -e COLD_LT_F=50 -e HOT_GE_F=80 \
  go-weather-service:dev
```

3. Call:

```bash
curl "http://localhost:8080/weather?lat=38.8894&lon=-77.0352"
```

### Run with Docker Compose

```bash
docker compose up --build
# then:
curl "http://localhost:8080/weather?lat=38.8894&lon=-77.0352"
```

---

## Running The Service With the Makefile
* Run server locally with envs:

  ```bash
  make run NWS_USER_AGENT="go-weather-service/1.0 (user@company.com)"
  ```

* Build binary:

  ```bash
  make build
  ```

* Build & run Docker:

  ```bash
  make docker-build
  make docker-run NWS_USER_AGENT="go-weather-service/1.0 (you@company.com)"
  ```

* Compose:

  ```bash
  make compose-up
  # ...
  make compose-down
  ```
## Configuration (ENV variables)

| Name             | Required | Default | Example                                   | Description                                                      |
|------------------|----------|---------|-------------------------------------------|------------------------------------------------------------------|
| `NWS_USER_AGENT` | **Yes**  | —       | `go-weather-service/1.0 (ops@yourco.com)` | **Required by NWS**. Must identify your app + contact email/URL. |
| `COLD_LT_F`      | No       | `50`    | `45`                                      | Temp **strictly less than** this is **cold**.                    |
| `HOT_GE_F`       | No       | `80`    | `85`                                      | Temp **greater than or equal to** this is **hot**.               |

**Notes**

* Anything between `COLD_LT_F` and `HOT_GE_F` is **moderate**.
* If you’re behind a proxy or need custom DNS, configure Docker/host accordingly.

---

## Project structure

```
go-weather/
├─ cmd/
│  └─ server/           # main.go (wires HTTP server + services)
├─ internal/
│  ├─ httpapi/          # HTTP handler (no business logic)
│  ├─ util/             # time helpers (e.g., same-day checks)
│  └─ weather/
│     ├─ ports.go       # interfaces (ForecastClient, ShortForecastService, TemperatureClassifier)
│     ├─ services.go    # business services (short forecast selection, temperature classifier)
│     └─ nwsclient.go   # NWS HTTP client (points -> forecast)
├─ Dockerfile
├─ Makefile
├─ docker-compose.yml
└─ README.md
```

---

## Development

### Run tests (if/when you add them)

```bash
go test ./...
```

### Common checks

* `go vet ./...`
* `gofmt -s -w .` or `gofumpt -w .` (if you use `gofumpt`)

---

## Troubleshooting

* **`NWS_USER_AGENT env var is required`**
  Set `NWS_USER_AGENT` with an identifier and contact. NWS rejects anonymous clients.

* **`failed to get short forecast: no forecast URL returned by points endpoint`**
  Your client’s `Accept` header is wrong. Ensure it uses `Accept: application/geo+json` (or omit Accept).
  We intentionally prefer **GeoJSON**; the `points` response must include `properties.forecast`.

* **`502` from our service**
  Usually the upstream NWS failed or returned unexpected data. Check container logs.

* **Docker build fails on `COPY go.sum`**
  Run `go mod tidy` to generate `go.sum`, or update Dockerfile to copy only `go.mod` before `go mod download`.

---

## Possible optimizations / scalability roadmap

**Networking & HTTP**

* **Connection reuse + timeouts**: already using a single `http.Client` with timeouts; consider tuning `Transport` (idle conns, per-host limits).
* **Compression**: enable `Accept-Encoding: gzip` and decode to reduce payload size.
* **Client retry policy**: add bounded retries with backoff for transient 5xx/429 responses from NWS.

**Caching**

* **Response caching**: cache the **forecast URL** per `(lat,lon)` pair and the last **forecast JSON** for a short TTL (e.g., 60–300s).
  This can cut request rate to NWS dramatically and improve latency.
* **ETag / If-None-Match**: NWS supports conditional requests; leverage ETags to save bandwidth.

**Throughput & parallelism**

* **Worker pool** for batch queries (if you add bulk endpoints).
* **Rate limiting** (token bucket) to protect NWS and yourself from spikes.

**Observability**

* **Structured logging** with request IDs and upstream timing.
* **Metrics** (Prometheus/OpenTelemetry): request counts, latency histograms, error rates, cache hit ratio.
* **Health/ready endpoints**: add `/healthz` (no NWS call) and `/readyz` (optionally checks NWS).

**Config & DX**

* Feature flags for thresholds and cache behavior.
* Add a **Makefile** with `make run`, `make build`, `make docker`, `make test`.
* Add CI (lint, vet, test, build, container scan).

**Functionality**

* Add a **/forecast** endpoint returning more periods (e.g., next 24h summary).
* Support **units** switching (Fahrenheit/Celsius).
* Add **graceful shutdown** with context cancellation and timeouts around the HTTP server.

---

## Edge cases & hardening ideas

**Forecast selection**

* **“Today” ambiguity**: The first period starting “today” may be **“Tonight”** if you call in the evening. Decide whether you want the *daytime* period instead (e.g., prefer `"Today"` / `"This Afternoon"` label). Add a strategy: “first period with daylight” vs. “first period in same local date.”
* **Missing/parse-error on `startTime`**: If `startTime` fails to parse, we skip it. Add a fallback strategy or log at warn level.

**Data quality**

* `shortForecast` may be empty or generic; decide if you want to fallback to `detailedForecast`.
* Temperature may be missing for a given period; pick the next period with a valid temperature.

**Inputs**

* Validate `lat`/`lon` bounds (`-90..90`, `-180..180`); return `400` early.
* Reject obviously bad precision (e.g., strings or NaN) — already covered by parse, but add unit tests.

**Upstream failures**

* Handle `429` (rate limit) distinctly: set `Retry-After` to the client, include backoff in our client.
* Circuit breaker for repeated NWS failures to avoid thundering herd loops.

**Security / robustness**

* Limit request body size (even though it’s GET, still good practice).
* Validate headers; set `Content-Type: application/json` consistently.
* Run as non-root user in container (already done).
* Pin base image versions and enable image scanning in CI.

**Time zones**

* We use `time.Local`. For multi-region users, consider allowing `?tz=` to evaluate “today” in a specific IANA TZ.

---

## License

TBD (MIT/Apache-2.0 recommended).
