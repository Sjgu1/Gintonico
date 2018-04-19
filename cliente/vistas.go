package main

import "github.com/dtylman/gowd"

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
										<input type="text" id="usuario" class="form-control" placeholder="Usuario" autocomplete="new-password">
									</div>
									<div class="form-group">
										<input type="password" id="contraseña" class="form-control" placeholder="Contraseña" autocomplete="new-password">
									</div>
									<div class="form-group">
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
	<p id="texto"/><p id="texto1"/><p id="texto2"/>`
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
	<p id="texto"/><p id="texto1"/><p id="texto2"/>`
}

func vistaPrincipal() string {
	return `<header class="main-header"><nav class="navbar navbar-static-top">
		<div class="container-fluid">
			<div class="navbar-header">
				<button type="button" class="navbar-toggle collapsed" data-toggle="collapse" data-target="#navbar-collapse">
					<i class="fa fa-bars"></i>
				</button>
				<a class="navbar-brand" href="#" id="recargar">Gintónico</a>
			</div>
			<div class="collapse navbar-collapse" id="navbar-collapse">
				<ul class="nav navbar-nav">
				<li class="active"><a href="#">Principal <span class="sr-only">(current)</span></a></li>
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
	</nav></header>
	<div class="content-wrapper">
		</br>
		<div class="row" style="margin: 0 auto;">
			<div class="col-md-3">
				<div class="box">
					<div class="box-header">
						<h3 class="box-title">Gintónico</h3>
					</div>
					<div class="box-body">
						<input type="file" id="idFile" onchange="subirArchivo()" style="display: none"/>
						<input type="text" id="route" style="display: none" />
						<input type="text" id="filename" style="display: none" />
						<input type="button" onclick="document.getElementById('idFile').click();"  value="Seleccionar Archivo" id="file-selector" class="btn btn-primary btn-block"/>
						<button type="button"  style="display: none"id="buttonEnviar"  class="btn btn-primary btn-block"> Subir </button>
						<button type="button"  style="display: none" id="buttonPedir" class="btn btn-primary btn-block">Pedir</button>
						<input type="text" id="archivoPedido" style="display: none" />
							
						<div class="clearfix"></div>
					</div>
				</div>
			</div>
			<div class="col-md-9" style="margin-bottom: 40px;">
				<div class="box">
					<div class="box-header">
						<h3 class="box-title">Ficheros</h3>
					</div>
					<div class="box-body">
						<table id="tabla" class="table table-striped table-bordered dataTable no-footer" style="width:100%">
							<thead>
								<tr>
									<th>Archivo</th>
								</tr>
							</thead>
							<tbody>
								` + peticionNombreFicheros() + `
							</tbody>
						</table>
					</div>
				</div>
			</div>
			<p id="texto"/><p id="texto1"/><p id="texto2"/>
		</div>
	</div>
	<footer class="main-footer" style="bottom:0;position:fixed;width:100%">
		<div class="container">
			<div class="pull-right hidden-xs">
				<b>Version: </b>&nbsp;1.0.0
			</div>
			<strong>Copyright © 2018&nbsp;<a href="#"> Gintónico </a>.</strong>&nbsp;&nbsp;&nbsp;Todos los derechos reservados.
		</div>
	</footer>
	`
}

func goLogin(sender *gowd.Element, event *gowd.EventElement) {
	mostrar = "login"
	login = ""
	token = ""
	main()
}

func goRegister(sender *gowd.Element, event *gowd.EventElement) {
	mostrar = "register"
	main()
}

func goPrincipal(sender *gowd.Element, event *gowd.EventElement) {
	mostrar = "principal"
	main()
}
