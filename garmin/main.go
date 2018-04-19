package garmin

import (
	"encoding/xml"
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

// ToPoint converts a Placemark to a models.Point
func (p *Placemark) ToPoint() models.Point {
	return models.Point{
		ID:        p.ExtendedData.GetID(),
		Lat:       p.ExtendedData.GetLatitude(),
		Lon:       p.ExtendedData.GetLongitude(),
		Timestamp: p.TimeStamp,
	}
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
func (k *Kml) GetAllPoints(startDate, endDate time.Time) []models.Point {
	if !startDate.IsZero() || !endDate.IsZero() {
		return k.getAllPointsFiltered(startDate, endDate)
	}

	// we ignore the last placemark since it does not include extended data and
	// is more of a summary
	size := len(k.Document.Folder.Placemarks) - 1
	points := make([]models.Point, size, size)

	for i := 0; i < size; i++ {
		points[i] = k.Document.Folder.Placemarks[i].ToPoint()
	}

	return points
}

// getAllPointsFiltered gets all the points from the kml data filtered using
// the passed in dates
func (k *Kml) getAllPointsFiltered(startDate, endDate time.Time) []models.Point {
	points := []models.Point{}

	filterFunc := getDateFilterFunc(startDate, endDate)

	// we ignore the last placemark since it does not include extended data and
	// is more of a summary
	for i := 0; i < len(k.Document.Folder.Placemarks)-1; i++ {
		if filterFunc(k.Document.Folder.Placemarks[i].TimeStamp) {
			points = append(points, k.Document.Folder.Placemarks[i].ToPoint())
		}
	}

	return points
}

func getDateFilterFunc(startDate, endDate time.Time) func(time.Time) bool {
	if startDate.IsZero() && endDate.IsZero() {
		return func(t time.Time) bool {
			return true
		}
	} else if !startDate.IsZero() && endDate.IsZero() {
		return func(t time.Time) bool {
			return t.After(startDate)
		}
	} else if startDate.IsZero() && !endDate.IsZero() {
		return func(t time.Time) bool {
			return t.Before(startDate)
		}
	}

	return func(t time.Time) bool {
		return t.After(startDate) && t.Before(endDate)
	}
}
