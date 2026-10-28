package vista

import (
	"regexp"
	"strconv"
	"strings"

	"yodo131/nucleo/estadisticas"
)

var patronSingular = regexp.MustCompile(`^(\S+?)s\b`)

func formatoNumero(n int) string {
	negativo := n < 0
	if negativo {
		n = -n
	}
	digitos := strconv.Itoa(n)
	var partes []string
	for len(digitos) > 3 {
		partes = append([]string{digitos[len(digitos)-3:]}, partes...)
		digitos = digitos[:len(digitos)-3]
	}
	partes = append([]string{digitos}, partes...)
	resultado := strings.Join(partes, ".")
	if negativo {
		return "-" + resultado
	}
	return resultado
}

func porcentajeTexto(cantidad, total int) string {
	if total == 0 {
		return "—"
	}
	valor := float64(cantidad) / float64(total) * 100
	redondeado := int(valor + 0.5)
	return strconv.Itoa(redondeado) + " %"
}

func fraccion(cantidad, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(cantidad) / float64(total)
}

func singular(unidad string) string {
	return patronSingular.ReplaceAllString(unidad, "$1")
}

func describirCantidad(cantidad int, unidad string) string {
	u := unidad
	if cantidad == 1 {
		u = singular(unidad)
	}
	return formatoNumero(cantidad) + " " + u
}

const unidadResumen = "tratamientos"

func describirCobertura(conDato, total int, unidad string) string {
	if total == 0 {
		return ""
	}
	if conDato == total {
		return "Dato disponible en todos los " + unidad + " (" + formatoNumero(total) + ")."
	}
	return "Dato disponible en " + formatoNumero(conDato) + " de " + formatoNumero(total) + " " + unidad + " (" + porcentajeTexto(conDato, total) + "). Sin dato: " + formatoNumero(total-conDato) + " (" + porcentajeTexto(total-conDato, total) + "). Los porcentajes del gráfico se calculan sobre los " + formatoNumero(conDato) + " con dato."
}

type masFrecuente struct {
	categoria  string
	detalle    string
	porcentaje string
}

func obtenerMasFrecuente(dist estadisticas.Distribucion, excluir []string) (masFrecuente, bool) {
	var mayor estadisticas.FilaDist
	encontrado := false
	for _, fila := range dist.Filas {
		if contieneTexto(excluir, fila.Categoria) {
			continue
		}
		if !encontrado || fila.Cantidad > mayor.Cantidad {
			mayor = fila
			encontrado = true
		}
	}
	if !encontrado {
		return masFrecuente{}, false
	}
	return masFrecuente{
		categoria:  mayor.Categoria,
		detalle:    formatoNumero(mayor.Cantidad) + " de " + formatoNumero(dist.ConDato),
		porcentaje: porcentajeTexto(mayor.Cantidad, dist.ConDato),
	}, true
}

func describirMasFrecuente(dist estadisticas.Distribucion, introduccion, unidadConDato string, excluir []string) string {
	mayor, ok := obtenerMasFrecuente(dist, excluir)
	if !ok {
		return ""
	}
	return introduccion + ": " + mayor.categoria + ", con " + mayor.detalle + " " + unidadConDato + " (" + mayor.porcentaje + ")."
}

func sumarCategorias(dist estadisticas.Distribucion, condicion func(estadisticas.FilaDist) bool) int {
	suma := 0
	for _, fila := range dist.Filas {
		if condicion(fila) {
			suma += fila.Cantidad
		}
	}
	return suma
}

func contieneTexto(lista []string, objetivo string) bool {
	for _, elemento := range lista {
		if elemento == objetivo {
			return true
		}
	}
	return false
}
