(function (ns) {
  const UMBRALES_POR_DEFECTO = Object.freeze({ trab: 1.75, tiroglobulina: 1.0, antiTiroglobulina: 20, antiTpo: 35 });

  const NOMBRES_FEMENINOS = new Set(('maria maría ana carmen rosa luisa patricia claudia carolina francisca javiera constanza catalina valentina camila fernanda daniela paula andrea alejandra veronica verónica marcela pamela paola macarena katherine catherine karen karina jessica jesica jacqueline jaqueline nicole nicol nathalie natalia sofia sofía antonia isidora florencia josefa martina trinidad gloria olga cecilia silvia teresa monica mónica sandra susana ximena lorena paulina gabriela marisol margarita elizabeth isabel ines inés amelia bernardita raquel cristina juana elena eva miriam myriam leonor rocio rocío fabiola tamara viviana yasna jimena jenny jennifer jenifer evelyn evelin angela ángela berta blanca carla carolina cinthya cynthia  debora débora elsa erika gisela hilda irma ivonne johanna johana julia laura lidia lilian lili loreto lucia lucía magdalena mariela marta nancy norma pilar priscila regina ruth sara  scarlett silvana tatiana valeria victoria yolanda            alexandra   ').split(/\s+/));
  const NOMBRES_MASCULINOS = new Set(('jose josé juan luis carlos jorge manuel pedro francisco antonio miguel ricardo sergio cristian cristián hector héctor raul raúl mario rodrigo patricio eduardo fernando alejandro andres andrés pablo diego felipe sebastian sebastián matias matías nicolas nicolás ignacio tomas tomás benjamin benjamín vicente joaquin joaquín gonzalo claudio marcelo mauricio victor víctor hugo oscar óscar rene rené ruben rubén roberto daniel david jaime jonathan  leonardo armando gabriel hernan hernán ivan iván enrique alberto arturo bastian bastián camilo cesar césar emilio esteban fabian fabián  german germán gustavo humberto jesus jesús julio  marco marcos nelson octavio oliver orlando osvaldo rafael ramon ramón samuel simon simón ulises waldo william   ').split(/\s+/));

  const CATEGORIAS_MOTIVO = [
    ['Seguimiento de cáncer ya tratado', /antecedentes? de (c[aá]ncer|tiroidectom)|tiroidectom[ií]a|recidiva|met[aá]stasis|c[aá]ncer papilar|carcinoma/i],
    ['Aumento de volumen cervical o nódulo', /aumento de volumen|se palpa|autopalpa|n[oó]dulo|bocio|masa cervical/i],
    ['Síntomas de hipertiroidismo', /palpitaci|temblor|baja de peso|nerviosismo|taquicard|diaforesis|sudoraci|insomnio|fatigabilidad/i],
    ['Oftalmopatía o exoftalmo', /oftalmopat|exoftalm/i],
    ['Control endocrinológico sin detalle', /control/i]
  ];

  function determinarSexoExplicito(informe) {
    if (informe.sexoRegistrado) return { sexoExplicito: informe.sexoRegistrado, origenSexoExplicito: 'Registrado en Excel' };
    if (informe.tieneDatosGinecologicos) return { sexoExplicito: 'F', origenSexoExplicito: 'Datos gineco-obstétricos' };
    return { sexoExplicito: null, origenSexoExplicito: null };
  }

  function determinarSexoInferido(informe) {
    const primerNombre = (informe.nombrePaciente || '').split(' ')[0]?.toLowerCase();
    if (NOMBRES_FEMENINOS.has(primerNombre)) return { sexoInferido: 'F', origenSexoInferido: 'Nombre' };
    if (NOMBRES_MASCULINOS.has(primerNombre)) return { sexoInferido: 'M', origenSexoInferido: 'Nombre' };
    if (/\b(derivada|la paciente|paciente femenina|operada)\b/i.test(informe.textoNarrativa || '')) return { sexoInferido: 'F', origenSexoInferido: 'Texto del informe' };
    if (/\b(derivado|operado)\b|(^|\s)él(\s|[,.])/i.test(informe.textoNarrativa || '')) return { sexoInferido: 'M', origenSexoInferido: 'Texto del informe' };
    return { sexoInferido: null, origenSexoInferido: null };
  }

  function determinarSexo(informe) {
    const explicito = determinarSexoExplicito(informe);
    const inferido = determinarSexoInferido(informe);
    return {
      ...explicito,
      ...inferido,
      sexo: explicito.sexoExplicito ?? inferido.sexoInferido,
      origenSexo: explicito.origenSexoExplicito ?? inferido.origenSexoInferido ?? 'Sin dato'
    };
  }

  function clasificarDiagnostico(textoClinico) {
    if (!textoClinico || !textoClinico.trim()) return null;
    if (/carcinoma|c[aá]ncer\s+(papilar|folicular|medular|de(l)?\s+tiroides|tiroideo)|comit[eé]\s+onc|^\s*ca\.?\s/i.test(textoClinico)) return 'Cáncer de tiroides';
    if (/multinodular|n[oó]dulo\s+(hiperfuncionante|caliente|t[oó]xico)/i.test(textoClinico)) return 'Bocio multinodular o nódulo hiperfuncionante';
    if (/graves|bocio difuso|hipertiroidismo|\btrab\b|thyrozol|tiamazol|tsh frenada/i.test(textoClinico)) return 'Hipertiroidismo (Graves / bocio difuso)';
    return 'Otro / no determinado';
  }

  function clasificarHistologia(informe) {
    const fuentes = [informe.textoBiopsia, informe.textoClinico];
    for (const texto of fuentes) {
      if (!texto) continue;
      if (/papilar[^.]{0,40}varie?(dad|nte)\s+folicular/i.test(texto)) return 'Papilar, variante folicular';
      if (/(carcinoma|c[aá]ncer)\s+papilar/i.test(texto)) return 'Papilar';
      if (/(carcinoma|c[aá]ncer)\s+folicular/i.test(texto)) return 'Folicular';
      if (/medular/i.test(texto)) return 'Medular';
      if (/anapl[aá]sico/i.test(texto)) return 'Anaplásico';
    }
    return 'No especificado';
  }

  function clasificarMotivo(textoNarrativa) {
    if (!textoNarrativa || !textoNarrativa.trim()) return null;
    const categoria = CATEGORIAS_MOTIVO.find(([, patron]) => patron.test(textoNarrativa));
    return categoria ? categoria[0] : 'Otro';
  }

  function quitarTildes(texto) {
    return texto.normalize('NFD').replace(/[\u0300-\u036f]/g, '');
  }

  function limpiarNombreMedico(textoMedico) {
    if (!textoMedico) return null;
    const sinRelleno = textoMedico.replace(/[-.…_]{3,}/g, ' ').replace(/\s+/g, ' ').trim();
    if (!sinRelleno) return null;
    const titulo = /^dra\b/i.test(sinRelleno) ? 'Dra.' : /^dr\b/i.test(sinRelleno) ? 'Dr.' : '';
    const nombre = sinRelleno.replace(/^dra?\.?\s*/i, '').toLowerCase()
      .replace(/(^|\s)(\p{L})/gu, (_, espacio, letra) => espacio + letra.toUpperCase());
    return nombre ? `${titulo} ${nombre}`.trim() : null;
  }

  function obtenerPalabrasNombre(nombreMedico) {
    return quitarTildes(nombreMedico.replace(/^dra?\.?\s*/i, '').toLowerCase()).split(' ').filter((palabra) => palabra.length > 1);
  }

  function unificarMedicos(registros) {
    const variantes = [...new Set(registros.map((registro) => registro.medicoTratante).filter(Boolean))];
    const palabrasPorVariante = new Map(variantes.map((variante) => [variante, obtenerPalabrasNombre(variante)]));
    const vecesPorVariante = new Map();
    for (const registro of registros) {
      if (registro.medicoTratante) vecesPorVariante.set(registro.medicoTratante, (vecesPorVariante.get(registro.medicoTratante) || 0) + 1);
    }
    const formaCompleta = new Map();
    for (const variante of variantes) {
      const palabras = palabrasPorVariante.get(variante);
      const candidatas = variantes.filter((otra) => {
        const palabrasOtra = palabrasPorVariante.get(otra);
        return palabras.length > 0 && palabras.every((palabra) => palabrasOtra.includes(palabra));
      });
      const largoMaximo = Math.max(...candidatas.map((candidata) => palabrasPorVariante.get(candidata).length));
      const masCompletas = candidatas.filter((candidata) => palabrasPorVariante.get(candidata).length === largoMaximo);
      const personasDistintas = new Set(masCompletas.map((candidata) => palabrasPorVariante.get(candidata).join(' ')));
      if (personasDistintas.size > 1) {
        formaCompleta.set(variante, variante);
        continue;
      }
      const cantidadTildes = (texto) => (texto.match(/[áéíóúñü]/gi) || []).length;
      masCompletas.sort((a, b) => cantidadTildes(b) - cantidadTildes(a) || vecesPorVariante.get(b) - vecesPorVariante.get(a) || a.localeCompare(b));
      formaCompleta.set(variante, masCompletas[0]);
    }
    const votosTitulo = new Map();
    for (const [variante, forma] of formaCompleta) {
      const titulo = (variante.match(/^dra?\./i) || [''])[0];
      if (!titulo) continue;
      const votos = votosTitulo.get(forma) || new Map();
      votos.set(titulo, (votos.get(titulo) || 0) + vecesPorVariante.get(variante));
      votosTitulo.set(forma, votos);
    }
    const conTituloMayoritario = (forma) => {
      const votos = votosTitulo.get(forma);
      if (!votos) return forma;
      const tituloGanador = [...votos].sort((a, b) => b[1] - a[1])[0][0];
      return `${tituloGanador} ${forma.replace(/^dra?\.\s*/i, '')}`;
    };
    return registros.map((registro) => (registro.medicoTratante ? { ...registro, medicoTratante: conTituloMayoritario(formaCompleta.get(registro.medicoTratante)) } : registro));
  }

  function evaluarElevado(medicion, umbral) {
    if (!medicion || medicion.valor == null) return null;
    if (medicion.comparador === '<') return medicion.valor > umbral ? null : false;
    return medicion.valor > umbral;
  }

  function evaluarTiroglobulina(tiroglobulina, umbral) {
    if (tiroglobulina.enAscenso) return true;
    if (tiroglobulina.valorMaximo == null) return null;
    return tiroglobulina.valorMaximo > umbral;
  }

  function agruparEdad(edad) {
    if (edad == null) return null;
    if (edad < 20) return '<20';
    if (edad >= 70) return '70+';
    const decada = Math.floor(edad / 10) * 10;
    return `${decada}–${decada + 9}`;
  }

  function agruparDosis(dosisMci) {
    if (dosisMci == null) return null;
    if (dosisMci <= 30) return 'Baja (≤30 mCi)';
    if (dosisMci < 100) return 'Intermedia (31–99 mCi)';
    return 'Alta (≥100 mCi)';
  }

  function listarCamposFaltantes(registro) {
    const faltantes = [];
    if (!registro.rut) faltantes.push('RUT');
    if (registro.edad == null) faltantes.push('edad');
    if (!registro.anio) faltantes.push('fecha');
    if (!registro.medicoTratante) faltantes.push('médico');
    if (registro.dosisMci == null && !registro.rechazoYodo) faltantes.push('dosis');
    if (!registro.sexo) faltantes.push('sexo');
    return faltantes;
  }

  function clasificarInforme(informe, umbrales = UMBRALES_POR_DEFECTO) {
    const { evaluarPresencia } = ns.ExtractorInforme;
    const diagnostico = clasificarDiagnostico(informe.textoClinico);
    const registro = {
      ...informe,
      ...determinarSexo(informe),
      diagnostico,
      histologia: diagnostico === 'Cáncer de tiroides' ? clasificarHistologia(informe) : null,
      motivoCategoria: clasificarMotivo(informe.textoNarrativa),
      medicoTratante: limpiarNombreMedico(informe.textoMedicoTratante),
      grupoEdad: agruparEdad(informe.edad),
      grupoDosis: informe.rechazoYodo && informe.dosisMci == null ? 'Rechazó / no recibió' : agruparDosis(informe.dosisMci),
      trabElevado: evaluarElevado(informe.trab, umbrales.trab),
      tiroglobulinaElevada: evaluarTiroglobulina(informe.tiroglobulina, umbrales.tiroglobulina),
      antiTiroglobulinaElevado: evaluarElevado(informe.antiTiroglobulina, umbrales.antiTiroglobulina),
      antiTpoElevado: evaluarElevado(informe.antiTpo, umbrales.antiTpo),
      tieneAntecedenteFamiliar: evaluarPresencia(informe.textoAntecedentesFamiliares),
      esFumador: evaluarPresencia(informe.textoCigarrillo)
    };
    registro.camposFaltantes = listarCamposFaltantes(registro);
    return registro;
  }

  ns.Clasificador = { clasificarInforme, unificarMedicos, limpiarNombreMedico, agruparEdad, agruparDosis, listarCamposFaltantes, UMBRALES_POR_DEFECTO };
})(globalThis.LectorYodo = globalThis.LectorYodo || {});
