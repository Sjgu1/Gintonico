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

const rutaArchivos = "./archivos/"
const rutaCertificados = "./certificados"
const rutaDatabases = "./databases"
const rutaUsersBD = rutaDatabases + "/users.json"
const rutaBlocksBD = rutaDatabases + "/blocks.json"
const rutaFilesBD = rutaDatabases + "/files.json"

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

	jsonBytes := leerJSON(rutaUsersBD)
	var users Users
	err = json.Unmarshal(jsonBytes, &users)
	check(err)

	if err == nil && validarLogin(user.Login[0], user.Password[0], &users) {
		token := createJWT(user.Login[0])
		w.Header().Add("Token", token)
		guardarToken(token, user.Login[0], &users)
		response(w, true, token)
	} else {
		response(w, false, "Error al loguear")
	}
}

func validarLogin(login string, password string, users *Users) bool {
	for i := 0; i < len(users.Users); i++ {
		if login == users.Users[i].User && encriptarScrypt(password, users.Users[i].Salt) == users.Users[i].Password {
			return true
		}
	}
	return false
}

func guardarToken(token string, user string, users *Users) {
	var encontrado = false
	for i := 0; i < len(users.Users) && !encontrado; i++ {
		if users.Users[i].User == user {
			users.Users[i].Token = token
			guardarJSON(rutaUsersBD, &users)
			encontrado = true
		}
	}
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
	if password != confirm || register == "" || password == "" {
		return false
	}

	if comprobarExisteUsuario(register) {
		return false
	}

	jsonBytes := leerJSON(rutaUsersBD)
	var users Users
	json.Unmarshal(jsonBytes, &users)

	salt := randomString(32)
	cifrado := randomString(32)

	users.Users = append(users.Users, User{User: register, Password: encriptarScrypt(password, salt), Salt: salt, Cifrado: cifrado})
	guardarJSON(rutaUsersBD, &users)

	return true
}

func comprobarExisteUsuario(usuario string) bool {
	jsonBytes := leerJSON(rutaUsersBD)
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
	parte := strconv.Itoa(cont)

	var position BlockPosition
	existeBloque, nombreBloque := existeBloqueHash(hash)

	if existeBloque {
		position.Block = nombreBloque
		position.Position = parte
		position.Size = strconv.Itoa(tam)
		registrarFileUsuario(user, filename, position)
		return true
	}
	return false
}

func existeBloqueHash(hash string) (bool, string) {
	jsonBytes := leerJSON(rutaBlocksBD)
	var blocks Blocks
	json.Unmarshal(jsonBytes, &blocks)

	for i := 0; i < len(blocks.Blocks); i++ {
		if hash == blocks.Blocks[i].Hash {
			return true, blocks.Blocks[i].Block
		}
	}
	return false, "nil"
}

func handlerUpload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("uploadfile")
	check(err)
	defer file.Close()

	var position BlockPosition

	last := getNombreUltimoFichero()
	value, err := strconv.Atoi(last)
	value++
	path := strconv.Itoa(value)

	f, err := os.OpenFile(rutaArchivos+path, os.O_WRONLY|os.O_CREATE, 0666)
	check(err)
	defer f.Close()
	io.Copy(f, file)
	position.Block = path
	position.Position = r.FormValue("Parte")
	position.Size = r.FormValue("Size")

	var block Block
	block.User = r.FormValue("Username")
	block.Hash = r.FormValue("Hash")
	block.Block = path
	registrarBloque(&block)                                                                 //lo guarda en Blocks
	registrarFileUsuario(r.FormValue("Username"), decodeURLB64(handler.Filename), position) //lo guarda en Files
	cifrarFichero(rutaArchivos+path, obtenerClaveCifrado(rutaArchivos+path))
}

func registrarBloque(bloque *Block) {
	jsonBytes := leerJSON(rutaBlocksBD)
	var blocks Blocks
	json.Unmarshal(jsonBytes, &blocks)

	blocks.Blocks = append(blocks.Blocks, Block{Block: bloque.Block, Hash: bloque.Hash, User: bloque.User})
	guardarJSON(rutaBlocksBD, &blocks)
}

