package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
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
	body = bootstrap.NewElement("div", "wrapper")
	body.SetAttribute("style", "background-color:#FF654E; height: 100%")

	logo := `<div style="margin:0 auto;width:40%;"><img src="assets/img/logo_alargado.png" style="width:100%;margin:0 auto"/></div>`

	switch mostrar {
	case "login":
		body.AddHTML(logo, nil)
		body.AddHTML(vistaLogin(), nil)
		body.Find("login-submit").OnEvent(gowd.OnClick, sendLogin)
		body.Find("register-form-link").OnEvent(gowd.OnClick, goRegister)
		break
	case "register":
		body.AddHTML(logo, nil)
		body.AddHTML(vistaRegister(), nil)
		body.Find("register-submit").OnEvent(gowd.OnClick, sendRegister)
		body.Find("login-form-link").OnEvent(gowd.OnClick, goLogin)
		break
	case "principal":
		body.AddHTML(vistaPrincipal(), nil)
		body.Find("recargar").OnEvent(gowd.OnClick, goPrincipal)
		body.Find("buttonEnviar").OnEvent(gowd.OnClick, seleccionarFichero)
		body.Find("logout-link").OnEvent(gowd.OnClick, goLogin)
		body.Find("buttonPedir").OnEvent(gowd.OnClick, pedirFichero)
		break
	}
	//start the ui loop
	err := gowd.Run(body)
	check(err)
}

func sendLogin(sender *gowd.Element, event *gowd.EventElement) {
	data := url.Values{} // estructura para contener los valores
	usuario := body.Find("usuario").GetValue()
	pass := body.Find("contraseña").GetValue()
	data.Set("login", usuario)
	data.Set("password", encriptarScrypt(pass, usuario))

	bytesJSON, err := json.Marshal(data)
	check(err)
	reader := bytes.NewReader(bytesJSON)

	response := sendServerPetition("POST", reader, "/login", "application/json")
	defer response.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)

	var respuesta resp
	err = json.Unmarshal(buf.Bytes(), &respuesta)
	check(err)

	body.Find("texto").SetText(buf.String())
	if respuesta.Ok == true {
		login = usuario
		token = response.Header.Get("Token")
		goPrincipal(nil, nil)
	}
}

func sendRegister(sender *gowd.Element, event *gowd.EventElement) {
	data := url.Values{} // estructura para contener los valores
	usuario := body.Find("registerUser").GetValue()
	pass := body.Find("registerPassword").GetValue()
	confirm := body.Find("confirmPassword").GetValue()
	data.Set("register", usuario)
	data.Set("password", encriptarScrypt(pass, usuario))
	data.Set("confirm", encriptarScrypt(confirm, usuario))

	bytesJSON, err := json.Marshal(data)
	check(err)
	reader := bytes.NewReader(bytesJSON)

	response := sendServerPetition("POST", reader, "/register", "application/json")
	defer response.Body.Close()

	s := streamToString(response.Body)
	body.Find("texto").SetText(s)
	body.Find("login-form-link").RemoveAttribute("active")
	body.Find("register-form-link").SetClass("active")
}

func seleccionarFichero(sender *gowd.Element, event *gowd.EventElement) {
	//fmt.Println(body.Find("archivo").GetValue())
	ruta := body.Find("route").GetValue()
	filename := body.Find("filename").GetValue()
	enviarFichero(ruta, encodeURLB64(filename))
	goPrincipal(nil, nil)
}

func enviarFichero(ruta string, filename string) {
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
	enviarParteFichero(contador, bytes, bytesLeidos, filename)

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
			enviarParteFichero(contador, bytes, bytesLeidos, filename)
		}
	}
	body.Find("texto").SetText(texto)
}

func enviarParteFichero(cont int, parte []byte, tam int, filename string) {
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

	bytesJSON, err := json.Marshal(data)
	check(err)
	reader := bytes.NewReader(bytesJSON)

	imprimir := "Pieza: " + contador + " hash: " + hex.EncodeToString(hash[:]) + " size: " + size + " user: " + login + " filename: " + filename
	body.Find("texto1").SetText(imprimir)

	/**************************** conseguir usuario *************************/
	response := sendServerPetition("POST", reader, "/checkhash", "application/json")
	defer response.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)

	var respuesta resp
	err = json.Unmarshal(buf.Bytes(), &respuesta)
	check(err)

	if err != nil || (respuesta.Ok == false && respuesta.Msg != "Hash comprobado") {
		//mostrar error y si es posible que esta funcion devuelva un error y el bucle de arriba pare
	} else if respuesta.Ok == false && respuesta.Msg == "Hash comprobado" { //el hash no existe en el servidor (la parte no se ha subido nunca)
		enviarDatos(parte, filename, contador, hex.EncodeToString(hash[:]))
	}
}

func enviarDatos(data []byte, filename string, parte string, hash string) {
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

	r := bytes.NewReader(data)
	_, err = io.Copy(fileWriter, r)
	check(err)

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	sendServerPetition("POST", bodyBuf, "/upload", contentType)
}

func pedirFichero(sender *gowd.Element, event *gowd.EventElement) {
	filename := encodeURLB64(body.Find("archivoPedido").GetValue())
	response := sendServerPetition("GET", nil, "/user/"+login+"/file/"+filename, "application/json")
	defer response.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	var respuestaJSON resp
	err := json.Unmarshal(buf.Bytes(), &respuestaJSON)
	if err == nil && respuestaJSON.Ok == false {
		//Cerrar sesion
		body.Find("texto").SetText(respuestaJSON.Msg)
	} else {
		respuesta := buf.String()
		//fmt.Printf("%s\n", string(contents))
		createDirIfNotExist("./descargas/")
		createFile("./descargas/" + body.Find("archivoPedido").GetValue())
		writeFile("./descargas/"+body.Find("archivoPedido").GetValue(), respuesta)
		body.Find("texto").SetText(body.Find("archivoPedido").GetValue())
	}
}

func peticionNombreFicheros() string {
	response := sendServerPetition("GET", nil, "/user/"+login, "application/json")
	defer response.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	var respuestaJSON resp
	err := json.Unmarshal(buf.Bytes(), &respuestaJSON)
	if err == nil && respuestaJSON.Ok == false {
		//Cerrar sesion
		return respuestaJSON.Msg
	}

	respuesta := ""
	a := strings.Split(buf.String(), "\"")
	for i, n := range a {
		if i%2 != 0 {
			respuesta += `<a href="#" class="list-group-item" 
				onclick="seleccionarArchivo('` + decodeURLB64(n) + `')">
					` + decodeURLB64(n) + `</a>`
		}
	}
	return respuesta
}
