package vista

import (
	"strconv"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

var mesesAbreviados = []string{"ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"}

func partesMes(mes string) (int, int, bool) {
	if len(mes) != 7 || mes[4] != '-' {
		return 0, 0, false
	}
	anio, err := strconv.Atoi(mes[:4])
	if err != nil {
		return 0, 0, false
	}
	numero, err := strconv.Atoi(mes[5:])
	if err != nil || numero < 1 || numero > 12 {
		return 0, 0, false
	}
	return anio, numero, true
}

func componerMes(anio, numero int) string {
	return strconv.Itoa(anio) + "-" + leadingCero(numero)
}

func leadingCero(numero int) string {
	if numero < 10 {
		return "0" + strconv.Itoa(numero)
	}
	return strconv.Itoa(numero)
}

func mesAnterior(mes string) string {
	anio, numero, ok := partesMes(mes)
	if !ok {
		return mes
	}
	if numero == 1 {
		return componerMes(anio-1, 12)
	}
	return componerMes(anio, numero-1)
}

func mesSiguiente(mes string) string {
	anio, numero, ok := partesMes(mes)
	if !ok {
		return mes
	}
	if numero == 12 {
		return componerMes(anio+1, 1)
	}
	return componerMes(anio, numero+1)
}

func formatoMes(mes string) string {
	anio, numero, ok := partesMes(mes)
	if !ok {
		return mes
	}
	return mesesAbreviados[numero-1] + " " + strconv.Itoa(anio)
}

func (v *Vista) desdeActual(datos Datos) string {
	if v.filtroDesde != "" {
		return v.filtroDesde
	}
	return datos.PeriodoMin
}

func (v *Vista) hastaActual(datos Datos) string {
	if v.filtroHasta != "" {
		return v.filtroHasta
	}
	return datos.PeriodoMax
}

func (v *Vista) describirPeriodo() string {
	desde, hasta := v.filtroDesde, v.filtroHasta
	switch {
	case desde == "" && hasta == "":
		return "Todos los tratamientos"
	case desde != "" && hasta != "":
		return "De " + formatoMes(desde) + " a " + formatoMes(hasta)
	case desde != "":
		return "Desde " + formatoMes(desde)
	default:
		return "Hasta " + formatoMes(hasta)
	}
}

func (v *Vista) notaPeriodo(datos Datos) string {
	texto := v.describirPeriodo()
	if (v.filtroDesde != "" || v.filtroHasta != "") && datos.CantidadSinFecha > 0 {
		texto += " · " + formatoNumero(datos.CantidadSinFecha) + " tratamiento(s) sin fecha quedan fuera"
	}
	return texto
}

func (v *Vista) procesarPeriodo(gtx C, datos Datos, acciones *Acciones) {
	if datos.PeriodoMin == "" || datos.PeriodoMax == "" {
		return
	}
	desde := v.desdeActual(datos)
	hasta := v.hastaActual(datos)
	cambio := false
	if v.periodoDesdeMenos.Clicked(gtx) {
		nuevo := mesAnterior(desde)
		if nuevo >= datos.PeriodoMin {
			v.filtroDesde = nuevo
			cambio = true
		}
	}
	if v.periodoDesdeMas.Clicked(gtx) {
		nuevo := mesSiguiente(desde)
		if nuevo <= hasta {
			v.filtroDesde = nuevo
			cambio = true
		}
	}
	if v.periodoHastaMenos.Clicked(gtx) {
		nuevo := mesAnterior(hasta)
		if nuevo >= desde {
			v.filtroHasta = nuevo
			cambio = true
		}
	}
	if v.periodoHastaMas.Clicked(gtx) {
		nuevo := mesSiguiente(hasta)
		if nuevo <= datos.PeriodoMax {
			v.filtroHasta = nuevo
			cambio = true
		}
	}
	if v.periodoTodo.Clicked(gtx) && (v.filtroDesde != "" || v.filtroHasta != "") {
		v.filtroDesde = ""
		v.filtroHasta = ""
		cambio = true
	}
	if cambio {
		acciones.CambiarPeriodo = true
		acciones.PeriodoDesde = v.filtroDesde
		acciones.PeriodoHasta = v.filtroHasta
	}
}

func (v *Vista) barraFiltroPeriodo(gtx C, t *Tema, datos Datos) D {
	if datos.PeriodoMin == "" || datos.PeriodoMax == "" {
		return D{}
	}
	desde := v.desdeActual(datos)
	hasta := v.hastaActual(datos)
	hayFiltro := v.filtroDesde != "" || v.filtroHasta != ""
	return layout.Inset{Bottom: 12}.Layout(gtx, func(gtx C) D {
		return widget.Border{Color: t.Borde, CornerRadius: 8, Width: 1}.Layout(gtx, func(gtx C) D {
			return fondoRedondeado(gtx, t.Panel, 8, func(gtx C) D {
				return layout.Inset{Top: 10, Bottom: 10, Left: 14, Right: 14}.Layout(gtx, func(gtx C) D {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									lbl := t.etiqueta(13, "Período", t.Texto, TipografiaCuerpo)
									lbl.Font.Weight = font.SemiBold
									return layout.Inset{Right: 14}.Layout(gtx, func(gtx C) D { return lbl.Layout(gtx) })
								}),
								layout.Rigid(func(gtx C) D {
									return v.selectorMes(gtx, t, "Desde", desde, &v.periodoDesdeMenos, &v.periodoDesdeMas, desde > datos.PeriodoMin, desde < hasta)
								}),
								layout.Rigid(func(gtx C) D { return layout.Spacer{Width: 18}.Layout(gtx) }),
								layout.Rigid(func(gtx C) D {
									return v.selectorMes(gtx, t, "Hasta", hasta, &v.periodoHastaMenos, &v.periodoHastaMas, hasta > desde, hasta < datos.PeriodoMax)
								}),
								layout.Rigid(func(gtx C) D { return layout.Spacer{Width: 18}.Layout(gtx) }),
								layout.Rigid(func(gtx C) D { return t.boton(gtx, &v.periodoTodo, "Ver todo", botonTexto, hayFiltro) }),
							)
						}),
						layout.Rigid(func(gtx C) D {
							return layout.Inset{Top: 8}.Layout(gtx, func(gtx C) D { return t.Nota(v.notaPeriodo(datos)).Layout(gtx) })
						}),
					)
				})
			})
		})
	})
}

