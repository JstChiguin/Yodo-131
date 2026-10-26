package exportacion

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"regexp"
	"strconv"
	"strings"

	"yodo131/nucleo/estadisticas"
	"yodo131/nucleo/modelo"
	"yodo131/nucleo/version"
)

type entradaZip struct {
	nombre string
	texto  string
}

func crearZipSinCompresion(entradas []entradaZip) []byte {
	var locales bytes.Buffer
	var directorio bytes.Buffer
	desplazamiento := 0
	for _, entrada := range entradas {
		nombre := []byte(entrada.nombre)
		datos := []byte(entrada.texto)
		crc := crc32.ChecksumIEEE(datos)
		cabecera := make([]byte, 30)
		binary.LittleEndian.PutUint32(cabecera[0:], 0x04034b50)
		binary.LittleEndian.PutUint16(cabecera[4:], 20)
		binary.LittleEndian.PutUint16(cabecera[6:], 0x0800)
		binary.LittleEndian.PutUint32(cabecera[14:], crc)
		binary.LittleEndian.PutUint32(cabecera[18:], uint32(len(datos)))
		binary.LittleEndian.PutUint32(cabecera[22:], uint32(len(datos)))
		binary.LittleEndian.PutUint16(cabecera[26:], uint16(len(nombre)))
		locales.Write(cabecera)
		locales.Write(nombre)
		locales.Write(datos)

		central := make([]byte, 46)
		binary.LittleEndian.PutUint32(central[0:], 0x02014b50)
		binary.LittleEndian.PutUint16(central[4:], 20)
		binary.LittleEndian.PutUint16(central[6:], 20)
		binary.LittleEndian.PutUint16(central[8:], 0x0800)
		binary.LittleEndian.PutUint32(central[16:], crc)
		binary.LittleEndian.PutUint32(central[20:], uint32(len(datos)))
		binary.LittleEndian.PutUint32(central[24:], uint32(len(datos)))
		binary.LittleEndian.PutUint16(central[28:], uint16(len(nombre)))
		binary.LittleEndian.PutUint32(central[42:], uint32(desplazamiento))
		directorio.Write(central)
		directorio.Write(nombre)
		desplazamiento += 30 + len(nombre) + len(datos)
	}
	fin := make([]byte, 22)
	binary.LittleEndian.PutUint32(fin[0:], 0x06054b50)
	binary.LittleEndian.PutUint16(fin[8:], uint16(len(entradas)))
	binary.LittleEndian.PutUint16(fin[10:], uint16(len(entradas)))
	binary.LittleEndian.PutUint32(fin[12:], uint32(directorio.Len()))
	binary.LittleEndian.PutUint32(fin[16:], uint32(desplazamiento))

	var salida bytes.Buffer
	salida.Write(locales.Bytes())
	salida.Write(directorio.Bytes())
	salida.Write(fin)
	return salida.Bytes()
}

var caracteresControl = regexp.MustCompile("[\u0000-\u0008\u000b\u000c\u000e-\u001f]")

func escaparXml(texto string) string {
	texto = caracteresControl.ReplaceAllString(texto, "")
	texto = strings.ReplaceAll(texto, "&", "&amp;")
	texto = strings.ReplaceAll(texto, "<", "&lt;")
	texto = strings.ReplaceAll(texto, ">", "&gt;")
	return strings.ReplaceAll(texto, "\"", "&quot;")
}

func letraColumna(indice int) string {
	letra := ""
	for numero := indice + 1; numero > 0; numero = (numero - 1) / 26 {
		letra = string(rune(65+((numero-1)%26))) + letra
	}
	return letra
}

func jsNumero(valor float64) string {
	return strconv.FormatFloat(valor, 'f', -1, 64)
}

