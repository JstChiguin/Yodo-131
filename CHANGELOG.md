# Historial de versiones

Formato basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/) y [versionado semántico](https://semver.org/lang/es/).

## [1.1.1] - 2026-10-24

### Corregido
- Si la firma «by Chiguin» quedaba oculta por estilos, el programa la recreaba una y otra vez y la página se congelaba. Ahora deja de intentarlo, avisa en la consola y la aplicación sigue funcionando.

### Agregado
- Prueba de integridad de la firma en el navegador. Verifica en escritorio, tablet y celular, en modo claro y oscuro y en todas las pestañas, que el texto sea exactamente "by Chiguin" y que esté visible, flotante, en la esquina inferior derecha y con la versión centrada debajo. Si algo de eso cambia, la prueba falla con un mensaje claro.
- Integración continua con GitHub Actions: las pruebas corren en cada push y pull request.

## [1.1.0] - 2026-10-24

### Agregado
- Botón para cambiar entre modo claro y oscuro. El programa siempre abre en modo claro.
- Cada gráfico tiene una explicación de lo que muestra y una conclusión destacada (por ejemplo, el diagnóstico más frecuente y su porcentaje).
- Cantidad y porcentaje en todas las barras; eje con líneas guía y números dentro de cada segmento en el gráfico por año.
- Significado de cada categoría de TIRADS y Bethesda junto a su barra.
- Detalle al pasar el mouse o al recorrer los gráficos con el teclado.

### Cambiado
- Las etiquetas de las barras se muestran completas sobre la barra, en vez de cortarse.
- Las barras de exámenes, ecografía y antecedentes se calculan sobre los informes que traen el dato; los "Sin dato" se informan aparte.
- El modo oscuro ya no depende de la configuración del sistema operativo.

## [1.0.0] - 2026-10-24

### Agregado
- Lectura de carpetas de informes Word (`.docx`) sin librerías ni internet, incluidos documentos con varios informes.
- Extracción de datos del paciente, motivo, diagnóstico, tipo histológico, ecografía (TIRADS), punción (Bethesda), laboratorio tiroideo, anticuerpos, antecedentes, médico tratante y dosis.
- Gráficos por año y sexo, diagnóstico, motivo, edad, dosis, exámenes elevados, ecografía, Bethesda, histología, antecedentes y médico tratante.
- Filtro por período, tabla de pacientes con búsqueda y orden, y pestaña de datos por revisar.
- Exportación a Excel con las hojas Informes, Resumen y No leídos.
- Sección de mantenimiento protegida por clave (solo se guarda la huella PBKDF2) para ajustar umbrales y ver el registro de errores.
- Firma «by Chiguin» y versión fijas en la esquina inferior derecha.
- Informes ficticios de ejemplo y pruebas automáticas.
