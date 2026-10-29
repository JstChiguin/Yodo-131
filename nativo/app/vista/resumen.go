package vista

import (
	"image"
	"image/color"
	"strconv"
	"strings"

	"gioui.org/layout"
	"gioui.org/unit"

	"yodo131/nucleo/estadisticas"
	"yodo131/nucleo/modelo"
)

var descripcionTirads = map[string]string{
	"1": "Normal", "2": "No sospechoso", "3": "Levemente sospechoso",
	"4": "Moderadamente sospechoso", "5": "Altamente sospechoso", "6": "Cáncer ya confirmado",
}
var descripcionBethesda = map[string]string{
	"I": "Muestra no diagnóstica", "II": "Benigno", "III": "Atipia de significado incierto",
	"IV": "Neoplasia folicular", "V": "Sospechoso de cáncer", "VI": "Maligno (cáncer)",
}
var nivelBethesda = map[string]int{"I": 1, "II": 2, "III": 3, "IV": 4, "V": 5, "VI": 6}

func (t *Tema) coloresDiagnostico(categoria string) color.NRGBA {
	switch categoria {
	case "Cáncer de tiroides":
		return t.Elevado
	case "Hipertiroidismo (Graves / bocio difuso)":
		return t.Acento
	case "Bocio multinodular o nódulo hiperfuncionante":
		return t.Serie3
	case "Otro / no determinado":
		return t.Serie5
	}
	return t.SinDato
}

func (t *Tema) coloresOrigen(categoria string) color.NRGBA {
	switch categoria {
	case "Word":
		return t.Acento
	case "Excel":
		return t.Hombres
	case "Word y Excel":
		return t.Serie3
	}
	return t.SinDato
}

func (t *Tema) seriesSexo() []serieApilada {
	return []serieApilada{
		{Clave: "mujeres", Nombre: "Mujeres", Color: t.Mujeres},
		{Clave: "hombres", Nombre: "Hombres", Color: t.Hombres},
		{Clave: "sinDato", Nombre: "Sexo sin dato", Color: t.SinDato, ColorTexto: t.Texto},
	}
}

func (t *Tema) leyendaSexo() []leyendaItem {
	return []leyendaItem{{"Mujeres", t.Mujeres}, {"Hombres", t.Hombres}, {"Sexo sin dato", t.SinDato}}
}

func colorPorNivel(t *Tema, nivel, alto, medio int) color.NRGBA {
	if nivel >= alto {
		return t.Elevado
	}
	if nivel >= medio {
		return t.Serie4
	}
	return t.Serie3
}

func formatoUmbral(valor float64) string {
	return strconv.FormatFloat(valor, 'g', -1, 64)
}

func totalFila(fila estadisticas.FilaAnio) int { return fila.Mujeres + fila.Hombres + fila.SinDato }

func conclusionAnios(est estadisticas.Estadisticas) string {
	conFecha := []estadisticas.FilaAnio{}
	for _, fila := range est.PorAnio {
		if fila.Categoria != "Sin fecha" {
			conFecha = append(conFecha, fila)
		}
	}
	if len(conFecha) == 0 {
		return ""
	}
	mujeres, conSexo := 0, 0
	for _, fila := range est.PorAnio {
		mujeres += fila.Mujeres
		conSexo += fila.Mujeres + fila.Hombres
	}
	mayor := conFecha[0]
	for _, fila := range conFecha {
		if totalFila(fila) > totalFila(mayor) {
			mayor = fila
		}
	}
	return "El año con más tratamientos fue " + mayor.Categoria + " (" + formatoNumero(totalFila(mayor)) + "). De los tratamientos con sexo registrado, " + formatoNumero(mujeres) + " de " + formatoNumero(conSexo) + " son de mujeres (" + porcentajeTexto(mujeres, conSexo) + ")."
}

func coberturaAnios(est estadisticas.Estadisticas) string {
	conSexo := 0
	for _, fila := range est.PorAnio {
		conSexo += fila.Mujeres + fila.Hombres
	}
	texto := "Sexo disponible en " + formatoNumero(conSexo) + " de " + formatoNumero(est.CantidadTratamientos) + " tratamientos (" + porcentajeTexto(conSexo, est.CantidadTratamientos) + ")."
	if est.RegistrosUnidos > 0 {
		texto += " Se unieron " + formatoNumero(est.RegistrosUnidos) + " registros que correspondían a un mismo tratamiento."
	}
	return texto
}

