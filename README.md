# Gintónico - SDS

Esta va a ser la mejor aplicacion en Go jamás vista.

## Guía de instalación

Siga estas instruccines para obtener una copia de este proyecto funcionando correctamente en tu sistema operativo.


### Prerrequisitos

Para empezar es necesario que tanto el lenguje esté bien instalado como los paths bien configurados (Depende de cada sistema operativo).

Una vez dispongamos de los medios para poder compilar y ejecutar Go necesitaremos descargarnos las librerías:

```
go get
```

Ahora necesitaremos descargarnos NW.js para poder ejecutar y poner en funcionamiento la interfaz gráfica de Gintónico:

```
https://nwjs.io/downloads/
```

Una vez tengamos NW para el sistema que necesitemos, lo meteremos por ejemplo dentro del proyecto en una carpeta llamada "nwjs".

### Instalando y ejecutando Gintónico

Primero comprobaremos por seguridad que están todas las librerías instaladas de nuevo con el comando:

```
go get
```

Primero compilamos Gintónico desde la carpeta raíz del proyecto con el comando: 

```
go build
```

Una vez compilado el proyecto ya podremos ejecutar la aplicación con NW.js pasandole al ejecutable del mismo la carpeta raíz del proyecto, por ejemplo:

(si se han seguido los pasos de prerrequisitos exactamente, tendremos una carpeta llamada nwjs dentro de la carpeta raíz del proyecto)
Windows:
```
.\nwjs\nw.exe .
```
MacOS:
```
./nwjs/nwjs.app/Contents/MacOS/nwjs .
```

## Deployment

Add additional notes about how to deploy this on a live system

## Built With

* [GoWD](https://github.com/dtylman/gowd) - Interfaz Gráfica utilizada

## Contributing

Please read [CONTRIBUTING.md](#) for details on our code of conduct, and the process for submitting pull requests to us.

## Versioning

We use [SemVer](#) for versioning. For the versions available, see the [tags on this repository](https://github.com/your/project/tags). 

## Authors

* **Sergio Julio García Urdiales** - *Programador* - [Serge](#)
* **Lawrence Rider García** - *Programador* - [Larry](http://www.larryrider.es)

See also the list of [contributors](https://github.com/Sjgu1/Gintonico/contributors) who participated in this project.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details

## Acknowledgments

* Hat tip to anyone who's code was used
* Inspiration
* etc
