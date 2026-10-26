package fusion

import (
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"yodo131/nucleo/modelo"
	"yodo131/nucleo/texto"
)

const ventanaMismoTratamientoDias = 60.0
const milisegundosPorDia = 86400000.0

var camposComparables = []struct {
	campo string
	tipo  string
}{
	{"edad", "edad"}, {"dosisMci", "exacto"}, {"medicoTratante", "medico"},
	{"diagnostico", "diagnostico"}, {"histologia", "histologia"}, {"motivoCategoria", "exacto"},
	{"tirads", "exacto"}, {"bethesda", "exacto"}, {"riesgo", "exacto"},
	{"tsh", "medicion"}, {"t4l", "medicion"}, {"t3", "medicion"}, {"trab", "medicion"},
	{"antiTiroglobulina", "medicion"}, {"antiTpo", "medicion"},
	{"trabElevado", "exacto"}, {"tiroglobulinaElevada", "exacto"}, {"antiTiroglobulinaElevado", "exacto"}, {"antiTpoElevado", "exacto"},
	{"ecoConNodulos", "exacto"}, {"ecoConAdenopatias", "exacto"},
	{"tieneAntecedenteFamiliar", "exacto"}, {"esFumador", "exacto"},
}

var camposTexto = []string{"textoMotivo", "textoEco", "textoPaaf", "textoBiopsia", "estadificacion", "textoCintigrama", "textoAntecedentesFamiliares", "textoMorbidos", "textoCigarrillo", "textoAlcohol", "textoDrogas", "textoDiagnosticoExcel"}

var nombresCampos = map[string]string{
	"edad": "Edad", "dosisMci": "Dosis (mCi)", "medicoTratante": "Médico tratante", "diagnostico": "Diagnóstico", "histologia": "Tipo histológico",
	"motivoCategoria": "Motivo", "tirads": "TIRADS", "bethesda": "Bethesda", "riesgo": "Riesgo", "tsh": "TSH", "t4l": "T4L", "t3": "T3", "trab": "TRAb",
	"antiTiroglobulina": "Anti-tiroglobulina", "antiTpo": "Anti-TPO", "trabElevado": "TRAb elevado", "tiroglobulinaElevada": "Tiroglobulina elevada",
	"antiTiroglobulinaElevado": "Anti-tiroglobulina elevada", "antiTpoElevado": "Anti-TPO elevado", "ecoConNodulos": "Nódulos en ecografía",
	"ecoConAdenopatias": "Adenopatías en ecografía", "tieneAntecedenteFamiliar": "Antecedente familiar", "esFumador": "Tabaquismo", "sexo": "Sexo",
}

var (
	patronTituloFusion = regexp.MustCompile(`(?i)^dra?\.?\s*`)
	patronNoNombre     = regexp.MustCompile(`[^a-zñ\s]`)
	patronEspacios     = regexp.MustCompile(`\s+`)
)

func normalizarNombre(nombre string) string {
	base := texto.QuitarDiacriticos(nombre)
	base = strings.ToLower(base)
	base = patronNoNombre.ReplaceAllString(base, " ")
	return strings.TrimSpace(patronEspacios.ReplaceAllString(base, " "))
}

func textoDe(valor any) string {
	s, _ := valor.(string)
	return s
}

func obtenerClavePaciente(registro modelo.Objeto) string {
	if rut, ok := registro.Get("rut").(string); ok && rut != "" {
		return "rut:" + rut
	}
	nombre := normalizarNombre(registro.Texto("nombrePaciente"))
	if nombre != "" {
		return "nombre:" + nombre
	}
	return "sin-identificar:" + registro.Texto("referencia")
}

func obtenerPalabrasMedico(nombreMedico string) []string {
	base := strings.ToLower(texto.QuitarDiacriticos(patronTituloFusion.ReplaceAllString(nombreMedico, "")))
	palabras := []string{}
	for _, palabra := range patronEspacios.Split(base, -1) {
		if utf8.RuneCountInString(palabra) > 1 {
			palabras = append(palabras, palabra)
		}
	}
	return palabras
}

