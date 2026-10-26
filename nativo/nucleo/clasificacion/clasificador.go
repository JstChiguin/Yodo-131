package clasificacion

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"yodo131/nucleo/extraccion"
	"yodo131/nucleo/modelo"
	"yodo131/nucleo/texto"
)

var UmbralesPorDefecto = map[string]float64{"trab": 1.75, "tiroglobulina": 1.0, "antiTiroglobulina": 20, "antiTpo": 35}

var categoriasMotivo = []struct {
	nombre string
	patron *regexp.Regexp
}{
	{"Seguimiento de cáncer ya tratado", regexp.MustCompile(`(?i)antecedentes? de (c[aá]ncer|tiroidectom)|tiroidectom[ií]a|recidiva|met[aá]stasis|c[aá]ncer papilar|carcinoma`)},
	{"Aumento de volumen cervical o nódulo", regexp.MustCompile(`(?i)aumento de volumen|se palpa|autopalpa|n[oó]dulo|bocio|masa cervical`)},
	{"Síntomas de hipertiroidismo", regexp.MustCompile(`(?i)palpitaci|temblor|baja de peso|nerviosismo|taquicard|diaforesis|sudoraci|insomnio|fatigabilidad`)},
	{"Oftalmopatía o exoftalmo", regexp.MustCompile(`(?i)oftalmopat|exoftalm`)},
	{"Control endocrinológico sin detalle", regexp.MustCompile(`(?i)control`)},
}

var (
	patronCancer          = regexp.MustCompile(`(?i)carcinoma|c[aá]ncer\s+(papilar|folicular|medular|de(l)?\s+tiroides|tiroideo)|comit[eé]\s+onc|^\s*ca\.?\s`)
	patronBocio           = regexp.MustCompile(`(?i)multinodular|n[oó]dulo\s+(hiperfuncionante|caliente|t[oó]xico)`)
	patronHiper           = regexp.MustCompile(`(?i)graves|bocio difuso|hipertiroidismo|\btrab\b|thyrozol|tiamazol|tsh frenada`)
	patronPapilarFolic    = regexp.MustCompile(`(?i)papilar[^.]{0,40}varie?(dad|nte)\s+folicular`)
	patronPapilar         = regexp.MustCompile(`(?i)(carcinoma|c[aá]ncer)\s+papilar`)
	patronFolicular       = regexp.MustCompile(`(?i)(carcinoma|c[aá]ncer)\s+folicular`)
	patronMedular         = regexp.MustCompile(`(?i)medular`)
	patronAnaplasico      = regexp.MustCompile(`(?i)anapl[aá]sico`)
	patronSexoFemTexto    = regexp.MustCompile(`(?i)\b(derivada|la paciente|paciente femenina|operada)\b`)
	patronSexoMascTexto   = regexp.MustCompile(`(?i)\b(derivado|operado)\b|(^|\s)él(\s|[,.])`)
	patronRellenoMedico   = regexp.MustCompile(`[-.…_]{3,}`)
	patronEspaciosMedico  = regexp.MustCompile(`\s+`)
	patronDraInicio       = regexp.MustCompile(`(?i)^dra\b`)
	patronDrInicio        = regexp.MustCompile(`(?i)^dr\b`)
	patronTituloMedico    = regexp.MustCompile(`(?i)^dra?\.?\s*`)
	patronTituloConPunto  = regexp.MustCompile(`(?i)^dra?\.\s*`)
	patronTituloCaptura   = regexp.MustCompile(`(?i)^dra?\.`)
	patronInicialPalabra  = regexp.MustCompile(`(^|\s)(\p{L})`)
	patronTildes          = regexp.MustCompile(`(?i)[áéíóúñü]`)
)

func textoDe(valor any) (string, bool) {
	s, ok := valor.(string)
	return s, ok
}

func determinarSexoExplicito(informe modelo.Objeto) (any, any) {
	if sexo := informe.Get("sexoRegistrado"); sexo != nil {
		return sexo, "Registrado en Excel"
	}
	if informe.Bool("tieneDatosGinecologicos") {
		return "F", "Datos gineco-obstétricos"
	}
	return nil, nil
}

