package diferencial

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"yodo131/nucleo/clasificacion"
	"yodo131/nucleo/estadisticas"
	"yodo131/nucleo/exportacion"
	"yodo131/nucleo/modelo"
	"yodo131/nucleo/procesamiento"
)

const carpetaEjemplos = "../../../ejemplos"
const descripcionPeriodo = "Todos los tratamientos"

func convertir(valor any) any {
	switch v := valor.(type) {
	case modelo.Medicion:
		return map[string]any{"valor": v.Valor, "comparador": v.Comparador}
	case modelo.Tiroglobulina:
		return map[string]any{"valorMaximo": v.ValorMaximo, "enAscenso": v.EnAscenso, "mencionada": v.Mencionada}
	case modelo.Fecha:
		return map[string]any{"fecha": v.Iso}
	default:
		return v
	}
}

func serializarRegistro(o modelo.Objeto) map[string]any {
	salida := map[string]any{}
	for clave, valor := range o.D {
		salida[clave] = convertir(valor)
	}
	return salida
}

func serializarTratamiento(o modelo.Objeto) map[string]any {
	salida := map[string]any{}
	for clave, valor := range o.D {
		switch clave {
		case "fuentes":
			arr := []any{}
			for _, fuente := range valor.([]modelo.Fuente) {
				arr = append(arr, map[string]any{"origen": fuente.Origen, "referencia": fuente.Referencia, "fechaIso": fuente.FechaIso, "anio": fuente.Anio})
			}
			salida[clave] = arr
		case "conflictos":
			salida[clave] = serializarConflictos(valor.([]modelo.Conflicto))
		case "tiroglobulina":
			tg := valor.(modelo.Tiroglobulina)
			salida[clave] = map[string]any{"valorMaximo": tg.ValorMaximo, "enAscenso": tg.EnAscenso}
		default:
			salida[clave] = convertir(valor)
		}
	}
	return salida
}

func serializarConflictos(conflictos []modelo.Conflicto) []any {
	arr := []any{}
	for _, conflicto := range conflictos {
		valores := []any{}
		for _, valor := range conflicto.Valores {
			valores = append(valores, map[string]any{"origen": valor.Origen, "referencia": valor.Referencia, "valor": valor.Valor})
		}
		obj := map[string]any{"campo": conflicto.Campo, "nombreCampo": conflicto.NombreCampo, "valores": valores}
		if conflicto.IdTratamiento != "" {
			obj["idTratamiento"] = conflicto.IdTratamiento
			obj["nombrePaciente"] = conflicto.NombrePaciente
			obj["rut"] = conflicto.Rut
			obj["fechaIso"] = conflicto.FechaIso
			obj["anio"] = conflicto.Anio
		}
		arr = append(arr, obj)
	}
	return arr
}

func serializarDocumentos(documentos []modelo.DocumentoNoLeido) []any {
	arr := []any{}
	for _, documento := range documentos {
		arr = append(arr, map[string]any{"archivo": documento.Archivo, "motivo": documento.Motivo})
	}
	return arr
}

func serializarDistribucion(d estadisticas.Distribucion) map[string]any {
	filas := []any{}
	for _, fila := range d.Filas {
		filas = append(filas, map[string]any{"categoria": fila.Categoria, "cantidad": fila.Cantidad})
	}
	return map[string]any{"filas": filas, "conDato": d.ConDato, "sinDato": d.SinDato, "total": d.Total}
}

func serializarAnios(filas []estadisticas.FilaAnio) []any {
	arr := []any{}
	for _, fila := range filas {
		arr = append(arr, map[string]any{"categoria": fila.Categoria, "mujeres": fila.Mujeres, "hombres": fila.Hombres, "sinDato": fila.SinDato})
	}
	return arr
}

func serializarGrupo(grupo map[string]map[string]int) map[string]any {
	salida := map[string]any{}
	for nombre, conteo := range grupo {
		interno := map[string]any{}
		for categoria, cantidad := range conteo {
			interno[categoria] = cantidad
		}
		salida[nombre] = interno
	}
	return salida
}

