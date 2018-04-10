package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/h2non/filetype"

	"github.com/kabukky/httpscerts"
)

// respuesta del servidor
type resp struct {
	Ok  bool   `json:"ok"`  // true -> correcto, false -> error
	Msg string `json:"msg"` // mensaje adicional
}

// Estructura de usuarios
type Users struct {
	Users []User `json:"users"`
}

//Estructura de usuario
type User struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Salt     string `json:"salt"`
	Cifrado  string `json:"cifrado"`
}

//Estructura de bloque
type Block struct {
	Block string `json:"block"`
	Hash  string `json:"hash"`
}

//Estructura de bloque
type Blocks struct {
	Blocks []Block `json:"block"`
}

// función para escribir una respuesta del servidor
func response(w io.Writer, ok bool, msg string) {
	r := resp{Ok: ok, Msg: msg}    // formateamos respuesta
	rJSON, err := json.Marshal(&r) // codificamos en JSON
	check(err)                     // comprobamos error
	w.Write(rJSON)                 // escribimos el JSON resultante
}

func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	// Redirect the incoming HTTP request. Note that "127.0.0.1:8081" will only work if you are accessing the server from your local machine.
	http.Redirect(w, r, "https://127.0.0.1:8081"+r.RequestURI, http.StatusMovedPermanently)
}

func handler(w http.ResponseWriter, r *http.Request) {
	response(w, true, "Bienvenido a Gintónico")
}

func handlerLogin(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Paso por handlerLogin")

	r.ParseForm()                                // es necesario parsear el formulario
	w.Header().Set("Content-Type", "text/plain") // cabecera estándar

	if validarLogin(r.Form.Get("login"), r.Form.Get("password")) {
		response(w, true, "Logeado")
	} else {
		response(w, false, "Error al loguear")
	}
}

func validarLogin(login string, password string) bool {
	// Abre el archivo json
	jsonFile, err := os.Open("users.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
		// detect if file exists
		var _, err = os.Stat("users.json")

		// create file if not exists
		if os.IsNotExist(err) {
			var file, err = os.Create("users.json")
			check(err)
			defer file.Close()
		}

		fmt.Println("==> done creating file", "users.json")
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Users array
	var users Users

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &users)
	// Comprueba si algun usuario coincide con el del login

	for i := 0; i < len(users.Users); i++ {
		if login == users.Users[i].User && encriptarScrypt(password, users.Users[i].Salt) == users.Users[i].Password {
			return true
		}
	}
	return false

}

func handlerRegister(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()                                // es necesario parsear el formulario
	w.Header().Set("Content-Type", "text/plain") // cabecera estándar

	if validarRegister(r.Form.Get("register"), r.Form.Get("password"), r.Form.Get("confirm")) {
		response(w, true, "Registrado")
	} else {
		response(w, false, "Error al registrar")
	}
}

func validarRegister(register string, password string, confirm string) bool {
	//Las password no coinciden
	if password != confirm || register == "" || password == "" {
		return false
	}

	//El usuario ya existe en la base de datos
	if comprobarExisteUsuario(register) {
		return false
	}

	// Abre el archivo json
	jsonFile, err := os.Open("users.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
		var _, err = os.Stat("users.json")

		// create file if not exists
		if os.IsNotExist(err) {
			var file, err = os.Create("users.json")
			check(err)
			defer file.Close()
		}
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Users array
	var users Users

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &users)

	salt := randomString(32)
	cifrado := randomString(32)

	users.Users = append(users.Users, User{User: register, Password: encriptarScrypt(password, salt), Salt: salt, Cifrado: cifrado})

	usersJSON, _ := json.Marshal(users)
	err = ioutil.WriteFile("users.json", usersJSON, 0644)

	// IMPRIMIR USUARIOS
	// now Marshal it
	check(err)

	return true
}

func comprobarExisteUsuario(usuario string) bool {
	// Abre el archivo json
	jsonFile, err := os.Open("users.json")
	check(err)
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Users array
	var users Users

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &users)

	// Comprueba si algun usuario coincide con el del login
	for i := 0; i < len(users.Users); i++ {
		if usuario == users.Users[i].User {
			return true
		}
	}
	return false
}

