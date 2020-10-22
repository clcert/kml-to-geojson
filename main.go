package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

type KML struct {
	XMLName  xml.Name `xml:"kml"`
	Document struct {
		Folder struct {
			Placemarks []struct {
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
			} `xml:"Placemark"`
		} `xml:"Folder"`
	} `xml:"Document"`
}

type Coord struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (c *Coord) Dist(c2 *Coord) float64 {
	return math.Sqrt(math.Abs(c.X-c2.X) + math.Abs(c.Y-c2.Y))
}

type Loc struct {
	ID     int     `json:"id"`
	Center Coord   `json:"center"`
	Radius float64 `json:"radius"`
}

func main() {
	xmlFile, err := os.Open("Microdatos_Censo_2017 _Manzana.kml")
	if err != nil {
		fmt.Println(err)
	}
	defer xmlFile.Close()
	var doc KML
	log.Printf("Loading XML file...")
	err = xml.NewDecoder(xmlFile).Decode(&doc)
	if err != nil {
		panic(err)
	}
	log.Printf("XML File loaded!")
	locations := make([]*Loc, 0)
	for i, placemark := range doc.Document.Folder.Placemarks {
		if i%1000 == 0 {
			log.Printf("[%d/%d] coords processed...", i, len(doc.Document.Folder.Placemarks))
		}

		loc := &Loc{}
		for _, data := range placemark.ExtendedData.SchemaData.Data {
			if data.Name == "MANZENT" {
				id, err := strconv.Atoi(data.Text)
				if err != nil {
					log.Printf("Error reading id: %s", err)
				} else {
					loc.ID = id
					break
				}
			}
		}
		coords := make([]*Coord, 0)
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
			x, err := strconv.ParseFloat(coordStr[0], 64)
			if err != nil {
				log.Printf("cannot parse coord %s as float: %s", coordStr[0], err)
			}
			y, err := strconv.ParseFloat(coordStr[1], 64)
			if err != nil {
				log.Printf("cannot parse coord %s as float: %s", coordStr[1], err)
			}
			coords = append(coords, &Coord{
				x,
				y,
			})
			loc.Center.X += x
			loc.Center.Y += y
		}
		loc.Center.X /= float64(len(coords))
		loc.Center.Y /= float64(len(coords))
		for _, coord := range coords {
			loc.Radius = math.Max(loc.Radius, loc.Center.Dist(coord))
		}
		locations = append(locations, loc)
	}
	outFile, err := os.Create("location.json")
	if err != nil {
		log.Printf("cannot create out file: %s", err)
		os.Exit(1)
	}
	defer outFile.Close()
	log.Printf("Marshaling to JSON")
	locsStr, err := json.MarshalIndent(&locations, " ", " ")
	if err != nil {
		log.Printf("cannot marshal locations: %s", err)
		os.Exit(1)
	}
	log.Printf("Done!")
	outFile.Write(locsStr)
}
