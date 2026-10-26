package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"

	"gioui.org/app"

	"yodo131/nucleo/estadisticas"
	"yodo131/nucleo/exportacion"
	"yodo131/nucleo/firma"
	"yodo131/nucleo/procesamiento"
	"yodo131/nucleo/resumen"
	"yodo131/nucleo/version"
)

func main() {
	autoverificacion := flag.String("autoverificacion", "", "carpeta de ejemplos ficticios para la autoverificación")
	salida := flag.String("salida", "salida.xlsx", "archivo Excel de salida en la autoverificación")
	captura := flag.String("captura", "", "archivo PNG donde guardar una captura de la interfaz")
	carpeta := flag.String("carpeta", "", "carpeta de datos para la captura")
	flag.Parse()

	if *autoverificacion != "" {
		if err := ejecutarAutoverificacion(*autoverificacion, *salida); err != nil {
			fmt.Fprintln(os.Stderr, "autoverificación:", err)
			os.Exit(1)
		}
		return
	}
	if *captura != "" {
		if err := ejecutarCaptura(*captura, *carpeta); err != nil {
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
	app.Main()
}

func ejecutarAutoverificacion(carpeta, salida string) error {
	if firma.Texto != "by Chiguin" {
		return fmt.Errorf("FIRMA DE AUTOR ALTERADA: el texto de la firma debe ser exactamente \"by Chiguin\"")
	}
	archivos, err := procesamiento.LeerCarpeta(carpeta)
	if err != nil {
		return err
	}
	if len(archivos) == 0 {
		return fmt.Errorf("no se encontraron archivos .docx ni .xlsx en %s", carpeta)
	}
	lectura := procesamiento.ProcesarArchivos(archivos, nil)
	resultado := procesamiento.ConstruirResultado(lectura, clasificacionUmbrales())
	est := estadisticas.Calcular(resultado, estadisticas.Filtro{})
	datos := exportacion.ExportarLibro(est, resultado, "Todos los tratamientos")
	if err := escribirArchivo(salida, datos); err != nil {
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

func ejecutarCaptura(destino, carpeta string) error {
	imagen, err := construirImagen(1000, 700, carpeta)
	if err != nil {
		return err
	}
	archivo, err := os.Create(destino)
	if err != nil {
		return err
	}
	defer archivo.Close()
	return png.Encode(archivo, imagen)
}