func handlerHash(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("entro handlerHash")
	r.ParseForm()                                // es necesario parsear el formulario
	w.Header().Set("Content-Type", "text/plain") // cabecera estándar

	contador, _ := strconv.Atoi(r.Form.Get("cont"))  // numero del orden de la parte del fichero
	hash := r.Form.Get("hash")                       // hash de la parte del fichero
	size, _ := strconv.Atoi(r.Form.Get("size"))      // tamaño de la parte del fichero
	user := r.Form.Get("user")                       // usuario que sube el fichero
	filename := decodeURLB64(r.Form.Get("filename")) // nombre del fichero original

	comprobar := comprobarHash(contador, hash, size, user, filename)
	fmt.Println("Hash recibido: " + hash + " usuario: " + user + " filename: " + filename)
	response(w, comprobar, "Hash comprobado")
}

func comprobarHash(cont int, hash string, tam int, user string, filename string) bool {
	//buscar el hash en la base de datos:
	//si ya existe, hay que asociar ese hash existente con el usuario al que pertence y tal y se devuelve true
	//si no existe, entonces simplemente se devuelve false

	return false //esta puesto en false para que el cliente ahora siempre suba los ficheros (para pruebas y tal)
}

func handlerUpload(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Paso por handlerUpload")

	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("uploadfile")
	check(err)
	defer file.Close()
	fmt.Fprintf(w, "%v", handler.Header)
	// Split on /.
	fichero := decodeURLB64(handler.Filename) + ".part" + r.FormValue("Parte")
	fmt.Println(fichero)
	createDirIfNotExist("./archivos/")
	//createDirIfNotExist("./archivos/" + r.FormValue("Username"))

	//path := "./archivos/" + r.FormValue("Username") + "/" + fichero
	last := getNombreUlitmoFichero() + 1
	path := strconv.Itoa(last)

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	check(err)
	defer f.Close()
	io.Copy(f, file)
}

func handlerShowUserFiles(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Paso por handlerUser")

	u, err := url.Parse(r.URL.String())
	check(err)
	result := strings.Split(u.Path, "/")
	createDirIfNotExist("./archivos/" + result[len(result)-1])
	files, err := ioutil.ReadDir("./archivos/" + result[len(result)-1] + "/")
	check(err)

	s := make([]string, len(files))
	for i, f := range files {
		s[i] = encodeURLB64(f.Name())
	}

	slc, _ := json.Marshal(s)
	w.Write(slc)
}

func handlerSendFile(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Paso por handlerFiles")

	u, err := url.Parse(r.URL.String())
	check(err)
	result := strings.Split(u.Path, "/")
	fmt.Println(result)
	if _, err := os.Stat("./archivos/" + result[len(result)-3] + "/" + decodeURLB64(result[len(result)-1])); err == nil {

		// grab the generated receipt.pdf file and stream it to browser
		streamBytes, err := ioutil.ReadFile("./archivos/" + result[len(result)-3] + "/" + decodeURLB64(result[len(result)-1]))
		check(err)

		kind, unknown := filetype.Match(streamBytes)
		if unknown != nil {
			fmt.Printf("Unknown: %s", unknown)
			return
		}

		fmt.Printf("File type: %s. MIME: %s\n", kind.Extension, kind.MIME.Value)
		b := bytes.NewBuffer(streamBytes)

		// stream straight to client(browser)

		w.Header().Set("Content-type", kind.MIME.Value)

		if _, err := b.WriteTo(w); err != nil { // <----- here!
			fmt.Fprintf(w, "%s", err)
		}

	} else {
		response(w, true, "El archivo No Existe")
	}

}

func getNombreUlitmoFichero() int {
	// Abre el archivo json
	jsonFile, err := os.Open("blocks.json")
	// if we os.Open returns an error then handle it
	check(err)
	//defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Blocks array
	var blocks Blocks

	json.Unmarshal(byteValue, &blocks)
	if len(blocks.Blocks) == 0 {
		return -1
	} else {
		result, err := strconv.Atoi(blocks.Blocks[len(blocks.Blocks)-1].Block)
		check(err)
		return result
	}

}

func cifrarFicherosUsuarios() {
	//recorrer todos los ficheros y cifrarlos con una contraseña maestra
	err := filepath.Walk("./archivos", visitEncrypt) //esta funcion recorre todos los directorios y ficheros recursivamente
	check(err)
}

