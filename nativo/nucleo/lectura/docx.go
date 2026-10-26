package lectura

import (
	"regexp"
	"strings"
)

var (
	bloqueParrafo    = regexp.MustCompile(`<w:p[ >][\s\S]*?</w:p>`)
	piezaTexto       = regexp.MustCompile(`<w:t(?: [^>]*)?>[^<]*</w:t>|<w:tab/>|<w:br/>`)
	etiquetaXml      = regexp.MustCompile(`<[^>]+>`)
	espaciosColapsar = regexp.MustCompile("[  ]+")
)

func extraerParrafos(xmlDocumento string) []string {
	parrafos := []string{}
	for _, bloque := range bloqueParrafo.FindAllString(xmlDocumento, -1) {
		var constructor strings.Builder
		for _, pieza := range piezaTexto.FindAllString(bloque, -1) {
			switch pieza {
			case "<w:tab/>":
				constructor.WriteString(" ")
			case "<w:br/>":
				constructor.WriteString("\n")
			default:
				constructor.WriteString(DecodificarEntidades(etiquetaXml.ReplaceAllString(pieza, "")))
			}
		}
		textoLimpio := strings.TrimSpace(espaciosColapsar.ReplaceAllString(constructor.String(), " "))
		if textoLimpio == "" {
			continue
		}
		for _, linea := range strings.Split(textoLimpio, "\n") {
			linea = strings.TrimSpace(linea)
			if linea != "" {
				parrafos = append(parrafos, linea)
			}
		}
	}
	return parrafos
}

func LeerParrafosDocx(datos []byte) ([]string, error) {
	if EsFormatoOfficeAntiguo(datos) {
		return nil, &FormatoNoSoportadoError{Mensaje: `Es un Word antiguo (.doc). Ábralo en Word y use "Guardar como" .docx`}
	}
	lector, err := abrirZip(datos)
	if err != nil {
		return nil, err
	}
	documento, existe, err := leerEntrada(lector, "word/document.xml")
	if err != nil {
		return nil, err
	}
	if !existe {
		return nil, &FormatoNoSoportadoError{Mensaje: "Falta la parte interna word/document.xml"}
	}
	return extraerParrafos(documento), nil
}
