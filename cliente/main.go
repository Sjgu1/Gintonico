package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/dtylman/gowd"
	"github.com/dtylman/gowd/bootstrap"
)

var body *gowd.Element
var mostrar = "login"
var login = ""

type resp struct {
	Ok  bool   `json:"ok"`  // true -> correcto, false -> error
	Msg string `json:"msg"` // mensaje adicional
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
	ruta := body.Find("route").GetValue()
	filename := body.Find("filename").GetValue()
	enviarFichero(ruta, encodeB64(filename), targetURL)
	//cambiarVista("principal")
	//actualizarVista(nil, nil)
}

func enviarFichero(ruta string, filename string, targetURL string) {
	f, err := os.Open(ruta)
	check(err)
	defer f.Close()
	bytes := make([]byte, 1024*1024*4) //byte -> kb -> mb
	bytesLeidos, err := f.Read(bytes)
	check(err)
	contador := 0
	contadorBytes := bytesLeidos
	texto := strconv.Itoa(contador) + ": " + strconv.Itoa(bytesLeidos) + ", "
	body.Find("texto").SetText(texto)
	enviarParteFichero(contador, bytes, bytesLeidos)
	for bytesLeidos > 0 {
		bytesLeidos, err = f.ReadAt(bytes, int64(contadorBytes))
		check(err)
		contador++
		contadorBytes += bytesLeidos
		//body.Find("texto").SetText(strconv.Itoa(bytesLeidos))
		if bytesLeidos > 0 {
			texto += strconv.Itoa(contador) + ": " + strconv.Itoa(bytesLeidos) + ", "
			enviarParteFichero(contador, bytes, bytesLeidos)
		}
	}

	body.Find("texto").SetText(string(bytes))
}

func enviarParteFichero(contador int, data []byte, size int) {
	//preparar peticion
	//hash := hashSHA256(data)
}

func pedirFichero(sender *gowd.Element, event *gowd.EventElement) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	filename := encodeB64(body.Find("archivoPedido").GetValue())

	response, err := client.Post("https://localhost:8081/user/"+login+"/file/"+filename, "application/json", nil) // Pedimos Por get
	check(err)

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	check(err)
	//fmt.Printf("%s\n", string(contents))
	body.Find("texto").SetText(string(contents))
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
		s := streamToString(r.Body)
		a := strings.Split(s, "\"")

		for i, n := range a {
			if i%2 != 0 {
				respuesta += `<div class="file-box">  
			<div class="file">
				<a href="#" onclick="seleccionarArchivo('` + decodeB64(n) + `')">
					<span class="corner"></span>
					<div class="icon">
						<i class="fa fa-file"></i>
					</div>
					<div class="file-name">
					` + decodeB64(n) + `
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
