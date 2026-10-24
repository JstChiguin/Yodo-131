import { readdirSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { cargarModulos } from './cargarModulos.mjs';

const carpetaInformes = process.argv[2];
if (!carpetaInformes) {
  console.error('Indique la carpeta con informes');
  process.exit(1);
}

const LectorYodo = cargarModulos();
const archivos = readdirSync(carpetaInformes).map((nombre) => ({
  nombre,
  obtenerBuffer: async () => {
    const contenido = readFileSync(path.join(carpetaInformes, nombre));
    return contenido.buffer.slice(contenido.byteOffset, contenido.byteOffset + contenido.byteLength);
  }
}));

const resultado = await LectorYodo.ProcesadorInformes.procesarArchivos(archivos, LectorYodo.Clasificador.UMBRALES_POR_DEFECTO);
const formatear = (medicion) => (medicion ? `${medicion.comparador}${medicion.valor}` : '·');
for (const registro of resultado.registros) {
  console.log(`\n== ${registro.archivo}${registro.indiceEnArchivo ? ` #${registro.indiceEnArchivo}` : ''}`);
  console.log({
    nombre: registro.nombrePaciente, rut: registro.rut, edad: registro.edad, fecha: registro.fechaIso || registro.anio,
    sexo: `${registro.sexo} (${registro.origenSexo})`, medico: registro.medicoTratante, dosis: registro.dosisMci, grupoDosis: registro.grupoDosis,
    diagnostico: registro.diagnostico, histologia: registro.histologia, motivo: registro.motivoCategoria,
    tirads: registro.tirads, nodulos: registro.ecoConNodulos, adenopatias: registro.ecoConAdenopatias, bethesda: registro.bethesda, riesgo: registro.riesgo,
    tsh: formatear(registro.tsh), t4l: formatear(registro.t4l), t3: formatear(registro.t3), trab: `${formatear(registro.trab)} → ${registro.trabElevado}`,
    tg: `${registro.tiroglobulina.valorMaximo} asc:${registro.tiroglobulina.enAscenso} → ${registro.tiroglobulinaElevada}`,
    antiTg: `${formatear(registro.antiTiroglobulina)} → ${registro.antiTiroglobulinaElevado}`,
    antecedenteFamiliar: `${registro.textoAntecedentesFamiliares} → ${registro.tieneAntecedenteFamiliar}`,
    fumador: `${registro.textoCigarrillo} → ${registro.esFumador}`,
    rechazo: registro.rechazoYodo, faltantes: registro.camposFaltantes.join(', ')
  });
}
console.log('\nNo leídos:', resultado.documentosNoLeidos);
console.log('Errores:', LectorYodo.RegistroErrores.obtenerErrores());
