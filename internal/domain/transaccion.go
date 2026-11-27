package domain

import (
	"errors"
	"regexp"
)

// Entidad de Dominio
type TransaccionRecarga struct {
	ID            string
	IdempotencyKey string
	Msisdn        string
	MontoCentavos int64
	Estado        string
}

// Estados permitidos
const (
	EstadoPendiente  = "PENDIENTE"
	EstadoAprobada   = "APROBADA"
	EstadoRechazada  = "RECHAZADA"
)

// Constructor (Fábrica) - Aplica las validaciones de negocio
func NuevaTransaccionRecarga(id, idempKey, msisdn string, montoCentavos int64) (*TransaccionRecarga, error) {
	// 1. Validar formato E.164 usando Regex
	// ^\+[1-9]\d{6,14}$ exige un '+' seguido de un dígito del 1 al 9, y de 6 a 14 dígitos más.
	re := regexp.MustCompile(`^\+[1-9]\d{6,14}$`)
	if !re.MatchString(msisdn) {
		return nil, errors.New("msisdn inválido: no cumple con el estándar E.164")
	}

	// 2. Validar monto estrictamente positivo (y al menos 100 centavos según el YAML)
	if montoCentavos < 100 {
		return nil, errors.New("monto inválido: debe ser mayor o igual a 100 centavos")
	}

	// 3. Crear entidad en estado inicial PENDIENTE
	return &TransaccionRecarga{
		ID:            id,
		IdempotencyKey: idempKey,
		Msisdn:        msisdn,
		MontoCentavos: montoCentavos,
		Estado:        EstadoPendiente,
	}, nil
}

// Método de negocio para cambiar estado (Máquina de estados)
func (t *TransaccionRecarga) Aprobar() error {
	if t.Estado != EstadoPendiente {
		return errors.New("solo se pueden aprobar transacciones PENDIENTES")
	}
	t.Estado = EstadoAprobada
	return nil
}

func (t *TransaccionRecarga) Rechazar() error {
	if t.Estado != EstadoPendiente {
		return errors.New("solo se pueden rechazar transacciones PENDIENTES")
	}
	t.Estado = EstadoRechazada
	return nil
}
