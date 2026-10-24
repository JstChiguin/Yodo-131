(function (ns) {
  const ORDEN_EDAD = ['<20', '20–29', '30–39', '40–49', '50–59', '60–69', '70+'];
  const ORDEN_DOSIS = ['Baja (≤30 mCi)', 'Intermedia (31–99 mCi)', 'Alta (≥100 mCi)', 'Rechazó / no recibió'];
  const ORDEN_BETHESDA = ['I', 'II', 'III', 'IV', 'V', 'VI'];

  function claveMes(anio, mes) {
    return anio ? `${anio}-${String(mes || 1).padStart(2, '0')}` : null;
  }

  function filtrarPorPeriodo(registros, filtro) {
    if (!filtro?.desde && !filtro?.hasta) return registros;
    return registros.filter((registro) => {
      if (!registro.anio) return false;
      const desdeMes = claveMes(registro.anio, registro.mes || 1);
      const hastaMes = claveMes(registro.anio, registro.mes || 12);
      if (filtro.desde && hastaMes < filtro.desde) return false;
      if (filtro.hasta && desdeMes > filtro.hasta) return false;
      return true;
    });
  }

  function calcularDistribucion(registros, obtenerCategoria, ordenFijo) {
    const conteo = new Map();
    for (const registro of registros) {
      const categoria = obtenerCategoria(registro);
      if (categoria == null) continue;
      conteo.set(categoria, (conteo.get(categoria) || 0) + 1);
    }
    const filas = [...conteo].map(([categoria, cantidad]) => ({ categoria, cantidad }));
    if (ordenFijo) filas.sort((a, b) => ordenFijo.indexOf(a.categoria) - ordenFijo.indexOf(b.categoria));
    else filas.sort((a, b) => b.cantidad - a.cantidad);
    const conDato = filas.reduce((suma, fila) => suma + fila.cantidad, 0);
    return { filas, conDato, sinDato: registros.length - conDato, total: registros.length };
  }

  function contarEstadoExamen(registros, campo) {
    const etiquetas = { true: 'Elevado', false: 'Normal', null: 'Sin dato' };
    const conteo = { Elevado: 0, Normal: 0, 'Sin dato': 0 };
    for (const registro of registros) conteo[etiquetas[registro[campo] ?? null]] += 1;
    return conteo;
  }

  function contarSiNo(registros, campo) {
    const conteo = { 'Sí': 0, 'No': 0, 'Sin dato': 0 };
    for (const registro of registros) conteo[registro[campo] === true ? 'Sí' : registro[campo] === false ? 'No' : 'Sin dato'] += 1;
    return conteo;
  }

  function contarPorAnioYSexo(registros) {
    const porAnio = new Map();
    for (const registro of registros) {
      const clave = registro.anio ? String(registro.anio) : 'Sin fecha';
      const fila = porAnio.get(clave) || { categoria: clave, mujeres: 0, hombres: 0, sinDato: 0 };
      if (registro.sexo === 'F') fila.mujeres += 1;
      else if (registro.sexo === 'M') fila.hombres += 1;
      else fila.sinDato += 1;
      porAnio.set(clave, fila);
    }
    return [...porAnio.values()].sort((a, b) => a.categoria.localeCompare(b.categoria));
  }

  function filtrarPorAnios(registros, filtro) {
    const anioDesde = filtro?.desde ? Number(filtro.desde.slice(0, 4)) : null;
    const anioHasta = filtro?.hasta ? Number(filtro.hasta.slice(0, 4)) : null;
    if (!anioDesde && !anioHasta) return registros;
    return registros.filter((registro) => registro.anio && (!anioDesde || registro.anio >= anioDesde) && (!anioHasta || registro.anio <= anioHasta));
  }

  function calcularEstadisticas(resultado, filtro) {
    const tratamientos = filtrarPorPeriodo(resultado.tratamientos, filtro);
    const tratamientosPorPaciente = new Map();
    for (const tratamiento of tratamientos) {
      if (!tratamientosPorPaciente.has(tratamiento.clavePaciente)) tratamientosPorPaciente.set(tratamiento.clavePaciente, []);
      tratamientosPorPaciente.get(tratamiento.clavePaciente).push(tratamiento);
    }
    const pacientes = [...tratamientosPorPaciente.values()].map((tratamientosPaciente) => tratamientosPaciente[0]);
    const tratamientosCancer = tratamientos.filter((tratamiento) => tratamiento.diagnostico === 'Cáncer de tiroides');
    const pacientesCancer = filtrarPorAnios(resultado.pacientesCancer, filtro);
    const edades = tratamientos.map((tratamiento) => tratamiento.edad).filter((edad) => edad != null);
    const idsEnPeriodo = new Set(tratamientos.map((tratamiento) => tratamiento.idTratamiento));

    return {
      cantidadTratamientos: tratamientos.length,
      cantidadSinFecha: resultado.tratamientos.filter((tratamiento) => !tratamiento.anio).length,
      cantidadPacientes: pacientes.length,
      pacientesConVariosTratamientos: [...tratamientosPorPaciente.values()].filter((lista) => lista.length > 1).length,
      pacientesMujeres: pacientes.filter((paciente) => paciente.sexo === 'F').length,
      pacientesHombres: pacientes.filter((paciente) => paciente.sexo === 'M').length,
      pacientesSinSexo: pacientes.filter((paciente) => !paciente.sexo).length,
      registrosUnidos: tratamientos.reduce((suma, tratamiento) => suma + tratamiento.cantidadFuentes - 1, 0),
      cantidadConflictos: resultado.conflictos.filter((conflicto) => idsEnPeriodo.has(conflicto.idTratamiento)).length,
      cantidadPacientesCancer: pacientesCancer.length,
      edadPromedio: edades.length ? edades.reduce((suma, edad) => suma + edad, 0) / edades.length : null,
      edadMinima: edades.length ? Math.min(...edades) : null,
      edadMaxima: edades.length ? Math.max(...edades) : null,
      coberturaEdad: { conDato: edades.length, total: tratamientos.length },
      porAnio: contarPorAnioYSexo(tratamientos),
      cancerPorAnio: contarPorAnioYSexo(pacientesCancer),
      porDiagnostico: calcularDistribucion(tratamientos, (tratamiento) => tratamiento.diagnostico),
      porMotivo: calcularDistribucion(tratamientos, (tratamiento) => tratamiento.motivoCategoria),
      porEdad: calcularDistribucion(tratamientos, (tratamiento) => tratamiento.grupoEdad, ORDEN_EDAD),
      porDosis: calcularDistribucion(tratamientos, (tratamiento) => tratamiento.grupoDosis, ORDEN_DOSIS),
      porHistologia: calcularDistribucion(tratamientosCancer, (tratamiento) => (tratamiento.histologia === 'No especificado' ? null : tratamiento.histologia)),
      porBethesda: calcularDistribucion(tratamientos, (tratamiento) => tratamiento.bethesda, ORDEN_BETHESDA),
      porTirads: calcularDistribucion(tratamientos, (tratamiento) => (tratamiento.tirads ? `TIRADS ${tratamiento.tirads}` : null), [1, 2, 3, 4, 5, 6].map((nivel) => `TIRADS ${nivel}`)),
      porMedico: calcularDistribucion(tratamientos, (tratamiento) => tratamiento.medicoTratante),
      porOrigen: calcularDistribucion(tratamientos, (tratamiento) => (tratamiento.origenes.length > 1 ? 'Word y Excel' : tratamiento.origenes[0])),
      examenes: {
        'TRAb': contarEstadoExamen(tratamientos, 'trabElevado'),
        'Anti-tiroglobulina': contarEstadoExamen(tratamientos, 'antiTiroglobulinaElevado'),
        'Anti-TPO': contarEstadoExamen(tratamientos, 'antiTpoElevado'),
        'Tiroglobulina': contarEstadoExamen(tratamientos, 'tiroglobulinaElevada')
      },
      ecografia: {
        'Con nódulos': contarSiNo(tratamientos, 'ecoConNodulos'),
        'Con adenopatías': contarSiNo(tratamientos, 'ecoConAdenopatias')
      },
      antecedentes: {
        'Antecedente familiar tiroideo': contarSiNo(tratamientos, 'tieneAntecedenteFamiliar'),
        'Tabaquismo actual o previo': contarSiNo(tratamientos, 'esFumador')
      },
      tratamientos
    };
  }

  ns.Estadisticas = { calcularEstadisticas, filtrarPorPeriodo, calcularDistribucion };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
