package estadisticas

import (
	"sort"
	"strconv"
	"strings"

	"yodo131/nucleo/modelo"
)

var ordenEdad = []string{"<20", "20–29", "30–39", "40–49", "50–59", "60–69", "70+"}
var ordenDosis = []string{"Baja (≤30 mCi)", "Intermedia (31–99 mCi)", "Alta (≥100 mCi)", "Rechazó / no recibió"}
var ordenBethesda = []string{"I", "II", "III", "IV", "V", "VI"}

type Filtro struct {
	Desde string
	Hasta string
}

type FilaDist struct {
	Categoria string
	Cantidad  int
}

type Distribucion struct {
	Filas   []FilaDist
	ConDato int
	SinDato int
	Total   int
}

type FilaAnio struct {
	Categoria string
	Mujeres   int
	Hombres   int
	SinDato   int
}

type Cobertura struct {
	ConDato int
	Total   int
}

type Estadisticas struct {
	CantidadTratamientos           int
	CantidadSinFecha               int
	CantidadPacientes              int
	PacientesConVariosTratamientos int
	PacientesMujeres               int
	PacientesHombres               int
	PacientesSinSexo               int
	RegistrosUnidos                int
	CantidadConflictos             int
	CantidadPacientesCancer        int
	EdadPromedio                   any
	EdadMinima                     any
	EdadMaxima                     any
	CoberturaEdad                  Cobertura
	PorAnio                        []FilaAnio
	CancerPorAnio                  []FilaAnio
	PorDiagnostico                 Distribucion
	PorMotivo                      Distribucion
	PorEdad                        Distribucion
	PorDosis                       Distribucion
	PorHistologia                  Distribucion
	PorBethesda                    Distribucion
	PorTirads                      Distribucion
	PorMedico                      Distribucion
	PorOrigen                      Distribucion
	Examenes                       map[string]map[string]int
	OrdenExamenes                  []string
	Ecografia                      map[string]map[string]int
	OrdenEcografia                 []string
	Antecedentes                   map[string]map[string]int
	OrdenAntecedentes              []string
	Tratamientos                   []modelo.Objeto
}

func claveMes(anio, mes int) string {
	if anio == 0 {
		return ""
	}
	return strconv.Itoa(anio) + "-" + dosDigitos(mes)
}

