package extraccion

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"yodo131/nucleo/modelo"
	textoutil "yodo131/nucleo/texto"
)

var (
	patronRutExcel     = regexp.MustCompile(`^\s*\d{1,2}\.?\d{2,3}\.?\d{3}\s*[-.]\s*[\dkK]\s*$`)
	patronMedicoExcel  = regexp.MustCompile(`(?i)^\s*dra?\b\.?\s*\S`)
	patronDiagExcel    = regexp.MustCompile(`(?i)carcinoma\s+(papilar|folicular|medular|anapl)|c[aá]ncer\s+(de(l)?\s+)?tiroid|^\s*ca\.?\s|hipertiroid|bocio|tiroiditis|graves|rastreo`)
	patronC73          = regexp.MustCompile(`(?i)^C73`)
	patronAnio4        = regexp.MustCompile(`^\d{4}$`)
	patronSoloDigitos  = regexp.MustCompile(`^\d+$`)
	patronTieneDigito  = regexp.MustCompile(`\d`)
	patronFechaTexto   = regexp.MustCompile(`\d{1,2}\s*/\s*\d{1,2}\s*/\s*\d{4}`)
	patronRastreo      = regexp.MustCompile(`(?i)rastreo`)
	patronEspacios     = regexp.MustCompile(`\s+`)
	patronCombinantes  = regexp.MustCompile("[̀-ͯ]")
)

var encabezadosYodo = []struct {
	campo    string
	nombres  []string
}{
	{"paciente", []string{"PACIENTE"}},
	{"rut", []string{"RUT"}},
	{"edad", []string{"EDAD"}},
	{"sexo", []string{"F/M", "M/F", "SEXO"}},
	{"fechaSolicitud", []string{"FECHA DE SOLICITUD"}},
	{"solicitudEnviada", []string{"SOLICITUD ENVIADA"}},
	{"diagnostico", []string{"DIAGNOSTICO"}},
	{"medico", []string{"MEDICO TRATANTE"}},
	{"dosis", []string{"DOSIS"}},
	{"antecedentesFamiliares", []string{"ANT.FAMILIARES", "ANT. FAMILIARES", "ANTECEDENTES FAMILIARES"}},
	{"morbidos", []string{"MORBIDOS"}},
	{"cigarrillo", []string{"CIGARRILLO"}},
	{"alcohol", []string{"OH"}},
	{"drogas", []string{"DROGAS"}},
}

func jsNumero(valor float64) string {
	if math.IsInf(valor, 0) || math.IsNaN(valor) {
		return ""
	}
	return strconv.FormatFloat(valor, 'f', -1, 64)
}

func comoTexto(valor any) string {
	switch v := valor.(type) {
	case nil:
		return ""
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		return jsNumero(v)
	case int:
		return strconv.Itoa(v)
	case modelo.Fecha:
		return "[object Object]"
	}
	return ""
}

func normalizarEncabezado(valor any) string {
	texto := comoTexto(valor)
	texto = patronCombinantes.ReplaceAllString(textoutil.QuitarDiacriticos(texto), "")
	texto = strings.ToUpper(texto)
	return recortar(patronEspacios.ReplaceAllString(texto, " "))
}

func esFecha(valor any) bool {
	_, ok := valor.(modelo.Fecha)
	return ok
}

func esTexto(valor any) bool {
	s, ok := valor.(string)
	return ok && strings.TrimSpace(s) != ""
}

func textoCelda(valor any) any {
	if valor == nil {
		return nil
	}
	if fecha, ok := valor.(modelo.Fecha); ok {
		return fecha.Iso
	}
	texto := recortar(patronEspacios.ReplaceAllString(comoTexto(valor), " "))
	if texto == "" {
		return nil
	}
	return texto
}

type triFecha struct {
	fechaIso any
	anio     any
	mes      any
}

func obtenerFecha(valor any) (triFecha, bool) {
	if fecha, ok := valor.(modelo.Fecha); ok {
		partes := strings.Split(fecha.Iso, "-")
		anio, _ := strconv.Atoi(partes[0])
		mes, _ := strconv.Atoi(partes[1])
		if anio >= 1990 && anio <= 2100 {
			return triFecha{fechaIso: fecha.Iso, anio: anio, mes: mes}, true
		}
		return triFecha{}, false
	}
	if esTexto(valor) && patronFechaTexto.MatchString(valor.(string)) {
		fechaIso, anio, mes := interpretarFecha(valor.(string))
		return triFecha{fechaIso: fechaIso, anio: anio, mes: mes}, true
	}
	return triFecha{}, false
}

