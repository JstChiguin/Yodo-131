package extraccion

import (
	"regexp"
	"strconv"
	"strings"

	"yodo131/nucleo/modelo"
)

var (
	patronInicioInforme   = regexp.MustCompile(`(?i)^NOMBRE\s*:`)
	patronLineaIgnorada   = regexp.MustCompile(`(?i)^(MINISTERIO DE SALUD|SERVICIO DE SALUD|CENTRO DE COSTO|HOSPITAL DR|HOSPITAL CARLOS|ENDOCRINOLOG[IÍ]A$|\.+$)`)
	patronLineaEstruct    = regexp.MustCompile(`(?i)^(Ecotomograf|ECO\b|PAAF|[UÚ]ltima TSH|TSH\b|Cintigrama|Biopsia|Antecedentes\s+(familiares|de patolog|patolog)|M[oó]rbidos|Estadificaci|M[eé]dico tratante|TAC\b|Disecci|Laboratorio|\d{1,2}/\d{1,2}/\d{2,4}\s*:)`)
	patronMes             = regexp.MustCompile(`(?i)^\s*(de\s+)?(enero|febrero|marzo|abril|mayo|junio|julio|agosto|septiembre|octubre|noviembre|diciembre)`)
	patronNombre          = regexp.MustCompile(`(?i)^NOMBRE\s*:`)
	patronRut             = regexp.MustCompile(`(?i)^RUT\s*:`)
	patronEdad            = regexp.MustCompile(`(?i)^EDAD\s*:`)
	patronFecha           = regexp.MustCompile(`(?i)^FECHA\s*:`)
	patronMorbidosLinea   = regexp.MustCompile(`(?i)^M[oó]rbidos`)
	patronAntecOMorbidos  = regexp.MustCompile(`(?i)^M[oó]rbidos|^Antecedentes\s+(familiares|de patolog|patolog)`)
	patronEco             = regexp.MustCompile(`(?i)Ecotomograf[ií]a[^:]{0,30}:|^ECO\b`)
	patronTirads          = regexp.MustCompile(`(?i)ti-?\s?rads\s*(\d)`)
	patronBethesda        = regexp.MustCompile(`(?i)Be\w{0,3}sda\s*([IVX]{1,3}|[1-6])\b`)
	patronEstadificacion  = regexp.MustCompile(`(?i)^Estadificaci[oó]n\s*:`)
	patronRiesgo          = regexp.MustCompile(`(?i)riesgo\s+(bajo|intermedio|alto)|(bajo|intermedio|alto)\s+riesgo`)
	patronPaaf            = regexp.MustCompile(`(?i)^PAAF[^:]*:`)
	patronBiopsiaLinea    = regexp.MustCompile(`(?i)^Biopsia`)
	patronCintigrama      = regexp.MustCompile(`(?i)Cintigrama[^:]*:`)
	patronTsh             = regexp.MustCompile(`(?i)\bTSH\b\s*[:=]?`)
	patronT4l             = regexp.MustCompile(`(?i)\bT4\s*L\b\s*[:=]`)
	patronT3              = regexp.MustCompile(`(?i)\bT3\b\s*[:=]`)
	patronTrab            = regexp.MustCompile(`(?i)\bTra[bB]\b\s*[:=]?`)
	patronAntiTg          = regexp.MustCompile(`(?i)anti\s*-?\s*tiroglobulina|anti\s*-?\s*tg\b`)
	patronAntiTpo         = regexp.MustCompile(`(?i)anti\s*-?\s*tpo|\batpo\b|anti\s*-?\s*peroxidasa`)
	patronAntecFamiliar   = regexp.MustCompile(`(?i)^Antecedentes[^:]*(familia|tiroidea)[^:]*:`)
	patronCigarrillo      = regexp.MustCompile(`(?i)Cigarrillo[^:]*:`)
	patronOh              = regexp.MustCompile(`(?i)\bOH\s*:`)
	patronCigarrilloYOh   = regexp.MustCompile(`(?i)Cigarrillo y\s+OH\s*:`)
	patronDrogas          = regexp.MustCompile(`(?i)Drogas\s*:`)
	patronMedicoTratante  = regexp.MustCompile(`(?i)^M[eé]dico tratante\s*:`)
	patronGineco          = regexp.MustCompile(`(?i)^(Gesta|G)\s*:|^FUR\s*:|^MAC\s*:|^TRH\s*:`)
	patronRechazoYodo     = regexp.MustCompile(`(?i)rechaza\s+(la\s+)?yodo|no acude a recibirlo`)
	patronValorLab        = regexp.MustCompile(`(?i)^[^0-9<>\n;]{0,30}?(<|>|menos de|menor a|mayor a)?\s*(\d+(?:[.,]\d+)?)`)
	patronComparadorMenor = regexp.MustCompile(`(?i)^(<|menos|menor)`)
	patronCorteAnticuerpo = regexp.MustCompile(`(?i);|\n|\bTSH\b`)
	patronSinAnticuerpo   = regexp.MustCompile(`(?i)no aparece|sin dato|pendiente`)
	patronValorAnticuerpo = regexp.MustCompile(`(?i)(menos de|menor a|<|>|mayor a)?\s*(\d+(?:[.,]\d+)?)`)
	patronCorteTg         = regexp.MustCompile(`(?i)anticuerpo|antitiroglobulina|\bac\b|\bTSH\b|;|\n|\.\s`)
	patronAscenso         = regexp.MustCompile(`(?i)ascenso|aumenta|elevad`)
	patronNumeroTg        = regexp.MustCompile(`(\d+(?:[.,]\d+)?)`)
	patronAnio            = regexp.MustCompile(`^(19|20)\d\d$`)
	patronEmpiezaBarra    = regexp.MustCompile(`^/`)
	patronTerminaBarra    = regexp.MustCompile(`/$`)
	patronElAntes         = regexp.MustCompile(`(?i)\bel\s$`)
	patronNodulos         = regexp.MustCompile(`(?i)n[oó]dul|imagen\s+s[oó]lida|lesi[oó]n\s+nodular|masa\s+s[oó]lida`)
	patronSinNodulos      = regexp.MustCompile(`(?i)sin\s+n[oó]dul\w*`)
	patronAdenopatias     = regexp.MustCompile(`(?i)adenopat|linfonodos?\s+(sospechos|prominent|aumentad)`)
	patronSinAdenopatias  = regexp.MustCompile(`(?i)sin\s+adenopat\w*`)
	patronDecide          = regexp.MustCompile(`(?i)decid`)
	patronErrorOeneDosis  = regexp.MustCompile(`(?i)(\d)[oO](\s*m\s*c)`)
	patronDosis           = regexp.MustCompile(`(?i)(\d{1,3})\s*m\s*C(?:i)?\b`)
	patronMorbidosInicio  = regexp.MustCompile(`(?i)M[oó]rbidos\s*:\s*`)
	patronCorteMorbidos   = regexp.MustCompile(`(?i);\s*(Alergia|Cigarrillo|OH)`)
	patronCampoEnLinea    = regexp.MustCompile(`(?i);|\n|\bOH\s*:|Drogas\s*:|Alergias?\s*:`)
	patronSoloGuiones     = regexp.MustCompile(`^-+$`)
	patronPresenciaSi     = regexp.MustCompile(`(?i)\d|cajetilla|ocasional|hasta|fum|exagerad|exceso|moderad|diario|semana|leve|marihuana|coca[ií]na`)
	patronPresenciaNo     = regexp.MustCompile(`(?i)^\(\s*-\s*\)|^no\b|^niega|^\(-`)
	patronRellenoPresencia = regexp.MustCompile(`[.…\s]`)
	patronGuionesTriples  = regexp.MustCompile(`-{3,}`)
	patronPrimerDigito13  = regexp.MustCompile(`\d{1,3}`)
)

