package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/dtylman/gowd"
	"github.com/dtylman/gowd/bootstrap"
	"golang.org/x/crypto/scrypt"
)

var body *gowd.Element
var mostrar = "login"
var login = ""

type resp struct {
	Ok  bool   `json:"ok"`  // true -> correcto, false -> error
	Msg string `json:"msg"` // mensaje adicional
}

// función para comprobar errores (ahorra escritura)
func check(e error) {
	if e != nil {
		fmt.Println(e.Error())
	}
}

func sendServerPetition(data map[string][]string, route string) *http.Response {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	r, err := client.PostForm("https://localhost:8081"+route, data) // enviamos por POST
	check(err)
	return r
}

func main() {
	body = bootstrap.NewContainer(false)
	//body.SetAttribute("style", "background-color:#FF654E")

	logo := `<div style="margin:0 auto;width:40%;"><img src="img/logo_alargado.png" style="width:100%;margin:0 auto"/></div><br/><br/>`

	switch mostrar {
	case "login":
		body.AddHTML(logo, nil)
		body.AddHTML(vistaLogin(), nil)
		body.Find("login-submit").OnEvent(gowd.OnClick, sendLogin)
		body.Find("register-form-link").OnEvent(gowd.OnClick, actualizarVista)
		cambiarVista("register")
		break
	case "register":
		body.AddHTML(logo, nil)
		body.AddHTML(vistaRegister(), nil)
		body.Find("register-submit").OnEvent(gowd.OnClick, sendRegister)
		body.Find("login-form-link").OnEvent(gowd.OnClick, actualizarVista)
		cambiarVista("login")
		break
	case "principal":
		body.AddHTML(vistaPrincipal(), nil)
		body.Find("buttonEnviar").OnEvent(gowd.OnClick, seleccionarFichero)
		body.Find("logout-link").OnEvent(gowd.OnClick, actualizarVista)
		body.Find("buttonPedir").OnEvent(gowd.OnClick, pedirFichero)
		cambiarVista("login")
		break
	}
	//start the ui loop
	err := gowd.Run(body)
	check(err)
}

func StreamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}

func StreamToString(stream io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.String()
}

func peticionNombreFicheros() string {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	respuesta := ""
	if login != "" {
		r, err := client.Get("https://localhost:8081/user/" + login) // Pedimos Por get
		check(err)

		//` `
		s := StreamToString(r.Body)
		a := strings.Split(s, "\"")

		for i, n := range a {
			if i%2 != 0 {
				respuesta += `<div class="file-box">  
			<div class="file">
				<a href="#" onclick="seleccionarArchivo('` + n + `')">
					<span class="corner"></span>
					<div class="icon">
						<i class="fa fa-file"></i>
					</div>
					<div class="file-name">
					` + n + `
						<br>
					</div>
				</a>
			</div>
		</div>`
			}
		}
	}

	return respuesta
}

func actualizarVista(sender *gowd.Element, event *gowd.EventElement) { //por si necesitamos hacer algo especial a la hora de actualizar
	main()
}

func cambiarVista(vista string) {
	mostrar = vista
}

func sendLogin(sender *gowd.Element, event *gowd.EventElement) {
	// ** ejemplo de login
	data := url.Values{} // estructura para contener los valores
	usuario := body.Find("usuario").GetValue()
	pass := body.Find("contraseña").GetValue()
	data.Set("login", usuario)
	data.Set("password", encriptarScrypt(pass, usuario))

	response := sendServerPetition(data, "/login")
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)

	var respuesta resp
	err := json.Unmarshal(buf.Bytes(), &respuesta)
	check(err)

	if respuesta.Ok == true {
		login = usuario
		cambiarVista("principal")
		actualizarVista(nil, nil)
	}
}

func sendRegister(sender *gowd.Element, event *gowd.EventElement) {
	// ** ejemplo de registro
	data := url.Values{} // estructura para contener los valores
	usuario := body.Find("registerUser").GetValue()
	pass := body.Find("registerPassword").GetValue()
	confirm := body.Find("confirmPassword").GetValue()
	data.Set("register", usuario)
	data.Set("password", encriptarScrypt(pass, usuario))
	data.Set("confirm", encriptarScrypt(confirm, usuario))

	response := sendServerPetition(data, "/register")

	//io.Copy(os.Stdout, r.Body) // mostramos el cuerpo de la respuesta (es un reader)
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	s := buf.String()
	body.Find("texto").SetText(s)
	body.Find("login-form-link").RemoveAttribute("active")
	body.Find("register-form-link").SetClass("active")
}

func seleccionarFichero(sender *gowd.Element, event *gowd.EventElement) {
	//fmt.Println(body.Find("archivo").GetValue())

	targetURL := "https://localhost:8081/upload"
	filename := body.Find("archivo").GetValue()
	postFile(filename, targetURL)
	cambiarVista("principal")
	actualizarVista(nil, nil)
}

func pedirFichero(sender *gowd.Element, event *gowd.EventElement) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	response, err := client.Post("https://localhost:8081/user/"+login+"/file/"+body.Find("archivoPedido").GetValue(), "application/json", nil) // Pedimos Por get
	check(err)

	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		//fmt.Printf("%s\n", string(contents))
		body.Find("texto").SetText(string(contents))
	}

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

func postFile(filename string, targetURL string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	err := bodyWriter.WriteField("Username", login)
	check(err)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	// open file handle
	fh, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file")
		return err
	}
	defer fh.Close()

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Post(targetURL, contentType, bodyBuf)
	check(err)

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	fmt.Println(string(respBody))
	return nil
}
