(function (ns) {
  const { leerTextoEntrada, esFormatoOfficeAntiguo, decodificarEntidades, FormatoNoSoportadoError } = ns.LectorZip;
  const FORMATOS_FECHA_INCORPORADOS = new Set([14, 15, 16, 17, 18, 19, 20, 21, 22, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 45, 46, 47, 50, 51, 52, 53, 54, 55, 56, 57, 58]);
  const MILISEGUNDOS_POR_DIA = 86400000;

  function leerAtributos(textoEtiqueta) {
    const atributos = {};
    for (const [, nombre, valor] of textoEtiqueta.matchAll(/([\w:]+)="([^"]*)"/g)) atributos[nombre] = decodificarEntidades(valor);
    return atributos;
  }

  function extraerTextoRico(fragmentoXml) {
    const sinFonetica = fragmentoXml.replace(/<rPh\b[\s\S]*?<\/rPh>/g, '');
    return [...sinFonetica.matchAll(/<t(?:\s[^>]*)?>([^<]*)<\/t>/g)].map(([, texto]) => decodificarEntidades(texto)).join('');
  }

  function leerTextosCompartidos(xml) {
    if (!xml) return [];
    return [...xml.matchAll(/<si>([\s\S]*?)<\/si>/g)].map(([, contenido]) => extraerTextoRico(contenido));
  }

  function esCodigoFormatoFecha(codigoFormato) {
    const limpio = codigoFormato.replace(/"[^"]*"/g, '').replace(/\[[^\]]*\]/g, '').toLowerCase();
    return /[dy]/.test(limpio) || (/m/.test(limpio) && !/[hs]/.test(limpio));
  }

  function leerEstilosFecha(xml) {
    const estilosFecha = new Set();
    if (!xml) return estilosFecha;
    const formatosPropios = new Map([...xml.matchAll(/<numFmt\b([^>]*)\/?>/g)].map(([, atributos]) => {
      const { numFmtId, formatCode } = leerAtributos(atributos);
      return [Number(numFmtId), formatCode || ''];
    }));
    const bloqueEstilos = (xml.match(/<cellXfs\b[\s\S]*?<\/cellXfs>/) || [''])[0];
    [...bloqueEstilos.matchAll(/<xf\b([^>]*?)\/?>/g)].forEach(([, atributos], indice) => {
      const idFormato = Number(leerAtributos(atributos).numFmtId || 0);
      if (FORMATOS_FECHA_INCORPORADOS.has(idFormato) || (formatosPropios.has(idFormato) && esCodigoFormatoFecha(formatosPropios.get(idFormato)))) estilosFecha.add(indice);
    });
    return estilosFecha;
  }

  function convertirSerialAFecha(serial, usaSistema1904) {
    const origen = usaSistema1904 ? Date.UTC(1904, 0, 1) : Date.UTC(1899, 11, 30);
    const fecha = new Date(origen + Math.round(serial * MILISEGUNDOS_POR_DIA));
    return { fecha: fecha.toISOString().slice(0, 10) };
  }

  function indiceColumna(referencia) {
    const letras = referencia.replace(/\d+/g, '');
    let indice = 0;
    for (const letra of letras) indice = indice * 26 + (letra.charCodeAt(0) - 64);
    return indice - 1;
  }

  function leerHoja(xml, textosCompartidos, estilosFecha, usaSistema1904) {
    const filas = [];
    for (const [, atributosFila, contenidoFila] of xml.matchAll(/<row\b([^>]*)>([\s\S]*?)<\/row>/g)) {
      const numeroFila = Number(leerAtributos(atributosFila).r) - 1;
      const fila = [];
      for (const [, atributosCelda, contenidoCelda = ''] of contenidoFila.matchAll(/<c\b([^>]*?)(?:\/>|>([\s\S]*?)<\/c>)/g)) {
        const { r: referencia, t: tipo, s: estilo } = leerAtributos(atributosCelda);
        const valorCrudo = (contenidoCelda.match(/<v>([^<]*)<\/v>/) || [])[1];
        let valor = null;
        if (tipo === 's') valor = textosCompartidos[Number(valorCrudo)] ?? null;
        else if (tipo === 'inlineStr') valor = extraerTextoRico(contenidoCelda);
        else if (tipo === 'str' || tipo === 'e') valor = valorCrudo != null ? decodificarEntidades(valorCrudo) : null;
        else if (tipo === 'b') valor = valorCrudo === '1';
        else if (valorCrudo != null && valorCrudo !== '') {
          const numero = Number(valorCrudo);
          valor = estilosFecha.has(Number(estilo)) ? convertirSerialAFecha(numero, usaSistema1904) : numero;
        }
        if (valor !== null && valor !== '') fila[indiceColumna(referencia)] = valor;
      }
      filas[numeroFila] = fila;
    }
    return Array.from(filas, (fila) => fila || []);
  }

  function resolverRutaHoja(destino) {
    if (destino.startsWith('/')) return destino.slice(1);
    return `xl/${destino.replace(/^\.\//, '')}`;
  }

  async function leerLibro(buffer) {
    if (esFormatoOfficeAntiguo(buffer)) {
      throw new FormatoNoSoportadoError('Es un Excel antiguo (.xls). Ábralo en Excel y use "Guardar como" .xlsx');
    }
    const xmlLibro = await leerTextoEntrada(buffer, 'xl/workbook.xml');
    const xmlRelaciones = await leerTextoEntrada(buffer, 'xl/_rels/workbook.xml.rels');
    const textosCompartidos = leerTextosCompartidos(await leerTextoEntrada(buffer, 'xl/sharedStrings.xml', { esOpcional: true }));
    const estilosFecha = leerEstilosFecha(await leerTextoEntrada(buffer, 'xl/styles.xml', { esOpcional: true }));
    const usaSistema1904 = /<workbookPr\b[^>]*date1904="(1|true)"/.test(xmlLibro);
    const destinos = new Map([...xmlRelaciones.matchAll(/<Relationship\b([^>]*)\/?>/g)].map(([, atributos]) => {
      const { Id, Target } = leerAtributos(atributos);
      return [Id, Target];
    }));
    const hojas = [];
    for (const [, atributos] of xmlLibro.matchAll(/<sheet\b([^>]*)\/?>/g)) {
      const datosHoja = leerAtributos(atributos);
      const destino = destinos.get(datosHoja['r:id']);
      if (!destino) continue;
      const xmlHoja = await leerTextoEntrada(buffer, resolverRutaHoja(destino), { esOpcional: true });
      if (xmlHoja) hojas.push({ nombre: datosHoja.name, filas: leerHoja(xmlHoja, textosCompartidos, estilosFecha, usaSistema1904) });
    }
    return hojas;
  }

  ns.LectorXlsx = { leerLibro };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
