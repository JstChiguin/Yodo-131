import { test, before, after } from 'node:test';
import assert from 'node:assert/strict';
import { createRequire } from 'node:module';
import { fileURLToPath, pathToFileURL } from 'node:url';
import path from 'node:path';
import { cargarModulos } from './cargarModulos.mjs';

const TEXTO_FIRMA_OBLIGATORIO = 'by Chiguin';
const DISTANCIA_MAXIMA_AL_BORDE = 40;
const carpetaRaiz = path.join(path.dirname(fileURLToPath(import.meta.url)), '..');
const direccionPagina = pathToFileURL(path.join(carpetaRaiz, 'Sistema - no tocar', 'index.html')).href;
const carpetaEjemplos = path.join(carpetaRaiz, 'ejemplos');
const { VERSION_APLICACION } = cargarModulos();

function cargarPlaywright() {
  try {
    return createRequire(import.meta.url)('playwright');
  } catch (errorCarga) {
    throw new Error(`No se encontró Playwright, necesario para verificar la firma en el navegador. Ejecute "npm ci" y "npx playwright install chromium". Detalle: ${errorCarga.message}`);
  }
}

let navegador;

before(async () => {
  navegador = await cargarPlaywright().chromium.launch();
});

after(async () => {
  await navegador?.close();
});

async function medirFirma(pagina) {
  return pagina.evaluate(() => {
    const firmas = document.querySelectorAll('#firma-autor');
    const firma = firmas[0];
    const sello = firma?.querySelector('.sello-autor');
    const version = firma?.querySelector('.sello-version');
    const describir = (nodo) => {
      if (!nodo) return null;
      const caja = nodo.getBoundingClientRect();
      const estilo = getComputedStyle(nodo);
      return {
        texto: nodo.textContent, izquierda: caja.left, derecha: caja.right, arriba: caja.top, abajo: caja.bottom, ancho: caja.width, alto: caja.height,
        display: estilo.display, visibility: estilo.visibility, opacity: Number(estilo.opacity), position: estilo.position, pointerEvents: estilo.pointerEvents
      };
    };
    const nivelFirma = firma ? Number(getComputedStyle(firma).zIndex) || 0 : 0;
    const elementosEncima = firma ? [...document.body.querySelectorAll('*')].filter((nodo) => {
      if (firma.contains(nodo)) return false;
      const estilo = getComputedStyle(nodo);
      if (estilo.position === 'static' || estilo.display === 'none' || estilo.visibility === 'hidden') return false;
      return (Number(estilo.zIndex) || 0) >= nivelFirma;
    }).map((nodo) => nodo.id || nodo.className || nodo.tagName) : [];
    return {
      cantidadFirmas: firmas.length,
      firma: describir(firma),
      sello: describir(sello),
      version: describir(version),
      anchoVentana: innerWidth,
      altoVentana: innerHeight,
      elementosEncima
    };
  });
}