func conclusionCancer(est estadisticas.Estadisticas) (string, string) {
	total, mujeres, conSexo := 0, 0, 0
	for _, fila := range est.CancerPorAnio {
		total += totalFila(fila)
		mujeres += fila.Mujeres
		conSexo += fila.Mujeres + fila.Hombres
	}
	if total == 0 {
		return "No hay registros de cáncer de tiroides en este período.", ""
	}
	mayor := est.CancerPorAnio[0]
	for _, fila := range est.CancerPorAnio {
		if totalFila(fila) > totalFila(mayor) {
			mayor = fila
		}
	}
	conclusion := formatoNumero(total) + " pacientes con cáncer de tiroides en el período; " + formatoNumero(mujeres) + " de " + formatoNumero(conSexo) + " con sexo registrado son mujeres (" + porcentajeTexto(mujeres, conSexo) + "). El año con más casos nuevos fue " + mayor.Categoria + "."
	cobertura := "Sexo disponible en " + formatoNumero(conSexo) + " de " + formatoNumero(total) + " pacientes (" + porcentajeTexto(conSexo, total) + ")."
	return conclusion, cobertura
}

func describirExamen(conteo map[string]int) string {
	conDato := conteo["Elevado"] + conteo["Normal"]
	total := conDato + conteo["Sin dato"]
	if conDato == 0 {
		return "Ningún tratamiento trae este examen"
	}
	return "Elevado en " + formatoNumero(conteo["Elevado"]) + " de " + formatoNumero(conDato) + " con el examen (" + porcentajeTexto(conteo["Elevado"], conDato) + ") · registrado en " + formatoNumero(conDato) + " de " + formatoNumero(total) + " (" + porcentajeTexto(conDato, total) + ")"
}

func describirSiNo(conteo map[string]int) string {
	conDato := conteo["Sí"] + conteo["No"]
	total := conDato + conteo["Sin dato"]
	if conDato == 0 {
		return "Ningún tratamiento trae el dato"
	}
	return formatoNumero(conteo["Sí"]) + " de " + formatoNumero(conDato) + " con el dato (" + porcentajeTexto(conteo["Sí"], conDato) + ") · registrado en " + formatoNumero(conDato) + " de " + formatoNumero(total) + " (" + porcentajeTexto(conDato, total) + ")"
}

func detalleGrupo(conteo map[string]int, orden []string, fuera []string) string {
	partes := []string{}
	for _, categoria := range orden {
		if contieneTexto(fuera, categoria) {
			partes = append(partes, categoria+" (no se grafica): "+formatoNumero(conteo[categoria]))
		} else {
			partes = append(partes, categoria+": "+formatoNumero(conteo[categoria]))
		}
	}
	return strings.Join(partes, " · ")
}

func (t *Tema) grupoEstado(nombre string, conteo map[string]int, resumen string) grupoProp {
	return grupoProp{
		Nombre:  nombre,
		Resumen: resumen,
		Tramos:  []tramoProp{{"Elevado", conteo["Elevado"], t.Elevado}, {"Normal", conteo["Normal"], t.Normal}},
		Total:   conteo["Elevado"] + conteo["Normal"],
		Detalle: detalleGrupo(conteo, []string{"Elevado", "Normal", "Sin dato"}, []string{"Sin dato"}),
	}
}

func (t *Tema) grupoSiNo(nombre string, conteo map[string]int, resumen string) grupoProp {
	return grupoProp{
		Nombre:  nombre,
		Resumen: resumen,
		Tramos:  []tramoProp{{"Sí", conteo["Sí"], t.Serie4}, {"No", conteo["No"], t.Normal}},
		Total:   conteo["Sí"] + conteo["No"],
		Detalle: detalleGrupo(conteo, []string{"Sí", "No", "Sin dato"}, []string{"Sin dato"}),
	}
}

