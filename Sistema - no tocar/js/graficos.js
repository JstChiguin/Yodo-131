(function (ns) {
  const ESPACIO_SVG = 'http://www.w3.org/2000/svg';
  const formatoNumero = new Intl.NumberFormat('es-CL');
  const formatoPorcentaje = new Intl.NumberFormat('es-CL', { style: 'percent', maximumFractionDigits: 0 });
  let lienzoMedicion = null;
  let globoDetalle = null;

  function medirTexto(texto, tamanoFuente = 13, peso = 400) {
    lienzoMedicion = lienzoMedicion || document.createElement('canvas').getContext('2d');
    lienzoMedicion.font = `${peso} ${tamanoFuente}px "Segoe UI", system-ui, sans-serif`;
    return lienzoMedicion.measureText(texto).width;
  }

  function describirCantidad(cantidad, unidad) {
    const unidadAjustada = cantidad === 1 ? unidad.replace(/^(\S+?)s\b/, '$1') : unidad;
    return `${formatoNumero.format(cantidad)} ${unidadAjustada}`;
  }

  function calcularPorcentaje(cantidad, total) {
    return total ? formatoPorcentaje.format(cantidad / total) : '—';
  }

  function crearElementoSvg(etiqueta, atributos = {}, estilo = '') {
    const elementoSvg = document.createElementNS(ESPACIO_SVG, etiqueta);
    for (const [nombre, valor] of Object.entries(atributos)) elementoSvg.setAttribute(nombre, valor);
    if (estilo) elementoSvg.setAttribute('style', estilo);
    return elementoSvg;
  }

  function crearTextoSvg(texto, atributos, estilo) {
    const elementoTexto = crearElementoSvg('text', atributos, estilo);
    elementoTexto.textContent = texto;
    return elementoTexto;
  }

  function obtenerGloboDetalle() {
    if (globoDetalle?.isConnected) return globoDetalle;
    globoDetalle = document.createElement('div');
    globoDetalle.className = 'globo-detalle';
    globoDetalle.setAttribute('role', 'tooltip');
    globoDetalle.hidden = true;
    document.body.appendChild(globoDetalle);
    return globoDetalle;
  }

  function llenarGloboDetalle(contenido) {
    const globo = obtenerGloboDetalle();
    globo.textContent = '';
    const titulo = document.createElement('div');
    titulo.className = 'globo-titulo';
    titulo.textContent = contenido.titulo;
    globo.appendChild(titulo);
    for (const linea of contenido.lineas) {
      const fila = document.createElement('div');
      fila.className = 'globo-linea';
      if (linea.color) {
        const muestra = document.createElement('i');
        muestra.style.background = linea.color;
        fila.appendChild(muestra);
      }
      const texto = document.createElement('span');
      texto.textContent = linea.texto;
      fila.appendChild(texto);
      globo.appendChild(fila);
    }
    globo.hidden = false;
    return globo;
  }

  function ubicarGloboDetalle(globo, x, y) {
    const margen = 12;
    const ancho = globo.offsetWidth;
    const alto = globo.offsetHeight;
    const izquierda = Math.min(Math.max(margen, x + 14), window.innerWidth - ancho - margen);
    const arriba = y - alto - 12 < margen ? y + 18 : y - alto - 12;
    globo.style.left = `${izquierda}px`;
    globo.style.top = `${arriba}px`;
  }

  function conectarDetalle(marca, obtenerContenido) {
    marca.setAttribute('tabindex', '0');
    marca.classList.add('marca');
    const ocultar = () => { obtenerGloboDetalle().hidden = true; };
    marca.addEventListener('pointermove', (evento) => ubicarGloboDetalle(llenarGloboDetalle(obtenerContenido()), evento.clientX, evento.clientY));
    marca.addEventListener('pointerleave', ocultar);
    marca.addEventListener('focus', () => {
      const caja = marca.getBoundingClientRect();
      ubicarGloboDetalle(llenarGloboDetalle(obtenerContenido()), caja.left + caja.width / 2, caja.top);
    });
    marca.addEventListener('blur', ocultar);
    const contenido = obtenerContenido();
    marca.setAttribute('aria-label', [contenido.titulo, ...contenido.lineas.map((linea) => linea.texto)].join('. '));
  }

  function mostrarVacio(contenedor, mensaje = 'No hay datos en este período.') {
    contenedor.textContent = '';
    const aviso = document.createElement('p');
    aviso.className = 'grafico-vacio';
    aviso.textContent = mensaje;
    contenedor.appendChild(aviso);
  }

  function dibujarBarrasHorizontales(contenedor, filas, opciones = {}) {
    const { obtenerColor = () => 'var(--acento)', obtenerDescripcion = () => '', unidad = 'informes' } = opciones;
    if (!filas.length) return mostrarVacio(contenedor);
    contenedor.textContent = '';
    const total = opciones.total ?? filas.reduce((suma, fila) => suma + fila.cantidad, 0);
    const maximo = Math.max(...filas.map((fila) => fila.cantidad), 1);
    const lista = document.createElement('div');
    lista.className = 'barras-lista';

    filas.forEach((fila, indice) => {
      const descripcion = obtenerDescripcion(fila);
      const color = obtenerColor(fila, indice);
      const filaBarra = document.createElement('div');
      filaBarra.className = 'barra-fila';

      const titulo = document.createElement('div');
      titulo.className = 'barra-titulo';
      const nombre = document.createElement('strong');
      nombre.textContent = fila.categoria;
      titulo.appendChild(nombre);
      if (descripcion) {
        const textoDescripcion = document.createElement('span');
        textoDescripcion.textContent = descripcion;
        titulo.appendChild(textoDescripcion);
      }

      const pista = document.createElement('div');
      pista.className = 'barra-pista';
      const relleno = document.createElement('div');
      relleno.className = 'barra-relleno';
      relleno.style.width = `${Math.max(1.5, (fila.cantidad / maximo) * 100)}%`;
      relleno.style.background = color;
      pista.appendChild(relleno);

      const valor = document.createElement('span');
      valor.className = 'barra-valor';
      const cantidad = document.createElement('b');
      cantidad.textContent = formatoNumero.format(fila.cantidad);
      valor.append(cantidad, document.createTextNode(` · ${calcularPorcentaje(fila.cantidad, total)}`));

      filaBarra.append(titulo, pista, valor);
      conectarDetalle(filaBarra, () => ({
        titulo: fila.categoria,
        lineas: [
          ...(descripcion ? [{ texto: descripcion }] : []),
          { color, texto: `${describirCantidad(fila.cantidad, unidad)} de ${formatoNumero.format(total)} (${calcularPorcentaje(fila.cantidad, total)})` }
        ]
      }));
      lista.appendChild(filaBarra);
    });
    contenedor.appendChild(lista);
  }

  function calcularMarcasEje(maximo) {
    const pasoBruto = maximo / 4;
    const magnitud = 10 ** Math.floor(Math.log10(Math.max(pasoBruto, 1)));
    const paso = [1, 2, 5, 10].map((factor) => factor * magnitud).find((candidato) => candidato >= pasoBruto) || magnitud * 10;
    const pasoEntero = Math.max(1, Math.round(paso));
    const marcas = [];
    for (let valor = 0; valor <= maximo + pasoEntero - 1; valor += pasoEntero) marcas.push(valor);
    return marcas;
  }

  function dibujarBarrasApiladas(contenedor, filas, series, opciones = {}) {
    const { unidad = 'informes' } = opciones;
    if (!filas.length) return mostrarVacio(contenedor);
    contenedor.textContent = '';
    const ancho = Math.max(280, contenedor.clientWidth);
    const alto = 280;
    const totales = filas.map((fila) => series.reduce((suma, serie) => suma + (fila[serie.clave] || 0), 0));
    const marcas = calcularMarcasEje(Math.max(...totales, 1));
    const maximoEje = marcas[marcas.length - 1];
    const margen = { arriba: 24, abajo: 30, izquierda: Math.max(...marcas.map((marca) => medirTexto(formatoNumero.format(marca), 12))) + 14, derecha: 8 };
    const altoUtil = alto - margen.arriba - margen.abajo;
    const escalaY = (valor) => (valor / maximoEje) * altoUtil;
    const anchoBanda = (ancho - margen.izquierda - margen.derecha) / filas.length;
    const anchoBarra = Math.min(56, anchoBanda * 0.6);
    const anchoEtiqueta = Math.max(...filas.map((fila) => medirTexto(fila.categoria, 12))) + 8;
    const saltoEtiquetas = Math.max(1, Math.ceil(anchoEtiqueta / anchoBanda));
    const svg = crearElementoSvg('svg', { viewBox: `0 0 ${ancho} ${alto}`, width: '100%', role: 'group' });

    for (const marca of marcas) {
      const y = alto - margen.abajo - escalaY(marca);
      svg.appendChild(crearElementoSvg('line', { x1: margen.izquierda, x2: ancho - margen.derecha, y1: y, y2: y }, `stroke: ${marca === 0 ? 'var(--borde-fuerte)' : 'var(--rejilla)'}`));
      svg.appendChild(crearTextoSvg(formatoNumero.format(marca), { x: margen.izquierda - 8, y: y + 4, 'text-anchor': 'end', 'font-size': 12 }, 'fill: var(--texto-suave); font-variant-numeric: tabular-nums'));
    }

    filas.forEach((fila, indice) => {
      const x = margen.izquierda + indice * anchoBanda + (anchoBanda - anchoBarra) / 2;
      const grupo = crearElementoSvg('g');
      let base = alto - margen.abajo;
      for (const serie of series) {
        const cantidad = fila[serie.clave] || 0;
        if (!cantidad) continue;
        const altoTramo = escalaY(cantidad);
        grupo.appendChild(crearElementoSvg('rect', { x, y: base - altoTramo, width: anchoBarra, height: altoTramo }, `fill: ${serie.color}`));
        if (altoTramo >= 18 && anchoBarra >= 20) {
          grupo.appendChild(crearTextoSvg(formatoNumero.format(cantidad), { x: x + anchoBarra / 2, y: base - altoTramo / 2 + 4.5, 'text-anchor': 'middle', 'font-size': 12, 'font-weight': 600 }, `fill: ${serie.colorTexto || 'var(--panel)'}; pointer-events: none`));
        }
        base -= altoTramo;
      }
      grupo.appendChild(crearTextoSvg(formatoNumero.format(totales[indice]), { x: x + anchoBarra / 2, y: base - 7, 'text-anchor': 'middle', 'font-size': 13, 'font-weight': 700 }, 'fill: var(--texto); font-variant-numeric: tabular-nums'));
      const zonaActiva = crearElementoSvg('rect', { x: margen.izquierda + indice * anchoBanda, y: margen.arriba - 20, width: anchoBanda, height: altoUtil + 20 }, 'fill: transparent');
      conectarDetalle(zonaActiva, () => ({
        titulo: `${fila.categoria}: ${describirCantidad(totales[indice], unidad)}`,
        lineas: series.filter((serie) => fila[serie.clave]).map((serie) => ({
          color: serie.color,
          texto: `${serie.nombre}: ${formatoNumero.format(fila[serie.clave])} (${calcularPorcentaje(fila[serie.clave], totales[indice])})`
        }))
      }));
      grupo.appendChild(zonaActiva);
      svg.appendChild(grupo);
      if (indice % saltoEtiquetas === 0) {
        svg.appendChild(crearTextoSvg(fila.categoria, { x: x + anchoBarra / 2, y: alto - 9, 'text-anchor': 'middle', 'font-size': 12, 'font-weight': 600 }, 'fill: var(--texto-suave)'));
      }
    });
    contenedor.appendChild(svg);
  }

  function dibujarProporciones(contenedor, grupos, opciones) {
    const { colores, coloresTexto = {}, describirGrupo = () => '', unidad = 'informes', categoriasFueraDeBarra = ['Sin dato'] } = opciones;
    contenedor.textContent = '';
    for (const [nombreGrupo, conteo] of Object.entries(grupos)) {
      const categoriasEnBarra = Object.entries(conteo).filter(([categoria]) => !categoriasFueraDeBarra.includes(categoria));
      const total = categoriasEnBarra.reduce((suma, [, cantidad]) => suma + cantidad, 0);
      const fila = document.createElement('div');
      fila.className = 'proporcion';
      const encabezado = document.createElement('div');
      encabezado.className = 'proporcion-encabezado';
      const nombre = document.createElement('strong');
      nombre.textContent = nombreGrupo;
      const resumen = document.createElement('span');
      resumen.className = 'proporcion-resumen';
      resumen.textContent = describirGrupo(nombreGrupo, conteo);
      encabezado.append(nombre, resumen);

      const barra = document.createElement('div');
      barra.className = 'proporcion-barra';
      if (!total) {
        barra.classList.add('proporcion-barra-vacia');
        barra.textContent = 'Sin datos';
      }
      for (const [categoria, cantidad] of categoriasEnBarra) {
        if (!cantidad || !total) continue;
        const tramo = document.createElement('div');
        tramo.className = 'proporcion-tramo';
        tramo.style.width = `${(cantidad / total) * 100}%`;
        tramo.style.background = colores[categoria] || 'var(--sin-dato)';
        tramo.style.color = coloresTexto[categoria] || 'var(--panel)';
        if (cantidad / total >= 0.1) tramo.textContent = calcularPorcentaje(cantidad, total);
        conectarDetalle(tramo, () => ({
          titulo: `${nombreGrupo}: ${categoria}`,
          lineas: [{ color: colores[categoria], texto: `${describirCantidad(cantidad, unidad)} de ${formatoNumero.format(total)} (${calcularPorcentaje(cantidad, total)})` }]
        }));
        barra.appendChild(tramo);
      }

      const detalle = document.createElement('div');
      detalle.className = 'proporcion-detalle';
      detalle.textContent = Object.entries(conteo).map(([categoria, cantidad]) => (categoriasFueraDeBarra.includes(categoria) ? `${categoria} (no se grafica): ${formatoNumero.format(cantidad)}` : `${categoria}: ${formatoNumero.format(cantidad)}`)).join(' · ');
      fila.append(encabezado, barra, detalle);
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

  ns.Graficos = { dibujarBarrasHorizontales, dibujarBarrasApiladas, dibujarProporciones, dibujarLeyenda, calcularPorcentaje, formatoNumero };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
