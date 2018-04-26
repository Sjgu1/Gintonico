package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	Size     string `json:"size"`
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

func handlerLogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()                                // es necesario parsear el formulario
	w.Header().Set("Content-Type", "text/plain") // cabecera estándar

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.Bytes()

	type LoginJSON struct {
		Login    []string `json:"login"`
		Password []string `json:"password"`
	}
	var user LoginJSON
	err := json.Unmarshal(body, &user)
	check(err)
	if err == nil && validarLogin(user.Login[0], user.Password[0]) {
		token := createJWT(user.Login[0])
		w.Header().Add("Token", token)
		guardarToken(token, user.Login[0])
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

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.Bytes()

	type RegisterJSON struct {
		Register []string `json:"register"`
		Password []string `json:"password"`
		Confirm  []string `json:"confirm"`
	}
	var user RegisterJSON
	err := json.Unmarshal(body, &user)
	check(err)

	if err == nil && validarRegister(user.Register[0], user.Password[0], user.Confirm[0]) {
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
	r.ParseForm()                                // es necesario parsear el formulario
	w.Header().Set("Content-Type", "text/plain") // cabecera estándar

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.Bytes()

	type BodyJSON struct {
		Cont     []string `json:"cont"`
		Hash     []string `json:"hash"`
		Size     []string `json:"size"`
		User     []string `json:"user"`
		Filename []string `json:"filename"`
	}
	var bodyJSON BodyJSON
	err := json.Unmarshal(body, &bodyJSON)
	check(err)

	if err == nil {
		contador, _ := strconv.Atoi(bodyJSON.Cont[0])  // numero del orden de la parte del fichero
		hash := bodyJSON.Hash[0]                       // hash de la parte del fichero
		size, _ := strconv.Atoi(bodyJSON.Size[0])      // tamaño de la parte del fichero
		user := bodyJSON.User[0]                       // usuario que sube el fichero
		filename := decodeURLB64(bodyJSON.Filename[0]) // nombre del fichero original

		comprobar := comprobarHash(contador, hash, size, user, filename)
		response(w, comprobar, "Hash comprobado")
	} else {
		response(w, false, "Error al comprobar")
	}
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
		position.Size = strconv.Itoa(tam)
		registrarBloqueFicheroUsuario(user, filename, position)
		return true
	}
	return false
}

func handlerUpload(w http.ResponseWriter, r *http.Request) {
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
	position.Size = r.FormValue("Size")

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
	u, err := url.Parse(r.URL.String())
	check(err)
	result := strings.Split(u.Path, "/")
	username := result[len(result)-1]

	type FilesJSON struct {
		Filename []string `json:"filename"`
		Size     []string `json:"size"`
	}

	jsonBytes := leerJSON("./databases/files.json")
	var files Files
	json.Unmarshal(jsonBytes, &files)

	var filesUser []string
	var tamFiles []string
	for i := 0; i < len(files.Files); i++ {
		if username == files.Files[i].User {
			filesUser = append(filesUser, encodeURLB64(files.Files[i].File))
			tamanyo := 0
			for j := range files.Files[i].Order {
				x, _ := strconv.Atoi(files.Files[i].Order[j].Size)
				tamanyo += x
			}
			total := strconv.Itoa(tamanyo)
			tamFiles = append(tamFiles, total)
		}
	}

	var filesJSON = FilesJSON{Filename: filesUser, Size: tamFiles}

	if len(filesUser) > 0 {
		slc, _ := json.Marshal(filesJSON)
		w.Write(slc)
	} else {
		response(w, false, "No tienes ficheros subidos")
	}
}

func handlerDeleteFile(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	check(err)
	result := strings.Split(u.Path, "/")
	userSolicitante := result[len(result)-3]
	archivoSolicitado := decodeURLB64(result[len(result)-1])

	jsonBytes := leerJSON("./databases/files.json")
	var files Files
	json.Unmarshal(jsonBytes, &files)

	existe := false
	var bloquesDeArchivo []BlockPosition
	for i := 0; i < len(files.Files) && !existe; i++ {
		if files.Files[i].User == userSolicitante && files.Files[i].File == archivoSolicitado {
			existe = true
			bloquesDeArchivo = files.Files[i].Order
		}
	}

	if !existe {
		response(w, false, "El usuario no dispone de este archivo")
	} else {
		jsonBytes2 := leerJSON("./databases/blocks.json")
		var blocks Blocks
		json.Unmarshal(jsonBytes2, &blocks)

		for i := 0; i < len(bloquesDeArchivo); i++ {
			var bloqueCambiado = false
			for j := 0; j < len(files.Files); j++ {
				for k := 0; k < len(files.Files[j].Order) && !bloqueCambiado; k++ {
					if bloquesDeArchivo[i].Block == files.Files[j].Order[k].Block {
						otroUsuarioBloque, otroUsuarioTiene := checkUsersBlocks(userSolicitante, bloquesDeArchivo[i].Block)
						if !otroUsuarioTiene {
							deleteFile("./archivos/" + bloquesDeArchivo[i].Block)
							eliminarBloque(bloquesDeArchivo[i].Block, &blocks)
						} else {
							claveOriginal, nuevaClave, err := obtenerClavesUsuarios(bloquesDeArchivo[i].Block, otroUsuarioBloque)
							check(err)

							asignarNuevaClave("./archivos/"+bloquesDeArchivo[i].Block, claveOriginal, nuevaClave)

							blocks.Blocks[getPosicionBloque(bloquesDeArchivo[i].Block)].User = otroUsuarioBloque
						}
						bloqueCambiado = true
					} /*else { //no hace falta responder que no existe el bloque
						response(w, false, "No existe el bloque")
					}*/
				}
			}
		}
		blocksJSON, _ := json.Marshal(blocks)
		err := ioutil.WriteFile("./databases/blocks.json", blocksJSON, 0644)
		check(err)
		eliminarArchivoUsuario(userSolicitante, archivoSolicitado)
		response(w, true, "Borrado")
	}
}

