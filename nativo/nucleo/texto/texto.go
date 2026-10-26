package texto

import "strings"

var baseDiacriticos = map[rune]rune{
	'á': 'a', 'é': 'e', 'í': 'i', 'ó': 'o', 'ú': 'u', 'ü': 'u', 'ñ': 'n',
	'à': 'a', 'è': 'e', 'ì': 'i', 'ò': 'o', 'ù': 'u',
	'â': 'a', 'ê': 'e', 'î': 'i', 'ô': 'o', 'û': 'u',
	'ä': 'a', 'ë': 'e', 'ï': 'i', 'ö': 'o',
	'ã': 'a', 'õ': 'o', 'ç': 'c', 'ý': 'y', 'ÿ': 'y',
	'Á': 'A', 'É': 'E', 'Í': 'I', 'Ó': 'O', 'Ú': 'U', 'Ü': 'U', 'Ñ': 'N',
	'À': 'A', 'È': 'E', 'Ì': 'I', 'Ò': 'O', 'Ù': 'U',
	'Â': 'A', 'Ê': 'E', 'Î': 'I', 'Ô': 'O', 'Û': 'U',
	'Ä': 'A', 'Ë': 'E', 'Ï': 'I', 'Ö': 'O',
	'Ã': 'A', 'Õ': 'O', 'Ç': 'C',
}

func QuitarDiacriticos(texto string) string {
	var constructor strings.Builder
	for _, caracter := range texto {
		if caracter >= 0x0300 && caracter <= 0x036f {
			continue
		}
		if base, ok := baseDiacriticos[caracter]; ok {
			constructor.WriteRune(base)
			continue
		}
		constructor.WriteRune(caracter)
	}
	return constructor.String()
}
