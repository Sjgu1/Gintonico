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
		body.AddHTML(vistaPrincipal2(), nil)
		body.Find("buttonEnviar").OnEvent(gowd.OnClick, seleccionarFichero)
		body.Find("logout-link").OnEvent(gowd.OnClick, actualizarVista)
		//body.Find("file-selector").OnEvent(gowd.OnClick, seleccionarFichero)
		cambiarVista("login")
		break
	}

	//start the ui loop
	err := gowd.Run(body)
	check(err)
}

func vistaLogin() string {
	return `<div class="container">
		<div class="row">
			<div class="col-md-6 col-md-offset-3">
				<div class="panel panel-login">
					<div class="panel-heading">
						<div class="row">
							<div class="col-xs-6">
								<a id="login-form-link" href="#" class="active">Iniciar Sesión</a>
							</div>
							<div class="col-xs-6">
								<a id="register-form-link" href="#">Registro</a>
							</div>
						</div>
						<hr>
					</div>
					<div class="panel-body">
						<div class="row">
							<div class="col-lg-12">
								<div id="login-form">
									<div class="form-group">
										<input type="text" id="usuario" class="form-control" placeholder="Usuario">
									</div>
									<div class="form-group">
										<input type="password" id="contraseña" class="form-control" placeholder="Contraseña">
									</div>
									<div class="form-group">
										<div class="row">
											<div class="col-sm-6 col-sm-offset-3">
												<button id="login-submit" class="form-control btn btn-login">Iniciar Sesión</button>
											</div>
										</div>
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
	<p id="texto"/>`
}

func vistaRegister() string {
	return `<div class="container">
		<div class="row">
			<div class="col-md-6 col-md-offset-3">
				<div class="panel panel-login">
					<div class="panel-heading">
						<div class="row">
							<div class="col-xs-6">
								<a id="login-form-link" href="#">Iniciar Sesión</a>
							</div>
							<div class="col-xs-6">
								<a id="register-form-link" href="#" class="active">Registro</a>
							</div>
						</div>
						<hr>
					</div>
					<div class="panel-body">
						<div class="row">
							<div class="col-lg-12">
								<div id="register-form">
									<div class="form-group">
										<input type="text" id="registerUser" class="form-control" placeholder="Username" autocomplete="off">
									</div>
									<div class="form-group">
										<input type="password" id="registerPassword" class="form-control" placeholder="Password" autocomplete="off">
									</div>
									<div class="form-group">
										<input type="password" id="confirmPassword" class="form-control" placeholder="Confirm Password" autocomplete="off">
									</div>
									<div class="form-group">
										<div class="row">
											<div class="col-sm-6 col-sm-offset-3">
												<button id="register-submit" class="form-control btn btn-register">Regístrate ya!</button>
											</div>
										</div>
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
	<p id="texto"/>`
}

func vistaPrincipal() string {
	return `<nav class="navbar navbar-default">
	<div class="container-fluid">
	  <div class="navbar-header">
		<button type="button" class="navbar-toggle collapsed" data-toggle="collapse" data-target="#bs-example-navbar-collapse-1" aria-expanded="false">
		  <span class="sr-only">Toggle navigation</span>
		  <span class="icon-bar"></span>
		  <span class="icon-bar"></span>
		  <span class="icon-bar"></span>
		</button>
		<a class="navbar-brand" href="#">Gintónico</a>
	  </div>
	  <div class="collapse navbar-collapse" id="bs-example-navbar-collapse-1">
		<ul class="nav navbar-nav">
		  <li class="active"><a href="#">Principal <span class="sr-only">(current)</span></a></li>
		  <li><a href="#">Otra página</a></li>
		</ul>
		<ul class="nav navbar-nav navbar-right">
		  <li><a> Bienvenido/a ` + login + ` !</a></li>
		  <li class="dropdown">
			<a href="#" class="dropdown-toggle" data-toggle="dropdown" role="button" aria-haspopup="true" aria-expanded="false">Ajustes <span class="caret"></span></a>
			<ul class="dropdown-menu">
			  <li><a href="#">Accion increíble</a></li>
			  <li><a href="#">Esta es mejor</a></li>
			  <li role="separator" class="divider"></li>
			  <li><a href="#" id="logout-link"><i class="icon-off"></i>Cerrar sesión</a></li>
			</ul>
		  </li>
		</ul>
	  </div>
	  </br>
	  </br>
	  </br>
	  <!--<button id="file-selector" type="button" class="btn btn-primary btn-md">Selecciona un fichero</button>-->
	  </br>
	  </br>
	  </br>
	</div>
  </nav>`
}