func DividirEnInformes(parrafos []string) [][]string {
	indicesInicio := []int{}
	for indice, linea := range parrafos {
		if patronInicioInforme.MatchString(linea) {
			indicesInicio = append(indicesInicio, indice)
		}
	}
	informes := [][]string{}
	for posicion, inicio := range indicesInicio {
		fin := len(parrafos)
		if posicion+1 < len(indicesInicio) {
			fin = indicesInicio[posicion+1]
		}
		bloque := []string{}
		for _, linea := range parrafos[inicio:fin] {
			if !patronLineaIgnorada.MatchString(linea) {
				bloque = append(bloque, linea)
			}
		}
		informes = append(informes, bloque)
	}
	return informes
}

func textoTrasEtiqueta(lineas []string, patron *regexp.Regexp) any {
	for _, linea := range lineas {
		if loc := patron.FindStringIndex(linea); loc != nil {
			return recortar(desdeRuna(runas(linea), indiceRunaDesdeByte(linea, loc[1])))
		}
	}
	return nil
}

func comparadorDe(grupo string) string {
	if grupo == "" {
		return ""
	}
	if patronComparadorMenor.MatchString(grupo) {
		return "<"
	}
	return ">"
}

func extraerValorLaboratorio(texto string, patron *regexp.Regexp) any {
	runes := runas(texto)
	for _, loc := range patron.FindAllStringIndex(texto, -1) {
		inicio := indiceRunaDesdeByte(texto, loc[1])
		ventana := cortar(runes, inicio, inicio+40)
		if valor := patronValorLab.FindStringSubmatch(ventana); valor != nil {
			return nuevaMedicion(convertirNumeroValor(valor[2]), comparadorDe(valor[1]))
		}
	}
	return nil
}

