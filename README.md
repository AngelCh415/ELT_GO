# ELT_GO# Admira ETL (Go)

## 📦 Requisitos

- Go 1.22+
- Docker + Docker Compose
- Make

## ⚙️ Configuración

Copia el archivo de ejemplo:

```bash
cp .env.example .env
```

## 🚀 Correr

```bash
# servidor local
make run

# ejecutar tests
make test

# levantar con Docker
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

- En CRM:  
  - Cada oportunidad creada cuenta como **lead**.  
  - `stage="opportunity"` acumula en **opportunities**.  
  - `stage="closed_won"` acumula tanto en **opportunities** como en **closed_won** y suma al **revenue**.  
- El cruce Ads↔CRM se hace por **día + triple UTM (`utm_campaign`, `utm_source`, `utm_medium`)**.  
  - Si existen Ads y CRM en la misma fecha/UTMs, se **unen en un solo agregado**.  
  - Si no hay match de Ads, CRM cae a una clave “vacía” (`channel=""`), para no perder leads.  
- Normalización:  
  - UTMs ausentes se transforman a `unknown`.  
  - Fechas se truncan al día (`YYYY-MM-DD`).  
  - Valores negativos de clicks, impresiones y costos se normalizan a cero.  

---

## 🧱 Limitaciones

- **Persistencia:** El almacenamiento es solo **en memoria**; al reiniciar se pierden datos. En producción → usar DB o data lake.  
- **Unión Ads↔CRM:**  
  - Se basa únicamente en **día + UTM triple**.  
  - No distingue cuando múltiples campañas comparten UTMs → posible agregación conjunta.  
- **Escalabilidad:** Procesamiento secuencial; no hay worker pools ni particionamiento implementado.  
- **Validación de datos:** Se asume que los payloads cumplen el contrato; faltan validaciones estrictas de tipos y rangos.  
- **Exportación:** `/export/run` requiere `SINK_URL` y `SINK_SECRET`; si no están configurados, responde `"sink not configured"`.  

---

## 🔍 Observabilidad

- **Healthchecks**:  
  - `GET /healthz` → confirma que el servicio está vivo.  
  - `GET /readyz` → confirma que el servicio está listo para recibir tráfico.  
- **Logging estructurado**:  
  - Salida en JSON con nivel de log y `X-Request-ID` para trazabilidad.  
  - Incluye latencias por request.  
- **Manejo de errores de red**:  
  - Timeouts configurables en cliente HTTP.  
  - Retries con backoff exponencial en ingesta.  
  - Tests unitarios cubren casos 4xx, 5xx y timeouts.  
- **Métricas (opcional)**:  
  - Puede integrarse `/metrics` para Prometheus usando `promhttp.Handler()`.  