func filasDistribucion(dist estadisticas.Distribucion, color func(estadisticas.FilaDist) color.NRGBA, desc func(estadisticas.FilaDist) string) []filaBarra {
	filas := make([]filaBarra, len(dist.Filas))
	for i, fila := range dist.Filas {
		descripcion := ""
		if desc != nil {
			descripcion = desc(fila)
		}
		filas[i] = filaBarra{Categoria: fila.Categoria, Cantidad: fila.Cantidad, Descripcion: descripcion, Color: color(fila)}
	}
	return filas
}

type especificacionTarjeta struct {
	ancha     bool
	contenido layout.Widget
}

func (t *Tema) tarjetaDistribucion(titulo, nota, conclusion string, dist estadisticas.Distribucion, colorear func(estadisticas.FilaDist) color.NRGBA, desc func(estadisticas.FilaDist) string, unidadCobertura string, presentacion bool) layout.Widget {
	filas := filasDistribucion(dist, colorear, desc)
	cobertura := describirCobertura(dist.ConDato, dist.Total, unidadCobertura)
	return func(gtx C) D {
		return t.tarjeta(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D { return t.encabezadoTarjeta(gtx, titulo, nota) }),
				layout.Rigid(func(gtx C) D { return t.espacio(gtx, 8) }),
				layout.Rigid(func(gtx C) D { return t.cajaConclusion(gtx, conclusion, presentacion) }),
				layout.Rigid(func(gtx C) D { return t.espacioSi(gtx, 6, conclusion != "") }),
				layout.Rigid(func(gtx C) D { return t.textoCobertura(gtx, cobertura) }),
				layout.Rigid(func(gtx C) D { return t.espacio(gtx, 8) }),
				layout.Rigid(func(gtx C) D { return t.barrasHorizontales(gtx, filas, dist.ConDato) }),
			)
		})
	}
}

func (t *Tema) espacio(gtx C, dp unit.Dp) D {
	return layout.Spacer{Height: dp}.Layout(gtx)
}

func (t *Tema) espacioSi(gtx C, dp unit.Dp, condicion bool) D {
	if !condicion {
		return D{}
	}
	return layout.Spacer{Height: dp}.Layout(gtx)
}

func (t *Tema) tarjetaAnios(titulo, nota, conclusion, cobertura string, filas []estadisticas.FilaAnio, presentacion bool) layout.Widget {
	series := t.seriesSexo()
	leyenda := t.leyendaSexo()
	return func(gtx C) D {
		return t.tarjeta(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D { return t.encabezadoTarjeta(gtx, titulo, nota) }),
				layout.Rigid(func(gtx C) D { return t.espacio(gtx, 8) }),
				layout.Rigid(func(gtx C) D { return t.cajaConclusion(gtx, conclusion, presentacion) }),
				layout.Rigid(func(gtx C) D { return t.espacioSi(gtx, 6, conclusion != "") }),
				layout.Rigid(func(gtx C) D { return t.textoCobertura(gtx, cobertura) }),
				layout.Rigid(func(gtx C) D { return t.espacio(gtx, 8) }),
				layout.Rigid(func(gtx C) D { return t.leyenda(gtx, leyenda) }),
				layout.Rigid(func(gtx C) D { return t.espacio(gtx, 6) }),
				layout.Rigid(func(gtx C) D { return t.barrasApiladas(gtx, filas, series) }),
			)
		})
	}
}

func (t *Tema) tarjetaProporciones(titulo, nota, conclusion string, leyenda []leyendaItem, grupos []grupoProp, presentacion bool) layout.Widget {
	return func(gtx C) D {
		return t.tarjeta(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D { return t.encabezadoTarjeta(gtx, titulo, nota) }),
				layout.Rigid(func(gtx C) D { return t.espacio(gtx, 8) }),
				layout.Rigid(func(gtx C) D { return t.cajaConclusion(gtx, conclusion, presentacion) }),
				layout.Rigid(func(gtx C) D { return t.espacioSi(gtx, 8, conclusion != "" || len(leyenda) > 0) }),
				layout.Rigid(func(gtx C) D {
					if len(leyenda) == 0 {
						return D{}
					}
					return t.leyenda(gtx, leyenda)
				}),
				layout.Rigid(func(gtx C) D { return t.espacioSi(gtx, 6, len(leyenda) > 0) }),
				layout.Rigid(func(gtx C) D { return t.proporciones(gtx, grupos) }),
			)
		})
	}
}

