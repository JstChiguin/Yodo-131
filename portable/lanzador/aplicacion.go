package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var patronVersion = regexp.MustCompile(`VERSION_APLICACION\s*=\s*'([0-9]+\.[0-9]+\.[0-9]+)'`)

var errVersionNoEncontrada = errors.New("no se encontró la versión en js/version.js")

func obtenerVersionAplicacion(archivos map[string]string) (string, error) {
	contenidoVersion, existe := archivos["js/version.js"]
	if !existe {
		return "", errVersionNoEncontrada
	}
	coincidencia := patronVersion.FindStringSubmatch(contenidoVersion)
	if coincidencia == nil {
		return "", errVersionNoEncontrada
	}
	return coincidencia[1], nil
}

func esRutaRelativaSegura(ruta string) bool {
	if ruta == "" || strings.HasPrefix(ruta, "/") || strings.Contains(ruta, `\`) || strings.Contains(ruta, ":") {
		return false
	}
	for _, parte := range strings.Split(ruta, "/") {
		if parte == "" || parte == "." || parte == ".." {
			return false
		}
	}
	return true
}

func extraerAplicacion(archivos map[string]string, carpetaDestino string) (int, error) {
	rutas := make([]string, 0, len(archivos))
	for ruta := range archivos {
		rutas = append(rutas, ruta)
	}
	sort.Strings(rutas)

	cantidadEscritos := 0
	for _, ruta := range rutas {
		if !esRutaRelativaSegura(ruta) {
			return cantidadEscritos, fmt.Errorf("ruta no permitida en la aplicación empaquetada: %q", ruta)
		}
		rutaDestino := filepath.Join(carpetaDestino, filepath.FromSlash(ruta))
		contenidoEsperado := []byte(archivos[ruta])
		contenidoActual, errorLectura := os.ReadFile(rutaDestino)
		if errorLectura == nil && bytes.Equal(contenidoActual, contenidoEsperado) {
			continue
		}
		if errorEscritura := escribirArchivoCompleto(rutaDestino, contenidoEsperado); errorEscritura != nil {
			return cantidadEscritos, errorEscritura
		}
		cantidadEscritos++
	}
	return cantidadEscritos, nil
}

func escribirArchivoCompleto(rutaDestino string, contenido []byte) error {
	carpeta := filepath.Dir(rutaDestino)
	if errorCarpeta := os.MkdirAll(carpeta, 0o755); errorCarpeta != nil {
		return fmt.Errorf("no se pudo crear la carpeta de la aplicación: %w", errorCarpeta)
	}
	archivoTemporal, errorCreacion := os.CreateTemp(carpeta, ".yodo131-*")
	if errorCreacion != nil {
		return fmt.Errorf("no se pudo crear un archivo temporal: %w", errorCreacion)
	}
	rutaTemporal := archivoTemporal.Name()
	_, errorEscritura := archivoTemporal.Write(contenido)
	errorCierre := archivoTemporal.Close()
	if errorEscritura == nil {
		errorEscritura = errorCierre
	}
	if errorEscritura == nil {
		errorEscritura = os.Rename(rutaTemporal, rutaDestino)
	}
	if errorEscritura != nil {
		os.Remove(rutaTemporal)
		return fmt.Errorf("no se pudo escribir %s: %w", filepath.Base(rutaDestino), errorEscritura)
	}
	return nil
}
