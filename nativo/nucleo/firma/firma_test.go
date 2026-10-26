package firma

import (
	"regexp"
	"testing"

	"yodo131/nucleo/version"
)

func TestFirmaExacta(t *testing.T) {
	if Texto != "by Chiguin" {
		t.Fatalf("FIRMA DE AUTOR ALTERADA: el texto de la firma debe ser exactamente \"by Chiguin\" y es %q", Texto)
	}
}

func TestVersionSemantica(t *testing.T) {
	if !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(version.Aplicacion) {
		t.Fatalf("la versión central debe seguir versionado semántico y es %q", version.Aplicacion)
	}
}
