package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// ==========================================
// 1.2 DDD: VALUE OBJECT Y AGREGADO
// ==========================================

// Value Object Inmutable: Dinero
type Dinero struct {
	MontoCentavos int64
	Moneda        string
}

// Método Sumar: Retorna una NUEVA instancia (Inmutabilidad)
func (d Dinero) Sumar(otro Dinero) (Dinero, error) {
	if d.Moneda != otro.Moneda {
		return Dinero{}, errors.New("no se pueden sumar monedas diferentes")
	}
	return Dinero{MontoCentavos: d.MontoCentavos + otro.MontoCentavos, Moneda: d.Moneda}, nil
}

// Método Restar: Impide saldos negativos
func (d Dinero) Restar(otro Dinero) (Dinero, error) {
	if d.Moneda != otro.Moneda {
		return Dinero{}, errors.New("no se pueden restar monedas diferentes")
	}
	if d.MontoCentavos < otro.MontoCentavos {
		return Dinero{}, errors.New("saldo insuficiente: la resta resultaría en un saldo negativo")
	}
	return Dinero{MontoCentavos: d.MontoCentavos - otro.MontoCentavos, Moneda: d.Moneda}, nil
}

// Evento de Dominio
type CargoAplicadoEvent struct {
	SuscripcionID string
	Monto         Dinero
	Fecha         time.Time
}

// Raíz de Agregado
type Suscripcion5G struct {
	ID     string
	Saldo  Dinero
	eventos []interface{}
}

func NuevaSuscripcion(id string, saldoInicial int64) *Suscripcion5G {
	return &Suscripcion5G{
		ID:    id,
		Saldo: Dinero{MontoCentavos: saldoInicial, Moneda: "USD"},
	}
}

// Protege la invariante: No dejar saldo negativo
func (s *Suscripcion5G) AplicarCargo(monto Dinero) error {
	nuevoSaldo, err := s.Saldo.Restar(monto)
	if err != nil {
		return err // Rechaza el cargo
	}
	s.Saldo = nuevoSaldo
	s.eventos = append(s.eventos, CargoAplicadoEvent{
		SuscripcionID: s.ID,
		Monto:         monto,
		Fecha:         time.Now(),
	})
	return nil
}

// ==========================================
// 1.3 CIRCUIT BREAKER & BACKOFF
// ==========================================

type CircuitBreaker struct {
	mu           sync.Mutex
	state        string // "CLOSED", "OPEN", "HALF-OPEN"
	failures     int
	lastFailTime time.Time
	threshold    int
	timeout      time.Duration
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{state: "CLOSED", threshold: threshold, timeout: timeout}
}

func (cb *CircuitBreaker) Execute(accion func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Si está OPEN, verificar si ya pasó el tiempo para intentar HALF-OPEN
	if cb.state == "OPEN" {
		if time.Since(cb.lastFailTime) > cb.timeout {
			cb.state = "HALF-OPEN"
		} else {
			return errors.New("circuit breaker OPEN: llamada bloqueada para proteger el sistema")
		}
	}

	err := accion()
	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()
		if cb.failures >= cb.threshold || cb.state == "HALF-OPEN" {
			cb.state = "OPEN"
		}
		return err
	}

	// Éxito: Reiniciar contadores
	cb.failures = 0
	cb.state = "CLOSED"
	return nil
}

// Función Matemática: Exponential Backoff + Full Jitter
func BackoffFullJitter(intento int) time.Duration {
	T_base := 100 * time.Millisecond
	T_max := 5 * time.Second
	
	// Fórmula: T = Uniform(0, min(T_max, T_base * 2^k))
	exp := math.Pow(2, float64(intento))
	backoff := float64(T_base) * exp
	if backoff > float64(T_max) {
		backoff = float64(T_max)
	}
	// Full Jitter (Aleatoriedad entre 0 y backoff)
	sleepTime := time.Duration(rand.Float64() * backoff)
	fmt.Printf("   [Backoff] Intento %d falló. Reintentando en %v...\n", intento, sleepTime)
	time.Sleep(sleepTime)
	return sleepTime
}

// ==========================================
// 1.3 ORQUESTADOR SAGA
// ==========================================

type SagaOrchestrator struct {
	suscripcion *Suscripcion5G
	cb          *CircuitBreaker
}

func (s *SagaOrchestrator) EjecutarSaga(montoCargo int64) error {
	fmt.Println("\n--- Iniciando SAGA Transaccional ---")
	monto := Dinero{MontoCentavos: montoCargo, Moneda: "USD"}

	// T1: Debitar Billetera (Usa Circuit Breaker)
	fmt.Println("T1: Debitando billetera...")
	err := s.cb.Execute(func() error {
		return s.suscripcion.AplicarCargo(monto)
	})
	if err != nil {
		return fmt.Errorf("SAGA falló en T1: %w", err)
	}

	// T2: Aprovisionar QoS 5G
	fmt.Println("T2: Aprovisionando QoS 5G...")
	// Simulamos un fallo si el monto es mayor a 5000 para probar la compensación
	if montoCargo > 5000 {
		fmt.Println("   [!] Fallo simulado en T2 (Red 5G caída).")
		// COMPENSACIÓN C1: Revertir el débito
		fmt.Println("   [C1] Compensando T1: Reembolsando saldo al cliente...")
		s.suscripcion.Saldo, _ = s.suscripcion.Saldo.Sumar(monto)
		return errors.New("SAGA abortada: T2 falló, compensación ejecutada")
	}

	// T3: Emitir Factura
	fmt.Println("T3: Emitiendo factura...")
	fmt.Println("--- SAGA Completada Exitosamente ---")
	return nil
}


