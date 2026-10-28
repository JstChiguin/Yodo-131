package main

import (
	"image"
	"image/png"
	"os"

	"gioui.org/gpu/headless"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"yodo131/app/vista"
)

func render(ancho, alto int, dibujar func(layout.Context) layout.Dimensions) (image.Image, error) {
	ventana, err := headless.NewWindow(ancho, alto)
	if err != nil {
		return nil, err
	}
	defer ventana.Release()
	var ops op.Ops
	gtx := layout.Context{
		Ops:         &ops,
		Constraints: layout.Exact(image.Pt(ancho, alto)),
		Metric:      unit.Metric{PxPerDp: 1.4, PxPerSp: 1.4},
	}
	dibujar(gtx)
	if err := ventana.Frame(gtx.Ops); err != nil {
		return nil, err
	}
	imagen := image.NewRGBA(image.Rect(0, 0, ancho, alto))
	if err := ventana.Screenshot(imagen); err != nil {
		return nil, err
	}
	return imagen, nil
}

func altoPantalla(pantalla string) int {
	switch pantalla {
	case "resumen":
		return 3400
	case "pacientes", "revisar", "mantenimiento":
		return 1500
	default:
		return 820
	}
}

func capturarPantalla(destino, pantalla, temaNombre, carpeta string, presentacion bool) error {
	a := nuevaAplicacion()
	if carpeta != "" {
		a.cargarCarpeta(carpeta)
	}
	if temaNombre == "oscuro" {
		a.oscuro = true
		a.tema = vista.NuevoTema(true)
		a.vista.Oscuro = true
	}
	if pantalla == "bienvenida" {
		a.vista.Seccion = "resumen"
	} else {
		a.vista.Seccion = pantalla
	}
	if pantalla == "mantenimiento" {
		a.vista.SetClaveValida(true, "")
	}
	a.vista.Presentacion = presentacion
	ancho := 1180
	alto := altoPantalla(pantalla)
	imagen, err := render(ancho, alto, func(gtx layout.Context) layout.Dimensions {
		a.vista.Layout(gtx, a.tema, a.datos)
		return layout.Dimensions{Size: gtx.Constraints.Max}
	})
	if err != nil {
		return err
	}
	archivo, err := os.Create(destino)
	if err != nil {
		return err
	}
	defer archivo.Close()
	return png.Encode(archivo, imagen)
}
