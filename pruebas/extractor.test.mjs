import { test } from 'node:test';
import assert from 'node:assert/strict';
import { cargarModulos } from './cargarModulos.mjs';

const { ExtractorInforme, Clasificador } = cargarModulos();

function procesarTexto(texto) {
  const bloques = ExtractorInforme.dividirEnInformes(texto.trim().split('\n').map((linea) => linea.trim()));
  return bloques.map((lineas) => Clasificador.clasificarInforme(ExtractorInforme.extraerInforme(lineas, 'prueba.docx')));
}

const INFORME_HIPERTIROIDISMO = `
MINISTERIO DE SALUD
NOMBRE: Paciente Ficticio Uno
RUT: 11.111.111-1
EDAD: 44 años
FECHA: 11 /09/2026
En control desde abril 2024. Baja de peso, palpitaciones, temblor. Laboratorio muestra TSH frenada.
Ecotomografía tiroidea: leve aumento de tamaño del tiroides, sin nódulos.
Ultima TSH:< 0,01uUI /ml; T4L: 1,48 ng/dl; T3: 136ng/dl; Trab :22.5ng/dl
Antecedentes familiares patología tiroidea: madre y padre
Mórbidos: HTA (Losartan)
Cigarrillo: (-)   OH: (-); Drogas: (-)
Médico tratante: Dra. Ana Pérez
En reunión Endocrinología HCVB se decide administrar Yodo 131 en dosis de 15mC
`;

const INFORME_CANCER = `
NOMBRE: Paciente Ficticia Dos
RUT: 22222222-2
EDAD: 51 años
FECHA: /05/2022
Se autopalpa aumento de volumen cervical.
Ecotomografía tiroidea: nódulo sólido hipoecogénico ACR Ti-rads 5
PAAF: carcinoma papilar Behesda 6
Biopsia: Carcinoma papilar, variedad folicular. Total 12 ganglios.
Estadificación: T3N1Mx Riesgo Intermedio
05/05/2022: Tiroglobulina en ascenso, el 05 Mayo 2022 =1,56 ng/ml; Anticuerpos Antitiroglobulina: menos de 20 UI/ml
Antecedentes familiares patología tiroidea: (-)
Mórbidos: Cáncer de mama operado
Gesta: 2; Para: 2
Médico tratante: Dr. Luis Soto
En reunión de Comité Oncológico de tiroides decide administrar 15O mC de Yodo 131
`;

test('lee datos básicos y laboratorio de un informe de hipertiroidismo', () => {
  const [informe] = procesarTexto(INFORME_HIPERTIROIDISMO);
  assert.equal(informe.rut, '11111111-1');
  assert.equal(informe.edad, 44);
  assert.equal(informe.fechaIso, '2026-09-11');
  assert.deepEqual(informe.tsh, { valor: 0.01, comparador: '<' });
  assert.equal(informe.t4l.valor, 1.48);
  assert.equal(informe.trab.valor, 22.5);
  assert.equal(informe.trabElevado, true);
  assert.equal(informe.dosisMci, 15);
  assert.equal(informe.diagnostico, 'Hipertiroidismo (Graves / bocio difuso)');
  assert.equal(informe.motivoCategoria, 'Síntomas de hipertiroidismo');
  assert.equal(informe.ecoConNodulos, false, '"sin nódulos" no cuenta como nódulo');
  assert.equal(informe.tieneAntecedenteFamiliar, true);
  assert.equal(informe.esFumador, false);
  assert.equal(informe.medicoTratante, 'Dra. Ana Pérez');
});

test('lee cáncer con errores de tipeo frecuentes', () => {
  const [informe] = procesarTexto(INFORME_CANCER);
  assert.equal(informe.anio, 2022);
  assert.equal(informe.mes, 5);
  assert.equal(informe.fechaIso, null);
  assert.equal(informe.dosisMci, 150, '"15O" con letra O debe leerse como 150');
  assert.equal(informe.bethesda, 'VI');
  assert.equal(informe.tirads, 5);
  assert.equal(informe.diagnostico, 'Cáncer de tiroides');
  assert.equal(informe.histologia, 'Papilar, variante folicular');
  assert.equal(informe.riesgo, 'intermedio');
  assert.equal(informe.tiroglobulina.valorMaximo, 1.56);
  assert.equal(informe.tiroglobulinaElevada, true);
  assert.deepEqual(informe.antiTiroglobulina, { valor: 20, comparador: '<' });
  assert.equal(informe.antiTiroglobulinaElevado, false);
  assert.equal(informe.sexo, 'F');
  assert.equal(informe.motivoCategoria, 'Aumento de volumen cervical o nódulo');
});

test('un cáncer mencionado solo en mórbidos no cambia el diagnóstico tiroideo', () => {
  const [informe] = procesarTexto(INFORME_HIPERTIROIDISMO.replace('HTA (Losartan)', 'Carcinoma de próstata'));
  assert.equal(informe.diagnostico, 'Hipertiroidismo (Graves / bocio difuso)');
});

test('divide un Word con varios informes y descarta documentos sin formato', () => {
  assert.equal(procesarTexto(INFORME_HIPERTIROIDISMO + INFORME_CANCER).length, 2);
  assert.equal(procesarTexto('Revisión ficha clínica\nRUT: 1-9').length, 0);
});

test('campos de plantilla sin llenar quedan como sin dato', () => {
  const [informe] = procesarTexto(INFORME_HIPERTIROIDISMO
    .replace('RUT: 11.111.111-1', 'RUT: ')
    .replace('Antecedentes familiares patología tiroidea: madre y padre', 'Antecedentes familiares patología tiroidea: -------')
    .replace('Médico tratante: Dra. Ana Pérez', 'Médico tratante: '));
  assert.equal(informe.rut, null);
  assert.equal(informe.tieneAntecedenteFamiliar, null);
  assert.deepEqual(informe.camposFaltantes, ['RUT', 'médico', 'sexo']);
});

test('detecta rechazo de yodoterapia', () => {
  const [informe] = procesarTexto(`${INFORME_HIPERTIROIDISMO}\nPaciente rechaza yodoterapia.`);
  assert.equal(informe.rechazoYodo, true);
});

test('umbrales configurables cambian la evaluación', () => {
  const [informe] = procesarTexto(INFORME_HIPERTIROIDISMO);
  const reclasificado = Clasificador.clasificarInforme(informe, { ...Clasificador.UMBRALES_POR_DEFECTO, trab: 30 });
  assert.equal(reclasificado.trabElevado, false);
});

test('une variantes del nombre del mismo médico sin mezclar médicos distintos', () => {
  const nombres = ['Dra. Pérez', 'Dra Ana Pérez', 'DRA. ANA PEREZ', 'Dr. Luis Soto', 'Dra. Soto'];
  const registros = nombres.map((medico) => ({ medicoTratante: Clasificador.limpiarNombreMedico(medico) }));
  assert.deepEqual(Clasificador.unificarMedicos(registros).map((registro) => registro.medicoTratante),
    ['Dra. Ana Pérez', 'Dra. Ana Pérez', 'Dra. Ana Pérez', 'Dr. Luis Soto', 'Dr. Luis Soto']);
});

test('no toma el RUT de la firma cuando el paciente no tiene RUT', () => {
  const [informe] = procesarTexto(INFORME_HIPERTIROIDISMO.replace('RUT: 11.111.111-1', '') + 'DRA. FIRMANTE EJEMPLO\nRut: 9.999.999-9');
  assert.equal(informe.rut, null);
});
