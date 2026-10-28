package vista

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"yodo131/nucleo/estadisticas"
	"yodo131/nucleo/firma"
	"yodo131/nucleo/modelo"
	"yodo131/nucleo/version"
)

type Datos struct {
	TieneDatos   bool
	Cargando     bool
	Mensaje      string
	NotaArchivos string
	Est          estadisticas.Estadisticas
	Resultado    modelo.Resultado
	Umbrales     map[string]float64
}

type Acciones struct {
	ElegirCarpeta        bool
	Exportar             bool
	EmpezarDeNuevo       bool
	CambiarTema          bool
	AlternarPresentacion bool
	IntentarClave        bool
	Clave                string
	GuardarUmbrales      bool
	UmbralesNuevos       map[string]float64
	RestablecerUmbrales  bool
}

type Vista struct {
	Oscuro       bool
	Presentacion bool
	Seccion      string
	ClaveValida  bool

	lista           widget.List
	botonCarpeta    widget.Clickable
	botonExcel      widget.Clickable
	botonEmpezar    widget.Clickable
	botonTema       widget.Clickable
	botonPresent    widget.Clickable
	botonBienvenida widget.Clickable

	pestanaResumen       widget.Clickable
	pestanaPacientes     widget.Clickable
	pestanaRevisar       widget.Clickable
	pestanaMantenimiento widget.Clickable

	pacientes     pacientesEstado
	mantenimiento mantenimientoEstado
}

func NuevaVista() *Vista {
	v := &Vista{Seccion: "resumen"}
	v.lista.Axis = layout.Vertical
	v.pacientes.inicializar()
	v.mantenimiento.inicializar()
	return v
}

func (v *Vista) Layout(gtx C, t *Tema, datos Datos) Acciones {
	var acciones Acciones
	if v.botonCarpeta.Clicked(gtx) {
		acciones.ElegirCarpeta = true
	}
	if v.botonBienvenida.Clicked(gtx) {
		acciones.ElegirCarpeta = true
	}
	if v.botonExcel.Clicked(gtx) {
		acciones.Exportar = true
	}
	if v.botonEmpezar.Clicked(gtx) {
		acciones.EmpezarDeNuevo = true
	}
	if v.botonTema.Clicked(gtx) {
		acciones.CambiarTema = true
	}
	if v.botonPresent.Clicked(gtx) {
		v.Presentacion = !v.Presentacion
		if v.Presentacion {
			v.Seccion = "resumen"
		}
	}
	if !v.Presentacion {
		if v.pestanaResumen.Clicked(gtx) {
			v.Seccion = "resumen"
		}
		if v.pestanaPacientes.Clicked(gtx) {
			v.Seccion = "pacientes"
		}
		if v.pestanaRevisar.Clicked(gtx) {
			v.Seccion = "revisar"
		}
		if v.pestanaMantenimiento.Clicked(gtx) {
			v.Seccion = "mantenimiento"
		}
	}

	rellenarFondo(gtx, t.Fondo)

	bloques := []layout.Widget{
		func(gtx C) D { return v.cabecera(gtx, t, datos) },
	}
	if datos.TieneDatos {
		bloques = append(bloques,
			func(gtx C) D { return v.pestanas(gtx, t, datos) },
			func(gtx C) D { return v.contenidoSeccion(gtx, t, datos, &acciones) },
		)
	} else {
		bloques = append(bloques, func(gtx C) D { return v.bienvenida(gtx, t, datos) })
	}

	v.conFirma(gtx, t, func(gtx C) D {
		return material.List(t.Th, &v.lista).Layout(gtx, len(bloques), func(gtx C, indice int) D {
			return layout.Inset{Top: 10, Bottom: 14}.Layout(gtx, func(gtx C) D {
				return t.centrado(gtx, bloques[indice])
			})
		})
	})
	return acciones
}

func (v *Vista) SetClaveValida(ok bool, mensaje string) {
	v.ClaveValida = ok
	v.mantenimiento.mensajeClave = mensaje
}

func (v *Vista) SetMensajeUmbrales(mensaje string) {
	v.mantenimiento.mensajeUmbrales = mensaje
}

func (v *Vista) PrefillUmbrales(umbrales map[string]float64) {
	v.mantenimiento.prefill(umbrales)
}

func (t *Tema) centrado(gtx C, w layout.Widget) D {
	gutter := gtx.Dp(16)
	maxAncho := gtx.Dp(1120)
	disponible := gtx.Constraints.Max.X
	contenido := disponible - 2*gutter
	if contenido > maxAncho {
		contenido = maxAncho
	}
	if contenido < 1 {
		contenido = disponible
	}
	lateral := (disponible - contenido) / 2
	defer op.Offset(image.Pt(lateral, 0)).Push(gtx.Ops).Pop()
	gtx.Constraints.Min.X = contenido
	gtx.Constraints.Max.X = contenido
	dims := w(gtx)
	return D{Size: image.Pt(disponible, dims.Size.Y)}
}

