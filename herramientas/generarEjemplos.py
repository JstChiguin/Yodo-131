from pathlib import Path

from docx import Document

CARPETA_EJEMPLOS = Path(__file__).resolve().parent.parent / "ejemplos"

INFORMES_FICTICIOS = [
    {
        "archivo": "EJEMPLO_01_HIPERTIROIDISMO.docx", "nombre": "María Ejemplo Rojas", "rut": "11.111.111-1", "edad": 44, "fecha": "11/03/2024",
        "motivo": "En control en Endocrinología desde 2023. Cuadro de baja de peso, palpitaciones y temblor fino. En la actualidad con Tiamazol 10 mg/día.",
        "lineas": ["Ecotomografía tiroidea: tiroides aumentado de tamaño en forma difusa, sin nódulos.", "Ultima TSH: < 0,01 uUI/ml; T4L: 2,1 ng/dl; T3: 210 ng/dl; Trab: 12,4"],
        "familia": "madre", "morbidos": "(-)", "cigarrillo": "(-)", "gineco": True, "medico": "Dra. Ana Pérez", "dosis": "15 mC",
    },
    {
        "archivo": "EJEMPLO_02_CANCER_PAPILAR.docx", "nombre": "Carlos Ejemplo Soto", "rut": "12.222.222-2", "edad": 52, "fecha": "05/06/2024",
        "motivo": "Hace un año nota aumento de volumen cervical derecho de crecimiento lento.",
        "lineas": ["Ecotomografía tiroidea: en LD nódulo sólido hipoecogénico de 24 mm con microcalcificaciones. ACR Ti-rads 5",
                   "PAAF: Carcinoma papilar Bethesda VI", "02/04/2024: Tiroidectomía total",
                   "Biopsia: Carcinoma papilar clásico de 22 mm, márgenes libres. 2 de 8 ganglios con metástasis.",
                   "Estadificación: T2N1aMo Riesgo Intermedio", "Tiroglobulina: 3,2 ng/ml; Anticuerpos Antitiroglobulina: menos de 20 UI/ml"],
        "familia": "(-)", "morbidos": "HTA (Losartán)", "cigarrillo": "10/día hasta 2020", "gineco": False, "medico": "Dr. Luis Soto", "dosis": "100 mC",
    },
    {
        "archivo": "EJEMPLO_03_BOCIO_MULTINODULAR.docx", "nombre": "Rosa Ejemplo Díaz", "rut": "13.333.333-3", "edad": 67, "fecha": "20/09/2024",
        "motivo": "Derivada por bocio de larga data con TSH frenada en controles.",
        "lineas": ["Ecotomografía tiroidea: múltiples nódulos bilaterales, el mayor de 18 mm. Tirads 3", "PAAF: Bethesda II",
                   "Cintigrama tiroideo: bocio multinodular hiperfuncionante", "Ultima TSH: 0,05 uUI/ml; T4L: 1,6 ng/dl; T3: 160 ng/dl; Trab: 0,9"],
        "familia": "una hermana", "morbidos": "DM2 (Metformina)", "cigarrillo": "(-)", "gineco": True, "medico": "Dra Ana Pérez", "dosis": "30 mC",
    },
    {
        "archivo": "EJEMPLO_04_CANCER_SEGUIMIENTO.docx", "nombre": "Valentina Ejemplo Muñoz", "rut": "14.444.444-4", "edad": 35, "fecha": "14/01/2025",
        "motivo": "Antecedentes de cáncer papilar del tiroides tratado con tiroidectomía total en 2022. En controles presenta tiroglobulina en ascenso.",
        "lineas": ["Ecotomografía cervical: adenopatía sospechosa en grupo IV derecho de 9 mm.", "PAAF: Carcinoma papilar Bethesda VI",
                   "Tiroglobulina en ascenso de 2,1 a 8,7 ng/ml; Anticuerpos Antitiroglobulina: 45 UI/ml", "Estadificación: Riesgo Alto"],
        "familia": "(-)", "morbidos": "Asma bronquial", "cigarrillo": "(-)", "gineco": True, "medico": "Dr. Luis Soto", "dosis": "150 mC",
    },
    {
        "archivo": "EJEMPLO_05_GRAVES.docx", "nombre": "Pedro Ejemplo Vera", "rut": "15.555.555-5", "edad": 29, "fecha": "03/04/2025",
        "motivo": "Consulta por nerviosismo, taquicardia y baja de peso de 6 kg. Al examen exoftalmo bilateral.",
        "lineas": ["Ultima TSH: < 0,01 uUI/ml; T4L: 3,4 ng/dl; T3: 320 ng/dl; Trab: 18", "Cintigrama tiroideo: bocio difuso hiperfuncionante"],
        "familia": "(-)", "morbidos": "(-)", "cigarrillo": "5/día", "gineco": False, "medico": "Dra. Carla Muñoz", "dosis": "20 mC",
    },
    {
        "archivo": "EJEMPLO_06_CANCER_FOLICULAR.docx", "nombre": "Elena Ejemplo Castro", "rut": "16.666.666-6", "edad": 58, "fecha": "22/07/2025",
        "motivo": "Hallazgo de nódulo tiroideo palpable en control de medicina general.",
        "lineas": ["Ecotomografía tiroidea: nódulo sólido isoecogénico de 35 mm en LI. Tirads 4", "PAAF: Neoplasia folicular Bethesda IV",
                   "10/05/2025: Tiroidectomía total", "Biopsia: Carcinoma folicular mínimamente invasivo de 32 mm.",
                   "Estadificación: T2NxMo Riesgo Bajo", "Tiroglobulina: 0,4 ng/ml; Anti TPO: 12 UI/ml"],
        "familia": "(-)", "morbidos": "Hipotiroidismo", "cigarrillo": "(-)", "gineco": True, "medico": "Dr. Luis Soto", "dosis": "50 mC",
    },
    {
        "archivo": "EJEMPLO_07_HIPERTIROIDISMO_RECIDIVA.docx", "nombre": "Jorge Ejemplo Lagos", "rut": "17.777.777-7", "edad": 61, "fecha": "15/10/2025",
        "motivo": "Hipertiroidismo con recaída tras suspender Tiamazol en dos oportunidades. Palpitaciones y fatigabilidad.",
        "lineas": ["Ecotomografía tiroidea: leve aumento de tamaño, sin nódulos.", "Ultima TSH: 0,02 uUI/ml; T4L: 1,9 ng/dl; T3: 190 ng/dl; Trab: 6,5; Anti TPO: 120 UI/ml"],
        "familia": "un hermano", "morbidos": "HTA; Dislipidemia", "cigarrillo": "(-)", "gineco": False, "medico": "Dra. Carla Muñoz", "dosis": "15 mC",
    },
    {
        "archivo": "EJEMPLO_08_PLANTILLA_INCOMPLETA.docx", "nombre": "Isabel Ejemplo Fuentes", "rut": "", "edad": "", "fecha": "--/--/2026",
        "motivo": "En control en Endocrinología desde octubre 2025 ----------------.",
        "lineas": ["Ultima TSH: 0,3 uUI/ml; T4L: 1,2 ng/dl; T3: 140 ng/dl"],
        "familia": "-------", "morbidos": "-------", "cigarrillo": "-------", "gineco": False, "medico": "", "dosis": "30 mC",
    },
]


