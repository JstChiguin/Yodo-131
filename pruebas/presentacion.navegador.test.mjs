import { test, before, after } from 'node:test';
import assert from 'node:assert/strict';
import { createRequire } from 'node:module';
import { fileURLToPath, pathToFileURL } from 'node:url';
import path from 'node:path';

const carpetaRaiz = path.join(path.dirname(fileURLToPath(import.meta.url)), '..');
const direccionPagina = pathToFileURL(path.join(carpetaRaiz, 'Sistema - no tocar', 'index.html')).href;
const carpetaEjemplos = path.join(carpetaRaiz, 'ejemplos');

let navegador;

before(async () => {
  navegador = await createRequire(import.meta.url)('playwright').chromium.launch();
});

after(async () => {
  await navegador?.close();
});

const SELECTORES_SOLO_REVISION = [
  '.pestana[data-seccion="revisar"]',
  '#contador-revisar',
  '.pestana[data-seccion="pacientes"]',
  '.pestana[data-seccion="mantenimiento"]',
  '#etiqueta-selector',
  '#boton-excel',
  '#boton-empezar',
  '#nota-archivos',
  '#tarjeta-origen'
];

async function listarVisibles(pagina, selectores) {
  const visibles = [];
  for (const selector of selectores) {
    if (await pagina.locator(selector).isVisible()) visibles.push(selector);
  }
  return visibles;
}

async function obtenerTextoCifras(pagina) {
  return (await pagina.locator('#cifras .cifra').evaluateAll((cifras) => cifras.filter((cifra) => cifra.offsetParent).map((cifra) => cifra.textContent))).join(' | ');
}

test('el modo presentación oculta lo que solo debe ver la doctora y se puede deshacer', async () => {
  const pagina = await navegador.newPage({ viewport: { width: 1366, height: 768 } });
  const erroresPagina = [];
  pagina.on('pageerror', (error) => erroresPagina.push(error.message));
  try {
    await pagina.goto(direccionPagina);
    assert.ok(await pagina.locator('#boton-presentacion').isDisabled(), 'sin datos cargados no hay nada que presentar');

    await pagina.setInputFiles('#selector-carpeta', carpetaEjemplos);
    await pagina.waitForSelector('#seccion-resumen:not([hidden])');
    assert.match(await obtenerTextoCifras(pagina), /Datos que no coinciden/);
    await pagina.click('.pestana[data-seccion="revisar"]');
    assert.ok(await pagina.locator('#contador-revisar').isVisible(), 'fuera de la presentación el contador rojo sigue a la vista');

    await pagina.click('#boton-presentacion');
    assert.ok(await pagina.locator('#seccion-resumen').isVisible(), 'al presentar se vuelve a los gráficos');
    assert.deepEqual(await listarVisibles(pagina, SELECTORES_SOLO_REVISION), [], 'en presentación no debe verse nada de revisión interna');
    assert.doesNotMatch(await obtenerTextoCifras(pagina), /Datos que no coinciden/);
    assert.ok(await pagina.locator('#grafico-anios').isVisible(), 'los gráficos siguen visibles');
    assert.ok(await pagina.locator('#boton-presentacion').isVisible());
    assert.equal(await pagina.locator('#boton-presentacion').textContent(), 'Salir de presentación');

    const cifra = pagina.locator('#cifras .cifra').first();
    await cifra.hover();
    await pagina.waitForTimeout(300);
    assert.notEqual(await cifra.evaluate((nodo) => getComputedStyle(nodo).transform), 'none', 'la tarjeta bajo el mouse se resalta');

    const barras = pagina.locator('#grafico-diagnostico .barra-fila.marca');
    await barras.first().hover();
    await pagina.waitForTimeout(300);
    const opacidades = await barras.evaluateAll((filas) => filas.map((fila) => Number(getComputedStyle(fila).opacity)));
    assert.equal(opacidades[0], 1, 'la barra bajo el mouse queda destacada');
    assert.ok(opacidades.slice(1).every((opacidad) => opacidad < 1), 'las demás barras del gráfico se atenúan');

    await pagina.click('#boton-presentacion');
    assert.deepEqual(await listarVisibles(pagina, ['.pestana[data-seccion="revisar"]', '#contador-revisar', '#boton-excel']), ['.pestana[data-seccion="revisar"]', '#contador-revisar', '#boton-excel']);
    assert.match(await obtenerTextoCifras(pagina), /Datos que no coinciden/);
    assert.deepEqual(erroresPagina, []);
  } finally {
    await pagina.close();
  }
});
