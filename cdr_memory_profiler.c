#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <stddef.h>

// Estructura 1: Desalineada (Intercala tipos de 1 byte con 8 y 4 bytes)
// El compilador insertará "padding" (bytes basura) para alinear los uint32_t y uint64_t.
struct CDR_Desalineado {
    char tipo_red;            // 1 byte
    uint64_t imsi_abonado;    // 8 bytes (requiere alineación de 8, por lo que habrá 7 bytes de padding después de tipo_red)
    char bandera_roaming;     // 1 byte
    uint32_t celda_id;        // 4 bytes (requiere alineación de 4, habrá 3 bytes de padding)
    char codigo_estado;       // 1 byte
    uint16_t puerto_origen;   // 2 bytes (requiere alineación de 2, habrá 1 byte de padding)
};

// Estructura 2: Optimizada (Reordenada de mayor a menor requerimiento de alineación)
// Minimiza el padding. Los campos de 8 van primero, luego 4, 2 y 1.
struct CDR_Optimizado {
    uint64_t imsi_abonado;    // 8 bytes
    uint32_t celda_id;        // 4 bytes
    uint16_t puerto_origen;   // 2 bytes
    char tipo_red;            // 1 byte
    char bandera_roaming;     // 1 byte
    char codigo_estado;       // 1 byte
    // Total: 8 + 4 + 2 + 1 + 1 + 1 = 17 bytes. 
    // El compilador añadirá 7 bytes de padding al final para que el tamaño total sea múltiplo de 8 (24 bytes).
};

// Variable Global
int data = 42; 

int main() {
    // (a) Direcciones de memoria virtuales
    int local_var = 10;
    int *heap_var = (int *)malloc(sizeof(int));
    if (heap_var == NULL) return 1;
    *heap_var = 20;

    printf("=== (a) Direcciones de Memoria Virtual ===\n");
    printf("Global 'data'  (Segmento de Datos): %p\n", (void*)&data);
    printf("Local  'local' (Stack):             %p\n", (void*)&local_var);
    printf("Heap   'heap'  (Heap):              %p\n\n", (void*)heap_var);

    // Liberamos la memoria para evitar Memory Leaks
    free(heap_var);
    heap_var = NULL; 

    // (b) Tamaños y desplazamientos (Offsets)
    printf("=== (b) Análisis de Memoria (Sizeof y Offsets) ===\n");
    printf("Tamaño struct CDR_Desalineado: %zu bytes\n", sizeof(struct CDR_Desalineado));
    printf("  Offset imsi_abonado: %zu\n", offsetof(struct CDR_Desalineado, imsi_abonado));
    printf("  Offset celda_id:     %zu\n", offsetof(struct CDR_Desalineado, celda_id));
    
    printf("\nTamaño struct CDR_Optimizado:  %zu bytes\n", sizeof(struct CDR_Optimizado));
    printf("  Offset imsi_abonado: %zu\n", offsetof(struct CDR_Optimizado, imsi_abonado));
    printf("  Offset celda_id:     %zu\n\n", offsetof(struct CDR_Optimizado, celda_id));

    // (c) Cálculos de RAM y Caché L1
    size_t lote = 10000000; // 10 millones
    size_t mem_desalineada = sizeof(struct CDR_Desalineado) * lote;
    size_t mem_optimizada = sizeof(struct CDR_Optimizado) * lote;
    size_t ahorro = mem_desalineada - mem_optimizada;

    printf("=== (c) Impacto en RAM y Caché L1 ===\n");
    printf("RAM para 10M Desalineados: %.2f MB\n", (double)mem_desalineada / (1024 * 1024));
    printf("RAM para 10M Optimizados:  %.2f MB\n", (double)mem_optimizada / (1024 * 1024));
    printf("Ahorro total de RAM:       %.2f MB (%.0f bytes)\n", (double)ahorro / (1024 * 1024), (double)ahorro);
    
    // L1 Cache = 64 bytes. Cuántos registros caben por línea.
    int regs_desalineados_cache = 64 / sizeof(struct CDR_Desalineado);
    int regs_optimizados_cache = 64 / sizeof(struct CDR_Optimizado);
    printf("Registros por Línea de Caché L1 (64 bytes):\n");
    printf("  Desalineado: %d registros\n", regs_desalineados_cache);
    printf("  Optimizado:  %d registros\n", regs_optimizados_cache);

    return 0;
}
