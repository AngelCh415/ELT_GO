# Admira ETL (Go)

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
- **Nota**: aunque la prueba pedía Mocky, se utilizó **MockAPI** para exponer los endpoints de prueba (más estable).


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

---
## 📤 Ejemplos de llamadas

Puedes probar los endpoints de dos formas:

### 🔹 1) Colección de Postman
Importa el archivo [`Admira-ETL.postman_collection.json`](Admira-ETL.postman_collection.json) en Postman.  
Incluye ejemplos listos para:
- Ingestar datos (`POST /ingest/run`)
- Métricas por canal (`GET /metrics/channel`)
- Funnel por UTM (`GET /metrics/funnel`)

La colección usa la variable `{{baseUrl}}` con valor por defecto `http://localhost:8080`.

---

### 🔹 2) Archivo `requests.http` (VSCode/IntelliJ)
Si usas **VSCode** o **IntelliJ** con la extensión *REST Client*, puedes ejecutar los requests desde el archivo  
[`examples/requests.http`](examples/requests.http).  

Ejemplo (legible en GitHub):

```http
### Ingestar datos desde una fecha
POST http://localhost:8080/ingest/run?since=2025-08-01

### Métricas por canal
GET http://localhost:8080/metrics/channel?from=2025-08-01&to=2025-08-31&channel=google_ads

### Funnel por UTM
GET http://localhost:8080/metrics/funnel?from=2025-08-01&to=2025-08-31&utm_campaign=back_to_school

```



## 📤 Ejemplos de llamadas y respuestas

### 1) Ingestar datos
**Request**
```bash
POST http://localhost:8080/ingest/run?since=2025-08-01

Response:

{
  "status": "ingest started"
}
```

```bash
2) Métricas por canal

Request

GET http://localhost:8080/metrics/channel?from=2025-08-01&to=2025-08-31&channel=google_ads&limit=10&offset=0


Response (ejemplo)

[
  {
    "date": "2025-08-01",
    "channel": "google_ads",
    "campaign_id": "C-1001",
    "clicks": 1200,
    "impressions": 45000,
    "cost": 350.75,
    "leads": 25,
    "opportunities": 8,
    "closed_won": 3,
    "revenue": 5000,
    "cpc": 0.292,
    "cpa": 14.03,
    "cvr_lead_to_opp": 0.32,
    "cvr_opp_to_won": 0.375,
    "roas": 14.25
  }
]
```

```bash

3) Funnel por UTM

Request

GET http://localhost:8080/metrics/funnel?from=2025-08-01&to=2025-08-31&utm_campaign=back_to_school


Response (ejemplo)

{
  "utm_campaign": "back_to_school",
  "leads": 25,
  "opportunities": 8,
  "closed_won": 3,
  "revenue": 5000,
  "cvr_lead_to_opp": 0.32,
  "cvr_opp_to_won": 0.375,
  "roas": 14.25
}
```

```bash
4) Exportación (opcional)

Request

POST http://localhost:8080/export/run?date=2025-08-01


Response (ejemplo)

{
  "date": "2025-08-01",
  "channel": "google_ads",
  "campaign_id": "C-1001",
  "clicks": 1200,
  "impressions": 45000,
  "cost": 350.75,
  "leads": 25,
  "opportunities": 8,
  "closed_won": 3,
  "revenue": 5000,
  "cpc": 0.292,
  "cpa": 14.03,
  "cvr_lead_to_opp": 0.32,
  "cvr_opp_to_won": 0.375,
  "roas": 14.25
}
```