func registrarFileUsuario(usuario string, fichero string, bloque BlockPosition) {
	jsonBytes := leerJSON(rutaFilesBD)
	var files Files
	json.Unmarshal(jsonBytes, &files)

	existe, bloquesDeArchivo, count := existeFicheroUsuario(usuario, fichero, &files)

	if !existe { // Primer bloque de un nuevo archivo
		bloquesDeArchivo = append(bloquesDeArchivo, bloque)
		files.Files = append(files.Files, File{User: usuario, File: fichero, Order: bloquesDeArchivo})
		guardarJSON(rutaFilesBD, &files)
	} else {
		// Si ya existe un usuario-file, comprueba que el bloque-posicion existe, si no existe, lo crea, sino lo sobrescribe
		asignado := false
		var nuevosBloquesDeArchivo []BlockPosition
		for i := 0; i < len(bloquesDeArchivo); i++ {
			if bloque.Position == bloquesDeArchivo[i].Position {
				bloquesDeArchivo[i] = bloque
				asignado = true
			}
			nuevosBloquesDeArchivo = append(nuevosBloquesDeArchivo, bloquesDeArchivo[i])
		}
		if asignado {
			bloquesDeArchivo = nuevosBloquesDeArchivo
		} else {
			bloquesDeArchivo = append(bloquesDeArchivo, bloque)
		}
		files.Files[count].Order = bloquesDeArchivo
		guardarJSON(rutaFilesBD, &files)
	}
}

