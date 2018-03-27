package main

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
							<button type="button"  style="display: none"id="buttonEnviar"  class="btn btn-primary btn-block " > Subir </button>
							<!--<button  ype="button" class="btn btn-primary btn-md">Subir un fichero</button>-->
							<button type="button"  style="display: none"id="buttonPedir"  class="btn btn-primary btn-block " > Subir </button>
							<input type="text" id="archivoPedido" style="display: none" />
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
					` + peticionNombreFicheros() + `
					</div>
				</div>
			</div>
		</div>
	</div>
	<p id="texto"/>
	</nav>`
}
