(function (ns) {
  const CLAVE_ALMACEN = 'lectorYodo131.umbrales';

  function validarUmbrales(umbralesPropuestos) {
    const porDefecto = ns.Clasificador.UMBRALES_POR_DEFECTO;
    const umbralesValidos = {};
    for (const clave of Object.keys(porDefecto)) {
      const valor = Number(umbralesPropuestos?.[clave]);
      umbralesValidos[clave] = Number.isFinite(valor) && valor > 0 ? valor : porDefecto[clave];
    }
    return umbralesValidos;
  }

  function obtenerUmbrales() {
    try {
      const guardados = globalThis.localStorage?.getItem(CLAVE_ALMACEN);
      return validarUmbrales(guardados ? JSON.parse(guardados) : null);
    } catch (error) {
      ns.RegistroErrores.registrarError({ componente: 'Configuracion', metodo: 'obtenerUmbrales', error });
      return { ...ns.Clasificador.UMBRALES_POR_DEFECTO };
    }
  }

  function guardarUmbrales(umbralesPropuestos) {
    const umbralesValidos = validarUmbrales(umbralesPropuestos);
    try {
      globalThis.localStorage?.setItem(CLAVE_ALMACEN, JSON.stringify(umbralesValidos));
      return { umbrales: umbralesValidos, guardadoPermanente: true };
    } catch (error) {
      ns.RegistroErrores.registrarError({ componente: 'Configuracion', metodo: 'guardarUmbrales', error });
      return { umbrales: umbralesValidos, guardadoPermanente: false };
    }
  }

  function restablecerUmbrales() {
    try {
      globalThis.localStorage?.removeItem(CLAVE_ALMACEN);
    } catch (error) {
      ns.RegistroErrores.registrarError({ componente: 'Configuracion', metodo: 'restablecerUmbrales', error });
    }
    return { ...ns.Clasificador.UMBRALES_POR_DEFECTO };
  }

  ns.Configuracion = { obtenerUmbrales, guardarUmbrales, restablecerUmbrales };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
