package vista

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

func rellenarRect(gtx C, rect image.Rectangle, c color.NRGBA, radio int) {
	defer clip.UniformRRect(rect, radio).Push(gtx.Ops).Pop()
	paint.ColorOp{Color: c}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

func rellenarFondo(gtx C, c color.NRGBA) {
	rellenarRect(gtx, image.Rectangle{Max: gtx.Constraints.Max}, c, 0)
}

func fondoRedondeado(gtx C, c color.NRGBA, radioDp unit.Dp, w layout.Widget) D {
	macro := op.Record(gtx.Ops)
	dims := w(gtx)
	call := macro.Stop()
	rellenarRect(gtx, image.Rectangle{Max: dims.Size}, c, gtx.Dp(radioDp))
	call.Add(gtx.Ops)
	return dims
}

func conBarraIzquierda(gtx C, fondo, barra color.NRGBA, radioDp unit.Dp, anchoBarra unit.Dp, w layout.Widget) D {
	macro := op.Record(gtx.Ops)
	dims := w(gtx)
	call := macro.Stop()
	rellenarRect(gtx, image.Rectangle{Max: dims.Size}, fondo, gtx.Dp(radioDp))
	rellenarRect(gtx, image.Rectangle{Max: image.Pt(gtx.Dp(anchoBarra), dims.Size.Y)}, barra, 0)
	call.Add(gtx.Ops)
	return dims
}

func (t *Tema) tarjeta(gtx C, w layout.Widget) D {
	return widget.Border{Color: t.Borde, CornerRadius: 10, Width: 1}.Layout(gtx, func(gtx C) D {
		return fondoRedondeado(gtx, t.Panel, 10, func(gtx C) D {
			return layout.Inset{Top: 16, Bottom: 16, Left: 18, Right: 18}.Layout(gtx, w)
		})
	})
}

func (t *Tema) encabezadoTarjeta(gtx C, titulo, nota string) D {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D { return t.TituloTarjeta(titulo).Layout(gtx) }),
		layout.Rigid(func(gtx C) D {
			if nota == "" {
				return D{}
			}
			return layout.Inset{Top: 2}.Layout(gtx, func(gtx C) D { return t.Nota(nota).Layout(gtx) })
		}),
	)
}

func (t *Tema) cajaConclusion(gtx C, texto string, destacada bool) D {
	if texto == "" {
		return D{}
	}
	tam := unit.Sp(15)
	var anchoBarra unit.Dp = 3
	padV, padH := unit.Dp(9), unit.Dp(13)
	if destacada {
		tam = 20
		anchoBarra = 5
		padV, padH = 13, 16
	}
	return conBarraIzquierda(gtx, t.AcentoSuave, t.Acento, 8, anchoBarra, func(gtx C) D {
		return layout.Inset{Top: padV, Bottom: padV, Left: padH, Right: padH}.Layout(gtx, func(gtx C) D {
			lbl := t.etiqueta(tam, texto, t.Texto, TipografiaCuerpo)
			lbl.Font.Weight = font.Medium
			return lbl.Layout(gtx)
		})
	})
}

func (t *Tema) textoCobertura(gtx C, texto string) D {
	if texto == "" {
		return D{}
	}
	return layout.Inset{Left: 10}.Layout(gtx, func(gtx C) D {
		return conBarraIzquierda(gtx, color.NRGBA{}, t.BordeFuerte, 0, 2, func(gtx C) D {
			return layout.Inset{Left: 8}.Layout(gtx, func(gtx C) D {
				return t.Pequeno(texto, t.Suave).Layout(gtx)
			})
		})
	})
}

type leyendaItem struct {
	Nombre string
	Color  color.NRGBA
}

func (t *Tema) leyenda(gtx C, items []leyendaItem) D {
	hijos := make([]layout.FlexChild, 0, len(items))
	for _, item := range items {
		item := item
		hijos = append(hijos, layout.Rigid(func(gtx C) D {
			return layout.Inset{Right: 14, Top: 2, Bottom: 2}.Layout(gtx, func(gtx C) D {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						lado := gtx.Dp(10)
						rellenarRect(gtx, image.Rectangle{Max: image.Pt(lado, lado)}, item.Color, gtx.Dp(3))
						return D{Size: image.Pt(lado, lado)}
					}),
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Left: 6}.Layout(gtx, func(gtx C) D { return t.Pequeno(item.Nombre, t.Suave).Layout(gtx) })
					}),
				)
			})
		}))
	}
	return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceEnd}.Layout(gtx, envolver(hijos)...)
}

func envolver(hijos []layout.FlexChild) []layout.FlexChild {
	return hijos
}

func (t *Tema) badge(gtx C, texto string, fondo, frente color.NRGBA) D {
	return fondoRedondeado(gtx, fondo, 10, func(gtx C) D {
		return layout.Inset{Left: 6, Right: 6, Top: 1, Bottom: 1}.Layout(gtx, func(gtx C) D {
			lbl := t.etiqueta(12, texto, frente, TipografiaCuerpo)
			lbl.Alignment = 2
			return lbl.Layout(gtx)
		})
	})
}

const (
	botonNormal = iota
	botonPrincipal
	botonTexto
)

func (t *Tema) boton(gtx C, clic *widget.Clickable, texto string, tipo int, habilitado bool) D {
	if !habilitado {
		gtx = gtx.Disabled()
	}
	return material.Clickable(gtx, clic, func(gtx C) D {
		fondo := t.Panel
		borde := t.BordeFuerte
		colorTexto := t.Texto
		switch tipo {
		case botonPrincipal:
			fondo = t.Acento
			borde = t.Acento
			colorTexto = t.AcentoTexto
		case botonTexto:
			fondo = color.NRGBA{}
			borde = color.NRGBA{}
			colorTexto = t.Acento
		}
		return widget.Border{Color: borde, CornerRadius: 8, Width: 1}.Layout(gtx, func(gtx C) D {
			return fondoRedondeado(gtx, fondo, 8, func(gtx C) D {
				return layout.Inset{Left: 16, Right: 16, Top: 10, Bottom: 10}.Layout(gtx, func(gtx C) D {
					lbl := t.etiqueta(14, texto, colorTexto, TipografiaCuerpo)
					lbl.Font.Weight = font.SemiBold
					return lbl.Layout(gtx)
				})
			})
		})
	})
}

func rellenarRectLinea(gtx C, x0, y0, x1, y1 int, c color.NRGBA) {
	rellenarRect(gtx, image.Rect(x0, y0, x1, y1), c, 0)
}