func extraerAnticuerpo(texto string, patron *regexp.Regexp) any {
	loc := patron.FindStringIndex(texto)
	if loc == nil {
		return nil
	}
	runes := runas(texto)
	inicio := indiceRunaDesdeByte(texto, loc[1])
	ventana := cortar(runes, inicio, inicio+60)
	if corte := patronCorteAnticuerpo.FindStringIndex(ventana); corte != nil {
		ventana = cortar(runas(ventana), 0, indiceRunaDesdeByte(ventana, corte[0]))
	}
	if patronSinAnticuerpo.MatchString(ventana) {
		return nil
	}
	valor := patronValorAnticuerpo.FindStringSubmatch(ventana)
	if valor == nil {
		return nil
	}
	return nuevaMedicion(convertirNumeroValor(valor[2]), comparadorDe(valor[1]))
}

func extraerTiroglobulina(texto string) modelo.Tiroglobulina {
	runes := runas(texto)
	bajo := strings.ToLower(texto)
	bajoRunes := []rune(bajo)
	valores := []float64{}
	enAscenso := false
	mencionada := false
	for _, loc := range regexp.MustCompile(`(?i)tiroglobulina`).FindAllStringIndex(texto, -1) {
		inicioRuna := indiceRunaDesdeByte(texto, loc[0])
		if precedidoPorAnti(bajoRunes, inicioRuna) {
			continue
		}
		mencionada = true
		fin := indiceRunaDesdeByte(texto, loc[1])
		ventana := cortar(runes, fin, fin+100)
		if corte := patronCorteTg.FindStringIndex(ventana); corte != nil {
			ventana = cortar(runas(ventana), 0, indiceRunaDesdeByte(ventana, corte[0]))
		}
		if patronAscenso.MatchString(ventana) {
			enAscenso = true
		}
		ventanaRunes := runas(ventana)
		for _, numeroLoc := range patronNumeroTg.FindAllStringSubmatchIndex(ventana, -1) {
			inicioNum := indiceRunaDesdeByte(ventana, numeroLoc[0])
			finNum := indiceRunaDesdeByte(ventana, numeroLoc[1])
			textoNum := string(ventanaRunes[inicioNum:finNum])
			antes := cortar(ventanaRunes, inicioNum-4, inicioNum)
			despues := desdeRuna(ventanaRunes, finNum)
			esAnio := patronAnio.MatchString(textoNum)
			esFecha := patronEmpiezaBarra.MatchString(despues) || patronTerminaBarra.MatchString(antes) || patronElAntes.MatchString(antes) || patronMes.MatchString(despues)
			if !esAnio && !esFecha {
				if valor, ok := convertirNumero(textoNum); ok {
					valores = append(valores, valor)
				}
			}
		}
	}
	tg := modelo.Tiroglobulina{EnAscenso: enAscenso, Mencionada: mencionada}
	if len(valores) > 0 {
		maximo := valores[0]
		for _, valor := range valores[1:] {
			if valor > maximo {
				maximo = valor
			}
		}
		tg.ValorMaximo = maximo
	}
	return tg
}

func precedidoPorAnti(bajoRunes []rune, indice int) bool {
	tiene := func(inicio, fin int) string {
		if inicio < 0 || fin > len(bajoRunes) {
			return ""
		}
		return string(bajoRunes[inicio:fin])
	}
	if tiene(indice-4, indice) == "anti" {
		return true
	}
	if indice >= 5 && tiene(indice-5, indice-1) == "anti" {
		caracter := bajoRunes[indice-1]
		if caracter == ' ' || caracter == '\t' || caracter == '\n' || caracter == '\r' || caracter == '\f' || caracter == '\v' || caracter == ' ' {
			return true
		}
	}
	if tiene(indice-5, indice) == "anti-" {
		return true
	}
	return false
}

func evaluarPresencia(texto any) any {
	s, ok := texto.(string)
	if !ok {
		return nil
	}
	sinRelleno := patronRellenoPresencia.ReplaceAllString(s, "")
	if sinRelleno == "" || patronSoloGuiones.MatchString(sinRelleno) {
		return nil
	}
	if patronPresenciaSi.MatchString(s) {
		return true
	}
	if patronPresenciaNo.MatchString(sinRelleno) {
		return false
	}
	return true
}

