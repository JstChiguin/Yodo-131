(function (ns) {
  const SAL_HEX = 'e81dc673beaf9aed204c53f62e22a657';
  const HUELLA_HEX = 'cc5d58cdae068134fd66186eae0bf91a91adc8f69ad98fb186145cf2feb259e1';
  const ITERACIONES = 310000;
  const MAXIMO_INTENTOS = 5;
  const ESPERA_BLOQUEO_MS = 30000;

  let intentosFallidos = 0;
  let bloqueadoHasta = 0;

  function hexABytes(hex) {
    return new Uint8Array(hex.match(/../g).map((par) => parseInt(par, 16)));
  }

  function bytesAHex(bytes) {
    return [...new Uint8Array(bytes)].map((byte) => byte.toString(16).padStart(2, '0')).join('');
  }

  function compararSinFiltrarTiempo(textoA, textoB) {
    if (textoA.length !== textoB.length) return false;
    let diferencia = 0;
    for (let indice = 0; indice < textoA.length; indice++) diferencia |= textoA.charCodeAt(indice) ^ textoB.charCodeAt(indice);
    return diferencia === 0;
  }

  function segundosDeEspera() {
    return Math.max(0, Math.ceil((bloqueadoHasta - Date.now()) / 1000));
  }

  async function verificarClave(claveIngresada) {
    if (segundosDeEspera() > 0) return { valida: false, esperaSegundos: segundosDeEspera() };
    if (!globalThis.crypto?.subtle) throw new Error('Este navegador no permite verificar la clave. Use Chrome o Edge actualizados.');
    const material = await crypto.subtle.importKey('raw', new TextEncoder().encode(claveIngresada), 'PBKDF2', false, ['deriveBits']);
    const bits = await crypto.subtle.deriveBits({ name: 'PBKDF2', hash: 'SHA-256', salt: hexABytes(SAL_HEX), iterations: ITERACIONES }, material, 256);
    const valida = compararSinFiltrarTiempo(bytesAHex(bits), HUELLA_HEX);
    if (valida) {
      intentosFallidos = 0;
      return { valida: true, esperaSegundos: 0 };
    }
    intentosFallidos += 1;
    if (intentosFallidos >= MAXIMO_INTENTOS) {
      intentosFallidos = 0;
      bloqueadoHasta = Date.now() + ESPERA_BLOQUEO_MS;
    }
    return { valida: false, esperaSegundos: segundosDeEspera() };
  }

  ns.Seguridad = { verificarClave };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
