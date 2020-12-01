package main

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

var errNoSelecc = fmt.Errorf("Manzana no seleccionada")

func leerManzanasSeleccionadas(r *csv.Reader) map[uint32][]uint32 {
	r.Read() // Ignorando primera fila.
	manzents := make(map[uint32][]uint32)
	var i uint32 = 1
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("⚠️ No se pudo leer la fila %d: %s", i, err)
			continue
		}
		if len(record) != 1 {
			log.Printf("⚠️ La fila %d tiene un largo distinto al esperado (%d)", i, 1)
			continue
		}
		id, err := strconv.ParseUint(record[0], 10, 32)
		if err != nil {
			log.Printf("⚠️ No se pudo transformar la fila del CSV en un entero: %v", err)
			continue
		}
		_, ok := manzents[uint32(id)]
		if !ok {
			manzents[uint32(id)] = make([]uint32, 0)
		}
		manzents[uint32(id)] = append(manzents[uint32(id)], i)
		i++
	}
	return manzents
}

func posicionarKML(dec *xml.Decoder) error {
	for {
		t, err := dec.Token()
		if err != nil {
			return err
		}
		switch s := t.(type) {
		case xml.StartElement:
			if s.Name.Local == "name" {
				dec.Skip()
				return nil
			}
		}
	}
}

func procesarManzana(placemark Placemark, fid map[uint32][]uint32) (*Manzana, error) {
	manzana := &Manzana{}
	for _, data := range placemark.ExtendedData.SchemaData.Data {
		switch data.Name {
		case "FID":
			id, err := strconv.ParseUint(data.Text, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("no se puede transformar en entero el ID de la manzana: %s", err)
			}
			fids, ok := fid[uint32(id)]
			if !ok {
				return nil, errNoSelecc
			}
			manzana.ID = id
			manzana.Seleccion = fids
		case "REGION":
			manzana.Region = data.Text
		case "PROVINCIA":
			manzana.Provincia = data.Text
		case "COMUNA":
			manzana.Comuna = data.Text
		case "NOMBRE_DISTRITO":
			manzana.Distrito = data.Text
		case "TOTAL_VIVIENDAS":
			viviendas, err := strconv.ParseUint(data.Text, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("no se puede transformar en entero el número total de viviendas en la manzana: %s", err)
			}
			manzana.Viviendas = uint32(viviendas)
		case "TOTAL_PERSONAS":
			habitantes, err := strconv.ParseUint(data.Text, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("no se puede transformar en entero el número total de personas: %s", err)
			}
			manzana.Habitantes = uint32(habitantes)
		}
	}
	poli, err := procesarPoligono(placemark)
	if err != nil {
		return nil, fmt.Errorf("error en manzana %d: %s", manzana.ID, err)
	}
	manzana.Poligono = poli
	return manzana, nil
}

func procesarPoligono(placemark Placemark) ([][2]float32, error) {
	coordenadas := make([][2]float32, 0)
	coordsStr := strings.Split(placemark.MultiGeometry.Polygon.OuterBoundaryIs.LinearRing.Coordinates, " ")
	if len(coordsStr) == 0 {
		return nil, fmt.Errorf("no hay ninguna coordenada en esta manzana")
	}
	for _, coord := range coordsStr {
		coordStr := strings.Split(coord, ",")
		if len(coordStr) != 2 {
			return nil, fmt.Errorf("coordenada mal formada (su valor es %s)", coordStr)
		}
		x, err := strconv.ParseFloat(coordStr[0], 32)
		if err != nil {
			return nil, fmt.Errorf("no es posible parsear la primera coordenada como un número Float: %s", err)
		}
		y, err := strconv.ParseFloat(coordStr[1], 32)
		if err != nil {
			return nil, fmt.Errorf("no es posible parsear la segunda coordenada como un número Float: %s", err)
		}
		coordenadas = append(coordenadas, [2]float32{float32(y), float32(x)})
	}
	return coordenadas, nil
}
