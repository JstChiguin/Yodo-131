package seguridad

import "testing"

func TestClaveCorrecta(t *testing.T) {
	if !VerificarClave("chiguin") {
		t.Fatal("la clave correcta debería validar")
	}
}

func TestClaveIncorrecta(t *testing.T) {
	if VerificarClave("otra") {
		t.Fatal("una clave incorrecta no debería validar")
	}
}
