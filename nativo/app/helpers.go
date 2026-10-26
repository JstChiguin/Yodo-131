package main

import (
	"image/color"
	"os"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"

	"yodo131/nucleo/clasificacion"
)

func rellenar(gtx layout.Context, fondo color.NRGBA) {
	defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: fondo}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

func clasificacionUmbrales() map[string]float64 {
	return clasificacion.UmbralesPorDefecto
}

func escribirArchivo(ruta string, datos []byte) error {
	return os.WriteFile(ruta, datos, 0o644)
}
