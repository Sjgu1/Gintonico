package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
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
}

// función para escribir una respuesta del servidor
func response(w io.Writer, ok bool, msg string) {
	r := resp{Ok: ok, Msg: msg}    // formateamos respuesta
	rJSON, err := json.Marshal(&r) // codificamos en JSON
	check(err)                     // comprobamos error
	w.Write(rJSON)                 // escribimos el JSON resultante
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Paso por el handler")
}

func handlerUser(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Paso por handlerUser")

	u, err := url.Parse(r.URL.String())
	if err != nil {
		log.Fatal(err)
	}
	result := strings.Split(u.Path, "/")
	createDirIfNotExist("./archivos/" + result[len(result)-1])
	files, err := ioutil.ReadDir("./archivos/" + result[len(result)-1] + "/")
	if err != nil {
		log.Fatal(err)
	}

	s := make([]string, len(files))
	for i, f := range files {
		s[i] = encodeURLB64(f.Name())
	}

	slc, _ := json.Marshal(s)
	w.Write(slc)
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
			if err != nil {
				fmt.Println(err)
			}
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

func comprobarExisteUsuario(usuario string) bool {
	// Abre el archivo json
	jsonFile, err := os.Open("users.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
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
		if usuario == users.Users[i].User {
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

func handlerUpload(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Paso por handlerUpload")

	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Fprintf(w, "%v", handler.Header)
	// Split on /.
	fichero := decodeURLB64(handler.Filename) + ".part" + r.FormValue("Parte")
	fmt.Println(fichero)
	createDirIfNotExist("./archivos/")
	createDirIfNotExist("./archivos/" + r.FormValue("Username"))
	f, err := os.OpenFile("./archivos/"+r.FormValue("Username")+"/"+fichero, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, file)
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
	//buscar en la base de datos, si ya existe devolver true, si no existe false
	//además si ya existe, hay que asociar ese hash existente con el usuario y to eso
	return false
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
			if err != nil {
				fmt.Println(err)
				return false
			}
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

	salt := randomString(30)

	users.Users = append(users.Users, User{User: register, Password: encriptarScrypt(password, salt), Salt: salt})

	usersJSON, _ := json.Marshal(users)
	err = ioutil.WriteFile("users.json", usersJSON, 0644)

	// IMPRIMIR USUARIOS
	// now Marshal it
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	// Redirect the incoming HTTP request. Note that "127.0.0.1:8081" will only work if you are accessing the server from your local machine.
	http.Redirect(w, r, "https://127.0.0.1:8081"+r.RequestURI, http.StatusMovedPermanently)
}

func handlerFiles(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Paso por handlerFiles")

	u, err := url.Parse(r.URL.String())
	if err != nil {
		log.Fatal(err)
	}
	result := strings.Split(u.Path, "/")
	fmt.Println(result)
	if _, err := os.Stat("./archivos/" + result[len(result)-3] + "/" + decodeURLB64(result[len(result)-1])); err == nil {

		// grab the generated receipt.pdf file and stream it to browser
		streamBytes, err := ioutil.ReadFile("./archivos/" + result[len(result)-3] + "/" + decodeURLB64(result[len(result)-1]))

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

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

func main() {
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
	muxa.HandleFunc("/user/{username}", handlerUser)
	muxa.HandleFunc("/user/{username}/file/{filename}", handlerFiles)

	srv := &http.Server{Addr: ":8081", Handler: muxa}

	go func() {
		log.Println("Poniendo en marcha servidor HTTPS, escuchando puerto 8081")
		if err := srv.ListenAndServeTLS("cert.pem", "key.pem"); err != nil {
			log.Printf("Error al poner en funcionamiento el servidor TLS: %s\n", err)
		}
	}()
	// Inicia el servidor HTTP y redirige todas las peticiones a HTTPS
	go func() {
		log.Println("Poniendo en marcha redireccionamiento HTTP->HTTPS, escuchando puerto 8080")
		if err := http.ListenAndServe(":8080", http.HandlerFunc(redirectToHTTPS)); err != nil {
			log.Printf("Error al redireccionar http a https: %s\n", err)
		}
	}()

	<-stopChan // espera señal SIGINT
	log.Println("Apagando servidor ...")
	// apagar servidor de forma segura
	ctx, fnc := context.WithTimeout(context.Background(), 5*time.Second)
	fnc()
	srv.Shutdown(ctx)

	log.Println("Servidor detenido correctamente")
}
