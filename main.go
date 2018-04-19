package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/grounded042/hike_map_translator/garmin"
	"github.com/grounded042/hike_map_translator/models"
	"github.com/spf13/cobra"
)

var source string
var sourceURL string
var startDate string
var endDate string
var tripName string

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
	rootCmd.Flags().StringVarP(&startDate, "start_date", "d", "", "The day on which to start pulling data from in the format of <month>/<day>/<year> with leading zeros")
	rootCmd.Flags().StringVarP(&endDate, "end_date", "e", "", "The day on which to stop pulling data from in the format of <month>/<day>/<year> with leading zeros")
	rootCmd.Flags().StringVarP(&tripName, "trip_name", "n", "", "The name of the trip")

	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		timeStartDate := convertDate(startDate, "`start_date` was not supplied - will not filter based on a start date")
		timeEndDate := convertDate(endDate, "`end_date` was not supplied - will not filter based on a end date")

		if !timeEndDate.IsZero() {
			timeEndDate = timeEndDate.AddDate(0, 0, 1)
		}

		translate(source, sourceURL, timeStartDate, timeEndDate, tripName)
	}

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(2)
	}
}

func convertDate(date string, errMessage string) time.Time {
	if date == "" {
		fmt.Println(errMessage)
		return time.Time{}
	}

	t, err := time.Parse(timestampFormat, date)

	if err != nil {
		panic(err)
	}

	return getStartOfDayTime(t)
}

func getStartOfDayTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

// buildDaysFromPoints takes a slice of points and groups a slice of points
// into a day. It returns a map of the days with their corresponding slice of
// points and an ordered slice of map keys of days.
func buildDaysFromPoints(points []models.Point) (map[string][]models.Point, []string) {
	// sort the points so they are in order by timestamp
	sort.Sort(models.ByTimestamp(points))

	previousDay := points[0].Timestamp.Format(timestampFormat)
	dayStartingIndex := 0
	days := map[string][]models.Point{}
	dayKeys := []string{}
	for i := 0; i < len(points); i++ {
		curDay := points[i].Timestamp.Format(timestampFormat)

		if curDay != previousDay {
			days[previousDay] = points[dayStartingIndex : i-1]
			dayKeys = append(dayKeys, previousDay)
			previousDay = curDay
			dayStartingIndex = i
		}
	}
	days[previousDay] = points[dayStartingIndex:len(points)]
	dayKeys = append(dayKeys, previousDay)

	return days, dayKeys
}

// writeDayJSON writes the passed in day info to a JSON file and returns the
// location of the file
func writeDayJSON(day *models.Day, dayName string) string {
	dJSON, _ := json.MarshalIndent(day, "", "	")
	detailsFilePath := tripDetailsFolder + dayName + ".json"
	ioutil.WriteFile(tripsFolder+detailsFilePath, dJSON, 0644)

	return detailsFilePath
}

// generateJSON takes a map of days and their points along with the ordered
// slice of the map keys so the days can be properly ordered
func generateJSON(days map[string][]models.Point, dayKeys []string, tName string) {
	os.MkdirAll(tripsFolder+tripDetailsFolder, os.ModePerm)

	cumulativeCoords := [][]float64{}

	index := make([]models.IndexDay, len(dayKeys)+1)

	// create the individual days
	for i, key := range dayKeys {
		dayNum := i + 1
		day := models.DayFromSliceOfPoints(key, dayNum, days[key])
		cumulativeCoords = append(cumulativeCoords, day.Coordinates...)
		dayName := "Day " + strconv.Itoa(dayNum)
		detailsFilePath := writeDayJSON(day, dayName)

		index[dayNum] = models.NewIndexDay(dayNum, dayName, key, detailsFilePath)
	}

	// create the all day
	allDay := models.DayFromCoords("All", 0, cumulativeCoords)
	dayName := "All"
	detailsFilePath := writeDayJSON(allDay, dayName)

	index[0] = models.NewIndexDay(0, dayName, tName, detailsFilePath)

	// write the index
	iJSON, _ := json.MarshalIndent(index, "", "	")
	indexFilePath := tripsFolder + "index.json"
	ioutil.WriteFile(indexFilePath, iJSON, 0644)

}

type generateFrom interface {
	GetAllPoints(time.Time, time.Time) []models.Point
}

func translate(tSource, url string, sDate, eDate time.Time, tName string) {
	var gFrom generateFrom
	var err error

	switch strings.ToLower(tSource) {
	case "garmin":
		gFrom, err = garmin.LoadURL(url)
	default:
		panic(fmt.Sprintf("unknown source: %v", source))
	}

	if err != nil {
		panic(err)
	}

	points := gFrom.GetAllPoints(sDate, eDate)

	if len(points) == 0 {
		fmt.Println("No points to translate!")
		return
	}

	days, dayKeys := buildDaysFromPoints(points)

	generateJSON(days, dayKeys, tName)
}