func (v *Vista) cabecera(gtx C, t *Tema, datos Datos) D {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.End, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Flexed(1, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					lbl := t.Pequeno("MEDICINA INTERNA · HOSPITAL DR. EDUARDO PEREIRA", t.Suave)
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Top: 2}.Layout(gtx, func(gtx C) D { return t.TituloPrincipal("Lector Yodo 131").Layout(gtx) })
				}),
			)
		}),
		layout.Rigid(func(gtx C) D { return v.accionesCabecera(gtx, t, datos) }),
	)
}

func (v *Vista) accionesCabecera(gtx C, t *Tema, datos Datos) D {
	hijos := []layout.FlexChild{}
	agregar := func(w layout.Widget) {
		hijos = append(hijos, layout.Rigid(func(gtx C) D {
			return layout.Inset{Left: 8}.Layout(gtx, w)
		}))
	}
	if !v.Presentacion {
		agregar(func(gtx C) D {
			etiqueta := "Elegir carpeta de informes"
			if datos.TieneDatos {
				etiqueta = "Agregar otra carpeta"
			}
			return t.boton(gtx, &v.botonCarpeta, etiqueta, botonPrincipal, true)
		})
		if datos.TieneDatos {
			agregar(func(gtx C) D { return t.boton(gtx, &v.botonEmpezar, "Empezar de nuevo", botonNormal, true) })
			agregar(func(gtx C) D { return t.boton(gtx, &v.botonExcel, "Descargar Excel", botonNormal, true) })
		}
	}
	if datos.TieneDatos {
		agregar(func(gtx C) D {
			etiqueta := "Modo presentación"
			tipo := botonNormal
			if v.Presentacion {
				etiqueta = "Salir de presentación"
				tipo = botonPrincipal
			}
			return t.boton(gtx, &v.botonPresent, etiqueta, tipo, true)
		})
	}
	agregar(func(gtx C) D {
		etiqueta := "Modo oscuro"
		if v.Oscuro {
			etiqueta = "Modo claro"
		}
		return t.boton(gtx, &v.botonTema, etiqueta, botonNormal, true)
	})
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, hijos...)
}

func (v *Vista) pestanas(gtx C, t *Tema, datos Datos) D {
	contador := len(datos.Resultado.Conflictos)
	for _, tr := range datos.Resultado.Tratamientos {
		if campos, ok := tr.Get("camposFaltantes").([]string); ok && len(campos) > 0 {
			contador++
		}
	}
	contador += len(datos.Resultado.DocumentosNoLeidos)

	tipo := []struct {
		clave  string
		texto  string
		clic   *widget.Clickable
		conteo int
	}{
		{"resumen", "Gráficos", &v.pestanaResumen, 0},
		{"pacientes", "Pacientes", &v.pestanaPacientes, 0},
		{"revisar", "Por revisar", &v.pestanaRevisar, contador},
		{"mantenimiento", "Mantenimiento", &v.pestanaMantenimiento, 0},
	}
	hijos := []layout.FlexChild{}
	for _, p := range tipo {
		p := p
		if v.Presentacion && p.clave != "resumen" {
			continue
		}
		hijos = append(hijos, layout.Rigid(func(gtx C) D { return v.pestana(gtx, t, p.texto, p.clave, p.clic, p.conteo) }))
	}
	return func(gtx C) D {
		dims := layout.Flex{Axis: layout.Horizontal}.Layout(gtx, hijos...)
		rellenarRect(gtx, image.Rect(0, dims.Size.Y-1, gtx.Constraints.Max.X, dims.Size.Y), t.Borde, 0)
		return dims
	}(gtx)
}

func (v *Vista) pestana(gtx C, t *Tema, texto, clave string, clic *widget.Clickable, conteo int) D {
	activa := v.Seccion == clave
	return material.Clickable(gtx, clic, func(gtx C) D {
		return layout.Inset{Left: 14, Right: 14, Top: 10, Bottom: 10}.Layout(gtx, func(gtx C) D {
			colorTexto := t.Suave
			if activa {
				colorTexto = t.Texto
			}
			dims := layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					lbl := t.etiqueta(14, texto, colorTexto, TipografiaCuerpo)
					lbl.Font.Weight = font.SemiBold
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					if conteo == 0 {
						return D{}
					}
					return layout.Inset{Left: 4}.Layout(gtx, func(gtx C) D {
						return t.badge(gtx, formatoNumero(conteo), t.Elevado, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff})
					})
				}),
			)
			if activa {
				rellenarRect(gtx, image.Rect(0, dims.Size.Y-3, dims.Size.X, dims.Size.Y), t.Acento, 0)
			}
			return dims
		})
	})
}

func (v *Vista) contenidoSeccion(gtx C, t *Tema, datos Datos, acc *Acciones) D {
	switch v.Seccion {
	case "pacientes":
		return v.layoutPacientes(gtx, t, datos)
	case "revisar":
		return v.layoutRevisar(gtx, t, datos)
	case "mantenimiento":
		return v.layoutMantenimiento(gtx, t, datos, acc)
	default:
		return v.layoutResumen(gtx, t, datos)
	}
}

