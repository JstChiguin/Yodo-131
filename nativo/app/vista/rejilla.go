package vista

import (
	"gioui.org/layout"
	"gioui.org/unit"
)

func (t *Tema) enRejilla(gtx C, cantidad, anchoMinimo int, celda func(C, int) D) D {
	if cantidad == 0 {
		return D{}
	}
	columnas := gtx.Constraints.Max.X / anchoMinimo
	if columnas < 1 {
		columnas = 1
	}
	if columnas > cantidad {
		columnas = cantidad
	}
	return t.rejillaColumnas(gtx, cantidad, columnas, celda)
}

func (t *Tema) rejillaColumnas(gtx C, cantidad, columnas int, celda func(C, int) D) D {
	gap := unit.Dp(12)
	filas := (cantidad + columnas - 1) / columnas
	hijosFila := make([]layout.FlexChild, 0, filas)
	for fila := 0; fila < filas; fila++ {
		fila := fila
		hijosFila = append(hijosFila, layout.Rigid(func(gtx C) D {
			columnasHijos := make([]layout.FlexChild, 0, columnas)
			for col := 0; col < columnas; col++ {
				indice := fila*columnas + col
				col := col
				indice2 := indice
				columnasHijos = append(columnasHijos, layout.Flexed(1, func(gtx C) D {
					izquierda := gap
					if col == 0 {
						izquierda = 0
					}
					return layout.Inset{Left: izquierda}.Layout(gtx, func(gtx C) D {
						if indice2 >= cantidad {
							return D{}
						}
						return celda(gtx, indice2)
					})
				}))
			}
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx, columnasHijos...)
		}))
		if fila < filas-1 {
			hijosFila = append(hijosFila, layout.Rigid(layout.Spacer{Height: gap}.Layout))
		}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, hijosFila...)
}

func (t *Tema) layoutTarjetas(gtx C, tarjetas []especificacionTarjeta) D {
	hijos := []layout.FlexChild{}
	agregarEspacio := func() {
		if len(hijos) > 0 {
			hijos = append(hijos, layout.Rigid(layout.Spacer{Height: 12}.Layout))
		}
	}
	indice := 0
	for indice < len(tarjetas) {
		if tarjetas[indice].ancha {
			w := tarjetas[indice].contenido
			agregarEspacio()
			hijos = append(hijos, layout.Rigid(w))
			indice++
			continue
		}
		primera := tarjetas[indice].contenido
		var segunda layout.Widget
		indice++
		if indice < len(tarjetas) && !tarjetas[indice].ancha {
			segunda = tarjetas[indice].contenido
			indice++
		}
		agregarEspacio()
		hijos = append(hijos, layout.Rigid(func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx,
				layout.Flexed(1, primera),
				layout.Flexed(1, func(gtx C) D {
					if segunda == nil {
						return D{}
					}
					return layout.Inset{Left: 12}.Layout(gtx, segunda)
				}),
			)
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, hijos...)
}
