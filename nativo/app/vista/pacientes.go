package vista

import (
	"image/color"
	"sort"
	"strconv"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"yodo131/nucleo/modelo"
)

type pacientesEstado struct {
	buscador     widget.Editor
	ordenCampo   string
	ordenAsc     bool
	botonesOrden map[string]*widget.Clickable
	filaClic     map[string]*widget.Clickable
	expandidas   map[string]bool
}

func (p *pacientesEstado) inicializar() {
	p.ordenCampo = "fechaOrden"
	p.ordenAsc = false
	p.buscador.SingleLine = true
	p.botonesOrden = map[string]*widget.Clickable{}
	p.filaClic = map[string]*widget.Clickable{}
	p.expandidas = map[string]bool{}
}

type columnaTabla struct {
	titulo string
	orden  string
	peso   float32
}

var columnasTabla = []columnaTabla{
	{"Paciente", "nombrePaciente", 2.2},
	{"RUT", "rut", 1.2},
	{"Edad", "edad", 0.5},
	{"Sexo", "sexo", 0.6},
	{"Fecha", "fechaOrden", 1.0},
	{"Diagnóstico", "diagnostico", 2.2},
	{"Dosis", "dosisMci", 1.0},
	{"TIRADS", "tirads", 0.7},
	{"Bethesda", "bethesda", 0.8},
	{"Fuente", "", 1.0},
	{"Exámenes", "", 1.8},
	{"Revisar", "", 1.6},
}

func textoMedicionVista(valor any) string {
	m, ok := valor.(modelo.Medicion)
	if !ok {
		return ""
	}
	return m.Comparador + strings.Replace(strconv.FormatFloat(m.Valor, 'f', -1, 64), ".", ",", 1)
}

func fechaBonita(iso string) string {
	partes := strings.Split(iso, "-")
	if len(partes) != 3 {
		return iso
	}
	return partes[2] + "/" + partes[1] + "/" + partes[0]
}

func textoOpcional(r modelo.Objeto, clave string) string {
	if s, ok := r.Get(clave).(string); ok {
		return s
	}
	return ""
}

func (p *pacientesEstado) botonOrden(campo string) *widget.Clickable {
	if p.botonesOrden[campo] == nil {
		p.botonesOrden[campo] = &widget.Clickable{}
	}
	return p.botonesOrden[campo]
}

func (p *pacientesEstado) clicFila(clave string) *widget.Clickable {
	if p.filaClic[clave] == nil {
		p.filaClic[clave] = &widget.Clickable{}
	}
	return p.filaClic[clave]
}

func valorCeldaTexto(r modelo.Objeto, indice int) string {
	switch indice {
	case 0:
		return r.Texto("nombrePaciente")
	case 1:
		return textoOpcional(r, "rut")
	case 2:
		if e, ok := r.Entero("edad"); ok {
			return strconv.Itoa(e)
		}
		return ""
	case 3:
		return textoOpcional(r, "sexo")
	case 4:
		if iso, ok := r.Get("fechaIso").(string); ok && iso != "" {
			return fechaBonita(iso)
		}
		if a, ok := r.Entero("anio"); ok {
			return strconv.Itoa(a)
		}
		return ""
	case 5:
		diag := textoOpcional(r, "diagnostico")
		histo := textoOpcional(r, "histologia")
		if histo != "" && histo != "No especificado" {
			return diag + " · " + histo
		}
		return diag
	case 6:
		if d, ok := r.Numero("dosisMci"); ok {
			texto := strconv.FormatFloat(d, 'f', -1, 64) + " mCi"
			if r.Bool("rechazoYodo") {
				texto += " (rechazó)"
			}
			return texto
		}
		if r.Bool("rechazoYodo") {
			return "Rechazó"
		}
		return ""
	case 7:
		if v, ok := r.Entero("tirads"); ok {
			return strconv.Itoa(v)
		}
		return ""
	case 8:
		return textoOpcional(r, "bethesda")
	case 9:
		origenes, _ := r.Get("origenes").([]string)
		if len(origenes) > 1 {
			return "Word y Excel"
		}
		if len(origenes) == 1 {
			return origenes[0]
		}
		return ""
	}
	return ""
}