func generarHojaXml(filas [][]any, anchos []int) string {
	var columnas strings.Builder
	for indice, ancho := range anchos {
		columnas.WriteString("<col min=\"" + strconv.Itoa(indice+1) + "\" max=\"" + strconv.Itoa(indice+1) + "\" width=\"" + strconv.Itoa(ancho) + "\" customWidth=\"1\"/>")
	}
	var filasXml strings.Builder
	for indiceFila, fila := range filas {
		filasXml.WriteString("<row r=\"" + strconv.Itoa(indiceFila+1) + "\">")
		for indiceColumna, valor := range fila {
			if valor == nil || valor == "" {
				continue
			}
			referencia := letraColumna(indiceColumna) + strconv.Itoa(indiceFila+1)
			estilo := ""
			if indiceFila == 0 {
				estilo = " s=\"1\""
			}
			switch n := valor.(type) {
			case int:
				filasXml.WriteString("<c r=\"" + referencia + "\"" + estilo + "><v>" + strconv.Itoa(n) + "</v></c>")
			case float64:
				filasXml.WriteString("<c r=\"" + referencia + "\"" + estilo + "><v>" + jsNumero(n) + "</v></c>")
			case string:
				filasXml.WriteString("<c r=\"" + referencia + "\" t=\"inlineStr\"" + estilo + "><is><t xml:space=\"preserve\">" + escaparXml(n) + "</t></is></c>")
			}
		}
		filasXml.WriteString("</row>")
	}
	return "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>" +
		"<worksheet xmlns=\"http://schemas.openxmlformats.org/spreadsheetml/2006/main\">" +
		"<sheetViews><sheetView workbookViewId=\"0\"><pane ySplit=\"1\" topLeftCell=\"A2\" activePane=\"bottomLeft\" state=\"frozen\"/></sheetView></sheetViews>" +
		"<cols>" + columnas.String() + "</cols><sheetData>" + filasXml.String() + "</sheetData></worksheet>"
}

type hojaExcel struct {
	nombre string
	filas  [][]any
	anchos []int
}

func crearLibroXlsx(hojas []hojaExcel) []byte {
	var tipos, relaciones, listaHojas strings.Builder
	for indice, hoja := range hojas {
		n := strconv.Itoa(indice + 1)
		tipos.WriteString("<Override PartName=\"/xl/worksheets/hoja" + n + ".xml\" ContentType=\"application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml\"/>")
		relaciones.WriteString("<Relationship Id=\"rId" + n + "\" Type=\"http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet\" Target=\"worksheets/hoja" + n + ".xml\"/>")
		listaHojas.WriteString("<sheet name=\"" + escaparXml(hoja.nombre) + "\" sheetId=\"" + n + "\" r:id=\"rId" + n + "\"/>")
	}
	entradas := []entradaZip{
		{"[Content_Types].xml", "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?><Types xmlns=\"http://schemas.openxmlformats.org/package/2006/content-types\"><Default Extension=\"rels\" ContentType=\"application/vnd.openxmlformats-package.relationships+xml\"/><Default Extension=\"xml\" ContentType=\"application/xml\"/><Override PartName=\"/xl/workbook.xml\" ContentType=\"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml\"/><Override PartName=\"/xl/styles.xml\" ContentType=\"application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml\"/>" + tipos.String() + "</Types>"},
		{"_rels/.rels", "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?><Relationships xmlns=\"http://schemas.openxmlformats.org/package/2006/relationships\"><Relationship Id=\"rId1\" Type=\"http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument\" Target=\"xl/workbook.xml\"/></Relationships>"},
		{"xl/workbook.xml", "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?><workbook xmlns=\"http://schemas.openxmlformats.org/spreadsheetml/2006/main\" xmlns:r=\"http://schemas.openxmlformats.org/officeDocument/2006/relationships\"><sheets>" + listaHojas.String() + "</sheets></workbook>"},
		{"xl/_rels/workbook.xml.rels", "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?><Relationships xmlns=\"http://schemas.openxmlformats.org/package/2006/relationships\">" + relaciones.String() + "<Relationship Id=\"rId" + strconv.Itoa(len(hojas)+1) + "\" Type=\"http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles\" Target=\"styles.xml\"/></Relationships>"},
		{"xl/styles.xml", "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?><styleSheet xmlns=\"http://schemas.openxmlformats.org/spreadsheetml/2006/main\"><fonts count=\"2\"><font><sz val=\"11\"/><name val=\"Calibri\"/></font><font><b/><sz val=\"11\"/><name val=\"Calibri\"/></font></fonts><fills count=\"3\"><fill><patternFill patternType=\"none\"/></fill><fill><patternFill patternType=\"gray125\"/></fill><fill><patternFill patternType=\"solid\"><fgColor rgb=\"FFE3ECEA\"/></patternFill></fill></fills><borders count=\"1\"><border/></borders><cellStyleXfs count=\"1\"><xf numFmtId=\"0\" fontId=\"0\" fillId=\"0\" borderId=\"0\"/></cellStyleXfs><cellXfs count=\"2\"><xf numFmtId=\"0\" fontId=\"0\" fillId=\"0\" borderId=\"0\" xfId=\"0\"/><xf numFmtId=\"0\" fontId=\"1\" fillId=\"2\" borderId=\"0\" xfId=\"0\" applyFont=\"1\" applyFill=\"1\"/></cellXfs><cellStyles count=\"1\"><cellStyle name=\"Normal\" xfId=\"0\" builtinId=\"0\"/></cellStyles></styleSheet>"},
	}
	for indice, hoja := range hojas {
		entradas = append(entradas, entradaZip{"xl/worksheets/hoja" + strconv.Itoa(indice+1) + ".xml", generarHojaXml(hoja.filas, hoja.anchos)})
	}
	return crearZipSinCompresion(entradas)
}

