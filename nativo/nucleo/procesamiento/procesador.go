package procesamiento

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"yodo131/nucleo/clasificacion"
	"yodo131/nucleo/extraccion"
	"yodo131/nucleo/fusion"
	"yodo131/nucleo/lectura"
	"yodo131/nucleo/modelo"
)

type Archivo struct {
	Nombre string
	Datos  []byte
}

var (
	extensionesWord  = regexp.MustCompile(`(?i)\.(docx|doc)$`)
	extensionesExcel = regexp.MustCompile(`(?i)\.(xlsx|xls)$`)
)

func nombreBase(nombre string) string {
	partes := strings.Split(nombre, "/")
	return partes[len(partes)-1]
}

func esArchivoTemporal(nombre string) bool {
	return strings.HasPrefix(nombreBase(nombre), "~$")
}

func esArchivoWordValido(nombre string) bool {
	return extensionesWord.MatchString(nombre) && !esArchivoTemporal(nombre)
}

func esArchivoExcelValido(nombre string) bool {
	return extensionesExcel.MatchString(nombre) && !esArchivoTemporal(nombre)
}

func EsArchivoSoportado(nombre string) bool {
	return esArchivoWordValido(nombre) || esArchivoExcelValido(nombre)
}

func leerArchivoWord(archivo Archivo) ([]modelo.Objeto, error) {
	parrafos, err := lectura.LeerParrafosDocx(archivo.Datos)
	if err != nil {
		return nil, err
	}
	bloques := extraccion.DividirEnInformes(parrafos)
	informes := make([]modelo.Objeto, 0, len(bloques))
	for indice, lineas := range bloques {
		informe := extraccion.ExtraerInforme(lineas, archivo.Nombre)
		if len(bloques) > 1 {
			numero := indice + 1
			informe.Set("indiceEnArchivo", numero)
			informe.Set("referencia", archivo.Nombre+" › informe "+strconv.Itoa(numero))
		} else {
			informe.Set("indiceEnArchivo", nil)
		}
		informes = append(informes, informe)
	}
	return informes, nil
}

func convertirHojas(hojas []lectura.Hoja) []extraccion.HojaLeida {
	convertidas := make([]extraccion.HojaLeida, len(hojas))
	for indice, hoja := range hojas {
		convertidas[indice] = extraccion.HojaLeida{Nombre: hoja.Nombre, Filas: hoja.Filas}
	}
	return convertidas
}

func ProcesarArchivos(archivos []Archivo, alAvanzar func(int, int)) modelo.Lectura {
	lect := modelo.Lectura{
		RegistrosLeidos:    []modelo.Objeto{},
		RegistrosCancer:    []modelo.Objeto{},
		DocumentosNoLeidos: []modelo.DocumentoNoLeido{},
	}
	soportados := []Archivo{}
	for _, archivo := range archivos {
		if EsArchivoSoportado(archivo.Nombre) {
			soportados = append(soportados, archivo)
		}
	}
	for indice, archivo := range soportados {
		var formato *lectura.FormatoNoSoportadoError
		if esArchivoWordValido(archivo.Nombre) {
			lect.Conteo.ArchivosWord++
			informes, err := leerArchivoWord(archivo)
			if err != nil {
				if errors.As(err, &formato) {
					lect.DocumentosNoLeidos = append(lect.DocumentosNoLeidos, modelo.DocumentoNoLeido{Archivo: archivo.Nombre, Motivo: formato.Mensaje})
				} else {
					lect.DocumentosNoLeidos = append(lect.DocumentosNoLeidos, modelo.DocumentoNoLeido{Archivo: archivo.Nombre, Motivo: "No se pudo leer el archivo"})
				}
			} else {
				if len(informes) == 0 {
					lect.DocumentosNoLeidos = append(lect.DocumentosNoLeidos, modelo.DocumentoNoLeido{Archivo: archivo.Nombre, Motivo: `No tiene el formato de informe (falta la línea "NOMBRE:")`})
				}
				lect.Conteo.InformesWord += len(informes)
				lect.RegistrosLeidos = append(lect.RegistrosLeidos, informes...)
			}
		} else {
			lect.Conteo.ArchivosExcel++
			hojas, err := lectura.LeerLibro(archivo.Datos)
			if err != nil {
				if errors.As(err, &formato) {
					lect.DocumentosNoLeidos = append(lect.DocumentosNoLeidos, modelo.DocumentoNoLeido{Archivo: archivo.Nombre, Motivo: formato.Mensaje})
				} else {
					lect.DocumentosNoLeidos = append(lect.DocumentosNoLeidos, modelo.DocumentoNoLeido{Archivo: archivo.Nombre, Motivo: "No se pudo leer el archivo"})
				}
			} else {
				resultado := extraccion.ExtraerLibro(convertirHojas(hojas), archivo.Nombre)
				lect.Conteo.FilasExcelYodo += len(resultado.RegistrosYodo)
				lect.Conteo.FilasRegistroCancer += len(resultado.RegistrosCancer)
				lect.RegistrosLeidos = append(lect.RegistrosLeidos, resultado.RegistrosYodo...)
				lect.RegistrosCancer = append(lect.RegistrosCancer, resultado.RegistrosCancer...)
				lect.DocumentosNoLeidos = append(lect.DocumentosNoLeidos, resultado.HojasOmitidas...)
			}
		}
		if alAvanzar != nil {
			alAvanzar(indice+1, len(soportados))
		}
	}
	lect.CantidadArchivos = len(soportados)
	return lect
}

