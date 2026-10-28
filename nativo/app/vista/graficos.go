package vista

import (
	"image"
	"image/color"
	"math"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"yodo131/nucleo/estadisticas"
)

type filaBarra struct {
	Categoria   string
	Cantidad    int
	Descripcion string
	Color       color.NRGBA
}

type serieApilada struct {
	Clave      string
	Nombre     string
	Color      color.NRGBA
	ColorTexto color.NRGBA
}

type tramoProp struct {
	Categoria string
	Cantidad  int
	Color     color.NRGBA
}

type grupoProp struct {
	Nombre  string
	Resumen string
	Tramos  []tramoProp
	Total   int
	Detalle string
}

func anchoAprox(texto string, tamano float64) int {
	return int(float64(len([]rune(texto))) * tamano * 0.55)
}

func (t *Tema) graficoVacio(gtx C, mensaje string) D {
	return layout.Inset{Top: 12, Bottom: 12}.Layout(gtx, func(gtx C) D {
		return t.Pequeno(mensaje, t.Suave).Layout(gtx)
	})
}

func (t *Tema) barrasHorizontales(gtx C, filas []filaBarra, total int) D {
	if len(filas) == 0 {
		return t.graficoVacio(gtx, "No hay datos en este período.")
	}
	maximo := 1
	for _, fila := range filas {
		if fila.Cantidad > maximo {
			maximo = fila.Cantidad
		}
	}
	hijos := make([]layout.FlexChild, 0, len(filas))
	for _, fila := range filas {
		fila := fila
		hijos = append(hijos, layout.Rigid(func(gtx C) D {
			return layout.Inset{Top: 5, Bottom: 5}.Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D { return t.tituloBarra(gtx, fila.Categoria, fila.Descripcion) }),
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Top: 3}.Layout(gtx, func(gtx C) D {
							return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
								layout.Flexed(1, func(gtx C) D { return t.pistaBarra(gtx, fila.Cantidad, maximo, fila.Color) }),
								layout.Rigid(func(gtx C) D {
									return layout.Inset{Left: 10}.Layout(gtx, func(gtx C) D { return t.valorBarra(gtx, fila.Cantidad, total) })
								}),
							)
						})
					}),
				)
			})
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, hijos...)
}

func (t *Tema) tituloBarra(gtx C, categoria, descripcion string) D {
	nombre := t.etiqueta(13, categoria, t.Texto, TipografiaCuerpo)
	if descripcion == "" {
		return nombre.Layout(gtx)
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Baseline}.Layout(gtx,
		layout.Rigid(func(gtx C) D { return nombre.Layout(gtx) }),
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Left: 8}.Layout(gtx, func(gtx C) D { return t.Pequeno(descripcion, t.Suave).Layout(gtx) })
		}),
	)
}

func (t *Tema) pistaBarra(gtx C, cantidad, maximo int, c color.NRGBA) D {
	ancho := gtx.Constraints.Max.X
	alto := gtx.Dp(14)
	rellenarRect(gtx, image.Rectangle{Max: image.Pt(ancho, alto)}, t.Rejilla, gtx.Dp(5))
	fraccionBarra := float64(cantidad) / float64(maximo)
	anchoRelleno := int(fraccionBarra * float64(ancho))
	minimo := int(0.015 * float64(ancho))
	if anchoRelleno < minimo {
		anchoRelleno = minimo
	}
	if anchoRelleno > ancho {
		anchoRelleno = ancho
	}
	if anchoRelleno < gtx.Dp(3) {
		anchoRelleno = gtx.Dp(3)
	}
	rellenarRect(gtx, image.Rectangle{Max: image.Pt(anchoRelleno, alto)}, c, gtx.Dp(5))
	return D{Size: image.Pt(ancho, alto)}
}

func (t *Tema) valorBarra(gtx C, cantidad, total int) D {
	gtx.Constraints.Min.X = gtx.Dp(72)
	return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceStart}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			lbl := t.etiqueta(13, formatoNumero(cantidad), t.Texto, TipografiaCuerpo)
			lbl.Font.Weight = font.Bold
			return lbl.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return t.etiqueta(13, " · "+porcentajeTexto(cantidad, total), t.Suave, TipografiaCuerpo).Layout(gtx)
		}),
	)
}

