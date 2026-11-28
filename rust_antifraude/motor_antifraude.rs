use std::sync::{Arc, Mutex};
use std::thread;

#[derive(Debug)]
struct AlertaFraude<'a> {
    id_tx: &'a str,
    msisdn: &'a str,
    motivo: &'a str,
    monto_centavos: u64,
}

#[derive(Debug, Default)]
struct EstadisticasRiesgo {
    total_evaluadas: usize,
    alertas_detectadas: usize,
    monto_riesgo_acumulado: u64,
}

fn main() {
    // Lote de transacciones en memoria estática/stack (Zero-Copy)
    let lote_crudo: Vec<&str> = vec![
        "TX-1001|584121112233|150000",
        "TX-1002|584149998877|4500",
        "TX-1003|584245556677|890000",
        "TX-1004|584163334455|12000",
        "TX-1005|584127778899|620000",
        "TX-1006|584141110000|3000",
        "TX-1007|584248889900|1250000",
        "TX-1008|584162221100|9500",
    ];

    let lote_compartido = Arc::new(lote_crudo);
    let stats = Arc::new(Mutex::new(EstadisticasRiesgo::default()));
    let num_hilos = 4;
    let tamano_chunk = (lote_compartido.len() + num_hilos - 1) / num_hilos;

    let mut handles = Vec::with_capacity(num_hilos);

    for hilo_id in 0..num_hilos {
        let datos = Arc::clone(&lote_compartido);
        let estado_riesgo = Arc::clone(&stats);

        let handle = thread::spawn(move || {
            let inicio = hilo_id * tamano_chunk;
            let fin = (inicio + tamano_chunk).min(datos.len());

            if inicio >= fin {
                return;
            }

            let mut alertas_locales: Vec<AlertaFraude<'_>> = Vec::new();
            let mut evaluadas_local = 0;

            for &registro in &datos[inicio..fin] {
                evaluadas_local += 1;
                let partes: Vec<&str> = registro.split('|').collect();
                if partes.len() == 3 {
                    let monto: u64 = partes[2].parse().unwrap_or(0);
                    // Regla antifraude: transacciones mayores a 500,000 centavos
                    if monto > 500_000 {
                        alertas_locales.push(AlertaFraude {
                            id_tx: partes[0],
                            msisdn: partes[1],
                            motivo: "MONTO_INUSUAL_ALTO",
                            monto_centavos: monto,
                        });
                    }
                }
            }

            let mut guard = estado_riesgo.lock().expect("Mutex envenenado");
            guard.total_evaluadas += evaluadas_local;
            guard.alertas_detectadas += alertas_locales.len();
            for alerta in &alertas_locales {
                guard.monto_riesgo_acumulado += alerta.monto_centavos;
                println!(
                    "[Hilo {}] Alerta Zero-Copy -> ID: {}, Tel: {}, Motivo: {}, Monto: {}",
                    hilo_id, alerta.id_tx, alerta.msisdn, alerta.motivo, alerta.monto_centavos
                );
            }
        });

        handles.push(handle);
    }

    for handle in handles {
        handle.join().expect("Error esperando hilo");
    }

    let resumen = stats.lock().unwrap();
    println!("\n=== Resumen Auditoría Rust ===");
    println!("{:?}", *resumen);
}