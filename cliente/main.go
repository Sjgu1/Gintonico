package main

import (
	"bytes"
	"crypto/tls"
	"net/http"
	"net/url"

	"github.com/dtylman/gowd"
	"github.com/dtylman/gowd/bootstrap"
)

var body *gowd.Element
var mostrar = "login"

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

	body.AddHTML(`<div style="margin:0 auto;width:40%;"><img src="img/logo.png" style="width:100%;margin:0 auto"/></div><br/><br/>`, nil)

	switch mostrar {
	case "login":
		body.AddHTML(vistaLogin(), nil)
		body.Find("login-submit").OnEvent(gowd.OnClick, sendLogin)
		body.Find("register-form-link").OnEvent(gowd.OnClick, mostrarRegister)
		body.Find("login-form-link").OnEvent(gowd.OnClick, mostrarLogin)
		break
	case "register":
		body.AddHTML(vistaRegister(), nil)
		body.Find("register-submit").OnEvent(gowd.OnClick, sendRegister)
		body.Find("login-form-link").OnEvent(gowd.OnClick, mostrarLogin)
		body.Find("register-form-link").OnEvent(gowd.OnClick, mostrarRegister)
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

func mostrarLogin(sender *gowd.Element, event *gowd.EventElement) {
	mostrar = "login"
	main()
}

func mostrarRegister(sender *gowd.Element, event *gowd.EventElement) {
	mostrar = "register"
	main()
}

func sendLogin(sender *gowd.Element, event *gowd.EventElement) {

	// ** ejemplo de registro
	data := url.Values{} // estructura para contener los valores
	data.Set("login", body.Find("usuario").GetValue())
	data.Set("password", body.Find("contraseña").GetValue())

	response := sendServerPetition(data, "/login")

	//io.Copy(os.Stdout, r.Body) // mostramos el cuerpo de la respuesta (es un reader)
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	s := buf.String()
	body.Find("texto").SetText(s)
}

func sendRegister(sender *gowd.Element, event *gowd.EventElement) {

	// ** ejemplo de registro
	data := url.Values{} // estructura para contener los valores
	data.Set("register", body.Find("registerUser").GetValue())
	data.Set("password", body.Find("registerPassword").GetValue())
	data.Set("confirm", body.Find("confirmPassword").GetValue())

	response := sendServerPetition(data, "/register")

	//io.Copy(os.Stdout, r.Body) // mostramos el cuerpo de la respuesta (es un reader)
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	s := buf.String()
	body.Find("texto").SetText(s)
	body.Find("login-form-link").RemoveAttribute("active")
	body.Find("register-form-link").SetClass("active")
}
