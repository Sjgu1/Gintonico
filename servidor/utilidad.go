package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	mathrand "math/rand"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"golang.org/x/crypto/scrypt"
)

//UserStruct para el token
type UserStruct struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// función para comprobar errores (ahorra escritura)
func check(e error) {
	if e != nil {
		fmt.Println(e.Error())
	}
}

func createJWT(username string) string {
	// Embed User information to `token`
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), &UserStruct{
		Username: username})

	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	claims["iat"] = time.Now().Unix()
	claims["aud"] = username
	token.Claims = claims
	// token -> string. Only server knows this secret (foobar).
	clavemaestra := "{<J*l-&lG.f@GiNtOnIcO@B}%1ckFHb_"
	tokenstring, err := token.SignedString([]byte(clavemaestra))
	if err != nil {
		log.Fatalln(err)
	}
	return tokenstring
}

func validarToken(tokenRecibido string, username string) bool {
	clavemaestra := "{<J*l-&lG.f@GiNtOnIcO@B}%1ckFHb_"
	token, err := jwt.Parse(tokenRecibido, func(token *jwt.Token) (interface{}, error) {
		return []byte(clavemaestra), nil
	})
	//check(err)

	/*if claims["exp"].(float64) < float64(time.Now().Unix()) {
		//Aqui habria que deolver que el token ha expirado
		//fmt.Println(false)
		return false
	}*/
	if err != nil || token == nil { //ya valida tanto el tiempo de expiracion como si se ha firmado bien etc
		fmt.Println("Token incorrecto")
		return false
	}

	//claims := make(jwt.MapClaims)
	claims := token.Claims.(jwt.MapClaims)
	if claims["aud"].(string) != username {
		fmt.Println("Usuario de token incorrecto")
		return false
	}
	return true
}

// Devuelve el string de la cadena encriptada
func encriptarScrypt(cadena string, seed string) string {
	salt := []byte(seed)

	dk, err := scrypt.Key([]byte(cadena), salt, 1<<15, 10, 1, 32)
	check(err)
	return base64.StdEncoding.EncodeToString(dk)
}

func encodeURLB64(cadena string) string {
	return base64.URLEncoding.EncodeToString([]byte(cadena))
}

func decodeURLB64(cadena string) string {
	decode, _ := base64.URLEncoding.DecodeString(cadena)
	return string(decode[:])
}

const randomStringLetters = "0123abcdefghijklmnopqrstuvwxyz456ABCDEFGHIJKLMNOPQRSTUVWXYZ789"

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

func leerJSON(jsonNamefile string) []byte {
	// Abre el archivo json
	jsonFile, err := os.Open(jsonNamefile)
	// if we os.Open returns an error then handle it
	if err != nil {
		//fmt.Println(err)
		// detect if file exists
		var _, err = os.Stat(jsonNamefile)

		// create file if not exists
		if os.IsNotExist(err) {
			var file, err = os.Create(jsonNamefile)
			check(err)
			defer file.Close()
		}

		fmt.Println("==> done creating file", jsonNamefile)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)
	return byteValue
}
