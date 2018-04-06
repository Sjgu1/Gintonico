package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/scrypt"
)

// función para comprobar errores (ahorra escritura)
func check(e error) {
	if e != nil {
		fmt.Println(e.Error())
	}
}

// Devuelve el string de la cadena encriptada
func encriptarScrypt(cadena string, seed string) string {
	salt := []byte(seed)

	dk, err := scrypt.Key([]byte(cadena), salt, 1<<15, 10, 1, 32)
	check(err)
	return base64.StdEncoding.EncodeToString(dk)
}

func encodeB64(cadena string) string {
	//StdEncoding
	return base64.URLEncoding.EncodeToString([]byte(cadena))
}

func decodeB64(cadena string) string {
	//StdEncoding
	decode, _ := base64.URLEncoding.DecodeString(cadena)
	return string(decode[:])
}

func hashSHA256(datos []byte) [32]byte {
	return sha256.Sum256(datos)
}

func streamToString(stream io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.String()
}
