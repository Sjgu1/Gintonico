package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	mathrand "math/rand"

	"os"

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

func encodeURLB64(cadena string) string {
	//StdEncoding
	return base64.URLEncoding.EncodeToString([]byte(cadena))
}

func decodeURLB64(cadena string) string {
	//StdEncoding
	decode, _ := base64.URLEncoding.DecodeString(cadena)
	return string(decode[:])
}

func encodeB64(cadena string) string {
	//StdEncoding
	return base64.StdEncoding.EncodeToString([]byte(cadena))
}

func decodeB64(cadena string) string {
	//StdEncoding
	decode, _ := base64.StdEncoding.DecodeString(cadena)
	return string(decode[:])
}

const randomStringLetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = randomStringLetters[mathrand.Intn(len(randomStringLetters))]
	}
	return string(b)
}

//Comprueba que los directorios no existen
func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func decryptAESCFB(data []byte, keystring string) []byte {
	// Byte array of the string
	ciphertext := data
	// Key
	key := []byte(keystring)

	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// Before even testing the decryption,
	// if the text is too small, then it is incorrect
	if len(ciphertext) < aes.BlockSize {
		panic("Text is too short")
	}

	// Get the 16 byte IV
	iv := ciphertext[:aes.BlockSize]

	// Remove the IV from the ciphertext
	ciphertext = ciphertext[aes.BlockSize:]

	// Return a decrypted stream
	stream := cipher.NewCFBDecrypter(block, iv)

	// Decrypt bytes from ciphertext
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext
}

func encryptAESCFB(data []byte, keystring string) []byte {
	// Byte array of the string
	plaintext := data

	// Key
	key := []byte(keystring)

	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// Empty array of 16 + plaintext length
	// Include the IV at the beginning
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	// Slice of first 16 bytes
	iv := ciphertext[:aes.BlockSize]

	// Write 16 rand bytes to fill iv
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	// Return an encrypted stream
	stream := cipher.NewCFBEncrypter(block, iv)

	// Encrypt bytes from plaintext to ciphertext
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext
}

func deleteFile(path string) {
	os.Remove(path)
	//check(err)
}
