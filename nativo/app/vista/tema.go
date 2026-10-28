package vista

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

const TipografiaCuerpo = "Segoe UI"
const TipografiaTitulo = "Georgia"

type Tema struct {
	Th     *material.Theme
	Oscuro bool

	Fondo       color.NRGBA
	Panel       color.NRGBA
	Texto       color.NRGBA
	Suave       color.NRGBA
	Borde       color.NRGBA
	BordeFuerte color.NRGBA
	Acento      color.NRGBA
	AcentoTexto color.NRGBA
	AcentoSuave color.NRGBA
	Mujeres     color.NRGBA
	Hombres     color.NRGBA
	SinDato     color.NRGBA
	Serie3      color.NRGBA
	Serie4      color.NRGBA
	Serie5      color.NRGBA
	Elevado     color.NRGBA
	Normal      color.NRGBA
	Error       color.NRGBA
	Rejilla     color.NRGBA
}

func rgb(hex uint32) color.NRGBA {
	return color.NRGBA{R: uint8(hex >> 16), G: uint8(hex >> 8), B: uint8(hex), A: 0xff}
}

func conAlfa(c color.NRGBA, alfa uint8) color.NRGBA {
	c.A = alfa
	return c
}

func mezclar(frente color.NRGBA, fondo color.NRGBA, proporcion float64) color.NRGBA {
	mezcla := func(a, b uint8) uint8 {
		return uint8(float64(a)*proporcion + float64(b)*(1-proporcion))
	}
	return color.NRGBA{R: mezcla(frente.R, fondo.R), G: mezcla(frente.G, fondo.G), B: mezcla(frente.B, fondo.B), A: 0xff}
}

var crearShaper = func() *text.Shaper {
	return text.NewShaper(text.WithCollection(gofont.Collection()))
}

func NuevoTema(oscuro bool) *Tema {
	th := material.NewTheme()
	th.Shaper = crearShaper()
	th.TextSize = unit.Sp(14)
	t := &Tema{Th: th, Oscuro: oscuro}
	if oscuro {
		t.Fondo = rgb(0x0f1615)
		t.Panel = rgb(0x17201f)
		t.Texto = rgb(0xe4ecea)
		t.Suave = rgb(0x9fb0ac)
		t.Borde = rgb(0x2a3735)
		t.BordeFuerte = rgb(0x3c4c49)
		t.Acento = rgb(0x4fb3a8)
		t.AcentoTexto = rgb(0x08201d)
		t.AcentoSuave = rgb(0x1d3532)
		t.Mujeres = rgb(0x4fb3a8)
		t.Hombres = rgb(0x8ea4e6)
		t.SinDato = rgb(0x3d4a48)
		t.Serie3 = rgb(0xa3c36a)
		t.Serie4 = rgb(0xe09a5b)
		t.Serie5 = rgb(0xc39bd3)
		t.Elevado = rgb(0xf08a4b)
		t.Normal = rgb(0x6cc391)
		t.Error = rgb(0xf97066)
		t.Rejilla = rgb(0x26322f)
	} else {
		t.Fondo = rgb(0xf3f6f6)
		t.Panel = rgb(0xffffff)
		t.Texto = rgb(0x15211f)
		t.Suave = rgb(0x55625f)
		t.Borde = rgb(0xdbe3e1)
		t.BordeFuerte = rgb(0xb7c4c1)
		t.Acento = rgb(0x0f6b63)
		t.AcentoTexto = rgb(0xffffff)
		t.AcentoSuave = rgb(0xe1efed)
		t.Mujeres = rgb(0x0f6b63)
		t.Hombres = rgb(0x4a64a8)
		t.SinDato = rgb(0xc9d1cf)
		t.Serie3 = rgb(0x7a9a3c)
		t.Serie4 = rgb(0xb86b2c)
		t.Serie5 = rgb(0x8b5a9e)
		t.Elevado = rgb(0xc2410c)
		t.Normal = rgb(0x2f855a)
		t.Error = rgb(0xb42318)
		t.Rejilla = rgb(0xe6ecea)
	}
	th.Palette.Fg = t.Texto
	th.Palette.Bg = t.Fondo
	th.Palette.ContrastBg = t.Acento
	th.Palette.ContrastFg = t.AcentoTexto
	return t
}

func (t *Tema) ColorSerie(indice int) color.NRGBA {
	switch indice {
	case 0:
		return t.Acento
	case 1:
		return t.Hombres
	case 2:
		return t.Serie3
	case 3:
		return t.Serie4
	default:
		return t.Serie5
	}
}

func (t *Tema) etiqueta(tam unit.Sp, texto string, c color.NRGBA, tipografia string) material.LabelStyle {
	estilo := material.Label(t.Th, tam, texto)
	estilo.Color = c
	if tipografia != "" {
		estilo.Font.Typeface = font.Typeface(tipografia)
	}
	return estilo
}

func (t *Tema) Cuerpo(texto string) material.LabelStyle {
	return t.etiqueta(14, texto, t.Texto, TipografiaCuerpo)
}

func (t *Tema) CuerpoSuave(texto string) material.LabelStyle {
	return t.etiqueta(13, texto, t.Suave, TipografiaCuerpo)
}

func (t *Tema) Nota(texto string) material.LabelStyle {
	return t.etiqueta(13, texto, t.Suave, TipografiaCuerpo)
}

func (t *Tema) Pequeno(texto string, c color.NRGBA) material.LabelStyle {
	return t.etiqueta(12, texto, c, TipografiaCuerpo)
}

func (t *Tema) TituloTarjeta(texto string) material.LabelStyle {
	estilo := t.etiqueta(15, texto, t.Texto, TipografiaCuerpo)
	estilo.Font.Weight = font.SemiBold
	return estilo
}

func (t *Tema) TituloPrincipal(texto string) material.LabelStyle {
	estilo := t.etiqueta(30, texto, t.Texto, TipografiaTitulo)
	return estilo
}

func (t *Tema) NumeroGrande(texto string) material.LabelStyle {
	return t.etiqueta(30, texto, t.Texto, TipografiaTitulo)
}
