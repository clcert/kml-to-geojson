package main

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args[1:]) != 4 {
		log.Printf("⛔️ Cantidad de argumentos no coincide con los necesarios:")
		log.Printf("%s <fecha-pulso> <manzanas_seleccionadas> <kml> <json_salida>", os.Args[0])
		os.Exit(1)
	}

	log.Printf("🗺 Cargando archivo KML")
	xmlFile, err := os.Open(os.Args[3])
	if err != nil {
		fmt.Println(err)
	}
	defer xmlFile.Close()
	kmlDecoder := xml.NewDecoder(xmlFile)

	log.Printf("🍎 Cargando archivo con manzanas seleccionadas")
	selectedFile, err := os.Open(os.Args[2])
	if err != nil {
		log.Printf("⛔️ No se pudo abrir archivo con manzanas: %s", err)
		os.Exit(1)
	}
	defer selectedFile.Close()
	selectedCSV := csv.NewReader(selectedFile)

	log.Printf("📔 Creando archivo de salida")
	jsonSalida, err := os.Create(os.Args[4])
	if err != nil {
		log.Printf("⛔️ No se pudo crear archivo de salida: %s", err)
		os.Exit(1)
	}
	defer jsonSalida.Close()

	fids := leerManzanasSeleccionadas(selectedCSV)

	datos := Datos{
		Fecha:     os.Args[1],
		Regiones:  make(Regiones),
		Viviendas: make(map[string]uint32),
	}
	posicionarKML(kmlDecoder)
	i := 0
	for {
		placemark := Placemark{}
		err := kmlDecoder.Decode(&placemark)
		if err != nil {
			// TODO: diferenciar errores de fin de lista
			break
		}
		manzana, err := procesarManzana(placemark, fids)
		if err != nil {
			if err == errNoSelecc {
				// No alertar estas manzanas. Son muchas.
				continue
			}
			log.Printf("⛔️ Error procesando manzana: %s", err)
		}
		datos.agregar(manzana)
		if i%100 == 0 && i != 0 {
			log.Printf("🍎 %d manzanas procesadas", i)
		}
		i++
	}
	jsonCodificador := json.NewEncoder(jsonSalida)
	jsonCodificador.Encode(&datos)
	log.Printf("✅ Todo listo. El archivo de salida fue copiado a %s", os.Args[4])
}
