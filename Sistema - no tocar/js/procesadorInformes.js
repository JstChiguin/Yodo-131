(function (ns) {
  const EXTENSIONES_WORD = /\.(docx|doc)$/i;

  function esArchivoWordValido(nombreArchivo) {
    return EXTENSIONES_WORD.test(nombreArchivo) && !nombreArchivo.split('/').pop().startsWith('~$');
  }

  async function procesarDocumento(nombreArchivo, buffer, umbrales) {
    const { LectorDocx, ExtractorInforme, Clasificador } = ns;
    const parrafos = await LectorDocx.leerParrafosDocx(buffer);
    const bloquesInforme = ExtractorInforme.dividirEnInformes(parrafos);
    return bloquesInforme.map((lineas, indice) => {
      const informe = ExtractorInforme.extraerInforme(lineas, nombreArchivo);
      informe.indiceEnArchivo = bloquesInforme.length > 1 ? indice + 1 : null;
      return Clasificador.clasificarInforme(informe, umbrales);
    });
  }

  async function procesarArchivos(archivos, umbrales, alAvanzar) {
    const registros = [];
    const documentosNoLeidos = [];
    const archivosWord = archivos.filter((archivo) => esArchivoWordValido(archivo.nombre));
    for (let indice = 0; indice < archivosWord.length; indice++) {
      const archivo = archivosWord[indice];
      try {
        const informesDelArchivo = await procesarDocumento(archivo.nombre, await archivo.obtenerBuffer(), umbrales);
        if (informesDelArchivo.length === 0) {
          documentosNoLeidos.push({ archivo: archivo.nombre, motivo: 'No tiene el formato de informe (falta la línea "NOMBRE:")' });
        }
        registros.push(...informesDelArchivo);
      } catch (error) {
        const esFormato = error instanceof ns.LectorDocx.FormatoNoSoportadoError;
        if (!esFormato) {
          ns.RegistroErrores.registrarError({ componente: 'ProcesadorInformes', metodo: 'procesarArchivos', error, contexto: `archivo ${indice + 1} de ${archivosWord.length}` });
        }
        documentosNoLeidos.push({ archivo: archivo.nombre, motivo: esFormato ? error.message : 'No se pudo leer el archivo' });
      }
      alAvanzar?.(indice + 1, archivosWord.length);
    }
    return { registros: ns.Clasificador.unificarMedicos(registros), documentosNoLeidos, cantidadArchivos: archivosWord.length };
  }

  function reclasificar(registros, umbrales) {
    return ns.Clasificador.unificarMedicos(registros.map((registro) => ns.Clasificador.clasificarInforme(registro, umbrales)));
  }

  ns.ProcesadorInformes = { procesarArchivos, reclasificar, esArchivoWordValido };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