func extraerCampoEnLinea(texto string, patron *regexp.Regexp) any {
	loc := patron.FindStringIndex(texto)
	if loc == nil {
		return nil
	}
	resto := desdeRuna(runas(texto), indiceRunaDesdeByte(texto, loc[1]))
	partes := patronCampoEnLinea.Split(resto, -1)
	return recortar(partes[0])
}

func extraerDosis(lineas []string) any {
	for _, linea := range lineas {
		if !patronDecide.MatchString(linea) {
			continue
		}
		normalizada := patronErrorOeneDosis.ReplaceAllString(linea, "${1}0${2}")
		if dosis := patronDosis.FindStringSubmatch(normalizada); dosis != nil {
			numero, _ := strconv.Atoi(dosis[1])
			return float64(numero)
		}
	}
	return nil
}

var romanosBethesda = map[string]string{"1": "I", "2": "II", "3": "III", "4": "IV", "5": "V", "6": "VI"}

func convertirBethesdaRomano(valor string) any {
	if len(valor) == 1 && valor[0] >= '0' && valor[0] <= '9' {
		if romano, ok := romanosBethesda[valor]; ok {
			return romano
		}
		return nil
	}
	return strings.ToUpper(valor)
}

func extraerMorbidos(linea string) any {
	loc := patronMorbidosInicio.FindStringIndex(linea)
	if loc == nil {
		return nil
	}
	resto := desdeRuna(runas(linea), indiceRunaDesdeByte(linea, loc[1]))
	if corte := patronCorteMorbidos.FindStringIndex(resto); corte != nil {
		resto = cortar(runas(resto), 0, indiceRunaDesdeByte(resto, corte[0]))
	}
	recortado := recortar(resto)
	if recortado == "" {
		return nil
	}
	return recortado
}

func grupoOpcional(coincidencia []string, indice int) string {
	if coincidencia == nil || indice >= len(coincidencia) {
		return ""
	}
	return coincidencia[indice]
}

