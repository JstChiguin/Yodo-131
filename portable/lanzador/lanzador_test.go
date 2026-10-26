package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLaAplicacionEmpaquetadaTraeLaFirmaYLaVersion(t *testing.T) {
	selloAutor, existe := archivosAplicacion["js/selloAutor.js"]
	if !existe || !strings.Contains(selloAutor, "const TEXTO_SELLO = 'by Chiguin';") {
		t.Fatal("FIRMA DE AUTOR ALTERADA: la aplicación empaquetada no conserva TEXTO_SELLO = 'by Chiguin'")
	}
	version, errorVersion := obtenerVersionAplicacion(archivosAplicacion)
	if errorVersion != nil || version == "" {
		t.Fatalf("no se obtuvo la versión: %v", errorVersion)
	}
}

func TestObtenerVersionRechazaArchivoSinVersion(t *testing.T) {
	_, errorVersion := obtenerVersionAplicacion(map[string]string{"js/version.js": "sin versión"})
	if !errors.Is(errorVersion, errVersionNoEncontrada) {
		t.Fatalf("se esperaba errVersionNoEncontrada y se obtuvo %v", errorVersion)
	}
}

func TestExtraerAplicacionEscribeSoloLoQueCambio(t *testing.T) {
	carpetaDestino := t.TempDir()
	archivosFicticios := map[string]string{"index.html": "<p>inicio</p>", "js/version.js": "ns.VERSION_APLICACION = '9.9.9';"}

	cantidadEscritos, errorExtraccion := extraerAplicacion(archivosFicticios, carpetaDestino)
	if errorExtraccion != nil || cantidadEscritos != 2 {
		t.Fatalf("primera extracción: %d archivos, error %v", cantidadEscritos, errorExtraccion)
	}
	cantidadEscritos, _ = extraerAplicacion(archivosFicticios, carpetaDestino)
	if cantidadEscritos != 0 {
		t.Fatalf("una segunda extracción sin cambios no debe reescribir; reescribió %d", cantidadEscritos)
	}
	rutaIndice := filepath.Join(carpetaDestino, "index.html")
	os.WriteFile(rutaIndice, []byte("alterado"), 0o644)
	cantidadEscritos, _ = extraerAplicacion(archivosFicticios, carpetaDestino)
	contenido, _ := os.ReadFile(rutaIndice)
	if cantidadEscritos != 1 || string(contenido) != "<p>inicio</p>" {
		t.Fatalf("un archivo alterado debe restaurarse; escritos %d, contenido %q", cantidadEscritos, contenido)
	}
	temporales, _ := filepath.Glob(filepath.Join(carpetaDestino, ".yodo131-*"))
	if len(temporales) != 0 {
		t.Fatalf("quedaron archivos temporales: %v", temporales)
	}
}

func TestExtraerAplicacionRechazaRutasFueraDeLaCarpeta(t *testing.T) {
	for _, rutaPeligrosa := range []string{"../fuera.txt", "/absoluta.txt", `js\otro.js`, "C:/x.txt", "js//doble.js"} {
		if _, errorExtraccion := extraerAplicacion(map[string]string{rutaPeligrosa: "x"}, t.TempDir()); errorExtraccion == nil {
			t.Errorf("la ruta %q debía rechazarse", rutaPeligrosa)
		}
	}
}

func TestCandidatosWindowsPrefierenEdgeYLuegoChrome(t *testing.T) {
	variables := map[string]string{
		"ProgramFiles(x86)": `C:\Program Files (x86)`,
		"ProgramFiles":      `C:\Program Files\`,
		"LOCALAPPDATA":      `C:\Users\Ejemplo\AppData\Local`,
	}
	candidatos := candidatosNavegadorWindows(func(nombre string) string { return variables[nombre] })
	esperados := []string{
		`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
		`C:\Users\Ejemplo\AppData\Local\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		`C:\Users\Ejemplo\AppData\Local\Google\Chrome\Application\chrome.exe`,
	}
	if strings.Join(candidatos, "|") != strings.Join(esperados, "|") {
		t.Fatalf("orden de candidatos inesperado:\n%s", strings.Join(candidatos, "\n"))
	}
	soloChrome, encontrado := elegirNavegador(candidatos, func(ruta string) bool { return strings.HasSuffix(ruta, "chrome.exe") })
	if !encontrado || soloChrome != esperados[3] {
		t.Fatalf("sin Edge debía elegirse Chrome; se eligió %q", soloChrome)
	}
	if _, encontrado := elegirNavegador(candidatos, func(string) bool { return false }); encontrado {
		t.Fatal("sin navegadores no debía elegirse ninguno")
	}
}

func TestConvertirRutaEnDireccion(t *testing.T) {
	casos := map[string]string{
		`C:\Users\José Pérez\AppData\Local\Yodo131\aplicacion\1.6.0\index.html`: "file:///C:/Users/Jos%C3%A9%20P%C3%A9rez/AppData/Local/Yodo131/aplicacion/1.6.0/index.html",
		"/home/ejemplo/.cache/Yodo131/aplicacion/1.6.0/index.html":               "file:///home/ejemplo/.cache/Yodo131/aplicacion/1.6.0/index.html",
	}
	for ruta, esperada := range casos {
		if obtenida := convertirRutaEnDireccion(ruta); obtenida != esperada {
			t.Errorf("%s\n  obtenido: %s\n  esperado: %s", ruta, obtenida, esperada)
		}
	}
}

func TestArgumentosAbrenVentanaAisladaYSinRed(t *testing.T) {
	argumentos := strings.Join(construirArgumentosNavegador("file:///C:/app/index.html", `C:\perfil`), " ")
	for _, obligatorio := range []string{"--app=file:///C:/app/index.html", `--user-data-dir=C:\perfil`, "--no-first-run", "--disable-extensions", "--disable-sync", "--disable-background-networking"} {
		if !strings.Contains(argumentos, obligatorio) {
			t.Errorf("falta el argumento %s", obligatorio)
		}
	}
}

func TestRegistroDeErrorNoGuardaLaCarpetaPersonal(t *testing.T) {
	carpetaPersonal, _ := os.UserHomeDir()
	registro := crearRegistroError("lanzador", "iniciarAplicacion", errors.New("falló "+filepath.Join(carpetaPersonal, "x")), "1.0.0", time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	if strings.Contains(registro.Mensaje, carpetaPersonal) {
		t.Fatalf("el registro expone la carpeta personal: %s", registro.Mensaje)
	}
	if registro.FechaHora != "2026-01-02T03:04:05Z" || registro.TipoExcepcion != "*errors.errorString" || registro.VersionAplicacion != "1.0.0" {
		t.Fatalf("registro incompleto: %+v", registro)
	}
}
