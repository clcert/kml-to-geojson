package main

import "encoding/xml"

// Datos representa la estructura de salida de este programa.
type Datos struct {
	Timestamp string            `json:"ts"` // Incluye el timestamp en formato Unix Epoch del pulso usado
	Pulso     int               `json:"p"`  // ID del pulso usado
	Cadena    int               `json:"c"`  // ID de la cadena usada
	URI       string            `json:"u"`  // URI al pulso
	Viviendas map[string]uint32 `json:"v"`  // Incluye datos estadísticos sobre las viviendas elegidas
	Regiones  Regiones          `json:"r"`  // Agrupa todas las regiones que tienen manzanas seleccionadas.
}

func (d *Datos) agregar(m *Manzana) {
	if _, ok := d.Viviendas[m.Region]; !ok {
		d.Viviendas[m.Region] = 0
	}
	d.Viviendas[m.Region] += uint32(m.Viviendas)

	mapaRegion, ok := d.Regiones[m.Region]
	if !ok {
		mapaRegion = make(Provincias)
		d.Regiones[m.Region] = mapaRegion
	}
	mapaProvincia, ok := mapaRegion[m.Provincia]
	if !ok {
		mapaProvincia = make(Comunas)
		mapaRegion[m.Provincia] = mapaProvincia
	}
	mapaComuna, ok := mapaProvincia[m.Comuna]
	if !ok {
		mapaComuna = make(Distritos)
		mapaProvincia[m.Comuna] = mapaComuna
	}
	if _, ok = mapaComuna[m.Distrito]; !ok {
		mapaComuna[m.Distrito] = make([]*Manzana, 0)
	}
	mapaComuna[m.Distrito] = append(mapaComuna[m.Distrito], m)
}

// Regiones representa un mapa de provincias
type Regiones map[string]Provincias

// Provincias es un mapa de comunas
type Provincias map[string]Comunas

// Comunas es un mapa de distritos
type Comunas map[string]Distritos

// Distritos es un mapa con listas de manzanas.
type Distritos map[string][]*Manzana

// Manzana representa una cuadra según la información del KML del censo.
type Manzana struct {
	ID         uint32       `json:"i"` // equivale a MANZENT en el KML
	Seleccion  []uint32     `json:"s"` // Lista de posiciones en las cuales salió sorteada esta manzana.
	Poligono   [][2]float32 `json:"p"` // Lista de coordenadas que define la forma de la manzana.
	Viviendas  uint32       `json:"v"` // Número de viviendas en la manzana
	Habitantes uint32       `json:"h"` // Número de habitantes en la manzana.
	Region     string       `json:"-"` // Nombre de la región en que se ubica la manzana. No se exporta.
	Provincia  string       `json:"-"` // Nombre de la provincia en que se ubica la manzana. No se exporta.
	Comuna     string       `json:"-"` // Nombre de la comuna en que se ubica la manzana. No se exporta.
	Distrito   string       `json:"-"` // Nombre del distrito en que se ubica la manzana. No se exporta.
}

// Placemark agrupa la estructura XML de cada manzana.
type Placemark struct {
	ExtendedData struct {
		SchemaData struct {
			SchemaURL string `xml:"schemaURL,attr"`
			Data      []struct {
				Name string `xml:"name,attr"`
				Text string `xml:",chardata"`
			} `xml:"SimpleData"`
		} `xml:"SchemaData"`
	} `xml:"ExtendedData"`
	MultiGeometry struct {
		Polygon struct {
			OuterBoundaryIs struct {
				LinearRing struct {
					Coordinates string `xml:"coordinates"`
				} `xml:"LinearRing"`
			} `xml:"outerBoundaryIs"`
		} `xml:"Polygon"`
	} `xml:"MultiGeometry"`
}

// KML representa la estructura del archivo KML a revisar. Es necesario para
// poder desestructurar este archivo, aunque la mayoría de los campos no
// se usan.
type KML struct {
	XMLName  xml.Name `xml:"kml"`
	Document struct {
		Folder struct {
			Placemarks []Placemark `xml:"Placemark"`
		} `xml:"Folder"`
	} `xml:"Document"`
}

// PulseResponse extrae algunos campos importantes de una respuesta de la API del beacon.
type PulseResponse struct {
	Pulse Pulse `json:"pulse"`
}

type Pulse struct {
	Timestamp string `json:"timeStamp"`
	URI       string `json:"uri"`
	Pulse     int    `json:"pulseIndex"`
	Chain     int    `json:"chainIndex"`
}