def crear_informe(informe):
    documento = Document()
    for encabezado in ["MINISTERIO DE SALUD", "SERVICIO DE SALUD", "CENTRO DE COSTO MEDICINA INTERNA", "HOSPITAL DE EJEMPLO"]:
        documento.add_paragraph(encabezado)
    documento.add_paragraph(f"NOMBRE: {informe['nombre']}")
    documento.add_paragraph(f"RUT: {informe['rut']}")
    documento.add_paragraph(f"EDAD: {informe['edad']} años")
    documento.add_paragraph(f"FECHA: {informe['fecha']}")
    documento.add_paragraph(informe["motivo"])
    for linea in informe["lineas"]:
        documento.add_paragraph(linea)
    documento.add_paragraph(f"Antecedentes familiares patología tiroidea: {informe['familia']}")
    documento.add_paragraph(f"Mórbidos: {informe['morbidos']}")
    documento.add_paragraph(f"Cigarrillo: {informe['cigarrillo']}; OH: (-); Drogas: (-)")
    if informe["gineco"]:
        documento.add_paragraph("Gesta: 2; Para: 2")
    documento.add_paragraph(f"Médico tratante: {informe['medico']}")
    documento.add_paragraph(f"En reunión de la especialidad se decide administrar {informe['dosis']} de Yodo 131")
    documento.add_paragraph("DRA. FIRMANTE DE EJEMPLO")
    documento.save(CARPETA_EJEMPLOS / informe["archivo"])


if __name__ == "__main__":
    CARPETA_EJEMPLOS.mkdir(exist_ok=True)
    for informe_ficticio in INFORMES_FICTICIOS:
        crear_informe(informe_ficticio)
    print(f"{len(INFORMES_FICTICIOS)} informes ficticios creados en {CARPETA_EJEMPLOS}")