func calcularMarcasEje(maximo int) []int {
	pasoBruto := float64(maximo) / 4
	base := pasoBruto
	if base < 1 {
		base = 1
	}
	magnitud := math.Pow(10, math.Floor(math.Log10(base)))
	paso := magnitud * 10
	for _, factor := range []float64{1, 2, 5, 10} {
		candidato := factor * magnitud
		if candidato >= pasoBruto {
			paso = candidato
			break
		}
	}
	pasoEntero := int(math.Round(paso))
	if pasoEntero < 1 {
		pasoEntero = 1
	}
	marcas := []int{}
	for valor := 0; valor <= maximo+pasoEntero-1; valor += pasoEntero {
		marcas = append(marcas, valor)
	}
	return marcas
}

func valorSerie(fila estadisticas.FilaAnio, clave string) int {
	switch clave {
	case "mujeres":
		return fila.Mujeres
	case "hombres":
		return fila.Hombres
	default:
		return fila.SinDato
	}
}

func (t *Tema) textoEn(gtx C, izquierda, arriba, ancho, alto int, lbl material.LabelStyle, align text.Alignment) {
	defer op.Offset(image.Pt(izquierda, arriba)).Push(gtx.Ops).Pop()
	lbl.Alignment = align
	gtx.Constraints = layout.Exact(image.Pt(ancho, alto))
	lbl.Layout(gtx)
}

func (t *Tema) etiquetaTamano(tam unit.Sp, texto string, c color.NRGBA, peso font.Weight) material.LabelStyle {
	lbl := t.etiqueta(tam, texto, c, TipografiaCuerpo)
	lbl.Font.Weight = peso
	return lbl
}

func (t *Tema) barrasApiladas(gtx C, filas []estadisticas.FilaAnio, series []serieApilada) D {
	if len(filas) == 0 {
		return t.graficoVacio(gtx, "No hay datos en este período.")
	}
	ancho := gtx.Constraints.Max.X
	if ancho < gtx.Dp(280) {
		ancho = gtx.Dp(280)
	}
	alto := gtx.Dp(280)

	totales := make([]int, len(filas))
	maximoTotal := 1
	for i, fila := range filas {
		suma := 0
		for _, serie := range series {
			suma += valorSerie(fila, serie.Clave)
		}
		totales[i] = suma
		if suma > maximoTotal {
			maximoTotal = suma
		}
	}
	marcas := calcularMarcasEje(maximoTotal)
	maximoEje := marcas[len(marcas)-1]
	if maximoEje < 1 {
		maximoEje = 1
	}
	anchoMarca := 0
	for _, marca := range marcas {
		if a := anchoAprox(formatoNumero(marca), 12); a > anchoMarca {
			anchoMarca = a
		}
	}
	margenArriba := gtx.Dp(24)
	margenAbajo := gtx.Dp(30)
	margenIzq := anchoMarca + gtx.Dp(14)
	margenDer := gtx.Dp(8)
	altoUtil := alto - margenArriba - margenAbajo
	escalaY := func(valor int) int { return int(float64(valor) / float64(maximoEje) * float64(altoUtil)) }
	anchoBanda := float64(ancho-margenIzq-margenDer) / float64(len(filas))
	anchoBarra := anchoBanda * 0.6
	if anchoBarra > float64(gtx.Dp(56)) {
		anchoBarra = float64(gtx.Dp(56))
	}
	anchoEtiqueta := 0
	for _, fila := range filas {
		if a := anchoAprox(fila.Categoria, 12); a > anchoEtiqueta {
			anchoEtiqueta = a
		}
	}
	saltoEtiquetas := int(math.Ceil(float64(anchoEtiqueta+gtx.Dp(8)) / anchoBanda))
	if saltoEtiquetas < 1 {
		saltoEtiquetas = 1
	}

	for _, marca := range marcas {
		y := alto - margenAbajo - escalaY(marca)
		colorLinea := t.Rejilla
		if marca == 0 {
			colorLinea = t.BordeFuerte
		}
		rellenarRect(gtx, image.Rect(margenIzq, y, ancho-margenDer, y+1), colorLinea, 0)
		t.textoEn(gtx, 0, y-gtx.Dp(8), margenIzq-gtx.Dp(8), gtx.Dp(16), t.Pequeno(formatoNumero(marca), t.Suave), text.End)
	}

	for indice, fila := range filas {
		x := margenIzq + int(float64(indice)*anchoBanda+(anchoBanda-anchoBarra)/2)
		barra := int(anchoBarra)
		base := alto - margenAbajo
		for _, serie := range series {
			cantidad := valorSerie(fila, serie.Clave)
			if cantidad == 0 {
				continue
			}
			altoTramo := escalaY(cantidad)
			rellenarRect(gtx, image.Rect(x, base-altoTramo, x+barra, base), serie.Color, 0)
			if altoTramo >= gtx.Dp(18) && barra >= gtx.Dp(20) {
				colorTexto := serie.ColorTexto
				if colorTexto == (color.NRGBA{}) {
					colorTexto = t.Panel
				}
				t.textoEn(gtx, x, base-altoTramo/2-gtx.Dp(8), barra, gtx.Dp(16), t.etiquetaTamano(12, formatoNumero(cantidad), colorTexto, font.SemiBold), text.Middle)
			}
			base -= altoTramo
		}
		t.textoEn(gtx, x-gtx.Dp(10), base-gtx.Dp(18), barra+gtx.Dp(20), gtx.Dp(16), t.etiquetaTamano(13, formatoNumero(totales[indice]), t.Texto, font.Bold), text.Middle)
		if indice%saltoEtiquetas == 0 {
			t.textoEn(gtx, x-gtx.Dp(20), alto-margenAbajo+gtx.Dp(6), barra+gtx.Dp(40), gtx.Dp(16), t.etiquetaTamano(12, fila.Categoria, t.Suave, font.SemiBold), text.Middle)
		}
	}
	return D{Size: image.Pt(ancho, alto)}
}