func determinarSexoInferido(informe modelo.Objeto) (any, any) {
	nombre := informe.Texto("nombrePaciente")
	primerNombre := strings.ToLower(strings.SplitN(nombre, " ", 2)[0])
	if nombresFemeninos[primerNombre] {
		return "F", "Nombre"
	}
	if nombresMasculinos[primerNombre] {
		return "M", "Nombre"
	}
	narrativa := informe.Texto("textoNarrativa")
	if patronSexoFemTexto.MatchString(narrativa) {
		return "F", "Texto del informe"
	}
	if patronSexoMascTexto.MatchString(narrativa) {
		return "M", "Texto del informe"
	}
	return nil, nil
}

func determinarSexo(informe modelo.Objeto) map[string]any {
	sexoExplicito, origenExplicito := determinarSexoExplicito(informe)
	sexoInferido, origenInferido := determinarSexoInferido(informe)
	sexo := sexoExplicito
	if sexo == nil {
		sexo = sexoInferido
	}
	origen := origenExplicito
	if origen == nil {
		origen = origenInferido
	}
	if origen == nil {
		origen = "Sin dato"
	}
	return map[string]any{
		"sexoExplicito": sexoExplicito, "origenSexoExplicito": origenExplicito,
		"sexoInferido": sexoInferido, "origenSexoInferido": origenInferido,
		"sexo": sexo, "origenSexo": origen,
	}
}

func clasificarDiagnostico(valor any) any {
	textoClinico, ok := textoDe(valor)
	if !ok || strings.TrimSpace(textoClinico) == "" {
		return nil
	}
	if patronCancer.MatchString(textoClinico) {
		return "Cáncer de tiroides"
	}
	if patronBocio.MatchString(textoClinico) {
		return "Bocio multinodular o nódulo hiperfuncionante"
	}
	if patronHiper.MatchString(textoClinico) {
		return "Hipertiroidismo (Graves / bocio difuso)"
	}
	return "Otro / no determinado"
}

func clasificarHistologia(informe modelo.Objeto) any {
	fuentes := []any{informe.Get("textoBiopsia"), informe.Get("textoClinico")}
	for _, fuente := range fuentes {
		t, ok := textoDe(fuente)
		if !ok || t == "" {
			continue
		}
		if patronPapilarFolic.MatchString(t) {
			return "Papilar, variante folicular"
		}
		if patronPapilar.MatchString(t) {
			return "Papilar"
		}
		if patronFolicular.MatchString(t) {
			return "Folicular"
		}
		if patronMedular.MatchString(t) {
			return "Medular"
		}
		if patronAnaplasico.MatchString(t) {
			return "Anaplásico"
		}
	}
	return "No especificado"
}

func clasificarMotivo(valor any) any {
	narrativa, ok := textoDe(valor)
	if !ok || strings.TrimSpace(narrativa) == "" {
		return nil
	}
	for _, categoria := range categoriasMotivo {
		if categoria.patron.MatchString(narrativa) {
			return categoria.nombre
		}
	}
	return "Otro"
}

func capitalizarPalabras(base string) string {
	return patronInicialPalabra.ReplaceAllStringFunc(base, func(coincidencia string) string {
		r := []rune(coincidencia)
		r[len(r)-1] = []rune(strings.ToUpper(string(r[len(r)-1])))[0]
		return string(r)
	})
}

func limpiarNombreMedico(valor any) any {
	textoMedico, ok := textoDe(valor)
	if !ok || textoMedico == "" {
		return nil
	}
	sinRelleno := strings.TrimSpace(patronEspaciosMedico.ReplaceAllString(patronRellenoMedico.ReplaceAllString(textoMedico, " "), " "))
	if sinRelleno == "" {
		return nil
	}
	titulo := ""
	if patronDraInicio.MatchString(sinRelleno) {
		titulo = "Dra."
	} else if patronDrInicio.MatchString(sinRelleno) {
		titulo = "Dr."
	}
	nombre := capitalizarPalabras(strings.ToLower(patronTituloMedico.ReplaceAllString(sinRelleno, "")))
	if nombre == "" {
		return nil
	}
	return strings.TrimSpace(titulo + " " + nombre)
}

func obtenerPalabrasNombre(nombreMedico string) []string {
	base := texto.QuitarDiacriticos(strings.ToLower(patronTituloMedico.ReplaceAllString(nombreMedico, "")))
	palabras := []string{}
	for _, palabra := range strings.Split(base, " ") {
		if len([]rune(palabra)) > 1 {
			palabras = append(palabras, palabra)
		}
	}
	return palabras
}

