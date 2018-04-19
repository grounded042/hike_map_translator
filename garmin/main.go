package garmin

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/grounded042/hike_map_translator/util"

	"github.com/grounded042/hike_map_translator/models"
)

// Kml is the main struct to hold data from a garmin generated kml file
type Kml struct {
	XMLName  xml.Name `xml:"kml"`
	Document Document `xml:"Document"`
}

// Document corresponds to the Document kml element
type Document struct {
	XMLName xml.Name `xml:"Document"`
	Name    string   `xml:"name"`
	Folder  Folder   `xml:"Folder"`
}

// Folder corresponds to the Folder kml element
type Folder struct {
	XMLName    xml.Name    `xml:"Folder"`
	Name       string      `xml:"name"`
	Placemarks []Placemark `xml:"Placemark"`
}

// Placemark corresponds to the Placemark kml element
type Placemark struct {
	XMLName      xml.Name         `xml:"Placemark"`
	Name         string           `xml:"name"`
	TimeStamp    time.Time        `xml:"TimeStamp>when"`
	ExtendedData ExtendedDataData `xml:"ExtendedData>Data"`
	Coordinates  string           `xml:"Point>coordinates"`
}

// Data corresponds to the ExtendedData Data kml element
type Data struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value"`
}

// ExtendedDataData is a slice of ExtendedData Data structs
type ExtendedDataData []Data

// GetLatitude gets the latitude from a slice of ExtendedData Data structs
func (edd ExtendedDataData) GetLatitude() float64 {
	for _, data := range edd {
		if data.Name == "Latitude" {
			f, err := strconv.ParseFloat(data.Value, 64)

			if err != nil {
				return 0
			}

			return f
		}
	}

	return 0
}

// GetLongitude gets the longitude from a slice of ExtendedData Data structs
func (edd ExtendedDataData) GetLongitude() float64 {
	for _, data := range edd {
		if data.Name == "Longitude" {
			f, err := strconv.ParseFloat(data.Value, 64)

			if err != nil {
				return 0
			}

			return f
		}
	}

	return 0
}

// GetID gets the id from a slice of ExtendedData Data structs
func (edd ExtendedDataData) GetID() string {
	for _, data := range edd {
		if data.Name == "Id" {
			return data.Value
		}
	}

	return ""
}

// LoadFile loads a file from <filepath> and returns the Kml object with the
// data from the file
func LoadFile(filepath string) (*Kml, error) {
	kmlFile, err := os.Open(filepath)

	if err != nil {
		return nil, err
	}

	defer kmlFile.Close()

	byteValue, _ := ioutil.ReadAll(kmlFile)

	return LoadSliceOfBytes(byteValue), nil
}

// LoadURL loads data from the specified URL and returns the Kml object with
// the data
func LoadURL(url string) (*Kml, error) {
	feed, err := util.GetURLBody(url)

	if err != nil {
		return nil, err
	}

	return LoadSliceOfBytes(feed), nil
}

// LoadSliceOfBytes loads Garmin KML data from a slice of bytes
func LoadSliceOfBytes(bytes []byte) *Kml {
	var kml Kml
	xml.Unmarshal(bytes, &kml)

	return &kml
}

// GetAllPoints gets all the points from the kml data and satisfies the
// generateFrom interface
func (k *Kml) GetAllPoints() []models.Point {
	// we ignore the last placemark since it does not include extended data and
	// is more of a summary
	size := len(k.Document.Folder.Placemarks) - 1
	points := make([]models.Point, size, size)

	for i := 0; i < size; i++ {
		points[i] = models.Point{
			ID:        k.Document.Folder.Placemarks[i].ExtendedData.GetID(),
			Lat:       k.Document.Folder.Placemarks[i].ExtendedData.GetLatitude(),
			Lon:       k.Document.Folder.Placemarks[i].ExtendedData.GetLongitude(),
			Timestamp: k.Document.Folder.Placemarks[i].TimeStamp,
		}
	}

	return points
}

// GetAllPointsStartingAtDate gets all the points from the kml data as long as
// they are after the passed in time.
func (k *Kml) GetAllPointsStartingAtDate(startDate time.Time) []models.Point {
	points := []models.Point{}

	// we ignore the last placemark since it does not include extended data and
	// is more of a summary
	fmt.Println(len(k.Document.Folder.Placemarks))
	for i := 0; i < len(k.Document.Folder.Placemarks)-1; i++ {
		if k.Document.Folder.Placemarks[i].TimeStamp.After(startDate) {
			points = append(points, models.Point{
				ID:        k.Document.Folder.Placemarks[i].ExtendedData.GetID(),
				Lat:       k.Document.Folder.Placemarks[i].ExtendedData.GetLatitude(),
				Lon:       k.Document.Folder.Placemarks[i].ExtendedData.GetLongitude(),
				Timestamp: k.Document.Folder.Placemarks[i].TimeStamp,
			})
		}
	}

	return points
}
