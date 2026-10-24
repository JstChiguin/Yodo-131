import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { cargarModulos } from './cargarModulos.mjs';

const carpetaRaiz = path.join(path.dirname(fileURLToPath(import.meta.url)), '..');
const leerArchivo = (rutaRelativa) => readFileSync(path.join(carpetaRaiz, rutaRelativa), 'utf8');

test('la versión central sigue versionamiento semántico y está en el CHANGELOG', () => {
  const { VERSION_APLICACION } = cargarModulos();
  assert.match(VERSION_APLICACION, /^\d+\.\d+\.\d+$/);
  assert.ok(leerArchivo('CHANGELOG.md').includes(`## [${VERSION_APLICACION}]`), 'falta la entrada de la versión actual en CHANGELOG.md');
});

test('la página carga la versión y el sello antes de iniciar la aplicación', () => {
  const pagina = leerArchivo('Sistema - no tocar/index.html');
  const posicionVersion = pagina.indexOf('js/version.js');
  const posicionSello = pagina.indexOf('js/selloAutor.js');
  const posicionApp = pagina.indexOf('js/app.js');
  assert.ok(posicionVersion > -1 && posicionSello > -1 && posicionApp > -1);
  assert.ok(posicionVersion < posicionSello && posicionSello < posicionApp);
});

test('el sello conserva el texto exacto y se monta en la raíz de la aplicación', () => {
  assert.ok(leerArchivo('Sistema - no tocar/js/selloAutor.js').includes("const TEXTO_SELLO = 'by Chiguin';"), 'FIRMA DE AUTOR ALTERADA: selloAutor.js debe definir TEXTO_SELLO exactamente como \'by Chiguin\'.');
  assert.ok(leerArchivo('Sistema - no tocar/js/app.js').includes("SelloAutor.montarSello(elemento('app'))"), 'FIRMA DE AUTOR ALTERADA: app.js ya no monta la firma en la raíz de la aplicación (#app).');
  const estilos = leerArchivo('Sistema - no tocar/css/estilos.css');
  assert.match(estilos, /\.firma-autor\s*\{[^}]*position:\s*fixed/, 'FIRMA DE AUTOR ALTERADA: .firma-autor debe ser flotante (position: fixed).');
  assert.match(estilos, /\.firma-autor\s*\{[^}]*pointer-events:\s*none/, 'FIRMA DE AUTOR ALTERADA: .firma-autor debe tener pointer-events: none para no tapar botones.');
});
