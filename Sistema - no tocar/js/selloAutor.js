(function (ns) {
  const TEXTO_SELLO = 'by Chiguin';
  const ID_FIRMA = 'firma-autor';

  function obtenerTextoVersion() {
    return `v${ns.VERSION_APLICACION || '0.0.0'}`;
  }

  function crearFirma() {
    const firma = document.createElement('div');
    firma.id = ID_FIRMA;
    firma.className = 'firma-autor';
    firma.setAttribute('aria-hidden', 'true');
    const sello = document.createElement('span');
    sello.className = 'sello-autor';
    sello.textContent = TEXTO_SELLO;
    const version = document.createElement('span');
    version.className = 'sello-version';
    version.textContent = obtenerTextoVersion();
    firma.append(sello, version);
    return firma;
  }

  function estaVisible(nodo) {
    if (!nodo || nodo.hasAttribute('style') || nodo.hidden) return false;
    const estilo = getComputedStyle(nodo);
    return estilo.display !== 'none' && estilo.visibility !== 'hidden' && Number(estilo.opacity) > 0.2;
  }

  function firmaEstaIntacta(firma) {
    if (!firma?.isConnected || firma.className !== 'firma-autor' || firma.children.length !== 2) return false;
    const [sello, version] = firma.children;
    return sello.className === 'sello-autor' && sello.textContent === TEXTO_SELLO
      && version.className === 'sello-version' && version.textContent === obtenerTextoVersion()
      && [firma, sello, version].every(estaVisible);
  }

  function restaurarFirma(contenedorRaiz) {
    const actual = document.getElementById(ID_FIRMA);
    if (firmaEstaIntacta(actual)) return true;
    actual?.remove();
    const firmaNueva = crearFirma();
    contenedorRaiz.appendChild(firmaNueva);
    return firmaEstaIntacta(firmaNueva);
  }

  function avisarFirmaNoVisible() {
    console.warn('[SelloAutor] La firma by Chiguin no queda visible ni después de recrearla; revise los estilos. La aplicación sigue funcionando.');
  }

  function montarSello(contenedorRaiz) {
    if (!restaurarFirma(contenedorRaiz)) {
      avisarFirmaNoVisible();
      return;
    }
    const observador = new MutationObserver(() => {
      if (restaurarFirma(contenedorRaiz)) return;
      observador.disconnect();
      avisarFirmaNoVisible();
    });
    observador.observe(contenedorRaiz, { childList: true, subtree: true, characterData: true, attributes: true, attributeFilter: ['style', 'class', 'hidden', 'id'] });
  }

  ns.SelloAutor = { montarSello };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
