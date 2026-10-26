//go:build windows

package dialogos

import (
	"syscall"
	"unsafe"
)

var (
	shell32              = syscall.NewLazyDLL("shell32.dll")
	ole32                = syscall.NewLazyDLL("ole32.dll")
	comdlg32             = syscall.NewLazyDLL("comdlg32.dll")
	procBrowse           = shell32.NewProc("SHBrowseForFolderW")
	procPathFromList     = shell32.NewProc("SHGetPathFromIDListW")
	procCoInitializeEx   = ole32.NewProc("CoInitializeEx")
	procCoTaskMemFree    = ole32.NewProc("CoTaskMemFree")
	procGetSaveFileNameW = comdlg32.NewProc("GetSaveFileNameW")
)

type browseInfoW struct {
	hwndOwner      uintptr
	pidlRoot       uintptr
	pszDisplayName *uint16
	lpszTitle      *uint16
	ulFlags        uint32
	lpfn           uintptr
	lParam         uintptr
	iImage         int32
}

type openFileNameW struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
	pvReserved        uintptr
	dwReserved        uint32
	flagsEx           uint32
}

const (
	bifReturnOnlyFsDirs = 0x0001
	bifNewDialogStyle   = 0x0040
	coinitApartment     = 0x2
	ofnOverwritePrompt  = 0x00000002
	ofnPathMustExist    = 0x00000800
	ofnNoChangeDir      = 0x00000008
	ofnExplorer         = 0x00080000
)

func iniciarCom() {
	procCoInitializeEx.Call(0, coinitApartment)
}

func ElegirCarpeta(titulo string) (string, bool) {
	iniciarCom()
	tituloPtr, _ := syscall.UTF16PtrFromString(titulo)
	bufferNombre := make([]uint16, 260)
	info := browseInfoW{
		pszDisplayName: &bufferNombre[0],
		lpszTitle:      tituloPtr,
		ulFlags:        bifReturnOnlyFsDirs | bifNewDialogStyle,
	}
	pidl, _, _ := procBrowse.Call(uintptr(unsafe.Pointer(&info)))
	if pidl == 0 {
		return "", false
	}
	defer procCoTaskMemFree.Call(pidl)
	bufferRuta := make([]uint16, 260)
	ok, _, _ := procPathFromList.Call(pidl, uintptr(unsafe.Pointer(&bufferRuta[0])))
	if ok == 0 {
		return "", false
	}
	return syscall.UTF16ToString(bufferRuta), true
}

func GuardarComo(titulo, nombreSugerido, carpetaInicial string) (string, bool) {
	bufferArchivo := make([]uint16, 32768)
	if nombreSugerido != "" {
		sugerido, _ := syscall.UTF16FromString(nombreSugerido)
		copy(bufferArchivo, sugerido)
	}
	filtro, _ := syscall.UTF16FromString("Libro de Excel (*.xlsx)")
	comodin, _ := syscall.UTF16FromString("*.xlsx")
	filtroCompleto := append(filtro, comodin...)
	filtroCompleto = append(filtroCompleto, 0)
	tituloPtr, _ := syscall.UTF16PtrFromString(titulo)
	defExt, _ := syscall.UTF16PtrFromString("xlsx")
	var carpetaPtr *uint16
	if carpetaInicial != "" {
		carpetaPtr, _ = syscall.UTF16PtrFromString(carpetaInicial)
	}
	ofn := openFileNameW{
		lpstrFilter:     &filtroCompleto[0],
		nFilterIndex:    1,
		lpstrFile:       &bufferArchivo[0],
		nMaxFile:        uint32(len(bufferArchivo)),
		lpstrInitialDir: carpetaPtr,
		lpstrTitle:      tituloPtr,
		flags:           ofnOverwritePrompt | ofnPathMustExist | ofnNoChangeDir | ofnExplorer,
		lpstrDefExt:     defExt,
	}
	ofn.lStructSize = uint32(unsafe.Sizeof(ofn))
	exito, _, _ := procGetSaveFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if exito == 0 {
		return "", false
	}
	return syscall.UTF16ToString(bufferArchivo), true
}
