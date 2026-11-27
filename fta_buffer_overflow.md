graph TD
    A["A[Corrupción del registro RIP]"] --> B("B(Ausencia de validación de entrada)")
    A --> C("C(Falta de límites en funciones de copia)")
    A --> D("D(Configuración de compilador inseguro)")
    
    B --> B1("B1(No se valida la longitud del input del usuario)")
    
    C --> C1("C1(Uso de 'strcpy' o 'sprintf' en C)")
    C --> C2("C2(No usar 'strncpy' o 'snprintf')")
    
    D --> D1("D1(Falta de Stack Canaries)")
    D --> D2("D2(Deshabilitado ASLR)")
    
    ARRIBA(("EVENTO PRINCIPAL: Corrupción de RIP y Control de Flujo"))
    
    ARRIBA -->|Y| G1("G1(Desbordamiento del Buffer)")
    ARRIBA -->|Y| G2("G2(Falta de Mecanismos de Defensa)")
    
    G1 -->|O| CAUSA1["CAUSA1[Copiar datos sin verificar longitud]"]
    G1 -->|O| CAUSA2["CAUSA2[Input del atacante mayor al buffer asignado]"]
    
    CAUSA1 -->|O| C_1_1["C_1_1[Uso de función 'strcpy' vulnerable]"]
    CAUSA1 -->|O| C_1_2["C_1_2[Error de cálculo en 'sizeof']"]
    
    G2 -->|O| G2A("G2A(Falta de Stack Canary)")
    G2 -->|O| G2B("G2B(ASLR deshabilitado)")
    G2 -->|O| G2C("G2C(Sección .text con permisos de escritura - NX bit off)")
    
    CAUSA2 -->|Y| C_2_1["C_2_1[Atacante envía payload malicioso]"]
    CAUSA2 -->|Y| C_2_2["C_2_2[Payload supera límite del buffer local en Stack]"]
    
    %% Explicación del flujo lógico hacia RIP
    G1 --> FLUJO1["FLUJO1[Los datos sobrescritos variables locales en el Stack]"]
    FLUJO1 --> FLUJO2["FLUJO2[Los datos sobrescriben el 'Saved RBP']"]
    FLUJO2 --> FLUJO3["FLUJO3[Los datos sobrescriben la dirección de retorno]"]
    FLUJO3 --> RESULTADO["RESULTADO[El 'ret' de la función salta a la dirección controlada por el atacante]"]
    RESULTADO -.-> TOP
