-- ==========================================
-- PASO 2.1: Esquema Normalizado BCNF
-- ==========================================
BEGIN;

-- 1. Tabla de Catálogo de Planes (Dependencia directa de la llave primaria)
CREATE TABLE catalogo_planes_bcnf (
    plan_id INT PRIMARY KEY,
    nombre_plan VARCHAR(100) NOT NULL,
    costo_base_centavos BIGINT NOT NULL CHECK (costo_base_centavos >= 0)
);

-- 2. Tabla de Abonados (Datos del cliente, evita dependencias transitivas)
CREATE TABLE abonados_bcnf (
    abonado_id BIGSERIAL PRIMARY KEY,
    msisdn VARCHAR(15) UNIQUE NOT NULL,
    nombre_completo VARCHAR(150) NOT NULL,
    estado VARCHAR(20) NOT NULL DEFAULT 'ACTIVO'
);

-- 3. Tabla de Transacciones BSS (Tabla central)
-- Incluye columna semiestructurada JSONB NOT NULL y llaves foráneas
CREATE TABLE transacciones_bss (
    transaccion_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    abonado_id BIGINT NOT NULL REFERENCES abonados_bcnf(abonado_id),
    plan_id INT NOT NULL REFERENCES catalogo_planes_bcnf(plan_id),
    saldo_centavos BIGINT NOT NULL,
    metadata_jsonb JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. Tabla Outbox (Para el Ejercicio 4 - Patrón Outbox para microservicios)
-- Garantiza que los eventos de dominio se guarden en la misma transacción ACID
CREATE TABLE outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id VARCHAR(50) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed BOOLEAN NOT NULL DEFAULT FALSE
);

COMMIT;
