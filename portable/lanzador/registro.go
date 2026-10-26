package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	nombreRegistroErrores    = "registro-lanzador.log"
	tamanoMaximoRegistroByte = 256 * 1024
)

type registroErrorLanzador struct {
	FechaHora         string `json:"fechaHora"`
	Componente        string `json:"componente"`
	Metodo            string `json:"metodo"`
	TipoExcepcion     string `json:"tipoExcepcion"`
	Mensaje           string `json:"mensaje"`
	VersionAplicacion string `json:"versionAplicacion"`
}

func ocultarCarpetaPersonal(mensaje string) string {
	carpetaPersonal, errorConsulta := os.UserHomeDir()
	if errorConsulta != nil || carpetaPersonal == "" {
		return mensaje
	}
	return strings.ReplaceAll(mensaje, carpetaPersonal, "~")
}

func crearRegistroError(componente, metodo string, errorOcurrido error, versionAplicacion string, momento time.Time) registroErrorLanzador {
	return registroErrorLanzador{
		FechaHora:         momento.Format(time.RFC3339),
		Componente:        componente,
		Metodo:            metodo,
		TipoExcepcion:     fmt.Sprintf("%T", errorOcurrido),
		Mensaje:           ocultarCarpetaPersonal(errorOcurrido.Error()),
		VersionAplicacion: versionAplicacion,
	}
}

func registrarErrorLanzador(carpetaDatos, componente, metodo string, errorOcurrido error, versionAplicacion string) {
	registro := crearRegistroError(componente, metodo, errorOcurrido, versionAplicacion, time.Now())
	linea, errorFormato := json.Marshal(registro)
	if errorFormato != nil || carpetaDatos == "" {
		fmt.Fprintln(os.Stderr, "[Yodo131] no se pudo registrar el error:", registro.Mensaje)
		return
	}
	if errorCarpeta := os.MkdirAll(carpetaDatos, 0o755); errorCarpeta != nil {
		fmt.Fprintln(os.Stderr, "[Yodo131] no se pudo registrar el error:", registro.Mensaje)
		return
	}
	rutaRegistro := filepath.Join(carpetaDatos, nombreRegistroErrores)
	if informacion, errorConsulta := os.Stat(rutaRegistro); errorConsulta == nil && informacion.Size() > tamanoMaximoRegistroByte {
		os.Rename(rutaRegistro, rutaRegistro+".anterior")
	}
	archivo, errorApertura := os.OpenFile(rutaRegistro, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if errorApertura != nil {
		fmt.Fprintln(os.Stderr, "[Yodo131] no se pudo registrar el error:", registro.Mensaje)
		return
	}
	defer archivo.Close()
	if _, errorEscritura := archivo.Write(append(linea, '\n')); errorEscritura != nil {
		fmt.Fprintln(os.Stderr, "[Yodo131] no se pudo registrar el error:", registro.Mensaje)
	}
}
