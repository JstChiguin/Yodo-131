import { test } from 'node:test';
import assert from 'node:assert/strict';
import { execFileSync, spawnSync } from 'node:child_process';
import { chmodSync, existsSync, mkdtempSync, readFileSync, readdirSync, rmSync, statSync, writeFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { cargarModulos } from './cargarModulos.mjs';

const carpetaRaiz = path.join(path.dirname(fileURLToPath(import.meta.url)), '..');
const carpetaPortable = path.join(carpetaRaiz, 'portable');
const carpetaAplicacion = path.join(carpetaRaiz, 'Sistema - no tocar');
const carpetaEjemplos = path.join(carpetaRaiz, 'ejemplos');
const PUERTO_DEPURACION = 9333;
const { VERSION_APLICACION } = cargarModulos();
const hayGo = spawnSync('go', ['version']).status === 0;

function listarArchivos(carpeta, prefijo = '') {
  return readdirSync(carpeta).flatMap((nombre) => {
    const ruta = path.join(carpeta, nombre);
    return statSync(ruta).isDirectory() ? listarArchivos(ruta, `${prefijo}${nombre}/`) : [`${prefijo}${nombre}`];
  });
}

async function esperar(condicion, descripcion, limiteMs = 20000) {
  const inicio = Date.now();
  while (Date.now() - inicio < limiteMs) {
    const resultado = await condicion().catch(() => null);
    if (resultado) return resultado;
    await new Promise((resolver) => setTimeout(resolver, 200));
  }
  throw new Error(`Tiempo agotado esperando ${descripcion}`);
}

test('el lanzador portable abre la aplicación empaquetada sin internet y con la firma intacta', { skip: hayGo ? false : 'Go no está instalado: no se puede compilar el lanzador portable' }, async () => {
  const { chromium } = createRequire(import.meta.url)('playwright');
  const carpetaTemporal = mkdtempSync(path.join(tmpdir(), 'yodo131-portable-'));
  const carpetaBinarios = path.join(carpetaTemporal, 'bin');
  const carpetaCache = path.join(carpetaTemporal, 'cache');
  const rutaArgumentos = path.join(carpetaTemporal, 'argumentos.txt');
  const rutaProceso = path.join(carpetaTemporal, 'proceso.txt');
  const rutaSalidaNavegador = path.join(carpetaTemporal, 'navegador.log');
  let navegador;
  try {
    execFileSync('go', ['run', './generador', carpetaAplicacion, path.join('lanzador', 'aplicacion_empaquetada.go')], { cwd: carpetaPortable });
    const rutaLanzador = path.join(carpetaTemporal, 'Yodo131');
    execFileSync('go', ['build', '-o', rutaLanzador, './lanzador'], { cwd: carpetaPortable, env: { ...process.env, CGO_ENABLED: '0' } });

    execFileSync('mkdir', ['-p', carpetaBinarios]);
    const rutaEnvoltorio = path.join(carpetaBinarios, 'chromium');
    writeFileSync(rutaEnvoltorio, `#!/bin/sh\nprintf '%s\\n' "$@" > '${rutaArgumentos}'\necho $$ > '${rutaProceso}'\nexec '${chromium.executablePath()}' --headless=new${process.getuid?.() === 0 ? ' --no-sandbox' : ''} --remote-debugging-port=${PUERTO_DEPURACION} "$@" 2> '${rutaSalidaNavegador}'\n`);
    chmodSync(rutaEnvoltorio, 0o755);

    const ejecucion = spawnSync(rutaLanzador, [], {
      env: { ...process.env, PATH: `${carpetaBinarios}:/usr/bin:/bin`, XDG_CACHE_HOME: carpetaCache, HOME: carpetaTemporal },
      encoding: 'utf8',
      timeout: 20000
    });
    assert.equal(ejecucion.status, 0, `el lanzador terminó con error: ${ejecucion.stderr}`);

    const carpetaCopiada = path.join(carpetaCache, 'Yodo131', 'aplicacion', VERSION_APLICACION);
    assert.ok(existsSync(path.join(carpetaCopiada, 'index.html')), 'el lanzador debe copiar la aplicación a su carpeta de datos');
    const archivosOriginales = listarArchivos(carpetaAplicacion).sort();
    assert.deepEqual(listarArchivos(carpetaCopiada).sort(), archivosOriginales, 'la copia debe tener exactamente los archivos de la aplicación');
    for (const archivo of archivosOriginales) {
      assert.ok(readFileSync(path.join(carpetaCopiada, archivo)).equals(readFileSync(path.join(carpetaAplicacion, archivo))), `${archivo} cambió al empaquetarse`);
    }

    const argumentos = (await esperar(async () => existsSync(rutaArgumentos) && readFileSync(rutaArgumentos, 'utf8'), 'los argumentos del navegador')).trim().split('\n');
    assert.ok(argumentos.includes(`--app=file://${path.join(carpetaCopiada, 'index.html')}`), `se esperaba --app con la página copiada: ${argumentos.join(' ')}`);
    assert.ok(argumentos.includes(`--user-data-dir=${path.join(carpetaCache, 'Yodo131', 'perfil-navegador')}`));
    for (const obligatorio of ['--no-first-run', '--disable-extensions', '--disable-sync', '--disable-background-networking']) {
      assert.ok(argumentos.includes(obligatorio), `falta ${obligatorio}`);
    }

    navegador = await esperar(() => chromium.connectOverCDP(`http://127.0.0.1:${PUERTO_DEPURACION}`), 'el navegador abierto por el lanzador')
      .catch((errorConexion) => { throw new Error(`${errorConexion.message}. Salida del navegador: ${existsSync(rutaSalidaNavegador) ? readFileSync(rutaSalidaNavegador, 'utf8').slice(-800) : 'sin salida'}`); });
    const [contexto] = navegador.contexts();
    const pagina = await esperar(async () => contexto.pages().find((candidata) => candidata.url().endsWith('/index.html')), 'la ventana de la aplicación');
    const direccionesSolicitadas = [];
    pagina.on('request', (solicitud) => direccionesSolicitadas.push(solicitud.url()));
    await contexto.setOffline(true);
    await pagina.reload();
    await pagina.waitForSelector('#firma-autor .sello-autor');

    assert.equal(await pagina.locator('#firma-autor .sello-autor').textContent(), 'by Chiguin', 'FIRMA DE AUTOR ALTERADA en la versión portable');
    assert.equal(await pagina.locator('#firma-autor .sello-version').textContent(), `v${VERSION_APLICACION}`);

    await pagina.setInputFiles('#selector-carpeta', carpetaEjemplos);
    await pagina.waitForSelector('#seccion-resumen:not([hidden])');
    const cifraTratamientos = await pagina.locator('#cifras .cifra-valor').first().textContent();
    assert.equal(cifraTratamientos, '12', 'los ejemplos ficticios deben dar 12 tratamientos, igual que en la versión de navegador');

    const solicitudesExternas = direccionesSolicitadas.filter((direccion) => !/^(file|data|blob):/.test(direccion));
    assert.deepEqual(solicitudesExternas, [], 'la aplicación no debe pedir nada a internet');
  } finally {
    await navegador?.close().catch(() => {});
    if (existsSync(rutaProceso)) {
      const idProceso = Number(readFileSync(rutaProceso, 'utf8'));
      const sigueActivo = () => { try { process.kill(idProceso, 0); return true; } catch { return false; } };
      if (sigueActivo()) process.kill(idProceso, 'SIGKILL');
      await esperar(async () => !sigueActivo(), 'el cierre del navegador', 5000).catch(() => {});
    }
    rmSync(carpetaTemporal, { recursive: true, force: true, maxRetries: 10, retryDelay: 200 });
  }
});
