# ELT_GO# Admira ETL (Go)
## 🚀 Correr


```bash
cp .env.example .env
# edita ADS_API_URL / CRM_API_URL
make run
# o
docker compose up --build
```


## 🔌 Endpoints


- `POST /ingest/run?since=YYYY-MM-DD`
- `GET /metrics/channel?from=YYYY-MM-DD&to=YYYY-MM-DD&channel=google_ads&limit=50&offset=0`
- `GET /metrics/funnel?from=YYYY-MM-DD&to=YYYY-MM-DD&utm_campaign=back_to_school`
- `POST /export/run?date=YYYY-MM-DD` (opcional; requiere `SINK_URL` y `SINK_SECRET`)
- `GET /healthz`, `GET /readyz`


### Ejemplos (cURL)


```bash
# 1) Ingesta completa desde 2025-08-01
curl -XPOST "http://localhost:8080/ingest/run?since=2025-08-01"


# 2) Métricas por canal
curl "http://localhost:8080/metrics/channel?from=2025-08-01&to=2025-08-31&channel=google_ads&limit=20&offset=0"


# 3) Funnel por UTM
curl "http://localhost:8080/metrics/funnel?from=2025-08-01&to=2025-08-31&utm_campaign=back_to_school"


# 4) Exportación diaria (opcional)
curl -XPOST "http://localhost:8080/export/run?date=2025-08-01"
```


## 📐 Suposiciones


- En CRM, toda oportunidad creada cuenta como **lead**; `stage` acumula **opportunities** y **closed_won**.
- Cruce UTM: si faltan UTM en alguna fuente, se normaliza a `unknown` para mantener la agregación.
- Agregación **diaria**; la hora se trunca al día.


## 🔍 Observabilidad


- Logging estructurado JSON con `X-Request-ID` y latencias.
- (Opcional) Puedes añadir `/metrics` Prometheus fácilmente con `promhttp.Handler()`.


## 🧪 Tests


- Tests unitarios mínimos en `test/etl_transform_test.go`. Agrega casos adicionales según tus contratos de datos.


## 🧱 Limitaciones


- Almacenamiento **en memoria** (ideal para prueba). En producción: persistencia (SQL/OLAP) y esquemas.
- Unión de Ads↔CRM se hace por **triple UTM**; si hay múltiples canales para el mismo UTM en un día, se mantienen por clave (channel/campaign_id) del lado de Ads y por UTM del lado CRM.