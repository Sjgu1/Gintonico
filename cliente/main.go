package main

import (
	"bytes"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
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
var token = ""

type resp struct {
	Ok  bool   `json:"ok"`  // true -> correcto, false -> error
	Msg string `json:"msg"` // mensaje adicional
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

	response := sendServerPetition(data, "/login", usuario)
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)

	var respuesta resp
	err := json.Unmarshal(buf.Bytes(), &respuesta)
	check(err)

	body.Find("texto").SetText(buf.String())
	if respuesta.Ok == true {
		login = usuario
		token = response.Header.Get("Token")

		cambiarVista("principal")
		actualizarVista(nil, nil)

		//body.Find("texto").SetText(token)

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

	response := sendServerPetition(data, "/register", usuario)

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

	ruta := body.Find("route").GetValue()
	filename := body.Find("filename").GetValue()
	enviarFichero(ruta, encodeURLB64(filename))
	cambiarVista("principal")
	actualizarVista(nil, nil)
}

func enviarFichero(ruta string, filename string) {
	checkHashURL := "/checkhash"
	f, err := os.Open(ruta)
	check(err)
	defer f.Close()
	bytesTam := 1024 * 1024 * 4 //byte -> kb -> mb * 4
	bytes := make([]byte, bytesTam)
	bytesLeidos, err := f.Read(bytes)
	check(err)

	if bytesLeidos > 0 && bytesLeidos < bytesTam { //si solo hay una parte
		bytes = bytes[:bytesLeidos] // para que no ocupe 4mb siempre
	}

	contador := 0
	contadorBytes := bytesLeidos
	texto := strconv.Itoa(contador) + ": " + strconv.Itoa(bytesLeidos) + ", "
	enviarParteFichero(contador, bytes, bytesLeidos, checkHashURL, filename)

	for bytesLeidos > 0 {
		bytesLeidos, err = f.ReadAt(bytes, int64(contadorBytes))
		check(err)
		contador++
		contadorBytes += bytesLeidos
		if bytesLeidos > 0 {
			if bytesLeidos < bytesTam { //ultima parte
				bytes = bytes[:bytesLeidos] // para que no ocupe 4mb siempre
			}
			texto += strconv.Itoa(contador) + ": " + strconv.Itoa(bytesLeidos) + ", "
			enviarParteFichero(contador, bytes, bytesLeidos, checkHashURL, filename)
		}
	}

	body.Find("texto").SetText(texto)
}

func enviarParteFichero(cont int, parte []byte, tam int, checkHashURL string, filename string) {
	//preparar peticion
	data := url.Values{} // estructura para contener los valores
	contador := strconv.Itoa(cont)
	hash := hashSHA512(parte)
	size := strconv.Itoa(tam)
	data.Set("cont", contador)
	data.Set("hash", hex.EncodeToString(hash[:]))
	data.Set("size", size)
	data.Set("user", login)
	data.Set("filename", filename)

	imprimir := "Pieza: " + contador + " hash: " + hex.EncodeToString(hash[:]) + " size: " + size + " user: " + login + " filename: " + filename
	body.Find("texto1").SetText(imprimir)

	/**************************** conseguir usuario *************************/
	response := sendServerPetition(data, checkHashURL, login)
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)

	var respuesta resp
	err := json.Unmarshal(buf.Bytes(), &respuesta)
	check(err)

	if respuesta.Ok == false { //el hash no existe en el servidor (la parte no se ha subido nunca)
		enviarDatos(parte, filename, contador, hex.EncodeToString(hash[:]))
	}
}

func enviarDatos(data []byte, filename string, parte string, hash string) {
	targetURL := "https://localhost:8081/upload"
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	err := bodyWriter.WriteField("Username", login)
	check(err)
	err = bodyWriter.WriteField("Parte", parte)
	check(err)
	err = bodyWriter.WriteField("Hash", hash)
	check(err)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filename)
	check(err)

	//iocopy
	//fh, err := os.Open(ruta) ya no, ahora son bytes[]
	r := bytes.NewReader(data)
	_, err = io.Copy(fileWriter, r)
	check(err)

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("POST", targetURL, bodyBuf)
	req.Header.Set("Content-Type", contentType)

	req.Header.Add("Authorization", token)
	req.Header.Add("Username", login)

	resp, err := client.Do(req)
	//resp, err := client.Post(targetURL, contentType, bodyBuf)
	check(err)

	defer resp.Body.Close()
	/*respBody, err := ioutil.ReadAll(resp.Body)
	check(err)*/
}

func pedirFichero(sender *gowd.Element, event *gowd.EventElement) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	filename := encodeURLB64(body.Find("archivoPedido").GetValue())
	req, err := http.NewRequest("POST", "https://localhost:8081/user/"+login+"/file/"+filename, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", token)
	req.Header.Add("Username", login)
	response, err := client.Do(req)

	//response, err := client.Post("https://localhost:8081/user/"+login+"/file/"+filename, "application/json", nil) // Pedimos Por get
	check(err)

	defer response.Body.Close()
	//	contents, err := ioutil.ReadAll(response.Body)
	//	check(err)
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	newStr := buf.String()
	//fmt.Printf("%s\n", string(contents))
	createDirIfNotExist("./descargas/")
	createFile("./descargas/" + body.Find("archivoPedido").GetValue())
	writeFile("./descargas/"+body.Find("archivoPedido").GetValue(), newStr)

	body.Find("texto").SetText(body.Find("archivoPedido").GetValue())
}

func peticionNombreFicheros() string {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	respuesta := ""
	if login != "" {
		req, err := http.NewRequest("GET", "https://localhost:8081/user/"+login, nil)
		req.Header.Add("Authorization", token)
		req.Header.Add("Username", login)
		r, err := client.Do(req)
		//	r, err := client.Get("https://localhost:8081/user/" + login) // Pedimos Por get
		check(err)

		//` `
		s := streamToString(r.Body)
		a := strings.Split(s, "\"")

		for i, n := range a {
			if i%2 != 0 {
				respuesta += `<div class="file-box">  
			<div class="file">
				<a href="#" onclick="seleccionarArchivo('` + decodeURLB64(n) + `')">
					<span class="corner"></span>
					<div class="icon">
						<i class="fa fa-file"></i>
					</div>
					<div class="file-name">
					` + decodeURLB64(n) + `
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