func contiene(lista []string, objetivo string) bool {
	for _, elemento := range lista {
		if elemento == objetivo {
			return true
		}
	}
	return false
}

func sonEquivalentes(tipo string, valorA, valorB any) bool {
	switch tipo {
	case "edad":
		return math.Abs(numeroDe(valorA)-numeroDe(valorB)) <= 1
	case "medicion":
		a := valorA.(modelo.Medicion)
		b := valorB.(modelo.Medicion)
		return a.Valor == b.Valor && a.Comparador == b.Comparador
	case "medico":
		palabrasA := obtenerPalabrasMedico(textoDe(valorA))
		palabrasB := obtenerPalabrasMedico(textoDe(valorB))
		comparteApellido := len(palabrasA) > 0 && len(palabrasB) > 0 && palabrasA[len(palabrasA)-1] == palabrasB[len(palabrasB)-1]
		if comparteApellido {
			return true
		}
		todasAenB := true
		for _, p := range palabrasA {
			if !contiene(palabrasB, p) {
				todasAenB = false
				break
			}
		}
		if todasAenB {
			return true
		}
		for _, p := range palabrasB {
			if !contiene(palabrasA, p) {
				return false
			}
		}
		return true
	}
	return valoresIguales(valorA, valorB)
}

func valoresIguales(a, b any) bool {
	switch av := a.(type) {
	case modelo.Medicion:
		bv, ok := b.(modelo.Medicion)
		return ok && av == bv
	default:
		return a == b
	}
}

func numeroDe(valor any) float64 {
	switch v := valor.(type) {
	case int:
		return float64(v)
	case float64:
		return v
	}
	return 0
}

func jsNumero(valor float64) string {
	return strconv.FormatFloat(valor, 'f', -1, 64)
}

func describirValor(valor any) string {
	switch v := valor.(type) {
	case bool:
		if v {
			return "Sí"
		}
		return "No"
	case modelo.Medicion:
		return v.Comparador + strings.Replace(jsNumero(v.Valor), ".", ",", 1)
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case float64:
		return jsNumero(v)
	}
	return ""
}

type fuentePresente struct {
	registro modelo.Objeto
	valor    any
}

func valorPreferido(tipo string, valores []any) any {
	switch tipo {
	case "medico":
		orden := append([]any{}, valores...)
		sort.SliceStable(orden, func(i, j int) bool {
			return len(obtenerPalabrasMedico(textoDe(orden[j]))) < len(obtenerPalabrasMedico(textoDe(orden[i])))
		})
		return orden[0]
	case "edad":
		minimo := numeroDe(valores[0])
		esEntero := true
		for _, valor := range valores {
			if _, ok := valor.(int); !ok {
				esEntero = false
			}
			if numeroDe(valor) < minimo {
				minimo = numeroDe(valor)
			}
		}
		if esEntero {
			return int(minimo)
		}
		return minimo
	}
	return valores[0]
}

