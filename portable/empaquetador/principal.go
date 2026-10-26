package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "uso: empaquetador <archivo .zip de salida> <carpeta a comprimir>")
		os.Exit(2)
	}
	if errorEmpaquetado := comprimirCarpeta(os.Args[1], os.Args[2]); errorEmpaquetado != nil {
		fmt.Fprintln(os.Stderr, "empaquetador:", errorEmpaquetado)
		os.Exit(1)
	}
}

func comprimirCarpeta(rutaZip, carpetaOrigen string) error {
	archivoZip, errorCreacion := os.Create(rutaZip)
	if errorCreacion != nil {
		return errorCreacion
	}
	escritorZip := zip.NewWriter(archivoZip)
	carpetaBase := filepath.Dir(carpetaOrigen)

	errorRecorrido := filepath.WalkDir(carpetaOrigen, func(ruta string, entrada fs.DirEntry, errorEntrada error) error {
		if errorEntrada != nil || entrada.IsDir() {
			return errorEntrada
		}
		return agregarArchivo(escritorZip, carpetaBase, ruta)
	})
	errorCierreZip := escritorZip.Close()
	errorCierreArchivo := archivoZip.Close()
	for _, errorPaso := range []error{errorRecorrido, errorCierreZip, errorCierreArchivo} {
		if errorPaso != nil {
			os.Remove(rutaZip)
			return errorPaso
		}
	}
	return nil
}

func agregarArchivo(escritorZip *zip.Writer, carpetaBase, ruta string) error {
	informacion, errorConsulta := os.Stat(ruta)
	if errorConsulta != nil {
		return errorConsulta
	}
	encabezado, errorEncabezado := zip.FileInfoHeader(informacion)
	if errorEncabezado != nil {
		return errorEncabezado
	}
	rutaRelativa, errorRelativa := filepath.Rel(carpetaBase, ruta)
	if errorRelativa != nil {
		return errorRelativa
	}
	encabezado.Name = filepath.ToSlash(rutaRelativa)
	encabezado.Method = zip.Deflate
	destino, errorDestino := escritorZip.CreateHeader(encabezado)
	if errorDestino != nil {
		return errorDestino
	}
	origen, errorApertura := os.Open(ruta)
	if errorApertura != nil {
		return errorApertura
	}
	defer origen.Close()
	_, errorCopia := io.Copy(destino, origen)
	return errorCopia
}
