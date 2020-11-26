package main

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type Datos struct {
	Viviendas map[string]uint32 `json:"v"`
	Regiones  Regiones          `json:"r"`
}
type Regiones map[string]Provincias
type Provincias map[string]Comunas
type Comunas map[string]Distritos
type Distritos map[string][]*Manzana

type Manzana struct {
	ID         uint32       `json:"i"`
	Seleccion  []uint32     `json:"s"`
	Poligono   [][2]float32 `json:"p"`
	Viviendas  uint32       `json:"v"`
	Habitantes uint32       `json:"h"`
}

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

type KML struct {
	XMLName  xml.Name `xml:"kml"`
	Document struct {
		Folder struct {
			Placemarks []Placemark `xml:"Placemark"`
		} `xml:"Folder"`
	} `xml:"Document"`
}

func main() {
	if len(os.Args[1:]) != 3 {
		log.Printf("%s <manzanas_seleccionadas> <kml> <json_salida>", os.Args[0])
		os.Exit(1)
	}
	xmlFile, err := os.Open(os.Args[2])
	if err != nil {
		fmt.Println(err)
	}
	defer xmlFile.Close()
	log.Printf("Loading XML file...")
	dec := xml.NewDecoder(xmlFile)
	out, err := os.Create(os.Args[3])
	if err != nil {
		log.Printf("cannot create out file: %s", err)
		os.Exit(1)
	}
	defer out.Close()

	selected, err := os.Open(os.Args[1])
	if err != nil {
		log.Printf("cannot create out file: %s", err)
		os.Exit(1)
	}
	defer selected.Close()
	csvSelected := csv.NewReader(selected)
	manzents := make(map[uint32][]uint32)
	csvSelected.Read() // CSV Header
	var i uint32 = 1
	for {
		record, err := csvSelected.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("cannot read row: %s", err)
			continue
		}
		if len(record) != 2 {
			log.Printf("row shorter than expected: %v", record)
			continue
		}
		id, err := strconv.ParseUint(record[1], 10, 32)
		if err != nil {
			log.Printf("cannot transform fid id to int: %v", err)
			continue
		}
		_, ok := manzents[uint32(id)]
		if !ok {
			manzents[uint32(id)] = make([]uint32, 0)
		}
		manzents[uint32(id)] = append(manzents[uint32(id)], i)
		i++
	}
	datos := Datos{
		Regiones:  make(Regiones),
		Viviendas: make(map[string]uint32),
	}
L:
	for {
		t, err := dec.Token()
		if err != nil {
			panic(err)
		}
		switch s := t.(type) {
		case xml.StartElement:
			if s.Name.Local == "name" {
				break L
			}
		}
	}
	dec.Skip()
	i = 0
	for {
		placemark := Placemark{}
		err := dec.Decode(&placemark)
		if err != nil {
			break
		}
		var region, provincia, comuna, distrito string
		var personas, viviendas, id uint64
		for _, data := range placemark.ExtendedData.SchemaData.Data {
			switch data.Name {
			case "FID":
				id, err = strconv.ParseUint(data.Text, 10, 32)
				if err != nil {
					log.Printf("Error reading id: %s", err)
				}
			case "REGION":
				region = data.Text
			case "PROVINCIA":
				provincia = data.Text
			case "COMUNA":
				comuna = data.Text
			case "NOMBRE_DISTRITO":
				distrito = data.Text
			case "TOTAL_VIVIENDAS":
				viviendas, err = strconv.ParseUint(data.Text, 10, 32)
				if err != nil {
					log.Printf("Error reading viviendas: %s", err)
				}
			case "TOTAL_PERSONAS":
				personas, err = strconv.ParseUint(data.Text, 10, 32)
				if err != nil {
					log.Printf("Error reading personas: %s", err)
				}
			}
		}
		if _, ok := datos.Viviendas[region]; !ok {
			datos.Viviendas[region] = 0
		}
		datos.Viviendas[region] += uint32(viviendas)
		_, ok := manzents[uint32(id)]
		if !ok {
			continue
		}
		if i%1000 == 0 {
			log.Printf("%d coords processed...", i)
		}
		loc := &Manzana{
			ID:         uint32(id),
			Habitantes: uint32(personas),
			Viviendas:  uint32(viviendas),
			Seleccion:  manzents[uint32(id)],
			Poligono:   make([][2]float32, 0),
		}
		coordsStr := strings.Split(placemark.MultiGeometry.Polygon.OuterBoundaryIs.LinearRing.Coordinates, " ")
		if len(coordsStr) == 0 {
			log.Printf("empty coords for id=%d", loc.ID)
			continue
		}
		for _, coord := range coordsStr {
			coordStr := strings.Split(coord, ",")
			if len(coordStr) != 2 {
				log.Printf("bad formed coord: id=%d coord=%s", loc.ID, coordStr)
				continue
			}
			x, err := strconv.ParseFloat(coordStr[0], 32)
			if err != nil {
				log.Printf("cannot parse coord %s as float: %s", coordStr[0], err)
			}
			y, err := strconv.ParseFloat(coordStr[1], 32)
			if err != nil {
				log.Printf("cannot parse coord %s as float: %s", coordStr[1], err)
			}
			loc.Poligono = append(loc.Poligono, [2]float32{float32(y), float32(x)})
		}
		reg, ok := datos.Regiones[region]
		if !ok {
			reg = make(Provincias)
			datos.Regiones[region] = reg
		}
		prov, ok := reg[provincia]
		if !ok {
			prov = make(Comunas)
			reg[provincia] = prov
		}
		com, ok := prov[comuna]
		if !ok {
			com = make(Distritos)
			prov[comuna] = com
		}
		if _, ok = com[distrito]; !ok {
			com[distrito] = make([]*Manzana, 0)
		}
		com[distrito] = append(com[distrito], loc)
		i++
	}
	enc := json.NewEncoder(out)
	enc.Encode(&datos)
	log.Printf("Done!")
}
