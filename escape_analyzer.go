package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Sesion5G representa la telemetría
type Sesion5G struct {
	ID       uint64
	IMSI     uint64
	Operador string
}

// Constructor 1: Retorna la estructura por valor. Go probablemente la asigne en el Stack.
func CrearEnStackPorValor() Sesion5G {
	s := Sesion5G{ID: 1, IMSI: 123456789, Operador: "TelcoA"}
	return s
}

// Constructor 2: Retorna un puntero a una variable local. 
// Go detecta que el puntero "escapa" de la función, por lo que la mueve al Heap.
func CrearConEscapeAlHeap() *Sesion5G {
	s := Sesion5G{ID: 2, IMSI: 987654321, Operador: "TelcoB"}
	return &s
}

func main() {
	// 1. Demostración de Escape Analysis (conceptual en ejecución)
	_ = CrearEnStackPorValor()
	_ = CrearConEscapeAlHeap()
	fmt.Println("--- Benchmarks de Búsqueda ---")

	// 2. Preparación de datos: N = 100,000 abonados
	const N = 100000
	const busquedas = 1000

	// Datos para búsqueda lineal (Slice)
	loteSlice := make([]Sesion5G, N)
	// Datos para búsqueda O(1) (Mapa)
	loteMapa := make(map[uint64]Sesion5G, N)

	for i := 0; i < N; i++ {
		s := Sesion5G{ID: uint64(i), IMSI: uint64(rand.Intn(10000000)), Operador: "Telco"}
		loteSlice[i] = s
		loteMapa[s.IMSI] = s
	}

	// Generar 1000 IMSIs aleatorios a buscar
	busquedasIMSI := make([]uint64, busquedas)
	for i := 0; i < busquedas; i++ {
		busquedasIMSI[i] = uint64(rand.Intn(10000000))
	}

	// A. Benchmark Búsqueda Lineal O(n)
	inicioLineal := time.Now()
	for _, imsiBuscado := range busquedasIMSI {
		for _, s := range loteSlice {
			if s.IMSI == imsiBuscado {
				break // Encontrado
			}
		}
	}
	tiempoLineal := time.Since(inicioLineal).Microseconds()

	// B. Benchmark Búsqueda Hash Map O(1)
	inicioMapa := time.Now()
	for _, imsiBuscado := range busquedasIMSI {
		_, _ = loteMapa[imsiBuscado] // Acceso directo O(1)
	}
	tiempoMapa := time.Since(inicioMapa).Microseconds()

	fmt.Printf("Búsqueda Lineal O(n) en %d elementos: %d microsegundos\n", N, tiempoLineal)
	fmt.Printf("Búsqueda Hash Map O(1) en %d elementos: %d microsegundos\n", N, tiempoMapa)
}