func celda(fila []any, indice int) any {
	if indice >= 0 && indice < len(fila) {
		return fila[indice]
	}
	return nil
}

func buscarFilaEncabezado(filas [][]any) int {
	limite := len(filas)
	if limite > 6 {
		limite = 6
	}
	for indice := 0; indice < limite; indice++ {
		tienePaciente := false
		tieneDosis := false
		for _, valor := range filas[indice] {
			encabezado := normalizarEncabezado(valor)
			if encabezado == "PACIENTE" {
				tienePaciente = true
			}
			if strings.HasPrefix(encabezado, "DOSIS") {
				tieneDosis = true
			}
		}
		if tienePaciente && tieneDosis {
			return indice
		}
	}
	return -1
}

func detectarTipo(filas [][]any) string {
	if buscarFilaEncabezado(filas) >= 0 {
		return "yodo"
	}
	for _, fila := range filas {
		if len(fila) > 0 && esTexto(fila[0]) && patronC73.MatchString(strings.TrimSpace(fila[0].(string))) {
			return "cancer"
		}
	}
	return "desconocida"
}

func mapearColumnas(filaEncabezado []any) map[string]int {
	encabezados := make([]string, len(filaEncabezado))
	for indice, valor := range filaEncabezado {
		encabezados[indice] = normalizarEncabezado(valor)
	}
	columnas := map[string]int{}
	for _, entrada := range encabezadosYodo {
		for indice, encabezado := range encabezados {
			coincide := false
			for _, nombre := range entrada.nombres {
				if encabezado == nombre || strings.HasPrefix(encabezado, nombre) {
					coincide = true
					break
				}
			}
			if coincide {
				columnas[entrada.campo] = indice
				break
			}
		}
	}
	return columnas
}

func interpretarSexo(valor any) any {
	texto := normalizarEncabezado(valor)
	if texto == "F" || strings.HasPrefix(texto, "FEM") {
		return "F"
	}
	if texto == "M" || strings.HasPrefix(texto, "MASC") {
		return "M"
	}
	return nil
}

func interpretarDosis(valor any) any {
	if numero, ok := valor.(float64); ok {
		if numero > 0 && numero <= 500 {
			return numero
		}
		return nil
	}
	if m := regexp.MustCompile(`\d+(?:[.,]\d+)?`).FindString(comoTexto(valor)); m != "" {
		return convertirNumeroValor(m)
	}
	return nil
}

func crearRegistroBase(referencia string, nombrePaciente string) modelo.Objeto {
	o := modelo.Nuevo()
	o.Set("origen", "Excel")
	o.Set("archivo", referencia)
	o.Set("referencia", referencia)
	o.Set("nombrePaciente", nombrePaciente)
	for _, campo := range []string{"textoMotivo", "textoNarrativa", "textoClinico", "textoEco", "textoBiopsia"} {
		o.Set(campo, "")
	}
	for _, campo := range []string{"textoPaaf", "textoCintigrama", "tirads", "ecoConNodulos", "ecoConAdenopatias", "bethesda", "estadificacion", "riesgo", "tsh", "t4l", "t3", "trab", "antiTiroglobulina", "antiTpo", "indiceEnArchivo"} {
		o.Set(campo, nil)
	}
	o.Set("tiroglobulina", modelo.Tiroglobulina{EnAscenso: false, Mencionada: false})
	o.Set("tieneDatosGinecologicos", false)
	o.Set("rechazoYodo", false)
	return o
}

