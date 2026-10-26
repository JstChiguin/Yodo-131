package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func obtenerCarpetaDatos() (string, error) {
	carpetaCache, errorConsulta := os.UserCacheDir()
	if errorConsulta != nil {
		return "", errorConsulta
	}
	return filepath.Join(carpetaCache, "Yodo131"), nil
}

func listarCandidatosNavegador() []string {
	candidatos := []string{}
	for _, nombre := range []string{"microsoft-edge", "chromium", "chromium-browser", "google-chrome"} {
		if ruta, errorBusqueda := exec.LookPath(nombre); errorBusqueda == nil {
			candidatos = append(candidatos, ruta)
		}
	}
	return candidatos
}

func abrirConNavegadorPredeterminado(rutaPagina string) error {
	return exec.Command("xdg-open", rutaPagina).Start()
}

func mostrarMensaje(titulo, mensaje string, esError bool) {
	nivel := "Aviso"
	if esError {
		nivel = "Error"
	}
	fmt.Fprintf(os.Stderr, "%s - %s: %s\n", titulo, nivel, mensaje)
}
