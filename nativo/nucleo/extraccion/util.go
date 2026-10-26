package extraccion

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"yodo131/nucleo/modelo"
)

func runas(texto string) []rune {
	return []rune(texto)
}

func cortar(runes []rune, inicio, fin int) string {
	if inicio < 0 {
		inicio = 0
	}
	if inicio > len(runes) {
		inicio = len(runes)
	}
	if fin > len(runes) {
		fin = len(runes)
	}
	if fin < inicio {
		fin = inicio
	}
	return string(runes[inicio:fin])
}

func desdeRuna(runes []rune, inicio int) string {
	return cortar(runes, inicio, len(runes))
}

func indiceRunaDesdeByte(texto string, indiceByte int) int {
	return utf8.RuneCountInString(texto[:indiceByte])
}

func recortar(texto string) string {
	return strings.TrimSpace(texto)
}

func convertirNumero(texto string) (float64, bool) {
	limpio := strings.TrimSpace(strings.Replace(texto, ",", ".", 1))
	if limpio == "" {
		return 0, true
	}
	numero, err := strconv.ParseFloat(limpio, 64)
	if err != nil {
		return 0, false
	}
	return numero, true
}

func convertirNumeroValor(texto string) any {
	if numero, ok := convertirNumero(texto); ok {
		return numero
	}
	return nil
}

var noRut = regexp.MustCompile(`[^0-9kK]`)

func normalizarRut(texto string) any {
	if texto == "" {
		return nil
	}
	limpio := strings.ToUpper(noRut.ReplaceAllString(texto, ""))
	if len([]rune(limpio)) < 7 {
		return nil
	}
	r := []rune(limpio)
	return string(r[:len(r)-1]) + "-" + string(r[len(r)-1])
}

var (
	espaciosMultiples = regexp.MustCompile(`\s+`)
	inicialPalabra    = regexp.MustCompile(`(^|\s)(\p{L})`)
)

func capitalizarNombre(nombre string) string {
	base := recortar(espaciosMultiples.ReplaceAllString(strings.ToLower(nombre), " "))
	return inicialPalabra.ReplaceAllStringFunc(base, func(coincidencia string) string {
		r := []rune(coincidencia)
		r[len(r)-1] = unicode.ToUpper(r[len(r)-1])
		return string(r)
	})
}

var noDigito = regexp.MustCompile(`\D`)

func interpretarFecha(texto string) (fechaIso any, anio any, mes any) {
	if texto == "" {
		return nil, nil, nil
	}
	partes := strings.Split(espaciosMultiples.ReplaceAllString(texto, ""), "/")
	parte := func(indice int) string {
		if indice < len(partes) {
			return noDigito.ReplaceAllString(partes[indice], "")
		}
		return ""
	}
	numero := func(s string) int {
		if s == "" {
			return 0
		}
		n, _ := strconv.Atoi(s)
		return n
	}
	a := numero(parte(2))
	m := numero(parte(1))
	d := numero(parte(0))
	if a >= 1990 && a <= 2100 {
		anio = a
	}
	if anio != nil && m >= 1 && m <= 12 {
		mes = m
	}
	if mes != nil && d >= 1 && d <= 31 {
		fechaIso = a2(a) + "-" + dosDigitos(m) + "-" + dosDigitos(d)
	}
	return fechaIso, anio, mes
}

func a2(n int) string { return strconv.Itoa(n) }

func dosDigitos(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

func medicionNula() any { return nil }

func nuevaMedicion(valor any, comparador string) any {
	if valor == nil {
		return nil
	}
	return modelo.Medicion{Valor: valor.(float64), Comparador: comparador}
}