func (t *Tema) construirTarjetasResumen(est estadisticas.Estadisticas, umbrales map[string]float64, presentacion bool) []especificacionTarjeta {
	tarjetas := []especificacionTarjeta{}
	agregar := func(ancha bool, solorevision bool, w layout.Widget) {
		if solorevision && presentacion {
			return
		}
		tarjetas = append(tarjetas, especificacionTarjeta{ancha: ancha, contenido: w})
	}

	agregar(true, false, t.tarjetaAnios("Tratamientos con yodo 131 por año, según sexo",
		"Cada barra es un año. Una persona tratada dos veces cuenta dos tratamientos; un mismo tratamiento registrado en varios archivos cuenta una sola vez.",
		conclusionAnios(est), coberturaAnios(est), est.PorAnio, presentacion))

	conclusionCan, coberturaCan := conclusionCancer(est)
	agregar(true, false, t.tarjetaAniosCancer("Pacientes con cáncer de tiroides por año",
		"Cada paciente se cuenta una vez, en el año de su primer registro de cáncer (registro C73, informe o planilla de yodo).",
		conclusionCan, coberturaCan, est.CancerPorAnio, presentacion))

	agregar(false, false, t.tarjetaDistribucion("Diagnóstico", "Diagnóstico de cada tratamiento.",
		describirMasFrecuente(est.PorDiagnostico, "Lo más frecuente", "tratamientos con diagnóstico", nil),
		est.PorDiagnostico, func(f estadisticas.FilaDist) color.NRGBA { return t.coloresDiagnostico(f.Categoria) }, nil, unidadResumen, presentacion))

	agregar(false, false, t.tarjetaDistribucion("Motivo de consulta", "Por qué consultó el paciente, según el informe.",
		describirMasFrecuente(est.PorMotivo, "El motivo más frecuente", "tratamientos con motivo escrito", []string{"Otro"}),
		est.PorMotivo, func(estadisticas.FilaDist) color.NRGBA { return t.Acento }, nil, unidadResumen, presentacion))

	agregar(false, false, t.tarjetaDistribucion("Edad", "Edad del paciente en cada tratamiento, en rangos de 10 años.",
		conclusionEdad(est), est.PorEdad, func(estadisticas.FilaDist) color.NRGBA { return t.Hombres }, nil, unidadResumen, presentacion))

	agregar(false, false, t.tarjetaDistribucion("Dosis de yodo 131", "Las dosis bajas suelen usarse en hipertiroidismo; las altas, en el tratamiento del cáncer.",
		describirMasFrecuente(est.PorDosis, "La dosis más indicada", "tratamientos con dosis", []string{"Rechazó / no recibió"}),
		est.PorDosis, func(f estadisticas.FilaDist) color.NRGBA { return t.colorDosis(f.Categoria) }, nil, unidadResumen, presentacion))

	agregar(true, false, t.tarjetaProporciones("Resultados de exámenes", notaUmbrales(umbrales),
		conclusionExamenes(est),
		[]leyendaItem{{"Elevado", t.Elevado}, {"Normal", t.Normal}},
		[]grupoProp{
			t.grupoEstado("TRAb", est.Examenes["TRAb"], describirExamen(est.Examenes["TRAb"])),
			t.grupoEstado("Anti-tiroglobulina", est.Examenes["Anti-tiroglobulina"], describirExamen(est.Examenes["Anti-tiroglobulina"])),
			t.grupoEstado("Anti-TPO", est.Examenes["Anti-TPO"], describirExamen(est.Examenes["Anti-TPO"])),
			t.grupoEstado("Tiroglobulina", est.Examenes["Tiroglobulina"], describirExamen(est.Examenes["Tiroglobulina"])),
		}, presentacion))

	agregar(false, false, t.tarjetaTirads(est, presentacion))

	agregar(false, false, t.tarjetaDistribucion("Punción (Bethesda)", "Resultado de la punción (PAAF): a mayor número, mayor sospecha de cáncer.",
		conclusionBethesda(est), distribucionBethesda(est.PorBethesda),
		func(f estadisticas.FilaDist) color.NRGBA {
			return colorPorNivel(t, nivelBethesda[strings.TrimPrefix(f.Categoria, "Bethesda ")], 5, 3)
		},
		func(f estadisticas.FilaDist) string {
			return descripcionBethesda[strings.TrimPrefix(f.Categoria, "Bethesda ")]
		}, unidadResumen, presentacion))

	agregar(false, false, t.tarjetaDistribucion("Tipo de cáncer de tiroides", "Según biopsia o punción, solo en los tratamientos con cáncer.",
		describirMasFrecuente(est.PorHistologia, "El tipo más frecuente", "con tipo informado", nil),
		est.PorHistologia, func(estadisticas.FilaDist) color.NRGBA { return t.Elevado }, nil, "tratamientos con cáncer", presentacion))

	agregar(false, false, t.tarjetaProporciones("Antecedentes", "Antecedente familiar de enfermedad tiroidea y tabaquismo.",
		conclusionAntecedentes(est),
		[]leyendaItem{{"Sí", t.Serie4}, {"No", t.Normal}},
		[]grupoProp{
			t.grupoSiNo("Antecedente familiar tiroideo", est.Antecedentes["Antecedente familiar tiroideo"], describirSiNo(est.Antecedentes["Antecedente familiar tiroideo"])),
			t.grupoSiNo("Tabaquismo actual o previo", est.Antecedentes["Tabaquismo actual o previo"], describirSiNo(est.Antecedentes["Tabaquismo actual o previo"])),
		}, presentacion))

	agregar(false, false, t.tarjetaDistribucion("Médico tratante", "Tratamientos por médico tratante.",
		describirMasFrecuente(est.PorMedico, "Con más tratamientos", "con médico registrado", nil),
		est.PorMedico, func(estadisticas.FilaDist) color.NRGBA { return t.Hombres }, nil, unidadResumen, presentacion))

	agregar(false, true, t.tarjetaDistribucion("Origen de los datos", "De dónde sale cada tratamiento: informe Word, planilla Excel o ambos unidos.",
		conclusionOrigen(est), est.PorOrigen, func(f estadisticas.FilaDist) color.NRGBA { return t.coloresOrigen(f.Categoria) }, nil, unidadResumen, presentacion))

	return tarjetas
}

