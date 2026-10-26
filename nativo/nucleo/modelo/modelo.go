package modelo

type Medicion struct {
	Valor      float64
	Comparador string
}

type Tiroglobulina struct {
	ValorMaximo any
	EnAscenso   bool
	Mencionada  bool
}

type Objeto struct {
	D map[string]any
}

func Nuevo() Objeto {
	return Objeto{D: map[string]any{}}
}

func (o Objeto) Get(clave string) any {
	return o.D[clave]
}

func (o Objeto) Set(clave string, valor any) Objeto {
	o.D[clave] = valor
	return o
}

func (o Objeto) Tiene(clave string) bool {
	return o.D[clave] != nil
}

func (o Objeto) Texto(clave string) string {
	if valor, ok := o.D[clave].(string); ok {
		return valor
	}
	return ""
}

func (o Objeto) TextoOpcional(clave string) (string, bool) {
	valor, ok := o.D[clave].(string)
	return valor, ok
}

func (o Objeto) Bool(clave string) bool {
	valor, _ := o.D[clave].(bool)
	return valor
}

func (o Objeto) Entero(clave string) (int, bool) {
	valor, ok := o.D[clave].(int)
	return valor, ok
}

func (o Objeto) Numero(clave string) (float64, bool) {
	switch valor := o.D[clave].(type) {
	case float64:
		return valor, true
	case int:
		return float64(valor), true
	}
	return 0, false
}

func (o Objeto) Medicion(clave string) (Medicion, bool) {
	valor, ok := o.D[clave].(Medicion)
	return valor, ok
}

func (o Objeto) Tiroglobulina() Tiroglobulina {
	if valor, ok := o.D["tiroglobulina"].(Tiroglobulina); ok {
		return valor
	}
	return Tiroglobulina{}
}

func (o Objeto) Clonar() Objeto {
	copia := map[string]any{}
	for clave, valor := range o.D {
		copia[clave] = valor
	}
	return Objeto{D: copia}
}

type Conflicto struct {
	Campo        string
	NombreCampo  string
	Valores      []ValorConflicto
	IdTratamiento string
	NombrePaciente string
	Rut          any
	FechaIso     any
	Anio         any
}

type ValorConflicto struct {
	Origen     string
	Referencia string
	Valor      string
}

type Fuente struct {
	Origen     string
	Referencia string
	FechaIso   any
	Anio       any
}

type Paciente struct {
	Clave             string
	NombrePaciente    string
	Rut               any
	Sexo              any
	OrigenSexo        string
	Tratamientos      []Objeto
	TieneCancerTiroides bool
}

type PacienteCancer struct {
	Clave      string
	Anio       any
	Sexo       any
	EnRegistro bool
	EnInformes bool
}

type DocumentoNoLeido struct {
	Archivo string
	Motivo  string
}

type Conteo struct {
	ArchivosWord       int
	ArchivosExcel      int
	InformesWord       int
	FilasExcelYodo     int
	FilasRegistroCancer int
}

type Lectura struct {
	RegistrosLeidos    []Objeto
	RegistrosCancer    []Objeto
	DocumentosNoLeidos []DocumentoNoLeido
	Conteo             Conteo
	CantidadArchivos   int
}

type Resultado struct {
	Pacientes          []*Paciente
	Tratamientos       []Objeto
	PacientesCancer    []PacienteCancer
	Conflictos         []Conflicto
	RegistrosUnidos    int
	DocumentosNoLeidos []DocumentoNoLeido
	Conteo             Conteo
	CantidadArchivos   int
}

type Fecha struct {
	Iso string
}
