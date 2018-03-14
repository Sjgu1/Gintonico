package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"net/url"

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
		panic(e)
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
		body.Find("logout-link").OnEvent(gowd.OnClick, actualizarVista)
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
								<a id="login-form-link" href="#" class="active">Login</a>
							</div>
							<div class="col-xs-6">
								<a id="register-form-link" href="#">Register</a>
							</div>
						</div>
						<hr>
					</div>
					<div class="panel-body">
						<div class="row">
							<div class="col-lg-12">
								<div id="login-form">
									<div class="form-group">
										<input type="text" id="usuario" class="form-control" placeholder="Usuario" autocomplete="off">
									</div>
									<div class="form-group">
										<input type="password" id="contraseña" class="form-control" placeholder="Contraseña" autocomplete="off">
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
								<a id="login-form-link" href="#">Login</a>
							</div>
							<div class="col-xs-6">
								<a id="register-form-link" href="#" class="active">Register</a>
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
			  <li><a href="#" id="logout-link" >Cerrar sesión</a></li>
			</ul>
		  </li>
		</ul>
	  </div>
	</div>
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

// Devuelve el string de la cadena encriptada
func encriptarScrypt(cadena string, seed string) string {
	salt := []byte(seed)

	dk, err := scrypt.Key([]byte(cadena), salt, 1<<15, 10, 1, 32)
	if err != nil {
		log.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(dk)
}
