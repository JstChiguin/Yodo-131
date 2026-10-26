package main

import (
	"image"
	"image/color"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"yodo131/nucleo/dialogos"
	"yodo131/nucleo/estadisticas"
	"yodo131/nucleo/exportacion"
	"yodo131/nucleo/firma"
	"yodo131/nucleo/modelo"
	"yodo131/nucleo/procesamiento"
	"yodo131/nucleo/resumen"
	"yodo131/nucleo/version"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

var (
	colorFondo  = color.NRGBA{R: 0xf3, G: 0xf6, B: 0xf6, A: 0xff}
	colorAcento = color.NRGBA{R: 0x0f, G: 0x6b, B: 0x63, A: 0xff}
	colorTexto  = color.NRGBA{R: 0x15, G: 0x21, B: 0x1f, A: 0xff}
	colorSuave  = color.NRGBA{R: 0x55, G: 0x62, B: 0x5f, A: 0xff}
	colorFirma  = color.NRGBA{R: 0x55, G: 0x62, B: 0x5f, A: 0xcc}
)

type estado struct {
	mu            sync.Mutex
	botonCarpeta  widget.Clickable
	botonExcel    widget.Clickable
	lista         widget.List
	lineas        []string
	mensaje       string
	tieneDatos    bool
	cargando      bool
	est           estadisticas.Estadisticas
	resultado     modelo.Resultado
	carpetaActual string
	ventana       *app.Window
}

func nuevoEstado() *estado {
	st := &estado{mensaje: "Elija la carpeta con los informes Word y las planillas Excel."}
	st.lista.Axis = layout.Vertical
	return st
}

func (st *estado) procesarCarpeta(carpeta string) {
	st.mu.Lock()
	st.cargando = true
	st.mensaje = "Leyendo archivos…"
	st.mu.Unlock()
	st.invalidar()

	mensaje := ""
	var est estadisticas.Estadisticas
	var resultado modelo.Resultado
	var lineas []string
	tieneDatos := false
	archivos, err := procesamiento.LeerCarpeta(carpeta)
	if err != nil {
		mensaje = "No se pudo leer la carpeta."
	} else if len(archivos) == 0 {
		mensaje = "En esa carpeta no hay archivos Word (.docx) ni Excel (.xlsx)."
	} else {
		lectura := procesamiento.ProcesarArchivos(archivos, nil)
		resultado = procesamiento.ConstruirResultado(lectura, clasificacionUmbrales())
		est = estadisticas.Calcular(resultado, estadisticas.Filtro{})
		lineas = resumen.Lineas(est)
		tieneDatos = len(resultado.Tratamientos) > 0 || len(resultado.DocumentosNoLeidos) > 0
		mensaje = "Datos cargados desde el pendrive. Los archivos solo se leyeron; no se modificó ninguno."
	}

	st.mu.Lock()
	st.cargando = false
	st.carpetaActual = carpeta
	st.est = est
	st.resultado = resultado
	st.lineas = lineas
	st.tieneDatos = tieneDatos
	st.mensaje = mensaje
	st.mu.Unlock()
	st.invalidar()
}

func (st *estado) exportar(destino string) {
	st.mu.Lock()
	est := st.est
	resultado := st.resultado
	st.mu.Unlock()
	datos := exportacion.ExportarLibro(est, resultado, "Todos los tratamientos")
	mensaje := "Excel guardado donde usted eligió. Los datos no salieron del computador."
	if err := escribirArchivo(destino, datos); err != nil {
		mensaje = "No se pudo guardar el Excel en esa ubicación."
	}
	st.mu.Lock()
	st.mensaje = mensaje
	st.mu.Unlock()
	st.invalidar()
}

func (st *estado) invalidar() {
	st.mu.Lock()
	ventana := st.ventana
	st.mu.Unlock()
	if ventana != nil {
		ventana.Invalidate()
	}
}

func tema() *material.Theme {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	th.Palette.Fg = colorTexto
	return th
}

func (st *estado) dibujar(gtx C, th *material.Theme) D {
	if st.botonCarpeta.Clicked(gtx) {
		go st.alElegirCarpeta()
	}
	if st.botonExcel.Clicked(gtx) {
		go st.alExportar()
	}

	rellenar(gtx, colorFondo)

	st.mu.Lock()
	lineas := append([]string{}, st.lineas...)
	mensaje := st.mensaje
	tieneDatos := st.tieneDatos
	cargando := st.cargando
	st.mu.Unlock()

	return layout.Stack{Alignment: layout.SE}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			return layout.UniformInset(unit.Dp(24)).Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						titulo := material.H4(th, "Lector Yodo 131")
						titulo.Color = colorTexto
						return titulo.Layout(gtx)
					}),
					layout.Rigid(func(gtx C) D {
						sub := material.Caption(th, "Versión nativa para Windows · Medicina Interna · Hospital Dr. Eduardo Pereira")
						sub.Color = colorSuave
						return sub.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
					layout.Rigid(func(gtx C) D {
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								boton := material.Button(th, &st.botonCarpeta, "Elegir carpeta de informes")
								boton.Background = colorAcento
								if cargando {
									boton.Text = "Leyendo…"
								}
								return boton.Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
							layout.Rigid(func(gtx C) D {
								if !tieneDatos {
									gtx = gtx.Disabled()
								}
								return material.Button(th, &st.botonExcel, "Exportar Excel").Layout(gtx)
							}),
						)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
					layout.Rigid(func(gtx C) D {
						nota := material.Body2(th, mensaje)
						nota.Color = colorSuave
						return nota.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
					layout.Flexed(1, func(gtx C) D {
						return material.List(th, &st.lista).Layout(gtx, len(lineas), func(gtx C, indice int) D {
							return layout.Inset{Top: unit.Dp(4), Bottom: unit.Dp(4)}.Layout(gtx, func(gtx C) D {
								etiqueta := material.Body1(th, lineas[indice])
								etiqueta.Color = colorTexto
								return etiqueta.Layout(gtx)
							})
						})
					}),
				)
			})
		}),
		layout.Stacked(func(gtx C) D {
			return layout.Inset{Right: unit.Dp(14), Bottom: unit.Dp(10)}.Layout(gtx, func(gtx C) D {
				return dibujarFirma(gtx, th)
			})
		}),
	)
}

