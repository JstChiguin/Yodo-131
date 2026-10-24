(function (ns) {
  const EXTENSIONES_WORD = /\.(docx|doc)$/i;
  const EXTENSIONES_EXCEL = /\.(xlsx|xls)$/i;

  function esArchivoTemporal(nombreArchivo) {
    return nombreArchivo.split('/').pop().startsWith('~$');
  }

  function esArchivoWordValido(nombreArchivo) {
    return EXTENSIONES_WORD.test(nombreArchivo) && !esArchivoTemporal(nombreArchivo);
  }

  function esArchivoExcelValido(nombreArchivo) {
    return EXTENSIONES_EXCEL.test(nombreArchivo) && !esArchivoTemporal(nombreArchivo);
  }

  function esArchivoSoportado(nombreArchivo) {
    return esArchivoWordValido(nombreArchivo) || esArchivoExcelValido(nombreArchivo);
  }

  async function leerArchivoWord(archivo) {
    const { LectorDocx, ExtractorInforme } = ns;
    const parrafos = await LectorDocx.leerParrafosDocx(await archivo.obtenerBuffer());
    const bloquesInforme = ExtractorInforme.dividirEnInformes(parrafos);
    return bloquesInforme.map((lineas, indice) => {
      const informe = ExtractorInforme.extraerInforme(lineas, archivo.nombre);
      informe.indiceEnArchivo = bloquesInforme.length > 1 ? indice + 1 : null;
      if (informe.indiceEnArchivo) informe.referencia = `${archivo.nombre} › informe ${informe.indiceEnArchivo}`;
      return informe;
    });
  }

  async function leerArchivoExcel(archivo) {
    const hojas = await ns.LectorXlsx.leerLibro(await archivo.obtenerBuffer());
    return ns.ExtractorExcel.extraerLibro(hojas, archivo.nombre);
  }

  async function procesarArchivos(archivos, alAvanzar) {
    const lectura = {
      registrosLeidos: [], registrosCancer: [], documentosNoLeidos: [],
      conteo: { archivosWord: 0, archivosExcel: 0, informesWord: 0, filasExcelYodo: 0, filasRegistroCancer: 0 }
    };
    const archivosSoportados = archivos.filter((archivo) => esArchivoSoportado(archivo.nombre));
    for (let indice = 0; indice < archivosSoportados.length; indice++) {
      const archivo = archivosSoportados[indice];
      try {
        if (esArchivoWordValido(archivo.nombre)) {
          lectura.conteo.archivosWord += 1;
          const informes = await leerArchivoWord(archivo);
          if (!informes.length) lectura.documentosNoLeidos.push({ archivo: archivo.nombre, motivo: 'No tiene el formato de informe (falta la línea "NOMBRE:")' });
          lectura.conteo.informesWord += informes.length;
          lectura.registrosLeidos.push(...informes);
        } else {
          lectura.conteo.archivosExcel += 1;
          const { registrosYodo, registrosCancer, hojasOmitidas } = await leerArchivoExcel(archivo);
          lectura.conteo.filasExcelYodo += registrosYodo.length;
          lectura.conteo.filasRegistroCancer += registrosCancer.length;
          lectura.registrosLeidos.push(...registrosYodo);
          lectura.registrosCancer.push(...registrosCancer);
          lectura.documentosNoLeidos.push(...hojasOmitidas);
        }
      } catch (error) {
        const esFormato = error instanceof ns.LectorZip.FormatoNoSoportadoError;
        if (!esFormato) {
          ns.RegistroErrores.registrarError({ componente: 'ProcesadorInformes', metodo: 'procesarArchivos', error, contexto: `archivo ${indice + 1} de ${archivosSoportados.length}` });
        }
        lectura.documentosNoLeidos.push({ archivo: archivo.nombre, motivo: esFormato ? error.message : 'No se pudo leer el archivo' });
      }
      alAvanzar?.(indice + 1, archivosSoportados.length);
    }
    lectura.cantidadArchivos = archivosSoportados.length;
    return lectura;
  }

  function completarTratamiento(tratamiento) {
    const { Clasificador } = ns;
    tratamiento.grupoEdad = Clasificador.agruparEdad(tratamiento.edad);
    tratamiento.grupoDosis = tratamiento.rechazoYodo && tratamiento.dosisMci == null ? 'Rechazó / no recibió' : Clasificador.agruparDosis(tratamiento.dosisMci);
    const camposEnConflicto = new Set(tratamiento.conflictos.map((conflicto) => conflicto.campo));
    const equivalencias = { RUT: 'rut', edad: 'edad', fecha: 'fecha', médico: 'medicoTratante', dosis: 'dosisMci', sexo: 'sexo' };
    tratamiento.camposFaltantes = Clasificador.listarCamposFaltantes(tratamiento).filter((campo) => !camposEnConflicto.has(equivalencias[campo]));
    return tratamiento;
  }

  function construirResultado(lectura, umbrales) {
    const { Clasificador, FusionTratamientos } = ns;
    const clasificados = lectura.registrosLeidos.map((registro) => Clasificador.clasificarInforme(registro, umbrales));
    const fusion = FusionTratamientos.fusionar(clasificados, lectura.registrosCancer);
    const medicosUnificados = Clasificador.unificarMedicos(fusion.tratamientos);
    fusion.tratamientos.forEach((tratamiento, indice) => {
      tratamiento.medicoTratante = medicosUnificados[indice].medicoTratante;
      completarTratamiento(tratamiento);
    });
    return {
      ...fusion,
      documentosNoLeidos: lectura.documentosNoLeidos,
      conteo: lectura.conteo,
      cantidadArchivos: lectura.cantidadArchivos
    };
  }

  ns.ProcesadorInformes = { procesarArchivos, construirResultado, esArchivoSoportado };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
