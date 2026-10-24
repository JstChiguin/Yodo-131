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

  function esDocAntiguo(bytes) {
    return bytes[0] === 0xd0 && bytes[1] === 0xcf && bytes[2] === 0x11 && bytes[3] === 0xe0;
  }

  function buscarFinDirectorio(vista) {
    const inicioBusqueda = Math.max(0, vista.byteLength - 65557);
    for (let posicion = vista.byteLength - 22; posicion >= inicioBusqueda; posicion--) {
      if (vista.getUint32(posicion, true) === FIRMA_FIN_DIRECTORIO) return posicion;
    }
    throw new FormatoNoSoportadoError('El archivo no es un Word .docx válido');
  }

  function listarEntradasZip(buffer) {
    const vista = new DataView(buffer);
    const posicionFin = buscarFinDirectorio(vista);
    const cantidadEntradas = vista.getUint16(posicionFin + 10, true);
    let posicion = vista.getUint32(posicionFin + 16, true);
    const decodificador = new TextDecoder();
    const entradas = [];
    for (let indice = 0; indice < cantidadEntradas; indice++) {
      if (vista.getUint32(posicion, true) !== FIRMA_DIRECTORIO_CENTRAL) break;
      const metodoCompresion = vista.getUint16(posicion + 10, true);
      const tamanoComprimido = vista.getUint32(posicion + 20, true);
      const largoNombre = vista.getUint16(posicion + 28, true);
      const largoExtra = vista.getUint16(posicion + 30, true);
      const largoComentario = vista.getUint16(posicion + 32, true);
      const posicionCabeceraLocal = vista.getUint32(posicion + 42, true);
      const nombre = decodificador.decode(new Uint8Array(buffer, posicion + 46, largoNombre));
      entradas.push({ nombre, metodoCompresion, tamanoComprimido, posicionCabeceraLocal });
      posicion += 46 + largoNombre + largoExtra + largoComentario;
    }
    return entradas;
  }

  async function descomprimir(bytesComprimidos) {
    const flujo = new Blob([bytesComprimidos]).stream().pipeThrough(new DecompressionStream('deflate-raw'));
    return new Uint8Array(await new Response(flujo).arrayBuffer());
  }

  async function leerEntradaZip(buffer, nombreBuscado) {
    const entrada = listarEntradasZip(buffer).find((item) => item.nombre === nombreBuscado);
    if (!entrada) throw new FormatoNoSoportadoError(`El Word no contiene ${nombreBuscado}`);
    const vista = new DataView(buffer);
    if (vista.getUint32(entrada.posicionCabeceraLocal, true) !== FIRMA_CABECERA_LOCAL) {
      throw new FormatoNoSoportadoError('Estructura interna del Word dañada');
    }
    const largoNombre = vista.getUint16(entrada.posicionCabeceraLocal + 26, true);
    const largoExtra = vista.getUint16(entrada.posicionCabeceraLocal + 28, true);
    const inicioDatos = entrada.posicionCabeceraLocal + 30 + largoNombre + largoExtra;
    const datos = new Uint8Array(buffer, inicioDatos, entrada.tamanoComprimido);
    if (entrada.metodoCompresion === 0) return new TextDecoder().decode(datos);
    if (entrada.metodoCompresion === 8) return new TextDecoder().decode(await descomprimir(datos));
    throw new FormatoNoSoportadoError('Compresión del Word no soportada');
  }

  function decodificarEntidades(texto) {
    return texto
      .replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&quot;/g, '"').replace(/&apos;/g, "'")
      .replace(/&#(\d+);/g, (_, codigo) => String.fromCodePoint(Number(codigo)))
      .replace(/&#x([0-9a-f]+);/gi, (_, codigo) => String.fromCodePoint(parseInt(codigo, 16)))
      .replace(/&amp;/g, '&');
  }

  function extraerParrafos(xmlDocumento) {
    const parrafos = [];
    const bloques = xmlDocumento.match(/<w:p[ >][\s\S]*?<\/w:p>/g) || [];
    for (const bloque of bloques) {
      const texto = (bloque.match(/<w:t(?: [^>]*)?>[^<]*<\/w:t>|<w:tab\/>|<w:br\/>/g) || [])
        .map((pieza) => {
          if (pieza === '<w:tab/>') return ' ';
          if (pieza === '<w:br/>') return '\n';
          return decodificarEntidades(pieza.replace(/<[^>]+>/g, ''));
        })
        .join('');
      const textoLimpio = texto.replace(/[  ]+/g, ' ').trim();
      if (textoLimpio) parrafos.push(...textoLimpio.split('\n').map((linea) => linea.trim()).filter(Boolean));
    }
    return parrafos;
  }

  async function leerParrafosDocx(buffer) {
    const bytesIniciales = new Uint8Array(buffer, 0, Math.min(4, buffer.byteLength));
    if (esDocAntiguo(bytesIniciales)) {
      throw new FormatoNoSoportadoError('Es un Word antiguo (.doc). Ábralo en Word y use "Guardar como" .docx');
    }
    const xmlDocumento = await leerEntradaZip(buffer, 'word/document.xml');
    return extraerParrafos(xmlDocumento);
  }

  ns.LectorDocx = { leerParrafosDocx, FormatoNoSoportadoError };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