func (v *Vista) filasPacientes(datos Datos) []modelo.Objeto {
	busqueda := strings.ToLower(strings.TrimSpace(v.pacientes.buscador.Text()))
	filas := []modelo.Objeto{}
	for _, r := range datos.Resultado.Tratamientos {
		if busqueda == "" {
			filas = append(filas, r)
			continue
		}
		campos := []string{r.Texto("nombrePaciente"), textoOpcional(r, "rut"), textoOpcional(r, "diagnostico"), textoOpcional(r, "histologia"), textoOpcional(r, "motivoCategoria"), textoOpcional(r, "medicoTratante")}
		if fuentes, ok := r.Get("fuentes").([]modelo.Fuente); ok {
			for _, f := range fuentes {
				campos = append(campos, f.Referencia)
			}
		}
		for _, campo := range campos {
			if campo != "" && strings.Contains(strings.ToLower(campo), busqueda) {
				filas = append(filas, r)
				break
			}
		}
	}
	p := &v.pacientes
	valor := func(r modelo.Objeto) any {
		if p.ordenCampo == "fechaOrden" {
			if iso, ok := r.Get("fechaIso").(string); ok && iso != "" {
				return iso
			}
			if a, ok := r.Entero("anio"); ok {
				return strconv.Itoa(a)
			}
			return ""
		}
		return r.Get(p.ordenCampo)
	}
	vacio := func(x any) bool {
		if x == nil {
			return true
		}
		if s, ok := x.(string); ok {
			return s == ""
		}
		return false
	}
	sort.SliceStable(filas, func(i, j int) bool {
		a := valor(filas[i])
		b := valor(filas[j])
		if vacio(a) && vacio(b) {
			return false
		}
		if vacio(a) {
			return false
		}
		if vacio(b) {
			return true
		}
		menor := comparar(a, b)
		if !p.ordenAsc {
			menor = -menor
		}
		return menor < 0
	})
	return filas
}

func comparar(a, b any) int {
	na, okA := numeroComparable(a)
	nb, okB := numeroComparable(b)
	if okA && okB {
		switch {
		case na < nb:
			return -1
		case na > nb:
			return 1
		default:
			return 0
		}
	}
	sa := strings.ToLower(textoComparable(a))
	sb := strings.ToLower(textoComparable(b))
	return strings.Compare(sa, sb)
}

func numeroComparable(x any) (float64, bool) {
	switch v := x.(type) {
	case int:
		return float64(v), true
	case float64:
		return v, true
	}
	return 0, false
}

func textoComparable(x any) string {
	switch v := x.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	return ""
}

func (v *Vista) layoutPacientes(gtx C, t *Tema, datos Datos) D {
	for campo, clic := range v.pacientes.botonesOrden {
		if clic.Clicked(gtx) {
			if v.pacientes.ordenCampo == campo {
				v.pacientes.ordenAsc = !v.pacientes.ordenAsc
			} else {
				v.pacientes.ordenCampo = campo
				v.pacientes.ordenAsc = true
			}
		}
	}
	filas := v.filasPacientes(datos)
	for clave, clic := range v.pacientes.filaClic {
		if clic.Clicked(gtx) {
			v.pacientes.expandidas[clave] = !v.pacientes.expandidas[clave]
		}
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D { return v.barraTabla(gtx, t, len(filas)) }),
		layout.Rigid(func(gtx C) D { return layout.Spacer{Height: 10}.Layout(gtx) }),
		layout.Rigid(func(gtx C) D {
			return t.tarjeta(gtx, func(gtx C) D {
				hijos := []layout.FlexChild{layout.Rigid(func(gtx C) D { return v.encabezadoTabla(gtx, t) })}
				for _, r := range filas {
					r := r
					hijos = append(hijos, layout.Rigid(func(gtx C) D { return v.filaTabla(gtx, t, r) }))
				}
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx, hijos...)
			})
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Top: 8}.Layout(gtx, func(gtx C) D {
				return t.Nota("Cada fila es un tratamiento. Haga clic en una fila para ver de qué archivos salió y el texto que se leyó.").Layout(gtx)
			})
		}),
	)
}

