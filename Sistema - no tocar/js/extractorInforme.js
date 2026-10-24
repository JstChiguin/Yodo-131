(function (ns) {
  const PATRON_INICIO_INFORME = /^NOMBRE\s*:/i;
  const PATRON_LINEA_IGNORADA = /^(MINISTERIO DE SALUD|SERVICIO DE SALUD|CENTRO DE COSTO|HOSPITAL DR|HOSPITAL CARLOS|ENDOCRINOLOG[IÍ]A$|\.+$)/i;
  const PATRON_LINEA_ESTRUCTURADA = /^(Ecotomograf|ECO\b|PAAF|[UÚ]ltima TSH|TSH\b|Cintigrama|Biopsia|Antecedentes\s+(familiares|de patolog|patolog)|M[oó]rbidos|Estadificaci|M[eé]dico tratante|TAC\b|Disecci|Laboratorio|\d{1,2}\/\d{1,2}\/\d{2,4}\s*:)/i;
  const MESES = ['enero', 'febrero', 'marzo', 'abril', 'mayo', 'junio', 'julio', 'agosto', 'septiembre', 'octubre', 'noviembre', 'diciembre'];
  const PATRON_MES = new RegExp(`^\\s*(de\\s+)?(${MESES.join('|')})`, 'i');

  function dividirEnInformes(parrafos) {
    const indicesInicio = parrafos.map((linea, indice) => (PATRON_INICIO_INFORME.test(linea) ? indice : -1)).filter((indice) => indice >= 0);
    return indicesInicio.map((inicio, posicion) => {
      const fin = indicesInicio[posicion + 1] ?? parrafos.length;
      return parrafos.slice(inicio, fin).filter((linea) => !PATRON_LINEA_IGNORADA.test(linea));
    });
  }

  function textoTrasEtiqueta(lineas, patronEtiqueta) {
    for (const linea of lineas) {
      const coincidencia = linea.match(patronEtiqueta);
      if (coincidencia) return linea.slice(coincidencia.index + coincidencia[0].length).trim();
    }
    return null;
  }

  function convertirNumero(textoNumero) {
    if (textoNumero == null) return null;
    const normalizado = String(textoNumero).replace(',', '.');
    const numero = Number(normalizado);
    return Number.isFinite(numero) ? numero : null;
  }

  function normalizarRut(textoRut) {
    if (!textoRut) return null;
    const limpio = textoRut.replace(/[^0-9kK]/g, '').toUpperCase();
    if (limpio.length < 7) return null;
    return `${limpio.slice(0, -1)}-${limpio.slice(-1)}`;
  }

  function capitalizarNombre(nombre) {
    return nombre.toLowerCase().replace(/\s+/g, ' ').trim()
      .replace(/(^|\s)(\p{L})/gu, (_, espacio, letra) => espacio + letra.toUpperCase());
  }

  function interpretarFecha(textoFecha) {
    const resultado = { fechaIso: null, anio: null, mes: null };
    if (!textoFecha) return resultado;
    const partes = textoFecha.replace(/\s+/g, '').split('/');
    const anio = Number((partes[2] || '').replace(/\D/g, ''));
    const mes = Number((partes[1] || '').replace(/\D/g, ''));
    const dia = Number((partes[0] || '').replace(/\D/g, ''));
    if (anio >= 1990 && anio <= 2100) resultado.anio = anio;
    if (resultado.anio && mes >= 1 && mes <= 12) resultado.mes = mes;
    if (resultado.mes && dia >= 1 && dia <= 31) {
      resultado.fechaIso = `${anio}-${String(mes).padStart(2, '0')}-${String(dia).padStart(2, '0')}`;
    }
    return resultado;
  }

  function extraerValorLaboratorio(texto, patronEtiqueta) {
    const patronGlobal = new RegExp(patronEtiqueta.source, 'gi');
    let coincidencia;
    while ((coincidencia = patronGlobal.exec(texto)) !== null) {
      const inicio = coincidencia.index + coincidencia[0].length;
      const resto = texto.slice(inicio, inicio + 40);
      const valor = resto.match(/^[^0-9<>\n;]{0,30}?(<|>|menos de|menor a|mayor a)?\s*(\d+(?:[.,]\d+)?)/i);
      if (valor) {
        const comparador = valor[1] ? (/^(<|menos|menor)/i.test(valor[1]) ? '<' : '>') : '';
        return { valor: convertirNumero(valor[2]), comparador };
      }
    }
    return null;
  }

  function extraerTiroglobulina(texto) {
    const valores = [];
    let enAscenso = false;
    let mencionada = false;
    const patron = /(?<!anti\s?)(?<!anti-)tiroglobulina/gi;
    let coincidencia;
    while ((coincidencia = patron.exec(texto)) !== null) {
      mencionada = true;
      let ventana = texto.slice(coincidencia.index + coincidencia[0].length, coincidencia.index + coincidencia[0].length + 100);
      const corte = ventana.search(/anticuerpo|antitiroglobulina|\bac\b|\bTSH\b|;|\n|\.\s/i);
      if (corte >= 0) ventana = ventana.slice(0, corte);
      if (/ascenso|aumenta|elevad/i.test(ventana)) enAscenso = true;
      const patronNumero = /(\d+(?:[.,]\d+)?)/g;
      let numero;
      while ((numero = patronNumero.exec(ventana)) !== null) {
        const antes = ventana.slice(Math.max(0, numero.index - 4), numero.index);
        const despues = ventana.slice(numero.index + numero[0].length);
        const esAnio = /^(19|20)\d\d$/.test(numero[0]);
        const esFecha = /^\//.test(despues) || /\/$/.test(antes) || /\bel\s$/i.test(antes) || PATRON_MES.test(despues);
        if (!esAnio && !esFecha) valores.push(convertirNumero(numero[0]));
      }
    }
    const valoresValidos = valores.filter((valor) => valor !== null);
    return { valorMaximo: valoresValidos.length ? Math.max(...valoresValidos) : null, enAscenso, mencionada };
  }

  function extraerAnticuerpo(texto, patronEtiqueta) {
    const coincidencia = texto.match(patronEtiqueta);
    if (!coincidencia) return null;
    let ventana = texto.slice(coincidencia.index + coincidencia[0].length, coincidencia.index + coincidencia[0].length + 60);
    const corte = ventana.search(/;|\n|\bTSH\b/i);
    if (corte >= 0) ventana = ventana.slice(0, corte);
    if (/no aparece|sin dato|pendiente/i.test(ventana)) return null;
    const valor = ventana.match(/(menos de|menor a|<|>|mayor a)?\s*(\d+(?:[.,]\d+)?)/i);
    if (!valor) return null;
    const comparador = valor[1] ? (/^(<|menos|menor)/i.test(valor[1]) ? '<' : '>') : '';
    return { valor: convertirNumero(valor[2]), comparador };
  }

  function evaluarPresencia(textoCampo) {
    if (textoCampo == null) return null;
    const sinRelleno = textoCampo.replace(/[.…\s]/g, '');
    if (!sinRelleno || /^-+$/.test(sinRelleno)) return null;
    if (/\d|cajetilla|ocasional|hasta|fum|exagerad|exceso|moderad|diario|semana|leve|marihuana|coca[ií]na/i.test(textoCampo)) return true;
    if (/^\(\s*-\s*\)|^no\b|^niega|^\(-/i.test(sinRelleno)) return false;
    return true;
  }

  function extraerCampoEnLinea(texto, patronEtiqueta) {
    const coincidencia = texto.match(patronEtiqueta);
    if (!coincidencia) return null;
    const resto = texto.slice(coincidencia.index + coincidencia[0].length);
    return resto.split(/;|\n|\bOH\s*:|Drogas\s*:|Alergias?\s*:/i)[0].trim();
  }

  function extraerDosis(lineas) {
    for (const linea of lineas) {
      if (!/decid/i.test(linea)) continue;
      const normalizada = linea.replace(/(\d)[oO](?=\s*m\s*C)/g, '$10');
      const dosis = normalizada.match(/(\d{1,3})\s*m\s*C(?:i)?\b/i);
      if (dosis) return Number(dosis[1]);
    }
    return null;
  }

  function convertirBethesdaRomano(valor) {
    const romanos = { 1: 'I', 2: 'II', 3: 'III', 4: 'IV', 5: 'V', 6: 'VI' };
    return /^\d$/.test(valor) ? romanos[valor] || null : valor.toUpperCase();
  }

  function extraerInforme(lineas, nombreArchivo) {
    const textoCompleto = lineas.join('\n');
    const lineaMorbidos = lineas.find((linea) => /^M[oó]rbidos/i.test(linea)) || '';
    const lineasClinicas = lineas.filter((linea) => !/^M[oó]rbidos|^Antecedentes\s+(familiares|de patolog|patolog)/i.test(linea));
    const textoClinico = lineasClinicas.join('\n');

    const indiceFecha = lineas.findIndex((linea) => /^FECHA\s*:/i.test(linea));
    const lineasNarrativa = [];
    for (let indice = indiceFecha + 1; indice > 0 && indice < lineas.length; indice++) {
      if (PATRON_LINEA_ESTRUCTURADA.test(lineas[indice])) break;
      lineasNarrativa.push(lineas[indice]);
    }

    const lineasEco = lineas.filter((linea) => /Ecotomograf[ií]a[^:]{0,30}:|^ECO\b/i.test(linea));
    const textoEco = lineasEco.join(' ');
    const tirads = (textoEco.match(/ti-?\s?rads\s*(\d)/i) || textoClinico.match(/ti-?\s?rads\s*(\d)/i) || [])[1];
    const bethesda = (textoClinico.match(/Be\w{0,3}sda\s*([IVX]{1,3}|[1-6])\b/i) || [])[1];
    const estadificacion = textoTrasEtiqueta(lineas, /^Estadificaci[oó]n\s*:/i);
    const riesgo = (textoClinico.match(/riesgo\s+(bajo|intermedio|alto)|(bajo|intermedio|alto)\s+riesgo/i) || []).slice(1).find(Boolean);

    return {
      archivo: nombreArchivo,
      nombrePaciente: capitalizarNombre(textoTrasEtiqueta(lineas, /^NOMBRE\s*:/i) || ''),
      rut: normalizarRut(textoTrasEtiqueta(lineas.slice(0, indiceFecha >= 0 ? indiceFecha + 1 : 4), /^RUT\s*:/i)),
      edad: convertirNumero((textoTrasEtiqueta(lineas, /^EDAD\s*:/i) || '').match(/\d{1,3}/)?.[0]),
      ...interpretarFecha(textoTrasEtiqueta(lineas, /^FECHA\s*:/i)),
      textoMotivo: lineasNarrativa.join(' ').replace(/-{3,}/g, ' ').trim(),
      textoEco,
      tirads: tirads ? Number(tirads) : null,
      ecoConNodulos: lineasEco.length ? /n[oó]dul|imagen\s+s[oó]lida|lesi[oó]n\s+nodular|masa\s+s[oó]lida/i.test(textoEco.replace(/sin\s+n[oó]dul\w*/gi, '')) : null,
      ecoConAdenopatias: lineasEco.length ? /adenopat|linfonodos?\s+(sospechos|prominent|aumentad)/i.test(textoEco.replace(/sin\s+adenopat\w*/gi, '')) : null,
      bethesda: bethesda ? convertirBethesdaRomano(bethesda) : null,
      textoPaaf: textoTrasEtiqueta(lineas, /^PAAF[^:]*:/i),
      textoBiopsia: lineas.filter((linea) => /^Biopsia/i.test(linea)).join(' '),
      textoCintigrama: textoTrasEtiqueta(lineas, /Cintigrama[^:]*:/i),
      estadificacion: estadificacion || null,
      riesgo: riesgo ? riesgo.toLowerCase() : null,
      tsh: extraerValorLaboratorio(textoClinico, /\bTSH\b\s*[:=]?/i),
      t4l: extraerValorLaboratorio(textoClinico, /\bT4\s*L\b\s*[:=]/i),
      t3: extraerValorLaboratorio(textoClinico, /\bT3\b\s*[:=]/i),
      trab: extraerValorLaboratorio(textoClinico, /\bTra[bB]\b\s*[:=]?/i),
      tiroglobulina: extraerTiroglobulina(textoClinico),
      antiTiroglobulina: extraerAnticuerpo(textoClinico, /anti\s*-?\s*tiroglobulina|anti\s*-?\s*tg\b/i),
      antiTpo: extraerAnticuerpo(textoClinico, /anti\s*-?\s*tpo|\batpo\b|anti\s*-?\s*peroxidasa/i),
      textoAntecedentesFamiliares: textoTrasEtiqueta(lineas, /^Antecedentes[^:]*(familia|tiroidea)[^:]*:/i),
      textoMorbidos: (lineaMorbidos.match(/M[oó]rbidos\s*:\s*(.*?)(?=;\s*(Alergia|Cigarrillo|OH)|$)/i) || [])[1]?.trim() || null,
      textoCigarrillo: extraerCampoEnLinea(textoCompleto, /Cigarrillo[^:]*:/i),
      textoAlcohol: extraerCampoEnLinea(textoCompleto, /\bOH\s*:/) ?? (/Cigarrillo y\s+OH\s*:/i.test(textoCompleto) ? extraerCampoEnLinea(textoCompleto, /Cigarrillo[^:]*:/i) : null),
      textoDrogas: extraerCampoEnLinea(textoCompleto, /Drogas\s*:/i),
      textoMedicoTratante: textoTrasEtiqueta(lineas, /^M[eé]dico tratante\s*:/i),
      tieneDatosGinecologicos: lineas.some((linea) => /^(Gesta|G)\s*:|^FUR\s*:|^MAC\s*:|^TRH\s*:/i.test(linea)),
      dosisMci: extraerDosis(lineas),
      rechazoYodo: /rechaza\s+(la\s+)?yodo|no acude a recibirlo/i.test(textoCompleto),
      textoClinico,
      textoNarrativa: lineasNarrativa.join(' ')
    };
  }

  ns.ExtractorInforme = { dividirEnInformes, extraerInforme, evaluarPresencia, interpretarFecha, normalizarRut };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
