package resumen

import (
	"fmt"
	"strconv"

	"yodo131/nucleo/estadisticas"
)

func numero(valor int) string {
	return strconv.Itoa(valor)
}

func Lineas(est estadisticas.Estadisticas) []string {
	lineas := []string{
		"Tratamientos con yodo 131: " + numero(est.CantidadTratamientos),
		"Pacientes únicos: " + numero(est.CantidadPacientes) + " (" + numero(est.PacientesConVariosTratamientos) + " con más de un tratamiento)",
		"Mujeres: " + numero(est.PacientesMujeres) + " · Hombres: " + numero(est.PacientesHombres) + " · Sexo sin dato: " + numero(est.PacientesSinSexo),
		"Pacientes con cáncer de tiroides: " + numero(est.CantidadPacientesCancer),
		"Registros repetidos unidos: " + numero(est.RegistrosUnidos),
		"Datos que no coinciden (por revisar): " + numero(est.CantidadConflictos),
	}
	if est.EdadPromedio != nil {
		lineas = append(lineas, fmt.Sprintf("Edad promedio: %.1f años (dato en %d de %d)", est.EdadPromedio.(float64), est.CoberturaEdad.ConDato, est.CoberturaEdad.Total))
	}
	if len(est.PorDiagnostico.Filas) > 0 {
		principal := est.PorDiagnostico.Filas[0]
		lineas = append(lineas, "Diagnóstico más frecuente: "+principal.Categoria+" ("+numero(principal.Cantidad)+")")
	}
	return lineas
}
