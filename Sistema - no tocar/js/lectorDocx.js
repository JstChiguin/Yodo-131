(function (ns) {
  const { leerTextoEntrada, esFormatoOfficeAntiguo, decodificarEntidades, FormatoNoSoportadoError } = ns.LectorZip;

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
    if (esFormatoOfficeAntiguo(buffer)) {
      throw new FormatoNoSoportadoError('Es un Word antiguo (.doc). Ábralo en Word y use "Guardar como" .docx');
    }
    return extraerParrafos(await leerTextoEntrada(buffer, 'word/document.xml'));
  }

  ns.LectorDocx = { leerParrafosDocx };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
