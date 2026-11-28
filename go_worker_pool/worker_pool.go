package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Transaccion struct {
	ID     string
	Monto  float64
	Origen string
}

type ResultadoAuditoria struct {
	TransaccionID string
	Aprobada      bool
	WorkerID      int
}

type ProcesadorTransaccion interface {
	Auditar(tx Transaccion) bool
}

type MotorReglasBasico struct {
	LimiteMaximo float64
}

func (m MotorReglasBasico) Auditar(tx Transaccion) bool {
	// Simulación ligera de validación
	time.Sleep(5 * time.Millisecond)
	return tx.Monto <= m.LimiteMaximo
}

func EjecutarWorkerPool(
	numWorkers int,
	procesador ProcesadorTransaccion,
	jobs <-chan Transaccion,
	results chan<- ResultadoAuditoria,
	procesadas *atomic.Int64,
	rechazadas *atomic.Int64,
) {
	var wg sync.WaitGroup

	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for tx := range jobs {
				aprobada := procesador.Auditar(tx)
				procesadas.Add(1)
				if !aprobada {
					rechazadas.Add(1)
				}
				results <- ResultadoAuditoria{
					TransaccionID: tx.ID,
					Aprobada:      aprobada,
					WorkerID:      workerID,
				}
			}
		}(w)
	}

	wg.Wait()
	close(results)
}

func main() {
	const W = 4
	const totalJobs = 12

	jobs := make(chan Transaccion, totalJobs)
	results := make(chan ResultadoAuditoria, totalJobs)

	var totalProcesadas atomic.Int64
	var totalRechazadas atomic.Int64

	motor := MotorReglasBasico{LimiteMaximo: 1000.0}

	// Cargar trabajos
	for i := 1; i <= totalJobs; i++ {
		jobs <- Transaccion{
			ID:     fmt.Sprintf("GO-TX-%03d", i),
			Monto:  float64(i * 250),
			Origen: "API_GATEWAY",
		}
	}
	close(jobs)

	go EjecutarWorkerPool(W, motor, jobs, results, &totalProcesadas, &totalRechazadas)

	for res := range results {
		estado := "APROBADA"
		if !res.Aprobada {
			estado = "RECHAZADA"
		}
		fmt.Printf("[Worker %d] %s -> %s\n", res.WorkerID, res.TransaccionID, estado)
	}

	fmt.Printf("\nTotal Procesadas: %d | Rechazadas: %d\n", totalProcesadas.Load(), totalRechazadas.Load())
}