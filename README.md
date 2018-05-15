<p align="center"><img src="/cliente/assets/img/logo_redondeado.png" width="400"></img></p>

# Gintónico &nbsp; <img src="https://travis-ci.org/golang/dep.svg?branch=master" alt="Build Status"></img> <img src="https://img.shields.io/badge/license-GPL%20v3-blue.svg" alt="License"></img> <img src="https://ci.appveyor.com/api/projects/status/github/golang/dep?svg=true&branch=master&passingText=Windows%20-%20OK&failingText=Windows%20-%20failed&pendingText=Windows%20-%20pending" alt="Windows Build Status"></img> <img src="https://ci.appveyor.com/api/projects/status/github/golang/dep?svg=true&branch=master&passingText=MacOS%20-%20OK&failingText=MacOS%20-%20failed&pendingText=MacOS%20-%20pending" alt="MacOS Build Status"></img> <img src="https://ci.appveyor.com/api/projects/status/github/golang/dep?svg=true&branch=master&passingText=Linux%20-%20OK&failingText=Linux%20-%20failed&pendingText=Linux%20-%20pending" alt="Linux Build Status"></img>

Gintónico es un sistema de alojamiento de archivos en la nube (Arquitectura cliente/servidor). Está implementado con una de las mejores prácticas de seguridad en la actualidad. Esta disponible para Windows, MacOS y Linux, y el cliente dispone de una interfaz gráfica usable, simple y eficiente.

<p align="center"><img src="/cliente/assets/img/captura_app.png"></img></p>

## Guía de instalación

Siga estas instrucciones para obtener una copia de este proyecto funcionando correctamente en tu sistema operativo.

### Prerrequisitos <img align="right" width="200" src="/cliente/assets/img/golang_gintonico.png"></img> 
Para empezar es necesario que tanto el lenguaje (Golang) esté bien instalado como los paths bien configurados (Depende de cada sistema operativo). 

Una vez dispongamos de los medios para poder compilar y ejecutar Go necesitaremos descargarnos las librerías que se usan en el proyecto:

*Ejecutar este comando tanto en la carpeta servidor como en el cliente*
```
go get
```

Ahora necesitaremos descargarnos NW.js para poder ejecutar y poner en funcionamiento la interfaz gráfica del cliente de Gintónico:

*Descargar el correspondiente a nuestro sistema operativo*
```
https://nwjs.io/downloads/
```

Una vez tengamos NW para el sistema que necesitemos, lo copiaremos dentro del cliente en una carpeta llamada "nwjs".

### Instalando y ejecutando Gintónico

Primero compilamos Gintónico con el comando:

*Habrá que compilar el código tanto del cliente como del servidor*
```
go build
```

Una vez compilado el proyecto ya podremos ejecutar el cliente con NW.js. Para hacer esto (suponiendo que el ejecutable de NW está dentro de una carpeta llamada nwjs), habrá que ejecutar este comando:

*Desde la carpeta cliente*
Windows:
```
.\nwjs\nw.exe .
```
MacOS:
```
./nwjs/nwjs.app/Contents/MacOS/nwjs .
```

Por último, para poner en funcionamiento el servidor, solo habrá que iniciar el ejecutable que se habrá generado al ejecutar el comando go build anteriormente.

A continuación se muestra una pequeña y rápida demo:
<p align="center"><img src="/cliente/assets/img/gif.gif" width="750"></img></p>

## Librerías

* [GoWD](https://github.com/dtylman/gowd) - Interfaz Gráfica utilizada
* [Mux](https://github.com/gorilla/mux) - Manejador de rutas HTTP
* [HTTPS Certs](https://github.com/kabukky/httpscerts) - Gestor de certificados HTTPS
* [JWT Token](https://github.com/dgrijalva/jwt-go) - Implementación de JSON Web Tokens

## Autores

* **Sergio Julio García Urdiales** - *Programador* - [Sergio](https://github.com/Sjgu1)
* **Lawrence Rider García** - *Programador* - [Larry](http://www.larryrider.es)

Puedes ver también la lista de los [contribuidores](https://github.com/Sjgu1/Gintonico/contributors) que han participado en este proyecto.

## Licencia

Este proyecto está bajo la licencia GNU GPL v3 - revisa [LICENSE](LICENSE) para ver más detalles.
