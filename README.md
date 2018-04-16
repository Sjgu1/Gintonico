# Gintónico - SDS

Gintónico es un sistema de alojamiento de archivos en la nube (Arquitectura cliente/servidor). Está implementado con una de las mejores prácticas de seguridad en la actualidad. Esta disponible para Windows, MacOS y Linux, y tiene una interfaz gráfica en el cliente.


## Guía de instalación

Siga estas instrucciones para obtener una copia de este proyecto funcionando correctamente en tu sistema operativo.


### Prerrequisitos

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

Por último, para poner en funcionamiento el servidor solo habrá que iniciar el ejecutable que se habrá generado al ejecutar el comando go build anteriormente

## Librerías

* [GoWD](https://github.com/dtylman/gowd) - Interfaz Gráfica utilizada
* [Mux](https://github.com/gorilla/mux) - Manejador de rutas HTTP
* [HTTPS Certs](https://github.com/kabukky/httpscerts) - Gestor de certificados HTTPS


## Autores

* **Sergio Julio García Urdiales** - *Programador* - [Serge](#)
* **Lawrence Rider García** - *Programador* - [Larry](http://www.larryrider.es)

Puedes ver también la lista de los [contribuidores](https://github.com/Sjgu1/Gintonico/contributors) que han participado en este proyecto.


## Licencia

Este proyecto esta bajo la MIT Licencia - see the [LICENSE.md](LICENSE.md) file for details


## Acknowledgments

* Hat tip to anyone who's code was used
* Inspiration
* etc
