//go:build !windows

package dialogos

import "errors"

var ErrNoSoportado = errors.New("los diálogos nativos solo están disponibles en Windows")

func ElegirCarpeta(titulo string) (string, bool) {
	return "", false
}

func GuardarComo(titulo, nombreSugerido, carpetaInicial string) (string, bool) {
	return "", false
}
