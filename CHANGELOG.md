# Historial de versiones

Formato basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/) y [versionado semántico](https://semver.org/lang/es/).

## [1.2.0] - 2026-10-24

### Agregado
- Lectura de planillas Excel (`.xlsx`) sin librerías: la planilla de yodo 131, con una hoja por año, y el registro de cáncer de tiroides (C73). El tipo de cada hoja se reconoce por su contenido. Se toleran filas con columnas corridas, como en la hoja "LISTADO".
- Distinción entre **pacientes únicos** (una persona cuenta una vez), **tratamientos** (paciente + fecha; una persona puede tener varios) y **registros repetidos** del mismo tratamiento. Dos registros del mismo paciente con 60 días o menos de diferencia se consideran el mismo tratamiento y no aumentan las estadísticas.
- Combinación de datos entre Word y Excel del mismo tratamiento: se conserva lo que aporta cada fuente (por ejemplo, la fecha del Excel y el diagnóstico del Word).
- Detección de discrepancias entre fuentes: dosis, edad, médico, diagnóstico, exámenes, antecedentes y sexo, este último comparando lo registrado con lo deducido del nombre. También se avisa cuando un mismo RUT aparece con nombres distintos. El dato en discrepancia queda fuera de los gráficos y aparece en "Por revisar", con el valor y la ubicación de cada fuente.
- Cobertura en cada gráfico: cuántos tratamientos tienen el dato y cuántos no. Los porcentajes se calculan solo sobre los que tienen el dato, y un examen no registrado no se cuenta como normal.
- Gráfico de pacientes con cáncer de tiroides por año de primer registro, y gráfico del origen de los datos (Word, Excel o ambos).
- Cifras de pacientes únicos, pacientes con más de un tratamiento, registros unidos y datos que no coinciden.
- Excel exportado con las hojas Tratamientos (con sus fuentes), Resumen (con cobertura), Por revisar y No leídos.
- Planillas Excel ficticias de ejemplo y pruebas de fusión y cobertura.

### Cambiado
- Un dato ausente se informa como "sin dato", en vez de usar categorías como "Otro" o "Sin motivo escrito".
- El lector de archivos comprimidos se comparte entre Word y Excel (`lectorZip.js`).

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
