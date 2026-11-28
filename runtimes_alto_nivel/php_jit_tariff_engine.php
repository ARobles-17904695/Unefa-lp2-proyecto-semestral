<?php
declare(strict_types=1);

enum TipoRed: string
{
    case VISA = 'visa';
    case MASTERCARD = 'mastercard';
    case TRANSFERENCIA_INTERBANCARIA = 'ach';
    case CRIPTO_STABLE = 'usdt';
}

readonly class CalculadorTarifas
{
    public function __construct(
        public float $tarifaBaseFija,
        public float $impuestoFijo
    ) {}

    public function calcularComision(TipoRed $red, float $monto): float
    {
        $porcentaje = match ($red) {
            TipoRed::VISA => 0.029,
            TipoRed::MASTERCARD => 0.027,
            TipoRed::TRANSFERENCIA_INTERBANCARIA => 0.005,
            TipoRed::CRIPTO_STABLE => 0.010,
        };

        return round(($monto * $porcentaje) + $this->tarifaBaseFija + $this->impuestoFijo, 2);
    }
}

$motorTarifas = new CalculadorTarifas(tarifaBaseFija: 0.30, impuestoFijo: 0.15);
$montoPrueba = 500.00;

foreach (TipoRed::cases() as $red) {
    $comision = $motorTarifas->calcularComision($red, $montoPrueba);
    echo sprintf(
        "Red: %-12s | Monto: $%.2f | Comisión Total: $%.2f\n",
        strtoupper($red->value),
        $montoPrueba,
        $comision
    );
}