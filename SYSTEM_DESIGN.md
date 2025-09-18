# System Design — Admira ETL (Go)

## Idempotencia & Reprocesamiento
- Se mantiene un `seen` en memoria por fuente (`ads|date|campaign|channel`, `crm|opportunity_id`).  
- Reprocesar un rango (`since`) no duplica registros.  
- Para reprocesos limpios a futuro, se podría añadir un `version key` por ventana para invalidar/agregar.  

---

## Particionamiento & Retención
- Partición **por día** y **triple UTM** (`utm_campaign`, `utm_source`, `utm_medium`) más canal/campaign.  
- En almacenamiento durable se usaría partición por `date` y clustering por `channel,campaign_id,utm_*`.  
- Retención configurable por políticas (ej: 18 meses en caliente, histórico en data lake).  

---

## Concurrencia & Throughput
- El ETL actual es **secuencial** (suficiente para la prueba).  
- Escalable con **worker pools** para parseo/lotes y un `http.Client` ajustado (Keep-Alive, límites de conexiones).  
- La agregación es **O(n)** sobre los registros.  

---

## Calidad de datos (UTMs ausentes & fallbacks)
- Normalización: `strings.TrimSpace`, `strings.ToLower`, `unknown` para UTMs faltantes.  
- Fechas truncadas al día (`YYYY-MM-DD`).  
- Valores negativos (clicks, impresiones, costos) se recortan a 0.  
- División protegida para métricas (evita `NaN`/`Inf`).  
- CRM se cruza con Ads por **día + UTM triple**. Si no hay Ads, CRM cae en una clave “vacía” (`channel=""`).  

---

## Observabilidad
- **Logging estructurado JSON** con `request_id` y latencias por request.  
- Endpoints `GET /healthz` y `GET /readyz`.  
- Manejo de errores de red: timeouts, retries con backoff, captura de 4xx/5xx.  
- Métricas Prometheus opcionales:  
  - `ingest_requests_total`  
  - `ingest_failures_total`  
  - `ingest_latency_seconds` (histograma)  

---

## Evolución en el ecosistema Admira
- **Contratos:** versionar payloads con OpenAPI/JSON Schema.  
- **Data Lake:** persistir datos crudos en S3/GCS y derivar métricas en BigQuery/Snowflake.  
- **CDC/Upserts:** usar claves naturales (`campaign_id`, `opportunity_id`) junto con `ingested_at`.  
- **Persistencia real:** reemplazar `MemoryStore` por SQL/OLAP con particiones y retención.  