func fusionarCampo(campo, tipo string, registros []modelo.Objeto) (any, *modelo.Conflicto) {
	presentes := []fuentePresente{}
	for _, registro := range registros {
		if valor := registro.Get(campo); valor != nil {
			presentes = append(presentes, fuentePresente{registro: registro, valor: valor})
		}
	}
	if tipo == "diagnostico" {
		hayOtro := false
		for _, p := range presentes {
			if p.valor != "Otro / no determinado" {
				hayOtro = true
			}
		}
		if hayOtro {
			presentes = filtrarValor(presentes, "Otro / no determinado")
		}
	}
	if tipo == "histologia" {
		hayEspecifico := false
		for _, p := range presentes {
			if p.valor != "No especificado" {
				hayEspecifico = true
			}
		}
		if hayEspecifico {
			presentes = filtrarValor(presentes, "No especificado")
		}
	}
	if len(presentes) == 0 {
		return nil, nil
	}
	distintos := []fuentePresente{}
	for _, presente := range presentes {
		nuevo := true
		for _, otro := range distintos {
			if sonEquivalentes(tipo, otro.valor, presente.valor) {
				nuevo = false
				break
			}
		}
		if nuevo {
			distintos = append(distintos, presente)
		}
	}
	if len(distintos) == 1 {
		valores := make([]any, len(presentes))
		for indice, p := range presentes {
			valores[indice] = p.valor
		}
		return valorPreferido(tipo, valores), nil
	}
	valores := make([]modelo.ValorConflicto, len(presentes))
	for indice, p := range presentes {
		valores[indice] = modelo.ValorConflicto{
			Origen:     p.registro.Texto("origen"),
			Referencia: p.registro.Texto("referencia"),
			Valor:      describirValor(p.valor),
		}
	}
	nombreCampo := nombresCampos[campo]
	if nombreCampo == "" {
		nombreCampo = campo
	}
	return nil, &modelo.Conflicto{Campo: campo, NombreCampo: nombreCampo, Valores: valores}
}

func filtrarValor(presentes []fuentePresente, excluir any) []fuentePresente {
	filtrados := []fuentePresente{}
	for _, p := range presentes {
		if p.valor != excluir {
			filtrados = append(filtrados, p)
		}
	}
	return filtrados
}

func unirTextos(registros []modelo.Objeto, campo string) any {
	vistos := map[string]bool{}
	textos := []string{}
	for _, registro := range registros {
		valor := registro.Get(campo)
		s, ok := valor.(string)
		if !ok || strings.TrimSpace(s) == "" {
			continue
		}
		if !vistos[s] {
			vistos[s] = true
			textos = append(textos, s)
		}
	}
	if len(textos) == 0 {
		return nil
	}
	return strings.Join(textos, " | ")
}

type diaAprox struct {
	dia      float64
	esExacta bool
	soloAnio bool
	ok       bool
}

func diasDesdeUTC(anio, mes, dia int) float64 {
	return float64(time.Date(anio, time.Month(mes), dia, 0, 0, 0, 0, time.UTC).UnixMilli()) / milisegundosPorDia
}

func obtenerDiaAproximado(registro modelo.Objeto) diaAprox {
	if iso, ok := registro.Get("fechaIso").(string); ok && iso != "" {
		t, err := time.Parse("2006-01-02", iso)
		if err == nil {
			return diaAprox{dia: float64(t.UnixMilli()) / milisegundosPorDia, esExacta: true, ok: true}
		}
	}
	anio, hayAnio := registro.Entero("anio")
	mes, hayMes := registro.Entero("mes")
	if hayAnio && hayMes {
		return diaAprox{dia: diasDesdeUTC(anio, mes-1+1, 15), esExacta: false, ok: true}
	}
	if hayAnio {
		return diaAprox{dia: diasDesdeUTC(anio, 7, 1), esExacta: false, soloAnio: true, ok: true}
	}
	return diaAprox{}
}

func sonFechasCompatibles(a, b modelo.Objeto) bool {
	fechaA := obtenerDiaAproximado(a)
	fechaB := obtenerDiaAproximado(b)
	if !fechaA.ok || !fechaB.ok {
		return false
	}
	if fechaA.esExacta && fechaB.esExacta {
		return math.Abs(fechaA.dia-fechaB.dia) <= ventanaMismoTratamientoDias
	}
	anioA, _ := a.Entero("anio")
	anioB, _ := b.Entero("anio")
	if anioA != anioB {
		return false
	}
	if fechaA.soloAnio || fechaB.soloAnio {
		return true
	}
	mesA, _ := a.Entero("mes")
	mesB, _ := b.Entero("mes")
	return math.Abs(float64(mesA-mesB)) <= 2
}

