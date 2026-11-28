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
