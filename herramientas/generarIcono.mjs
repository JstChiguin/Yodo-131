import { readFileSync, writeFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { chromium } from 'playwright';

const carpetaRecursos = path.join(path.dirname(fileURLToPath(import.meta.url)), '..', 'portable', 'recursos');
const TAMANOS = [16, 24, 32, 48, 64, 128, 256];

function crearIco(imagenesPng) {
  const encabezado = Buffer.alloc(6 + imagenesPng.length * 16);
  encabezado.writeUInt16LE(0, 0);
  encabezado.writeUInt16LE(1, 2);
  encabezado.writeUInt16LE(imagenesPng.length, 4);
  let desplazamiento = encabezado.length;
  imagenesPng.forEach(({ tamano, datos }, indice) => {
    const posicion = 6 + indice * 16;
    encabezado.writeUInt8(tamano >= 256 ? 0 : tamano, posicion);
    encabezado.writeUInt8(tamano >= 256 ? 0 : tamano, posicion + 1);
    encabezado.writeUInt16LE(1, posicion + 4);
    encabezado.writeUInt16LE(32, posicion + 6);
    encabezado.writeUInt32LE(datos.length, posicion + 8);
    encabezado.writeUInt32LE(desplazamiento, posicion + 12);
    desplazamiento += datos.length;
  });
  return Buffer.concat([encabezado, ...imagenesPng.map(({ datos }) => datos)]);
}

const dibujo = readFileSync(path.join(carpetaRecursos, 'icono.svg'), 'utf8');
const navegador = await chromium.launch();
try {
  const imagenesPng = [];
  for (const tamano of TAMANOS) {
    const pagina = await navegador.newPage({ viewport: { width: tamano, height: tamano } });
    await pagina.setContent(`<style>html,body{margin:0;background:transparent}svg{display:block;width:${tamano}px;height:${tamano}px}</style>${dibujo}`);
    imagenesPng.push({ tamano, datos: await pagina.screenshot({ omitBackground: true }) });
    await pagina.close();
  }
  writeFileSync(path.join(carpetaRecursos, 'icono.ico'), crearIco(imagenesPng));
} finally {
  await navegador.close();
}