func (t *Tema) tarjetaAniosCancer(titulo, nota, conclusion, cobertura string, filas []estadisticas.FilaAnio, presentacion bool) layout.Widget {
	return t.tarjetaAnios(titulo, nota, conclusion, cobertura, filas, presentacion)
}

func (t *Tema) colorDosis(categoria string) color.NRGBA {
	switch {
	case strings.HasPrefix(categoria, "Alta"):
		return t.Elevado
	case strings.HasPrefix(categoria, "Intermedia"):
		return t.Serie4
	case strings.HasPrefix(categoria, "Baja"):
		return t.Acento
	}
	return t.SinDato
}

func conclusionEdad(est estadisticas.Estadisticas) string {
	texto := ""
	if mayor, ok := obtenerMasFrecuente(est.PorEdad, nil); ok {
		texto = "El grupo más numeroso es el de " + mayor.categoria + " años, con " + mayor.detalle + " tratamientos con edad (" + mayor.porcentaje + ")."
	}
	if est.EdadPromedio != nil {
		promedio := int(est.EdadPromedio.(float64) + 0.5)
		texto = strings.TrimSpace(texto + " Edad promedio: " + strconv.Itoa(promedio) + " años.")
	}
	return strings.TrimSpace(texto)
}

func notaUmbrales(umbrales map[string]float64) string {
	return "Se considera elevado: TRAb > " + formatoUmbral(umbrales["trab"]) + "; tiroglobulina > " + formatoUmbral(umbrales["tiroglobulina"]) + " ng/ml o \"en ascenso\"; anti-tiroglobulina > " + formatoUmbral(umbrales["antiTiroglobulina"]) + "; anti-TPO > " + formatoUmbral(umbrales["antiTpo"]) + ". Cada barra muestra solo los tratamientos que traen el examen; un examen no registrado no cuenta como normal."
}

