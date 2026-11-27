package ports

import "lab2_arquitectura_hexagonal/internal/domain"

// Puerto de Salida (Repositorio)
type RepositorioTransacciones interface {
	ObtenerPorIdempotencyKey(key string) (*domain.TransaccionRecarga, error)
	Guardar(t *domain.TransaccionRecarga) error
}

// Puerto de Salida (Pasarela de Pago - Servicio externo)
type PasarelaPagoBancaria interface {
	Cobrar(msisdn string, montoCentavos int64) error
}

// Puerto de Salida (Notificador)
type NotificadorAuditoria interface {
	Registrar(evento string, t *domain.TransaccionRecarga)
}

package application

import (
	"lab2_arquitectura_hexagonal/internal/domain"
	"lab2_arquitectura_hexagonal/internal/ports"
)

type ServicioRecarga struct {
	repo      ports.RepositorioTransacciones
	pasarela  ports.PasarelaPagoBancaria
	notificador ports.NotificadorAuditoria
}

// Inyección de dependencias
func NuevoServicioRecarga(r ports.RepositorioTransacciones, p ports.PasarelaPagoBancaria, n ports.NotificadorAuditoria) *ServicioRecarga {
	return &ServicioRecarga{repo: r, pasarela: p, notificador: n}
}

// EjecutarRecarga maneja la lógica de idempotencia
func (s *ServicioRecarga) EjecutarRecarga(id, idempKey, msisdn string, monto int64) (*domain.TransaccionRecarga, error) {
	// 1. Verificar Idempotencia: ¿Ya existe esta llave?
	tExistente, _ := s.repo.ObtenerPorIdempotencyKey(idempKey)
	if tExistente != nil {
		// Si ya existe, retornamos la original sin volver a debitar
		return tExistente, nil 
	}

	// 2. Si no existe, creamos la entidad de dominio (valida E.164 y monto)
	nuevaT, err := domain.NuevaTransaccionRecarga(id, idempKey, msisdn, monto)
	if err != nil {
		s.notificador.Registrar("RECHAZO_VALIDACION", nuevaT)
		return nil, err
	}

	// 3. Guardar en estado PENDIENTE antes de ir al banco (Patrón Saga/Outbox simple)
	if err := s.repo.Guardar(nuevaT); err != nil {
		return nil, err
	}

	// 4. Llamar a la pasarela de pago (Servicio externo)
	errPago := s.pasarela.Cobrar(msisdn, monto)
	
	// 5. Actualizar estado según el resultado
	if errPago != nil {
		nuevaT.Rechazar()
		s.notificador.Registrar("PAGO_FALLIDO", nuevaT)
	} else {
		nuevaT.Aprobar()
		s.notificador.Registrar("PAGO_EXITOSO", nuevaT)
	}
	
	// 6. Guardar estado final
	s.repo.Guardar(nuevaT)
	return nuevaT, nil
}

package main

import (
	"errors"
	"fmt"
	"lab2_arquitectura_hexagonal/internal/application"
	"lab2_arquitectura_hexagonal/internal/domain"
	"sync"
)

// ==========================================
// ADAPTADORES (Mocks en memoria)
// ==========================================

// Adaptador de Repositorio
type MemRepo struct {
	mu    sync.Mutex
	datos map[string]*domain.TransaccionRecarga
}

func (m *MemRepo) ObtenerPorIdempotencyKey(key string) (*domain.TransaccionRecarga, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.datos[key]; ok {
		return t, nil
	}
	return nil, nil // No encontrado
}
func (m *MemRepo) Guardar(t *domain.TransaccionRecarga) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.datos[t.IdempotencyKey] = t
	return nil
}

// Adaptador de Pasarela (Simula una API bancaria)
type MockPasarela struct{}
func (m *MockPasarela) Cobrar(msisdn string, monto int64) error {
	if monto > 100000 {
		return errors.New("fondos insuficientes en el banco")
	}
	return nil // Cobro exitoso
}

// Adaptador de Notificador
type MockNotificador struct{}
func (m *MockNotificador) Registrar(evento string, t *domain.TransaccionRecarga) {
	fmt.Printf("[AUDITORÍA] Evento: %s | ID: %s | Monto: %d\n", evento, t.ID, t.MontoCentavos)
}

// ==========================================
// ENSAMBLAJE (Main)
// ==========================================
func main() {
	// 1. Instanciar Adaptadores (Infraestructura)
	repo := &MemRepo{datos: make(map[string]*domain.TransaccionRecarga)}
	pasarela := &MockPasarela{}
	notificador := &MockNotificador{}

	// 2. Instanciar Caso de Uso inyectando Puertos
	servicio := application.NuevoServicioRecarga(repo, pasarela, notificador)

	fmt.Println("--- (a) Recarga Exitosa ---")
	t1, err := servicio.EjecutarRecarga("tx-001", "uuid-key-123", "+1234567890", 5000)
	if err != nil { fmt.Println("Error:", err) } else { fmt.Printf("Transacción creada: Estado %s\n", t1.Estado) }

	fmt.Println("\n--- (b) Rechazo por violar E.164 y Monto ---")
	_, err = servicio.EjecutarRecarga("tx-002", "uuid-key-456", "12345", 50) // Sin '+' y monto < 100
	if err != nil { fmt.Println("Rechazado correctamente:", err) }

	fmt.Println("\n--- (c) Reintentos con la misma Idempotency-Key ---")
	// Simulamos 3 reintentos de red con la misma key "uuid-key-123"
	for i := 1; i <= 3; i++ {
		tRetry, _ := servicio.EjecutarRecarga("tx-001-retry", "uuid-key-123", "+1234567890", 5000)
		fmt.Printf("Intento %d: Devuelve ID original (%s) - Estado: %s - ¿Se cobró de nuevo?: NO\n", i, tRetry.ID, tRetry.Estado)
	}
}