func vistaPrincipal2() string {
	return `<nav class="navbar navbar-default">
	<div class="container-fluid">
	  <div class="navbar-header">
		<button type="button" class="navbar-toggle collapsed" data-toggle="collapse" data-target="#bs-example-navbar-collapse-1" aria-expanded="false">
		  <span class="sr-only">Toggle navigation</span>
		  <span class="icon-bar"></span>
		  <span class="icon-bar"></span>
		  <span class="icon-bar"></span>
		</button>
		<a class="navbar-brand" href="#">Gintónico</a>
	  </div>
	  <div class="collapse navbar-collapse" id="bs-example-navbar-collapse-1">
		<ul class="nav navbar-nav">
		  <li class="active"><a href="#">Principal <span class="sr-only">(current)</span></a></li>
		  <li><a href="#">Otra página</a></li>
		</ul>
		<ul class="nav navbar-nav navbar-right">
		  <li><a> Bienvenido/a ` + login + ` !</a></li>
		  <li class="dropdown">
			<a href="#" class="dropdown-toggle" data-toggle="dropdown" role="button" aria-haspopup="true" aria-expanded="false">Ajustes <span class="caret"></span></a>
			<ul class="dropdown-menu">
			  <li><a href="#">Accion increíble</a></li>
			  <li><a href="#">Esta es mejor</a></li>
			  <li role="separator" class="divider"></li>
			  <li><a href="#" id="logout-link"><i class="icon-off"></i>Cerrar sesión</a></li>
			</ul>
		  </li>
		</ul>
	  </div>
	</div>
	</br>
	</br>
	<div class="container" style="background-color:#F8F8F8;width:97%;margin-right: 0px;margin-left:15px;">
		<div class="row" style="margin: 0 auto;">
			<div class="col-md-3">
				<div class="ibox float-e-margins">
					<div class="ibox-content">
						<div class="file-manager">
							<h5>Show:</h5>
							<a href="#" class="file-control active">Ale</a>
							<a href="#" class="file-control">Documents</a>
							<a href="#" class="file-control">Audio</a>
							<a href="#" class="file-control">Images</a>
							<div class="hr-line-dashed"></div>
							<input type="file" id="idFile" onchange="subirArchivo()" style="display: none"/>
							<input type="text" id="archivo" style="display: none" />
							<input type="button" onclick="document.getElementById('idFile').click();"  value="Seleccionar Archivo" id="file-selector" class="btn btn-primary btn-block"/>
							<button type="button" id="buttonEnviar"  class="btn btn-primary btn-block " > Subir </button>
							<!--<button  ype="button" class="btn btn-primary btn-md">Selecciona un fichero</button>-->
							<div class="hr-line-dashed"></div>
							<h5>Folders</h5>
							<ul class="folder-list" style="padding: 0">
								<li><a href=""><i class="fa fa-folder"></i> Files</a></li>
								<li><a href=""><i class="fa fa-folder"></i> Pictures</a></li>
								<li><a href=""><i class="fa fa-folder"></i> Web pages</a></li>
								<li><a href=""><i class="fa fa-folder"></i> Illustrations</a></li>
								<li><a href=""><i class="fa fa-folder"></i> Films</a></li>
								<li><a href=""><i class="fa fa-folder"></i> Books</a></li>
							</ul>
							<h5 class="tag-title">Tags</h5>
							<ul class="tag-list" style="padding: 0">
								<li><a href="">Family</a></li>
								<li><a href="">Work</a></li>
								<li><a href="">Home</a></li>
								<li><a href="">Children</a></li>
								<li><a href="">Holidays</a></li>
								<li><a href="">Music</a></li>
								<li><a href="">Photography</a></li>
								<li><a href="">Film</a></li>
							</ul>
							<div class="clearfix"></div>
						</div>
					</div>
				</div>
			</div>
			<div class="col-md-9">
				<div class="row" style="margin: 0 auto;">
					<div class="col-lg-12">
						<div class="file-box">
							<div class="file">
								<a href="#">
									<span class="corner"></span>
									<div class="icon">
										<i class="fa fa-file"></i>
									</div>
									<div class="file-name">
										Document_2014.doc
										<br>
										<small>Added: Jan 11, 2014</small>
									</div>
								</a>
							</div>
						</div>
						<div class="file-box">
							<div class="file">
								<a href="#">
									<span class="corner"></span>
									<div class="image">
										<img alt="image" class="img-responsive" src="https://lorempixel.com/400/300/nature/1">
									</div>
									<div class="file-name">
										Italy street.jpg
										<br>
										<small>Added: Jan 6, 2014</small>
									</div>
								</a>
							</div>
						</div>
						<div class="file-box">
							<div class="file">
								<a href="#">
									<span class="corner"></span>
									<div class="image">
										<img alt="image" class="img-responsive" src="https://lorempixel.com/400/300/nature/2">
									</div>
									<div class="file-name">
										My feel.png
										<br>
										<small>Added: Jan 7, 2014</small>
									</div>
								</a>
							</div>
						</div>
						<div class="file-box">
							<div class="file">
								<a href="#">
									<span class="corner"></span>
									<div class="icon">
										<i class="fa fa-music"></i>
									</div>
									<div class="file-name">
										Michal Jackson.mp3
										<br>
										<small>Added: Jan 22, 2014</small>
									</div>
								</a>
							</div>
						</div>
						<div class="file-box">
							<div class="file">
								<a href="#">
									<span class="corner"></span>
									<div class="image">
										<img alt="image" class="img-responsive" src="https://lorempixel.com/400/300/nature/3">
									</div>
									<div class="file-name">
										Document_2014.doc
										<br>
										<small>Added: Fab 11, 2014</small>
									</div>
								</a>
							</div>
						</div>
						<div class="file-box">
							<div class="file">
								<a href="#">
									<span class="corner"></span>
									<div class="icon">
										<i class="img-responsive fa fa-film"></i>
									</div>
									<div class="file-name">
										Monica's birthday.mpg4
										<br>
										<small>Added: Fab 18, 2014</small>
									</div>
								</a>
							</div>
						</div>
						<div class="file-box">
							<a href="#">
								<div class="file">
									<span class="corner"></span>
									<div class="icon">
										<i class="fa fa-bar-chart-o"></i>
									</div>
									<div class="file-name">
										Annual report 2014.xls
										<br>
										<small>Added: Fab 22, 2014</small>
									</div>
								</div>
							</a>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
	<p id="texto"/>
	</nav>`
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
	body.Find("texto").SetText("Hola que tal") //filename, err := dialog.File().Filter("Mp3 audio file", "mp3").Load()

	targetURL := "https://localhost:8081/upload"
	filename := body.Find("archivo").GetValue()
	postFile(filename, targetURL)

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
