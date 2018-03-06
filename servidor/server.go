package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/kabukky/httpscerts"
)

// respuesta del servidor
type resp struct {
	Ok  bool   // true -> correcto, false -> error
	Msg string // mensaje adicional
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

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there!")
}

func handlerLogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()                                // es necesario parsear el formulario
	w.Header().Set("Content-Type", "text/plain") // cabecera estándar

	if validarLogin(r.Form.Get("login"), r.Form.Get("password")) {
		response(w, true, "Logeado")
	} else {
		response(w, false, "Error")
	}

}

func validarLogin(login string, password string) bool {

	log.Println("Usuario: " + login)
	log.Println("Password: " + password)
	return true

}

func handlerRegister(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there!")
}

func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	// Redirect the incoming HTTP request. Note that "127.0.0.1:8081" will only work if you are accessing the server from your local machine.
	http.Redirect(w, r, "https://127.0.0.1:8081"+r.RequestURI, http.StatusMovedPermanently)
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

	srv := &http.Server{Addr: ":8081", Handler: mux}

	go func() {
		if err := srv.ListenAndServeTLS("cert.pem", "key.pem"); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()
	// Inicia el servidor HTTP y redirige todas las peticiones a HTTPS
	go func() {
		if err := http.ListenAndServe(":8080", http.HandlerFunc(redirectToHTTPS)); err != nil {
			log.Printf("listen: %s\n", err)
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
