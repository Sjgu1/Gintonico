package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"golang.org/x/crypto/scrypt"
)

// función para comprobar errores (ahorra escritura)
func check(e error) {
	if e != nil {
		fmt.Println(e.Error())
	}
}

func sendServerPetition(data map[string][]string, route string, username string) *http.Response {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	datos, err := json.Marshal(data)
	req, err := http.NewRequest("POST", "https://localhost:8081"+route, bytes.NewReader(datos))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Username", username)
	r, err := client.Do(req)

	//r, err := client.PostForm("https://localhost:8081"+route, data) // enviamos por POST
	//r.Header.Add()
	check(err)
	return r
}

// Devuelve el string de la cadena encriptada
func encriptarScrypt(cadena string, salt string) string {
	saltBytes := []byte(salt)

	dk, err := scrypt.Key([]byte(cadena), saltBytes, 1<<15, 10, 1, 32)
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

func hashSHA256(datos []byte) [32]byte {
	return sha256.Sum256(datos)
}

func streamToString(stream io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.String()
}

func createFile(path string) {
	// detect if file exists
	var _, err = os.Stat(path)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		check(err)
		defer file.Close()
	}

	//fmt.Println("==> done creating file", path)
}

func writeFile(path string, content string) {
	// open file using READ & WRITE permission
	var file, err = os.OpenFile(path, os.O_RDWR, 0644)
	check(err)
	defer file.Close()

	// write some text line-by-line to file
	_, err = file.WriteString(content)

	// save changes
	err = file.Sync()
	check(err)

	//fmt.Println("==> done writing to file")
}

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}
