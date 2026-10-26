package lectura

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type FormatoNoSoportadoError struct{ Mensaje string }

func (e *FormatoNoSoportadoError) Error() string { return e.Mensaje }

func EsFormatoOfficeAntiguo(datos []byte) bool {
	return len(datos) >= 4 && datos[0] == 0xd0 && datos[1] == 0xcf && datos[2] == 0x11 && datos[3] == 0xe0
}

func abrirZip(datos []byte) (*zip.Reader, error) {
	lector, err := zip.NewReader(bytes.NewReader(datos), int64(len(datos)))
	if err != nil {
		return nil, &FormatoNoSoportadoError{Mensaje: "El archivo está dañado o no es un documento de Office válido"}
	}
	return lector, nil
}

func leerEntrada(lector *zip.Reader, nombre string) (string, bool, error) {
	for _, archivo := range lector.File {
		if archivo.Name == nombre {
			flujo, err := archivo.Open()
			if err != nil {
				return "", false, &FormatoNoSoportadoError{Mensaje: "La estructura interna del archivo está dañada"}
			}
			defer flujo.Close()
			contenido, err := io.ReadAll(flujo)
			if err != nil {
				return "", false, &FormatoNoSoportadoError{Mensaje: "La estructura interna del archivo está dañada"}
			}
			return string(contenido), true, nil
		}
	}
	return "", false, nil
}

var (
	entidadDecimal     = regexp.MustCompile(`&#(\d+);`)
	entidadHexadecimal = regexp.MustCompile(`&#x([0-9a-fA-F]+);`)
)

func DecodificarEntidades(texto string) string {
	texto = strings.ReplaceAll(texto, "&lt;", "<")
	texto = strings.ReplaceAll(texto, "&gt;", ">")
	texto = strings.ReplaceAll(texto, "&quot;", "\"")
	texto = strings.ReplaceAll(texto, "&apos;", "'")
	texto = entidadDecimal.ReplaceAllStringFunc(texto, func(coincidencia string) string {
		digitos := entidadDecimal.FindStringSubmatch(coincidencia)[1]
		codigo, err := strconv.Atoi(digitos)
		if err != nil {
			return coincidencia
		}
		return string(rune(codigo))
	})
	texto = entidadHexadecimal.ReplaceAllStringFunc(texto, func(coincidencia string) string {
		hex := entidadHexadecimal.FindStringSubmatch(coincidencia)[1]
		codigo, err := strconv.ParseInt(hex, 16, 32)
		if err != nil {
			return coincidencia
		}
		return string(rune(codigo))
	})
	return strings.ReplaceAll(texto, "&amp;", "&")
}

var errNoEsZip = errors.New("no es zip")