func textoSiNo(valor any) any {
	if v, ok := valor.(bool); ok {
		if v {
			return "Sí"
		}
		return "No"
	}
	return ""
}

func textoElevado(valor any) any {
	if v, ok := valor.(bool); ok {
		if v {
			return "Elevado"
		}
		return "Normal"
	}
	return ""
}

func textoMedicion(valor any) any {
	if m, ok := valor.(modelo.Medicion); ok {
		return m.Comparador + strings.Replace(jsNumero(m.Valor), ".", ",", 1)
	}
	return ""
}

func oDefecto(valor any) any {
	if valor == nil {
		return ""
	}
	return valor
}

type columna struct {
	titulo  string
	ancho   int
	obtener func(modelo.Objeto) any
}

var columnasTratamientos = []columna{
	{"Fuentes", 50, func(t modelo.Objeto) any {
		partes := []string{}
		for _, f := range t.Get("fuentes").([]modelo.Fuente) {
			partes = append(partes, f.Origen+": "+f.Referencia)
		}
		return strings.Join(partes, " | ")
	}},
	{"Origen", 12, func(t modelo.Objeto) any { return textoOrigen(t) }},
	{"Tratamiento N° del paciente", 12, func(t modelo.Objeto) any {
		n, _ := t.Entero("numeroTratamientoPaciente")
		c, _ := t.Entero("cantidadTratamientosPaciente")
		return strconv.Itoa(n) + " de " + strconv.Itoa(c)
	}},
	{"Datos que no coinciden", 34, func(t modelo.Objeto) any {
		visto := map[string]bool{}
		nombres := []string{}
		for _, c := range t.Get("conflictos").([]modelo.Conflicto) {
			if !visto[c.NombreCampo] {
				visto[c.NombreCampo] = true
				nombres = append(nombres, c.NombreCampo)
			}
		}
		return strings.Join(nombres, ", ")
	}},
	{"Paciente", 28, func(t modelo.Objeto) any { return t.Get("nombrePaciente") }},
	{"RUT", 13, func(t modelo.Objeto) any { return t.Get("rut") }},
	{"Edad", 6, func(t modelo.Objeto) any { return t.Get("edad") }},
	{"Sexo", 6, func(t modelo.Objeto) any { return t.Get("sexo") }},
	{"Origen del sexo", 20, func(t modelo.Objeto) any { return t.Get("origenSexo") }},
	{"Fecha", 11, func(t modelo.Objeto) any {
		if iso, ok := t.Get("fechaIso").(string); ok {
			return iso
		}
		if anio, ok := t.Entero("anio"); ok {
			return strconv.Itoa(anio)
		}
		return ""
	}},
	{"Año", 6, func(t modelo.Objeto) any { return t.Get("anio") }},
	{"Médico tratante", 24, func(t modelo.Objeto) any { return t.Get("medicoTratante") }},
	{"Dosis (mCi)", 10, func(t modelo.Objeto) any { return t.Get("dosisMci") }},
	{"Grupo dosis", 20, func(t modelo.Objeto) any { return t.Get("grupoDosis") }},
	{"Rechazó yodo", 12, func(t modelo.Objeto) any { return textoSiNo(t.Get("rechazoYodo")) }},
	{"Diagnóstico", 34, func(t modelo.Objeto) any { return t.Get("diagnostico") }},
	{"Tipo histológico", 24, func(t modelo.Objeto) any { return t.Get("histologia") }},
	{"Motivo (categoría)", 32, func(t modelo.Objeto) any { return t.Get("motivoCategoria") }},
	{"Motivo (texto)", 60, func(t modelo.Objeto) any { return t.Get("textoMotivo") }},
	{"Ecografía", 60, func(t modelo.Objeto) any { return t.Get("textoEco") }},
	{"TIRADS", 7, func(t modelo.Objeto) any { return t.Get("tirads") }},
	{"Eco con nódulos", 9, func(t modelo.Objeto) any { return textoSiNo(t.Get("ecoConNodulos")) }},
	{"Eco con adenopatías", 9, func(t modelo.Objeto) any { return textoSiNo(t.Get("ecoConAdenopatias")) }},
	{"PAAF", 40, func(t modelo.Objeto) any { return t.Get("textoPaaf") }},
	{"Bethesda", 9, func(t modelo.Objeto) any { return t.Get("bethesda") }},
	{"Biopsia", 60, func(t modelo.Objeto) any { return t.Get("textoBiopsia") }},
	{"Estadificación", 20, func(t modelo.Objeto) any { return t.Get("estadificacion") }},
	{"Riesgo", 11, func(t modelo.Objeto) any { return t.Get("riesgo") }},
	{"Cintigrama", 40, func(t modelo.Objeto) any { return t.Get("textoCintigrama") }},
	{"TSH", 8, func(t modelo.Objeto) any { return textoMedicion(t.Get("tsh")) }},
	{"T4L", 8, func(t modelo.Objeto) any { return textoMedicion(t.Get("t4l")) }},
	{"T3", 8, func(t modelo.Objeto) any { return textoMedicion(t.Get("t3")) }},
	{"TRAb", 8, func(t modelo.Objeto) any { return textoMedicion(t.Get("trab")) }},
	{"TRAb estado", 10, func(t modelo.Objeto) any { return textoElevado(t.Get("trabElevado")) }},
	{"Tiroglobulina (máx.)", 10, func(t modelo.Objeto) any { return t.Tiroglobulina().ValorMaximo }},
	{"Tiroglobulina estado", 11, func(t modelo.Objeto) any { return textoElevado(t.Get("tiroglobulinaElevada")) }},
	{"Anti-tiroglobulina", 10, func(t modelo.Objeto) any { return textoMedicion(t.Get("antiTiroglobulina")) }},
	{"Anti-Tg estado", 10, func(t modelo.Objeto) any { return textoElevado(t.Get("antiTiroglobulinaElevado")) }},
	{"Anti-TPO", 9, func(t modelo.Objeto) any { return textoMedicion(t.Get("antiTpo")) }},
	{"Anti-TPO estado", 10, func(t modelo.Objeto) any { return textoElevado(t.Get("antiTpoElevado")) }},
	{"Antecedentes familiares", 28, func(t modelo.Objeto) any { return t.Get("textoAntecedentesFamiliares") }},
	{"Antecedente familiar", 10, func(t modelo.Objeto) any { return textoSiNo(t.Get("tieneAntecedenteFamiliar")) }},
	{"Mórbidos", 40, func(t modelo.Objeto) any { return t.Get("textoMorbidos") }},
	{"Cigarrillo", 24, func(t modelo.Objeto) any { return t.Get("textoCigarrillo") }},
	{"Fumador", 9, func(t modelo.Objeto) any { return textoSiNo(t.Get("esFumador")) }},
	{"OH", 20, func(t modelo.Objeto) any { return t.Get("textoAlcohol") }},
	{"Drogas", 18, func(t modelo.Objeto) any { return t.Get("textoDrogas") }},
	{"Diagnóstico en Excel", 30, func(t modelo.Objeto) any { return t.Get("textoDiagnosticoExcel") }},
	{"En registro de cáncer (C73)", 12, func(t modelo.Objeto) any { return textoSiNo(t.Get("pacienteConCancerRegistrado")) }},
	{"Datos que faltan", 24, func(t modelo.Objeto) any {
		campos, _ := t.Get("camposFaltantes").([]string)
		return strings.Join(campos, ", ")
	}},
}

