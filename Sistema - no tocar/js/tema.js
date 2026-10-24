(function (ns) {
  const TEMA_CLARO = 'claro';
  const TEMA_OSCURO = 'oscuro';

  function aplicarTema(tema) {
    document.documentElement.dataset.tema = tema;
  }

  function obtenerTemaActual() {
    return document.documentElement.dataset.tema === TEMA_OSCURO ? TEMA_OSCURO : TEMA_CLARO;
  }

  function alternarTema() {
    const temaNuevo = obtenerTemaActual() === TEMA_OSCURO ? TEMA_CLARO : TEMA_OSCURO;
    aplicarTema(temaNuevo);
    return temaNuevo;
  }

  aplicarTema(TEMA_CLARO);
  ns.Tema = { alternarTema, obtenerTemaActual, TEMA_OSCURO };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
