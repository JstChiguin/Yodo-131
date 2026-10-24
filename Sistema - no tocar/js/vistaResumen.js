(function (ns) {
  const { Graficos } = ns;
  const { formatoNumero, calcularPorcentaje } = Graficos;
  const elemento = (id) => document.getElementById(id);

  const DESCRIPCION_TIRADS = {
    1: 'Normal', 2: 'No sospechoso', 3: 'Levemente sospechoso',
    4: 'Moderadamente sospechoso', 5: 'Altamente sospechoso', 6: 'Cáncer ya confirmado'
  };
  const DESCRIPCION_BETHESDA = {
    I: 'Muestra no diagnóstica', II: 'Benigno', III: 'Atipia de significado incierto',
    IV: 'Neoplasia folicular', V: 'Sospechoso de cáncer', VI: 'Maligno (cáncer)'
  };
  const COLORES_DIAGNOSTICO = {
    'Cáncer de tiroides': 'var(--elevado)',
    'Hipertiroidismo (Graves / bocio difuso)': 'var(--acento)',
    'Bocio multinodular o nódulo hiperfuncionante': 'var(--serie-3)',
    'Otro / no determinado': 'var(--sin-dato)'
  };
  const COLORES_ESTADO = { Elevado: 'var(--elevado)', Normal: 'var(--normal)', 'Sin dato': 'var(--sin-dato)' };
  const COLORES_SI_NO = { 'Sí': 'var(--serie-4)', 'No': 'var(--normal)', 'Sin dato': 'var(--sin-dato)' };
  const TEXTO_SIN_DATO = { 'Sin dato': 'var(--texto)' };
  const SERIES_SEXO = [
    { clave: 'mujeres', nombre: 'Mujeres', color: 'var(--mujeres)' },
    { clave: 'hombres', nombre: 'Hombres', color: 'var(--hombres)' },
    { clave: 'sinDato', nombre: 'Sexo sin dato', color: 'var(--sin-dato)', colorTexto: 'var(--texto)' }
  ];

  function sumarCantidades(filas) {
    return filas.reduce((suma, fila) => suma + fila.cantidad, 0);
  }

  function escribirConclusion(id, texto) {
    const parrafo = elemento(`conclusion-${id}`);
    parrafo.textContent = texto;
    parrafo.hidden = !texto;
  }

  function obtenerMasFrecuente(filas, excluir = ['Sin dato']) {
    const consideradas = filas.filter((fila) => !excluir.includes(fila.categoria));
    if (!consideradas.length) return null;
    const mayor = consideradas.reduce((actual, fila) => (fila.cantidad > actual.cantidad ? fila : actual));
    const total = sumarCantidades(filas);
    return { ...mayor, total, detalle: `${formatoNumero.format(mayor.cantidad)} de ${formatoNumero.format(total)}`, porcentaje: calcularPorcentaje(mayor.cantidad, total) };
  }

  function describirMasFrecuente(filas, introduccion, unidad = 'informes', excluir = ['Sin dato']) {
    const mayor = obtenerMasFrecuente(filas, excluir);
    return mayor ? `${introduccion}: ${mayor.categoria}, con ${mayor.detalle} ${unidad} (${mayor.porcentaje}).` : '';
  }

  function describirProporcionSobreConDato(cantidad, conDato, frase) {
    if (!conDato) return '';
    return `${frase}: ${formatoNumero.format(cantidad)} de ${formatoNumero.format(conDato)} (${calcularPorcentaje(cantidad, conDato)}).`;
  }

  function colorPorNivel(nivel, umbralAlto, umbralMedio) {
    if (nivel >= umbralAlto) return 'var(--elevado)';
    if (nivel >= umbralMedio) return 'var(--serie-4)';
    return 'var(--serie-3)';
  }

  function renderizarAnios(estadisticas) {
    Graficos.dibujarLeyenda(elemento('leyenda-anios'), SERIES_SEXO);
    Graficos.dibujarBarrasApiladas(elemento('grafico-anios'), estadisticas.porAnio, SERIES_SEXO);
    const conFecha = estadisticas.porAnio.filter((fila) => fila.categoria !== 'Sin fecha');
    if (!conFecha.length) return escribirConclusion('anios', '');
    const totalDe = (fila) => fila.mujeres + fila.hombres + fila.sinDato;
    const anioMayor = conFecha.reduce((actual, fila) => (totalDe(fila) > totalDe(actual) ? fila : actual));
    const totalMujeres = estadisticas.porAnio.reduce((suma, fila) => suma + fila.mujeres, 0);
    escribirConclusion('anios', `El año con más informes fue ${anioMayor.categoria} (${formatoNumero.format(totalDe(anioMayor))}). En todo el período, ${formatoNumero.format(totalMujeres)} de ${formatoNumero.format(estadisticas.cantidadInformes)} informes son de mujeres (${calcularPorcentaje(totalMujeres, estadisticas.cantidadInformes)}).`);
  }

  function renderizarDistribuciones(estadisticas) {
    Graficos.dibujarBarrasHorizontales(elemento('grafico-diagnostico'), estadisticas.porDiagnostico, { obtenerColor: (fila) => COLORES_DIAGNOSTICO[fila.categoria] || 'var(--sin-dato)' });
    escribirConclusion('diagnostico', describirMasFrecuente(estadisticas.porDiagnostico, 'Lo más frecuente'));

    Graficos.dibujarBarrasHorizontales(elemento('grafico-motivo'), estadisticas.porMotivo, { obtenerColor: () => 'var(--acento)' });
    escribirConclusion('motivo', describirMasFrecuente(estadisticas.porMotivo, 'El motivo más frecuente', 'informes', ['Sin motivo escrito', 'Otro']));

    Graficos.dibujarBarrasHorizontales(elemento('grafico-edad'), estadisticas.porEdad, { obtenerColor: (fila) => (fila.categoria === 'Sin dato' ? 'var(--sin-dato)' : 'var(--hombres)') });
    const grupoMayor = obtenerMasFrecuente(estadisticas.porEdad);
    const textoEdad = grupoMayor ? `El grupo más numeroso es el de ${grupoMayor.categoria} años, con ${grupoMayor.detalle} informes (${grupoMayor.porcentaje}).` : '';
    const textoPromedio = estadisticas.edadPromedio == null ? '' : ` Edad promedio: ${formatoNumero.format(Math.round(estadisticas.edadPromedio))} años.`;
    escribirConclusion('edad', `${textoEdad}${textoPromedio}`.trim());

    Graficos.dibujarBarrasHorizontales(elemento('grafico-dosis'), estadisticas.porDosis, {
      obtenerColor: (fila) => (fila.categoria.startsWith('Alta') ? 'var(--elevado)' : fila.categoria.startsWith('Intermedia') ? 'var(--serie-4)' : fila.categoria.startsWith('Baja') ? 'var(--acento)' : 'var(--sin-dato)')
    });
    escribirConclusion('dosis', describirMasFrecuente(estadisticas.porDosis, 'La dosis más indicada', 'informes', ['Sin dato', 'Rechazó / no recibió']));

    Graficos.dibujarBarrasHorizontales(elemento('grafico-histologia'), estadisticas.porHistologia, { obtenerColor: () => 'var(--elevado)' });
    escribirConclusion('histologia', describirMasFrecuente(estadisticas.porHistologia, 'El tipo más frecuente', 'informes con cáncer', ['No especificado']));

    Graficos.dibujarBarrasHorizontales(elemento('grafico-medico'), estadisticas.porMedico, { obtenerColor: () => 'var(--hombres)' });
    escribirConclusion('medico', describirMasFrecuente(estadisticas.porMedico, 'Con más informes'));
  }

  function renderizarExamenes(estadisticas, umbrales) {
    elemento('nota-umbrales').textContent = `Se considera elevado: TRAb > ${umbrales.trab}; tiroglobulina > ${umbrales.tiroglobulina} ng/ml o "en ascenso"; anti-tiroglobulina > ${umbrales.antiTiroglobulina}; anti-TPO > ${umbrales.antiTpo}. Cada barra muestra solo los informes que traen el examen; los que no lo traen se cuentan aparte como "Sin dato".`;
    Graficos.dibujarLeyenda(elemento('leyenda-examenes'), Object.entries(COLORES_ESTADO).filter(([nombre]) => nombre !== 'Sin dato').map(([nombre, color]) => ({ nombre, color })));
    Graficos.dibujarProporciones(elemento('grafico-examenes'), estadisticas.examenes, {
      colores: COLORES_ESTADO,
      coloresTexto: TEXTO_SIN_DATO,
      describirGrupo: (nombre, conteo) => {
        const conDato = conteo.Elevado + conteo.Normal;
        return conDato ? `Elevado en ${formatoNumero.format(conteo.Elevado)} de ${formatoNumero.format(conDato)} informes con el examen (${calcularPorcentaje(conteo.Elevado, conDato)})` : 'Ningún informe trae este examen';
      }
    });
    const examenesConDato = Object.entries(estadisticas.examenes).map(([nombre, conteo]) => ({ nombre, conDato: conteo.Elevado + conteo.Normal })).filter((examen) => examen.conDato);
    if (!examenesConDato.length) return escribirConclusion('examenes', 'Ningún informe del período trae estos exámenes.');
    const masInformado = examenesConDato.reduce((actual, examen) => (examen.conDato > actual.conDato ? examen : actual));
    escribirConclusion('examenes', `El examen que más aparece es ${masInformado.nombre}: está en ${formatoNumero.format(masInformado.conDato)} de ${formatoNumero.format(estadisticas.cantidadInformes)} informes.`);
  }

  function renderizarImagenesYPuncion(estadisticas) {
    Graficos.dibujarBarrasHorizontales(elemento('grafico-tirads'), estadisticas.porTirads, {
      obtenerColor: (fila) => colorPorNivel(Number(fila.categoria.slice(-1)), 5, 4),
      obtenerDescripcion: (fila) => DESCRIPCION_TIRADS[fila.categoria.slice(-1)] || ''
    });
    const totalTirads = sumarCantidades(estadisticas.porTirads);
    const altaSospecha = estadisticas.porTirads.filter((fila) => Number(fila.categoria.slice(-1)) >= 5).reduce((suma, fila) => suma + fila.cantidad, 0);
    escribirConclusion('tirads', describirProporcionSobreConDato(altaSospecha, totalTirads, 'Alta sospecha (TIRADS 5 o más) en ecografías clasificadas'));
    Graficos.dibujarProporciones(elemento('grafico-eco'), estadisticas.ecografia, {
      colores: COLORES_SI_NO,
      coloresTexto: TEXTO_SIN_DATO,
      describirGrupo: (nombre, conteo) => {
        const conDato = conteo['Sí'] + conteo.No;
        return conDato ? `${formatoNumero.format(conteo['Sí'])} de ${formatoNumero.format(conDato)} ecografías (${calcularPorcentaje(conteo['Sí'], conDato)})` : 'Sin ecografías descritas';
      }
    });

    const romanoBethesda = (fila) => fila.categoria.replace('Bethesda ', '');
    const filasBethesda = estadisticas.porBethesda.map((fila) => ({ ...fila, categoria: `Bethesda ${fila.categoria}` }));
    const nivelBethesda = { I: 1, II: 2, III: 3, IV: 4, V: 5, VI: 6 };
    Graficos.dibujarBarrasHorizontales(elemento('grafico-bethesda'), filasBethesda, {
      obtenerColor: (fila) => colorPorNivel(nivelBethesda[romanoBethesda(fila)], 5, 3),
      obtenerDescripcion: (fila) => DESCRIPCION_BETHESDA[romanoBethesda(fila)] || ''
    });
    const totalBethesda = sumarCantidades(filasBethesda);
    const sospechosas = filasBethesda.filter((fila) => nivelBethesda[romanoBethesda(fila)] >= 5).reduce((suma, fila) => suma + fila.cantidad, 0);
    escribirConclusion('bethesda', describirProporcionSobreConDato(sospechosas, totalBethesda, 'Sospechosas o malignas (Bethesda V y VI)'));
  }

  function renderizarAntecedentes(estadisticas) {
    Graficos.dibujarLeyenda(elemento('leyenda-antecedentes'), Object.entries(COLORES_SI_NO).filter(([nombre]) => nombre !== 'Sin dato').map(([nombre, color]) => ({ nombre, color })));
    Graficos.dibujarProporciones(elemento('grafico-antecedentes'), estadisticas.antecedentes, {
      colores: COLORES_SI_NO,
      coloresTexto: TEXTO_SIN_DATO,
      describirGrupo: (nombre, conteo) => {
        const conDato = conteo['Sí'] + conteo.No;
        return conDato ? `${formatoNumero.format(conteo['Sí'])} de ${formatoNumero.format(conDato)} con el dato (${calcularPorcentaje(conteo['Sí'], conDato)})` : 'Ningún informe trae el dato';
      }
    });
    const familiar = estadisticas.antecedentes['Antecedente familiar tiroideo'];
    escribirConclusion('antecedentes', describirProporcionSobreConDato(familiar['Sí'], familiar['Sí'] + familiar.No, 'Tienen familiares con enfermedad tiroidea'));
  }

  function renderizarGraficos(estadisticas, umbrales) {
    renderizarAnios(estadisticas);
    renderizarDistribuciones(estadisticas);
    renderizarExamenes(estadisticas, umbrales);
    renderizarImagenesYPuncion(estadisticas);
    renderizarAntecedentes(estadisticas);
  }

  ns.VistaResumen = { renderizarGraficos };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