func contarTildes(s string) int {
	return len(patronTildes.FindAllString(s, -1))
}

func UnificarMedicos(tratamientos []modelo.Objeto) []modelo.Objeto {
	variantes := []string{}
	vistos := map[string]bool{}
	for _, t := range tratamientos {
		if medico, ok := textoDe(t.Get("medicoTratante")); ok && medico != "" {
			if !vistos[medico] {
				vistos[medico] = true
				variantes = append(variantes, medico)
			}
		}
	}
	palabrasPorVariante := map[string][]string{}
	for _, variante := range variantes {
		palabrasPorVariante[variante] = obtenerPalabrasNombre(variante)
	}
	veces := map[string]int{}
	for _, t := range tratamientos {
		if medico, ok := textoDe(t.Get("medicoTratante")); ok && medico != "" {
			veces[medico]++
		}
	}
	contiene := func(lista []string, objetivo string) bool {
		for _, elemento := range lista {
			if elemento == objetivo {
				return true
			}
		}
		return false
	}
	formaCompleta := map[string]string{}
	for _, variante := range variantes {
		palabras := palabrasPorVariante[variante]
		candidatas := []string{}
		for _, otra := range variantes {
			palabrasOtra := palabrasPorVariante[otra]
			todas := len(palabras) > 0
			for _, palabra := range palabras {
				if !contiene(palabrasOtra, palabra) {
					todas = false
					break
				}
			}
			if todas {
				candidatas = append(candidatas, otra)
			}
		}
		largoMaximo := 0
		for _, candidata := range candidatas {
			if l := len(palabrasPorVariante[candidata]); l > largoMaximo {
				largoMaximo = l
			}
		}
		masCompletas := []string{}
		for _, candidata := range candidatas {
			if len(palabrasPorVariante[candidata]) == largoMaximo {
				masCompletas = append(masCompletas, candidata)
			}
		}
		personas := map[string]bool{}
		for _, candidata := range masCompletas {
			personas[strings.Join(palabrasPorVariante[candidata], " ")] = true
		}
		if len(personas) > 1 {
			formaCompleta[variante] = variante
			continue
		}
		masCompletasOrden := append([]string{}, masCompletas...)
		sort.SliceStable(masCompletasOrden, func(i, j int) bool {
			a, b := masCompletasOrden[i], masCompletasOrden[j]
			if dt := contarTildes(b) - contarTildes(a); dt != 0 {
				return dt < 0
			}
			if dv := veces[b] - veces[a]; dv != 0 {
				return dv < 0
			}
			return strings.Compare(a, b) < 0
		})
		formaCompleta[variante] = masCompletasOrden[0]
	}

	tituloDe := func(variante string) string {
		return patronTituloCaptura.FindString(variante)
	}
	ordenFormas := []string{}
	vistasFormas := map[string]bool{}
	votos := map[string][]struct {
		titulo string
		cuenta int
	}{}
	for _, variante := range variantes {
		forma := formaCompleta[variante]
		titulo := tituloDe(variante)
		if titulo == "" {
			continue
		}
		if !vistasFormas[forma] {
			vistasFormas[forma] = true
			ordenFormas = append(ordenFormas, forma)
		}
		lista := votos[forma]
		encontrado := false
		for indice := range lista {
			if lista[indice].titulo == titulo {
				lista[indice].cuenta += veces[variante]
				encontrado = true
				break
			}
		}
		if !encontrado {
			lista = append(lista, struct {
				titulo string
				cuenta int
			}{titulo, veces[variante]})
		}
		votos[forma] = lista
	}

	conTituloMayoritario := func(forma string) string {
		lista, existe := votos[forma]
		if !existe || len(lista) == 0 {
			return forma
		}
		orden := append([]struct {
			titulo string
			cuenta int
		}{}, lista...)
		sort.SliceStable(orden, func(i, j int) bool { return orden[i].cuenta > orden[j].cuenta })
		return orden[0].titulo + " " + patronTituloConPunto.ReplaceAllString(forma, "")
	}

	resultado := make([]modelo.Objeto, len(tratamientos))
	for indice, t := range tratamientos {
		if medico, ok := textoDe(t.Get("medicoTratante")); ok && medico != "" {
			copia := t.Clonar()
			copia.Set("medicoTratante", conTituloMayoritario(formaCompleta[medico]))
			resultado[indice] = copia
		} else {
			resultado[indice] = t
		}
	}
	return resultado
}