func existeFicheroUsuario(usuario string, fichero string, files *Files) (bool, []BlockPosition, int) {
	for i := 0; i < len(files.Files); i++ {
		if usuario == files.Files[i].User && fichero == files.Files[i].File {
			return true, files.Files[i].Order, i
		}
	}
	return false, nil, 0
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

	jsonBytes := leerJSON(rutaFilesBD)
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

	jsonBytes := leerJSON(rutaFilesBD)
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
		jsonBytes2 := leerJSON(rutaBlocksBD)
		var blocks Blocks
		json.Unmarshal(jsonBytes2, &blocks)

		for i := 0; i < len(bloquesDeArchivo); i++ {
			var bloqueCambiado = false
			for j := 0; j < len(files.Files); j++ {
				for k := 0; k < len(files.Files[j].Order) && !bloqueCambiado; k++ {
					if bloquesDeArchivo[i].Block == files.Files[j].Order[k].Block {
						otroUsuarioBloque, otroUsuarioTiene := checkUsersBlocks(userSolicitante, bloquesDeArchivo[i].Block, &files)
						if !otroUsuarioTiene {
							deleteFile(rutaArchivos + bloquesDeArchivo[i].Block)
							eliminarBloque(bloquesDeArchivo[i].Block, &blocks)
						} else {
							claveOriginal, nuevaClave, err := obtenerClavesUsuarios(bloquesDeArchivo[i].Block, otroUsuarioBloque)
							check(err)

							asignarNuevaClave(rutaArchivos+bloquesDeArchivo[i].Block, claveOriginal, nuevaClave)

							blocks.Blocks[getPosicionBloque(bloquesDeArchivo[i].Block, &blocks)].User = otroUsuarioBloque
						}
						bloqueCambiado = true
					}
				}
			}
		}
		guardarJSON(rutaBlocksBD, &blocks)
		eliminarArchivoUsuario(userSolicitante, archivoSolicitado, &files)
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

func getPosicionBloque(bloque string, blocks *Blocks) int {
	for i := 0; i < len(blocks.Blocks); i++ {
		if blocks.Blocks[i].Block == bloque {
			return i
		}
	}
	return 0
}

func eliminarArchivoUsuario(usuario string, archivo string, files *Files) bool {
	borrado := false
	for i := 0; i < len(files.Files) && !borrado; i++ {
		if files.Files[i].File == archivo && files.Files[i].User == usuario {
			files.Files = append(files.Files[:i], files.Files[i+1:]...)
			guardarJSON(rutaFilesBD, &files)
			borrado = true
		}
	}
	return borrado
}

func eliminarBloque(bloque string, blocks *Blocks) bool {
	borrado := false
	for i := 0; i < len(blocks.Blocks) && !borrado; i++ {
		if blocks.Blocks[i].Block == bloque {
			blocks.Blocks = append(blocks.Blocks[:i], blocks.Blocks[i+1:]...)
			borrado = true
		}
	}
	return borrado
}

func obtenerClavesUsuarios(bloque string, nuevoUsuario string) (string, string, error) {
	claveUsuarioOriginal := obtenerClaveCifrado(rutaArchivos + bloque)
	jsonBytes := leerJSON(rutaUsersBD)
	var users Users
	json.Unmarshal(jsonBytes, &users)

	for i := 0; i < len(users.Users); i++ {
		if nuevoUsuario == users.Users[i].User {
			claveNuevoUsuario := users.Users[i].Cifrado
			return claveUsuarioOriginal, claveNuevoUsuario, nil
		}
	}

	err := errors.New("Error al obtener las claves")
	return "", "", err
}

func checkUsersBlocks(username string, block string, files *Files) (string, bool) {
	//comprueba si alquien a parte de ti tiene el bloque
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

	jsonBytes := leerJSON(rutaFilesBD)
	var files Files
	json.Unmarshal(jsonBytes, &files)

	existe, bloquesDeArchivo, _ := existeFicheroUsuario(userSolicitante, archivoSolicitado, &files)

	if !existe {
		response(w, false, "El archivo No Existe")
	} else {
		formatoArchivo := strings.Split(archivoSolicitado, ".")
		var streamBytesTotal []byte
		for i := 0; i < len(bloquesDeArchivo); i++ {
			ruta := rutaArchivos + bloquesDeArchivo[i].Block
			streamBytes, err := ioutil.ReadFile(ruta)
			streamBytes = decryptAESCFB(streamBytes, obtenerClaveCifrado(ruta))
			check(err)
			streamBytesTotal = append(streamBytesTotal[:], streamBytes[:]...)
		}
		kind := mime.TypeByExtension("." + formatoArchivo[len(formatoArchivo)-1])

		b := bytes.NewBuffer(streamBytesTotal)
		w.Header().Set("Content-type", kind)

		if _, err := b.WriteTo(w); err != nil {
			fmt.Fprintf(w, "%s", err)
		}
	}
}

func getNombreUltimoFichero() string {
	jsonBytes := leerJSON(rutaBlocksBD)
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
	jsonBytes := leerJSON(rutaBlocksBD)
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
	jsonBytes = leerJSON(rutaUsersBD)
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
	createDirIfNotExist(rutaArchivos)
	createDirIfNotExist(rutaCertificados)
	createDirIfNotExist(rutaDatabases)
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	// Comprueba los certificados, si no existen se generan nuevos
	err := httpscerts.Check(rutaCertificados+"/cert.pem", rutaCertificados+"/key.pem")
	if err != nil {
		err = httpscerts.Generate(rutaCertificados+"/cert.pem", rutaCertificados+"/key.pem", ":8081")
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
		if err := srv.ListenAndServeTLS(rutaCertificados+"/cert.pem", rutaCertificados+"/key.pem"); err != nil {
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
	descifrarCarpeta(rutaDatabases)

	<-stopChan // espera señal SIGINT
	log.Println("Apagando servidor ...")
	// apagar servidor de forma segura
	ctx, fnc := context.WithTimeout(context.Background(), 5*time.Second)
	fnc()
	srv.Shutdown(ctx)

	log.Println("Cifrando bases de datos...")
	cifrarCarpeta(rutaDatabases)

	log.Println("Servidor detenido correctamente")
}
