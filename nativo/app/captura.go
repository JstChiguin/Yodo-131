package main

import (
	"image"

	"gioui.org/gpu/headless"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
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
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
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
