import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readdirSync, readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { cargarModulos } from './cargarModulos.mjs';

const LectorYodo = cargarModulos();
const carpetaEjemplos = path.join(path.dirname(fileURLToPath(import.meta.url)), '..', 'ejemplos');

async function procesarEjemplos() {
  const archivos = readdirSync(carpetaEjemplos).map((nombre) => ({
    nombre,
    obtenerBuffer: async () => {
      const contenido = readFileSync(path.join(carpetaEjemplos, nombre));
      return contenido.buffer.slice(contenido.byteOffset, contenido.byteOffset + contenido.byteLength);
    }
  }));
  const lectura = await LectorYodo.ProcesadorInformes.procesarArchivos(archivos);
  return { lectura, resultado: LectorYodo.ProcesadorInformes.construirResultado(lectura, LectorYodo.Clasificador.UMBRALES_POR_DEFECTO) };
}

const tratamientosDe = (resultado, nombre) => resultado.tratamientos.filter((tratamiento) => tratamiento.nombrePaciente === nombre);

test('lee los Excel ficticios con fechas reales y reconoce cada tipo de planilla', async () => {
  const { lectura } = await procesarEjemplos();
  assert.equal(lectura.conteo.archivosExcel, 2);
  assert.equal(lectura.conteo.filasExcelYodo, 10);
  assert.equal(lectura.conteo.filasRegistroCancer, 6);
  const filaExcel = lectura.registrosLeidos.find((registro) => registro.origen === 'Excel' && registro.nombrePaciente === 'Luis Ejemplo Morales');
  assert.equal(filaExcel.fechaIso, '2024-04-10');
  assert.equal(filaExcel.sexoRegistrado, 'M');
  assert.equal(filaExcel.dosisMci, 15);
});

test('una persona tratada en dos fechas distintas cuenta como un paciente y dos tratamientos', async () => {
  const { resultado } = await procesarEjemplos();
  const tratamientos = tratamientosDe(resultado, 'María Ejemplo Rojas');
  assert.equal(tratamientos.length, 2);
  assert.deepEqual(tratamientos.map((tratamiento) => tratamiento.anio), [2024, 2025]);
  assert.equal(new Set(tratamientos.map((tratamiento) => tratamiento.clavePaciente)).size, 1);
  assert.equal(tratamientos[1].numeroTratamientoPaciente, 2);
});

test('dos registros del mismo tratamiento no aumentan las estadísticas', async () => {
  const { resultado } = await procesarEjemplos();
  const [tratamiento] = tratamientosDe(resultado, 'Luis Ejemplo Morales');
  assert.equal(tratamientosDe(resultado, 'Luis Ejemplo Morales').length, 1);
  assert.equal(tratamiento.cantidadFuentes, 2);
  const [conWordYExcel] = tratamientosDe(resultado, 'Jorge Ejemplo Lagos');
  assert.deepEqual([...conWordYExcel.origenes].sort(), ['Excel', 'Word']);
});

test('combina datos complementarios de Word y Excel del mismo tratamiento', async () => {
  const { resultado } = await procesarEjemplos();
  const [tratamiento] = tratamientosDe(resultado, 'Rosa Ejemplo Díaz');
  assert.equal(tratamiento.diagnostico, 'Bocio multinodular o nódulo hiperfuncionante', 'el diagnóstico viene solo del Word');
  assert.equal(tratamiento.tirads, 3, 'la ecografía viene solo del Word');
  assert.equal(tratamiento.fechaIso, '2024-09-18', 'la fecha exacta viene del Excel');
  assert.equal(tratamiento.medicoTratante, 'Dra. Ana Pérez', 'un error de tipeo en el nombre del médico no es una discrepancia');
  assert.deepEqual(tratamiento.conflictos, []);
});

test('una discrepancia entre fuentes queda por revisar y no se decide en silencio', async () => {
  const { resultado } = await procesarEjemplos();
  const [carlos] = tratamientosDe(resultado, 'Carlos Ejemplo Soto');
  assert.equal(carlos.dosisMci, null);
  const conflictoDosis = carlos.conflictos.find((conflicto) => conflicto.campo === 'dosisMci');
  assert.deepEqual(conflictoDosis.valores.map((valor) => valor.valor).sort(), ['100', '150']);

  const [pedro] = tratamientosDe(resultado, 'Pedro Ejemplo Vera');
  assert.equal(pedro.sexo, null, 'el Excel dice F y el nombre indica M: no se elige');
  assert.ok(pedro.conflictos.some((conflicto) => conflicto.campo === 'sexo'));
  assert.ok(resultado.conflictos.some((conflicto) => conflicto.nombrePaciente === 'Pedro Ejemplo Vera' && conflicto.campo === 'sexo'));
});

