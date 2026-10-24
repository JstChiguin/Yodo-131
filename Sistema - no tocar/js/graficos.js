(function (ns) {
  const ESPACIO_SVG = 'http://www.w3.org/2000/svg';
  const formatoNumero = new Intl.NumberFormat('es-CL');
  let lienzoMedicion = null;

  function medirTexto(texto, tamanoFuente = 12) {
    lienzoMedicion = lienzoMedicion || document.createElement('canvas').getContext('2d');
    lienzoMedicion.font = `${tamanoFuente}px "Segoe UI", system-ui, sans-serif`;
    return lienzoMedicion.measureText(texto).width;
  }

  function crearElementoSvg(etiqueta, atributos = {}, estilo = '') {
    const elemento = document.createElementNS(ESPACIO_SVG, etiqueta);
    for (const [nombre, valor] of Object.entries(atributos)) elemento.setAttribute(nombre, valor);
    if (estilo) elemento.setAttribute('style', estilo);
    return elemento;
  }

  function mostrarVacio(contenedor, mensaje = 'Sin datos en este período') {
    contenedor.textContent = '';
    const aviso = document.createElement('p');
    aviso.className = 'grafico-vacio';
    aviso.textContent = mensaje;
    contenedor.appendChild(aviso);
  }

  function recortarEtiqueta(texto, anchoMaximo) {
    if (medirTexto(texto) <= anchoMaximo) return texto;
    let recortado = texto;
    while (recortado.length > 4 && medirTexto(`${recortado}…`) > anchoMaximo) recortado = recortado.slice(0, -1);
    return `${recortado}…`;
  }

  function dibujarBarrasHorizontales(contenedor, filas, obtenerColor) {
    if (!filas.length) return mostrarVacio(contenedor);
    contenedor.textContent = '';
    const ancho = Math.max(280, contenedor.clientWidth);
    const altoFila = 30;
    const alto = filas.length * altoFila + 6;
    const anchoEtiquetas = Math.min(Math.max(...filas.map((fila) => medirTexto(fila.categoria))), ancho * 0.46);
    const anchoValores = Math.max(...filas.map((fila) => medirTexto(formatoNumero.format(fila.cantidad)))) + 10;
    const inicioBarras = anchoEtiquetas + 12;
    const anchoUtil = Math.max(20, ancho - inicioBarras - anchoValores);
    const maximo = Math.max(...filas.map((fila) => fila.cantidad), 1);
    const svg = crearElementoSvg('svg', { viewBox: `0 0 ${ancho} ${alto}`, width: '100%', role: 'img' });

    filas.forEach((fila, indice) => {
      const y = indice * altoFila + 4;
      const anchoBarra = Math.max(2, (fila.cantidad / maximo) * anchoUtil);
      const etiqueta = crearElementoSvg('text', { x: inicioBarras - 8, y: y + 15, 'text-anchor': 'end', 'font-size': 12 }, 'fill: var(--texto)');
      etiqueta.textContent = recortarEtiqueta(fila.categoria, anchoEtiquetas);
      const titulo = crearElementoSvg('title');
      titulo.textContent = `${fila.categoria}: ${formatoNumero.format(fila.cantidad)}`;
      const barra = crearElementoSvg('rect', { x: inicioBarras, y: y + 3, width: anchoBarra, height: altoFila - 12, rx: 4 }, `fill: ${obtenerColor(fila, indice)}`);
      barra.appendChild(titulo);
      const valor = crearElementoSvg('text', { x: inicioBarras + anchoBarra + 6, y: y + 15, 'font-size': 12 }, 'fill: var(--texto-suave); font-variant-numeric: tabular-nums');
      valor.textContent = formatoNumero.format(fila.cantidad);
      svg.append(etiqueta, barra, valor);
    });
    contenedor.appendChild(svg);
  }

  function dibujarBarrasApiladas(contenedor, filas, series) {
    if (!filas.length) return mostrarVacio(contenedor);
    contenedor.textContent = '';
    const ancho = Math.max(280, contenedor.clientWidth);
    const alto = 230;
    const margen = { arriba: 20, abajo: 26, izquierda: 8, derecha: 8 };
    const totales = filas.map((fila) => series.reduce((suma, serie) => suma + (fila[serie.clave] || 0), 0));
    const maximo = Math.max(...totales, 1);
    const anchoBanda = (ancho - margen.izquierda - margen.derecha) / filas.length;
    const anchoBarra = Math.min(36, anchoBanda * 0.62);
    const escalaY = (valor) => (valor / maximo) * (alto - margen.arriba - margen.abajo);
    const anchoEtiqueta = Math.max(...filas.map((fila) => medirTexto(fila.categoria))) + 6;
    const saltoEtiquetas = Math.max(1, Math.ceil(anchoEtiqueta / anchoBanda));
    const svg = crearElementoSvg('svg', { viewBox: `0 0 ${ancho} ${alto}`, width: '100%', role: 'img' });
    svg.appendChild(crearElementoSvg('line', { x1: margen.izquierda, x2: ancho - margen.derecha, y1: alto - margen.abajo, y2: alto - margen.abajo }, 'stroke: var(--borde-fuerte)'));

    filas.forEach((fila, indice) => {
      const x = margen.izquierda + indice * anchoBanda + (anchoBanda - anchoBarra) / 2;
      let base = alto - margen.abajo;
      for (const serie of series) {
        const cantidad = fila[serie.clave] || 0;
        if (!cantidad) continue;
        const altoTramo = escalaY(cantidad);
        const tramo = crearElementoSvg('rect', { x, y: base - altoTramo, width: anchoBarra, height: altoTramo }, `fill: ${serie.color}`);
        const titulo = crearElementoSvg('title');
        titulo.textContent = `${fila.categoria} · ${serie.nombre}: ${formatoNumero.format(cantidad)}`;
        tramo.appendChild(titulo);
        svg.appendChild(tramo);
        base -= altoTramo;
      }
      const total = crearElementoSvg('text', { x: x + anchoBarra / 2, y: base - 6, 'text-anchor': 'middle', 'font-size': 11 }, 'fill: var(--texto-suave); font-variant-numeric: tabular-nums');
      total.textContent = formatoNumero.format(totales[indice]);
      svg.appendChild(total);
      if (indice % saltoEtiquetas === 0) {
        const etiqueta = crearElementoSvg('text', { x: x + anchoBarra / 2, y: alto - 8, 'text-anchor': 'middle', 'font-size': 12 }, 'fill: var(--texto-suave)');
        etiqueta.textContent = fila.categoria;
        svg.appendChild(etiqueta);
      }
    });
    contenedor.appendChild(svg);
  }

  function dibujarProporciones(contenedor, grupos, colores) {
    contenedor.textContent = '';
    for (const [nombreGrupo, conteo] of Object.entries(grupos)) {
      const total = Object.values(conteo).reduce((suma, cantidad) => suma + cantidad, 0);
      const fila = document.createElement('div');
      fila.className = 'proporcion';
      const encabezado = document.createElement('div');
      encabezado.className = 'proporcion-encabezado';
      const nombre = document.createElement('span');
      nombre.textContent = nombreGrupo;
      const detalle = document.createElement('span');
      detalle.className = 'proporcion-detalle';
      detalle.textContent = Object.entries(conteo).map(([categoria, cantidad]) => `${categoria} ${formatoNumero.format(cantidad)}`).join(' · ');
      encabezado.append(nombre, detalle);
      const barra = document.createElement('div');
      barra.className = 'proporcion-barra';
      barra.setAttribute('role', 'img');
      barra.setAttribute('aria-label', `${nombreGrupo}: ${detalle.textContent}`);
      for (const [categoria, cantidad] of Object.entries(conteo)) {
        if (!cantidad || !total) continue;
        const tramo = document.createElement('div');
        tramo.style.width = `${(cantidad / total) * 100}%`;
        tramo.style.background = colores[categoria] || 'var(--sin-dato)';
        tramo.title = `${categoria}: ${cantidad}`;
        barra.appendChild(tramo);
      }
      fila.append(encabezado, barra);
      contenedor.appendChild(fila);
    }
  }

  function dibujarLeyenda(contenedor, elementos) {
    contenedor.textContent = '';
    for (const { nombre, color } of elementos) {
      const item = document.createElement('span');
      const muestra = document.createElement('i');
      muestra.style.background = color;
      item.append(muestra, document.createTextNode(nombre));
      contenedor.appendChild(item);
    }
  }

  ns.Graficos = { dibujarBarrasHorizontales, dibujarBarrasApiladas, dibujarProporciones, dibujarLeyenda, formatoNumero };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
