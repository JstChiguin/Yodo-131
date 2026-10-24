(function (ns) {
  const CLAVE_ALMACEN = 'lectorYodo131.errores';
  const MAXIMO_ERRORES = 200;
  const erroresRegistrados = [];

  function cargarErroresGuardados() {
    try {
      const guardados = JSON.parse(globalThis.localStorage?.getItem(CLAVE_ALMACEN) || '[]');
      if (Array.isArray(guardados)) erroresRegistrados.push(...guardados.slice(-MAXIMO_ERRORES));
    } catch (errorLectura) {
      console.warn('[RegistroErrores] No se pudo leer el historial local', errorLectura);
    }
  }

  function guardarErrores() {
    try {
      globalThis.localStorage?.setItem(CLAVE_ALMACEN, JSON.stringify(erroresRegistrados.slice(-MAXIMO_ERRORES)));
    } catch (errorEscritura) {
      console.warn('[RegistroErrores] No se pudo guardar el historial local', errorEscritura);
    }
  }

  function registrarError({ componente, metodo, error, contexto }) {
    try {
      const registro = {
        fechaHora: new Date().toISOString(),
        componente: componente || 'desconocido',
        metodo: metodo || 'desconocido',
        tipoExcepcion: error?.name || typeof error,
        mensaje: String(error?.message ?? error).slice(0, 500),
        stackTrace: String(error?.stack || '').split('\n').slice(0, 6).join('\n'),
        contexto: contexto ? String(contexto).slice(0, 200) : '',
        versionAplicacion: ns.VERSION_APLICACION || ''
      };
      erroresRegistrados.push(registro);
      if (erroresRegistrados.length > MAXIMO_ERRORES) erroresRegistrados.shift();
      console.error(`[${registro.componente}.${registro.metodo}]`, error);
      guardarErrores();
      return registro;
    } catch (errorInterno) {
      console.error('[RegistroErrores] Falla del registro', errorInterno, error);
      return null;
    }
  }

  function obtenerErrores() {
    return [...erroresRegistrados];
  }

  function limpiarErrores() {
    erroresRegistrados.length = 0;
    guardarErrores();
  }

  cargarErroresGuardados();
  ns.RegistroErrores = { registrarError, obtenerErrores, limpiarErrores };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
