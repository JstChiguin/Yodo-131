# Prototipo nativo en Go

Este directorio contiene un **prototipo** de la versión nativa de Yodo 131 para Windows, en evaluación. La versión web en `Sistema - no tocar/` sigue siendo la versión funcional y no se modificó.

## Estructura

- `nucleo/`: la lógica portada a Go, solo con la biblioteca estándar. Lee DOCX y XLSX, extrae, clasifica, fusiona, calcula estadísticas y exporta Excel. Incluye los diálogos nativos de Windows y la prueba de firma.
- `app/`: la interfaz nativa con Gio (sin navegador). Depende de `gioui.org`, que solo se descarga al compilar en GitHub Actions.

## Pruebas diferenciales (JS vs Go)

`nucleo/diferencial` compara, sobre los ejemplos ficticios, la salida de Go con un oráculo generado desde el JavaScript (`nucleo/diferencial/testdata/oraculo.json` y `oraculo.xlsx`). Compara extracción, clasificación, fusión, estadísticas y el Excel exportado. El oráculo se regenera con:

```
node herramientas/generarOraculo.mjs
cd nativo/nucleo && go test ./...
```

## Compilar el ejecutable (requiere internet para Gio)

```
cd nativo/app
go mod tidy
go build -ldflags "-H windowsgui" -o Yodo131.exe .
```

## Autoverificación

```
Yodo131.exe --autoverificacion ejemplos --salida salida.xlsx
```

Lee los ejemplos ficticios, exporta el Excel y confirma la firma, sin escribir datos en AppData ni Temp.

by Chiguin