func cargarEjemplos(t *testing.T) []procesamiento.Archivo {
	entradas, err := os.ReadDir(carpetaEjemplos)
	if err != nil {
		t.Fatalf("no se pudieron leer los ejemplos: %v", err)
	}
	nombres := []string{}
	for _, entrada := range entradas {
		nombres = append(nombres, entrada.Name())
	}
	sort.Strings(nombres)
	archivos := []procesamiento.Archivo{}
	for _, nombre := range nombres {
		if !strings.HasSuffix(nombre, ".docx") && !strings.HasSuffix(nombre, ".xlsx") {
			continue
		}
		datos, err := os.ReadFile(filepath.Join(carpetaEjemplos, nombre))
		if err != nil {
			t.Fatalf("no se pudo leer %s: %v", nombre, err)
		}
		archivos = append(archivos, procesamiento.Archivo{Nombre: nombre, Datos: datos})
	}
	return archivos
}

func construirSalida(t *testing.T) (map[string]any, []byte) {
	lectura := procesamiento.ProcesarArchivos(cargarEjemplos(t), nil)
	resultado := procesamiento.ConstruirResultado(lectura, clasificacion.UmbralesPorDefecto)
	est := estadisticas.Calcular(resultado, estadisticas.Filtro{})

	registros := []any{}
	for _, registro := range lectura.RegistrosLeidos {
		registros = append(registros, serializarRegistro(registro))
	}
	cancer := []any{}
	for _, registro := range lectura.RegistrosCancer {
		cancer = append(cancer, serializarRegistro(registro))
	}
	etapa1 := map[string]any{
		"registrosLeidos": registros, "registrosCancer": cancer,
		"documentosNoLeidos": serializarDocumentos(lectura.DocumentosNoLeidos),
		"conteo": map[string]any{
			"archivosWord": lectura.Conteo.ArchivosWord, "archivosExcel": lectura.Conteo.ArchivosExcel,
			"informesWord": lectura.Conteo.InformesWord, "filasExcelYodo": lectura.Conteo.FilasExcelYodo,
			"filasRegistroCancer": lectura.Conteo.FilasRegistroCancer,
		},
		"cantidadArchivos": lectura.CantidadArchivos,
	}

	tratamientos := []any{}
	for _, tratamiento := range resultado.Tratamientos {
		tratamientos = append(tratamientos, serializarTratamiento(tratamiento))
	}
	pacientes := []any{}
	for _, paciente := range resultado.Pacientes {
		ids := []any{}
		for _, tratamiento := range paciente.Tratamientos {
			ids = append(ids, tratamiento.Get("idTratamiento"))
		}
		pacientes = append(pacientes, map[string]any{
			"clave": paciente.Clave, "nombrePaciente": paciente.NombrePaciente, "rut": paciente.Rut,
			"sexo": paciente.Sexo, "origenSexo": paciente.OrigenSexo, "tieneCancerTiroides": paciente.TieneCancerTiroides,
			"idsTratamientos": ids,
		})
	}
	pacientesCancer := []any{}
	for _, paciente := range resultado.PacientesCancer {
		obj := map[string]any{"clave": paciente.Clave, "anio": paciente.Anio, "sexo": paciente.Sexo}
		if paciente.EnRegistro {
			obj["enRegistro"] = true
		}
		if paciente.EnInformes {
			obj["enInformes"] = true
		}
		pacientesCancer = append(pacientesCancer, obj)
	}
	etapa2 := map[string]any{
		"tratamientos": tratamientos, "pacientes": pacientes, "pacientesCancer": pacientesCancer,
		"conflictos": serializarConflictos(resultado.Conflictos), "registrosUnidos": resultado.RegistrosUnidos,
	}

	etapa3 := map[string]any{
		"cantidadTratamientos": est.CantidadTratamientos, "cantidadSinFecha": est.CantidadSinFecha,
		"cantidadPacientes": est.CantidadPacientes, "pacientesConVariosTratamientos": est.PacientesConVariosTratamientos,
		"pacientesMujeres": est.PacientesMujeres, "pacientesHombres": est.PacientesHombres, "pacientesSinSexo": est.PacientesSinSexo,
		"registrosUnidos": est.RegistrosUnidos, "cantidadConflictos": est.CantidadConflictos, "cantidadPacientesCancer": est.CantidadPacientesCancer,
		"edadPromedio": est.EdadPromedio, "edadMinima": est.EdadMinima, "edadMaxima": est.EdadMaxima,
		"coberturaEdad": map[string]any{"conDato": est.CoberturaEdad.ConDato, "total": est.CoberturaEdad.Total},
		"porAnio": serializarAnios(est.PorAnio), "cancerPorAnio": serializarAnios(est.CancerPorAnio),
		"porDiagnostico": serializarDistribucion(est.PorDiagnostico), "porMotivo": serializarDistribucion(est.PorMotivo),
		"porEdad": serializarDistribucion(est.PorEdad), "porDosis": serializarDistribucion(est.PorDosis),
		"porHistologia": serializarDistribucion(est.PorHistologia), "porBethesda": serializarDistribucion(est.PorBethesda),
		"porTirads": serializarDistribucion(est.PorTirads), "porMedico": serializarDistribucion(est.PorMedico),
		"porOrigen": serializarDistribucion(est.PorOrigen),
		"examenes": serializarGrupo(est.Examenes), "ecografia": serializarGrupo(est.Ecografia), "antecedentes": serializarGrupo(est.Antecedentes),
	}

	salida := map[string]any{"etapa1": etapa1, "etapa2": etapa2, "etapa3": etapa3}
	xlsx := exportacion.ExportarLibro(est, resultado, descripcionPeriodo)
	return salida, xlsx
}

