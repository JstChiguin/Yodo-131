package main

import (
	"flag"
	"fmt"
	"os"

	gioapp "gioui.org/app"

	"yodo131/nucleo/clasificacion"
	"yodo131/nucleo/estadisticas"
	"yodo131/nucleo/exportacion"
	"yodo131/nucleo/firma"
	"yodo131/nucleo/procesamiento"
	"yodo131/nucleo/resumen"
	"yodo131/nucleo/version"
)

func main() {
	autoverificacion := flag.String("autoverificacion", "", "carpeta de ejemplos ficticios")
	salida := flag.String("salida", "salida.xlsx", "Excel de salida de la autoverificación")
	captura := flag.String("captura", "", "archivo PNG de salida")
	pantalla := flag.String("pantalla", "resumen", "pantalla a capturar")
	temaNombre := flag.String("tema", "claro", "claro u oscuro")
	carpeta := flag.String("carpeta", "", "carpeta de datos para la captura")
	presentacion := flag.Bool("presentacion", false, "modo presentación en la captura")
	desde := flag.String("desde", "", "mes inicial del período (yyyy-mm)")
	hasta := flag.String("hasta", "", "mes final del período (yyyy-mm)")
	resaltar := flag.String("resaltar", "", "id de tarjeta o cifra a resaltar en la captura")
	flag.Parse()

	if *autoverificacion != "" {
		if err := ejecutarAutoverificacion(*autoverificacion, *salida); err != nil {
			fmt.Fprintln(os.Stderr, "autoverificación:", err)
			os.Exit(1)
		}
		return
	}
	if *captura != "" {
		if err := capturarPantalla(*captura, *pantalla, *temaNombre, *carpeta, *presentacion, *desde, *hasta, *resaltar); err != nil {
			fmt.Fprintln(os.Stderr, "captura:", err)
			os.Exit(1)
		}
		return
	}

	go func() {
		if err := ejecutarInterfaz(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	gioapp.Main()
}

func ejecutarAutoverificacion(carpeta, salida string) error {
	if firma.Texto != "by Chiguin" {
		return fmt.Errorf("FIRMA DE AUTOR ALTERADA: debe ser exactamente \"by Chiguin\"")
	}
	archivos, err := procesamiento.LeerCarpeta(carpeta)
	if err != nil {
		return err
	}
	if len(archivos) == 0 {
		return fmt.Errorf("no se encontraron archivos .docx ni .xlsx en %s", carpeta)
	}
	lectura := procesamiento.ProcesarArchivos(archivos, nil)
	resultado := procesamiento.ConstruirResultado(lectura, clasificacion.UmbralesPorDefecto)
	est := estadisticas.Calcular(resultado, estadisticas.Filtro{})
	datos := exportacion.ExportarLibro(est, resultado, "Todos los tratamientos")
	if err := os.WriteFile(salida, datos, 0o644); err != nil {
		return err
	}
	fmt.Println("Lector Yodo 131 nativo · v" + version.Aplicacion + " · " + firma.Texto)
	fmt.Printf("Archivos leídos: %d (%d Word, %d Excel)\n", lectura.CantidadArchivos, lectura.Conteo.ArchivosWord, lectura.Conteo.ArchivosExcel)
	for _, linea := range resumen.Lineas(est) {
		fmt.Println("  " + linea)
	}
	fmt.Printf("Excel exportado: %s (%d bytes)\n", salida, len(datos))
	return nil
}
