-- ==========================================
-- PASO 2.2: Poblado, Auditoría MVCC e Índices
-- ==========================================
BEGIN;

-- 1. Insertar catálogo y abonados base
INSERT INTO catalogo_planes_bcnf (plan_id, nombre_plan, costo_base_centavos) 
VALUES (1, 'Básico 5G', 1500), (2, 'Premium 5G', 3000);

-- Usamos generate_series para crear 50,000 abonados rápidamente
INSERT INTO abonados_bcnf (msisdn, nombre_completo)
SELECT 
    '+1234' || LPAD(i::text, 8, '0'),
    'Usuario ' || i
FROM generate_series(1, 50000) AS i;

-- Insertar 50,000 transacciones con datos JSON aleatorios
INSERT INTO transacciones_bss (abonado_id, plan_id, saldo_centavos, metadata_jsonb, created_at)
SELECT 
    i,
    CASE WHEN i % 2 = 0 THEN 1 ELSE 2 END,
    (random() * 10000)::bigint,
    jsonb_build_object(
        'antena_5g', CASE WHEN i % 3 = 0 THEN 'CCS-NORTE-01' ELSE 'CCS-SUR-02' END,
        'roaming', CASE WHEN i % 4 = 0 THEN true ELSE false END,
        'nivel_senal', (random() * 100)::int
    ),
    NOW() - (i || ' minutes')::interval
FROM generate_series(1, 50000) AS i;

COMMIT;

-- ==========================================
-- 2. Demostración de MVCC (Actualización)
-- ==========================================
-- Ejecutamos un UPDATE sobre un abonado específico
UPDATE abonados_bcnf SET estado = 'SUSPENDIDO' WHERE abonado_id = 1;

-- Consultamos las columnas ocultas del sistema (xmin, xmax)
-- xmin: ID de la transacción que insertó esta versión.
-- xmax: ID de la transacción que la borró (o 0 si está viva).
-- Nota: Usamos un SELECT normal, PostgreSQL no suele mostrar estas columnas a menos que se pidan explícitamente.
SELECT ctid, xmin, xmax, saldo_centavos 
FROM transacciones_bss 
WHERE abonado_id = 1;

-- En MVCC, el UPDATE no sobreescribe el dato en el mismo lugar. 
-- Escribe una nueva tupla (fila) al final de la tabla (un nuevo ctid) y marca la vieja con el xmax.

-- ==========================================
-- 3. Optimización con Índices (GIN y BRIN)
-- ==========================================

-- --- ANTES DE CREAR LOS ÍNDICES ---
-- Forzamos un Seq Scan (Escaneo Secuencial) desactivando el uso de índices
SET enable_seqscan = on;
SET enable_indexscan = off;
SET enable_bitmapscan = off;

EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM transacciones_bss 
WHERE metadata_jsonb @> '{"antena_5g": "CCS-NORTE-01", "roaming": true}';

-- Aquí verás un costo alto y un "Seq Scan" que lee toda la tabla (50k filas).


-- --- CREACIÓN DE ÍNDICES ---

-- Índice GIN: Ideal para buscar dentro de documentos JSONB (clave-valor)
CREATE INDEX idx_tx_metadata_gin ON transacciones_bss USING GIN (metadata_jsonb jsonb_path_ops);

-- Índice BRIN: Ideal para tablas gigantes ordenadas por fecha (como logs o telemetría).
-- Es minúsculo en tamaño comparado con un B-Tree y divide la tabla en bloques.
CREATE INDEX idx_tx_created_at_brin ON transacciones_bss USING BRIN (created_at);


-- --- DESPUÉS DE CREAR LOS ÍNDICES ---
-- Restauramos la configuración del planner para que use los índices
RESET enable_seqscan;
RESET enable_indexscan;
RESET enable_bitmapscan;

-- Volvemos a ejecutar la consulta
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM transacciones_bss 
WHERE metadata_jsonb @> '{"antena_5g": "CCS-NORTE-01", "roaming": true}';

-- Ahora deberías ver un "Bitmap Index Scan" usando idx_tx_metadata_gin.
-- El costo será drásticamente menor y el tiempo de ejecución (Execution Time) caerá.
