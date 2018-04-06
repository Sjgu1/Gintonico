package main

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"

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

const randomStringLetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString(n int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = randomStringLetters[rand.Intn(len(randomStringLetters))]
	}
	return string(b)
}