func CombinarLecturas(anterior, nueva modelo.Lectura) modelo.Lectura {
	return modelo.Lectura{
		RegistrosLeidos:    append(append([]modelo.Objeto{}, anterior.RegistrosLeidos...), nueva.RegistrosLeidos...),
		RegistrosCancer:    append(append([]modelo.Objeto{}, anterior.RegistrosCancer...), nueva.RegistrosCancer...),
		DocumentosNoLeidos: append(append([]modelo.DocumentoNoLeido{}, anterior.DocumentosNoLeidos...), nueva.DocumentosNoLeidos...),
		Conteo: modelo.Conteo{
			ArchivosWord:        anterior.Conteo.ArchivosWord + nueva.Conteo.ArchivosWord,
			ArchivosExcel:       anterior.Conteo.ArchivosExcel + nueva.Conteo.ArchivosExcel,
			InformesWord:        anterior.Conteo.InformesWord + nueva.Conteo.InformesWord,
			FilasExcelYodo:      anterior.Conteo.FilasExcelYodo + nueva.Conteo.FilasExcelYodo,
			FilasRegistroCancer: anterior.Conteo.FilasRegistroCancer + nueva.Conteo.FilasRegistroCancer,
		},
		CantidadArchivos: anterior.CantidadArchivos + nueva.CantidadArchivos,
	}
}

func completarTratamiento(tratamiento modelo.Objeto) {
	tratamiento.Set("grupoEdad", clasificacion.AgruparEdad(tratamiento.Get("edad")))
	if tratamiento.Bool("rechazoYodo") && tratamiento.Get("dosisMci") == nil {
		tratamiento.Set("grupoDosis", "Rechazó / no recibió")
	} else {
		tratamiento.Set("grupoDosis", clasificacion.AgruparDosis(tratamiento.Get("dosisMci")))
	}
	camposEnConflicto := map[string]bool{}
	if conflictos, ok := tratamiento.Get("conflictos").([]modelo.Conflicto); ok {
		for _, conflicto := range conflictos {
			camposEnConflicto[conflicto.Campo] = true
		}
	}
	equivalencias := map[string]string{"RUT": "rut", "edad": "edad", "fecha": "fecha", "médico": "medicoTratante", "dosis": "dosisMci", "sexo": "sexo"}
	faltantes := []string{}
	for _, campo := range clasificacion.ListarCamposFaltantes(tratamiento) {
		if !camposEnConflicto[equivalencias[campo]] {
			faltantes = append(faltantes, campo)
		}
	}
	tratamiento.Set("camposFaltantes", faltantes)
}

func ConstruirResultado(lect modelo.Lectura, umbrales map[string]float64) modelo.Resultado {
	clasificados := make([]modelo.Objeto, len(lect.RegistrosLeidos))
	for indice, registro := range lect.RegistrosLeidos {
		clasificados[indice] = clasificacion.ClasificarInforme(registro, umbrales)
	}
	fus := fusion.Fusionar(clasificados, lect.RegistrosCancer)
	medicosUnificados := clasificacion.UnificarMedicos(fus.Tratamientos)
	for indice, tratamiento := range fus.Tratamientos {
		tratamiento.Set("medicoTratante", medicosUnificados[indice].Get("medicoTratante"))
		completarTratamiento(tratamiento)
	}
	return modelo.Resultado{
		Pacientes:          fus.Pacientes,
		Tratamientos:       fus.Tratamientos,
		PacientesCancer:    fus.PacientesCancer,
		Conflictos:         fus.Conflictos,
		RegistrosUnidos:    fus.RegistrosUnidos,
		DocumentosNoLeidos: lect.DocumentosNoLeidos,
		Conteo:             lect.Conteo,
		CantidadArchivos:   lect.CantidadArchivos,
	}
}

func LeerCarpeta(ruta string) ([]Archivo, error) {
	archivos := []Archivo{}
	raiz := filepath.Clean(ruta)
	err := filepath.WalkDir(raiz, func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entrada.IsDir() {
			return nil
		}
		if !EsArchivoSoportado(entrada.Name()) {
			return nil
		}
		datos, err := os.ReadFile(ruta)
		if err != nil {
			return err
		}
		relativa, err := filepath.Rel(filepath.Dir(raiz), ruta)
		if err != nil {
			relativa = entrada.Name()
		}
		archivos = append(archivos, Archivo{Nombre: filepath.ToSlash(relativa), Datos: datos})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(archivos, func(i, j int) bool { return archivos[i].Nombre < archivos[j].Nombre })
	return archivos, nil
}
