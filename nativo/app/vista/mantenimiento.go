package vista

import (
	"strconv"
	"strings"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type mantenimientoEstado struct {
	campoClave       widget.Editor
	botonEntrar      widget.Clickable
	editores         map[string]*widget.Editor
	orden            []string
	etiquetas        map[string]string
	botonGuardar     widget.Clickable
	botonRestablecer widget.Clickable
	mensajeClave     string
	mensajeUmbrales  string
	prefilled        bool
}

func (m *mantenimientoEstado) inicializar() {
	m.campoClave.SingleLine = true
	m.campoClave.Mask = '•'
	m.orden = []string{"trab", "tiroglobulina", "antiTiroglobulina", "antiTpo"}
	m.etiquetas = map[string]string{
		"trab": "TRAb (sobre)", "tiroglobulina": "Tiroglobulina ng/ml (sobre)",
		"antiTiroglobulina": "Anti-tiroglobulina UI/ml (sobre)", "antiTpo": "Anti-TPO UI/ml (sobre)",
	}
	m.editores = map[string]*widget.Editor{}
	for _, clave := range m.orden {
		ed := &widget.Editor{SingleLine: true}
		m.editores[clave] = ed
	}
}

func (m *mantenimientoEstado) prefill(umbrales map[string]float64) {
	for _, clave := range m.orden {
		m.editores[clave].SetText(strconv.FormatFloat(umbrales[clave], 'g', -1, 64))
	}
	m.prefilled = true
}

func (v *Vista) layoutMantenimiento(gtx C, t *Tema, datos Datos, acc *Acciones) D {
	m := &v.mantenimiento
	if !v.ClaveValida {
		if v.mantenimiento.botonEntrar.Clicked(gtx) {
			acc.IntentarClave = true
			acc.Clave = m.campoClave.Text()
		}
		return v.tarjetaClave(gtx, t)
	}
	if !m.prefilled {
		m.prefill(datos.Umbrales)
	}
	if m.botonGuardar.Clicked(gtx) {
		acc.GuardarUmbrales = true
		acc.UmbralesNuevos = map[string]float64{}
		for _, clave := range m.orden {
			valor, _ := strconv.ParseFloat(strings.Replace(strings.TrimSpace(m.editores[clave].Text()), ",", ".", 1), 64)
			acc.UmbralesNuevos[clave] = valor
		}
	}
	if m.botonRestablecer.Clicked(gtx) {
		acc.RestablecerUmbrales = true
		m.prefilled = false
	}
	return v.tarjetaUmbrales(gtx, t)
}

func (v *Vista) tarjetaClave(gtx C, t *Tema) D {
	gtx.Constraints.Max.X = gtx.Dp(540)
	return t.tarjeta(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D { return t.TituloTarjeta("Mantenimiento").Layout(gtx) }),
			layout.Rigid(func(gtx C) D {
				return layout.Inset{Top: 2, Bottom: 10}.Layout(gtx, func(gtx C) D {
					return t.Nota("Esta sección cambia cómo se interpretan los exámenes. Requiere clave.").Layout(gtx)
				})
			}),
			layout.Rigid(func(gtx C) D { return v.campoEtiquetado(gtx, t, "Clave", &v.mantenimiento.campoClave, "") }),
			layout.Rigid(func(gtx C) D { return layout.Spacer{Height: 10}.Layout(gtx) }),
			layout.Rigid(func(gtx C) D {
				gtx.Constraints.Max.X = gtx.Dp(140)
				return t.boton(gtx, &v.mantenimiento.botonEntrar, "Entrar", botonPrincipal, true)
			}),
			layout.Rigid(func(gtx C) D {
				if v.mantenimiento.mensajeClave == "" {
					return D{}
				}
				return layout.Inset{Top: 8}.Layout(gtx, func(gtx C) D {
					return t.etiqueta(13, v.mantenimiento.mensajeClave, t.Error, TipografiaCuerpo).Layout(gtx)
				})
			}),
		)
	})
}

func (v *Vista) tarjetaUmbrales(gtx C, t *Tema) D {
	m := &v.mantenimiento
	return t.tarjeta(gtx, func(gtx C) D {
		hijos := []layout.FlexChild{
			layout.Rigid(func(gtx C) D {
				return t.TituloTarjeta("Valores sobre los que un examen se considera elevado").Layout(gtx)
			}),
			layout.Rigid(func(gtx C) D { return layout.Spacer{Height: 10}.Layout(gtx) }),
		}
		for _, clave := range m.orden {
			clave := clave
			hijos = append(hijos, layout.Rigid(func(gtx C) D {
				return layout.Inset{Bottom: 10}.Layout(gtx, func(gtx C) D {
					gtx.Constraints.Max.X = gtx.Dp(360)
					return v.campoEtiquetado(gtx, t, m.etiquetas[clave], m.editores[clave], "")
				})
			}))
		}
		hijos = append(hijos, layout.Rigid(func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					gtx.Constraints.Max.X = gtx.Dp(170)
					return t.boton(gtx, &m.botonGuardar, "Guardar valores", botonPrincipal, true)
				}),
				layout.Rigid(func(gtx C) D { return layout.Spacer{Width: 8}.Layout(gtx) }),
				layout.Rigid(func(gtx C) D {
					gtx.Constraints.Max.X = gtx.Dp(260)
					return t.boton(gtx, &m.botonRestablecer, "Volver a los valores iniciales", botonNormal, true)
				}),
			)
		}))
		hijos = append(hijos, layout.Rigid(func(gtx C) D {
			if m.mensajeUmbrales == "" {
				return D{}
			}
			return layout.Inset{Top: 8}.Layout(gtx, func(gtx C) D { return t.Nota(m.mensajeUmbrales).Layout(gtx) })
		}))
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, hijos...)
	})
}

func (v *Vista) campoEtiquetado(gtx C, t *Tema, etiqueta string, editor *widget.Editor, hint string) D {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D { return t.Pequeno(etiqueta, t.Suave).Layout(gtx) }),
		layout.Rigid(func(gtx C) D { return layout.Spacer{Height: 4}.Layout(gtx) }),
		layout.Rigid(func(gtx C) D {
			return widget.Border{Color: t.BordeFuerte, CornerRadius: 8, Width: 1}.Layout(gtx, func(gtx C) D {
				return fondoRedondeado(gtx, t.Panel, 8, func(gtx C) D {
					return layout.Inset{Left: 10, Right: 10, Top: 8, Bottom: 8}.Layout(gtx, func(gtx C) D {
						ed := material.Editor(t.Th, editor, hint)
						ed.Color = t.Texto
						ed.HintColor = t.Suave
						ed.Font.Typeface = TipografiaCuerpo
						return ed.Layout(gtx)
					})
				})
			})
		}),
	)
}