func extraerFilaYodo(fila []any, columnas map[string]int, referencia string) modelo.Objeto {
	valorEn := func(campo string) any {
		if indice, ok := columnas[campo]; ok {
			return celda(fila, indice)
		}
		return nil
	}
	nombreCelda := textoCelda(valorEn("paciente"))
	nombre, ok := nombreCelda.(string)
	if !ok || nombre == "" || patronSoloDigitos.MatchString(nombre) {
		return modelo.Objeto{}
	}

	columnasConocidas := map[int]bool{}
	for campo, indice := range columnas {
		if campo == "diagnostico" || campo == "medico" || campo == "solicitudEnviada" || campo == "fechaSolicitud" {
			continue
		}
		columnasConocidas[indice] = true
	}
	textos := []string{}
	for indice, valor := range fila {
		if esTexto(valor) && !columnasConocidas[indice] {
			textos = append(textos, strings.TrimSpace(valor.(string)))
		}
	}

	medicoEnColumna := textoCelda(valorEn("medico"))
	var textoMedico any
	if s, ok := medicoEnColumna.(string); ok && patronMedicoExcel.MatchString(s) {
		textoMedico = s
	} else {
		textoMedico = medicoEnColumna
		for _, texto := range textos {
			if patronMedicoExcel.MatchString(texto) {
				textoMedico = texto
				break
			}
		}
	}

	diagnosticoEnColumna := textoCelda(valorEn("diagnostico"))
	var textoDiagnostico any
	if s, ok := diagnosticoEnColumna.(string); ok && !patronMedicoExcel.MatchString(s) && !esFecha(valorEn("diagnostico")) {
		textoDiagnostico = s
	} else {
		textoDiagnostico = nil
		for _, texto := range textos {
			if texto != nombre && patronDiagExcel.MatchString(texto) && !patronMedicoExcel.MatchString(texto) {
				textoDiagnostico = texto
				break
			}
		}
	}

	fecha, hay := obtenerFecha(valorEn("fechaSolicitud"))
	if !hay {
		fecha, hay = obtenerFecha(valorEn("solicitudEnviada"))
	}
	if !hay {
		for _, valor := range fila {
			if f, ok := obtenerFecha(valor); ok {
				fecha = f
				hay = true
				break
			}
		}
	}
	if !hay {
		fecha = triFecha{fechaIso: nil, anio: nil, mes: nil}
	}

	valorDosis := valorEn("dosis")
	rutEnColumna := textoCelda(valorEn("rut"))

	o := crearRegistroBase(referencia, capitalizarNombre(nombre))

	var rutTexto string
	if s, ok := rutEnColumna.(string); ok && patronTieneDigito.MatchString(s) {
		rutTexto = s
	} else {
		for _, valor := range fila {
			if esTexto(valor) && patronRutExcel.MatchString(valor.(string)) {
				rutTexto = valor.(string)
				break
			}
		}
	}
	o.Set("rut", normalizarRut(rutTexto))

	if edad, ok := valorEn("edad").(float64); ok && edad > 0 && edad < 120 {
		o.Set("edad", int(math.Round(edad)))
	} else {
		o.Set("edad", nil)
	}

	o.Set("fechaIso", fecha.fechaIso)
	o.Set("anio", fecha.anio)
	o.Set("mes", fecha.mes)
	o.Set("sexoRegistrado", interpretarSexo(valorEn("sexo")))
	if textoDiagnostico == nil {
		o.Set("textoClinico", "")
	} else {
		o.Set("textoClinico", textoDiagnostico)
	}
	o.Set("textoDiagnosticoExcel", textoDiagnostico)
	o.Set("textoMedicoTratante", textoMedico)
	o.Set("dosisMci", interpretarDosis(valorDosis))
	o.Set("esRastreo", patronRastreo.MatchString(comoTexto(valorDosis)))
	o.Set("textoAntecedentesFamiliares", textoCelda(valorEn("antecedentesFamiliares")))
	o.Set("textoMorbidos", textoCelda(valorEn("morbidos")))
	o.Set("textoCigarrillo", textoCelda(valorEn("cigarrillo")))
	o.Set("textoAlcohol", textoCelda(valorEn("alcohol")))
	o.Set("textoDrogas", textoCelda(valorEn("drogas")))
	return o
}

func extraerHojaYodo(nombreHoja string, filas [][]any, nombreArchivo string) []modelo.Objeto {
	indiceEncabezado := buscarFilaEncabezado(filas)
	columnas := mapearColumnas(filas[indiceEncabezado])
	registros := []modelo.Objeto{}
	for indice := indiceEncabezado + 1; indice < len(filas); indice++ {
		referencia := nombreArchivo + " › hoja " + nombreHoja + " › fila " + strconv.Itoa(indice+1)
		registro := extraerFilaYodo(filas[indice], columnas, referencia)
		if registro.D != nil {
			registros = append(registros, registro)
		}
	}
	return registros
}

