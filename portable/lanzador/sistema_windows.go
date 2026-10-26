package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

const (
	estiloMensajeError       = 0x10
	estiloMensajeInformacion = 0x40
)

func obtenerCarpetaDatos() (string, error) {
	carpetaLocal := os.Getenv("LOCALAPPDATA")
	if carpetaLocal == "" {
		carpetaCache, errorConsulta := os.UserCacheDir()
		if errorConsulta != nil {
			return "", errorConsulta
		}
		carpetaLocal = carpetaCache
	}
	return filepath.Join(carpetaLocal, "Yodo131"), nil
}

func listarCandidatosNavegador() []string {
	return candidatosNavegadorWindows(os.Getenv)
}

func abrirConNavegadorPredeterminado(rutaPagina string) error {
	return exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", rutaPagina).Start()
}

func mostrarMensaje(titulo, mensaje string, esError bool) {
	punteroTitulo, errorTitulo := syscall.UTF16PtrFromString(titulo)
	punteroMensaje, errorMensaje := syscall.UTF16PtrFromString(mensaje)
	if errorTitulo != nil || errorMensaje != nil {
		return
	}
	estilo := uintptr(estiloMensajeInformacion)
	if esError {
		estilo = estiloMensajeError
	}
	cuadroMensaje := syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW")
	if cuadroMensaje.Find() != nil {
		return
	}
	cuadroMensaje.Call(0, uintptr(unsafe.Pointer(punteroMensaje)), uintptr(unsafe.Pointer(punteroTitulo)), estilo)
}
