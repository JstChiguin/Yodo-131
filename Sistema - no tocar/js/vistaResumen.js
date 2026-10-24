(function (ns) {
  const { Graficos } = ns;
  const { formatoNumero, calcularPorcentaje } = Graficos;
  const elemento = (id) => document.getElementById(id);
  const UNIDAD = 'tratamientos';

  const DESCRIPCION_TIRADS = {
    1: 'Normal', 2: 'No sospechoso', 3: 'Levemente sospechoso',
    4: 'Moderadamente sospechoso', 5: 'Altamente sospechoso', 6: 'Cáncer ya confirmado'
  };
  const DESCRIPCION_BETHESDA = {
    I: 'Muestra no diagnóstica', II: 'Benigno', III: 'Atipia de significado incierto',
    IV: 'Neoplasia folicular', V: 'Sospechoso de cáncer', VI: 'Maligno (cáncer)'
  };
  const NIVEL_BETHESDA = { I: 1, II: 2, III: 3, IV: 4, V: 5, VI: 6 };
  const COLORES_DIAGNOSTICO = {
    'Cáncer de tiroides': 'var(--elevado)',
    'Hipertiroidismo (Graves / bocio difuso)': 'var(--acento)',
    'Bocio multinodular o nódulo hiperfuncionante': 'var(--serie-3)',
    'Otro / no determinado': 'var(--serie-5)'
  };
  const COLORES_ORIGEN = { Word: 'var(--acento)', Excel: 'var(--hombres)', 'Word y Excel': 'var(--serie-3)' };
  const COLORES_ESTADO = { Elevado: 'var(--elevado)', Normal: 'var(--normal)' };
  const COLORES_SI_NO = { 'Sí': 'var(--serie-4)', 'No': 'var(--normal)' };
  const SERIES_SEXO = [
    { clave: 'mujeres', nombre: 'Mujeres', color: 'var(--mujeres)' },
    { clave: 'hombres', nombre: 'Hombres', color: 'var(--hombres)' },
    { clave: 'sinDato', nombre: 'Sexo sin dato', color: 'var(--sin-dato)', colorTexto: 'var(--texto)' }
  ];

  function escribirTexto(id, texto) {
    const parrafo = elemento(id);
    parrafo.textContent = texto;
    parrafo.hidden = !texto;
  }

  function describirCobertura(conDato, total, unidad = UNIDAD) {
    if (!total) return '';
    if (conDato === total) return `Dato disponible en todos los ${unidad} (${formatoNumero.format(total)}).`;
    return `Dato disponible en ${formatoNumero.format(conDato)} de ${formatoNumero.format(total)} ${unidad} (${calcularPorcentaje(conDato, total)}). Sin dato: ${formatoNumero.format(total - conDato)} (${calcularPorcentaje(total - conDato, total)}). Los porcentajes del gráfico se calculan sobre los ${formatoNumero.format(conDato)} con dato.`;
  }

  function obtenerMasFrecuente(distribucion, excluir = []) {
    const consideradas = distribucion.filas.filter((fila) => !excluir.includes(fila.categoria));
    if (!consideradas.length) return null;
    const mayor = consideradas.reduce((actual, fila) => (fila.cantidad > actual.cantidad ? fila : actual));
    return { ...mayor, detalle: `${formatoNumero.format(mayor.cantidad)} de ${formatoNumero.format(distribucion.conDato)}`, porcentaje: calcularPorcentaje(mayor.cantidad, distribucion.conDato) };
  }

  function describirMasFrecuente(distribucion, introduccion, unidadConDato = 'con dato', excluir = []) {
    const mayor = obtenerMasFrecuente(distribucion, excluir);
    return mayor ? `${introduccion}: ${mayor.categoria}, con ${mayor.detalle} ${unidadConDato} (${mayor.porcentaje}).` : '';
  }

  function sumarCategorias(distribucion, condicion) {
    return distribucion.filas.filter(condicion).reduce((suma, fila) => suma + fila.cantidad, 0);
  }

  function colorPorNivel(nivel, umbralAlto, umbralMedio) {
    if (nivel >= umbralAlto) return 'var(--elevado)';
    if (nivel >= umbralMedio) return 'var(--serie-4)';
    return 'var(--serie-3)';
  }

  function dibujarDistribucion(id, distribucion, opciones) {
    Graficos.dibujarBarrasHorizontales(elemento(`grafico-${id}`), distribucion.filas, { unidad: UNIDAD, ...opciones, total: distribucion.conDato });
    escribirTexto(`cobertura-${id}`, describirCobertura(distribucion.conDato, distribucion.total, opciones?.unidadCobertura));
  }

  function renderizarPorAnio(estadisticas) {
    Graficos.dibujarLeyenda(elemento('leyenda-anios'), SERIES_SEXO);
    Graficos.dibujarBarrasApiladas(elemento('grafico-anios'), estadisticas.porAnio, SERIES_SEXO, { unidad: UNIDAD });
    const conFecha = estadisticas.porAnio.filter((fila) => fila.categoria !== 'Sin fecha');
    const totalDe = (fila) => fila.mujeres + fila.hombres + fila.sinDato;
    const conSexo = estadisticas.porAnio.reduce((suma, fila) => suma + fila.mujeres + fila.hombres, 0);
    const mujeres = estadisticas.porAnio.reduce((suma, fila) => suma + fila.mujeres, 0);
    if (!conFecha.length) {
      escribirTexto('conclusion-anios', '');
    } else {
      const anioMayor = conFecha.reduce((actual, fila) => (totalDe(fila) > totalDe(actual) ? fila : actual));
      escribirTexto('conclusion-anios', `El año con más tratamientos fue ${anioMayor.categoria} (${formatoNumero.format(totalDe(anioMayor))}). De los tratamientos con sexo registrado, ${formatoNumero.format(mujeres)} de ${formatoNumero.format(conSexo)} son de mujeres (${calcularPorcentaje(mujeres, conSexo)}).`);
    }
    escribirTexto('cobertura-anios', `Sexo disponible en ${formatoNumero.format(conSexo)} de ${formatoNumero.format(estadisticas.cantidadTratamientos)} tratamientos (${calcularPorcentaje(conSexo, estadisticas.cantidadTratamientos)}).${estadisticas.registrosUnidos ? ` Se unieron ${formatoNumero.format(estadisticas.registrosUnidos)} registros que correspondían a un mismo tratamiento.` : ''}`);
  }

  function renderizarCancerPorAnio(estadisticas) {
    Graficos.dibujarLeyenda(elemento('leyenda-cancer'), SERIES_SEXO);
    Graficos.dibujarBarrasApiladas(elemento('grafico-cancer'), estadisticas.cancerPorAnio, SERIES_SEXO, { unidad: 'pacientes' });
    const total = estadisticas.cancerPorAnio.reduce((suma, fila) => suma + fila.mujeres + fila.hombres + fila.sinDato, 0);
    if (!total) {
      escribirTexto('conclusion-cancer', 'No hay registros de cáncer de tiroides en este período.');
      escribirTexto('cobertura-cancer', '');
      return;
    }
    const mujeres = estadisticas.cancerPorAnio.reduce((suma, fila) => suma + fila.mujeres, 0);
    const conSexo = estadisticas.cancerPorAnio.reduce((suma, fila) => suma + fila.mujeres + fila.hombres, 0);
    const anioMayor = estadisticas.cancerPorAnio.reduce((actual, fila) => (fila.mujeres + fila.hombres + fila.sinDato > actual.mujeres + actual.hombres + actual.sinDato ? fila : actual));
    escribirTexto('conclusion-cancer', `${formatoNumero.format(total)} pacientes con cáncer de tiroides en el período; ${formatoNumero.format(mujeres)} de ${formatoNumero.format(conSexo)} con sexo registrado son mujeres (${calcularPorcentaje(mujeres, conSexo)}). El año con más casos nuevos fue ${anioMayor.categoria}.`);
    escribirTexto('cobertura-cancer', `Sexo disponible en ${formatoNumero.format(conSexo)} de ${formatoNumero.format(total)} pacientes (${calcularPorcentaje(conSexo, total)}).`);
  }

  function renderizarDistribuciones(estadisticas) {
    dibujarDistribucion('diagnostico', estadisticas.porDiagnostico, { obtenerColor: (fila) => COLORES_DIAGNOSTICO[fila.categoria] || 'var(--sin-dato)' });
    escribirTexto('conclusion-diagnostico', describirMasFrecuente(estadisticas.porDiagnostico, 'Lo más frecuente', 'tratamientos con diagnóstico'));

    dibujarDistribucion('motivo', estadisticas.porMotivo, { obtenerColor: () => 'var(--acento)' });
    escribirTexto('conclusion-motivo', describirMasFrecuente(estadisticas.porMotivo, 'El motivo más frecuente', 'tratamientos con motivo escrito', ['Otro']));

    dibujarDistribucion('edad', estadisticas.porEdad, { obtenerColor: () => 'var(--hombres)' });
    const grupoMayor = obtenerMasFrecuente(estadisticas.porEdad);
    const textoEdad = grupoMayor ? `El grupo más numeroso es el de ${grupoMayor.categoria} años, con ${grupoMayor.detalle} tratamientos con edad (${grupoMayor.porcentaje}).` : '';
    const textoPromedio = estadisticas.edadPromedio == null ? '' : ` Edad promedio: ${formatoNumero.format(Math.round(estadisticas.edadPromedio))} años.`;
    escribirTexto('conclusion-edad', `${textoEdad}${textoPromedio}`.trim());

    dibujarDistribucion('dosis', estadisticas.porDosis, {
      obtenerColor: (fila) => (fila.categoria.startsWith('Alta') ? 'var(--elevado)' : fila.categoria.startsWith('Intermedia') ? 'var(--serie-4)' : fila.categoria.startsWith('Baja') ? 'var(--acento)' : 'var(--sin-dato)')
    });
    escribirTexto('conclusion-dosis', describirMasFrecuente(estadisticas.porDosis, 'La dosis más indicada', 'tratamientos con dosis', ['Rechazó / no recibió']));

    dibujarDistribucion('histologia', estadisticas.porHistologia, { obtenerColor: () => 'var(--elevado)', unidadCobertura: 'tratamientos con cáncer' });
    escribirTexto('conclusion-histologia', describirMasFrecuente(estadisticas.porHistologia, 'El tipo más frecuente', 'con tipo informado'));

    dibujarDistribucion('medico', estadisticas.porMedico, { obtenerColor: () => 'var(--hombres)' });
    escribirTexto('conclusion-medico', describirMasFrecuente(estadisticas.porMedico, 'Con más tratamientos', 'con médico registrado'));

    dibujarDistribucion('origen', estadisticas.porOrigen, { obtenerColor: (fila) => COLORES_ORIGEN[fila.categoria] || 'var(--sin-dato)' });
    const unidos = sumarCategorias(estadisticas.porOrigen, (fila) => fila.categoria === 'Word y Excel');
    escribirTexto('conclusion-origen', unidos ? `${formatoNumero.format(unidos)} tratamientos tienen datos de Word y de Excel a la vez; sus datos se combinaron y las diferencias quedaron en "Por revisar".` : '');
  }

  function describirExamen(nombre, conteo) {
    const conDato = conteo.Elevado + conteo.Normal;
    const total = conDato + conteo['Sin dato'];
    if (!conDato) return 'Ningún tratamiento trae este examen';
    return `Elevado en ${formatoNumero.format(conteo.Elevado)} de ${formatoNumero.format(conDato)} con el examen (${calcularPorcentaje(conteo.Elevado, conDato)}) · registrado en ${formatoNumero.format(conDato)} de ${formatoNumero.format(total)} (${calcularPorcentaje(conDato, total)})`;
  }

  function describirSiNo(nombre, conteo) {
    const conDato = conteo['Sí'] + conteo.No;
    const total = conDato + conteo['Sin dato'];
    if (!conDato) return 'Ningún tratamiento trae el dato';
    return `${formatoNumero.format(conteo['Sí'])} de ${formatoNumero.format(conDato)} con el dato (${calcularPorcentaje(conteo['Sí'], conDato)}) · registrado en ${formatoNumero.format(conDato)} de ${formatoNumero.format(total)} (${calcularPorcentaje(conDato, total)})`;
  }

  function renderizarExamenes(estadisticas, umbrales) {
    elemento('nota-umbrales').textContent = `Se considera elevado: TRAb > ${umbrales.trab}; tiroglobulina > ${umbrales.tiroglobulina} ng/ml o "en ascenso"; anti-tiroglobulina > ${umbrales.antiTiroglobulina}; anti-TPO > ${umbrales.antiTpo}. Cada barra muestra solo los tratamientos que traen el examen; un examen no registrado no cuenta como normal.`;
    Graficos.dibujarLeyenda(elemento('leyenda-examenes'), Object.entries(COLORES_ESTADO).map(([nombre, color]) => ({ nombre, color })));
    Graficos.dibujarProporciones(elemento('grafico-examenes'), estadisticas.examenes, { colores: COLORES_ESTADO, describirGrupo: describirExamen, unidad: UNIDAD });
    const examenesConDato = Object.entries(estadisticas.examenes).map(([nombre, conteo]) => ({ nombre, conDato: conteo.Elevado + conteo.Normal })).filter((examen) => examen.conDato);
    escribirTexto('cobertura-examenes', '');
    if (!examenesConDato.length) return escribirTexto('conclusion-examenes', 'Ningún tratamiento del período trae estos exámenes. Las planillas Excel no los incluyen; solo vienen en los informes Word.');
    const masInformado = examenesConDato.reduce((actual, examen) => (examen.conDato > actual.conDato ? examen : actual));
    escribirTexto('conclusion-examenes', `El examen más registrado es ${masInformado.nombre}: está en ${formatoNumero.format(masInformado.conDato)} de ${formatoNumero.format(estadisticas.cantidadTratamientos)} tratamientos (${calcularPorcentaje(masInformado.conDato, estadisticas.cantidadTratamientos)}).`);
  }

  function renderizarImagenesYPuncion(estadisticas) {
    dibujarDistribucion('tirads', estadisticas.porTirads, {
      obtenerColor: (fila) => colorPorNivel(Number(fila.categoria.slice(-1)), 5, 4),
      obtenerDescripcion: (fila) => DESCRIPCION_TIRADS[fila.categoria.slice(-1)] || ''
    });
    const altaSospecha = sumarCategorias(estadisticas.porTirads, (fila) => Number(fila.categoria.slice(-1)) >= 5);
    escribirTexto('conclusion-tirads', estadisticas.porTirads.conDato ? `Alta sospecha (TIRADS 5 o más): ${formatoNumero.format(altaSospecha)} de ${formatoNumero.format(estadisticas.porTirads.conDato)} ecografías clasificadas (${calcularPorcentaje(altaSospecha, estadisticas.porTirads.conDato)}).` : '');
    Graficos.dibujarProporciones(elemento('grafico-eco'), estadisticas.ecografia, { colores: COLORES_SI_NO, describirGrupo: describirSiNo, unidad: UNIDAD });

    const romano = (fila) => fila.categoria.replace('Bethesda ', '');
    const distribucionBethesda = { ...estadisticas.porBethesda, filas: estadisticas.porBethesda.filas.map((fila) => ({ ...fila, categoria: `Bethesda ${fila.categoria}` })) };
    dibujarDistribucion('bethesda', distribucionBethesda, {
      obtenerColor: (fila) => colorPorNivel(NIVEL_BETHESDA[romano(fila)], 5, 3),
      obtenerDescripcion: (fila) => DESCRIPCION_BETHESDA[romano(fila)] || ''
    });
    const sospechosas = sumarCategorias(distribucionBethesda, (fila) => NIVEL_BETHESDA[romano(fila)] >= 5);
    escribirTexto('conclusion-bethesda', distribucionBethesda.conDato ? `Sospechosas o malignas (Bethesda V y VI): ${formatoNumero.format(sospechosas)} de ${formatoNumero.format(distribucionBethesda.conDato)} punciones informadas (${calcularPorcentaje(sospechosas, distribucionBethesda.conDato)}).` : '');
  }

  function renderizarAntecedentes(estadisticas) {
    Graficos.dibujarLeyenda(elemento('leyenda-antecedentes'), Object.entries(COLORES_SI_NO).map(([nombre, color]) => ({ nombre, color })));
    Graficos.dibujarProporciones(elemento('grafico-antecedentes'), estadisticas.antecedentes, { colores: COLORES_SI_NO, describirGrupo: describirSiNo, unidad: UNIDAD });
    const familiar = estadisticas.antecedentes['Antecedente familiar tiroideo'];
    const conDato = familiar['Sí'] + familiar.No;
    escribirTexto('conclusion-antecedentes', conDato ? `Tienen familiares con enfermedad tiroidea: ${formatoNumero.format(familiar['Sí'])} de ${formatoNumero.format(conDato)} tratamientos con el dato (${calcularPorcentaje(familiar['Sí'], conDato)}).` : '');
    escribirTexto('cobertura-antecedentes', '');
  }

  function renderizarGraficos(estadisticas, umbrales) {
    renderizarPorAnio(estadisticas);
    renderizarCancerPorAnio(estadisticas);
    renderizarDistribuciones(estadisticas);
    renderizarExamenes(estadisticas, umbrales);
    renderizarImagenesYPuncion(estadisticas);
    renderizarAntecedentes(estadisticas);
  }

  ns.VistaResumen = { renderizarGraficos };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
