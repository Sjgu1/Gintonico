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
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/kabukky/httpscerts"
)

// respuesta del servidor
type resp struct {
	Ok  bool   `json:"ok"`  // true -> correcto, false -> error
	Msg string `json:"msg"` // mensaje adicional
}

// Users Estructura de usuarios
type Users struct {
	Users []User `json:"users"`
}

// User Estructura de usuario
type User struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Salt     string `json:"salt"`
	Cifrado  string `json:"cifrado"`
	Token    string `json:"token"`
}

// Block Estructura de bloque
type Block struct {
	Block string `json:"block"`
	Hash  string `json:"hash"`
	User  string `json:"user"`
}

// Blocks Estructura de bloque
type Blocks struct {
	Blocks []Block `json:"blocks"`
}

//BlockPosition Posicion del bloque
type BlockPosition struct {
	Block    string `json:"block"`
	Position string `json:"position"`
}

// File Estructura de file
type File struct {
	User  string          `json:"user"`
	File  string          `json:"file"`
	Order []BlockPosition `json:"order"`
}

//Files estructura de files
type Files struct {
	Files []File `json:"files"`
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

//Person strutc
type Person struct {
	Login    []string `json:"login"`
	Password []string `json:"password"`
}

func handlerLogin(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Paso por handlerLogin")
	r.ParseForm()
	// es necesario parsear el formulario
	w.Header().Set("Content-Type", "text/plain") // cabecera estándar

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	s := buf.String()
	bytes := []byte(s)

	var u Person
	err := json.Unmarshal(bytes, &u)
	check(err)
	if validarLogin(u.Login[0], u.Password[0]) {

		token := createJWT(u.Login[0])
		w.Header().Add("Token", token)
		guardarToken(token, u.Login[0])
		//validarToken(token, r.Form.Get("login"))
		response(w, true, token)

	} else {
		response(w, false, "Error al loguear")
	}
}

func validarLogin(login string, password string) bool {
	jsonBytes := leerJSON("./databases/users.json")
	var users Users
	json.Unmarshal(jsonBytes, &users)

	// Comprueba si algun usuario coincide con el del login
	for i := 0; i < len(users.Users); i++ {
		if login == users.Users[i].User && encriptarScrypt(password, users.Users[i].Salt) == users.Users[i].Password {
			return true
		}
	}
	return false
}

func guardarToken(token string, user string) {
	jsonBytes := leerJSON("./databases/users.json")
	var users Users
	json.Unmarshal(jsonBytes, &users)

	var contador = 0
	for i := 0; i < len(users.Users); i++ {
		if users.Users[i].User == user {
			contador = i
		}
	}
	users.Users[contador].Token = token
	usersJSON, _ := json.Marshal(users)
	err := ioutil.WriteFile("./databases/users.json", usersJSON, 0644)
	check(err)
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

	jsonBytes := leerJSON("./databases/users.json")
	var users Users
	json.Unmarshal(jsonBytes, &users)

	salt := randomString(32)
	cifrado := randomString(32)

	users.Users = append(users.Users, User{User: register, Password: encriptarScrypt(password, salt), Salt: salt, Cifrado: cifrado})

	usersJSON, _ := json.Marshal(users)
	err := ioutil.WriteFile("./databases/users.json", usersJSON, 0644)
	check(err)

	return true
}

func comprobarExisteUsuario(usuario string) bool {
	jsonBytes := leerJSON("./databases/users.json")
	var users Users
	json.Unmarshal(jsonBytes, &users)

	// Comprueba si algun usuario coincide con el del login
	for i := 0; i < len(users.Users); i++ {
		if usuario == users.Users[i].User {
			return true
		}
	}
	return false
}

func handlerHash(w http.ResponseWriter, r *http.Request) {
	validarToken(r.Header.Get("Authorization"), r.Header.Get("Username"))

	//fmt.Println("entro handlerHash")
	r.ParseForm()                                // es necesario parsear el formulario
	w.Header().Set("Content-Type", "text/plain") // cabecera estándar

	contador, _ := strconv.Atoi(r.Form.Get("cont"))  // numero del orden de la parte del fichero
	hash := r.Form.Get("hash")                       // hash de la parte del fichero
	size, _ := strconv.Atoi(r.Form.Get("size"))      // tamaño de la parte del fichero
	user := r.Form.Get("user")                       // usuario que sube el fichero
	filename := decodeURLB64(r.Form.Get("filename")) // nombre del fichero original

	comprobar := comprobarHash(contador, hash, size, user, filename)
	//fmt.Println("Hash recibido: " + hash + " usuario: " + user + " filename: " + filename)
	response(w, comprobar, "Hash comprobado")
}

func handlerValidarToken(w http.ResponseWriter, r *http.Request) {
	//	tokenRecibido := r.Header.Get("Authorization")
	//	clavemaestra := "{<J*l-&lG.f@GiNtOnIcO@B}%1ckFHb_"
	//	valido := validarToken(tokenRecibido, r.Header.Get("Username"))

}

func comprobarHash(cont int, hash string, tam int, user string, filename string) bool {
	//buscar el hash en la base de datos:
	//si ya existe, hay que asociar ese hash existente con el usuario al que pertence y tal y se devuelve true
	//si no existe, entonces simplemente se devuelve false
	parte := strconv.Itoa(cont)

	var position BlockPosition
	existeBloque, nombreBloque := existeBloqueHash(hash)

	if existeBloque {
		position.Block = nombreBloque
		position.Position = parte
		registrarBloqueFicheroUsuario(user, filename, position)
		return true
	}
	return false
}

func handlerUpload(w http.ResponseWriter, r *http.Request) {
	validarToken(r.Header.Get("Authorization"), r.Header.Get("Username"))

	//fmt.Println("Paso por handlerUpload")
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("uploadfile")
	check(err)
	defer file.Close()
	//fmt.Fprintf(w, "%v", handler.Header)

	//Se crea un bloque con los datos recibidos
	var position BlockPosition

	last := getNombreUltimoFichero()
	value, err := strconv.Atoi(last)
	value++
	path := strconv.Itoa(value)

	f, err := os.OpenFile("./archivos/"+path, os.O_WRONLY|os.O_CREATE, 0666)
	check(err)
	defer f.Close()
	io.Copy(f, file)
	position.Block = path
	position.Position = r.FormValue("Parte")

	//Se registra el bloque en la base de datos
	var block Block
	block.User = r.FormValue("Username")
	block.Hash = r.FormValue("Hash")
	block.Block = path
	registrarBloque(block)
	//Se le asigna el bloque al par fichero-usuario
	registrarBloqueFicheroUsuario(r.FormValue("Username"), decodeURLB64(handler.Filename), position)
	cifrarFichero("./archivos/"+path, obtenerClaveCifrado("./archivos/"+path))
}

func registrarBloque(bloque Block) {
	jsonBytes := leerJSON("./databases/blocks.json")
	var blocks Blocks
	json.Unmarshal(jsonBytes, &blocks)

	blocks.Blocks = append(blocks.Blocks, Block{Block: bloque.Block, Hash: bloque.Hash, User: bloque.User})

	blocksJSON, _ := json.Marshal(blocks)
	err := ioutil.WriteFile("./databases/blocks.json", blocksJSON, 0644)
	check(err)
}

func existeBloqueHash(hash string) (bool, string) {
	jsonBytes := leerJSON("./databases/blocks.json")
	var blocks Blocks
	json.Unmarshal(jsonBytes, &blocks)

	// Comprueba si algun usuario coincide con el del login
	for i := 0; i < len(blocks.Blocks); i++ {
		if hash == blocks.Blocks[i].Hash {
			return true, blocks.Blocks[i].Block
		}
	}
	return false, "nil"
}

func existeFicheroUsuario(usuario string, fichero string) bool {
	jsonBytes := leerJSON("./databases/files.json")
	var files Files
	json.Unmarshal(jsonBytes, &files)
	// Comprueba si algun usuario coincide con el del login

	for i := 0; i < len(files.Files); i++ {
		if usuario == files.Files[i].User && fichero == files.Files[i].File {
			return true
		}
	}
	return false
}

func registrarBloqueFicheroUsuario(usuario string, fichero string, bloque BlockPosition) {
	jsonBytes := leerJSON("./databases/files.json")
	var files Files
	json.Unmarshal(jsonBytes, &files)

	existe := false

	var order []BlockPosition
	var count int
	for i := 0; i < len(files.Files) && !existe; i++ {
		if usuario == files.Files[i].User && fichero == files.Files[i].File {
			existe = true
			order = files.Files[i].Order
			count = i
		}
	}

	if !existe { // Primer bloque de un nuevo archivo
		order = append(order, bloque)
		files.Files = append(files.Files, File{User: usuario, File: fichero, Order: order})

		filesJSON, _ := json.Marshal(files)
		err := ioutil.WriteFile("./databases/files.json", filesJSON, 0644)
		// now Marshal it
		check(err)
	} else {
		// Si ya existe un usuario-file, comprueba que el bloque-posicion existe, si no existe, lo crea, sino lo sobrescribe
		asignado := false
		var newOrder []BlockPosition
		for i := 0; i < len(order) && !asignado; i++ {
			if bloque.Position == order[i].Position {
				order[i] = bloque
				asignado = true
			}
			newOrder = append(newOrder, order[i])
		}
		if asignado {
			order = newOrder
		} else {
			order = append(order, bloque)
		}
		files.Files[count].Order = order
		filesJSON, _ := json.Marshal(files)
		err := ioutil.WriteFile("./databases/files.json", filesJSON, 0644)
		// now Marshal it
		check(err)
	}
}

func handlerShowUserFiles(w http.ResponseWriter, r *http.Request) {
	validarToken(r.Header.Get("Authorization"), r.Header.Get("Username"))
	jsonBytes := leerJSON("./databases/files.json")
	var files Files
	json.Unmarshal(jsonBytes, &files)

	u, err := url.Parse(r.URL.String())
	check(err)
	result := strings.Split(u.Path, "/")
	username := result[len(result)-1]
	var filesUser []string
	for i := 0; i < len(files.Files); i++ {
		if username == files.Files[i].User {
			filesUser = append(filesUser, encodeURLB64(files.Files[i].File))
		}
	}
	slc, _ := json.Marshal(filesUser)
	w.Write(slc)
}

func handlerSendFile(w http.ResponseWriter, r *http.Request) {

	validarToken(r.Header.Get("Authorization"), r.Header.Get("Username"))

	//fmt.Println("Paso por handlerFiles")

	u, err := url.Parse(r.URL.String())
	check(err)
	result := strings.Split(u.Path, "/")
	//fmt.Println(result)
	userSolicitante := result[len(result)-3]
	archivoSolicitado := decodeURLB64(result[len(result)-1])

	jsonBytes := leerJSON("./databases/files.json")
	var files Files
	json.Unmarshal(jsonBytes, &files)

	existe := false
	var bloquesDeArchivo []BlockPosition
	for i := 0; i < len(files.Files); i++ {
		if files.Files[i].User == userSolicitante && files.Files[i].File == archivoSolicitado && !existe {
			existe = true
			bloquesDeArchivo = files.Files[i].Order
		}
	}

	if !existe {
		response(w, false, "El archivo No Existe")
	} else {
		formatoArchivo := strings.Split(archivoSolicitado, ".")
		var streamBytesTotal []byte
		for i := 0; i < len(bloquesDeArchivo); i++ {
			ruta := "./archivos/" + bloquesDeArchivo[i].Block
			streamBytes, err := ioutil.ReadFile(ruta)
			streamBytes = decryptAESCFB(streamBytes, obtenerClaveCifrado(ruta))
			check(err)
			streamBytesTotal = append(streamBytesTotal[:], streamBytes[:]...)
		}
		kind := mime.TypeByExtension("." + formatoArchivo[len(formatoArchivo)-1])

		b := bytes.NewBuffer(streamBytesTotal)
		// stream straight to client(browser)

		w.Header().Set("Content-type", kind)

		if _, err := b.WriteTo(w); err != nil { // <----- here!
			fmt.Fprintf(w, "%s", err)
		}
	}
}

func getNombreUltimoFichero() string {
	jsonBytes := leerJSON("./databases/blocks.json")
	var blocks Blocks
	json.Unmarshal(jsonBytes, &blocks)

	final := "-1"
	for i := 0; i < len(blocks.Blocks); i++ {
		final = blocks.Blocks[i].Block
	}
	return final
}

func cifrarCarpeta(ruta string) {
	//recorrer todos los ficheros y cifrarlos con una contraseña maestra
	err := filepath.Walk(ruta, visitEncrypt) //esta funcion recorre todos los directorios y ficheros recursivamente
	check(err)
}

func visitEncrypt(path string, f os.FileInfo, err error) error { //funcion para cifrarFicherosUsuarios
	if f != nil && f.IsDir() == false { //para coger solo los ficheros y no las carpetas
		clavemaestra := "{<J*l-&lG.f@GiNtOnIcO@B}%1ckFHb_" //32 bytes para que sea AES256
		//clavemaestra := obtenerClaveCifrado(path)
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

func descifrarCarpeta(ruta string) {
	//recorrer todos los ficheros y cifrarlos con una contraseña maestra
	err := filepath.Walk(ruta, visitDecrypt) //esta funcion recorre todos los directorios y ficheros recursivamente
	check(err)
}

func visitDecrypt(path string, f os.FileInfo, err error) error {
	//funcion para descifrarFicherosUsuarios
	if f != nil && f.IsDir() == false { //para coger solo los ficheros y no las carpetas
		clavemaestra := "{<J*l-&lG.f@GiNtOnIcO@B}%1ckFHb_" //32 bytes para que sea AES256
		//clavemaestra := obtenerClaveCifrado(path)
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

func obtenerClaveCifrado(path string) string {
	//fmt.Println("Path: " + path)
	nombreBloque := strings.Split(path, "/")
	bloque := nombreBloque[len(nombreBloque)-1]
	/* Obtener quien cifro el bloque*/
	jsonBytes := leerJSON("./databases/blocks.json")
	var blocks Blocks
	json.Unmarshal(jsonBytes, &blocks)

	var userPropietarioClave string
	var encontrado = false
	for i := 0; i < len(blocks.Blocks) && !encontrado; i++ {
		if bloque == blocks.Blocks[i].Block {
			userPropietarioClave = blocks.Blocks[i].User
			encontrado = true
		}
	}

	/* FIN Obtener quien cifro el bloque*/
	/* Obtener clave de cifrado el bloque*/
	jsonBytes = leerJSON("./databases/users.json")
	var users Users
	json.Unmarshal(jsonBytes, &users)

	var claveCifrado string
	encontrado = false
	for i := 0; i < len(users.Users) && !encontrado; i++ {
		if userPropietarioClave == users.Users[i].User {
			claveCifrado = users.Users[i].Cifrado
			encontrado = true
		}
	}
	/* FIN Obtener clave de cifrado el bloque*/
	return claveCifrado
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano()) //para que el aleatorio funcione bien
	createDirIfNotExist("./archivos/")
	createDirIfNotExist("./certificados/")
	createDirIfNotExist("./databases/")
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	// Comprueba los certificados, si no existen se generan nuevos
	err := httpscerts.Check("./certificados/cert.pem", "./certificados/key.pem")

	if err != nil {
		err = httpscerts.Generate("./certificados/cert.pem", "./certificados/key.pem", ":8081")
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

	muxa.HandleFunc("/validarToken", handlerValidarToken)

	srv := &http.Server{Addr: ":8081", Handler: muxa}

	go func() {
		log.Println("Poniendo en marcha servidor HTTPS, escuchando puerto 8081")
		if err := srv.ListenAndServeTLS("./certificados/cert.pem", "./certificados/key.pem"); err != nil {
			log.Printf("Error al poner en funcionamiento el servidor TLS: %s\n", err)
		}
	}()
	go func() {
		log.Println("Poniendo en marcha redireccionamiento HTTP->HTTPS, escuchando puerto 8080")
		if err := http.ListenAndServe(":8080", http.HandlerFunc(redirectToHTTPS)); err != nil {
			log.Printf("Error al redireccionar http a https: %s\n", err)
		}
	}()

	log.Println("Descifrando bases de datos...")
	descifrarCarpeta("./databases")

	<-stopChan // espera señal SIGINT
	log.Println("Apagando servidor ...")
	// apagar servidor de forma segura
	ctx, fnc := context.WithTimeout(context.Background(), 5*time.Second)
	fnc()
	srv.Shutdown(ctx)

	log.Println("Cifrando bases de datos...")
	cifrarCarpeta("./databases")

	log.Println("Servidor detenido correctamente")
}
