# Generador de datos para el mapa de Lxs400

Este script transforma de forma eficiente una lista de manzanas seleccionadas en un archivo JSON mostrable en la página de "Los 400"

## Instrucciones

* Descargar base de datos con las manzanas desde [acá](https://geoine-ine-chile.opendata.arcgis.com/datasets/54e0c40680054efaabeb9d53b09e1e7a_0/data) (Download -> KML File). No cambiar el nombre del archivo.
* Descargar archivo de texto con las manzanas sorteadas en orden, una por línea y colocar en la misma carpeta.
* Clonar este repositorio, compilar y ejecutar programa: `lxs400-datos-mapa manzanas.csv 16000000 Microdatos_Censo_2017\ _Manzana.kml manzanas.json`:
    * `16000000`: Pulso usado para generar la selección de manzanas.
    * `manzanas.csv`: Archivo de entrada: Lista de manzanas seleccionadas, una por línea.
    * `Microdatos_Censo_2017 _Manzana.kml`: KML con datos de manzanas.
    * `manzanas.json`: Archivo de salida: JSON con coordenadas de manzanas, organizadas de forma geográfica.