func dosDigitos(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

func FiltrarPorPeriodo(tratamientos []modelo.Objeto, filtro Filtro) []modelo.Objeto {
	if filtro.Desde == "" && filtro.Hasta == "" {
		return tratamientos
	}
	filtrados := []modelo.Objeto{}
	for _, tratamiento := range tratamientos {
		anio, ok := tratamiento.Entero("anio")
		if !ok {
			continue
		}
		mes := 1
		if m, ok := tratamiento.Entero("mes"); ok {
			mes = m
		}
		desdeMes := claveMes(anio, mes)
		mesHasta := 12
		if m, ok := tratamiento.Entero("mes"); ok {
			mesHasta = m
		}
		hastaMes := claveMes(anio, mesHasta)
		if filtro.Desde != "" && hastaMes < filtro.Desde {
			continue
		}
		if filtro.Hasta != "" && desdeMes > filtro.Hasta {
			continue
		}
		filtrados = append(filtrados, tratamiento)
	}
	return filtrados
}

func CalcularDistribucion(registros []modelo.Objeto, categoria func(modelo.Objeto) any, ordenFijo []string) Distribucion {
	orden := []string{}
	conteo := map[string]int{}
	for _, registro := range registros {
		cat := categoria(registro)
		if cat == nil {
			continue
		}
		clave := cat.(string)
		if _, existe := conteo[clave]; !existe {
			orden = append(orden, clave)
		}
		conteo[clave]++
	}
	filas := make([]FilaDist, len(orden))
	for indice, clave := range orden {
		filas[indice] = FilaDist{Categoria: clave, Cantidad: conteo[clave]}
	}
	if ordenFijo != nil {
		indiceEn := func(categoria string) int {
			for i, c := range ordenFijo {
				if c == categoria {
					return i
				}
			}
			return -1
		}
		sort.SliceStable(filas, func(i, j int) bool {
			return indiceEn(filas[i].Categoria) < indiceEn(filas[j].Categoria)
		})
	} else {
		sort.SliceStable(filas, func(i, j int) bool {
			return filas[i].Cantidad > filas[j].Cantidad
		})
	}
	conDato := 0
	for _, fila := range filas {
		conDato += fila.Cantidad
	}
	return Distribucion{Filas: filas, ConDato: conDato, SinDato: len(registros) - conDato, Total: len(registros)}
}

func contarEstadoExamen(registros []modelo.Objeto, campo string) map[string]int {
	conteo := map[string]int{"Elevado": 0, "Normal": 0, "Sin dato": 0}
	for _, registro := range registros {
		switch v := registro.Get(campo).(type) {
		case bool:
			if v {
				conteo["Elevado"]++
			} else {
				conteo["Normal"]++
			}
		default:
			conteo["Sin dato"]++
		}
	}
	return conteo
}

func contarSiNo(registros []modelo.Objeto, campo string) map[string]int {
	conteo := map[string]int{"Sí": 0, "No": 0, "Sin dato": 0}
	for _, registro := range registros {
		switch v := registro.Get(campo).(type) {
		case bool:
			if v {
				conteo["Sí"]++
			} else {
				conteo["No"]++
			}
		default:
			conteo["Sin dato"]++
		}
	}
	return conteo
}

func esAnio(s string) bool {
	if len(s) != 4 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func ordenarAnios(filas []FilaAnio) {
	sort.SliceStable(filas, func(i, j int) bool {
		a, b := filas[i].Categoria, filas[j].Categoria
		anioA, anioB := esAnio(a), esAnio(b)
		if anioA && anioB {
			na, _ := strconv.Atoi(a)
			nb, _ := strconv.Atoi(b)
			return na < nb
		}
		if anioA != anioB {
			return anioA
		}
		return strings.Compare(a, b) < 0
	})
}

func contarPorAnioYSexo(items []struct {
	anio any
	sexo any
}) []FilaAnio {
	orden := []string{}
	porAnio := map[string]*FilaAnio{}
	for _, item := range items {
		clave := "Sin fecha"
		if anio, ok := item.anio.(int); ok && anio != 0 {
			clave = strconv.Itoa(anio)
		}
		fila, existe := porAnio[clave]
		if !existe {
			fila = &FilaAnio{Categoria: clave}
			porAnio[clave] = fila
			orden = append(orden, clave)
		}
		if item.sexo == "F" {
			fila.Mujeres++
		} else if item.sexo == "M" {
			fila.Hombres++
		} else {
			fila.SinDato++
		}
	}
	filas := make([]FilaAnio, len(orden))
	for indice, clave := range orden {
		filas[indice] = *porAnio[clave]
	}
	ordenarAnios(filas)
	return filas
}

func filtrarPorAnios(pacientes []modelo.PacienteCancer, filtro Filtro) []modelo.PacienteCancer {
	var anioDesde, anioHasta int
	if filtro.Desde != "" {
		anioDesde, _ = strconv.Atoi(filtro.Desde[:4])
	}
	if filtro.Hasta != "" {
		anioHasta, _ = strconv.Atoi(filtro.Hasta[:4])
	}
	if anioDesde == 0 && anioHasta == 0 {
		return pacientes
	}
	filtrados := []modelo.PacienteCancer{}
	for _, paciente := range pacientes {
		anio, ok := paciente.Anio.(int)
		if !ok {
			continue
		}
		if (anioDesde == 0 || anio >= anioDesde) && (anioHasta == 0 || anio <= anioHasta) {
			filtrados = append(filtrados, paciente)
		}
	}
	return filtrados
}

func Calcular(resultado modelo.Resultado, filtro Filtro) Estadisticas {
	tratamientos := FiltrarPorPeriodo(resultado.Tratamientos, filtro)

	ordenPacientes := []string{}
	porPaciente := map[string][]modelo.Objeto{}
	for _, tratamiento := range tratamientos {
		clave := tratamiento.Texto("clavePaciente")
		if _, existe := porPaciente[clave]; !existe {
			ordenPacientes = append(ordenPacientes, clave)
		}
		porPaciente[clave] = append(porPaciente[clave], tratamiento)
	}
	primerosPaciente := make([]modelo.Objeto, len(ordenPacientes))
	variosTratamientos := 0
	for indice, clave := range ordenPacientes {
		primerosPaciente[indice] = porPaciente[clave][0]
		if len(porPaciente[clave]) > 1 {
			variosTratamientos++
		}
	}

	tratamientosCancer := []modelo.Objeto{}
	for _, tratamiento := range tratamientos {
		if tratamiento.Get("diagnostico") == "Cáncer de tiroides" {
			tratamientosCancer = append(tratamientosCancer, tratamiento)
		}
	}
	pacientesCancer := filtrarPorAnios(resultado.PacientesCancer, filtro)

	edades := []int{}
	for _, tratamiento := range tratamientos {
		if edad, ok := tratamiento.Entero("edad"); ok {
			edades = append(edades, edad)
		}
	}
	idsEnPeriodo := map[string]bool{}
	for _, tratamiento := range tratamientos {
		idsEnPeriodo[tratamiento.Texto("idTratamiento")] = true
	}

	contarSexo := func(valor string) int {
		total := 0
		for _, paciente := range primerosPaciente {
			if paciente.Get("sexo") == valor {
				total++
			}
		}
		return total
	}
	sinSexo := 0
	for _, paciente := range primerosPaciente {
		if paciente.Get("sexo") == nil {
			sinSexo++
		}
	}

	registrosUnidos := 0
	for _, tratamiento := range tratamientos {
		cantidad, _ := tratamiento.Entero("cantidadFuentes")
		registrosUnidos += cantidad - 1
	}
	conflictos := 0
	for _, conflicto := range resultado.Conflictos {
		if idsEnPeriodo[conflicto.IdTratamiento] {
			conflictos++
		}
	}
	sinFecha := 0
	for _, tratamiento := range resultado.Tratamientos {
		if tratamiento.Get("anio") == nil {
			sinFecha++
		}
	}

	var edadPromedio, edadMinima, edadMaxima any
	if len(edades) > 0 {
		suma := 0
		minimo := edades[0]
		maximo := edades[0]
		for _, edad := range edades {
			suma += edad
			if edad < minimo {
				minimo = edad
			}
			if edad > maximo {
				maximo = edad
			}
		}
		edadPromedio = float64(suma) / float64(len(edades))
		edadMinima = minimo
		edadMaxima = maximo
	}

	aniosSexoTratamientos := make([]struct {
		anio any
		sexo any
	}, len(tratamientos))
	for indice, tratamiento := range tratamientos {
		aniosSexoTratamientos[indice] = struct {
			anio any
			sexo any
		}{anio: tratamiento.Get("anio"), sexo: tratamiento.Get("sexo")}
	}
	aniosSexoCancer := make([]struct {
		anio any
		sexo any
	}, len(pacientesCancer))
	for indice, paciente := range pacientesCancer {
		aniosSexoCancer[indice] = struct {
			anio any
			sexo any
		}{anio: paciente.Anio, sexo: paciente.Sexo}
	}

	return Estadisticas{
		CantidadTratamientos:           len(tratamientos),
		CantidadSinFecha:               sinFecha,
		CantidadPacientes:              len(primerosPaciente),
		PacientesConVariosTratamientos: variosTratamientos,
		PacientesMujeres:               contarSexo("F"),
		PacientesHombres:               contarSexo("M"),
		PacientesSinSexo:               sinSexo,
		RegistrosUnidos:                registrosUnidos,
		CantidadConflictos:             conflictos,
		CantidadPacientesCancer:        len(pacientesCancer),
		EdadPromedio:                   edadPromedio,
		EdadMinima:                     edadMinima,
		EdadMaxima:                     edadMaxima,
		CoberturaEdad:                  Cobertura{ConDato: len(edades), Total: len(tratamientos)},
		PorAnio:                        contarPorAnioYSexo(aniosSexoTratamientos),
		CancerPorAnio:                  contarPorAnioYSexo(aniosSexoCancer),
		PorDiagnostico:                 CalcularDistribucion(tratamientos, func(t modelo.Objeto) any { return t.Get("diagnostico") }, nil),
		PorMotivo:                      CalcularDistribucion(tratamientos, func(t modelo.Objeto) any { return t.Get("motivoCategoria") }, nil),
		PorEdad:                        CalcularDistribucion(tratamientos, func(t modelo.Objeto) any { return t.Get("grupoEdad") }, ordenEdad),
		PorDosis:                       CalcularDistribucion(tratamientos, func(t modelo.Objeto) any { return t.Get("grupoDosis") }, ordenDosis),
		PorHistologia: CalcularDistribucion(tratamientosCancer, func(t modelo.Objeto) any {
			if t.Get("histologia") == "No especificado" {
				return nil
			}
			return t.Get("histologia")
		}, nil),
		PorBethesda: CalcularDistribucion(tratamientos, func(t modelo.Objeto) any { return t.Get("bethesda") }, ordenBethesda),
		PorTirads: CalcularDistribucion(tratamientos, func(t modelo.Objeto) any {
			if tirads, ok := t.Entero("tirads"); ok {
				return "TIRADS " + strconv.Itoa(tirads)
			}
			return nil
		}, []string{"TIRADS 1", "TIRADS 2", "TIRADS 3", "TIRADS 4", "TIRADS 5", "TIRADS 6"}),
		PorMedico: CalcularDistribucion(tratamientos, func(t modelo.Objeto) any { return t.Get("medicoTratante") }, nil),
		PorOrigen: CalcularDistribucion(tratamientos, func(t modelo.Objeto) any {
			origenes, _ := t.Get("origenes").([]string)
			if len(origenes) > 1 {
				return "Word y Excel"
			}
			if len(origenes) == 1 {
				return origenes[0]
			}
			return nil
		}, nil),
		Examenes: map[string]map[string]int{
			"TRAb":               contarEstadoExamen(tratamientos, "trabElevado"),
			"Anti-tiroglobulina": contarEstadoExamen(tratamientos, "antiTiroglobulinaElevado"),
			"Anti-TPO":           contarEstadoExamen(tratamientos, "antiTpoElevado"),
			"Tiroglobulina":      contarEstadoExamen(tratamientos, "tiroglobulinaElevada"),
		},
		OrdenExamenes: []string{"TRAb", "Anti-tiroglobulina", "Anti-TPO", "Tiroglobulina"},
		Ecografia: map[string]map[string]int{
			"Con nódulos":     contarSiNo(tratamientos, "ecoConNodulos"),
			"Con adenopatías": contarSiNo(tratamientos, "ecoConAdenopatias"),
		},
		OrdenEcografia: []string{"Con nódulos", "Con adenopatías"},
		Antecedentes: map[string]map[string]int{
			"Antecedente familiar tiroideo": contarSiNo(tratamientos, "tieneAntecedenteFamiliar"),
			"Tabaquismo actual o previo":    contarSiNo(tratamientos, "esFumador"),
		},
		OrdenAntecedentes: []string{"Antecedente familiar tiroideo", "Tabaquismo actual o previo"},
		Tratamientos:      tratamientos,
	}
}