func visitEncrypt(path string, f os.FileInfo, err error) error { //funcion para cifrarFicherosUsuarios
	if f != nil && f.IsDir() == false { //para coger solo los ficheros y no las carpetas
		clavemaestra := "{<J*l-&lG.f@GiNtOnIcO@B}%1ckFHb_" //32 bytes para que sea AES256
		cifrarFichero(path, clavemaestra)
	}
	return nil
}

func cifrarFichero(path string, clave string) {
	file, err := ioutil.ReadFile(path)
	check(err)

	if len(file) > 0 {
		encryptedFile := encryptAESCFB(file, clave)

		deleteFile(path)

		filenew, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
		check(err)
		defer filenew.Close()
		io.Copy(filenew, bytes.NewReader(encryptedFile))
	}
}

func descifrarFicherosUsuarios() {
	//recorrer todos los ficheros y cifrarlos con una contraseña maestra
	err := filepath.Walk("./archivos", visitDecrypt) //esta funcion recorre todos los directorios y ficheros recursivamente
	check(err)
}

func visitDecrypt(path string, f os.FileInfo, err error) error {
	//funcion para descifrarFicherosUsuarios
	if f != nil && f.IsDir() == false { //para coger solo los ficheros y no las carpetas
		clavemaestra := "{<J*l-&lG.f@GiNtOnIcO@B}%1ckFHb_" //32 bytes para que sea AES256
		descifrarFichero(path, clavemaestra)
	}
	return nil
}

func descifrarFichero(path string, clave string) {
	file, err := ioutil.ReadFile(path)
	check(err)

	if len(file) > 0 {
		encryptedFile := decryptAESCFB(file, clave)

		deleteFile(path)

		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
		check(err)
		defer f.Close()
		io.Copy(f, bytes.NewReader(encryptedFile))
	}
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano()) //para que el aleatorio funcione bien
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	// Comprueba los certificados, si no existen se generan nuevos
	err := httpscerts.Check("cert.pem", "key.pem")

	if err != nil {
		err = httpscerts.Generate("cert.pem", "key.pem", ":8081")
		if err != nil {
			log.Fatal("Error: No se han podido crear los certificados https.")
		}
	}

	muxa := mux.NewRouter()
	muxa.HandleFunc("/", handler)
	muxa.HandleFunc("/login", handlerLogin)
	muxa.HandleFunc("/register", handlerRegister)
	muxa.HandleFunc("/checkhash", handlerHash)
	muxa.HandleFunc("/upload", handlerUpload)
	muxa.HandleFunc("/user/{username}", handlerShowUserFiles)
	muxa.HandleFunc("/user/{username}/file/{filename}", handlerSendFile)

	srv := &http.Server{Addr: ":8081", Handler: muxa}

	go func() {
		log.Println("Poniendo en marcha servidor HTTPS, escuchando puerto 8081")
		if err := srv.ListenAndServeTLS("cert.pem", "key.pem"); err != nil {
			log.Printf("Error al poner en funcionamiento el servidor TLS: %s\n", err)
		}
	}()
	go func() {
		log.Println("Poniendo en marcha redireccionamiento HTTP->HTTPS, escuchando puerto 8080")
		if err := http.ListenAndServe(":8080", http.HandlerFunc(redirectToHTTPS)); err != nil {
			log.Printf("Error al redireccionar http a https: %s\n", err)
		}
	}()

	log.Println("Descifrando ficheros...")
	descifrarFicherosUsuarios()
	descifrarFichero("users.json", "{<J*l-&lG.f@GiNtOnIcO@B}%1ckFHb_")

	<-stopChan // espera señal SIGINT
	log.Println("Apagando servidor ...")
	// apagar servidor de forma segura
	ctx, fnc := context.WithTimeout(context.Background(), 5*time.Second)
	fnc()
	srv.Shutdown(ctx)

	log.Println("Cifrando ficheros...")
	cifrarFicherosUsuarios()
	cifrarFichero("./users.json", "{<J*l-&lG.f@GiNtOnIcO@B}%1ckFHb_")

	log.Println("Servidor detenido correctamente")
}