func (v *Vista) barraTabla(gtx C, t *Tema, cantidad int) D {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Flexed(1, func(gtx C) D {
			return widget.Border{Color: t.BordeFuerte, CornerRadius: 8, Width: 1}.Layout(gtx, func(gtx C) D {
				return fondoRedondeado(gtx, t.Panel, 8, func(gtx C) D {
					return layout.Inset{Left: 12, Right: 12, Top: 8, Bottom: 8}.Layout(gtx, func(gtx C) D {
						ed := material.Editor(t.Th, &v.pacientes.buscador, "Buscar por nombre, RUT, diagnóstico…")
						ed.Color = t.Texto
						ed.HintColor = t.Suave
						ed.Font.Typeface = TipografiaCuerpo
						return ed.Layout(gtx)
					})
				})
			})
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Left: 16}.Layout(gtx, func(gtx C) D {
				return t.Nota(formatoNumero(cantidad) + " tratamiento(s) · Todos los tratamientos").Layout(gtx)
			})
		}),
	)
}

func (v *Vista) encabezadoTabla(gtx C, t *Tema) D {
	hijos := make([]layout.FlexChild, 0, len(columnasTabla))
	for indice, col := range columnasTabla {
		col := col
		indice := indice
		hijos = append(hijos, layout.Flexed(col.peso, func(gtx C) D {
			return layout.Inset{Right: 8, Top: 6, Bottom: 8}.Layout(gtx, func(gtx C) D {
				if col.orden == "" {
					lbl := t.etiquetaTamano(13, col.titulo, t.Texto, font.SemiBold)
					return lbl.Layout(gtx)
				}
				texto := col.titulo
				if v.pacientes.ordenCampo == col.orden {
					if v.pacientes.ordenAsc {
						texto += " ▲"
					} else {
						texto += " ▼"
					}
				}
				return material.Clickable(gtx, v.pacientes.botonOrden(col.orden), func(gtx C) D {
					lbl := t.etiquetaTamano(13, texto, t.Texto, font.SemiBold)
					return lbl.Layout(gtx)
				})
			})
		}))
		_ = indice
	}
	return func(gtx C) D {
		dims := layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx, hijos...)
		rellenarRectBorde(gtx, t, dims)
		return dims
	}(gtx)
}

func rellenarRectBorde(gtx C, t *Tema, dims D) {
	rellenarRectLinea(gtx, 0, dims.Size.Y-1, gtx.Constraints.Max.X, dims.Size.Y, t.BordeFuerte)
}

func (v *Vista) filaTabla(gtx C, t *Tema, r modelo.Objeto) D {
	clave := r.Texto("idTratamiento")
	expandida := v.pacientes.expandidas[clave]
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return material.Clickable(gtx, v.pacientes.clicFila(clave), func(gtx C) D {
				hijos := make([]layout.FlexChild, 0, len(columnasTabla))
				for indice, col := range columnasTabla {
					indice := indice
					col := col
					hijos = append(hijos, layout.Flexed(col.peso, func(gtx C) D {
						return layout.Inset{Right: 8, Top: 8, Bottom: 8}.Layout(gtx, func(gtx C) D {
							switch indice {
							case 10:
								return v.celdaExamenes(gtx, t, r)
							case 11:
								return v.celdaRevisar(gtx, t, r)
							default:
								return t.etiqueta(13, valorCeldaTexto(r, indice), t.Texto, TipografiaCuerpo).Layout(gtx)
							}
						})
					}))
				}
				dims := layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx, hijos...)
				rellenarRectLinea(gtx, 0, dims.Size.Y-1, gtx.Constraints.Max.X, dims.Size.Y, t.Borde)
				return dims
			})
		}),
		layout.Rigid(func(gtx C) D {
			if !expandida {
				return D{}
			}
			return v.filaDetalle(gtx, t, r)
		}),
	)
}

