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
	body.AddHTML(`<div class="container">
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
								<div id="login-form" style="display: block;">
									<div class="form-group">
										<input type="text" name="username" id="username" tabindex="1" class="form-control" placeholder="Username" value="">
									</div>
									<div class="form-group">
										<input type="password" name="password" id="password" tabindex="2" class="form-control" placeholder="Password">
									</div>
									<div class="form-group">
										<div class="row">
											<div class="col-sm-6 col-sm-offset-3">
												<button name="login-submit" id="login-submit" tabindex="4" class="form-control btn btn-login">Iniciar Sesión</button>
											</div>
										</div>
									</div>
								</div>
								<div id="register-form" style="display: none;">
									<div class="form-group">
										<input type="text" name="registerUser" id="registerUser" tabindex="1" class="form-control" placeholder="Username" value="">
									</div>
									<div class="form-group">
										<input type="password" name="registerPassword" id="registerPassword" tabindex="2" class="form-control" placeholder="Password">
									</div>
									<div class="form-group">
										<input type="password" name="confirmPassword" id="confirmPassword" tabindex="2" class="form-control" placeholder="Confirm Password">
									</div>
									<div class="form-group">
										<div class="row">
											<div class="col-sm-6 col-sm-offset-3">
												<button name="register-submit" id="register-submit" tabindex="4" class="form-control btn btn-register">Regístrate ya!</button>
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
	<p id="texto"/>`, nil)

	body.Find("login-submit").OnEvent(gowd.OnClick, btnLogin)
	body.Find("register-submit").OnEvent(gowd.OnClick, btnRegister)

	//start the ui loop
	gowd.Run(body)
}

func btnLogin(sender *gowd.Element, event *gowd.EventElement) {

	// ** ejemplo de registro
	data := url.Values{} // estructura para contener los valores
	data.Set("login", body.Find("username").GetValue())
	data.Set("password", body.Find("password").GetValue())

	response := sendServerPetition(data, "/login")

	//io.Copy(os.Stdout, r.Body) // mostramos el cuerpo de la respuesta (es un reader)
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	s := buf.String()
	body.Find("texto").SetText(s)
}

func btnRegister(sender *gowd.Element, event *gowd.EventElement) {

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
