package main

import (
	"errors"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var archivosObligatorios = []string{"index.html", "js/version.js", "js/selloAutor.js"}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "uso: generador <carpeta de la aplicación> <archivo .go de salida>")
		os.Exit(2)
	}
	if errorGeneracion := generarArchivoEmpaquetado(os.Args[1], os.Args[2]); errorGeneracion != nil {
		fmt.Fprintln(os.Stderr, "generador:", errorGeneracion)
		os.Exit(1)
	}
}

func generarArchivoEmpaquetado(carpetaAplicacion, rutaSalida string) error {
	archivos, errorLectura := leerArchivosAplicacion(carpetaAplicacion)
	if errorLectura != nil {
		return errorLectura
	}
	codigo, errorCodigo := generarCodigo(archivos)
	if errorCodigo != nil {
		return errorCodigo
	}
	return os.WriteFile(rutaSalida, codigo, 0o644)
}

func leerArchivosAplicacion(carpetaAplicacion string) (map[string]string, error) {
	archivos := map[string]string{}
	errorRecorrido := filepath.WalkDir(carpetaAplicacion, func(ruta string, entrada fs.DirEntry, errorEntrada error) error {
		if errorEntrada != nil {
			return errorEntrada
		}
		if entrada.IsDir() {
			return nil
		}
		rutaRelativa, errorRelativa := filepath.Rel(carpetaAplicacion, ruta)
		if errorRelativa != nil {
			return errorRelativa
		}
		contenido, errorArchivo := os.ReadFile(ruta)
		if errorArchivo != nil {
			return errorArchivo
		}
		archivos[filepath.ToSlash(rutaRelativa)] = string(contenido)
		return nil
	})
	if errorRecorrido != nil {
		return nil, fmt.Errorf("no se pudo leer la aplicación: %w", errorRecorrido)
	}
	for _, obligatorio := range archivosObligatorios {
		if _, existe := archivos[obligatorio]; !existe {
			return nil, errors.New("falta el archivo obligatorio " + obligatorio)
		}
	}
	return archivos, nil
}

func generarCodigo(archivos map[string]string) ([]byte, error) {
	rutas := make([]string, 0, len(archivos))
	for ruta := range archivos {
		rutas = append(rutas, ruta)
	}
	sort.Strings(rutas)

	var codigo strings.Builder
	codigo.WriteString("package main\n\nvar archivosAplicacion = map[string]string{\n")
	for _, ruta := range rutas {
		codigo.WriteString("\t" + strconv.Quote(ruta) + ": " + strconv.Quote(archivos[ruta]) + ",\n")
	}
	codigo.WriteString("}\n")
	return format.Source([]byte(codigo.String()))
}