func (v *Vista) celdaExamenes(gtx C, t *Tema, r modelo.Objeto) D {
	chips := []layout.FlexChild{}
	agregar := func(nombre string, elevado any, valorTexto string) {
		b, ok := elevado.(bool)
		if !ok {
			return
		}
		texto := strings.TrimSpace(nombre + " " + valorTexto)
		colorChip := t.Normal
		if b {
			texto += " ↑"
			colorChip = t.Elevado
		}
		chips = append(chips, layout.Rigid(func(gtx C) D {
			return layout.Inset{Bottom: 3}.Layout(gtx, func(gtx C) D { return t.chip(gtx, texto, colorChip) })
		}))
	}
	tg := ""
	if vm := r.Tiroglobulina().ValorMaximo; vm != nil {
		tg = strings.Replace(strconv.FormatFloat(vm.(float64), 'f', -1, 64), ".", ",", 1)
	}
	agregar("TRAb", r.Get("trabElevado"), textoMedicionVista(r.Get("trab")))
	agregar("Tg", r.Get("tiroglobulinaElevada"), tg)
	agregar("Anti-Tg", r.Get("antiTiroglobulinaElevado"), textoMedicionVista(r.Get("antiTiroglobulina")))
	agregar("Anti-TPO", r.Get("antiTpoElevado"), textoMedicionVista(r.Get("antiTpo")))
	if len(chips) == 0 {
		return D{}
	}
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}.Layout(gtx, chips...)
}

func (v *Vista) celdaRevisar(gtx C, t *Tema, r modelo.Objeto) D {
	chips := []layout.FlexChild{}
	conflictos, _ := r.Get("conflictos").([]modelo.Conflicto)
	if len(conflictos) > 0 {
		vistos := map[string]bool{}
		nombres := []string{}
		for _, c := range conflictos {
			if !vistos[c.NombreCampo] {
				vistos[c.NombreCampo] = true
				nombres = append(nombres, c.NombreCampo)
			}
		}
		chips = append(chips, layout.Rigid(func(gtx C) D {
			return layout.Inset{Bottom: 2}.Layout(gtx, func(gtx C) D { return t.chip(gtx, "No coincide: "+strings.Join(nombres, ", "), t.Elevado) })
		}))
	}
	if campos, ok := r.Get("camposFaltantes").([]string); ok && len(campos) > 0 {
		chips = append(chips, layout.Rigid(func(gtx C) D {
			return t.chip(gtx, "Falta "+strings.Join(campos, ", "), t.Elevado)
		}))
	}
	if len(chips) == 0 {
		return D{}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, chips...)
}

func (t *Tema) chip(gtx C, texto string, c color.NRGBA) D {
	fondo := mezclar(c, t.Panel, 0.16)
	return fondoRedondeado(gtx, fondo, 10, func(gtx C) D {
		return layout.Inset{Left: 8, Right: 8, Top: 1, Bottom: 1}.Layout(gtx, func(gtx C) D {
			lbl := t.etiqueta(12, texto, c, TipografiaCuerpo)
			lbl.MaxLines = 1
			return lbl.Layout(gtx)
		})
	})
}

func (v *Vista) filaDetalle(gtx C, t *Tema, r modelo.Objeto) D {
	campos := detalleTratamiento(r)
	hijos := make([]layout.FlexChild, 0, len(campos))
	for _, campo := range campos {
		campo := campo
		hijos = append(hijos, layout.Rigid(func(gtx C) D {
			return layout.Inset{Top: 3}.Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Baseline}.Layout(gtx,
					layout.Rigid(func(gtx C) D { return t.etiquetaTamano(13, campo[0]+": ", t.Texto, font.SemiBold).Layout(gtx) }),
					layout.Flexed(1, func(gtx C) D { return t.etiqueta(13, campo[1], t.Texto, TipografiaCuerpo).Layout(gtx) }),
				)
			})
		}))
	}
	return fondoRedondeado(gtx, t.Fondo, 8, func(gtx C) D {
		return layout.Inset{Top: 8, Bottom: 10, Left: 12, Right: 12}.Layout(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, hijos...)
		})
	})
}

