(function (ns) {
  const FIRMA_FIN_DIRECTORIO = 0x06054b50;
  const FIRMA_DIRECTORIO_CENTRAL = 0x02014b50;
  const FIRMA_CABECERA_LOCAL = 0x04034b50;

  class FormatoNoSoportadoError extends Error {
    constructor(mensaje) {
      super(mensaje);
      this.name = 'FormatoNoSoportadoError';
    }
  }

  function esFormatoOfficeAntiguo(buffer) {
    const bytes = new Uint8Array(buffer, 0, Math.min(4, buffer.byteLength));
    return bytes[0] === 0xd0 && bytes[1] === 0xcf && bytes[2] === 0x11 && bytes[3] === 0xe0;
  }

  function buscarFinDirectorio(vista) {
    const inicioBusqueda = Math.max(0, vista.byteLength - 65557);
    for (let posicion = vista.byteLength - 22; posicion >= inicioBusqueda; posicion--) {
      if (vista.getUint32(posicion, true) === FIRMA_FIN_DIRECTORIO) return posicion;
    }
    throw new FormatoNoSoportadoError('El archivo está dañado o no es un documento de Office válido');
  }

  function listarEntradas(buffer) {
    const vista = new DataView(buffer);
    const posicionFin = buscarFinDirectorio(vista);
    const cantidadEntradas = vista.getUint16(posicionFin + 10, true);
    let posicion = vista.getUint32(posicionFin + 16, true);
    const decodificador = new TextDecoder();
    const entradas = [];
    for (let indice = 0; indice < cantidadEntradas; indice++) {
      if (vista.getUint32(posicion, true) !== FIRMA_DIRECTORIO_CENTRAL) break;
      const largoNombre = vista.getUint16(posicion + 28, true);
      entradas.push({
        nombre: decodificador.decode(new Uint8Array(buffer, posicion + 46, largoNombre)),
        metodoCompresion: vista.getUint16(posicion + 10, true),
        tamanoComprimido: vista.getUint32(posicion + 20, true),
        posicionCabeceraLocal: vista.getUint32(posicion + 42, true)
      });
      posicion += 46 + largoNombre + vista.getUint16(posicion + 30, true) + vista.getUint16(posicion + 32, true);
    }
    return entradas;
  }

  async function descomprimir(bytesComprimidos) {
    const flujo = new Blob([bytesComprimidos]).stream().pipeThrough(new DecompressionStream('deflate-raw'));
    return new Uint8Array(await new Response(flujo).arrayBuffer());
  }

  async function leerTextoEntrada(buffer, nombreBuscado, { esOpcional = false } = {}) {
    const entrada = listarEntradas(buffer).find((item) => item.nombre === nombreBuscado);
    if (!entrada) {
      if (esOpcional) return null;
      throw new FormatoNoSoportadoError(`Falta la parte interna ${nombreBuscado}`);
    }
    const vista = new DataView(buffer);
    if (vista.getUint32(entrada.posicionCabeceraLocal, true) !== FIRMA_CABECERA_LOCAL) {
      throw new FormatoNoSoportadoError('La estructura interna del archivo está dañada');
    }
    const largoNombre = vista.getUint16(entrada.posicionCabeceraLocal + 26, true);
    const largoExtra = vista.getUint16(entrada.posicionCabeceraLocal + 28, true);
    const datos = new Uint8Array(buffer, entrada.posicionCabeceraLocal + 30 + largoNombre + largoExtra, entrada.tamanoComprimido);
    if (entrada.metodoCompresion === 0) return new TextDecoder().decode(datos);
    if (entrada.metodoCompresion === 8) return new TextDecoder().decode(await descomprimir(datos));
    throw new FormatoNoSoportadoError('El archivo usa una compresión no soportada');
  }

  function decodificarEntidades(texto) {
    return texto
      .replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&quot;/g, '"').replace(/&apos;/g, "'")
      .replace(/&#(\d+);/g, (_, codigo) => String.fromCodePoint(Number(codigo)))
      .replace(/&#x([0-9a-f]+);/gi, (_, codigo) => String.fromCodePoint(parseInt(codigo, 16)))
      .replace(/&amp;/g, '&');
  }

  ns.LectorZip = { leerTextoEntrada, esFormatoOfficeAntiguo, decodificarEntidades, FormatoNoSoportadoError };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