func conclusionExamenes(est estadisticas.Estadisticas) string {
	masNombre := ""
	masConDato := -1
	for _, nombre := range est.OrdenExamenes {
		conteo := est.Examenes[nombre]
		conDato := conteo["Elevado"] + conteo["Normal"]
		if conDato > 0 && conDato > masConDato {
			masConDato = conDato
			masNombre = nombre
		}
	}
	if masNombre == "" {
		return "Ningún tratamiento del período trae estos exámenes. Las planillas Excel no los incluyen; solo vienen en los informes Word."
	}
	return "El examen más registrado es " + masNombre + ": está en " + formatoNumero(masConDato) + " de " + formatoNumero(est.CantidadTratamientos) + " tratamientos (" + porcentajeTexto(masConDato, est.CantidadTratamientos) + ")."
}

func (t *Tema) tarjetaTirads(est estadisticas.Estadisticas, presentacion bool) layout.Widget {
	ultimo := func(categoria string) int {
		r := []rune(categoria)
		n, _ := strconv.Atoi(string(r[len(r)-1]))
		return n
	}
	filas := filasDistribucion(est.PorTirads,
		func(f estadisticas.FilaDist) color.NRGBA { return colorPorNivel(t, ultimo(f.Categoria), 5, 4) },
		func(f estadisticas.FilaDist) string {
			r := []rune(f.Categoria)
			return descripcionTirads[string(r[len(r)-1])]
		})
	conclusion := ""
	if est.PorTirads.ConDato > 0 {
		altaSospecha := sumarCategorias(est.PorTirads, func(f estadisticas.FilaDist) bool { return ultimo(f.Categoria) >= 5 })
		conclusion = "Alta sospecha (TIRADS 5 o más): " + formatoNumero(altaSospecha) + " de " + formatoNumero(est.PorTirads.ConDato) + " ecografías clasificadas (" + porcentajeTexto(altaSospecha, est.PorTirads.ConDato) + ")."
	}
	cobertura := describirCobertura(est.PorTirads.ConDato, est.PorTirads.Total, unidadResumen)
	grupos := []grupoProp{
		t.grupoSiNo("Con nódulos", est.Ecografia["Con nódulos"], describirSiNo(est.Ecografia["Con nódulos"])),
		t.grupoSiNo("Con adenopatías", est.Ecografia["Con adenopatías"], describirSiNo(est.Ecografia["Con adenopatías"])),
	}
	return func(gtx C) D {
		return t.tarjeta(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return t.encabezadoTarjeta(gtx, "Ecografía tiroidea (TIRADS)", "Clasificación TI-RADS de la ecografía: a mayor número, mayor sospecha de cáncer.")
				}),
				layout.Rigid(func(gtx C) D { return t.espacio(gtx, 8) }),
				layout.Rigid(func(gtx C) D { return t.cajaConclusion(gtx, conclusion, presentacion) }),
				layout.Rigid(func(gtx C) D { return t.espacioSi(gtx, 6, conclusion != "") }),
				layout.Rigid(func(gtx C) D { return t.textoCobertura(gtx, cobertura) }),
				layout.Rigid(func(gtx C) D { return t.espacio(gtx, 8) }),
				layout.Rigid(func(gtx C) D { return t.barrasHorizontales(gtx, filas, est.PorTirads.ConDato) }),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Top: 8}.Layout(gtx, func(gtx C) D {
						rellenarRect(gtx, image.Rect(0, 0, gtx.Constraints.Max.X, 1), t.Rejilla, 0)
						return layout.Inset{Top: 8}.Layout(gtx, func(gtx C) D { return t.proporciones(gtx, grupos) })
					})
				}),
			)
		})
	}
}

func distribucionBethesda(dist estadisticas.Distribucion) estadisticas.Distribucion {
	filas := make([]estadisticas.FilaDist, len(dist.Filas))
	for i, fila := range dist.Filas {
		filas[i] = estadisticas.FilaDist{Categoria: "Bethesda " + fila.Categoria, Cantidad: fila.Cantidad}
	}
	return estadisticas.Distribucion{Filas: filas, ConDato: dist.ConDato, SinDato: dist.SinDato, Total: dist.Total}
}

func conclusionBethesda(est estadisticas.Estadisticas) string {
	dist := distribucionBethesda(est.PorBethesda)
	if dist.ConDato == 0 {
		return ""
	}
	sospechosas := sumarCategorias(dist, func(f estadisticas.FilaDist) bool {
		return nivelBethesda[strings.TrimPrefix(f.Categoria, "Bethesda ")] >= 5
	})
	return "Sospechosas o malignas (Bethesda V y VI): " + formatoNumero(sospechosas) + " de " + formatoNumero(dist.ConDato) + " punciones informadas (" + porcentajeTexto(sospechosas, dist.ConDato) + ")."
}