func (v *Vista) layoutResumen(gtx C, t *Tema, datos Datos) D {
	tarjetas := t.construirTarjetasResumen(datos.Est, datos.Umbrales, v.Presentacion)
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			if v.Presentacion || datos.NotaArchivos == "" {
				return D{}
			}
			return layout.Inset{Bottom: 12}.Layout(gtx, func(gtx C) D {
				return widget.Border{Color: t.Borde, CornerRadius: 8, Width: 1}.Layout(gtx, func(gtx C) D {
					return fondoRedondeado(gtx, t.Panel, 8, func(gtx C) D {
						return layout.Inset{Top: 8, Bottom: 8, Left: 12, Right: 12}.Layout(gtx, func(gtx C) D { return t.Nota(datos.NotaArchivos).Layout(gtx) })
					})
				})
			})
		}),
		layout.Rigid(func(gtx C) D { return v.cifrasYLuego(gtx, t, datos) }),
		layout.Rigid(func(gtx C) D { return layout.Spacer{Height: 16}.Layout(gtx) }),
		layout.Rigid(func(gtx C) D { return t.layoutTarjetas(gtx, tarjetas) }),
	)
}

func (v *Vista) cifrasYLuego(gtx C, t *Tema, datos Datos) D {
	return t.cifrasResumen(gtx, datos.Est, datos.Resultado.Conteo, v.Presentacion)
}

func (v *Vista) bienvenida(gtx C, t *Tema, datos Datos) D {
	return widget.Border{Color: t.BordeFuerte, CornerRadius: 10, Width: 2}.Layout(gtx, func(gtx C) D {
		return fondoRedondeado(gtx, t.Panel, 10, func(gtx C) D {
			return layout.Inset{Top: 28, Bottom: 28, Left: 28, Right: 28}.Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return t.etiqueta(22, "Elija la carpeta con los informes Word y las planillas Excel", t.Texto, TipografiaTitulo).Layout(gtx)
					}),
					layout.Rigid(func(gtx C) D { return layout.Spacer{Height: 12}.Layout(gtx) }),
					layout.Rigid(func(gtx C) D { return t.Nota(datos.Mensaje).Layout(gtx) }),
					layout.Rigid(func(gtx C) D { return layout.Spacer{Height: 16}.Layout(gtx) }),
					layout.Rigid(func(gtx C) D {
						gtx.Constraints.Max.X = gtx.Dp(300)
						return t.boton(gtx, &v.botonBienvenida, "Elegir carpeta de informes", botonPrincipal, true)
					}),
				)
			})
		})
	})
}

func (v *Vista) conFirma(gtx C, t *Tema, w layout.Widget) D {
	gtx.Constraints.Min = gtx.Constraints.Max
	dims := w(gtx)

	gtxFirma := gtx
	gtxFirma.Constraints.Min = image.Point{}
	macro := op.Record(gtx.Ops)
	dimsFirma := v.firma(gtxFirma, t)
	grabado := macro.Stop()

	desplazamiento := image.Pt(
		gtx.Constraints.Max.X-dimsFirma.Size.X-gtx.Dp(14),
		gtx.Constraints.Max.Y-dimsFirma.Size.Y-gtx.Dp(10),
	)
	trans := op.Offset(desplazamiento).Push(gtx.Ops)
	grabado.Add(gtx.Ops)
	trans.Pop()
	return dims
}

func (v *Vista) firma(gtx C, t *Tema) D {
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return widget.Border{Color: t.Borde, CornerRadius: 20, Width: 1}.Layout(gtx, func(gtx C) D {
				return fondoRedondeado(gtx, conAlfa(t.Panel, 210), 20, func(gtx C) D {
					return layout.Inset{Left: 10, Right: 10, Top: 3, Bottom: 3}.Layout(gtx, func(gtx C) D {
						lbl := t.etiqueta(14, firma.Texto, conAlfa(t.Texto, 150), TipografiaTitulo)
						lbl.Font.Style = font.Italic
						return lbl.Layout(gtx)
					})
				})
			})
		}),
		layout.Rigid(func(gtx C) D { return layout.Spacer{Height: 3}.Layout(gtx) }),
		layout.Rigid(func(gtx C) D {
			return widget.Border{Color: t.Borde, CornerRadius: 20, Width: 1}.Layout(gtx, func(gtx C) D {
				return fondoRedondeado(gtx, conAlfa(t.Panel, 190), 20, func(gtx C) D {
					return layout.Inset{Left: 8, Right: 8, Top: 1, Bottom: 1}.Layout(gtx, func(gtx C) D {
						return t.etiqueta(11, "v"+version.Aplicacion, conAlfa(t.Suave, 190), TipografiaCuerpo).Layout(gtx)
					})
				})
			})
		}),
	)
}

var _ = unit.Dp(0)