function verificarFirma(medicion, escenario) {
  const prefijo = `FIRMA DE AUTOR ALTERADA (${escenario})`;
  assert.equal(medicion.cantidadFirmas, 1, `${prefijo}: debe existir exactamente una firma #firma-autor y se encontraron ${medicion.cantidadFirmas}.`);
  assert.ok(medicion.sello, `${prefijo}: falta el elemento .sello-autor con el texto "${TEXTO_FIRMA_OBLIGATORIO}".`);
  assert.equal(medicion.sello.texto, TEXTO_FIRMA_OBLIGATORIO, `${prefijo}: el texto debe ser exactamente "${TEXTO_FIRMA_OBLIGATORIO}" (mayúsculas, minúsculas y espacios) y se encontró "${medicion.sello.texto}".`);
  assert.ok(medicion.version, `${prefijo}: falta la versión debajo de la firma.`);
  assert.equal(medicion.version.texto, `v${VERSION_APLICACION}`, `${prefijo}: la versión mostrada debe ser "v${VERSION_APLICACION}" y se encontró "${medicion.version.texto}".`);

  for (const [nombre, nodo] of [['contenedor', medicion.firma], ['texto', medicion.sello], ['versión', medicion.version]]) {
    assert.notEqual(nodo.display, 'none', `${prefijo}: el ${nombre} de la firma está oculto con display: none.`);
    assert.notEqual(nodo.visibility, 'hidden', `${prefijo}: el ${nombre} de la firma está oculto con visibility: hidden.`);
    assert.ok(nodo.opacity >= 0.3, `${prefijo}: el ${nombre} de la firma es casi invisible (opacidad ${nodo.opacity}).`);
    assert.ok(nodo.ancho > 0 && nodo.alto > 0, `${prefijo}: el ${nombre} de la firma no tiene tamaño visible.`);
  }
  assert.deepEqual(medicion.elementosEncima, [], `${prefijo}: hay elementos visibles por encima de la firma que pueden taparla: ${medicion.elementosEncima.join(', ')}.`);

  assert.equal(medicion.firma.position, 'fixed', `${prefijo}: la firma debe ser flotante (position: fixed) y es "${medicion.firma.position}".`);
  assert.equal(medicion.firma.pointerEvents, 'none', `${prefijo}: la firma no debe capturar clics (pointer-events: none) para no tapar botones.`);
  const distanciaDerecha = medicion.anchoVentana - medicion.firma.derecha;
  const distanciaInferior = medicion.altoVentana - medicion.firma.abajo;
  assert.ok(distanciaDerecha >= 0 && distanciaDerecha <= DISTANCIA_MAXIMA_AL_BORDE, `${prefijo}: la firma debe estar en la esquina inferior DERECHA; está a ${Math.round(distanciaDerecha)} px del borde derecho (máximo ${DISTANCIA_MAXIMA_AL_BORDE}).`);
  assert.ok(distanciaInferior >= 0 && distanciaInferior <= DISTANCIA_MAXIMA_AL_BORDE, `${prefijo}: la firma debe estar en la esquina INFERIOR derecha; está a ${Math.round(distanciaInferior)} px del borde inferior (máximo ${DISTANCIA_MAXIMA_AL_BORDE}).`);
  assert.ok(medicion.firma.izquierda > medicion.anchoVentana / 2, `${prefijo}: la firma quedó en la mitad izquierda de la pantalla.`);

  assert.ok(medicion.version.arriba >= medicion.sello.abajo - 1, `${prefijo}: la versión debe ir debajo del texto "${TEXTO_FIRMA_OBLIGATORIO}".`);
  const centroSello = medicion.sello.izquierda + medicion.sello.ancho / 2;
  const centroVersion = medicion.version.izquierda + medicion.version.ancho / 2;
  assert.ok(Math.abs(centroSello - centroVersion) <= 2, `${prefijo}: la versión debe estar centrada bajo la firma (diferencia de ${Math.round(Math.abs(centroSello - centroVersion))} px).`);
  assert.ok(medicion.version.alto < medicion.sello.alto, `${prefijo}: la versión debe verse más pequeña que la firma.`);
}

const ESCENARIOS = [
  { nombre: 'escritorio, modo claro', vista: { width: 1366, height: 768 }, cambiarAOscuro: false },
  { nombre: 'escritorio, modo oscuro', vista: { width: 1366, height: 768 }, cambiarAOscuro: true },
  { nombre: 'tablet', vista: { width: 820, height: 1180 }, cambiarAOscuro: false },
  { nombre: 'celular', vista: { width: 390, height: 844 }, cambiarAOscuro: false }
];

for (const escenario of ESCENARIOS) {
  test(`la firma "${TEXTO_FIRMA_OBLIGATORIO}" está intacta en todas las pantallas (${escenario.nombre})`, async () => {
    const pagina = await navegador.newPage({ viewport: escenario.vista });
    const erroresPagina = [];
    pagina.on('pageerror', (error) => erroresPagina.push(error.message));
    try {
      await pagina.goto(direccionPagina);
      if (escenario.cambiarAOscuro) await pagina.click('#boton-tema');
      verificarFirma(await medirFirma(pagina), `${escenario.nombre}, pantalla de inicio`);

      await pagina.setInputFiles('#selector-carpeta', carpetaEjemplos);
      await pagina.waitForSelector('#seccion-resumen:not([hidden])');
      for (const seccion of ['resumen', 'pacientes', 'revisar', 'mantenimiento']) {
        await pagina.click(`.pestana[data-seccion="${seccion}"]`);
        verificarFirma(await medirFirma(pagina), `${escenario.nombre}, pestaña ${seccion}`);
      }
      await pagina.click('.pestana[data-seccion="resumen"]');
      await pagina.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
      verificarFirma(await medirFirma(pagina), `${escenario.nombre}, final de la página`);
      await pagina.click('#boton-presentacion');
      verificarFirma(await medirFirma(pagina), `${escenario.nombre}, modo presentación`);
      assert.deepEqual(erroresPagina, [], `La página mostró errores: ${erroresPagina.join(' | ')}`);
    } finally {
      await pagina.close();
    }
  });
}
