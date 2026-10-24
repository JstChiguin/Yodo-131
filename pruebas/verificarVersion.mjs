import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { cargarModulos } from './cargarModulos.mjs';

const carpetaRaiz = path.join(path.dirname(fileURLToPath(import.meta.url)), '..');
const versionPaquete = JSON.parse(readFileSync(path.join(carpetaRaiz, 'package.json'), 'utf8')).version;
const { VERSION_APLICACION } = cargarModulos();

if (versionPaquete !== VERSION_APLICACION) {
  console.error(`La versión de package.json (${versionPaquete}) no coincide con version.js (${VERSION_APLICACION}). Actualice ambas.`);
  process.exit(1);
}
