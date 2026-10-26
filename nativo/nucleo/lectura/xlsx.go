package lectura

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"yodo131/nucleo/modelo"
)

type Hoja struct {
	Nombre string
	Filas  [][]any
}

var formatosFechaIncorporados = map[int]bool{
	14: true, 15: true, 16: true, 17: true, 18: true, 19: true, 20: true, 21: true, 22: true,
	27: true, 28: true, 29: true, 30: true, 31: true, 32: true, 33: true, 34: true, 35: true, 36: true,
	45: true, 46: true, 47: true, 50: true, 51: true, 52: true, 53: true, 54: true, 55: true, 56: true, 57: true, 58: true,
}

const milisegundosPorDia = 86400000.0

var (
	atributoXml        = regexp.MustCompile(`([\w:]+)="([^"]*)"`)
	bloqueRph          = regexp.MustCompile(`<rPh\b[\s\S]*?</rPh>`)
	piezaTextoRico     = regexp.MustCompile(`<t(?:\s[^>]*)?>([^<]*)</t>`)
	bloqueSi           = regexp.MustCompile(`<si>([\s\S]*?)</si>`)
	numeroFormatoPropio = regexp.MustCompile(`<numFmt\b([^>]*)/?>`)
	bloqueCellXfs      = regexp.MustCompile(`<cellXfs\b[\s\S]*?</cellXfs>`)
	entradaXf          = regexp.MustCompile(`<xf\b([^>]*?)/?>`)
	filaXml            = regexp.MustCompile(`<row\b([^>]*)>([\s\S]*?)</row>`)
	celdaXml           = regexp.MustCompile(`<c\b([^>]*?)(?:/>|>([\s\S]*?)</c>)`)
	valorCeldaXml      = regexp.MustCompile(`<v>([^<]*)</v>`)
	comillasFormato    = regexp.MustCompile(`"[^"]*"`)
	corchetesFormato   = regexp.MustCompile(`\[[^\]]*\]`)
	tieneDiaAnio       = regexp.MustCompile(`[dy]`)
	tieneMes           = regexp.MustCompile(`m`)
	tieneHoraSegundo   = regexp.MustCompile(`[hs]`)
	soloDigitos        = regexp.MustCompile(`\d`)
	date1904           = regexp.MustCompile(`<workbookPr\b[^>]*date1904="(1|true)"`)
	hojaXmlRegex       = regexp.MustCompile(`<sheet\b([^>]*)/?>`)
	relacionXml        = regexp.MustCompile(`<Relationship\b([^>]*)/?>`)
)

func leerAtributos(texto string) map[string]string {
	atributos := map[string]string{}
	for _, par := range atributoXml.FindAllStringSubmatch(texto, -1) {
		atributos[par[1]] = DecodificarEntidades(par[2])
	}
	return atributos
}

func extraerTextoRico(fragmento string) string {
	sinFonetica := bloqueRph.ReplaceAllString(fragmento, "")
	var constructor strings.Builder
	for _, pieza := range piezaTextoRico.FindAllStringSubmatch(sinFonetica, -1) {
		constructor.WriteString(DecodificarEntidades(pieza[1]))
	}
	return constructor.String()
}

func leerTextosCompartidos(xml string) []string {
	textos := []string{}
	for _, si := range bloqueSi.FindAllStringSubmatch(xml, -1) {
		textos = append(textos, extraerTextoRico(si[1]))
	}
	return textos
}

func esCodigoFormatoFecha(codigo string) bool {
	limpio := strings.ToLower(corchetesFormato.ReplaceAllString(comillasFormato.ReplaceAllString(codigo, ""), ""))
	return tieneDiaAnio.MatchString(limpio) || (tieneMes.MatchString(limpio) && !tieneHoraSegundo.MatchString(limpio))
}

func leerEstilosFecha(xml string) map[int]bool {
	estilos := map[int]bool{}
	if xml == "" {
		return estilos
	}
	formatosPropios := map[int]string{}
	for _, propio := range numeroFormatoPropio.FindAllStringSubmatch(xml, -1) {
		atributos := leerAtributos(propio[1])
		id, _ := strconv.Atoi(atributos["numFmtId"])
		formatosPropios[id] = atributos["formatCode"]
	}
	bloque := bloqueCellXfs.FindString(xml)
	for indice, xf := range entradaXf.FindAllStringSubmatch(bloque, -1) {
		atributos := leerAtributos(xf[1])
		id := 0
		if texto, ok := atributos["numFmtId"]; ok {
			id, _ = strconv.Atoi(texto)
		}
		codigo, existe := formatosPropios[id]
		if formatosFechaIncorporados[id] || (existe && esCodigoFormatoFecha(codigo)) {
			estilos[indice] = true
		}
	}
	return estilos
}

func convertirSerialAFecha(serial float64, usaSistema1904 bool) modelo.Fecha {
	origen := time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)
	if usaSistema1904 {
		origen = time.Date(1904, time.January, 1, 0, 0, 0, 0, time.UTC)
	}
	milisegundos := math.Floor(serial*milisegundosPorDia + 0.5)
	fecha := origen.Add(time.Duration(milisegundos) * time.Millisecond)
	return modelo.Fecha{Iso: fecha.UTC().Format("2006-01-02")}
}

func indiceColumna(referencia string) int {
	letras := soloDigitos.ReplaceAllString(referencia, "")
	indice := 0
	for _, letra := range letras {
		indice = indice*26 + int(letra-'A'+1)
	}
	return indice - 1
}

