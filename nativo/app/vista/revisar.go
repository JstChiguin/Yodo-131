package vista

import (
	"strconv"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/widget"

	"yodo131/nucleo/modelo"
)

type itemRevisar struct {
	titulo  string
	detalle string
}

func (v *Vista) layoutRevisar(gtx C, t *Tema, datos Datos) D {
	conflictos := []itemRevisar{}
	for _, c := range datos.Resultado.Conflictos {
		fecha := "sin fecha"
		if iso, ok := c.FechaIso.(string); ok && iso != "" {
			fecha = fechaBonita(iso)
		} else if a, ok := c.Anio.(int); ok {
			fecha = strconv.Itoa(a)
		}
		nombre := c.NombrePaciente
		if nombre == "" {
			nombre = "Sin nombre"
		}
		if rut, ok := c.Rut.(string); ok && rut != "" {
			nombre += " (" + rut + ")"
		}
		valores := []string{}
		for _, val := range c.Valores {
			valores = append(valores, val.Origen+": «"+val.Valor+"» en "+val.Referencia)
		}
		conflictos = append(conflictos, itemRevisar{nombre + " — " + c.NombreCampo + ", tratamiento del " + fecha, strings.Join(valores, " · ")})
	}
	if len(conflictos) == 0 {
		conflictos = append(conflictos, itemRevisar{"No hay datos contradictorios entre fuentes.", ""})
	}

	faltantes := []itemRevisar{}
	for _, tr := range datos.Resultado.Tratamientos {
		campos, ok := tr.Get("camposFaltantes").([]string)
		if !ok || len(campos) == 0 {
			continue
		}
		nombre := tr.Texto("nombrePaciente")
		if nombre == "" {
			nombre = "Sin nombre"
		}
		refs := []string{}
		if fuentes, ok := tr.Get("fuentes").([]modelo.Fuente); ok {
			for _, f := range fuentes {
				refs = append(refs, f.Referencia)
			}
		}
		faltantes = append(faltantes, itemRevisar{nombre + " — falta " + strings.Join(campos, ", "), strings.Join(refs, " · ")})
	}
	if len(faltantes) == 0 {
		faltantes = append(faltantes, itemRevisar{"Todos los tratamientos tienen sus datos principales.", ""})
	}

	noLeidos := []itemRevisar{}
	for _, d := range datos.Resultado.DocumentosNoLeidos {
		noLeidos = append(noLeidos, itemRevisar{d.Archivo, d.Motivo})
	}
	if len(noLeidos) == 0 {
		noLeidos = append(noLeidos, itemRevisar{"Se pudieron leer todos los documentos.", ""})
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return v.tarjetaRevisar(gtx, t, "Datos que no coinciden entre fuentes",
				"Cuando el Word y el Excel (o dos registros del mismo tratamiento) dicen cosas distintas, el programa no elige por su cuenta: el dato queda fuera de los gráficos hasta que se corrija en el archivo de origen.", conflictos)
		}),
		layout.Rigid(func(gtx C) D { return layout.Spacer{Height: 12}.Layout(gtx) }),
		layout.Rigid(func(gtx C) D {
			return v.tarjetaRevisar(gtx, t, "Tratamientos con datos que faltan",
				"El programa no inventa datos: si algo no está en el Word ni en el Excel, aparece aquí. Se corrige en el archivo de origen y se vuelve a elegir la carpeta.", faltantes)
		}),
		layout.Rigid(func(gtx C) D { return layout.Spacer{Height: 12}.Layout(gtx) }),
		layout.Rigid(func(gtx C) D {
			return v.tarjetaRevisar(gtx, t, "Documentos que no se pudieron leer", "", noLeidos)
		}),
	)
}

func (v *Vista) tarjetaRevisar(gtx C, t *Tema, titulo, nota string, items []itemRevisar) D {
	return t.tarjeta(gtx, func(gtx C) D {
		hijos := []layout.FlexChild{
			layout.Rigid(func(gtx C) D { return t.TituloTarjeta(titulo).Layout(gtx) }),
		}
		if nota != "" {
			hijos = append(hijos, layout.Rigid(func(gtx C) D {
				return layout.Inset{Top: 2, Bottom: 4}.Layout(gtx, func(gtx C) D { return t.Nota(nota).Layout(gtx) })
			}))
		}
		for _, item := range items {
			item := item
			hijos = append(hijos, layout.Rigid(func(gtx C) D {
				return layout.Inset{Top: 6}.Layout(gtx, func(gtx C) D { return v.itemListaRevisar(gtx, t, item) })
			}))
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, hijos...)
	})
}

func (v *Vista) itemListaRevisar(gtx C, t *Tema, item itemRevisar) D {
	return widget.Border{Color: t.Borde, CornerRadius: 8, Width: 1}.Layout(gtx, func(gtx C) D {
		return layout.Inset{Top: 8, Bottom: 8, Left: 10, Right: 10}.Layout(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D { return t.etiqueta(13, item.titulo, t.Texto, TipografiaCuerpo).Layout(gtx) }),
				layout.Rigid(func(gtx C) D {
					if item.detalle == "" {
						return D{}
					}
					return layout.Inset{Top: 2}.Layout(gtx, func(gtx C) D { return t.Pequeno(item.detalle, t.Suave).Layout(gtx) })
				}),
			)
		})
	})
}

var _ = font.Bold
