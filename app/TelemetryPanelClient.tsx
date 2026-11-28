"use client";

import React, { useState, useMemo } from "react";

interface CellTelemetryDTO {
  cellId: string;
  latencyMs: number;
  packetLossPct: number;
  throughputMbps: number;
}

export default function TelemetryPanelClient() {
  const [packets, setPackets] = useState<CellTelemetryDTO[]>([]);
  const [renderTimeMs, setRenderTimeMs] = useState<number | null>(null);

  const { avgLatency, criticalCount } = useMemo(() => {
    if (packets.length === 0) {
      return { avgLatency: 0, criticalCount: 0 };
    }

    const totalLatency = packets.reduce((acc, item) => acc + item.latencyMs, 0);
    const criticals = packets.filter((item) => item.latencyMs > 50).length;

    return {
      avgLatency: totalLatency / packets.length,
      criticalCount: criticals,
    };
  }, [packets]);

  const handleInjectTelemetry = () => {
    const startTime = performance.now();

    const incomingBatch: CellTelemetryDTO[] = Array.from({ length: 10 }, () => ({
      cellId: `NODE-${Math.floor(100 + Math.random() * 900)}`,
      latencyMs: Number((Math.random() * 85 + 5).toFixed(2)),
      packetLossPct: Number((Math.random() * 4).toFixed(2)),
      throughputMbps: Number((Math.random() * 150 + 50).toFixed(1)),
    }));

    setPackets((prevPackets) => [...incomingBatch, ...prevPackets].slice(0, 50));

    const endTime = performance.now();
    setRenderTimeMs(Number((endTime - startTime).toFixed(3)));
  };

  return (
    <main style={{ padding: "2rem", fontFamily: "system-ui, sans-serif" }}>
      <h1>Monitor de Telemetría 5G</h1>

      <div style={{ marginBottom: "1.5rem" }}>
        <button
          onClick={handleInjectTelemetry}
          style={{
            padding: "10px 16px",
            backgroundColor: "#0070f3",
            color: "#fff",
            border: "none",
            borderRadius: "6px",
            cursor: "pointer",
            fontWeight: "bold",
          }}
        >
          Inyectar Ráfaga de Datos
        </button>

        {renderTimeMs !== null && (
          <p style={{ marginTop: "0.5rem", color: "#555" }}>
            <strong>Tiempo de ejecución (performance.now):</strong> {renderTimeMs} ms
          </p>
        )}
      </div>

      <section style={{ display: "flex", gap: "2rem", marginBottom: "1.5rem" }}>
        <div>
          <span>Latencia Promedio: </span>
          <strong>{avgLatency.toFixed(2)} ms</strong>
        </div>
        <div>
          <span>Celdas Críticas (&gt;50ms): </span>
          <strong style={{ color: criticalCount > 0 ? "crimson" : "green" }}>
            {criticalCount}
          </strong>
        </div>
        <div>
          <span>Ventana Activa: </span>
          <strong>{packets.length} / 50 paquetes</strong>
        </div>
      </section>

      <table border={1} cellPadding={8} style={{ width: "100%", borderCollapse: "collapse" }}>
        <thead>
          <tr style={{ backgroundColor: "#f4f4f4", textAlign: "left" }}>
            <th>ID Celda</th>
            <th>Latencia (ms)</th>
            <th>Pérdida (%)</th>
            <th>Velocidad (Mbps)</th>
          </tr>
        </thead>
        <tbody>
          {packets.length === 0 ? (
            <tr>
              <td colSpan={4} style={{ textAlign: "center" }}>
                Sin paquetes recibidos. Presiona el botón para iniciar.
              </td>
            </tr>
          ) : (
            packets.map((pkt, index) => (
              <tr
                key={`${pkt.cellId}-${index}`}
                style={{
                  backgroundColor: pkt.latencyMs > 50 ? "#ffe6e6" : "transparent",
                }}
              >
                <td>{pkt.cellId}</td>
                <td>{pkt.latencyMs} ms</td>
                <td>{pkt.packetLossPct}%</td>
                <td>{pkt.throughputMbps} Mbps</td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </main>
  );
}