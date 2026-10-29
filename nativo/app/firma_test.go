package main

import (
	"image"
	"testing"

	"gioui.org/layout"

	"yodo131/app/vista"
	"yodo131/nucleo/firma"
	"yodo131/nucleo/version"
)

func renderParaPrueba(t *testing.T, oscuro bool, ancho, alto int) *image.RGBA {
	t.Helper()
	a := nuevaAplicacion()
	if oscuro {
		a.oscuro = true
		a.tema = vista.NuevoTema(true)
		a.vista.Oscuro = true
	}
	a.vista.Seccion = "resumen"
	imagen, err := render(ancho, alto, func(gtx layout.Context) layout.Dimensions {
		a.vista.Layout(gtx, a.tema, a.datos)
		return layout.Dimensions{Size: gtx.Constraints.Max}
	})
	if err != nil {
		t.Fatalf("no se pudo renderizar: %v", err)
	}
	rgba, ok := imagen.(*image.RGBA)
	if !ok {
		t.Fatal("la imagen renderizada no es RGBA")
	}
	return rgba
}

func diferenciaColor(a, b [4]uint8) int {
	suma := 0
	for i := 0; i < 3; i++ {
		d := int(a[i]) - int(b[i])
		if d < 0 {
			d = -d
		}
		suma += d
	}
	return suma
}

func TestSelloIntegridad(t *testing.T) {
	if firma.Texto != "by Chiguin" {
		t.Fatalf("la firma de autor fue alterada: se esperaba \"by Chiguin\" y se obtuvo %q", firma.Texto)
	}
	if version.Aplicacion == "" {
		t.Fatal("la versión central de la aplicación está vacía")
	}
}

func TestSelloVisibleYEnEsquinaInferiorDerecha(t *testing.T) {
	for _, caso := range []struct {
		nombre string
		oscuro bool
	}{{"claro", false}, {"oscuro", true}} {
		t.Run(caso.nombre, func(t *testing.T) {
			ancho, alto := 900, 700
			img := renderParaPrueba(t, caso.oscuro, ancho, alto)
			muestraFondo := img.RGBAAt(6, 6)
			fondo := [4]uint8{muestraFondo.R, muestraFondo.G, muestraFondo.B, muestraFondo.A}

			contarDistintos := func(x0, y0, x1, y1 int) int {
				total := 0
				for y := y0; y < y1; y++ {
					for x := x0; x < x1; x++ {
						p := img.RGBAAt(x, y)
						if diferenciaColor([4]uint8{p.R, p.G, p.B, p.A}, fondo) > 30 {
							total++
						}
					}
				}
				return total
			}

			derecha := contarDistintos(ancho-180, alto-74, ancho-6, alto-6)
			izquierda := contarDistintos(6, alto-74, 180, alto-6)

			if derecha < 120 {
				t.Errorf("el sello no es visible en la esquina inferior derecha (%d píxeles distintos del fondo)", derecha)
			}
			if izquierda > derecha/4 {
				t.Errorf("hay contenido inesperado en la esquina inferior izquierda (%d píxeles); el sello debe ubicarse a la derecha", izquierda)
			}
		})
	}
}
