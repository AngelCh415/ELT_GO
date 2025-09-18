# System Design — Admira ETL (Go)


## Idempotencia & Reprocesamiento
- Se mantiene un `seen` en memoria por fuente (`ads|date|campaign|channel`, `crm|opportunity_id`).
- Reprocesar un rango (`since`) no duplica. Para reprocesos limpios, podría añadirse un `version key` por ventana para invalidar/agregar.


## Particionamiento & Retención
- Partición **por día** y **UTM triple** (+ canal/campaign). En almacenamiento durable usaríamos partición por `date` y clustering por `channel,campaign_id,utm_*`.
- Retención configurable por políticas (p.ej., 18 meses en caliente, histórico en lake).


## Concurrencia & Throughput
- El ETL actual es secuencial (suficiente para la prueba). Escalable con **worker pool** para parseo/loteo y con `http.Client` tunado (Keep‑Alive). Métricas y agregación son O(n).


## Calidad de datos (UTMs ausentes & fallbacks)
- Normalización: `strings.TrimSpace`, `unknown` para UTM faltantes, truncado de fechas al día.
- Valores negativos se recortan a 0. División protegida (no NaN/Inf).


## Observabilidad
- Logging JSON con `request_id`, latencias y contadores de ingesta. Añadir `/metrics` (Prometheus) para `ingest_requests_total`, `ingest_failures_total`, `latency_histogram`.


## Evolución en el ecosistema Admira
- **Contratos**: versionar payloads (OpenAPI/JSON Schema).
- **Data Lake**: escribir brutos a S3/GCS y derivar tablas en BigQuery/Snowflake.
- **CDC/Upserts**: usar claves naturales (campaign_id/opportunity_id) + `ingested_at`.