func evaluarElevado(valor any, umbral float64) any {
	medicion, ok := valor.(modelo.Medicion)
	if !ok {
		return nil
	}
	if medicion.Comparador == "<" {
		if medicion.Valor > umbral {
			return nil
		}
		return false
	}
	return medicion.Valor > umbral
}

func evaluarTiroglobulina(tg modelo.Tiroglobulina, umbral float64) any {
	if tg.EnAscenso {
		return true
	}
	if tg.ValorMaximo == nil {
		return nil
	}
	return tg.ValorMaximo.(float64) > umbral
}

func AgruparEdad(valor any) any {
	edad, ok := comoNumero(valor)
	if !ok {
		return nil
	}
	if edad < 20 {
		return "<20"
	}
	if edad >= 70 {
		return "70+"
	}
	decada := int(math.Floor(edad/10)) * 10
	return fmt.Sprintf("%d–%d", decada, decada+9)
}

func AgruparDosis(valor any) any {
	dosis, ok := comoNumero(valor)
	if !ok {
		return nil
	}
	if dosis <= 30 {
		return "Baja (≤30 mCi)"
	}
	if dosis < 100 {
		return "Intermedia (31–99 mCi)"
	}
	return "Alta (≥100 mCi)"
}

func comoNumero(valor any) (float64, bool) {
	switch v := valor.(type) {
	case int:
		return float64(v), true
	case float64:
		return v, true
	}
	return 0, false
}

func ListarCamposFaltantes(registro modelo.Objeto) []string {
	faltantes := []string{}
	if registro.Get("rut") == nil {
		faltantes = append(faltantes, "RUT")
	}
	if registro.Get("edad") == nil {
		faltantes = append(faltantes, "edad")
	}
	if registro.Get("anio") == nil {
		faltantes = append(faltantes, "fecha")
	}
	if registro.Get("medicoTratante") == nil {
		faltantes = append(faltantes, "médico")
	}
	if registro.Get("dosisMci") == nil && !registro.Bool("rechazoYodo") {
		faltantes = append(faltantes, "dosis")
	}
	if registro.Get("sexo") == nil {
		faltantes = append(faltantes, "sexo")
	}
	return faltantes
}

func ClasificarInforme(informe modelo.Objeto, umbrales map[string]float64) modelo.Objeto {
	registro := informe.Clonar()
	for clave, valor := range determinarSexo(informe) {
		registro.Set(clave, valor)
	}
	diagnostico := clasificarDiagnostico(informe.Get("textoClinico"))
	registro.Set("diagnostico", diagnostico)
	if diagnostico == "Cáncer de tiroides" {
		registro.Set("histologia", clasificarHistologia(informe))
	} else {
		registro.Set("histologia", nil)
	}
	registro.Set("motivoCategoria", clasificarMotivo(informe.Get("textoNarrativa")))
	registro.Set("medicoTratante", limpiarNombreMedico(informe.Get("textoMedicoTratante")))
	registro.Set("grupoEdad", AgruparEdad(informe.Get("edad")))
	if informe.Bool("rechazoYodo") && informe.Get("dosisMci") == nil {
		registro.Set("grupoDosis", "Rechazó / no recibió")
	} else {
		registro.Set("grupoDosis", AgruparDosis(informe.Get("dosisMci")))
	}
	registro.Set("trabElevado", evaluarElevado(informe.Get("trab"), umbrales["trab"]))
	registro.Set("tiroglobulinaElevada", evaluarTiroglobulina(informe.Tiroglobulina(), umbrales["tiroglobulina"]))
	registro.Set("antiTiroglobulinaElevado", evaluarElevado(informe.Get("antiTiroglobulina"), umbrales["antiTiroglobulina"]))
	registro.Set("antiTpoElevado", evaluarElevado(informe.Get("antiTpo"), umbrales["antiTpo"]))
	registro.Set("tieneAntecedenteFamiliar", extraccion.EvaluarPresencia(informe.Get("textoAntecedentesFamiliares")))
	registro.Set("esFumador", extraccion.EvaluarPresencia(informe.Get("textoCigarrillo")))
	registro.Set("camposFaltantes", ListarCamposFaltantes(registro))
	return registro
}