func normalizar(valor any) any {
	datos, _ := json.Marshal(valor)
	var normalizado any
	json.Unmarshal(datos, &normalizado)
	return normalizado
}

func comparar(esperado, obtenido any, ruta string) (string, bool) {
	switch esperadoTipo := esperado.(type) {
	case map[string]any:
		obtenidoMapa, ok := obtenido.(map[string]any)
		if !ok {
			return ruta + ": se esperaba un objeto", false
		}
		claves := map[string]bool{}
		for clave := range esperadoTipo {
			claves[clave] = true
		}
		for clave := range obtenidoMapa {
			claves[clave] = true
		}
		ordenadas := []string{}
		for clave := range claves {
			ordenadas = append(ordenadas, clave)
		}
		sort.Strings(ordenadas)
		for _, clave := range ordenadas {
			valorEsperado, existeEsperado := esperadoTipo[clave]
			valorObtenido, existeObtenido := obtenidoMapa[clave]
			if existeEsperado != existeObtenido {
				return fmt.Sprintf("%s/%s: la clave está en %s pero no en %s", ruta, clave, dondeExiste(existeEsperado), dondeExiste(existeObtenido)), false
			}
			if mensaje, igual := comparar(valorEsperado, valorObtenido, ruta+"/"+clave); !igual {
				return mensaje, false
			}
		}
		return "", true
	case []any:
		obtenidoLista, ok := obtenido.([]any)
		if !ok {
			return ruta + ": se esperaba una lista", false
		}
		if len(esperadoTipo) != len(obtenidoLista) {
			return fmt.Sprintf("%s: largo distinto (JS %d, Go %d)", ruta, len(esperadoTipo), len(obtenidoLista)), false
		}
		for indice := range esperadoTipo {
			if mensaje, igual := comparar(esperadoTipo[indice], obtenidoLista[indice], fmt.Sprintf("%s/%d", ruta, indice)); !igual {
				return mensaje, false
			}
		}
		return "", true
	default:
		if esperado != obtenido {
			return fmt.Sprintf("%s: JS=%v Go=%v", ruta, esperado, obtenido), false
		}
		return "", true
	}
}

func dondeExiste(existe bool) string {
	if existe {
		return "JS"
	}
	return "Go"
}

func TestLogicaGoCoincideConJavaScript(t *testing.T) {
	salida, _ := construirSalida(t)
	golden, err := os.ReadFile("testdata/oraculo.json")
	if err != nil {
		t.Fatalf("no se pudo leer el oráculo: %v. Genérelo con: node herramientas/generarOraculo.mjs", err)
	}
	var esperado any
	if err := json.Unmarshal(golden, &esperado); err != nil {
		t.Fatalf("oráculo inválido: %v", err)
	}
	obtenido := normalizar(salida)
	if mensaje, igual := comparar(esperado, obtenido, "resultado"); !igual {
		t.Fatalf("la lógica Go difiere del JavaScript en %s", mensaje)
	}
}

func TestExcelExportadoCoincideConJavaScript(t *testing.T) {
	_, xlsx := construirSalida(t)
	golden, err := os.ReadFile("testdata/oraculo.xlsx")
	if err != nil {
		t.Fatalf("no se pudo leer el Excel del oráculo: %v", err)
	}
	if !bytes.Equal(xlsx, golden) {
		t.Fatalf("el Excel exportado por Go difiere del exportado por JavaScript (Go %d bytes, JS %d bytes)", len(xlsx), len(golden))
	}
}
