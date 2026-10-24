(function (ns) {
  const TABLA_CRC = (() => {
    const tabla = new Uint32Array(256);
    for (let indice = 0; indice < 256; indice++) {
      let valor = indice;
      for (let bit = 0; bit < 8; bit++) valor = valor & 1 ? 0xedb88320 ^ (valor >>> 1) : valor >>> 1;
      tabla[indice] = valor >>> 0;
    }
    return tabla;
  })();

  function calcularCrc32(bytes) {
    let crc = 0xffffffff;
    for (const byte of bytes) crc = TABLA_CRC[(crc ^ byte) & 0xff] ^ (crc >>> 8);
    return (crc ^ 0xffffffff) >>> 0;
  }

  function crearZipSinCompresion(entradas) {
    const codificador = new TextEncoder();
    const partesLocales = [];
    const partesDirectorio = [];
    let desplazamiento = 0;
    for (const entrada of entradas) {
      const nombre = codificador.encode(entrada.nombre);
      const datos = codificador.encode(entrada.texto);
      const crc = calcularCrc32(datos);
      const cabecera = new DataView(new ArrayBuffer(30));
      cabecera.setUint32(0, 0x04034b50, true);
      cabecera.setUint16(4, 20, true);
      cabecera.setUint16(6, 0x0800, true);
      cabecera.setUint32(14, crc, true);
      cabecera.setUint32(18, datos.length, true);
      cabecera.setUint32(22, datos.length, true);
      cabecera.setUint16(26, nombre.length, true);
      partesLocales.push(new Uint8Array(cabecera.buffer), nombre, datos);

      const central = new DataView(new ArrayBuffer(46));
      central.setUint32(0, 0x02014b50, true);
      central.setUint16(4, 20, true);
      central.setUint16(6, 20, true);
      central.setUint16(8, 0x0800, true);
      central.setUint32(16, crc, true);
      central.setUint32(20, datos.length, true);
      central.setUint32(24, datos.length, true);
      central.setUint16(28, nombre.length, true);
      central.setUint32(42, desplazamiento, true);
      partesDirectorio.push(new Uint8Array(central.buffer), nombre);
      desplazamiento += 30 + nombre.length + datos.length;
    }
    const tamanoDirectorio = partesDirectorio.reduce((suma, parte) => suma + parte.length, 0);
    const fin = new DataView(new ArrayBuffer(22));
    fin.setUint32(0, 0x06054b50, true);
    fin.setUint16(8, entradas.length, true);
    fin.setUint16(10, entradas.length, true);
    fin.setUint32(12, tamanoDirectorio, true);
    fin.setUint32(16, desplazamiento, true);
    return new Blob([...partesLocales, ...partesDirectorio, new Uint8Array(fin.buffer)], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    });
  }

  function escaparXml(texto) {
    return String(texto)
      .replace(/[\u0000-\u0008\u000b\u000c\u000e-\u001f]/g, '')
      .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
  }

  function letraColumna(indice) {
    let letra = '';
    for (let numero = indice + 1; numero > 0; numero = Math.floor((numero - 1) / 26)) {
      letra = String.fromCharCode(65 + ((numero - 1) % 26)) + letra;
    }
    return letra;
  }

  function generarHojaXml(filas, anchosColumna) {
    const columnasXml = anchosColumna.map((ancho, indice) => `<col min="${indice + 1}" max="${indice + 1}" width="${ancho}" customWidth="1"/>`).join('');
    const filasXml = filas.map((fila, indiceFila) => {
      const celdas = fila.map((valor, indiceColumna) => {
        if (valor == null || valor === '') return '';
        const referencia = `${letraColumna(indiceColumna)}${indiceFila + 1}`;
        const estilo = indiceFila === 0 ? ' s="1"' : '';
        if (typeof valor === 'number' && Number.isFinite(valor)) return `<c r="${referencia}"${estilo}><v>${valor}</v></c>`;
        return `<c r="${referencia}" t="inlineStr"${estilo}><is><t xml:space="preserve">${escaparXml(valor)}</t></is></c>`;
      }).join('');
      return `<row r="${indiceFila + 1}">${celdas}</row>`;
    }).join('');
    return '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
      + '<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">'
      + '<sheetViews><sheetView workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews>'
      + `<cols>${columnasXml}</cols><sheetData>${filasXml}</sheetData></worksheet>`;
  }

  function crearLibroXlsx(hojas) {
    const tipos = hojas.map((_, indice) => `<Override PartName="/xl/worksheets/hoja${indice + 1}.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`).join('');
    const relaciones = hojas.map((_, indice) => `<Relationship Id="rId${indice + 1}" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/hoja${indice + 1}.xml"/>`).join('');
    const listaHojas = hojas.map((hoja, indice) => `<sheet name="${escaparXml(hoja.nombre)}" sheetId="${indice + 1}" r:id="rId${indice + 1}"/>`).join('');
    const entradas = [
      { nombre: '[Content_Types].xml', texto: '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>' + tipos + '</Types>' },
      { nombre: '_rels/.rels', texto: '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>' },
      { nombre: 'xl/workbook.xml', texto: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>${listaHojas}</sheets></workbook>` },
      { nombre: 'xl/_rels/workbook.xml.rels', texto: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">${relaciones}<Relationship Id="rId${hojas.length + 1}" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/></Relationships>` },
      { nombre: 'xl/styles.xml', texto: '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><fonts count="2"><font><sz val="11"/><name val="Calibri"/></font><font><b/><sz val="11"/><name val="Calibri"/></font></fonts><fills count="3"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill><fill><patternFill patternType="solid"><fgColor rgb="FFE3ECEA"/></patternFill></fill></fills><borders count="1"><border/></borders><cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs><cellXfs count="2"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/><xf numFmtId="0" fontId="1" fillId="2" borderId="0" xfId="0" applyFont="1" applyFill="1"/></cellXfs><cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles></styleSheet>' },
      ...hojas.map((hoja, indice) => ({ nombre: `xl/worksheets/hoja${indice + 1}.xml`, texto: generarHojaXml(hoja.filas, hoja.anchos) }))
    ];
    return crearZipSinCompresion(entradas);
  }

  const textoSiNo = (valor) => (valor === true ? 'Sí' : valor === false ? 'No' : '');
  const textoElevado = (valor) => (valor === true ? 'Elevado' : valor === false ? 'Normal' : '');
  const textoMedicion = (medicion) => (medicion ? `${medicion.comparador}${String(medicion.valor).replace('.', ',')}` : '');

  const COLUMNAS_TRATAMIENTOS = [
    ['Fuentes', 50, (r) => r.fuentes.map((fuente) => `${fuente.origen}: ${fuente.referencia}`).join(' | ')],
    ['Origen', 12, (r) => (r.origenes.length > 1 ? 'Word y Excel' : r.origenes[0])],
    ['Tratamiento N° del paciente', 12, (r) => `${r.numeroTratamientoPaciente} de ${r.cantidadTratamientosPaciente}`],
    ['Datos que no coinciden', 34, (r) => [...new Set(r.conflictos.map((conflicto) => conflicto.nombreCampo))].join(', ')],
    ['Paciente', 28, (r) => r.nombrePaciente], ['RUT', 13, (r) => r.rut], ['Edad', 6, (r) => r.edad],
    ['Sexo', 6, (r) => r.sexo], ['Origen del sexo', 20, (r) => r.origenSexo],
    ['Fecha', 11, (r) => r.fechaIso || (r.anio ? String(r.anio) : '')], ['Año', 6, (r) => r.anio],
    ['Médico tratante', 24, (r) => r.medicoTratante], ['Dosis (mCi)', 10, (r) => r.dosisMci], ['Grupo dosis', 20, (r) => r.grupoDosis],
    ['Rechazó yodo', 12, (r) => textoSiNo(r.rechazoYodo)],
    ['Diagnóstico', 34, (r) => r.diagnostico], ['Tipo histológico', 24, (r) => r.histologia],
    ['Motivo (categoría)', 32, (r) => r.motivoCategoria], ['Motivo (texto)', 60, (r) => r.textoMotivo],
    ['Ecografía', 60, (r) => r.textoEco], ['TIRADS', 7, (r) => r.tirads], ['Eco con nódulos', 9, (r) => textoSiNo(r.ecoConNodulos)],
    ['Eco con adenopatías', 9, (r) => textoSiNo(r.ecoConAdenopatias)],
    ['PAAF', 40, (r) => r.textoPaaf], ['Bethesda', 9, (r) => r.bethesda], ['Biopsia', 60, (r) => r.textoBiopsia],
    ['Estadificación', 20, (r) => r.estadificacion], ['Riesgo', 11, (r) => r.riesgo], ['Cintigrama', 40, (r) => r.textoCintigrama],
    ['TSH', 8, (r) => textoMedicion(r.tsh)], ['T4L', 8, (r) => textoMedicion(r.t4l)], ['T3', 8, (r) => textoMedicion(r.t3)],
    ['TRAb', 8, (r) => textoMedicion(r.trab)], ['TRAb estado', 10, (r) => textoElevado(r.trabElevado)],
    ['Tiroglobulina (máx.)', 10, (r) => r.tiroglobulina?.valorMaximo], ['Tiroglobulina estado', 11, (r) => textoElevado(r.tiroglobulinaElevada)],
    ['Anti-tiroglobulina', 10, (r) => textoMedicion(r.antiTiroglobulina)], ['Anti-Tg estado', 10, (r) => textoElevado(r.antiTiroglobulinaElevado)],
    ['Anti-TPO', 9, (r) => textoMedicion(r.antiTpo)], ['Anti-TPO estado', 10, (r) => textoElevado(r.antiTpoElevado)],
    ['Antecedentes familiares', 28, (r) => r.textoAntecedentesFamiliares], ['Antecedente familiar', 10, (r) => textoSiNo(r.tieneAntecedenteFamiliar)],
    ['Mórbidos', 40, (r) => r.textoMorbidos], ['Cigarrillo', 24, (r) => r.textoCigarrillo], ['Fumador', 9, (r) => textoSiNo(r.esFumador)],
    ['OH', 20, (r) => r.textoAlcohol], ['Drogas', 18, (r) => r.textoDrogas],
    ['Diagnóstico en Excel', 30, (r) => r.textoDiagnosticoExcel],
    ['En registro de cáncer (C73)', 12, (r) => textoSiNo(r.pacienteConCancerRegistrado)],
    ['Datos que faltan', 24, (r) => r.camposFaltantes.join(', ')]
  ];

  function crearHojaResumen(estadisticas, resultado, descripcionPeriodo) {
    const { conteo } = resultado;
    const filas = [['Indicador', 'Categoría', 'Cantidad', 'Con dato', 'Total', 'Cobertura']];
    const porcentaje = (parte, total) => (total ? Math.round((parte / total) * 1000) / 10 : '');
    filas.push(
      ['Período', descripcionPeriodo], ['Versión del lector', `v${ns.VERSION_APLICACION || ''}`],
      ['Informes Word leídos', '', conteo.informesWord], ['Filas de Excel de yodo leídas', '', conteo.filasExcelYodo], ['Filas del registro de cáncer leídas', '', conteo.filasRegistroCancer],
      ['Tratamientos', '', estadisticas.cantidadTratamientos], ['Registros repetidos unidos', '', estadisticas.registrosUnidos],
      ['Pacientes únicos', '', estadisticas.cantidadPacientes], ['Pacientes con más de un tratamiento', '', estadisticas.pacientesConVariosTratamientos],
      ['Pacientes', 'Mujeres', estadisticas.pacientesMujeres], ['Pacientes', 'Hombres', estadisticas.pacientesHombres], ['Pacientes', 'Sexo sin dato o por revisar', estadisticas.pacientesSinSexo],
      ['Pacientes con cáncer de tiroides', '', estadisticas.cantidadPacientesCancer], ['Datos que no coinciden entre fuentes', '', estadisticas.cantidadConflictos]
    );
    if (estadisticas.edadPromedio != null) filas.push(['Edad promedio', '', Math.round(estadisticas.edadPromedio * 10) / 10, estadisticas.coberturaEdad.conDato, estadisticas.coberturaEdad.total, porcentaje(estadisticas.coberturaEdad.conDato, estadisticas.coberturaEdad.total)]);
    const agregarDistribucion = (indicador, distribucion) => distribucion.filas.forEach((fila) => filas.push([indicador, fila.categoria, fila.cantidad, distribucion.conDato, distribucion.total, porcentaje(distribucion.conDato, distribucion.total)]));
    agregarDistribucion('Diagnóstico', estadisticas.porDiagnostico);
    agregarDistribucion('Motivo de consulta', estadisticas.porMotivo);
    agregarDistribucion('Grupo de edad', estadisticas.porEdad);
    agregarDistribucion('Dosis', estadisticas.porDosis);
    agregarDistribucion('Tipo histológico (cáncer)', estadisticas.porHistologia);
    agregarDistribucion('Bethesda', estadisticas.porBethesda);
    agregarDistribucion('TIRADS', estadisticas.porTirads);
    agregarDistribucion('Origen de los datos', estadisticas.porOrigen);
    for (const grupo of [estadisticas.examenes, estadisticas.ecografia, estadisticas.antecedentes]) {
      for (const [nombre, conteoGrupo] of Object.entries(grupo)) {
        const total = Object.values(conteoGrupo).reduce((suma, cantidad) => suma + cantidad, 0);
        const conDato = total - conteoGrupo['Sin dato'];
        for (const [categoria, cantidad] of Object.entries(conteoGrupo)) {
          if (categoria !== 'Sin dato') filas.push([nombre, categoria, cantidad, conDato, total, porcentaje(conDato, total)]);
        }
      }
    }
    estadisticas.cancerPorAnio.forEach((fila) => filas.push(['Pacientes con cáncer de tiroides por año', fila.categoria, fila.mujeres + fila.hombres + fila.sinDato]));
    return { nombre: 'Resumen', filas, anchos: [34, 36, 10, 10, 10, 11] };
  }

  function exportarLibro(estadisticas, resultado, descripcionPeriodo) {
    const hojaTratamientos = {
      nombre: 'Tratamientos',
      filas: [COLUMNAS_TRATAMIENTOS.map(([titulo]) => titulo), ...estadisticas.tratamientos.map((tratamiento) => COLUMNAS_TRATAMIENTOS.map(([, , obtener]) => obtener(tratamiento) ?? ''))],
      anchos: COLUMNAS_TRATAMIENTOS.map(([, ancho]) => ancho)
    };
    const idsEnPeriodo = new Set(estadisticas.tratamientos.map((tratamiento) => tratamiento.idTratamiento));
    const hojaConflictos = {
      nombre: 'Por revisar',
      filas: [['Paciente', 'RUT', 'Fecha del tratamiento', 'Dato', 'Fuente', 'Valor', 'Ubicación'],
        ...resultado.conflictos.filter((conflicto) => idsEnPeriodo.has(conflicto.idTratamiento)).flatMap((conflicto) => conflicto.valores.map((valor) => [
          conflicto.nombrePaciente, conflicto.rut, conflicto.fechaIso || (conflicto.anio ? String(conflicto.anio) : ''), conflicto.nombreCampo, valor.origen, valor.valor, valor.referencia
        ]))],
      anchos: [28, 13, 12, 24, 26, 30, 50]
    };
    const hojaNoLeidos = {
      nombre: 'No leídos',
      filas: [['Archivo', 'Motivo'], ...resultado.documentosNoLeidos.map((documento) => [documento.archivo, documento.motivo])],
      anchos: [40, 70]
    };
    return crearLibroXlsx([hojaTratamientos, crearHojaResumen(estadisticas, resultado, descripcionPeriodo), hojaConflictos, hojaNoLeidos]);
  }

  ns.ExportadorExcel = { exportarLibro };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
