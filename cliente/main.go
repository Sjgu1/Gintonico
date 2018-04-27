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
	logo := `<div style="margin:0 auto;width:40%;"><img src="assets/img/logo_alargado.png" style="width:100%;margin:0 auto"/></div>`

	switch mostrar {
	case "login":
		body.SetAttribute("style", "background-color:#FF654E; height: 100%")
		body.AddHTML(logo, nil)
		body.AddHTML(vistaLogin(), nil)
		body.Find("login-submit").OnEvent(gowd.OnClick, sendLogin)
		body.Find("register-form-link").OnEvent(gowd.OnClick, goRegister)
		break
	case "register":
		body.SetAttribute("style", "background-color:#FF654E; height: 100%")
		body.AddHTML(logo, nil)
		body.AddHTML(vistaRegister(), nil)
		body.Find("register-submit").OnEvent(gowd.OnClick, sendRegister)
		body.Find("login-form-link").OnEvent(gowd.OnClick, goLogin)
		break
	case "principal":
		body.SetAttribute("style", "background-color:#ecf0f5; height: 100%")
		body.AddHTML(vistaPrincipal(), nil)
		body.Find("recargar").OnEvent(gowd.OnClick, goPrincipal)
		body.Find("buttonEnviar").OnEvent(gowd.OnClick, seleccionarFichero)
		body.Find("logout-link").OnEvent(gowd.OnClick, goLogin)
		body.Find("buttonPedir").OnEvent(gowd.OnClick, pedirFichero)
		body.Find("buttonEliminar").OnEvent(gowd.OnClick, eliminarFichero)
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
		goLogin(nil, nil)
		//mostrar error y si es posible que esta funcion devuelva un error y el bucle de arriba pare
	} else if respuesta.Ok == false && respuesta.Msg == "Hash comprobado" { //el hash no existe en el servidor (la parte no se ha subido nunca)
		enviarDatos(parte, filename, contador, hex.EncodeToString(hash[:]), size)
	}
}

func enviarDatos(data []byte, filename string, parte string, hash string, size string) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	err := bodyWriter.WriteField("Username", login)
	check(err)
	err = bodyWriter.WriteField("Parte", parte)
	check(err)
	err = bodyWriter.WriteField("Hash", hash)
	check(err)
	err = bodyWriter.WriteField("Size", size)
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
	if err == nil && respuestaJSON.Ok == false && respuestaJSON.Msg != "" {
		//Cerrar sesion
		goLogin(nil, nil)
		body.Find("texto").SetText(respuestaJSON.Msg)
	} else {
		respuesta := buf.String()
		//fmt.Printf("%s\n", string(contents))
		createDirIfNotExist("./descargas/" + login)
		createFile("./descargas/" + login + "/" + body.Find("archivoPedido").GetValue())
		writeFile("./descargas/"+login+"/"+body.Find("archivoPedido").GetValue(), respuesta)
		body.Find("texto").SetText("Fichero en descargas: " + body.Find("archivoPedido").GetValue())
	}
}

func peticionNombreFicheros() string {
	response := sendServerPetition("GET", nil, "/user/"+login, "application/json")
	defer response.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	var respuestaJSON resp
	err := json.Unmarshal(buf.Bytes(), &respuestaJSON)
	respuesta := ""

	if err == nil && respuestaJSON.Ok == false && respuestaJSON.Msg != "" {
		//Cerrar sesion
		//return respuestaJSON.Msg
		//goLogin(nil, nil)
		return respuesta
	}

	type FilesJSON struct {
		Filename []string `json:"filename"`
		Size     []string `json:"size"`
	}
	var filesJSON FilesJSON
	err = json.Unmarshal(buf.Bytes(), &filesJSON)
	if err == nil && len(filesJSON.Filename) != 0 && len(filesJSON.Size) != 0 && len(filesJSON.Filename) == len(filesJSON.Size) {
		for i := range filesJSON.Filename {
			//respuesta += filesJSON.Filename[i] + filesJSON.Size[i]
			/*<div class="dropdown">
				<a href="#" class="dropdown-toggle" data-toggle="dropdown" role="button" aria-haspopup="true" aria-expanded="false">` + decodeURLB64(filesJSON.Filename[i]) + `</a>
				<ul class="dropdown-menu dropdown-menu-files" style="background-color: #53A3CD;">
					<li><a href="#" onclick="seleccionarArchivo('` + decodeURLB64(filesJSON.Filename[i]) + `')">Descargar</a></li>
					<li><a href="#" onclick="eliminarArchivo('` + decodeURLB64(filesJSON.Filename[i]) + `')">Eliminar</a></li>
				</ul>
			</div>*/
			tamanyo, _ := strconv.Atoi(filesJSON.Size[i])
			respuesta += `<tr>
				<td>
					<a href="#">` + decodeURLB64(filesJSON.Filename[i]) + `</a>
					<span style="float:right;">&nbsp;</span>
					<span style="float:right;">&nbsp;</span>
					<button type="button" class="btn btn-danger btn-xs" style="float: right;" onclick="eliminarArchivo('` + decodeURLB64(filesJSON.Filename[i]) + `')">
						<span class="glyphicon glyphicon-trash" aria-hidden="true"></span>
					</button>
					<span style="float:right;">&nbsp;</span>
					<span style="float:right;">&nbsp;</span>
					<button type="button" class="btn btn-primary btn-xs" style="float: right;" onclick="seleccionarArchivo('` + decodeURLB64(filesJSON.Filename[i]) + `')">
						<span class="glyphicon glyphicon-download-alt" aria-hidden="true"></span>
					</button>
					<span style="float:right;">&nbsp;</span>
					<span style="float:right;">&nbsp;</span>
				</td>
				<td>
					` + formatBytesToString(tamanyo) + `
				</td>
			</tr>`
		}

	}
	return respuesta
}

func eliminarFichero(sender *gowd.Element, event *gowd.EventElement) {
	filename := encodeURLB64(body.Find("archivoEliminar").GetValue())
	body.Find("texto").SetText("Eliminando: " + decodeURLB64(filename))
	response := sendServerPetition("DELETE", nil, "/user/"+login+"/file/"+filename, "application/json")
	defer response.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	var respuestaJSON resp
	err := json.Unmarshal(buf.Bytes(), &respuestaJSON)
	if err == nil && respuestaJSON.Ok == false && respuestaJSON.Msg != "" {
		//Cerrar sesion
		goLogin(nil, nil)
		body.Find("texto").SetText(respuestaJSON.Msg)
	} else {
		goPrincipal(nil, nil)
	}
}