func distanciaEntre(a, b modelo.Objeto) float64 {
	fechaA := obtenerDiaAproximado(a)
	fechaB := obtenerDiaAproximado(b)
	if fechaA.ok && fechaB.ok {
		return math.Abs(fechaA.dia - fechaB.dia)
	}
	return math.Inf(1)
}

func agruparEnTratamientos(registros []modelo.Objeto) [][]modelo.Objeto {
	grupos := [][]modelo.Objeto{}
	sinFecha := []modelo.Objeto{}
	ordenados := append([]modelo.Objeto{}, registros...)
	claveOrden := func(registro modelo.Objeto) float64 {
		d := obtenerDiaAproximado(registro)
		if !d.ok {
			return math.Inf(1)
		}
		return d.dia
	}
	sort.SliceStable(ordenados, func(i, j int) bool {
		return claveOrden(ordenados[i]) < claveOrden(ordenados[j])
	})
	for _, registro := range ordenados {
		if !obtenerDiaAproximado(registro).ok {
			sinFecha = append(sinFecha, registro)
			continue
		}
		candidatos := []int{}
		for indiceGrupo, grupo := range grupos {
			for _, otro := range grupo {
				if sonFechasCompatibles(otro, registro) {
					candidatos = append(candidatos, indiceGrupo)
					break
				}
			}
		}
		if len(candidatos) == 0 {
			grupos = append(grupos, []modelo.Objeto{registro})
			continue
		}
		distanciaGrupo := func(indiceGrupo int) float64 {
			minimo := math.Inf(1)
			for _, otro := range grupos[indiceGrupo] {
				if d := distanciaEntre(otro, registro); d < minimo {
					minimo = d
				}
			}
			return minimo
		}
		sort.SliceStable(candidatos, func(i, j int) bool {
			return distanciaGrupo(candidatos[i]) < distanciaGrupo(candidatos[j])
		})
		masCercano := candidatos[0]
		grupos[masCercano] = append(grupos[masCercano], registro)
	}
	for _, registro := range sinFecha {
		if len(grupos) == 1 {
			grupos[0] = append(grupos[0], registro)
		} else {
			grupos = append(grupos, []modelo.Objeto{registro})
		}
	}
	return grupos
}

func construirTratamiento(grupo []modelo.Objeto, paciente *modelo.Paciente, indice int) modelo.Objeto {
	t := modelo.Nuevo()
	t.Set("idTratamiento", paciente.Clave+"#"+strconv.Itoa(indice+1))
	t.Set("clavePaciente", paciente.Clave)
	t.Set("nombrePaciente", paciente.NombrePaciente)
	t.Set("rut", paciente.Rut)
	fuentes := make([]modelo.Fuente, len(grupo))
	for i, registro := range grupo {
		fuentes[i] = modelo.Fuente{
			Origen:     registro.Texto("origen"),
			Referencia: registro.Texto("referencia"),
			FechaIso:   registro.Get("fechaIso"),
			Anio:       registro.Get("anio"),
		}
	}
	t.Set("fuentes", fuentes)
	conflictos := []modelo.Conflicto{}
	for _, cc := range camposComparables {
		valor, conflicto := fusionarCampo(cc.campo, cc.tipo, grupo)
		t.Set(cc.campo, valor)
		if conflicto != nil {
			conflictos = append(conflictos, *conflicto)
		}
	}
	for _, campo := range camposTexto {
		t.Set(campo, unirTextos(grupo, campo))
	}

	conAnio := []modelo.Objeto{}
	for _, registro := range grupo {
		if registro.Get("anio") != nil {
			conAnio = append(conAnio, registro)
		}
	}
	sort.SliceStable(conAnio, func(i, j int) bool {
		da := obtenerDiaAproximado(conAnio[i])
		db := obtenerDiaAproximado(conAnio[j])
		if da.esExacta == db.esExacta {
			return da.dia < db.dia
		}
		return da.esExacta
	})
	if len(conAnio) > 0 {
		t.Set("fechaIso", conAnio[0].Get("fechaIso"))
		t.Set("anio", conAnio[0].Get("anio"))
		t.Set("mes", conAnio[0].Get("mes"))
	} else {
		t.Set("fechaIso", nil)
		t.Set("anio", nil)
		t.Set("mes", nil)
	}

	var maximo any
	enAscenso := false
	for _, registro := range grupo {
		tg := registro.Tiroglobulina()
		if tg.EnAscenso {
			enAscenso = true
		}
		if tg.ValorMaximo != nil {
			valor := tg.ValorMaximo.(float64)
			if maximo == nil || valor > maximo.(float64) {
				maximo = valor
			}
		}
	}
	t.Set("tiroglobulina", modelo.Tiroglobulina{ValorMaximo: maximo, EnAscenso: enAscenso})

	rechazo := false
	rastreo := false
	origenes := []string{}
	vistosOrigen := map[string]bool{}
	for _, registro := range grupo {
		if registro.Bool("rechazoYodo") {
			rechazo = true
		}
		if registro.Bool("esRastreo") {
			rastreo = true
		}
		origen := registro.Texto("origen")
		if !vistosOrigen[origen] {
			vistosOrigen[origen] = true
			origenes = append(origenes, origen)
		}
	}
	t.Set("rechazoYodo", rechazo)
	t.Set("esRastreo", rastreo)
	t.Set("origenes", origenes)
	t.Set("cantidadFuentes", len(grupo))
	t.Set("conflictos", conflictos)
	return t
}

