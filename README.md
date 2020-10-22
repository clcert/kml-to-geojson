# KML-to-LOC

Transforms the coordinates in Microdatos_Censo_2017_Manzana.kml to points and radiuses in a JSON.

## Instructions

* Download database of blocks from [here](https://geoine-ine-chile.opendata.arcgis.com/datasets/54e0c40680054efaabeb9d53b09e1e7a_0/data) (Download -> KML File) Do not change the name of the file.
* Open a terminal in the folder where the KML file is
* execute `go run github.com/clcert/kml-to-loc`. The parsed file will be on the same directory, with the name location.json.