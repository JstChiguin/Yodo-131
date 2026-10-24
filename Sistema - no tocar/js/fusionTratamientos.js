(function (ns) {
  const VENTANA_MISMO_TRATAMIENTO_DIAS = 60;
  const MILISEGUNDOS_POR_DIA = 86400000;

  const CAMPOS_COMPARABLES = [
    ['edad', 'edad'], ['dosisMci', 'exacto'], ['medicoTratante', 'medico'],
    ['diagnostico', 'diagnostico'], ['histologia', 'histologia'], ['motivoCategoria', 'exacto'],
    ['tirads', 'exacto'], ['bethesda', 'exacto'], ['riesgo', 'exacto'],
    ['tsh', 'medicion'], ['t4l', 'medicion'], ['t3', 'medicion'], ['trab', 'medicion'],
    ['antiTiroglobulina', 'medicion'], ['antiTpo', 'medicion'],
    ['trabElevado', 'exacto'], ['tiroglobulinaElevada', 'exacto'], ['antiTiroglobulinaElevado', 'exacto'], ['antiTpoElevado', 'exacto'],
    ['ecoConNodulos', 'exacto'], ['ecoConAdenopatias', 'exacto'],
    ['tieneAntecedenteFamiliar', 'exacto'], ['esFumador', 'exacto']
  ];
  const CAMPOS_TEXTO = ['textoMotivo', 'textoEco', 'textoPaaf', 'textoBiopsia', 'estadificacion', 'textoCintigrama', 'textoAntecedentesFamiliares', 'textoMorbidos', 'textoCigarrillo', 'textoAlcohol', 'textoDrogas', 'textoDiagnosticoExcel'];
  const NOMBRES_CAMPOS = {
    edad: 'Edad', dosisMci: 'Dosis (mCi)', medicoTratante: 'Médico tratante', diagnostico: 'Diagnóstico', histologia: 'Tipo histológico',
    motivoCategoria: 'Motivo', tirads: 'TIRADS', bethesda: 'Bethesda', riesgo: 'Riesgo', tsh: 'TSH', t4l: 'T4L', t3: 'T3', trab: 'TRAb',
    antiTiroglobulina: 'Anti-tiroglobulina', antiTpo: 'Anti-TPO', trabElevado: 'TRAb elevado', tiroglobulinaElevada: 'Tiroglobulina elevada',
    antiTiroglobulinaElevado: 'Anti-tiroglobulina elevada', antiTpoElevado: 'Anti-TPO elevado', ecoConNodulos: 'Nódulos en ecografía',
    ecoConAdenopatias: 'Adenopatías en ecografía', tieneAntecedenteFamiliar: 'Antecedente familiar', esFumador: 'Tabaquismo', sexo: 'Sexo'
  };

  function quitarTildes(texto) {
    return String(texto).normalize('NFD').replace(/[̀-ͯ]/g, '');
  }

  function normalizarNombre(nombre) {
    return quitarTildes(nombre || '').toLowerCase().replace(/[^a-zñ\s]/g, ' ').replace(/\s+/g, ' ').trim();
  }

  function obtenerClavePaciente(registro) {
    if (registro.rut) return `rut:${registro.rut}`;
    const nombre = normalizarNombre(registro.nombrePaciente);
    return nombre ? `nombre:${nombre}` : `sin-identificar:${registro.referencia}`;
  }

  function obtenerPalabrasMedico(nombreMedico) {
    return quitarTildes(nombreMedico.replace(/^dra?\.?\s*/i, '')).toLowerCase().split(/\s+/).filter((palabra) => palabra.length > 1);
  }

  function sonEquivalentes(tipo, valorA, valorB) {
    if (tipo === 'edad') return Math.abs(valorA - valorB) <= 1;
    if (tipo === 'medicion') return valorA.valor === valorB.valor && valorA.comparador === valorB.comparador;
    if (tipo === 'medico') {
      const palabrasA = obtenerPalabrasMedico(valorA);
      const palabrasB = obtenerPalabrasMedico(valorB);
      const comparteApellido = palabrasA.length > 0 && palabrasB.length > 0 && palabrasA[palabrasA.length - 1] === palabrasB[palabrasB.length - 1];
      return comparteApellido || palabrasA.every((palabra) => palabrasB.includes(palabra)) || palabrasB.every((palabra) => palabrasA.includes(palabra));
    }
    return valorA === valorB;
  }

  function describirValor(valor) {
    if (valor === true) return 'Sí';
    if (valor === false) return 'No';
    if (valor && typeof valor === 'object' && 'valor' in valor) return `${valor.comparador}${String(valor.valor).replace('.', ',')}`;
    return String(valor);
  }

  function valorPreferido(tipo, valores) {
    if (tipo === 'medico') return [...valores].sort((a, b) => obtenerPalabrasMedico(b).length - obtenerPalabrasMedico(a).length)[0];
    if (tipo === 'edad') return Math.min(...valores);
    return valores[0];
  }

  function fusionarCampo(campo, tipo, fuentes) {
    let presentes = fuentes.filter((fuente) => fuente.registro[campo] != null).map((fuente) => ({ fuente, valor: fuente.registro[campo] }));
    if (tipo === 'diagnostico' && presentes.some(({ valor }) => valor !== 'Otro / no determinado')) presentes = presentes.filter(({ valor }) => valor !== 'Otro / no determinado');
    if (tipo === 'histologia' && presentes.some(({ valor }) => valor !== 'No especificado')) presentes = presentes.filter(({ valor }) => valor !== 'No especificado');
    if (!presentes.length) return { valor: null, conflicto: null };
    const distintos = [];
    for (const presente of presentes) {
      if (!distintos.some((otro) => sonEquivalentes(tipo, otro.valor, presente.valor))) distintos.push(presente);
    }
    if (distintos.length === 1) return { valor: valorPreferido(tipo, presentes.map(({ valor }) => valor)), conflicto: null };
    return {
      valor: null,
      conflicto: { campo, nombreCampo: NOMBRES_CAMPOS[campo] || campo, valores: presentes.map(({ fuente, valor }) => ({ origen: fuente.registro.origen, referencia: fuente.registro.referencia, valor: describirValor(valor) })) }
    };
  }

  function unirTextos(fuentes, campo) {
    const textos = [...new Set(fuentes.map((fuente) => fuente.registro[campo]).filter((texto) => texto && String(texto).trim()))];
    return textos.length ? textos.join(' | ') : null;
  }

  function obtenerDiaAproximado(registro) {
    if (registro.fechaIso) return { dia: Date.parse(registro.fechaIso) / MILISEGUNDOS_POR_DIA, esExacta: true };
    if (registro.anio && registro.mes) return { dia: Date.UTC(registro.anio, registro.mes - 1, 15) / MILISEGUNDOS_POR_DIA, esExacta: false };
    if (registro.anio) return { dia: Date.UTC(registro.anio, 6, 1) / MILISEGUNDOS_POR_DIA, esExacta: false, soloAnio: true };
    return null;
  }

  function sonFechasCompatibles(registroA, registroB) {
    const fechaA = obtenerDiaAproximado(registroA);
    const fechaB = obtenerDiaAproximado(registroB);
    if (!fechaA || !fechaB) return false;
    if (fechaA.esExacta && fechaB.esExacta) return Math.abs(fechaA.dia - fechaB.dia) <= VENTANA_MISMO_TRATAMIENTO_DIAS;
    if (registroA.anio !== registroB.anio) return false;
    if (fechaA.soloAnio || fechaB.soloAnio) return true;
    return Math.abs(registroA.mes - registroB.mes) <= 2;
  }

  function distanciaEntre(registroA, registroB) {
    const fechaA = obtenerDiaAproximado(registroA);
    const fechaB = obtenerDiaAproximado(registroB);
    return fechaA && fechaB ? Math.abs(fechaA.dia - fechaB.dia) : Infinity;
  }

  function agruparEnTratamientos(registrosPaciente) {
    const grupos = [];
    const sinFecha = [];
    const ordenados = [...registrosPaciente].sort((a, b) => (obtenerDiaAproximado(a)?.dia ?? Infinity) - (obtenerDiaAproximado(b)?.dia ?? Infinity));
    for (const registro of ordenados) {
      if (!obtenerDiaAproximado(registro)) {
        sinFecha.push(registro);
        continue;
      }
      const candidatos = grupos.filter((grupo) => grupo.some((otro) => sonFechasCompatibles(otro, registro)));
      if (!candidatos.length) {
        grupos.push([registro]);
        continue;
      }
      const masCercano = candidatos.sort((a, b) => Math.min(...a.map((otro) => distanciaEntre(otro, registro))) - Math.min(...b.map((otro) => distanciaEntre(otro, registro))))[0];
      masCercano.push(registro);
    }
    for (const registro of sinFecha) {
      if (grupos.length === 1) grupos[0].push(registro);
      else grupos.push([registro]);
    }
    return grupos;
  }

  function construirTratamiento(registrosGrupo, paciente, indice) {
    const fuentes = registrosGrupo.map((registro) => ({ registro }));
    const tratamiento = {
      idTratamiento: `${paciente.clave}#${indice + 1}`,
      clavePaciente: paciente.clave,
      nombrePaciente: paciente.nombrePaciente,
      rut: paciente.rut,
      fuentes: registrosGrupo.map((registro) => ({ origen: registro.origen, referencia: registro.referencia, fechaIso: registro.fechaIso, anio: registro.anio })),
      conflictos: []
    };
    for (const [campo, tipo] of CAMPOS_COMPARABLES) {
      const { valor, conflicto } = fusionarCampo(campo, tipo, fuentes);
      tratamiento[campo] = valor;
      if (conflicto) tratamiento.conflictos.push(conflicto);
    }
    for (const campo of CAMPOS_TEXTO) tratamiento[campo] = unirTextos(fuentes, campo);

    const fechaPrincipal = [...registrosGrupo].filter((registro) => registro.anio).sort((a, b) => (obtenerDiaAproximado(a).esExacta === obtenerDiaAproximado(b).esExacta ? obtenerDiaAproximado(a).dia - obtenerDiaAproximado(b).dia : obtenerDiaAproximado(a).esExacta ? -1 : 1))[0];
    tratamiento.fechaIso = fechaPrincipal?.fechaIso ?? null;
    tratamiento.anio = fechaPrincipal?.anio ?? null;
    tratamiento.mes = fechaPrincipal?.mes ?? null;
    tratamiento.tiroglobulina = {
      valorMaximo: registrosGrupo.map((registro) => registro.tiroglobulina?.valorMaximo).filter((valor) => valor != null).reduce((maximo, valor) => Math.max(maximo, valor), -Infinity),
      enAscenso: registrosGrupo.some((registro) => registro.tiroglobulina?.enAscenso)
    };
    if (tratamiento.tiroglobulina.valorMaximo === -Infinity) tratamiento.tiroglobulina.valorMaximo = null;
    tratamiento.rechazoYodo = registrosGrupo.some((registro) => registro.rechazoYodo);
    tratamiento.esRastreo = registrosGrupo.some((registro) => registro.esRastreo);
    tratamiento.origenes = [...new Set(registrosGrupo.map((registro) => registro.origen))];
    tratamiento.cantidadFuentes = registrosGrupo.length;
    return tratamiento;
  }

  function fusionarSexoPaciente(registros) {
    const explicitos = registros.filter((registro) => registro.sexoExplicito);
    const valoresExplicitos = [...new Set(explicitos.map((registro) => registro.sexoExplicito))];
    const inferidos = [...new Set(registros.map((registro) => registro.sexoInferido).filter(Boolean))];
    const describir = (registro, valor, origen) => ({ origen: `${registro.origen} (${origen})`, referencia: registro.referencia, valor });
    if (valoresExplicitos.length > 1) {
      return { sexo: null, origenSexo: 'Por revisar', conflicto: { campo: 'sexo', nombreCampo: 'Sexo', valores: explicitos.map((registro) => describir(registro, registro.sexoExplicito, registro.origenSexoExplicito)) } };
    }
    if (valoresExplicitos.length === 1) {
      if (inferidos.length === 1 && inferidos[0] !== valoresExplicitos[0]) {
        const conInferencia = registros.find((registro) => registro.sexoInferido);
        return {
          sexo: null, origenSexo: 'Por revisar',
          conflicto: { campo: 'sexo', nombreCampo: 'Sexo', valores: [describir(explicitos[0], valoresExplicitos[0], explicitos[0].origenSexoExplicito), describir(conInferencia, inferidos[0], `deducido del ${conInferencia.origenSexoInferido.toLowerCase()}`)] }
        };
      }
      return { sexo: valoresExplicitos[0], origenSexo: explicitos[0].origenSexoExplicito, conflicto: null };
    }
    if (inferidos.length === 1) return { sexo: inferidos[0], origenSexo: registros.find((registro) => registro.sexoInferido).origenSexoInferido, conflicto: null };
    return { sexo: null, origenSexo: 'Sin dato', conflicto: null };
  }

  function detectarNombresDistintos(registros) {
    const nombres = [...new Map(registros.filter((registro) => registro.nombrePaciente).map((registro) => [normalizarNombre(registro.nombrePaciente), registro])).values()];
    if (nombres.length < 2) return null;
    const palabrasDe = (registro) => normalizarNombre(registro.nombrePaciente).split(' ').filter((palabra) => palabra.length > 2);
    const base = palabrasDe(nombres[0]);
    const hayNombreIncompatible = nombres.slice(1).some((registro) => palabrasDe(registro).filter((palabra) => base.includes(palabra)).length < 2);
    if (!hayNombreIncompatible) return null;
    return { campo: 'nombre', nombreCampo: 'Mismo RUT con nombres distintos', valores: nombres.map((registro) => ({ origen: registro.origen, referencia: registro.referencia, valor: registro.nombrePaciente })) };
  }

  function asignarClavesPorNombre(registros) {
    const clavePorNombre = new Map();
    const nombresAmbiguos = new Set();
    for (const registro of registros) {
      if (!registro.rut) continue;
      const nombre = normalizarNombre(registro.nombrePaciente);
      if (!nombre) continue;
      const clave = `rut:${registro.rut}`;
      if (clavePorNombre.has(nombre) && clavePorNombre.get(nombre) !== clave) nombresAmbiguos.add(nombre);
      clavePorNombre.set(nombre, clave);
    }
    return registros.map((registro) => {
      if (registro.rut) return { registro, clave: obtenerClavePaciente(registro) };
      const nombre = normalizarNombre(registro.nombrePaciente);
      const clave = !nombresAmbiguos.has(nombre) && clavePorNombre.get(nombre);
      return { registro, clave: clave || obtenerClavePaciente(registro) };
    });
  }

  function fusionar(registrosTratamiento, registrosCancer) {
    const registrosPorPaciente = new Map();
    for (const { registro, clave } of asignarClavesPorNombre(registrosTratamiento)) {
      if (!registrosPorPaciente.has(clave)) registrosPorPaciente.set(clave, []);
      registrosPorPaciente.get(clave).push(registro);
    }
    const cancerPorPaciente = new Map();
    for (const registro of registrosCancer) {
      const clave = obtenerClavePaciente(registro);
      if (!cancerPorPaciente.has(clave)) cancerPorPaciente.set(clave, []);
      cancerPorPaciente.get(clave).push(registro);
    }

    const pacientes = [];
    const tratamientos = [];
    for (const [clave, registros] of registrosPorPaciente) {
      const registrosCancerPaciente = cancerPorPaciente.get(clave) || [];
      const conRut = registros.find((registro) => registro.rut);
      const nombrePaciente = (registros.find((registro) => registro.origen === 'Word') || registros[0]).nombrePaciente;
      const sexo = fusionarSexoPaciente([...registros, ...registrosCancerPaciente.map((registro) => ({ ...registro, sexoExplicito: registro.sexoRegistrado, origenSexoExplicito: 'Registro de cáncer' }))]);
      const paciente = { clave, nombrePaciente, rut: conRut?.rut ?? null, sexo: sexo.sexo, origenSexo: sexo.origenSexo, tratamientos: [] };
      const conflictoNombre = paciente.rut ? detectarNombresDistintos(registros) : null;

      agruparEnTratamientos(registros).forEach((grupo, indice) => {
        const tratamiento = construirTratamiento(grupo, paciente, indice);
        tratamiento.sexo = paciente.sexo;
        tratamiento.origenSexo = paciente.origenSexo;
        if (sexo.conflicto) tratamiento.conflictos.unshift(sexo.conflicto);
        if (conflictoNombre) tratamiento.conflictos.unshift(conflictoNombre);
        paciente.tratamientos.push(tratamiento);
        tratamientos.push(tratamiento);
      });
      pacientes.push(paciente);
    }

    const pacientesCancer = new Map();
    const registrarCancer = (clave, datos) => {
      const actual = pacientesCancer.get(clave);
      if (!actual || (datos.anio && (!actual.anio || datos.anio < actual.anio))) pacientesCancer.set(clave, { ...actual, ...datos, sexo: actual?.sexo ?? datos.sexo });
    };
    for (const [clave, registros] of cancerPorPaciente) {
      const anio = Math.min(...registros.map((registro) => registro.anio).filter(Boolean));
      const pacienteYodo = pacientes.find((paciente) => paciente.clave === clave);
      const sexosRegistro = [...new Set(registros.map((registro) => registro.sexoRegistrado).filter(Boolean))];
      registrarCancer(clave, { clave, anio: Number.isFinite(anio) ? anio : null, sexo: pacienteYodo ? pacienteYodo.sexo : sexosRegistro.length === 1 ? sexosRegistro[0] : null, enRegistro: true });
    }
    for (const tratamiento of tratamientos) {
      if (tratamiento.diagnostico === 'Cáncer de tiroides') registrarCancer(tratamiento.clavePaciente, { clave: tratamiento.clavePaciente, anio: tratamiento.anio, sexo: tratamiento.sexo, enInformes: true });
    }
    for (const paciente of pacientes) {
      const cancer = pacientesCancer.get(paciente.clave);
      paciente.tieneCancerTiroides = Boolean(cancer);
      for (const tratamiento of paciente.tratamientos) {
        tratamiento.pacienteConCancerRegistrado = Boolean(cancerPorPaciente.get(paciente.clave));
        tratamiento.numeroTratamientoPaciente = paciente.tratamientos.indexOf(tratamiento) + 1;
        tratamiento.cantidadTratamientosPaciente = paciente.tratamientos.length;
      }
    }

    return {
      pacientes,
      tratamientos,
      pacientesCancer: [...pacientesCancer.values()],
      conflictos: tratamientos.flatMap((tratamiento) => tratamiento.conflictos.map((conflicto) => ({ ...conflicto, idTratamiento: tratamiento.idTratamiento, nombrePaciente: tratamiento.nombrePaciente, rut: tratamiento.rut, fechaIso: tratamiento.fechaIso, anio: tratamiento.anio })))
        .filter((conflicto, indice, lista) => !['sexo', 'nombre'].includes(conflicto.campo) || lista.findIndex((otro) => otro.campo === conflicto.campo && otro.rut === conflicto.rut && otro.nombrePaciente === conflicto.nombrePaciente) === indice),
      registrosUnidos: tratamientos.reduce((suma, tratamiento) => suma + tratamiento.cantidadFuentes - 1, 0)
    };
  }

  ns.FusionTratamientos = { fusionar, VENTANA_MISMO_TRATAMIENTO_DIAS };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
