package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const tituloAplicacion = "Lector Yodo 131"

func main() {
	carpetaDatos, versionAplicacion, errorInicio := iniciarAplicacion()
	if errorInicio == nil {
		return
	}
	registrarErrorLanzador(carpetaDatos, "lanzador", "iniciarAplicacion", errorInicio, versionAplicacion)
	mostrarMensaje(tituloAplicacion, describirErrorParaUsuario(errorInicio), true)
	os.Exit(1)
}

func describirErrorParaUsuario(errorInicio error) string {
	if errors.Is(errorInicio, errSinNavegador) {
		return "No se pudo abrir el Lector Yodo 131 porque este computador no tiene Microsoft Edge ni Google Chrome.\n\nInstale uno de ellos o pida ayuda a Chiguin."
	}
	return "No se pudo abrir el Lector Yodo 131.\n\nVuelva a intentarlo. Si el problema continúa, avise a Chiguin: el detalle quedó guardado en el registro del programa."
}

func iniciarAplicacion() (string, string, error) {
	carpetaDatos, errorCarpeta := obtenerCarpetaDatos()
	if errorCarpeta != nil {
		return "", "", fmt.Errorf("no se encontró la carpeta de datos del usuario: %w", errorCarpeta)
	}
	versionAplicacion, errorVersion := obtenerVersionAplicacion(archivosAplicacion)
	if errorVersion != nil {
		return carpetaDatos, "", errorVersion
	}
	carpetaAplicacion := filepath.Join(carpetaDatos, "aplicacion", versionAplicacion)
	if _, errorExtraccion := extraerAplicacion(archivosAplicacion, carpetaAplicacion); errorExtraccion != nil {
		return carpetaDatos, versionAplicacion, fmt.Errorf("no se pudo preparar la aplicación: %w", errorExtraccion)
	}
	paginaInicio := filepath.Join(carpetaAplicacion, "index.html")

	rutaNavegador, hayNavegador := elegirNavegador(listarCandidatosNavegador(), existeArchivo)
	if !hayNavegador {
		if errorApertura := abrirConNavegadorPredeterminado(paginaInicio); errorApertura != nil {
			return carpetaDatos, versionAplicacion, fmt.Errorf("%w y no se pudo usar el navegador predeterminado: %v", errSinNavegador, errorApertura)
		}
		return carpetaDatos, versionAplicacion, nil
	}

	argumentos := construirArgumentosNavegador(convertirRutaEnDireccion(paginaInicio), filepath.Join(carpetaDatos, "perfil-navegador"))
	procesoNavegador := exec.Command(rutaNavegador, argumentos...)
	if errorArranque := procesoNavegador.Start(); errorArranque != nil {
		return carpetaDatos, versionAplicacion, fmt.Errorf("no se pudo iniciar %s: %w", filepath.Base(rutaNavegador), errorArranque)
	}
	return carpetaDatos, versionAplicacion, procesoNavegador.Process.Release()
}
