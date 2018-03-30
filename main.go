package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/grounded042/hike_map_translator/garmin"
	"github.com/grounded042/hike_map_translator/models"
)

const timestampFormat = "01/02/2006"

const tripsFolder = "trips/"
const tripDetailsFolder = "details/"

func main() {
	url := os.Args[1]
	feed, err := getURL(url)

	if err != nil {
		panic(err)
	}

	kml := garmin.LoadSliceOfBytes(feed)

	generateJSON(kml)
}

type generateFrom interface {
	GetAllPoints() []models.Point
}

func generateJSON(gFrom generateFrom) {
	points := gFrom.GetAllPoints()

	// sort
	sort.Sort(models.ByTimestamp(points))

	// get each day
	lastDay := points[0].Timestamp.Format(timestampFormat)
	dayStartingIndex := 0
	days := map[string][]models.Point{}
	dayKeys := []string{}
	for i := 0; i < len(points); i++ {
		curDay := points[i].Timestamp.Format(timestampFormat)

		if curDay != lastDay {
			days[lastDay] = points[dayStartingIndex : i-1]
			dayKeys = append(dayKeys, lastDay)
			lastDay = curDay
			dayStartingIndex = i
		}
	}
	days[lastDay] = points[dayStartingIndex:len(points)]
	dayKeys = append(dayKeys, lastDay)

	os.MkdirAll(tripsFolder+tripDetailsFolder, os.ModePerm)

	cumulativeCoords := [][]float64{}

	index := make([]models.IndexDay, len(dayKeys)+1)

	for i, key := range dayKeys {
		dayNum := i + 1
		day := models.DayFromSliceOfPoints(key, dayNum, days[key])
		cumulativeCoords = append(cumulativeCoords, day.Coordinates...)
		dJSON, _ := json.MarshalIndent(day, "", "	")
		dayName := "Day " + strconv.Itoa(dayNum)
		detailsFilePath := tripDetailsFolder + dayName + ".json"
		ioutil.WriteFile(tripsFolder+detailsFilePath, dJSON, 0644)

		index[dayNum].Index = dayNum
		index[dayNum].Label = dayName
		index[dayNum].SubLabel = key
		index[dayNum].DetailsLocation = detailsFilePath
	}

	// write the all day
	allDay := models.DayFromCoords("All", 0, cumulativeCoords)
	dJSON, _ := json.MarshalIndent(allDay, "", "	")
	dayName := "All"
	detailsFilePath := tripDetailsFolder + dayName + ".json"
	ioutil.WriteFile(tripsFolder+detailsFilePath, dJSON, 0644)

	index[0].Index = 0
	index[0].Label = dayName
	index[0].DetailsLocation = detailsFilePath

	// write the index
	iJSON, _ := json.MarshalIndent(index, "", "	")
	indexFilePath := tripsFolder + "index.json"
	ioutil.WriteFile(indexFilePath, iJSON, 0644)

}

func getURL(url string) ([]byte, error) {
	resp, err := http.Get(url)

	if err != nil {
		return nil, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Read body: %v", err)
	}

	return data, nil
}
