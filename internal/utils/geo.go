package utils

import (
	"math"
	"time"

	"github.com/4wings/cli/types"
	log "github.com/sirupsen/logrus"
)

const (
	MIN_LAT = -85.051128
	MAX_LAT = 85.051128
	MIN_LON = -180
	MAX_LON = 180
)

func clip(n float64, minValue float64, maxValue float64) float64 {
	return math.Min(math.Max(n, minValue), maxValue)
}

func getTileNums(lat float64, lon float64, maxZoom float64) []int64 {

	nums := make([]int64, int(maxZoom+1))

	for i := 1; i <= int(maxZoom); i++ {

		tile := FromCoordinate(lat, lon, float64(i))
		pos := int64(tile.X) + (int64(tile.Y) * int64(math.Pow(2, float64(i))))
		nums[i] = pos
	}
	return nums
}

func GetCellsByPos(pos int, z int) int {
	cellSizeLat := cellAreas[z]
	cellSizeLon := cellAreas[z]
	y := int(math.Floor(float64(pos) / math.Pow(2, float64(z))))
	x := int(pos) % int(math.Pow(2, float64(z)))
	bounds := TileToBBOX(x, y, z)
	numCellsLat := int(math.Ceil((bounds.MaxLat - bounds.MinLat) / cellSizeLat))
	numCellsLon := int(math.Ceil((bounds.MaxLon - bounds.MinLon) / cellSizeLon))
	return numCellsLat * numCellsLon
}

func GetCellByDatasetRowAndColumn(z, x, y int, cellColumn, cellRow int) int {
	cellSizeLon := cellAreas[z]
	bounds := TileToBBOX(x, y, z)
	numCellsLon := int(math.Ceil((bounds.MaxLon - bounds.MinLon) / cellSizeLon))
	return numCellsLon*cellRow + cellColumn
}

func GetCellsLatLonByPos(pos int, z int) (int, int) {
	cellSizeLat := cellAreas[z]
	cellSizeLon := cellAreas[z]
	y := int(math.Floor(float64(pos) / math.Pow(2, float64(z))))
	x := int(pos) % int(math.Pow(2, float64(z)))
	bounds := TileToBBOX(x, y, z)
	numCellsLat := int(math.Ceil((bounds.MaxLat - bounds.MinLat) / cellSizeLat))
	numCellsLon := int(math.Ceil((bounds.MaxLon - bounds.MinLon) / cellSizeLon))
	return numCellsLat, numCellsLon
}

func GetLatLonByTileAndCell(cell, x, y, z int) Point {
	cellSize := cellAreas[z]

	bounds := TileToBBOX(x, y, z)
	numCellsLat := math.Ceil((bounds.MaxLat - bounds.MinLat) / cellSize)
	numCellsLon := math.Ceil((bounds.MaxLon - bounds.MinLon) / cellSize)
	deltaLat := (bounds.MaxLat - bounds.MinLat) / numCellsLat
	deltaLon := (bounds.MaxLon - bounds.MinLon) / numCellsLon

	x1 := cell % int(numCellsLon)
	y1 := cell / int(numCellsLon)
	lat := float64(y1)*deltaLat + bounds.MinLat
	lon := float64(x1)*deltaLon + bounds.MinLon
	return Point{Lat: lat, Lon: lon}
}

func GetCell(lat float64, lon float64, zoom int, gridCellAreas []float64) int {
	//cellDefinition := []float64{2.8407093, 1.4203547, 0.898311, 0.449156, 0.200868, 0.089833, 0.04492, 0.02008, 0.00898, 0.00492025295, 0.0028407093, 0.0014203547, 0.0006352019, 0.00028407093, 0.000200868, 0.000089833}
	cellSizeLat := gridCellAreas[zoom] /// math.Pow(2, float64(zoom-1))
	cellSizeLon := gridCellAreas[zoom] /// math.Pow(2, float64(zoom-1))

	tile := FromCoordinate(lat, lon, float64(zoom))
	bounds := TileToBBOX(tile.X, tile.Y, zoom)
	numCellsLat := math.Ceil((bounds.MaxLat - bounds.MinLat) / cellSizeLat)
	numCellsLon := math.Ceil((bounds.MaxLon - bounds.MinLon) / cellSizeLon)
	deltaLat := (bounds.MaxLat - bounds.MinLat) / numCellsLat
	deltaLon := (bounds.MaxLon - bounds.MinLon) / numCellsLon

	y1 := math.Floor((lat - bounds.MinLat) / deltaLat)
	x1 := math.Floor((lon - bounds.MinLon) / deltaLon)

	return int(x1) + int(y1)*int(numCellsLon)
}

func getCells(maxZoom int, lat float64, lon float64) []int {
	cells := make([]int, maxZoom+1)
	for i := 0; i <= maxZoom; i++ {
		cell := GetCell(lat, lon, i, cellAreas)
		if cell == -1 {
			log.Debugf("lat: %f, lon: %f, zoom: %d, numCells: %d  -> %d \n", lat, lon, i, cellAreas, cell)
		}
		cells[i] = cell
	}
	return cells
}
func SanitizeRow(row map[string]interface{}, resolution string) (*types.Row, error) {
	var lat, lon, value float64
	if val, ok := row["lat"].(float64); ok {
		lat = val
	} else if val, ok := row["lat"].(int64); ok {
		lat = float64(val)
	}
	if valLon, ok := row["lon"].(float64); ok {
		lon = valLon
	} else if valLon, ok := row["lon"].(int64); ok {
		lon = float64(valLon)
	}

	if valValue, ok := row["value"].(float64); ok {
		value = valValue
	} else if valValue, ok := row["value"].(int64); ok {
		value = float64(valValue)
	}
	lat = clip(lat, MIN_LAT, MAX_LAT)

	lon = clip(lon, MIN_LON, MAX_LON)

	timestamp := row["timestamp"].(time.Time)

	truncateValue := 1 * time.Hour
	if resolution == "day" {
		truncateValue = 24 * time.Hour
	}
	timestamp = timestamp.Truncate(truncateValue)
	htime := timestamp.Unix() / int64(truncateValue.Seconds())

	nums := getTileNums(lat, lon, 12)

	cells := getCells(12, lat, lon)
	position := getTilePosition(lon, lat, 12)
	data := types.Row{
		Lat:       lat,
		Lon:       lon,
		Timestamp: timestamp,
		HTime:     htime,
		Positions: nums,
		Cells:     cells,
		Position:  position,
		Value:     value,
		Others:    row,
	}

	return &data, nil

}
