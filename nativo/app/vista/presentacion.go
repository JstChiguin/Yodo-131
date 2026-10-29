package vista

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
)

type resalto struct {
	clic     widget.Clickable
	progreso float32
	ultimo   time.Time
}

const segundosTransicion = 0.18

func (v *Vista) resaltoDe(id string) *resalto {
	r := v.resaltos[id]
	if r == nil {
		r = &resalto{}
		v.resaltos[id] = r
	}
	return r
}

func (v *Vista) animarResalto(gtx C, r *resalto, activo bool) float32 {
	destino := float32(0)
	if activo {
		destino = 1
	}
	if v.resaltoInstantaneo {
		r.progreso = destino
		r.ultimo = gtx.Now
		return r.progreso
	}
	if r.ultimo.IsZero() {
		r.ultimo = gtx.Now
	}
	transcurrido := float32(gtx.Now.Sub(r.ultimo).Seconds())
	r.ultimo = gtx.Now
	if transcurrido < 0 {
		transcurrido = 0
	}
	paso := transcurrido / segundosTransicion
	if r.progreso < destino {
		r.progreso += paso
		if r.progreso >= destino {
			r.progreso = destino
		} else {
			gtx.Execute(op.InvalidateCmd{})
		}
	} else if r.progreso > destino {
		r.progreso -= paso
		if r.progreso <= destino {
			r.progreso = destino
		} else {
			gtx.Execute(op.InvalidateCmd{})
		}
	}
	return r.progreso
}

func (v *Vista) conResalto(gtx C, t *Tema, id string, presentacion bool, interior func(gtx C, resaltado float32) D) D {
	if !presentacion {
		return interior(gtx, 0)
	}
	r := v.resaltoDe(id)
	activo := r.clic.Hovered() || (v.resaltoInstantaneo && v.forzarResalto == id)
	progreso := v.animarResalto(gtx, r, activo)
	return r.clic.Layout(gtx, func(gtx C) D {
		if progreso <= 0 {
			return interior(gtx, 0)
		}
		subida := int(float32(gtx.Dp(3)) * progreso)
		defer op.Offset(image.Pt(0, -subida)).Push(gtx.Ops).Pop()
		macro := op.Record(gtx.Ops)
		dims := interior(gtx, progreso)
		contenido := macro.Stop()
		t.dibujarSombraResalto(gtx, dims.Size, progreso)
		t.dibujarAnilloResalto(gtx, dims.Size, progreso)
		contenido.Add(gtx.Ops)
		return dims
	})
}

func (t *Tema) dibujarSombraResalto(gtx C, tamano image.Point, progreso float32) {
	desplazamiento := int(float32(gtx.Dp(8)) * progreso)
	sombra := color.NRGBA{A: uint8(42 * progreso)}
	rect := image.Rect(0, desplazamiento, tamano.X, tamano.Y+desplazamiento)
	rellenarRect(gtx, rect, sombra, gtx.Dp(10))
}

func (t *Tema) dibujarAnilloResalto(gtx C, tamano image.Point, progreso float32) {
	grosor := gtx.Dp(2)
	acento := conAlfa(t.Acento, uint8(255*progreso))
	rect := image.Rect(-grosor, -grosor, tamano.X+grosor, tamano.Y+grosor)
	rellenarRect(gtx, rect, acento, gtx.Dp(12))
}

func (t *Tema) cifraColor(gtx C, etiqueta, valor, nota string, colorValor color.NRGBA) D {
	return widget.Border{Color: t.Borde, CornerRadius: 10, Width: 1}.Layout(gtx, func(gtx C) D {
		return fondoRedondeado(gtx, t.Panel, 10, func(gtx C) D {
			return layout.Inset{Top: 14, Bottom: 14, Left: 16, Right: 16}.Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D { return t.Pequeno(etiqueta, t.Suave).Layout(gtx) }),
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Top: 2, Bottom: 2}.Layout(gtx, func(gtx C) D {
							return t.etiqueta(30, valor, colorValor, TipografiaTitulo).Layout(gtx)
						})
					}),
					layout.Rigid(func(gtx C) D { return t.Pequeno(nota, t.Suave).Layout(gtx) }),
				)
			})
		})
	})
}
