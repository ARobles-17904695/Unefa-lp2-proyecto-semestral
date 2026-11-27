# Árbol de Fallos (FTA) - Corrupción de RIP por Buffer Overflow

Este documento detalla el análisis de fallos (FTA) mediante un diagrama de Ishikawa y un Árbol de Fallos (FTA) lógico para un incidente de seguridad donde un desbordamiento de búfer corrompe el registro `RIP`.

## 1. Diagrama de Ishikawa (Causa-Efecto)
*Identificando los factores que contribuyen al fallo.*

```mermaid
graph TD
    A[Corrupción del registro RIP] --> B(Ausencia de validación de entrada)
    A --> C(Falta de límites en funciones de copia)
    A --> D(Configuración de compilador insegura)
    
    B --> B1(No se valida la longitud del input del usuario)
    
    C --> C1(Uso de 'strcpy' o 'sprintf' en C)
    C --> C2(No usar 'strncpy' o 'snprintf')
    
    D --> D1(Falta de Stack Canaries)
    D --> D2(Deshabilitado ASLR)
    
    
    graph TD
    TOP((EVENTO PRINCIPAL: \n Corrupción de RIP y \n Control de Flujo)) 
    
    TOP -->|AND| G1(Desbordamiento del Buffer)
    TOP -->|AND| G2(Falta de Mecanismos de Defensa)
    
    G1 -->|OR| CAUSA1[Copiar datos sin verificar longitud]
    G1 -->|OR| CAUSA2[Input del atacante mayor al buffer asignado]
    
    CAUSA1 -->|OR| C1_1[Uso de función 'strcpy' vulnerable]
    CAUSA1 -->|OR| C1_2[Error de cálculo en 'sizeof']
    
    G2 -->|OR| G2A(Falta de Stack Canary)
    G2 -->|OR| G2B(ASLR deshabilitado)
    G2 -->|OR| G2C(Sección .text con permisos de escritura - NX bit off)
    
    CAUSA2 -->|AND| C2_1[Atacante envía payload malicioso]
    CAUSA2 -->|AND| C2_2[Payload supera límite del buffer local en Stack]

    %% Explicación del flujo lógico hacia RIP
    G1 --> FLUJO1[Los datos sobrescriben variables locales en el Stack]
    FLUJO1 --> FLUJO2[Los datos sobrescriben el 'Saved RBP']
    FLUJO2 --> FLUJO3[Los datos sobrescriben la dirección de retorno]
    FLUJO3 --> RESULTADO[El 'ret' de la función salta a la dirección controlada por el atacante]
    RESULTADO -.-> TOP
