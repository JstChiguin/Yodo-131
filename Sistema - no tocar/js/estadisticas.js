(function (ns) {
  const ORDEN_EDAD = ['<20', '20–29', '30–39', '40–49', '50–59', '60–69', '70+', 'Sin dato'];
  const ORDEN_DOSIS = ['Baja (≤30 mCi)', 'Intermedia (31–99 mCi)', 'Alta (≥100 mCi)', 'Rechazó / no recibió', 'Sin dato'];
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

  function obtenerClavePaciente(registro) {
    if (registro.rut) return `rut:${registro.rut}`;
    return `nombre:${(registro.nombrePaciente || registro.archivo).toLowerCase()}`;
  }

  function contar(registros, obtenerCategoria, ordenFijo) {
    const conteo = new Map();
    for (const registro of registros) {
      const categoria = obtenerCategoria(registro);
      if (categoria == null) continue;
      conteo.set(categoria, (conteo.get(categoria) || 0) + 1);
    }
    const filas = [...conteo].map(([categoria, cantidad]) => ({ categoria, cantidad }));
    if (ordenFijo) return filas.sort((a, b) => ordenFijo.indexOf(a.categoria) - ordenFijo.indexOf(b.categoria));
    return filas.sort((a, b) => b.cantidad - a.cantidad);
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

  function obtenerPacientesUnicos(registros) {
    const pacientes = new Map();
    for (const registro of registros) {
      const clave = obtenerClavePaciente(registro);
      const existente = pacientes.get(clave);
      if (!existente || (!existente.sexo && registro.sexo)) pacientes.set(clave, registro);
    }
    return [...pacientes.values()];
  }

  function contarPorAnio(registros) {
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

  function calcularEstadisticas(todosLosRegistros, filtro) {
    const registros = filtrarPorPeriodo(todosLosRegistros, filtro);
    const pacientes = obtenerPacientesUnicos(registros);
    const registrosCancer = registros.filter((registro) => registro.diagnostico === 'Cáncer de tiroides');
    const pacientesCancer = obtenerPacientesUnicos(registrosCancer);
    const edades = registros.map((registro) => registro.edad).filter((edad) => edad != null);

    return {
      cantidadInformes: registros.length,
      cantidadSinFecha: todosLosRegistros.filter((registro) => !registro.anio).length,
      cantidadPacientes: pacientes.length,
      pacientesMujeres: pacientes.filter((paciente) => paciente.sexo === 'F').length,
      pacientesHombres: pacientes.filter((paciente) => paciente.sexo === 'M').length,
      pacientesSinSexo: pacientes.filter((paciente) => !paciente.sexo).length,
      cantidadPacientesCancer: pacientesCancer.length,
      edadPromedio: edades.length ? edades.reduce((suma, edad) => suma + edad, 0) / edades.length : null,
      edadMinima: edades.length ? Math.min(...edades) : null,
      edadMaxima: edades.length ? Math.max(...edades) : null,
      porAnio: contarPorAnio(registros),
      porDiagnostico: contar(registros, (registro) => registro.diagnostico),
      porMotivo: contar(registros, (registro) => registro.motivoCategoria),
      porEdad: contar(registros, (registro) => registro.grupoEdad, ORDEN_EDAD),
      porDosis: contar(registros, (registro) => registro.grupoDosis, ORDEN_DOSIS),
      porHistologia: contar(registrosCancer, (registro) => registro.histologia),
      porBethesda: contar(registros, (registro) => registro.bethesda, ORDEN_BETHESDA),
      porTirads: contar(registros, (registro) => (registro.tirads ? `TIRADS ${registro.tirads}` : null)).sort((a, b) => a.categoria.localeCompare(b.categoria)),
      porMedico: contar(registros, (registro) => registro.medicoTratante || 'Sin dato'),
      examenes: {
        'TRAb': contarEstadoExamen(registros, 'trabElevado'),
        'Anti-tiroglobulina': contarEstadoExamen(registros, 'antiTiroglobulinaElevado'),
        'Anti-TPO': contarEstadoExamen(registros, 'antiTpoElevado'),
        'Tiroglobulina': contarEstadoExamen(registros, 'tiroglobulinaElevada')
      },
      ecografia: {
        'Con nódulos': contarSiNo(registros, 'ecoConNodulos'),
        'Con adenopatías': contarSiNo(registros, 'ecoConAdenopatias')
      },
      antecedentes: {
        'Antecedente familiar tiroideo': contarSiNo(registros, 'tieneAntecedenteFamiliar'),
        'Tabaquismo actual o previo': contarSiNo(registros, 'esFumador')
      },
      registros
    };
  }

  ns.Estadisticas = { calcularEstadisticas, filtrarPorPeriodo };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
