package main

import (
	"fmt"
	"time"
)

func main() {
	// 1. Demostración de Value Object Inmutable
	fmt.Println("=== 1. Prueba de Value Object Dinero ===")
	saldo := Dinero{MontoCentavos: 10000, Moneda: "USD"}
	cargo := Dinero{MontoCentavos: 2000, Moneda: "USD"}
	nuevoSaldo, _ := saldo.Restar(cargo)
	fmt.Printf("Saldo original: %v | Saldo restado: %v | Saldo original intacto: %v\n", 
		saldo.MontoCentavos, nuevoSaldo.MontoCentavos, saldo.MontoCentavos)
	
	_, errNegativo := saldo.Restar(Dinero{MontoCentavos: 20000, Moneda: "USD"})
	fmt.Printf("Intento de resta negativa bloqueado: %v\n\n", errNegativo)

	// 2. Demostración SAGA Exitosa
	suscripcionExito := NuevaSuscripcion("SUB-001", 10000)
	cbExito := NewCircuitBreaker(3, 2*time.Second)
	sagaExito := &SagaOrchestrator{suscripcion: suscripcionExito, cb: cbExito}
	sagaExito.EjecutarSaga(3000)
	fmt.Printf("Saldo final tras SAGA exitosa: %d centavos\n\n", suscripcionExito.Saldo.MontoCentavos)

	// 3. Demostración SAGA Fallida (Compensación)
	suscripcionFallo := NuevaSuscripcion("SUB-002", 10000)
	cbFallo := NewCircuitBreaker(3, 2*time.Second)
	sagaFallo := &SagaOrchestrator{suscripcion: suscripcionFallo, cb: cbFallo}
	sagaFallo.EjecutarSaga(8000) // Forzará fallo en T2
	fmt.Printf("Saldo final tras SAGA fallida (debe volver a 10000): %d centavos\n\n", suscripcionFallo.Saldo.MontoCentavos)

	// 4. Demostración Circuit Breaker y Backoff
	fmt.Println("=== 4. Prueba de Circuit Breaker y Backoff ===")
	for i := 1; i <= 5; i++ {
		// Simulamos una llamada a red que siempre falla
		err := cbFallo.Execute(func() error {
			return fmt.Errorf("error de red simulado")
		})
		
		if err != nil {
			fmt.Printf("Intento %d: %v\n", i, err)
			if i <= 3 { // Aplicar backoff solo en los primeros intentos
				BackoffFullJitter(i)
			}
		}
	}
}
