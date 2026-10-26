package main

import (
	"errors"
	"net/url"
	"os"
	"strings"
)

var errSinNavegador = errors.New("no se encontró Microsoft Edge ni Google Chrome")

func candidatosNavegadorWindows(obtenerVariable func(string) string) []string {
	carpetasProgramas := []string{
		obtenerVariable("ProgramFiles(x86)"),
		obtenerVariable("ProgramFiles"),
		obtenerVariable("LOCALAPPDATA"),
	}
	ejecutables := []string{
		`Microsoft\Edge\Application\msedge.exe`,
		`Google\Chrome\Application\chrome.exe`,
	}
	candidatos := []string{}
	yaIncluidos := map[string]bool{}
	for _, ejecutable := range ejecutables {
		for _, carpeta := range carpetasProgramas {
			if carpeta == "" {
				continue
			}
			candidato := strings.TrimRight(carpeta, `\`) + `\` + ejecutable
			if yaIncluidos[strings.ToLower(candidato)] {
				continue
			}
			yaIncluidos[strings.ToLower(candidato)] = true
			candidatos = append(candidatos, candidato)
		}
	}
	return candidatos
}

func elegirNavegador(candidatos []string, existeArchivo func(string) bool) (string, bool) {
	for _, candidato := range candidatos {
		if existeArchivo(candidato) {
			return candidato, true
		}
	}
	return "", false
}

func existeArchivo(ruta string) bool {
	informacion, errorConsulta := os.Stat(ruta)
	return errorConsulta == nil && !informacion.IsDir()
}

func convertirRutaEnDireccion(ruta string) string {
	rutaConBarras := strings.ReplaceAll(ruta, `\`, "/")
	if !strings.HasPrefix(rutaConBarras, "/") {
		rutaConBarras = "/" + rutaConBarras
	}
	return (&url.URL{Scheme: "file", Path: rutaConBarras}).String()
}

func construirArgumentosNavegador(direccionPagina, carpetaPerfil string) []string {
	return []string{
		"--app=" + direccionPagina,
		"--user-data-dir=" + carpetaPerfil,
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-extensions",
		"--disable-sync",
		"--disable-background-networking",
		"--disable-component-update",
		"--start-maximized",
	}
}