test('el cáncer de tiroides se cuenta por paciente, en el año de su primer registro', async () => {
  const { resultado } = await procesarEjemplos();
  const porClave = new Map(resultado.pacientesCancer.map((paciente) => [paciente.clave, paciente]));
  assert.equal(resultado.pacientesCancer.length, 6);
  assert.equal(porClave.get('rut:12222222-2').anio, 2024);
  assert.ok(porClave.has('rut:15432109-8'), 'incluye pacientes que solo están en el registro C73');
});

test('las estadísticas separan pacientes, tratamientos y cobertura del dato', async () => {
  const { resultado } = await procesarEjemplos();
  const estadisticas = LectorYodo.Estadisticas.calcularEstadisticas(resultado, {});
  assert.equal(estadisticas.cantidadTratamientos, 12);
  assert.equal(estadisticas.cantidadPacientes, 11);
  assert.equal(estadisticas.pacientesConVariosTratamientos, 1);
  const { porTirads, examenes } = estadisticas;
  assert.equal(porTirads.total, 12);
  assert.ok(porTirads.conDato < porTirads.total);
  assert.equal(porTirads.filas.reduce((suma, fila) => suma + fila.cantidad, 0), porTirads.conDato, 'los porcentajes se calculan solo sobre los registros con dato');
  assert.equal(examenes.TRAb['Sin dato'] + examenes.TRAb.Elevado + examenes.TRAb.Normal, 12);
  assert.ok(examenes.TRAb['Sin dato'] > 0, 'un examen no registrado no se cuenta como normal');
});

test('la distribución informa cobertura con y sin dato', () => {
  const distribucion = LectorYodo.Estadisticas.calcularDistribucion(
    [{ examen: 'Elevado' }, { examen: 'Normal' }, { examen: null }, { examen: null }, { examen: 'Elevado' }],
    (registro) => registro.examen
  );
  assert.deepEqual(distribucion.filas, [{ categoria: 'Elevado', cantidad: 2 }, { categoria: 'Normal', cantidad: 1 }]);
  assert.equal(distribucion.conDato, 3);
  assert.equal(distribucion.sinDato, 2);
  assert.equal(distribucion.total, 5);
});

test('agregar una segunda carpeta da el mismo resultado que leer todo junto', async () => {
  const leerSeleccion = async (filtro) => {
    const archivos = readdirSync(carpetaEjemplos).filter(filtro).map((nombre) => ({
      nombre,
      obtenerBuffer: async () => {
        const contenido = readFileSync(path.join(carpetaEjemplos, nombre));
        return contenido.buffer.slice(contenido.byteOffset, contenido.byteOffset + contenido.byteLength);
      }
    }));
    return LectorYodo.ProcesadorInformes.procesarArchivos(archivos);
  };
  const soloWord = await leerSeleccion((nombre) => nombre.endsWith('.docx'));
  const soloExcel = await leerSeleccion((nombre) => nombre.endsWith('.xlsx'));
  const combinada = LectorYodo.ProcesadorInformes.combinarLecturas(soloWord, soloExcel);
  const { resultado: todoJunto } = await procesarEjemplos();
  const resultadoCombinado = LectorYodo.ProcesadorInformes.construirResultado(combinada, LectorYodo.Clasificador.UMBRALES_POR_DEFECTO);
  assert.equal(resultadoCombinado.tratamientos.length, todoJunto.tratamientos.length);
  assert.equal(resultadoCombinado.pacientes.length, todoJunto.pacientes.length);
  assert.equal(resultadoCombinado.registrosUnidos, todoJunto.registrosUnidos);
  assert.equal(resultadoCombinado.conflictos.length, todoJunto.conflictos.length);
  assert.deepEqual(combinada.conteo, (await procesarEjemplos()).lectura.conteo);
});
