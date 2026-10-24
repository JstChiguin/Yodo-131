(function (ns) {
  const { RegistroErrores, ProcesadorInformes, Estadisticas, ExportadorExcel, Configuracion, Seguridad, Graficos, SelloAutor, Tema, VistaResumen } = ns;
  const formatoNumero = Graficos.formatoNumero;
  const formatoPorcentaje = new Intl.NumberFormat('es-CL', { style: 'percent', maximumFractionDigits: 1 });
  const elemento = (id) => document.getElementById(id);

  const estado = {
    registrosLeidos: [],
    registros: [],
    documentosNoLeidos: [],
    cantidadArchivos: 0,
    filtro: { desde: '', hasta: '' },
    umbrales: Configuracion.obtenerUmbrales(),
    orden: { campo: 'fechaOrden', ascendente: false },
    textoBusqueda: '',
    seccionActiva: 'resumen',
    mantenimientoAbierto: false,
    ultimasEstadisticas: null
  };

  function mostrarAviso(mensaje, esError = false) {
    let aviso = elemento('aviso');
    if (!aviso) {
      aviso = document.createElement('p');
      aviso.id = 'aviso';
      aviso.setAttribute('role', 'status');
      document.querySelector('.contenido').prepend(aviso);
    }
    aviso.className = esError ? 'mensaje-error' : 'nota';
    aviso.textContent = mensaje;
    aviso.hidden = !mensaje;
  }

  function ejecutarConProteccion(componente, metodo, accion) {
    return async (...argumentos) => {
      try {
        await accion(...argumentos);
      } catch (error) {
        RegistroErrores.registrarError({ componente, metodo, error });
        mostrarAviso('Ocurrió un problema inesperado. Intente de nuevo; si se repite, avise a Chiguin.', true);
      }
    };
  }

  function hayDatos() {
    return estado.registrosLeidos.length > 0 || estado.documentosNoLeidos.length > 0;
  }

  function cambiarSeccion(seccion) {
    estado.seccionActiva = seccion;
    document.querySelectorAll('.pestana').forEach((pestana) => pestana.classList.toggle('activa', pestana.dataset.seccion === seccion));
    const mostrarBienvenida = !hayDatos() && seccion !== 'mantenimiento';
    elemento('seccion-bienvenida').hidden = !mostrarBienvenida;
    for (const nombre of ['resumen', 'pacientes', 'revisar', 'mantenimiento']) {
      elemento(`seccion-${nombre}`).hidden = mostrarBienvenida || nombre !== seccion;
    }
    if (!mostrarBienvenida) renderizarSeccionActiva();
  }

  function renderizarSeccionActiva() {
    if (estado.seccionActiva === 'resumen') renderizarResumen();
    if (estado.seccionActiva === 'pacientes') renderizarTabla();
    if (estado.seccionActiva === 'revisar') renderizarRevisar();
    if (estado.seccionActiva === 'mantenimiento') renderizarMantenimiento();
  }

  async function leerArchivos(archivos) {
    const cantidadWord = archivos.filter((archivo) => ProcesadorInformes.esArchivoWordValido(archivo.nombre)).length;
    if (!cantidadWord) {
      mostrarAviso('En esa carpeta no hay archivos Word (.docx). Revise que eligió la carpeta correcta.', true);
      return;
    }
    mostrarAviso('');
    elemento('estado-carga').hidden = false;
    const actualizarProgreso = (leidos, total) => {
      elemento('barra-progreso-relleno').style.width = `${(leidos / total) * 100}%`;
      elemento('texto-progreso').textContent = `Leyendo informes… ${leidos} de ${total}`;
    };
    actualizarProgreso(0, cantidadWord);
    try {
      const resultado = await ProcesadorInformes.procesarArchivos(archivos, estado.umbrales, actualizarProgreso);
      estado.registrosLeidos = resultado.registros;
      estado.registros = resultado.registros;
      estado.documentosNoLeidos = resultado.documentosNoLeidos;
      estado.cantidadArchivos = resultado.cantidadArchivos;
      elemento('boton-excel').disabled = estado.registros.length === 0;
      actualizarContadorRevisar();
      cambiarSeccion(estado.seccionActiva === 'mantenimiento' ? 'resumen' : estado.seccionActiva);
    } finally {
      elemento('estado-carga').hidden = true;
    }
  }

  function archivosDesdeSelector(listaArchivos) {
    return [...listaArchivos].map((archivo) => ({ nombre: archivo.webkitRelativePath || archivo.name, obtenerBuffer: () => archivo.arrayBuffer() }));
  }

  async function leerEntradaArrastrada(entrada, rutaPadre = '') {
    if (entrada.isFile) {
      const archivo = await new Promise((resolver, rechazar) => entrada.file(resolver, rechazar));
      return [{ nombre: `${rutaPadre}${archivo.name}`, obtenerBuffer: () => archivo.arrayBuffer() }];
    }
    if (!entrada.isDirectory) return [];
    const lector = entrada.createReader();
    const hijos = [];
    let lote;
    do {
      lote = await new Promise((resolver, rechazar) => lector.readEntries(resolver, rechazar));
      hijos.push(...lote);
    } while (lote.length);
    const archivos = [];
    for (const hijo of hijos) archivos.push(...await leerEntradaArrastrada(hijo, `${rutaPadre}${entrada.name}/`));
    return archivos;
  }

  function describirPeriodo() {
    const { desde, hasta } = estado.filtro;
    const formatear = (mesIso) => {
      const [anio, mes] = mesIso.split('-');
      return new Date(Number(anio), Number(mes) - 1, 1).toLocaleDateString('es-CL', { month: 'long', year: 'numeric' });
    };
    if (!desde && !hasta) return 'Todos los informes';
    if (desde && hasta) return `De ${formatear(desde)} a ${formatear(hasta)}`;
    return desde ? `Desde ${formatear(desde)}` : `Hasta ${formatear(hasta)}`;
  }

  function crearCifra(etiqueta, valor, nota) {
    const cifra = document.createElement('div');
    cifra.className = 'cifra';
    const etiquetaElemento = document.createElement('span');
    etiquetaElemento.className = 'cifra-etiqueta';
    etiquetaElemento.textContent = etiqueta;
    const valorElemento = document.createElement('span');
    valorElemento.className = 'cifra-valor';
    valorElemento.textContent = valor;
    const notaElemento = document.createElement('span');
    notaElemento.className = 'cifra-nota';
    notaElemento.textContent = nota;
    cifra.append(etiquetaElemento, valorElemento, notaElemento);
    return cifra;
  }

  function renderizarResumen() {
    const estadisticas = Estadisticas.calcularEstadisticas(estado.registros, estado.filtro);
    estado.ultimasEstadisticas = estadisticas;
    const sinFecha = estadisticas.cantidadSinFecha;
    const hayFiltro = estado.filtro.desde || estado.filtro.hasta;
    elemento('nota-periodo').textContent = `${describirPeriodo()}${hayFiltro && sinFecha ? ` · ${sinFecha} informe(s) sin fecha quedan fuera` : ''}`;

    const porcentajeCancer = estadisticas.cantidadPacientes ? formatoPorcentaje.format(estadisticas.cantidadPacientesCancer / estadisticas.cantidadPacientes) : '—';
    const cifras = elemento('cifras');
    cifras.textContent = '';
    cifras.append(
      crearCifra('Informes leídos', formatoNumero.format(estadisticas.cantidadInformes), `${formatoNumero.format(estado.cantidadArchivos)} archivos Word en la carpeta`),
      crearCifra('Pacientes distintos', formatoNumero.format(estadisticas.cantidadPacientes), 'Se cuentan una vez aunque tengan varios informes'),
      crearCifra('Mujeres', formatoNumero.format(estadisticas.pacientesMujeres), `Hombres: ${formatoNumero.format(estadisticas.pacientesHombres)} · Sin dato: ${formatoNumero.format(estadisticas.pacientesSinSexo)}`),
      crearCifra('Edad promedio', estadisticas.edadPromedio == null ? '—' : `${formatoNumero.format(Math.round(estadisticas.edadPromedio))} años`,
        estadisticas.edadMinima == null ? 'Sin edades' : `Entre ${estadisticas.edadMinima} y ${estadisticas.edadMaxima} años`),
      crearCifra('Pacientes con cáncer de tiroides', formatoNumero.format(estadisticas.cantidadPacientesCancer), `${porcentajeCancer} de los pacientes`)
    );

    VistaResumen.renderizarGraficos(estadisticas, estado.umbrales);
  }

  function textoMedicion(medicion) {
    return medicion ? `${medicion.comparador}${String(medicion.valor).replace('.', ',')}` : '';
  }

  function crearEtiquetaExamen(nombre, elevado, valorTexto) {
    if (elevado == null) return null;
    const etiqueta = document.createElement('span');
    etiqueta.className = `etiqueta ${elevado ? 'etiqueta-elevado' : 'etiqueta-normal'}`;
    etiqueta.textContent = `${nombre} ${valorTexto}${elevado ? ' ↑' : ''}`.trim();
    return etiqueta;
  }

  function obtenerRegistrosTabla() {
    const filtrados = Estadisticas.filtrarPorPeriodo(estado.registros, estado.filtro);
    const busqueda = estado.textoBusqueda.trim().toLowerCase();
    const encontrados = busqueda
      ? filtrados.filter((registro) => [registro.nombrePaciente, registro.rut, registro.diagnostico, registro.histologia, registro.motivoCategoria, registro.medicoTratante, registro.archivo]
        .some((valor) => valor && valor.toLowerCase().includes(busqueda)))
      : filtrados;
    const { campo, ascendente } = estado.orden;
    const valorOrden = (registro) => (campo === 'fechaOrden' ? registro.fechaIso || (registro.anio ? `${registro.anio}` : '') : registro[campo]);
    return [...encontrados].sort((registroA, registroB) => {
      const valorA = valorOrden(registroA);
      const valorB = valorOrden(registroB);
      if (valorA == null || valorA === '') return 1;
      if (valorB == null || valorB === '') return -1;
      const comparacion = typeof valorA === 'number' ? valorA - valorB : String(valorA).localeCompare(String(valorB), 'es');
      return ascendente ? comparacion : -comparacion;
    });
  }

  function crearFilaDetalle(registro, cantidadColumnas) {
    const fila = document.createElement('tr');
    fila.className = 'detalle';
    const celda = document.createElement('td');
    celda.colSpan = cantidadColumnas;
    const grilla = document.createElement('div');
    grilla.className = 'detalle-grilla';
    const campos = [
      ['Archivo', registro.archivo + (registro.indiceEnArchivo ? ` (informe ${registro.indiceEnArchivo} del archivo)` : '')],
      ['Motivo', `${registro.motivoCategoria}. ${registro.textoMotivo}`],
      ['Ecografía', registro.textoEco], ['PAAF', registro.textoPaaf], ['Biopsia', registro.textoBiopsia],
      ['Estadificación', registro.estadificacion], ['Cintigrama', registro.textoCintigrama],
      ['Laboratorio', [['TSH', registro.tsh], ['T4L', registro.t4l], ['T3', registro.t3], ['TRAb', registro.trab], ['Anti-Tg', registro.antiTiroglobulina], ['Anti-TPO', registro.antiTpo]]
        .filter(([, medicion]) => medicion).map(([nombre, medicion]) => `${nombre} ${textoMedicion(medicion)}`)
        .concat(registro.tiroglobulina?.valorMaximo != null ? [`Tiroglobulina máx. ${String(registro.tiroglobulina.valorMaximo).replace('.', ',')}`] : []).join(' · ')],
      ['Antecedentes familiares', registro.textoAntecedentesFamiliares], ['Mórbidos', registro.textoMorbidos],
      ['Hábitos', [registro.textoCigarrillo && `Cigarrillo: ${registro.textoCigarrillo}`, registro.textoAlcohol && `OH: ${registro.textoAlcohol}`, registro.textoDrogas && `Drogas: ${registro.textoDrogas}`].filter(Boolean).join(' · ')],
      ['Médico tratante', registro.medicoTratante], ['Sexo', registro.sexo ? `${registro.sexo} (según ${registro.origenSexo.toLowerCase()})` : 'Sin dato'],
      ['Yodo', registro.rechazoYodo ? 'El informe indica que la paciente rechazó o no recibió el yodo' : '']
    ];
    for (const [titulo, texto] of campos) {
      if (!texto) continue;
      const parrafo = document.createElement('div');
      const negrita = document.createElement('b');
      negrita.textContent = `${titulo}: `;
      parrafo.append(negrita, document.createTextNode(texto));
      grilla.appendChild(parrafo);
    }
    celda.appendChild(grilla);
    fila.appendChild(celda);
    return fila;
  }

  function renderizarTabla() {
    const registrosTabla = obtenerRegistrosTabla();
    const cuerpo = document.querySelector('#tabla-pacientes tbody');
    const cantidadColumnas = document.querySelectorAll('#tabla-pacientes thead th').length;
    cuerpo.textContent = '';
    elemento('conteo-tabla').textContent = `${formatoNumero.format(registrosTabla.length)} informe(s) · ${describirPeriodo()}`;
    document.querySelectorAll('#tabla-pacientes th').forEach((encabezado) => {
      const boton = encabezado.querySelector('button');
      if (!boton) return;
      if (boton.dataset.orden === estado.orden.campo) encabezado.setAttribute('aria-sort', estado.orden.ascendente ? 'ascending' : 'descending');
      else encabezado.removeAttribute('aria-sort');
    });

    for (const registro of registrosTabla) {
      const fila = document.createElement('tr');
      fila.className = 'fila-informe';
      fila.tabIndex = 0;
      const valores = [
        registro.nombrePaciente, registro.rut || '', registro.edad ?? '', registro.sexo || '',
        registro.fechaIso ? registro.fechaIso.split('-').reverse().join('/') : registro.anio ?? '',
        registro.histologia ? `${registro.diagnostico} · ${registro.histologia}` : registro.diagnostico,
        registro.dosisMci != null ? `${registro.dosisMci} mCi${registro.rechazoYodo ? ' (rechazó)' : ''}` : registro.rechazoYodo ? 'Rechazó' : '',
        registro.tirads ?? '', registro.bethesda ?? ''
      ];
      for (const valor of valores) {
        const celda = document.createElement('td');
        celda.textContent = String(valor);
        fila.appendChild(celda);
      }
      const celdaExamenes = document.createElement('td');
      [crearEtiquetaExamen('TRAb', registro.trabElevado, textoMedicion(registro.trab)),
        crearEtiquetaExamen('Tg', registro.tiroglobulinaElevada, registro.tiroglobulina?.valorMaximo != null ? String(registro.tiroglobulina.valorMaximo).replace('.', ',') : ''),
        crearEtiquetaExamen('Anti-Tg', registro.antiTiroglobulinaElevado, textoMedicion(registro.antiTiroglobulina)),
        crearEtiquetaExamen('Anti-TPO', registro.antiTpoElevado, textoMedicion(registro.antiTpo))]
        .filter(Boolean).forEach((etiqueta) => celdaExamenes.appendChild(etiqueta));
      const celdaRevisar = document.createElement('td');
      if (registro.camposFaltantes.length) {
        const etiqueta = document.createElement('span');
        etiqueta.className = 'etiqueta etiqueta-revisar';
        etiqueta.textContent = `Falta ${registro.camposFaltantes.join(', ')}`;
        celdaRevisar.appendChild(etiqueta);
      }
      fila.append(celdaExamenes, celdaRevisar);
      const alternarDetalle = () => {
        const siguiente = fila.nextElementSibling;
        if (siguiente?.classList.contains('detalle')) siguiente.remove();
        else fila.after(crearFilaDetalle(registro, cantidadColumnas));
      };
      fila.addEventListener('click', alternarDetalle);
      fila.addEventListener('keydown', (evento) => { if (evento.key === 'Enter') alternarDetalle(); });
      cuerpo.appendChild(fila);
    }
  }

  function actualizarContadorRevisar() {
    const cantidad = estado.registros.filter((registro) => registro.camposFaltantes.length).length + estado.documentosNoLeidos.length;
    const contador = elemento('contador-revisar');
    contador.hidden = cantidad === 0;
    contador.textContent = String(cantidad);
  }

  function agregarItemLista(lista, titulo, detalle) {
    const item = document.createElement('li');
    const texto = document.createElement('span');
    texto.textContent = titulo;
    const pequeno = document.createElement('small');
    pequeno.textContent = detalle;
    item.append(texto, pequeno);
    lista.appendChild(item);
  }

  function renderizarRevisar() {
    const listaFaltantes = elemento('lista-faltantes');
    listaFaltantes.textContent = '';
    const conFaltantes = estado.registros.filter((registro) => registro.camposFaltantes.length);
    if (!conFaltantes.length) agregarItemLista(listaFaltantes, 'Todos los informes tienen sus datos principales.', '');
    for (const registro of conFaltantes) {
      agregarItemLista(listaFaltantes, `${registro.nombrePaciente || 'Sin nombre'} — falta ${registro.camposFaltantes.join(', ')}`,
        registro.archivo + (registro.indiceEnArchivo ? ` (informe ${registro.indiceEnArchivo} del archivo)` : ''));
    }
    const listaNoLeidos = elemento('lista-no-leidos');
    listaNoLeidos.textContent = '';
    if (!estado.documentosNoLeidos.length) agregarItemLista(listaNoLeidos, 'Se pudieron leer todos los documentos.', '');
    for (const documento of estado.documentosNoLeidos) agregarItemLista(listaNoLeidos, documento.archivo, documento.motivo);
  }

  function renderizarMantenimiento() {
    elemento('panel-clave').hidden = estado.mantenimientoAbierto;
    elemento('panel-mantenimiento').hidden = !estado.mantenimientoAbierto;
    if (!estado.mantenimientoAbierto) return;
    document.querySelectorAll('[data-umbral]').forEach((campo) => { campo.value = estado.umbrales[campo.dataset.umbral]; });
    const listaErrores = elemento('lista-errores');
    listaErrores.textContent = '';
    const errores = RegistroErrores.obtenerErrores().reverse();
    if (!errores.length) agregarItemLista(listaErrores, 'Sin errores registrados.', '');
    for (const error of errores.slice(0, 50)) {
      agregarItemLista(listaErrores, `${error.componente}.${error.metodo}: ${error.tipoExcepcion} — ${error.mensaje}`,
        `${new Date(error.fechaHora).toLocaleString('es-CL')}${error.versionAplicacion ? ` · v${error.versionAplicacion}` : ''}${error.contexto ? ` · ${error.contexto}` : ''}`);
    }
  }

  function aplicarUmbrales(umbrales) {
    estado.umbrales = umbrales;
    estado.registros = ProcesadorInformes.reclasificar(estado.registrosLeidos, umbrales);
    actualizarContadorRevisar();
  }

  function descargarExcel() {
    const estadisticas = Estadisticas.calcularEstadisticas(estado.registros, estado.filtro);
    const libro = ExportadorExcel.exportarLibro(estadisticas, estado.documentosNoLeidos, describirPeriodo());
    const enlace = document.createElement('a');
    enlace.href = URL.createObjectURL(libro);
    enlace.download = `Yodo131_${new Date().toISOString().slice(0, 10)}.xlsx`;
    document.body.appendChild(enlace);
    enlace.click();
    enlace.remove();
    setTimeout(() => URL.revokeObjectURL(enlace.href), 10000);
    mostrarAviso('Excel descargado. Lo encuentra en la carpeta Descargas.');
  }

  function actualizarBotonTema() {
    const temaOscuroActivo = Tema.obtenerTemaActual() === Tema.TEMA_OSCURO;
    elemento('boton-tema').setAttribute('aria-pressed', String(temaOscuroActivo));
    elemento('texto-boton-tema').textContent = temaOscuroActivo ? 'Modo claro' : 'Modo oscuro';
  }

  function conectarEventos() {
    elemento('boton-tema').addEventListener('click', () => {
      Tema.alternarTema();
      actualizarBotonTema();
    });
    document.querySelectorAll('.pestana').forEach((pestana) => pestana.addEventListener('click', () => cambiarSeccion(pestana.dataset.seccion)));

    elemento('selector-carpeta').addEventListener('change', ejecutarConProteccion('App', 'elegirCarpeta', async (evento) => {
      await leerArchivos(archivosDesdeSelector(evento.target.files));
      evento.target.value = '';
    }));

    const zonaSoltar = elemento('zona-soltar');
    zonaSoltar.addEventListener('dragover', (evento) => { evento.preventDefault(); zonaSoltar.classList.add('arrastrando'); });
    zonaSoltar.addEventListener('dragleave', () => zonaSoltar.classList.remove('arrastrando'));
    zonaSoltar.addEventListener('drop', ejecutarConProteccion('App', 'soltarCarpeta', async (evento) => {
      evento.preventDefault();
      zonaSoltar.classList.remove('arrastrando');
      const entradas = [...evento.dataTransfer.items].map((item) => item.webkitGetAsEntry?.()).filter(Boolean);
      const archivos = [];
      for (const entrada of entradas) archivos.push(...await leerEntradaArrastrada(entrada));
      await leerArchivos(archivos);
    }));

    const alCambiarPeriodo = () => {
      estado.filtro = { desde: elemento('filtro-desde').value, hasta: elemento('filtro-hasta').value };
      if (estado.filtro.desde && estado.filtro.hasta && estado.filtro.desde > estado.filtro.hasta) {
        mostrarAviso('La fecha "Desde" es posterior a "Hasta". Revise el período.', true);
        return;
      }
      mostrarAviso('');
      renderizarSeccionActiva();
    };
    elemento('filtro-desde').addEventListener('change', ejecutarConProteccion('App', 'cambiarPeriodo', alCambiarPeriodo));
    elemento('filtro-hasta').addEventListener('change', ejecutarConProteccion('App', 'cambiarPeriodo', alCambiarPeriodo));
    elemento('boton-todo-periodo').addEventListener('click', ejecutarConProteccion('App', 'verTodo', () => {
      elemento('filtro-desde').value = '';
      elemento('filtro-hasta').value = '';
      alCambiarPeriodo();
    }));

    elemento('buscador').addEventListener('input', ejecutarConProteccion('App', 'buscar', (evento) => {
      estado.textoBusqueda = evento.target.value;
      renderizarTabla();
    }));
    document.querySelectorAll('#tabla-pacientes th button').forEach((boton) => boton.addEventListener('click', ejecutarConProteccion('App', 'ordenarTabla', () => {
      const campo = boton.dataset.orden;
      estado.orden = { campo, ascendente: estado.orden.campo === campo ? !estado.orden.ascendente : true };
      renderizarTabla();
    })));

    elemento('boton-excel').addEventListener('click', ejecutarConProteccion('App', 'descargarExcel', descargarExcel));

    elemento('formulario-clave').addEventListener('submit', ejecutarConProteccion('App', 'verificarClave', async (evento) => {
      evento.preventDefault();
      const botonEntrar = elemento('boton-entrar');
      const mensaje = elemento('mensaje-clave');
      botonEntrar.disabled = true;
      mensaje.textContent = '';
      try {
        const resultado = await Seguridad.verificarClave(elemento('campo-clave').value);
        elemento('campo-clave').value = '';
        if (resultado.valida) {
          estado.mantenimientoAbierto = true;
          renderizarMantenimiento();
        } else {
          mensaje.textContent = resultado.esperaSegundos ? `Demasiados intentos. Espere ${resultado.esperaSegundos} segundos.` : 'Clave incorrecta.';
        }
      } catch (error) {
        RegistroErrores.registrarError({ componente: 'App', metodo: 'verificarClave', error });
        mensaje.textContent = error.message;
      } finally {
        botonEntrar.disabled = false;
      }
    }));

    elemento('formulario-umbrales').addEventListener('submit', ejecutarConProteccion('App', 'guardarUmbrales', (evento) => {
      evento.preventDefault();
      const propuestos = {};
      document.querySelectorAll('[data-umbral]').forEach((campo) => { propuestos[campo.dataset.umbral] = campo.value; });
      const { umbrales, guardadoPermanente } = Configuracion.guardarUmbrales(propuestos);
      aplicarUmbrales(umbrales);
      renderizarMantenimiento();
      elemento('mensaje-umbrales').textContent = guardadoPermanente ? 'Valores guardados.' : 'Valores aplicados, pero este navegador no permite guardarlos para la próxima vez.';
    }));
    elemento('boton-restablecer').addEventListener('click', ejecutarConProteccion('App', 'restablecerUmbrales', () => {
      aplicarUmbrales(Configuracion.restablecerUmbrales());
      renderizarMantenimiento();
      elemento('mensaje-umbrales').textContent = 'Se volvió a los valores iniciales.';
    }));
    elemento('boton-limpiar-errores').addEventListener('click', ejecutarConProteccion('App', 'limpiarErrores', () => {
      RegistroErrores.limpiarErrores();
      renderizarMantenimiento();
    }));
    elemento('boton-bloquear').addEventListener('click', () => {
      estado.mantenimientoAbierto = false;
      renderizarMantenimiento();
    });

    let temporizadorRedimension;
    window.addEventListener('resize', () => {
      clearTimeout(temporizadorRedimension);
      temporizadorRedimension = setTimeout(() => { if (estado.seccionActiva === 'resumen' && hayDatos()) renderizarResumen(); }, 150);
    });
  }

  function iniciar() {
    SelloAutor.montarSello(elemento('app'));
    actualizarBotonTema();
    conectarEventos();
    cambiarSeccion('resumen');
    if (typeof DecompressionStream === 'undefined') {
      mostrarAviso('Este navegador es muy antiguo para leer los Word. Abra el programa con Google Chrome o Microsoft Edge.', true);
    }
  }

  iniciar();
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