func asignarNuevaClave(path string, claveOriginal string, claveNueva string) {

	file, err := ioutil.ReadFile(path)
	check(err)

	if len(file) > 0 {
		decryptedFile := decryptAESCFB(file, claveOriginal)

		deleteFile(path)

		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
		check(err)
		defer f.Close()

		encryptedFile := encryptAESCFB(decryptedFile, claveNueva)

		io.Copy(f, bytes.NewReader(encryptedFile))
	}

}

func getPosicionBloque(bloque string) int {
	jsonBytes2 := leerJSON("./databases/blocks.json")
	var blocks Blocks
	json.Unmarshal(jsonBytes2, &blocks)
	for i := 0; i < len(blocks.Blocks); i++ {
		if blocks.Blocks[i].Block == bloque {
			return i
		}
	}
	return 0
}

func eliminarArchivoUsuario(usuario string, archivo string) {
	jsonBytes := leerJSON("./databases/files.json")
	var files Files
	json.Unmarshal(jsonBytes, &files)

	existe := false
	for i := 0; i < len(files.Files) && !existe; i++ {
		if files.Files[i].File == archivo && files.Files[i].User == usuario {
			files.Files = append(files.Files[:i], files.Files[i+1:]...)
			filesJSON, _ := json.Marshal(files)
			err := ioutil.WriteFile("./databases/files.json", filesJSON, 0644)
			check(err)
			existe = true
		}
	}
}

func eliminarBloque(bloque string, blocks *Blocks) {
	existe := false
	for i := 0; i < len(blocks.Blocks) && !existe; i++ {
		if blocks.Blocks[i].Block == bloque {
			blocks.Blocks = append(blocks.Blocks[:i], blocks.Blocks[i+1:]...)
			existe = true
		}
	}
}

func obtenerClavesUsuarios(bloque string, nuevoUsuario string) (string, string, error) {
	claveUsuarioOriginal := obtenerClaveCifrado("./archivos/" + bloque)
	jsonBytes := leerJSON("./databases/users.json")
	var users Users
	json.Unmarshal(jsonBytes, &users)
	var claveNuevoUsuario string
	var encontrado = false
	for i := 0; i < len(users.Users) && !encontrado; i++ {
		if nuevoUsuario == users.Users[i].User {
			claveNuevoUsuario = users.Users[i].Cifrado
			encontrado = true
		}
	}

	if encontrado {
		return claveUsuarioOriginal, claveNuevoUsuario, nil
	}
	err := errors.New("Error al obtener las claves")
	return "", "", err

}

//comprueba si alquien a parte de ti tiene el bloque
func checkUsersBlocks(username string, block string) (string, bool) {
	jsonBytes := leerJSON("./databases/files.json")
	var files Files
	json.Unmarshal(jsonBytes, &files)
	for i := 0; i < len(files.Files); i++ {
		for j := 0; j < len(files.Files[i].Order); j++ {
			if files.Files[i].Order[j].Block == block && files.Files[i].User != username {
				return files.Files[i].User, true
			}
		}

	}
	return "false", false
}

func handlerSendFile(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	check(err)
	result := strings.Split(u.Path, "/")

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

func middlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenValido := validarToken(r.Header.Get("Authorization"), r.Header.Get("Username"))
		if tokenValido {
			next.ServeHTTP(w, r)
		} else {
			response(w, false, "Error de autenticación")
		}
	})
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
	muxa.Handle("/checkhash", middlewareAuth(http.HandlerFunc(handlerHash)))
	muxa.Handle("/upload", middlewareAuth(http.HandlerFunc(handlerUpload)))
	muxa.Handle("/user/{username}", middlewareAuth(http.HandlerFunc(handlerShowUserFiles)))
	muxa.Handle("/user/{username}/file/{filename}", middlewareAuth(http.HandlerFunc(handlerSendFile))).Methods("GET")
	muxa.Handle("/user/{username}/file/{filename}", middlewareAuth(http.HandlerFunc(handlerDeleteFile))).Methods("DELETE")

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
