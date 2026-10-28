package main

import (
	"os"
	"strconv"
	"strings"
	"sync"

	gioapp "gioui.org/app"
	"gioui.org/op"
	"gioui.org/unit"

	"yodo131/app/vista"
	"yodo131/nucleo/clasificacion"
	"yodo131/nucleo/dialogos"
	"yodo131/nucleo/estadisticas"
	"yodo131/nucleo/exportacion"
	"yodo131/nucleo/modelo"
	"yodo131/nucleo/procesamiento"
	"yodo131/nucleo/seguridad"
)

type aplicacion struct {
	mu       sync.Mutex
	vista    *vista.Vista
	oscuro   bool
	tema     *vista.Tema
	datos    vista.Datos
	umbrales map[string]float64
	lectura  *modelo.Lectura
	carpetas []string
	ventana  *gioapp.Window
}

func nuevaAplicacion() *aplicacion {
	a := &aplicacion{
		vista:    vista.NuevaVista(),
		tema:     vista.NuevoTema(false),
		umbrales: copiarUmbrales(clasificacion.UmbralesPorDefecto),
	}
	a.datos = vista.Datos{
		Mensaje:  "Elija la carpeta con los informes Word y las planillas Excel. El programa solo lee los archivos: no modifica ni borra ninguno.",
		Umbrales: a.umbrales,
	}
	return a
}

func copiarUmbrales(origen map[string]float64) map[string]float64 {
	copia := map[string]float64{}
	for clave, valor := range origen {
		copia[clave] = valor
	}
	return copia
}

func (a *aplicacion) reconstruir() {
	if a.lectura == nil {
		a.datos.TieneDatos = false
		return
	}
	resultado := procesamiento.ConstruirResultado(*a.lectura, a.umbrales)
	est := estadisticas.Calcular(resultado, estadisticas.Filtro{})
	a.datos.Resultado = resultado
	a.datos.Est = est
	a.datos.Umbrales = a.umbrales
	a.datos.TieneDatos = len(resultado.Tratamientos) > 0 || len(resultado.DocumentosNoLeidos) > 0
	a.datos.NotaArchivos = a.notaArchivos()
}

func (a *aplicacion) notaArchivos() string {
	if a.lectura == nil {
		return ""
	}
	return "Cargado: " + strings.Join(a.carpetas, ", ") + " · " + strconv.Itoa(a.lectura.Conteo.ArchivosWord) + " Word y " + strconv.Itoa(a.lectura.Conteo.ArchivosExcel) + " Excel, analizados en conjunto. Para sumar otra carpeta use «Agregar otra carpeta»."
}

func (a *aplicacion) cargarCarpeta(carpeta string) {
	archivos, err := procesamiento.LeerCarpeta(carpeta)
	a.mu.Lock()
	defer a.mu.Unlock()
	if err != nil || len(archivos) == 0 {
		a.datos.Mensaje = "En esa carpeta no hay archivos Word (.docx) ni Excel (.xlsx)."
		return
	}
	nueva := procesamiento.ProcesarArchivos(archivos, nil)
	if a.lectura == nil {
		a.lectura = &nueva
	} else {
		combinada := procesamiento.CombinarLecturas(*a.lectura, nueva)
		a.lectura = &combinada
	}
	a.carpetas = append(a.carpetas, nombreCarpeta(carpeta))
	a.reconstruir()
}

func nombreCarpeta(ruta string) string {
	ruta = strings.TrimRight(ruta, "/\\")
	for i := len(ruta) - 1; i >= 0; i-- {
		if ruta[i] == '/' || ruta[i] == '\\' {
			return ruta[i+1:]
		}
	}
	return ruta
}

func (a *aplicacion) exportar(destino string) {
	a.mu.Lock()
	est := a.datos.Est
	resultado := a.datos.Resultado
	a.mu.Unlock()
	datos := exportacion.ExportarLibro(est, resultado, "Todos los tratamientos")
	os.WriteFile(destino, datos, 0o644)
}

func (a *aplicacion) procesarAcciones(acc vista.Acciones) {
	switch {
	case acc.ElegirCarpeta:
		go func() {
			if carpeta, ok := dialogos.ElegirCarpeta("Elija la carpeta con los informes Word y las planillas Excel"); ok {
				a.cargarCarpeta(carpeta)
				a.invalidar()
			}
		}()
	case acc.Exportar:
		go func() {
			if destino, ok := dialogos.GuardarComo("Guardar el Excel", "Yodo131.xlsx", ""); ok {
				a.exportar(destino)
				a.invalidar()
			}
		}()
	case acc.EmpezarDeNuevo:
		a.mu.Lock()
		a.lectura = nil
		a.carpetas = nil
		a.datos = vista.Datos{Mensaje: "Se quitaron los archivos. Elija una carpeta para comenzar.", Umbrales: a.umbrales}
		a.vista.ClaveValida = false
		a.mu.Unlock()
	case acc.CambiarTema:
		a.mu.Lock()
		a.oscuro = !a.oscuro
		a.tema = vista.NuevoTema(a.oscuro)
		a.vista.Oscuro = a.oscuro
		a.mu.Unlock()
	case acc.IntentarClave:
		if seguridad.VerificarClave(acc.Clave) {
			a.vista.SetClaveValida(true, "")
		} else {
			a.vista.SetClaveValida(false, "Clave incorrecta.")
		}
	case acc.GuardarUmbrales:
		a.mu.Lock()
		for clave, porDefecto := range clasificacion.UmbralesPorDefecto {
			valor := acc.UmbralesNuevos[clave]
			if valor > 0 {
				a.umbrales[clave] = valor
			} else {
				a.umbrales[clave] = porDefecto
			}
		}
		a.reconstruir()
		a.mu.Unlock()
		a.vista.SetMensajeUmbrales("Valores guardados para esta sesión.")
	case acc.RestablecerUmbrales:
		a.mu.Lock()
		a.umbrales = copiarUmbrales(clasificacion.UmbralesPorDefecto)
		a.reconstruir()
		a.mu.Unlock()
		a.vista.SetMensajeUmbrales("Se volvió a los valores iniciales.")
	}
}

func (a *aplicacion) invalidar() {
	a.mu.Lock()
	ventana := a.ventana
	a.mu.Unlock()
	if ventana != nil {
		ventana.Invalidate()
	}
}

func ejecutarInterfaz() error {
	a := nuevaAplicacion()
	ventana := new(gioapp.Window)
	ventana.Option(gioapp.Title("Lector Yodo 131"), gioapp.Size(unit.Dp(1440), unit.Dp(860)), gioapp.MinSize(unit.Dp(1024), unit.Dp(640)))
	a.mu.Lock()
	a.ventana = ventana
	a.mu.Unlock()
	var ops op.Ops
	for {
		switch evento := ventana.Event().(type) {
		case gioapp.DestroyEvent:
			return evento.Err
		case gioapp.FrameEvent:
			gtx := gioapp.NewContext(&ops, evento)
			a.mu.Lock()
			tema := a.tema
			datos := a.datos
			a.mu.Unlock()
			acc := a.vista.Layout(gtx, tema, datos)
			a.procesarAcciones(acc)
			evento.Frame(gtx.Ops)
		}
	}
}