func dibujarFirma(gtx C, th *material.Theme) D {
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			sello := material.Body2(th, firma.Texto)
			sello.Color = colorFirma
			sello.Font.Style = font.Italic
			return sello.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			ver := material.Caption(th, "v"+version.Aplicacion)
			ver.Color = colorFirma
			return ver.Layout(gtx)
		}),
	)
}

func (st *estado) alElegirCarpeta() {
	if carpeta, ok := dialogos.ElegirCarpeta("Elija la carpeta con los informes Word y las planillas Excel"); ok {
		st.procesarCarpeta(carpeta)
	}
}

func (st *estado) alExportar() {
	st.mu.Lock()
	carpeta := st.carpetaActual
	st.mu.Unlock()
	nombre := "Yodo131_" + time.Now().Format("2006-01-02") + ".xlsx"
	if destino, ok := dialogos.GuardarComo("Guardar el Excel", nombre, carpeta); ok {
		st.exportar(destino)
	}
}

func ejecutarInterfaz() error {
	st := nuevoEstado()
	ventana := new(app.Window)
	ventana.Option(app.Title("Lector Yodo 131"), app.Size(unit.Dp(1000), unit.Dp(700)))
	st.mu.Lock()
	st.ventana = ventana
	st.mu.Unlock()
	th := tema()
	var ops op.Ops
	for {
		switch evento := ventana.Event().(type) {
		case app.DestroyEvent:
			return evento.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, evento)
			st.dibujar(gtx, th)
			evento.Frame(gtx.Ops)
		}
	}
}

func construirImagen(ancho, alto int, carpeta string) (image.Image, error) {
	st := nuevoEstado()
	if carpeta != "" {
		st.procesarCarpeta(carpeta)
	}
	th := tema()
	return render(ancho, alto, func(gtx C) D { return st.dibujar(gtx, th) })
}
