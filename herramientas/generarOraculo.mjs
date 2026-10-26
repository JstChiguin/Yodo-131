import { readdirSync, readFileSync, writeFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { cargarModulos } from '../pruebas/cargarModulos.mjs';

const raiz = path.join(path.dirname(fileURLToPath(import.meta.url)), '..');
const carpetaEjemplos = path.join(raiz, 'ejemplos');
const carpetaDestino = path.join(raiz, 'nativo', 'nucleo', 'diferencial', 'testdata');

const L = cargarModulos();
const archivos = readdirSync(carpetaEjemplos).sort().filter((nombre) => nombre.endsWith('.docx') || nombre.endsWith('.xlsx')).map((nombre) => ({
  nombre,
  obtenerBuffer: async () => {
    const contenido = readFileSync(path.join(carpetaEjemplos, nombre));
    return contenido.buffer.slice(contenido.byteOffset, contenido.byteOffset + contenido.byteLength);
  }
}));

const lectura = await L.ProcesadorInformes.procesarArchivos(archivos);
const resultado = L.ProcesadorInformes.construirResultado(lectura, L.Clasificador.UMBRALES_POR_DEFECTO);
const estadisticas = L.Estadisticas.calcularEstadisticas(resultado, {});
const { tratamientos: _tratamientos, ...etapa3 } = estadisticas;

const etapa1 = {
  registrosLeidos: lectura.registrosLeidos,
  registrosCancer: lectura.registrosCancer,
  documentosNoLeidos: lectura.documentosNoLeidos,
  conteo: lectura.conteo,
  cantidadArchivos: lectura.cantidadArchivos
};
const etapa2 = {
  tratamientos: resultado.tratamientos,
  pacientes: resultado.pacientes.map((paciente) => ({
    clave: paciente.clave, nombrePaciente: paciente.nombrePaciente, rut: paciente.rut,
    sexo: paciente.sexo, origenSexo: paciente.origenSexo, tieneCancerTiroides: paciente.tieneCancerTiroides,
    idsTratamientos: paciente.tratamientos.map((tratamiento) => tratamiento.idTratamiento)
  })),
  pacientesCancer: resultado.pacientesCancer,
  conflictos: resultado.conflictos,
  registrosUnidos: resultado.registrosUnidos
};

writeFileSync(path.join(carpetaDestino, 'oraculo.json'), JSON.stringify({ etapa1, etapa2, etapa3 }));
const blob = L.ExportadorExcel.exportarLibro(estadisticas, resultado, 'Todos los tratamientos');
writeFileSync(path.join(carpetaDestino, 'oraculo.xlsx'), Buffer.from(await blob.arrayBuffer()));
console.log('Oráculo generado desde el JavaScript con los ejemplos ficticios.');