func ExtraerInforme(lineas []string, nombreArchivo string) modelo.Objeto {
	textoCompleto := strings.Join(lineas, "\n")
	lineaMorbidos := ""
	for _, linea := range lineas {
		if patronMorbidosLinea.MatchString(linea) {
			lineaMorbidos = linea
			break
		}
	}
	lineasClinicas := []string{}
	for _, linea := range lineas {
		if !patronAntecOMorbidos.MatchString(linea) {
			lineasClinicas = append(lineasClinicas, linea)
		}
	}
	textoClinico := strings.Join(lineasClinicas, "\n")

	indiceFecha := -1
	for indice, linea := range lineas {
		if patronFecha.MatchString(linea) {
			indiceFecha = indice
			break
		}
	}
	lineasNarrativa := []string{}
	for indice := indiceFecha + 1; indice > 0 && indice < len(lineas); indice++ {
		if patronLineaEstruct.MatchString(lineas[indice]) {
			break
		}
		lineasNarrativa = append(lineasNarrativa, lineas[indice])
	}

	lineasEco := []string{}
	for _, linea := range lineas {
		if patronEco.MatchString(linea) {
			lineasEco = append(lineasEco, linea)
		}
	}
	textoEco := strings.Join(lineasEco, " ")

	var tirads any
	if m := patronTirads.FindStringSubmatch(textoEco); m != nil {
		n, _ := strconv.Atoi(m[1])
		tirads = n
	} else if m := patronTirads.FindStringSubmatch(textoClinico); m != nil {
		n, _ := strconv.Atoi(m[1])
		tirads = n
	}

	var bethesda any
	if m := patronBethesda.FindStringSubmatch(textoClinico); m != nil {
		bethesda = convertirBethesdaRomano(m[1])
	}

	var riesgo any
	if m := patronRiesgo.FindStringSubmatch(textoClinico); m != nil {
		valor := m[1]
		if valor == "" {
			valor = m[2]
		}
		riesgo = strings.ToLower(valor)
	}

	narrativaUnida := strings.Join(lineasNarrativa, " ")

	o := modelo.Nuevo()
	o.Set("origen", "Word")
	o.Set("archivo", nombreArchivo)
	o.Set("referencia", nombreArchivo)

	nombre := textoTrasEtiqueta(lineas, patronNombre)
	nombreTexto := ""
	if s, ok := nombre.(string); ok {
		nombreTexto = s
	}
	o.Set("nombrePaciente", capitalizarNombre(nombreTexto))

	lineasRut := lineas
	if indiceFecha >= 0 {
		lineasRut = lineas[:indiceFecha+1]
	} else if len(lineas) > 4 {
		lineasRut = lineas[:4]
	}
	rutTexto := ""
	if s, ok := textoTrasEtiqueta(lineasRut, patronRut).(string); ok {
		rutTexto = s
	}
	o.Set("rut", normalizarRut(rutTexto))

	var edad any
	if s, ok := textoTrasEtiqueta(lineas, patronEdad).(string); ok {
		if m := patronPrimerDigito13.FindString(s); m != "" {
			if n, err := strconv.Atoi(m); err == nil {
				edad = n
			}
		}
	}
	o.Set("edad", edad)

	var fechaEtiqueta string
	if s, ok := textoTrasEtiqueta(lineas, patronFecha).(string); ok {
		fechaEtiqueta = s
	}
	fechaIso, anio, mes := interpretarFecha(fechaEtiqueta)
	o.Set("fechaIso", fechaIso)
	o.Set("anio", anio)
	o.Set("mes", mes)

	o.Set("textoMotivo", recortar(patronGuionesTriples.ReplaceAllString(narrativaUnida, " ")))
	o.Set("textoEco", textoEco)
	o.Set("tirads", tirads)

	if len(lineasEco) > 0 {
		o.Set("ecoConNodulos", patronNodulos.MatchString(patronSinNodulos.ReplaceAllString(textoEco, "")))
		o.Set("ecoConAdenopatias", patronAdenopatias.MatchString(patronSinAdenopatias.ReplaceAllString(textoEco, "")))
	} else {
		o.Set("ecoConNodulos", nil)
		o.Set("ecoConAdenopatias", nil)
	}
	o.Set("bethesda", bethesda)
	o.Set("textoPaaf", textoTrasEtiqueta(lineas, patronPaaf))

	biopsia := []string{}
	for _, linea := range lineas {
		if patronBiopsiaLinea.MatchString(linea) {
			biopsia = append(biopsia, linea)
		}
	}
	o.Set("textoBiopsia", strings.Join(biopsia, " "))
	o.Set("textoCintigrama", textoTrasEtiqueta(lineas, patronCintigrama))
	o.Set("estadificacion", textoTrasEtiqueta(lineas, patronEstadificacion))
	o.Set("riesgo", riesgo)
	o.Set("tsh", extraerValorLaboratorio(textoClinico, patronTsh))
	o.Set("t4l", extraerValorLaboratorio(textoClinico, patronT4l))
	o.Set("t3", extraerValorLaboratorio(textoClinico, patronT3))
	o.Set("trab", extraerValorLaboratorio(textoClinico, patronTrab))
	o.Set("tiroglobulina", extraerTiroglobulina(textoClinico))
	o.Set("antiTiroglobulina", extraerAnticuerpo(textoClinico, patronAntiTg))
	o.Set("antiTpo", extraerAnticuerpo(textoClinico, patronAntiTpo))
	o.Set("textoAntecedentesFamiliares", textoTrasEtiqueta(lineas, patronAntecFamiliar))

	if lineaMorbidos != "" {
		o.Set("textoMorbidos", extraerMorbidos(lineaMorbidos))
	} else {
		o.Set("textoMorbidos", nil)
	}

	o.Set("textoCigarrillo", extraerCampoEnLinea(textoCompleto, patronCigarrillo))
	alcohol := extraerCampoEnLinea(textoCompleto, patronOh)
	if alcohol == nil && patronCigarrilloYOh.MatchString(textoCompleto) {
		alcohol = extraerCampoEnLinea(textoCompleto, patronCigarrillo)
	}
	o.Set("textoAlcohol", alcohol)
	o.Set("textoDrogas", extraerCampoEnLinea(textoCompleto, patronDrogas))
	o.Set("textoMedicoTratante", textoTrasEtiqueta(lineas, patronMedicoTratante))

	gineco := false
	for _, linea := range lineas {
		if patronGineco.MatchString(linea) {
			gineco = true
			break
		}
	}
	o.Set("tieneDatosGinecologicos", gineco)
	o.Set("dosisMci", extraerDosis(lineas))
	o.Set("rechazoYodo", patronRechazoYodo.MatchString(textoCompleto))
	o.Set("textoClinico", textoClinico)
	o.Set("textoNarrativa", narrativaUnida)
	o.Set("indiceEnArchivo", nil)
	return o
}

func EvaluarPresencia(texto any) any {
	return evaluarPresencia(texto)
}