func (v *Vista) selectorMes(gtx C, t *Tema, etiqueta, mes string, menos, mas *widget.Clickable, puedeMenos, puedeMas bool) D {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Right: 8}.Layout(gtx, func(gtx C) D { return t.Pequeno(etiqueta, t.Suave).Layout(gtx) })
		}),
		layout.Rigid(func(gtx C) D { return t.botonPaso(gtx, menos, "‹", puedeMenos) }),
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Left: 8, Right: 8}.Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Dp(78)
				gtx.Constraints.Max.X = gtx.Dp(78)
				lbl := t.etiqueta(13, formatoMes(mes), t.Texto, TipografiaCuerpo)
				lbl.Font.Weight = font.SemiBold
				lbl.Alignment = 2
				return lbl.Layout(gtx)
			})
		}),
		layout.Rigid(func(gtx C) D { return t.botonPaso(gtx, mas, "›", puedeMas) }),
	)
}

func (t *Tema) botonPaso(gtx C, clic *widget.Clickable, simbolo string, habilitado bool) D {
	if !habilitado {
		gtx = gtx.Disabled()
	}
	return material.Clickable(gtx, clic, func(gtx C) D {
		borde := t.BordeFuerte
		colorTexto := t.Texto
		if !habilitado {
			borde = t.Borde
			colorTexto = t.Suave
		}
		return widget.Border{Color: borde, CornerRadius: 6, Width: 1}.Layout(gtx, func(gtx C) D {
			return fondoRedondeado(gtx, t.Panel, 6, func(gtx C) D {
				return layout.Inset{Top: 5, Bottom: 5, Left: 11, Right: 11}.Layout(gtx, func(gtx C) D {
					lbl := t.etiqueta(16, simbolo, colorTexto, TipografiaCuerpo)
					lbl.Font.Weight = font.Bold
					return lbl.Layout(gtx)
				})
			})
		})
	})
}