func textoOrigen(t modelo.Objeto) any {
	origenes, _ := t.Get("origenes").([]string)
	if len(origenes) > 1 {
		return "Word y Excel"
	}
	if len(origenes) == 1 {
		return origenes[0]
	}
	return nil
}

func porcentaje(parte, total int) any {
	if total == 0 {
		return ""
	}
	valor := float64(parte) / float64(total) * 1000
	redondeado := int(valor + 0.5)
	resultado := float64(redondeado) / 10
	if resultado == float64(int(resultado)) {
		return int(resultado)
	}
	return resultado
}

func crearHojaResumen(est estadisticas.Estadisticas, res modelo.Resultado, descripcionPeriodo string) hojaExcel {
	conteo := res.Conteo
	filas := [][]any{{"Indicador", "Categoría", "Cantidad", "Con dato", "Total", "Cobertura"}}
	filas = append(filas,
		[]any{"Período", descripcionPeriodo},
		[]any{"Versión del lector", "v" + version.Aplicacion},
		[]any{"Informes Word leídos", "", conteo.InformesWord},
		[]any{"Filas de Excel de yodo leídas", "", conteo.FilasExcelYodo},
		[]any{"Filas del registro de cáncer leídas", "", conteo.FilasRegistroCancer},
		[]any{"Tratamientos", "", est.CantidadTratamientos},
		[]any{"Registros repetidos unidos", "", est.RegistrosUnidos},
		[]any{"Pacientes únicos", "", est.CantidadPacientes},
		[]any{"Pacientes con más de un tratamiento", "", est.PacientesConVariosTratamientos},
		[]any{"Pacientes", "Mujeres", est.PacientesMujeres},
		[]any{"Pacientes", "Hombres", est.PacientesHombres},
		[]any{"Pacientes", "Sexo sin dato o por revisar", est.PacientesSinSexo},
		[]any{"Pacientes con cáncer de tiroides", "", est.CantidadPacientesCancer},
		[]any{"Datos que no coinciden entre fuentes", "", est.CantidadConflictos},
	)
	if est.EdadPromedio != nil {
		promedio := est.EdadPromedio.(float64)
		redondeado := float64(int(promedio*10+0.5)) / 10
		var valorPromedio any = redondeado
		if redondeado == float64(int(redondeado)) {
			valorPromedio = int(redondeado)
		}
		filas = append(filas, []any{"Edad promedio", "", valorPromedio, est.CoberturaEdad.ConDato, est.CoberturaEdad.Total, porcentaje(est.CoberturaEdad.ConDato, est.CoberturaEdad.Total)})
	}
	agregar := func(indicador string, d estadisticas.Distribucion) {
		for _, fila := range d.Filas {
			filas = append(filas, []any{indicador, fila.Categoria, fila.Cantidad, d.ConDato, d.Total, porcentaje(d.ConDato, d.Total)})
		}
	}
	agregar("Diagnóstico", est.PorDiagnostico)
	agregar("Motivo de consulta", est.PorMotivo)
	agregar("Grupo de edad", est.PorEdad)
	agregar("Dosis", est.PorDosis)
	agregar("Tipo histológico (cáncer)", est.PorHistologia)
	agregar("Bethesda", est.PorBethesda)
	agregar("TIRADS", est.PorTirads)
	agregar("Origen de los datos", est.PorOrigen)

	grupos := []struct {
		conteo map[string]map[string]int
		orden  []string
		claves []string
	}{
		{est.Examenes, est.OrdenExamenes, []string{"Elevado", "Normal"}},
		{est.Ecografia, est.OrdenEcografia, []string{"Sí", "No"}},
		{est.Antecedentes, est.OrdenAntecedentes, []string{"Sí", "No"}},
	}
	for _, grupo := range grupos {
		for _, nombre := range grupo.orden {
			conteoGrupo := grupo.conteo[nombre]
			total := 0
			for _, cantidad := range conteoGrupo {
				total += cantidad
			}
			conDato := total - conteoGrupo["Sin dato"]
			for _, categoria := range grupo.claves {
				filas = append(filas, []any{nombre, categoria, conteoGrupo[categoria], conDato, total, porcentaje(conDato, total)})
			}
		}
	}
	for _, fila := range est.CancerPorAnio {
		filas = append(filas, []any{"Pacientes con cáncer de tiroides por año", fila.Categoria, fila.Mujeres + fila.Hombres + fila.SinDato})
	}
	return hojaExcel{nombre: "Resumen", filas: filas, anchos: []int{34, 36, 10, 10, 10, 11}}
}