func extraerHojaCancer(nombreHoja string, filas [][]any, nombreArchivo string) []modelo.Objeto {
	var anioHoja any
	if patronAnio4.MatchString(strings.TrimSpace(nombreHoja)) {
		n, _ := strconv.Atoi(strings.TrimSpace(nombreHoja))
		anioHoja = n
	}
	registros := []modelo.Objeto{}
	for indice, fila := range filas {
		if !(len(fila) > 0 && esTexto(fila[0]) && patronC73.MatchString(strings.TrimSpace(fila[0].(string)))) {
			continue
		}
		var rutTexto string
		if len(fila) > 5 && esTexto(fila[5]) && patronRutExcel.MatchString(fila[5].(string)) {
			rutTexto = fila[5].(string)
		} else {
			for _, valor := range fila {
				if esTexto(valor) && patronRutExcel.MatchString(valor.(string)) {
					rutTexto = valor.(string)
					break
				}
			}
		}
		fecha, hay := obtenerFecha(celda(fila, 13))
		if !hay {
			for _, valor := range fila {
				if f, ok := obtenerFecha(valor); ok {
					fecha = f
					hay = true
					break
				}
			}
		}
		if !hay {
			fecha = triFecha{fechaIso: nil, anio: anioHoja, mes: nil}
		}
		o := modelo.Nuevo()
		o.Set("origen", "Registro de cáncer")
		o.Set("referencia", nombreArchivo+" › hoja "+nombreHoja+" › fila "+strconv.Itoa(indice+1))
		o.Set("rut", normalizarRut(rutTexto))
		if esTexto(celda(fila, 6)) {
			o.Set("nombrePaciente", capitalizarNombre(fila[6].(string)))
		} else {
			o.Set("nombrePaciente", "")
		}
		o.Set("sexoRegistrado", interpretarSexo(celda(fila, 8)))
		if edad, ok := celda(fila, 12).(float64); ok {
			o.Set("edad", int(math.Round(edad)))
		} else {
			o.Set("edad", nil)
		}
		if esTexto(celda(fila, 11)) {
			o.Set("estadoVital", strings.TrimSpace(fila[11].(string)))
		} else {
			o.Set("estadoVital", nil)
		}
		o.Set("fechaIso", fecha.fechaIso)
		o.Set("mes", fecha.mes)
		if fecha.anio != nil {
			o.Set("anio", fecha.anio)
		} else {
			o.Set("anio", anioHoja)
		}
		registros = append(registros, o)
	}
	return registros
}

type ResultadoExcel struct {
	RegistrosYodo   []modelo.Objeto
	RegistrosCancer []modelo.Objeto
	HojasOmitidas   []modelo.DocumentoNoLeido
}

func filaTieneContenido(fila []any) bool {
	return len(fila) > 0
}

type HojaLeida struct {
	Nombre string
	Filas  [][]any
}

func ExtraerLibro(hojas []HojaLeida, nombreArchivo string) ResultadoExcel {
	resultado := ResultadoExcel{RegistrosYodo: []modelo.Objeto{}, RegistrosCancer: []modelo.Objeto{}, HojasOmitidas: []modelo.DocumentoNoLeido{}}
	for _, hoja := range hojas {
		tipo := detectarTipo(hoja.Filas)
		switch tipo {
		case "yodo":
			resultado.RegistrosYodo = append(resultado.RegistrosYodo, extraerHojaYodo(hoja.Nombre, hoja.Filas, nombreArchivo)...)
		case "cancer":
			resultado.RegistrosCancer = append(resultado.RegistrosCancer, extraerHojaCancer(hoja.Nombre, hoja.Filas, nombreArchivo)...)
		default:
			tieneContenido := false
			for _, fila := range hoja.Filas {
				if filaTieneContenido(fila) {
					tieneContenido = true
					break
				}
			}
			if tieneContenido {
				resultado.HojasOmitidas = append(resultado.HojasOmitidas, modelo.DocumentoNoLeido{
					Archivo: nombreArchivo + " › hoja " + hoja.Nombre,
					Motivo:  "No se reconoce como planilla de yodo 131 ni como registro de cáncer de tiroides (C73)",
				})
			}
		}
	}
	return resultado
}
