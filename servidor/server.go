package main

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/textproto"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/kabukky/httpscerts"
	"golang.org/x/crypto/scrypt"
)

//Estrucutra de ficheros
type FileHeader struct {
	Username string
	Filename string
	Header   textproto.MIMEHeader
	// contains filtered or unexported fields
}

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

// función para comprobar errores (ahorra escritura)
func chk(e error) {
	if e != nil {
		panic(e)
	}
}

// función para escribir una respuesta del servidor
func response(w io.Writer, ok bool, msg string) {
	r := resp{Ok: ok, Msg: msg}    // formateamos respuesta
	rJSON, err := json.Marshal(&r) // codificamos en JSON
	chk(err)                       // comprobamos error
	w.Write(rJSON)                 // escribimos el JSON resultante
}

// Int Aleatorio
func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

// String Aleatorio
func randomString(l int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi theresfsdgsfg!")
}

func handlerLogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()                                // es necesario parsear el formulario
	w.Header().Set("Content-Type", "text/plain") // cabecera estándar

	if validarLogin(r.Form.Get("login"), r.Form.Get("password")) {
		response(w, true, "Logeado")
	} else {
		response(w, false, "Error al loguear")
	}
}

func validarLogin(login string, password string) bool {
	fmt.Println("login: " + login + " pass: " + password)
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

//Comprueba que los directorios no existen
func CreateDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func handlerUpload(w http.ResponseWriter, r *http.Request) {

	fmt.Println("method:", r.Method)
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		fmt.Fprintf(w, "%v", handler.Header)
		// Split on /.

		result := strings.Split(handler.Filename, "/")
		fmt.Println(r.FormValue("Username"))
		CreateDirIfNotExist("./archivos/")
		CreateDirIfNotExist("./archivos/" + r.FormValue("Username"))
		f, err := os.OpenFile("./archivos/"+r.FormValue("Username")+"/"+result[len(result)-1], os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
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
	/*
		for _, v := range users.Users {
			fmt.Println(v)
		}
		fmt.Println()
	*/
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

// Devuelve el string de la cadena encriptada
func encriptarScrypt(cadena string, seed string) string {
	salt := []byte(seed)

	dk, err := scrypt.Key([]byte(cadena), salt, 1<<15, 10, 1, 32)
	if err != nil {
		log.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(dk)
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

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(handler))
	mux.Handle("/login", http.HandlerFunc(handlerLogin))
	mux.Handle("/register", http.HandlerFunc(handlerRegister))
	mux.Handle("/upload", http.HandlerFunc(handlerUpload))

	srv := &http.Server{Addr: ":8081", Handler: mux}

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