func ExportarLibro(est estadisticas.Estadisticas, res modelo.Resultado, descripcionPeriodo string) []byte {
	titulos := make([]any, len(columnasTratamientos))
	anchos := make([]int, len(columnasTratamientos))
	for indice, col := range columnasTratamientos {
		titulos[indice] = col.titulo
		anchos[indice] = col.ancho
	}
	filasTratamientos := [][]any{titulos}
	for _, tratamiento := range est.Tratamientos {
		fila := make([]any, len(columnasTratamientos))
		for indice, col := range columnasTratamientos {
			fila[indice] = oDefecto(col.obtener(tratamiento))
		}
		filasTratamientos = append(filasTratamientos, fila)
	}
	hojaTratamientos := hojaExcel{nombre: "Tratamientos", filas: filasTratamientos, anchos: anchos}

	idsEnPeriodo := map[string]bool{}
	for _, tratamiento := range est.Tratamientos {
		idsEnPeriodo[tratamiento.Texto("idTratamiento")] = true
	}
	filasConflictos := [][]any{{"Paciente", "RUT", "Fecha del tratamiento", "Dato", "Fuente", "Valor", "Ubicación"}}
	for _, conflicto := range res.Conflictos {
		if !idsEnPeriodo[conflicto.IdTratamiento] {
			continue
		}
		fecha := ""
		if iso, ok := conflicto.FechaIso.(string); ok {
			fecha = iso
		} else if anio, ok := conflicto.Anio.(int); ok {
			fecha = strconv.Itoa(anio)
		}
		var rut any = conflicto.Rut
		for _, valor := range conflicto.Valores {
			filasConflictos = append(filasConflictos, []any{conflicto.NombrePaciente, oDefecto(rut), fecha, conflicto.NombreCampo, valor.Origen, valor.Valor, valor.Referencia})
		}
	}
	hojaConflictos := hojaExcel{nombre: "Por revisar", filas: filasConflictos, anchos: []int{28, 13, 12, 24, 26, 30, 50}}

	filasNoLeidos := [][]any{{"Archivo", "Motivo"}}
	for _, documento := range res.DocumentosNoLeidos {
		filasNoLeidos = append(filasNoLeidos, []any{documento.Archivo, documento.Motivo})
	}
	hojaNoLeidos := hojaExcel{nombre: "No leídos", filas: filasNoLeidos, anchos: []int{40, 70}}

	return crearLibroXlsx([]hojaExcel{hojaTratamientos, crearHojaResumen(est, res, descripcionPeriodo), hojaConflictos, hojaNoLeidos})
}

