import { execFileSync } from 'node:child_process';
import { copyFileSync, mkdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { createHash } from 'node:crypto';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import * as ResEdit from 'resedit';

const carpetaPortable = path.dirname(fileURLToPath(import.meta.url));
const carpetaRaiz = path.join(carpetaPortable, '..');
const carpetaAplicacion = path.join(carpetaRaiz, 'Sistema - no tocar');
const carpetaSalida = path.join(carpetaPortable, 'salida');
const carpetaPaquete = path.join(carpetaSalida, 'Yodo131_Portable');
const rutaEjecutable = path.join(carpetaPaquete, 'Yodo131.exe');
const rutaZip = path.join(carpetaSalida, 'Yodo131_Portable.zip');
const NOMBRE_GUIA = 'LEEME - Yodo131 portable.txt';
const IDIOMA_ESPANOL_CHILE = 0x340a;
const PAGINA_CODIGOS_UNICODE = 1200;
const TIPO_RECURSO_MANIFIESTO = 24;
const MAQUINA_X64 = 0x8664;
const FORMATO_PE32_PLUS = 0x20b;
const SUBSISTEMA_VENTANAS = 2;

function ejecutar(comando, argumentos, entornoAdicional = {}) {
  execFileSync(comando, argumentos, { cwd: carpetaPortable, stdio: 'inherit', env: { ...process.env, ...entornoAdicional } });
}

function leerVersion() {
  const contenidoVersion = readFileSync(path.join(carpetaAplicacion, 'js', 'version.js'), 'utf8');
  const coincidencia = contenidoVersion.match(/VERSION_APLICACION = '(\d+\.\d+\.\d+)'/);
  if (!coincidencia) throw new Error('No se encontró la versión en js/version.js');
  const versionPaquete = JSON.parse(readFileSync(path.join(carpetaRaiz, 'package.json'), 'utf8')).version;
  if (versionPaquete !== coincidencia[1]) throw new Error(`La versión de package.json (${versionPaquete}) no coincide con version.js (${coincidencia[1]})`);
  return coincidencia[1];
}

function crearManifiesto(version) {
  return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<assembly xmlns="urn:schemas-microsoft-com:asm.v1" manifestVersion="1.0">
  <assemblyIdentity type="win32" name="Chiguin.Yodo131" version="${version}.0" processorArchitecture="amd64"/>
  <description>Lector Yodo 131 by Chiguin</description>
  <trustInfo xmlns="urn:schemas-microsoft-com:asm.v3">
    <security>
      <requestedPrivileges>
        <requestedExecutionLevel level="asInvoker" uiAccess="false"/>
      </requestedPrivileges>
    </security>
  </trustInfo>
  <compatibility xmlns="urn:schemas-microsoft-com:compatibility.v1">
    <application>
      <supportedOS Id="{8e0f7a12-bfb3-4fe8-b9a5-48fd50a15a9a}"/>
    </application>
  </compatibility>
  <application xmlns="urn:schemas-microsoft-com:asm.v3">
    <windowsSettings>
      <dpiAware xmlns="http://schemas.microsoft.com/SMI/2005/WindowsSettings">true/pm</dpiAware>
      <dpiAwareness xmlns="http://schemas.microsoft.com/SMI/2016/WindowsSettings">PerMonitorV2</dpiAwareness>
    </windowsSettings>
  </application>
</assembly>
`;
}

function aplicarRecursos(rutaExe, version) {
  const ejecutable = ResEdit.NtExecutable.from(readFileSync(rutaExe));
  const recursos = ResEdit.NtExecutableResource.from(ejecutable);
  const icono = ResEdit.Data.IconFile.from(readFileSync(path.join(carpetaPortable, 'recursos', 'icono.ico')));
  ResEdit.Resource.IconGroupEntry.replaceIconsForResource(recursos.entries, 1, IDIOMA_ESPANOL_CHILE, icono.icons.map((elemento) => elemento.data));

  const informacionVersion = ResEdit.Resource.VersionInfo.createEmpty();
  informacionVersion.setFileVersion(version, IDIOMA_ESPANOL_CHILE);
  informacionVersion.setProductVersion(version, IDIOMA_ESPANOL_CHILE);
  informacionVersion.setStringValues({ lang: IDIOMA_ESPANOL_CHILE, codepage: PAGINA_CODIGOS_UNICODE }, {
    FileDescription: 'Lector Yodo 131',
    ProductName: 'Lector Yodo 131',
    CompanyName: 'by Chiguin',
    LegalCopyright: 'by Chiguin',
    InternalName: 'Yodo131',
    OriginalFilename: 'Yodo131.exe',
    FileVersion: version,
    ProductVersion: version
  });
  informacionVersion.outputToResourceEntries(recursos.entries);

  const manifiesto = Buffer.from(crearManifiesto(version), 'utf8');
  recursos.entries.push({ type: TIPO_RECURSO_MANIFIESTO, id: 1, lang: IDIOMA_ESPANOL_CHILE, codepage: PAGINA_CODIGOS_UNICODE, bin: manifiesto.buffer.slice(manifiesto.byteOffset, manifiesto.byteOffset + manifiesto.byteLength) });
  recursos.outputResource(ejecutable);
  writeFileSync(rutaExe, Buffer.from(ejecutable.generate()));
}

function verificarEjecutable(rutaExe, version) {
  const datos = readFileSync(rutaExe);
  if (datos.toString('latin1', 0, 2) !== 'MZ') throw new Error('Yodo131.exe no es un ejecutable de Windows (falta la firma MZ)');
  const desplazamientoPe = datos.readUInt32LE(0x3c);
  if (datos.toString('latin1', desplazamientoPe, desplazamientoPe + 4) !== 'PE\0\0') throw new Error('Yodo131.exe no tiene encabezado PE');
  const maquina = datos.readUInt16LE(desplazamientoPe + 4);
  const desplazamientoOpcional = desplazamientoPe + 24;
  const formato = datos.readUInt16LE(desplazamientoOpcional);
  const subsistema = datos.readUInt16LE(desplazamientoOpcional + 68);
  if (maquina !== MAQUINA_X64 || formato !== FORMATO_PE32_PLUS) throw new Error('Yodo131.exe no es de 64 bits (x64)');
  if (subsistema !== SUBSISTEMA_VENTANAS) throw new Error('Yodo131.exe abriría una consola negra: el subsistema no es de ventanas');

  const recursos = ResEdit.NtExecutableResource.from(ResEdit.NtExecutable.from(datos));
  const [informacionVersion] = ResEdit.Resource.VersionInfo.fromEntries(recursos.entries);
  const valores = informacionVersion?.getStringValues({ lang: IDIOMA_ESPANOL_CHILE, codepage: PAGINA_CODIGOS_UNICODE }) || {};
  if (valores.ProductVersion !== version) throw new Error(`La versión del ejecutable (${valores.ProductVersion}) no es ${version}`);
  if (!recursos.entries.some((entrada) => entrada.type === 14)) throw new Error('Yodo131.exe no tiene ícono');
  const manifiesto = recursos.entries.find((entrada) => entrada.type === TIPO_RECURSO_MANIFIESTO);
  if (!manifiesto || !Buffer.from(manifiesto.bin).toString('utf8').includes('level="asInvoker"')) throw new Error('Yodo131.exe no declara que se ejecuta sin permisos de administrador');

  if (!datos.includes("const TEXTO_SELLO = 'by Chiguin';")) throw new Error('FIRMA DE AUTOR ALTERADA: el ejecutable no contiene TEXTO_SELLO = \'by Chiguin\'');
  if (!datos.includes(`VERSION_APLICACION = '${version}'`)) throw new Error(`El ejecutable no contiene la versión ${version}`);
  return { tamanoBytes: datos.length, descripcion: valores.FileDescription };
}

function calcularHuella(ruta) {
  return createHash('sha256').update(readFileSync(ruta)).digest('hex');
}

const version = leerVersion();
rmSync(carpetaSalida, { recursive: true, force: true });
mkdirSync(carpetaPaquete, { recursive: true });

ejecutar('go', ['run', './generador', carpetaAplicacion, path.join('lanzador', 'aplicacion_empaquetada.go')]);
ejecutar('go', ['vet', './...']);
ejecutar('go', ['vet', './...'], { GOOS: 'windows', GOARCH: 'amd64' });
ejecutar('go', ['test', './...']);
ejecutar('go', ['build', '-trimpath', '-buildvcs=false', '-ldflags', '-s -w -H windowsgui', '-o', rutaEjecutable, './lanzador'], { GOOS: 'windows', GOARCH: 'amd64', CGO_ENABLED: '0' });

aplicarRecursos(rutaEjecutable, version);
const { tamanoBytes } = verificarEjecutable(rutaEjecutable, version);
copyFileSync(path.join(carpetaPortable, NOMBRE_GUIA), path.join(carpetaPaquete, NOMBRE_GUIA));
ejecutar('go', ['run', './empaquetador', rutaZip, carpetaPaquete]);

const huellas = `SHA-256 de la versión ${version}\n${calcularHuella(rutaEjecutable)}  Yodo131.exe\n${calcularHuella(rutaZip)}  Yodo131_Portable.zip\n`;
writeFileSync(path.join(carpetaSalida, 'SHA256.txt'), huellas);
console.log(`\nYodo131.exe v${version} listo (${(tamanoBytes / 1024 / 1024).toFixed(1)} MB)\n${huellas}`);