func fusionarSexoPaciente(registros []modelo.Objeto) (any, string, *modelo.Conflicto) {
	explicitos := []modelo.Objeto{}
	for _, registro := range registros {
		if registro.Get("sexoExplicito") != nil {
			explicitos = append(explicitos, registro)
		}
	}
	valoresExplicitos := []string{}
	vistoExpl := map[string]bool{}
	for _, registro := range explicitos {
		valor := registro.Texto("sexoExplicito")
		if !vistoExpl[valor] {
			vistoExpl[valor] = true
			valoresExplicitos = append(valoresExplicitos, valor)
		}
	}
	inferidos := []string{}
	vistoInf := map[string]bool{}
	for _, registro := range registros {
		if valor, ok := registro.Get("sexoInferido").(string); ok && valor != "" {
			if !vistoInf[valor] {
				vistoInf[valor] = true
				inferidos = append(inferidos, valor)
			}
		}
	}
	describir := func(registro modelo.Objeto, valor, origen string) modelo.ValorConflicto {
		return modelo.ValorConflicto{Origen: registro.Texto("origen") + " (" + origen + ")", Referencia: registro.Texto("referencia"), Valor: valor}
	}
	if len(valoresExplicitos) > 1 {
		valores := make([]modelo.ValorConflicto, len(explicitos))
		for indice, registro := range explicitos {
			valores[indice] = describir(registro, registro.Texto("sexoExplicito"), registro.Texto("origenSexoExplicito"))
		}
		return nil, "Por revisar", &modelo.Conflicto{Campo: "sexo", NombreCampo: "Sexo", Valores: valores}
	}
	if len(valoresExplicitos) == 1 {
		if len(inferidos) == 1 && inferidos[0] != valoresExplicitos[0] {
			var conInferencia modelo.Objeto
			for _, registro := range registros {
				if registro.Get("sexoInferido") != nil {
					conInferencia = registro
					break
				}
			}
			valores := []modelo.ValorConflicto{
				describir(explicitos[0], valoresExplicitos[0], explicitos[0].Texto("origenSexoExplicito")),
				describir(conInferencia, inferidos[0], "deducido del "+strings.ToLower(conInferencia.Texto("origenSexoInferido"))),
			}
			return nil, "Por revisar", &modelo.Conflicto{Campo: "sexo", NombreCampo: "Sexo", Valores: valores}
		}
		return valoresExplicitos[0], explicitos[0].Texto("origenSexoExplicito"), nil
	}
	if len(inferidos) == 1 {
		for _, registro := range registros {
			if registro.Get("sexoInferido") != nil {
				return inferidos[0], registro.Texto("origenSexoInferido"), nil
			}
		}
	}
	return nil, "Sin dato", nil
}