func conclusionAntecedentes(est estadisticas.Estadisticas) string {
	familiar := est.Antecedentes["Antecedente familiar tiroideo"]
	conDato := familiar["Sí"] + familiar["No"]
	if conDato == 0 {
		return ""
	}
	return "Tienen familiares con enfermedad tiroidea: " + formatoNumero(familiar["Sí"]) + " de " + formatoNumero(conDato) + " tratamientos con el dato (" + porcentajeTexto(familiar["Sí"], conDato) + ")."
}

func conclusionOrigen(est estadisticas.Estadisticas) string {
	unidos := sumarCategorias(est.PorOrigen, func(f estadisticas.FilaDist) bool { return f.Categoria == "Word y Excel" })
	if unidos == 0 {
		return ""
	}
	return formatoNumero(unidos) + " tratamientos tienen datos de Word y de Excel a la vez; sus datos se combinaron y las diferencias quedaron en \"Por revisar\"."
}

func (v *Vista) cifrasResumen(gtx C, t *Tema, est estadisticas.Estadisticas, conteo modelo.Conteo, presentacion bool) D {
	fuentes := []string{}
	if conteo.InformesWord > 0 {
		fuentes = append(fuentes, formatoNumero(conteo.InformesWord)+" informes Word")
	}
	if conteo.FilasExcelYodo > 0 {
		fuentes = append(fuentes, formatoNumero(conteo.FilasExcelYodo)+" filas de Excel")
	}
	fuentesTexto := strings.Join(fuentes, " y ")
	if fuentesTexto == "" {
		fuentesTexto = "ningún registro"
	}
	edadValor := "—"
	edadNota := "Sin edades"
	if est.EdadPromedio != nil {
		edadValor = strconv.Itoa(int(est.EdadPromedio.(float64)+0.5)) + " años"
		edadNota = "Entre " + strconv.Itoa(est.EdadMinima.(int)) + " y " + strconv.Itoa(est.EdadMaxima.(int)) + " años · dato en " + formatoNumero(est.CoberturaEdad.ConDato) + " de " + formatoNumero(est.CoberturaEdad.Total)
	}
	cancerNota := "Según informes y planillas de yodo"
	if conteo.FilasRegistroCancer > 0 {
		cancerNota = "Registro C73, informes y planillas de yodo"
	}
	type tile struct {
		etiqueta, valor, nota string
		solorevision          bool
	}
	tiles := []tile{
		{"Tratamientos con yodo 131", formatoNumero(est.CantidadTratamientos), "De " + fuentesTexto + "; " + formatoNumero(est.RegistrosUnidos) + " registros repetidos se unieron", false},
		{"Pacientes únicos", formatoNumero(est.CantidadPacientes), formatoNumero(est.PacientesConVariosTratamientos) + " con más de un tratamiento", false},
		{"Mujeres", formatoNumero(est.PacientesMujeres), "Hombres: " + formatoNumero(est.PacientesHombres) + " · Sin dato: " + formatoNumero(est.PacientesSinSexo), false},
		{"Edad promedio", edadValor, edadNota, false},
		{"Pacientes con cáncer de tiroides", formatoNumero(est.CantidadPacientesCancer), cancerNota, false},
		{"Datos que no coinciden", formatoNumero(est.CantidadConflictos), "Entre fuentes del mismo tratamiento · ver Por revisar", true},
	}
	visibles := []tile{}
	for _, tl := range tiles {
		if tl.solorevision && presentacion {
			continue
		}
		visibles = append(visibles, tl)
	}
	return t.enRejilla(gtx, len(visibles), gtx.Dp(168), func(gtx C, indice int) D {
		tl := visibles[indice]
		return v.conResalto(gtx, t, "cifra-"+strconv.Itoa(indice), presentacion, func(gtx C, resaltado float32) D {
			colorValor := t.Texto
			if resaltado > 0 {
				colorValor = mezclar(t.Acento, t.Texto, float64(resaltado))
			}
			return t.cifraColor(gtx, tl.etiqueta, tl.valor, tl.nota, colorValor)
		})
	})
}
