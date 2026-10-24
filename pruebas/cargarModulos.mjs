import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import vm from 'node:vm';

const carpetaJs = path.join(path.dirname(fileURLToPath(import.meta.url)), '..', 'Sistema - no tocar', 'js');
const MODULOS = ['version.js', 'registroErrores.js', 'lectorZip.js', 'lectorDocx.js', 'lectorXlsx.js', 'extractorInforme.js', 'extractorExcel.js', 'clasificador.js', 'fusionTratamientos.js', 'procesadorInformes.js', 'estadisticas.js', 'exportadorExcel.js'];

export function cargarModulos() {
  for (const modulo of MODULOS) {
    vm.runInThisContext(readFileSync(path.join(carpetaJs, modulo), 'utf8'), { filename: modulo });
  }
  return globalThis.LectorYodo;
}