func leerNumero(texto string) (float64, bool) {
	valor, err := strconv.ParseFloat(strings.TrimSpace(texto), 64)
	if err != nil {
		return 0, false
	}
	return valor, true
}

func leerHoja(xml string, textosCompartidos []string, estilosFecha map[int]bool, usaSistema1904 bool) [][]any {
	filasPorNumero := map[int][]any{}
	maximaFila := -1
	for _, coincidenciaFila := range filaXml.FindAllStringSubmatch(xml, -1) {
		atributosFila := leerAtributos(coincidenciaFila[1])
		numeroFila, err := strconv.Atoi(atributosFila["r"])
		if err != nil {
			continue
		}
		numeroFila--
		fila := []any{}
		ponerEnColumna := func(indice int, valor any) {
			if indice < 0 {
				return
			}
			for len(fila) <= indice {
				fila = append(fila, nil)
			}
			fila[indice] = valor
		}
		for _, coincidenciaCelda := range celdaXml.FindAllStringSubmatch(coincidenciaFila[2], -1) {
			atributosCelda := leerAtributos(coincidenciaCelda[1])
			contenidoCelda := coincidenciaCelda[2]
			tipo := atributosCelda["t"]
			var crudo string
			var hayCrudo bool
			if m := valorCeldaXml.FindStringSubmatch(contenidoCelda); m != nil {
				crudo, hayCrudo = m[1], true
			}
			var valor any
			switch {
			case tipo == "s":
				indice, _ := strconv.Atoi(crudo)
				if indice >= 0 && indice < len(textosCompartidos) {
					valor = textosCompartidos[indice]
				}
			case tipo == "inlineStr":
				valor = extraerTextoRico(contenidoCelda)
			case tipo == "str" || tipo == "e":
				if hayCrudo {
					valor = DecodificarEntidades(crudo)
				}
			case tipo == "b":
				valor = crudo == "1"
			default:
				if hayCrudo && crudo != "" {
					numero, ok := leerNumero(crudo)
					if ok {
						estilo := -1
						if texto, existe := atributosCelda["s"]; existe {
							estilo, _ = strconv.Atoi(texto)
						}
						if estilosFecha[estilo] {
							valor = convertirSerialAFecha(numero, usaSistema1904)
						} else {
							valor = numero
						}
					}
				}
			}
			if valor != nil && valor != "" {
				ponerEnColumna(indiceColumna(atributosCelda["r"]), valor)
			}
		}
		filasPorNumero[numeroFila] = fila
		if numeroFila > maximaFila {
			maximaFila = numeroFila
		}
	}
	filas := make([][]any, maximaFila+1)
	for indice := range filas {
		if fila, existe := filasPorNumero[indice]; existe {
			filas[indice] = fila
		} else {
			filas[indice] = []any{}
		}
	}
	return filas
}

func resolverRutaHoja(destino string) string {
	if strings.HasPrefix(destino, "/") {
		return destino[1:]
	}
	return "xl/" + strings.TrimPrefix(destino, "./")
}

func LeerLibro(datos []byte) ([]Hoja, error) {
	if EsFormatoOfficeAntiguo(datos) {
		return nil, &FormatoNoSoportadoError{Mensaje: `Es un Excel antiguo (.xls). Ábralo en Excel y use "Guardar como" .xlsx`}
	}
	lector, err := abrirZip(datos)
	if err != nil {
		return nil, err
	}
	leerParte := func(nombre string, opcional bool) (string, error) {
		contenido, existe, err := leerEntrada(lector, nombre)
		if err != nil {
			return "", err
		}
		if !existe && !opcional {
			return "", &FormatoNoSoportadoError{Mensaje: "Falta la parte interna " + nombre}
		}
		return contenido, nil
	}
	xmlLibro, err := leerParte("xl/workbook.xml", false)
	if err != nil {
		return nil, err
	}
	xmlRelaciones, err := leerParte("xl/_rels/workbook.xml.rels", false)
	if err != nil {
		return nil, err
	}
	xmlCompartidos, _ := leerParte("xl/sharedStrings.xml", true)
	xmlEstilos, _ := leerParte("xl/styles.xml", true)
	textosCompartidos := leerTextosCompartidos(xmlCompartidos)
	estilosFecha := leerEstilosFecha(xmlEstilos)
	usaSistema1904 := date1904.MatchString(xmlLibro)

	destinos := map[string]string{}
	for _, relacion := range relacionXml.FindAllStringSubmatch(xmlRelaciones, -1) {
		atributos := leerAtributos(relacion[1])
		destinos[atributos["Id"]] = atributos["Target"]
	}

	hojas := []Hoja{}
	for _, coincidencia := range hojaXmlRegex.FindAllStringSubmatch(xmlLibro, -1) {
		atributos := leerAtributos(coincidencia[1])
		destino, existe := destinos[atributos["r:id"]]
		if !existe {
			continue
		}
		xmlHoja, err := leerParte(resolverRutaHoja(destino), true)
		if err != nil {
			return nil, err
		}
		if xmlHoja != "" {
			hojas = append(hojas, Hoja{Nombre: atributos["name"], Filas: leerHoja(xmlHoja, textosCompartidos, estilosFecha, usaSistema1904)})
		}
	}
	return hojas, nil
}
