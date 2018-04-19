package models

import (
	"sort"
	"time"

	"github.com/satori/go.uuid"
)

// Point represents a point on the Earth
type Point struct {
	ID        string
	Lat       float64   // latitude
	Lon       float64   // longitude
	Alt       float64   // altitude in meters
	Timestamp time.Time // time this point was tracked
}

// GetCoordinates gets a slice of the longitude and latitude coordinates from a
// point
func (p *Point) GetCoordinates() []float64 {
	return []float64{p.Lon, p.Lat}
}

// ByTimestamp implements sort.Interface to sort a slice of Points
type ByTimestamp []Point

func (bt ByTimestamp) Len() int {
	return len(bt)
}

func (bt ByTimestamp) Less(i, j int) bool {
	return bt[i].Timestamp.Before(bt[j].Timestamp)
}

func (bt ByTimestamp) Swap(i, j int) {
	bt[i], bt[j] = bt[j], bt[i]
}

// Day represents all the info needed for a day of hiking data
type Day struct {
	Index       int         `json:"index"`
	Name        string      `json:"name"`
	Coordinates [][]float64 `json:"coordinates"`
	Bounds      Bounds      `json:"outerPoints"`
}

// DayFromSliceOfPoints builds a day object from a slice of points
func DayFromSliceOfPoints(name string, index int, points []Point) *Day {
	pointsLen := len(points)
	coordinates := make([][]float64, pointsLen)
	pointCoords := make([]float64, pointsLen*2)

	for i := range coordinates {
		l := i * 2
		pointCoords[l] = points[i].Lon
		pointCoords[l+1] = points[i].Lat
		coordinates[i] = pointCoords[l : l+2]
	}

	return DayFromCoords(name, index, coordinates)
}

// DayFromCoords builds a day object from coords
func DayFromCoords(name string, index int, coords [][]float64) *Day {
	day := Day{
		Index:       index,
		Name:        name,
		Coordinates: coords,
		Bounds:      getBoundaryCoords(coords),
	}

	return &day
}

// Given a slice of coordinates, this function finds the most logical northern,
// southern, eastern, and western most coordinates in that list.
func getBoundaryCoords(coords [][]float64) Bounds {
	lon := make([]float64, len(coords))
	lat := make([]float64, len(coords))

	for i := range coords {
		lon[i] = coords[i][0]
		lat[i] = coords[i][1]
	}

	west, east := getWestAndEastCoords(lon)

	sort.Float64s(lat)

	return Bounds{
		North: lat[len(lat)-1],
		South: lat[0],
		East:  east,
		West:  west,
	}
}

// Given a slice of longitude coordinates, this func finds the most logical
// western and eastern most coordinates in that slice.
func getWestAndEastCoords(longitude []float64) (float64, float64) {
	sort.Float64s(longitude)

	var lowest = longitude[0]
	var highest = longitude[len(longitude)-1]

	if highest-lowest > 180 {
		return highest, lowest
	}

	return lowest, highest
}

// Bounds holds North, South, East, West outer coordinates
type Bounds struct {
	North float64 `json:"north"`
	South float64 `json:"south"`
	East  float64 `json:"east"`
	West  float64 `json:"west"`
}

// IndexDay holds the index details for a day
type IndexDay struct {
	Index           int       `json:"index"`
	ID              uuid.UUID `json:"id"`
	Label           string    `json:"label"`
	SubLabel        string    `json:"subLabel"`
	DetailsLocation string    `json:"detailsLocation"`
}

// NewIndexDay creates and returns a new index day
func NewIndexDay(dayNum int, label, subLabel, detailsLocation string) IndexDay {
	return IndexDay{
		Index:           dayNum,
		ID:              uuid.Must(uuid.NewV4()),
		Label:           label,
		SubLabel:        subLabel,
		DetailsLocation: detailsLocation,
	}
}

// Index is the type that holds an index entry for each day
type Index []IndexDay