func (t *Tema) proporciones(gtx C, grupos []grupoProp) D {
	hijos := make([]layout.FlexChild, 0, len(grupos))
	for indice, grupo := range grupos {
		grupo := grupo
		primero := indice == 0
		hijos = append(hijos, layout.Rigid(func(gtx C) D {
			return layout.Inset{Top: 6, Bottom: 6}.Layout(gtx, func(gtx C) D {
				if !primero {
					rellenarRect(gtx, image.Rect(0, 0, gtx.Constraints.Max.X, 1), t.Rejilla, 0)
				}
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Baseline, Spacing: layout.SpaceBetween}.Layout(gtx,
							layout.Rigid(func(gtx C) D { return t.etiquetaTamano(14, grupo.Nombre, t.Texto, font.SemiBold).Layout(gtx) }),
							layout.Rigid(func(gtx C) D { return t.etiqueta(13, grupo.Resumen, t.Texto, TipografiaCuerpo).Layout(gtx) }),
						)
					}),
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Top: 5, Bottom: 5}.Layout(gtx, func(gtx C) D { return t.barraProporcion(gtx, grupo) })
					}),
					layout.Rigid(func(gtx C) D { return t.Pequeno(grupo.Detalle, t.Suave).Layout(gtx) }),
				)
			})
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, hijos...)
}

func (t *Tema) barraProporcion(gtx C, grupo grupoProp) D {
	ancho := gtx.Constraints.Max.X
	alto := gtx.Dp(24)
	radio := gtx.Dp(6)
	rellenarRect(gtx, image.Rectangle{Max: image.Pt(ancho, alto)}, t.Rejilla, radio)
	if grupo.Total == 0 {
		t.textoEn(gtx, 0, gtx.Dp(4), ancho, gtx.Dp(16), t.Pequeno("Sin datos", t.Suave), text.Middle)
		return D{Size: image.Pt(ancho, alto)}
	}
	x := 0
	for _, tramo := range grupo.Tramos {
		if tramo.Cantidad == 0 {
			continue
		}
		anchoTramo := int(float64(tramo.Cantidad) / float64(grupo.Total) * float64(ancho))
		rellenarRect(gtx, image.Rect(x, 0, x+anchoTramo, alto), tramo.Color, 0)
		if float64(tramo.Cantidad)/float64(grupo.Total) >= 0.1 {
			t.textoEn(gtx, x, gtx.Dp(4), anchoTramo, gtx.Dp(16), t.etiquetaTamano(12, porcentajeTexto(tramo.Cantidad, grupo.Total), t.Panel, font.Bold), text.Middle)
		}
		x += anchoTramo
	}
	return D{Size: image.Pt(ancho, alto)}
}
