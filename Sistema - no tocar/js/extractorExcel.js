(function (ns) {
  const { normalizarRut, interpretarFecha, capitalizarNombre, convertirNumero } = ns.ExtractorInforme;
  const PATRON_RUT = /^\s*\d{1,2}\.?\d{2,3}\.?\d{3}\s*[-.]\s*[\dkK]\s*$/;
  const PATRON_MEDICO = /^\s*dra?\b\.?\s*\S/i;
  const PATRON_DIAGNOSTICO = /carcinoma\s+(papilar|folicular|medular|anapl)|c[aá]ncer\s+(de(l)?\s+)?tiroid|^\s*ca\.?\s|hipertiroid|bocio|tiroiditis|graves|rastreo/i;

  const ENCABEZADOS_YODO = {
    paciente: ['PACIENTE'],
    rut: ['RUT'],
    edad: ['EDAD'],
    sexo: ['F/M', 'M/F', 'SEXO'],
    fechaSolicitud: ['FECHA DE SOLICITUD'],
    solicitudEnviada: ['SOLICITUD ENVIADA'],
    diagnostico: ['DIAGNOSTICO'],
    medico: ['MEDICO TRATANTE'],
    dosis: ['DOSIS'],
    antecedentesFamiliares: ['ANT.FAMILIARES', 'ANT. FAMILIARES', 'ANTECEDENTES FAMILIARES'],
    morbidos: ['MORBIDOS'],
    cigarrillo: ['CIGARRILLO'],
    alcohol: ['OH'],
    drogas: ['DROGAS']
  };

  function normalizarEncabezado(texto) {
    return String(texto ?? '').normalize('NFD').replace(/[̀-ͯ]/g, '').toUpperCase().replace(/\s+/g, ' ').trim();
  }

  function esFecha(valor) {
    return Boolean(valor && typeof valor === 'object' && valor.fecha);
  }

  function esTexto(valor) {
    return typeof valor === 'string' && valor.trim() !== '';
  }

  function textoCelda(valor) {
    if (valor == null) return null;
    if (esFecha(valor)) return valor.fecha;
    const texto = String(valor).replace(/\s+/g, ' ').trim();
    return texto || null;
  }

  function obtenerFecha(valor) {
    if (esFecha(valor)) {
      const [anio, mes] = valor.fecha.split('-').map(Number);
      return anio >= 1990 && anio <= 2100 ? { fechaIso: valor.fecha, anio, mes } : null;
    }
    if (esTexto(valor) && /\d{1,2}\s*\/\s*\d{1,2}\s*\/\s*\d{4}/.test(valor)) return interpretarFecha(valor);
    return null;
  }

  function buscarFilaEncabezado(filas) {
    for (let indice = 0; indice < Math.min(filas.length, 6); indice++) {
      const encabezados = (filas[indice] || []).map(normalizarEncabezado);
      if (encabezados.includes('PACIENTE') && encabezados.some((encabezado) => encabezado.startsWith('DOSIS'))) return indice;
    }
    return -1;
  }

  function detectarTipoHoja(hoja) {
    if (buscarFilaEncabezado(hoja.filas) >= 0) return 'yodo';
    if (hoja.filas.some((fila) => esTexto(fila[0]) && /^C73/i.test(fila[0].trim()))) return 'cancer';
    return 'desconocida';
  }

  function mapearColumnas(filaEncabezado) {
    const encabezados = filaEncabezado.map(normalizarEncabezado);
    const columnas = {};
    for (const [campo, nombresPosibles] of Object.entries(ENCABEZADOS_YODO)) {
      const indice = encabezados.findIndex((encabezado) => nombresPosibles.some((nombre) => encabezado === nombre || encabezado.startsWith(nombre)));
      if (indice >= 0) columnas[campo] = indice;
    }
    return columnas;
  }

  function interpretarSexo(valor) {
    const texto = normalizarEncabezado(valor);
    if (texto === 'F' || texto.startsWith('FEM')) return 'F';
    if (texto === 'M' || texto.startsWith('MASC')) return 'M';
    return null;
  }

  function interpretarDosis(valor) {
    if (typeof valor === 'number') return valor > 0 && valor <= 500 ? valor : null;
    const numero = String(valor ?? '').match(/\d+(?:[.,]\d+)?/);
    return numero ? convertirNumero(numero[0]) : null;
  }

  function crearRegistroBase(referencia, nombrePaciente) {
    return {
      origen: 'Excel', archivo: referencia, referencia, nombrePaciente,
      textoMotivo: '', textoNarrativa: '', textoClinico: '', textoEco: '', textoPaaf: null, textoBiopsia: '', textoCintigrama: null,
      tirads: null, ecoConNodulos: null, ecoConAdenopatias: null, bethesda: null, estadificacion: null, riesgo: null,
      tsh: null, t4l: null, t3: null, trab: null, antiTiroglobulina: null, antiTpo: null,
      tiroglobulina: { valorMaximo: null, enAscenso: false, mencionada: false },
      tieneDatosGinecologicos: false, rechazoYodo: false, indiceEnArchivo: null
    };
  }

  function extraerFilaYodo(fila, columnas, referencia) {
    const valorEn = (campo) => (columnas[campo] != null ? fila[columnas[campo]] : null);
    const nombre = textoCelda(valorEn('paciente'));
    if (!nombre || /^\d+$/.test(nombre)) return null;

    const columnasConocidas = new Set(Object.entries(columnas).filter(([campo]) => !['diagnostico', 'medico', 'solicitudEnviada', 'fechaSolicitud'].includes(campo)).map(([, indice]) => indice));
    const textos = fila.map((valor, indice) => (esTexto(valor) && !columnasConocidas.has(indice) ? valor.trim() : null)).filter(Boolean);
    const medicoEnColumna = textoCelda(valorEn('medico'));
    const textoMedico = medicoEnColumna && PATRON_MEDICO.test(medicoEnColumna) ? medicoEnColumna : textos.find((texto) => PATRON_MEDICO.test(texto)) || medicoEnColumna;
    const diagnosticoEnColumna = textoCelda(valorEn('diagnostico'));
    const textoDiagnostico = diagnosticoEnColumna && !PATRON_MEDICO.test(diagnosticoEnColumna) && !esFecha(valorEn('diagnostico'))
      ? diagnosticoEnColumna
      : textos.find((texto) => texto !== nombre && PATRON_DIAGNOSTICO.test(texto) && !PATRON_MEDICO.test(texto)) || null;
    const fecha = obtenerFecha(valorEn('fechaSolicitud')) || obtenerFecha(valorEn('solicitudEnviada')) || fila.map(obtenerFecha).find(Boolean) || { fechaIso: null, anio: null, mes: null };
    const valorDosis = valorEn('dosis');
    const rutEnColumna = textoCelda(valorEn('rut'));

    return {
      ...crearRegistroBase(referencia, capitalizarNombre(nombre)),
      rut: normalizarRut(rutEnColumna && /\d/.test(rutEnColumna) ? rutEnColumna : fila.filter(esTexto).find((texto) => PATRON_RUT.test(texto))),
      edad: typeof valorEn('edad') === 'number' && valorEn('edad') > 0 && valorEn('edad') < 120 ? Math.round(valorEn('edad')) : null,
      ...fecha,
      sexoRegistrado: interpretarSexo(valorEn('sexo')),
      textoClinico: textoDiagnostico || '',
      textoDiagnosticoExcel: textoDiagnostico,
      textoMedicoTratante: textoMedico,
      dosisMci: interpretarDosis(valorDosis),
      esRastreo: /rastreo/i.test(String(valorDosis ?? '')),
      textoAntecedentesFamiliares: textoCelda(valorEn('antecedentesFamiliares')),
      textoMorbidos: textoCelda(valorEn('morbidos')),
      textoCigarrillo: textoCelda(valorEn('cigarrillo')),
      textoAlcohol: textoCelda(valorEn('alcohol')),
      textoDrogas: textoCelda(valorEn('drogas'))
    };
  }

  function extraerHojaYodo(hoja, nombreArchivo) {
    const indiceEncabezado = buscarFilaEncabezado(hoja.filas);
    const columnas = mapearColumnas(hoja.filas[indiceEncabezado]);
    const registros = [];
    for (let indice = indiceEncabezado + 1; indice < hoja.filas.length; indice++) {
      const registro = extraerFilaYodo(hoja.filas[indice] || [], columnas, `${nombreArchivo} › hoja ${hoja.nombre} › fila ${indice + 1}`);
      if (registro) registros.push(registro);
    }
    return registros;
  }

  function extraerHojaCancer(hoja, nombreArchivo) {
    const anioHoja = /^\d{4}$/.test(String(hoja.nombre).trim()) ? Number(hoja.nombre) : null;
    const registros = [];
    hoja.filas.forEach((fila, indice) => {
      if (!esTexto(fila[0]) || !/^C73/i.test(fila[0].trim())) return;
      const textoRut = esTexto(fila[5]) && PATRON_RUT.test(fila[5]) ? fila[5] : fila.filter(esTexto).find((texto) => PATRON_RUT.test(texto));
      const fecha = obtenerFecha(fila[13]) || fila.map(obtenerFecha).find(Boolean) || { fechaIso: null, anio: anioHoja, mes: null };
      registros.push({
        origen: 'Registro de cáncer',
        referencia: `${nombreArchivo} › hoja ${hoja.nombre} › fila ${indice + 1}`,
        rut: normalizarRut(textoRut),
        nombrePaciente: esTexto(fila[6]) ? capitalizarNombre(fila[6]) : '',
        sexoRegistrado: interpretarSexo(fila[8]),
        edad: typeof fila[12] === 'number' ? Math.round(fila[12]) : null,
        estadoVital: esTexto(fila[11]) ? fila[11].trim() : null,
        ...fecha,
        anio: fecha.anio || anioHoja
      });
    });
    return registros;
  }

  function extraerLibro(hojas, nombreArchivo) {
    const resultado = { registrosYodo: [], registrosCancer: [], hojasOmitidas: [] };
    for (const hoja of hojas) {
      const tipo = detectarTipoHoja(hoja);
      if (tipo === 'yodo') resultado.registrosYodo.push(...extraerHojaYodo(hoja, nombreArchivo));
      else if (tipo === 'cancer') resultado.registrosCancer.push(...extraerHojaCancer(hoja, nombreArchivo));
      else if (hoja.filas.some((fila) => fila.length)) resultado.hojasOmitidas.push({ archivo: `${nombreArchivo} › hoja ${hoja.nombre}`, motivo: 'No se reconoce como planilla de yodo 131 ni como registro de cáncer de tiroides (C73)' });
    }
    return resultado;
  }

  ns.ExtractorExcel = { extraerLibro, detectarTipoHoja };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
