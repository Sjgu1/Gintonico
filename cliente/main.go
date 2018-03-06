package main

import (
	"bytes"
	"net/http"
	"net/url"

	"github.com/dtylman/gowd"

	"github.com/dtylman/gowd/bootstrap"

	"crypto/tls"
)

var body *gowd.Element

// función para comprobar errores (ahorra escritura)
func check(e error) {
	if e != nil {
		panic(e)
	}
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
									<div class="form-group text-center">
										<input type="checkbox" tabindex="3" class="" name="remember" id="remember">
										<label for="remember"> Remember Me</label>
									</div>
									<div class="form-group">
										<div class="row">
											<div class="col-sm-6 col-sm-offset-3">
												<button name="login-submit" id="login-submit" tabindex="4" class="form-control btn btn-login">Iniciar Sesión</button>
											</div>
										</div>
									</div>
									<div class="form-group">
										<div class="row">
											<div class="col-lg-12">
												<div class="text-center">
													<a href="#" tabindex="5" class="forgot-password">Forgot Password?</a>
												</div>
											</div>
										</div>
									</div>
								</div>
								<div id="register-form" style="display: none;">
									<div class="form-group">
										<input type="text" name="username" id="username" tabindex="1" class="form-control" placeholder="Username" value="">
									</div>
									<div class="form-group">
										<input type="email" name="email" id="email" tabindex="1" class="form-control" placeholder="Email Address" value="">
									</div>
									<div class="form-group">
										<input type="password" name="password" id="password" tabindex="2" class="form-control" placeholder="Password">
									</div>
									<div class="form-group">
										<input type="password" name="confirm-password" id="confirm-password" tabindex="2" class="form-control" placeholder="Confirm Password">
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

	body.Find("login-submit").OnEvent(gowd.OnClick, btnPrueba)

	//start the ui loop
	gowd.Run(body)
}

func btnPrueba(sender *gowd.Element, event *gowd.EventElement) {
	/*log.SetFlags(log.Lshortfile)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	buf, err := client.Get("https://localhost:8081")
	check(err)
	var buffer = new(bytes.Buffer)
	buffer.ReadFrom(buf.Body)
	var resul = buffer.String()

	println(resul)
	body.Find("texto").SetText(resul)*/
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// ** ejemplo de registro
	data := url.Values{}            // estructura para contener los valores
	data.Set("login", "hola")       // comando (string)
	data.Set("password", "saludos") // usuario (string)

	r, err := client.PostForm("https://localhost:8081/login", data) // enviamos por POST
	check(err)
	//io.Copy(os.Stdout, r.Body) // mostramos el cuerpo de la respuesta (es un reader)
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	s := buf.String()
	body.Find("texto").SetText(s)
}
