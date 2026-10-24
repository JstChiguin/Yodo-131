# Historial de versiones

Formato basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/) y [versionado semántico](https://semver.org/lang/es/).

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
