import asyncio
from dataclasses import dataclass
from typing import Final


@dataclass(slots=True, frozen=True)
class LoteTransaccion:
    id_lote: str
    monto: float
    moneda: str


@dataclass(slots=True)
class ResultadoValidacion:
    id_lote: str
    valido: bool
    observacion: str


MONEDAS_PERMITIDAS: Final[set[str]] = {"USD", "EUR", "VES"}


async def validar_lote(lote: LoteTransaccion) -> ResultadoValidacion:
    # Simula latencia de I/O (consulta a microservicio o base de datos)
    await asyncio.sleep(0.05)

    if lote.moneda not in MONEDAS_PERMITIDAS:
        return ResultadoValidacion(lote.id_lote, False, f"Moneda no soportada: {lote.moneda}")
    if lote.monto <= 0:
        return ResultadoValidacion(lote.id_lote, False, "Monto debe ser positivo")

    return ResultadoValidacion(lote.id_lote, True, "OK")


async def main() -> None:
    lotes: list[LoteTransaccion] = [
        LoteTransaccion("LOTE-01", 1250.50, "USD"),
        LoteTransaccion("LOTE-02", -40.00, "EUR"),
        LoteTransaccion("LOTE-03", 9800.00, "VES"),
        LoteTransaccion("LOTE-04", 300.00, "BTC"),
    ]

    resultados: list[ResultadoValidacion] = await asyncio.gather(
        *(validar_lote(lote) for lote in lotes)
    )

    for res in resultados:
        estado: str = "VÁLIDO" if res.valido else "INVÁLIDO"
        print(f"[{res.id_lote}] {estado} -> {res.observacion}")


if __name__ == "__main__":
    asyncio.run(main())