func detectarNombresDistintos(registros []modelo.Objeto) *modelo.Conflicto {
	orden := []string{}
	ultimoPorNombre := map[string]modelo.Objeto{}
	for _, registro := range registros {
		if registro.Texto("nombrePaciente") == "" {
			continue
		}
		clave := normalizarNombre(registro.Texto("nombrePaciente"))
		if _, existe := ultimoPorNombre[clave]; !existe {
			orden = append(orden, clave)
		}
		ultimoPorNombre[clave] = registro
	}
	if len(orden) < 2 {
		return nil
	}
	nombres := make([]modelo.Objeto, len(orden))
	for indice, clave := range orden {
		nombres[indice] = ultimoPorNombre[clave]
	}
	palabrasDe := func(registro modelo.Objeto) []string {
		palabras := []string{}
		for _, palabra := range strings.Split(normalizarNombre(registro.Texto("nombrePaciente")), " ") {
			if utf8.RuneCountInString(palabra) > 2 {
				palabras = append(palabras, palabra)
			}
		}
		return palabras
	}
	base := palabrasDe(nombres[0])
	incompatible := false
	for _, registro := range nombres[1:] {
		coincidencias := 0
		for _, palabra := range palabrasDe(registro) {
			if contiene(base, palabra) {
				coincidencias++
			}
		}
		if coincidencias < 2 {
			incompatible = true
			break
		}
	}
	if !incompatible {
		return nil
	}
	valores := make([]modelo.ValorConflicto, len(nombres))
	for indice, registro := range nombres {
		valores[indice] = modelo.ValorConflicto{Origen: registro.Texto("origen"), Referencia: registro.Texto("referencia"), Valor: registro.Texto("nombrePaciente")}
	}
	return &modelo.Conflicto{Campo: "nombre", NombreCampo: "Mismo RUT con nombres distintos", Valores: valores}
}

func asignarClavesPorNombre(registros []modelo.Objeto) []string {
	clavePorNombre := map[string]string{}
	ambiguos := map[string]bool{}
	for _, registro := range registros {
		if registro.Get("rut") == nil {
			continue
		}
		nombre := normalizarNombre(registro.Texto("nombrePaciente"))
		if nombre == "" {
			continue
		}
		clave := "rut:" + registro.Texto("rut")
		if anterior, existe := clavePorNombre[nombre]; existe && anterior != clave {
			ambiguos[nombre] = true
		}
		clavePorNombre[nombre] = clave
	}
	claves := make([]string, len(registros))
	for indice, registro := range registros {
		if registro.Get("rut") != nil {
			claves[indice] = obtenerClavePaciente(registro)
			continue
		}
		nombre := normalizarNombre(registro.Texto("nombrePaciente"))
		if !ambiguos[nombre] {
			if clave, existe := clavePorNombre[nombre]; existe && clave != "" {
				claves[indice] = clave
				continue
			}
		}
		claves[indice] = obtenerClavePaciente(registro)
	}
	return claves
}

type Resultado struct {
	Pacientes       []*modelo.Paciente
	Tratamientos    []modelo.Objeto
	PacientesCancer []modelo.PacienteCancer
	Conflictos      []modelo.Conflicto
	RegistrosUnidos int
}

