package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/grounded042/hike_map_translator/garmin"
	"github.com/grounded042/hike_map_translator/models"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

// uuid "github.com/satori/go.uuid"

var source string
var sourceURL string

const timestampFormat = "01/02/2006"
const tripsFolder = "trips/"
const tripDetailsFolder = "details/"

func main() {

	rootCmd := &cobra.Command{
		Use:   "hike_map_translator",
		Short: "hike_map_translator translates data into a format for hike_map",
	}

	rootCmd.Flags().StringVarP(&source, "source", "s", "", "The source type to translate data from")
	rootCmd.Flags().StringVarP(&sourceURL, "source_url", "u", "", "The url used to pull the data to translate")

	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		translate(source, sourceURL)
	}

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(2)
	}
}

func translate(tSource, url string) {
	var gFrom generateFrom
	var err error

	switch strings.ToLower(tSource) {
	case "garmin":
		gFrom, err = garmin.LoadURL(url)
	default:
		panic(fmt.Sprintf("unknown source %v", source))
	}

	if err != nil {
		panic(err)
	}

	generateJSON(gFrom)
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
		index[dayNum].ID = uuid.Must(uuid.NewV4())
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
	index[0].ID = uuid.Must(uuid.NewV4())
	index[0].Label = dayName
	index[0].DetailsLocation = detailsFilePath

	// write the index
	iJSON, _ := json.MarshalIndent(index, "", "	")
	indexFilePath := tripsFolder + "index.json"
	ioutil.WriteFile(indexFilePath, iJSON, 0644)

}