func detalleTratamiento(r modelo.Objeto) [][2]string {
	campos := [][2]string{}
	agregar := func(titulo, texto string) {
		if strings.TrimSpace(texto) != "" {
			campos = append(campos, [2]string{titulo, texto})
		}
	}

	fuentesTexto := []string{}
	if fuentes, ok := r.Get("fuentes").([]modelo.Fuente); ok {
		for _, f := range fuentes {
			linea := f.Origen + ": " + f.Referencia
			if iso, ok := f.FechaIso.(string); ok && iso != "" {
				linea += " (" + fechaBonita(iso) + ")"
			}
			fuentesTexto = append(fuentesTexto, linea)
		}
	}
	agregar("Fuentes", strings.Join(fuentesTexto, " · "))

	num, _ := r.Entero("numeroTratamientoPaciente")
	cant, _ := r.Entero("cantidadTratamientosPaciente")
	agregar("Tratamiento", strconv.Itoa(num)+" de "+strconv.Itoa(cant)+" de este paciente")

	if conflictos, ok := r.Get("conflictos").([]modelo.Conflicto); ok {
		for _, c := range conflictos {
			valores := []string{}
			for _, v := range c.Valores {
				valores = append(valores, v.Origen+": "+v.Valor)
			}
			agregar("No coincide — "+c.NombreCampo, strings.Join(valores, " · "))
		}
	}

	if motivo := textoOpcional(r, "motivoCategoria"); motivo != "" {
		agregar("Motivo", motivo+". "+textoOpcional(r, "textoMotivo"))
	}
	agregar("Diagnóstico en Excel", textoOpcional(r, "textoDiagnosticoExcel"))
	agregar("Ecografía", textoOpcional(r, "textoEco"))
	agregar("PAAF", textoOpcional(r, "textoPaaf"))
	agregar("Biopsia", textoOpcional(r, "textoBiopsia"))
	agregar("Estadificación", textoOpcional(r, "estadificacion"))
	agregar("Cintigrama", textoOpcional(r, "textoCintigrama"))

	laboratorio := []string{}
	for _, par := range [][2]string{{"TSH", "tsh"}, {"T4L", "t4l"}, {"T3", "t3"}, {"TRAb", "trab"}, {"Anti-Tg", "antiTiroglobulina"}, {"Anti-TPO", "antiTpo"}} {
		if m, ok := r.Get(par[1]).(modelo.Medicion); ok {
			laboratorio = append(laboratorio, par[0]+" "+textoMedicionVista(m))
		}
	}
	if vm := r.Tiroglobulina().ValorMaximo; vm != nil {
		laboratorio = append(laboratorio, "Tiroglobulina máx. "+strings.Replace(strconv.FormatFloat(vm.(float64), 'f', -1, 64), ".", ",", 1))
	}
	agregar("Laboratorio", strings.Join(laboratorio, " · "))

	agregar("Antecedentes familiares", textoOpcional(r, "textoAntecedentesFamiliares"))
	agregar("Mórbidos", textoOpcional(r, "textoMorbidos"))

	habitos := []string{}
	if c := textoOpcional(r, "textoCigarrillo"); c != "" {
		habitos = append(habitos, "Cigarrillo: "+c)
	}
	if oh := textoOpcional(r, "textoAlcohol"); oh != "" {
		habitos = append(habitos, "OH: "+oh)
	}
	if d := textoOpcional(r, "textoDrogas"); d != "" {
		habitos = append(habitos, "Drogas: "+d)
	}
	agregar("Hábitos", strings.Join(habitos, " · "))

	agregar("Médico tratante", textoOpcional(r, "medicoTratante"))

	sexo := textoOpcional(r, "sexo")
	origenSexo := textoOpcional(r, "origenSexo")
	if sexo != "" {
		agregar("Sexo", sexo+" (según "+strings.ToLower(origenSexo)+")")
	} else if origenSexo == "Por revisar" {
		agregar("Sexo", "Por revisar")
	} else {
		agregar("Sexo", "Sin dato")
	}

	if r.Bool("pacienteConCancerRegistrado") {
		agregar("Cáncer de tiroides", "El paciente figura en el registro de cáncer de tiroides (C73)")
	}
	if r.Bool("rechazoYodo") {
		agregar("Yodo", "El informe indica que la paciente rechazó o no recibió el yodo")
	}
	return campos
}
