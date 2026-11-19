import { test } from 'node:test';
import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { mkdtempSync, writeFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const carpetaRaiz = path.join(path.dirname(fileURLToPath(import.meta.url)), '..');
const rutaGancho = path.join(carpetaRaiz, '.githooks', 'commit-msg');
const CORREO_AUTORIZADO = 'chiguinchiang@gmail.com';

function evaluarMensaje(mensaje, correoAutor = CORREO_AUTORIZADO) {
  const carpeta = mkdtempSync(path.join(tmpdir(), 'mensaje-'));
  const archivo = path.join(carpeta, 'MENSAJE');
  writeFileSync(archivo, mensaje);
  try {
    return spawnSync('sh', [rutaGancho, archivo], {
      encoding: 'utf8',
      env: { ...process.env, GIT_AUTHOR_NAME: 'Chiguin', GIT_AUTHOR_EMAIL: correoAutor, GIT_COMMITTER_NAME: 'Chiguin', GIT_COMMITTER_EMAIL: correoAutor }
    });
  } finally {
    rmSync(carpeta, { recursive: true, force: true });
  }
}

test('acepta un mensaje en español con la identidad autorizada', () => {
  assert.equal(evaluarMensaje('fix(prueba): corrige un detalle\n\nExplica el motivo.\n').status, 0);
});

test('rechaza líneas Co-Authored-By en cualquier formato de mayúsculas', () => {
  for (const linea of ['Co-Authored-By: Alguien <a@b.cl>', 'co-authored-by: Alguien <a@b.cl>', 'CO-AUTHORED-BY: Alguien <a@b.cl>']) {
    assert.equal(evaluarMensaje(`feat: algo\n\n${linea}\n`).status, 1, linea);
  }
});

test('rechaza referencias de sesión y de generación por IA', () => {
  assert.equal(evaluarMensaje('feat: algo\n\nClaude-Session: abc\n').status, 1);
  assert.equal(evaluarMensaje('feat: algo\n\nGenerado con Claude Code\n').status, 1);
  assert.equal(evaluarMensaje('feat: algo\n\nhttps://claude.ai/code/sesion\n').status, 1);
  assert.equal(evaluarMensaje('feat: algo\n\nAutor: noreply@anthropic.com\n').status, 1);
});

test('rechaza un commit cuyo autor use otro correo', () => {
  assert.equal(evaluarMensaje('feat: algo\n', 'otra.persona@ejemplo.cl').status, 1);
});