func Fusionar(registrosTratamiento, registrosCancer []modelo.Objeto) Resultado {
	ordenClaves := []string{}
	registrosPorPaciente := map[string][]modelo.Objeto{}
	claves := asignarClavesPorNombre(registrosTratamiento)
	for indice, registro := range registrosTratamiento {
		clave := claves[indice]
		if _, existe := registrosPorPaciente[clave]; !existe {
			ordenClaves = append(ordenClaves, clave)
		}
		registrosPorPaciente[clave] = append(registrosPorPaciente[clave], registro)
	}

	ordenCancer := []string{}
	cancerPorPaciente := map[string][]modelo.Objeto{}
	for _, registro := range registrosCancer {
		clave := obtenerClavePaciente(registro)
		if _, existe := cancerPorPaciente[clave]; !existe {
			ordenCancer = append(ordenCancer, clave)
		}
		cancerPorPaciente[clave] = append(cancerPorPaciente[clave], registro)
	}

	pacientes := []*modelo.Paciente{}
	tratamientos := []modelo.Objeto{}
	for _, clave := range ordenClaves {
		registros := registrosPorPaciente[clave]
		registrosCancerPaciente := cancerPorPaciente[clave]
		var rut any
		for _, registro := range registros {
			if registro.Get("rut") != nil {
				rut = registro.Get("rut")
				break
			}
		}
		nombrePaciente := registros[0].Texto("nombrePaciente")
		for _, registro := range registros {
			if registro.Texto("origen") == "Word" {
				nombrePaciente = registro.Texto("nombrePaciente")
				break
			}
		}
		registrosSexo := append([]modelo.Objeto{}, registros...)
		for _, registro := range registrosCancerPaciente {
			copia := registro.Clonar()
			copia.Set("sexoExplicito", registro.Get("sexoRegistrado"))
			copia.Set("origenSexoExplicito", "Registro de cáncer")
			registrosSexo = append(registrosSexo, copia)
		}
		sexoValor, origenSexo, conflictoSexo := fusionarSexoPaciente(registrosSexo)
		paciente := &modelo.Paciente{Clave: clave, NombrePaciente: nombrePaciente, Rut: rut, Sexo: sexoValor, OrigenSexo: origenSexo, Tratamientos: []modelo.Objeto{}}
		var conflictoNombre *modelo.Conflicto
		if paciente.Rut != nil {
			conflictoNombre = detectarNombresDistintos(registros)
		}

		for indice, grupo := range agruparEnTratamientos(registros) {
			tratamiento := construirTratamiento(grupo, paciente, indice)
			tratamiento.Set("sexo", paciente.Sexo)
			tratamiento.Set("origenSexo", paciente.OrigenSexo)
			conflictos := tratamiento.Get("conflictos").([]modelo.Conflicto)
			if conflictoSexo != nil {
				conflictos = append([]modelo.Conflicto{*conflictoSexo}, conflictos...)
			}
			if conflictoNombre != nil {
				conflictos = append([]modelo.Conflicto{*conflictoNombre}, conflictos...)
			}
			tratamiento.Set("conflictos", conflictos)
			paciente.Tratamientos = append(paciente.Tratamientos, tratamiento)
			tratamientos = append(tratamientos, tratamiento)
		}
		pacientes = append(pacientes, paciente)
	}

	ordenPacientesCancer := []string{}
	pacientesCancer := map[string]*modelo.PacienteCancer{}
	registrarCancer := func(clave string, anio any, sexo any, enRegistro, enInformes bool) {
		actual, existe := pacientesCancer[clave]
		anioMenor := false
		if anio != nil {
			if !existe || actual.Anio == nil {
				anioMenor = true
			} else if numeroDe(anio) < numeroDe(actual.Anio) {
				anioMenor = true
			}
		}
		if !existe || anioMenor {
			nuevo := &modelo.PacienteCancer{Clave: clave, Anio: anio}
			if existe {
				nuevo.EnRegistro = actual.EnRegistro
				nuevo.EnInformes = actual.EnInformes
				if actual.Sexo != nil {
					nuevo.Sexo = actual.Sexo
				} else {
					nuevo.Sexo = sexo
				}
			} else {
				nuevo.Sexo = sexo
			}
			if enRegistro {
				nuevo.EnRegistro = true
			}
			if enInformes {
				nuevo.EnInformes = true
			}
			if !existe {
				ordenPacientesCancer = append(ordenPacientesCancer, clave)
			}
			pacientesCancer[clave] = nuevo
		}
	}
	for _, clave := range ordenCancer {
		registros := cancerPorPaciente[clave]
		anioMin := math.Inf(1)
		for _, registro := range registros {
			if anio, ok := registro.Entero("anio"); ok {
				if float64(anio) < anioMin {
					anioMin = float64(anio)
				}
			}
		}
		var anio any
		if !math.IsInf(anioMin, 1) {
			anio = int(anioMin)
		}
		var sexo any
		var pacienteYodo *modelo.Paciente
		for _, paciente := range pacientes {
			if paciente.Clave == clave {
				pacienteYodo = paciente
				break
			}
		}
		if pacienteYodo != nil {
			sexo = pacienteYodo.Sexo
		} else {
			sexosRegistro := []string{}
			visto := map[string]bool{}
			for _, registro := range registros {
				if s, ok := registro.Get("sexoRegistrado").(string); ok && s != "" && !visto[s] {
					visto[s] = true
					sexosRegistro = append(sexosRegistro, s)
				}
			}
			if len(sexosRegistro) == 1 {
				sexo = sexosRegistro[0]
			}
		}
		registrarCancer(clave, anio, sexo, true, false)
	}
	for _, tratamiento := range tratamientos {
		if tratamiento.Get("diagnostico") == "Cáncer de tiroides" {
			registrarCancer(tratamiento.Texto("clavePaciente"), tratamiento.Get("anio"), tratamiento.Get("sexo"), false, true)
		}
	}
	for _, paciente := range pacientes {
		_, tieneCancer := pacientesCancer[paciente.Clave]
		paciente.TieneCancerTiroides = tieneCancer
		_, enRegistroCancer := cancerPorPaciente[paciente.Clave]
		for indice, tratamiento := range paciente.Tratamientos {
			tratamiento.Set("pacienteConCancerRegistrado", enRegistroCancer)
			tratamiento.Set("numeroTratamientoPaciente", indice+1)
			tratamiento.Set("cantidadTratamientosPaciente", len(paciente.Tratamientos))
		}
	}

	conflictos := []modelo.Conflicto{}
	for _, tratamiento := range tratamientos {
		for _, conflicto := range tratamiento.Get("conflictos").([]modelo.Conflicto) {
			c := conflicto
			c.IdTratamiento = tratamiento.Texto("idTratamiento")
			c.NombrePaciente = tratamiento.Texto("nombrePaciente")
			c.Rut = tratamiento.Get("rut")
			c.FechaIso = tratamiento.Get("fechaIso")
			c.Anio = tratamiento.Get("anio")
			conflictos = append(conflictos, c)
		}
	}
	conflictosFiltrados := []modelo.Conflicto{}
	for indice, conflicto := range conflictos {
		if conflicto.Campo != "sexo" && conflicto.Campo != "nombre" {
			conflictosFiltrados = append(conflictosFiltrados, conflicto)
			continue
		}
		primero := -1
		for otro, candidato := range conflictos {
			if candidato.Campo == conflicto.Campo && igualRut(candidato.Rut, conflicto.Rut) && candidato.NombrePaciente == conflicto.NombrePaciente {
				primero = otro
				break
			}
		}
		if primero == indice {
			conflictosFiltrados = append(conflictosFiltrados, conflicto)
		}
	}

	registrosUnidos := 0
	for _, tratamiento := range tratamientos {
		cantidad, _ := tratamiento.Entero("cantidadFuentes")
		registrosUnidos += cantidad - 1
	}

	ordenados := make([]modelo.PacienteCancer, len(ordenPacientesCancer))
	for indice, clave := range ordenPacientesCancer {
		ordenados[indice] = *pacientesCancer[clave]
	}

	return Resultado{
		Pacientes:       pacientes,
		Tratamientos:    tratamientos,
		PacientesCancer: ordenados,
		Conflictos:      conflictosFiltrados,
		RegistrosUnidos: registrosUnidos,
	}
}

func igualRut(a, b any) bool {
	return a == b
}
