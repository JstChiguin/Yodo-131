package seguridad

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
)

const salHex = "e81dc673beaf9aed204c53f62e22a657"
const huellaHex = "cc5d58cdae068134fd66186eae0bf91a91adc8f69ad98fb186145cf2feb259e1"
const iteraciones = 310000
const largoClave = 32

func derivar(clave, sal []byte) []byte {
	prf := hmac.New(sha256.New, clave)
	largoHash := prf.Size()
	bloques := (largoClave + largoHash - 1) / largoHash
	resultado := make([]byte, 0, bloques*largoHash)
	contador := make([]byte, 4)
	for bloque := 1; bloque <= bloques; bloque++ {
		binary.BigEndian.PutUint32(contador, uint32(bloque))
		prf.Reset()
		prf.Write(sal)
		prf.Write(contador)
		u := prf.Sum(nil)
		t := make([]byte, len(u))
		copy(t, u)
		for n := 2; n <= iteraciones; n++ {
			prf.Reset()
			prf.Write(u)
			u = prf.Sum(nil)
			for i := range t {
				t[i] ^= u[i]
			}
		}
		resultado = append(resultado, t...)
	}
	return resultado[:largoClave]
}

func VerificarClave(claveIngresada string) bool {
	sal, err := hex.DecodeString(salHex)
	if err != nil {
		return false
	}
	huella, err := hex.DecodeString(huellaHex)
	if err != nil {
		return false
	}
	derivada := derivar([]byte(claveIngresada), sal)
	return subtle.ConstantTimeCompare(derivada, huella) == 1
}
