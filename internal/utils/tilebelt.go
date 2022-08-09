package utils

import (
	"fmt"
	"math"
	"strconv"

	"github.com/paulmach/orb"
)

var gridCellAreas []float64 = []float64{128000000000, 32000000000, 8000000000, 2000000000, 500000000, 125000000, 31250000, 7812500, 1953125, 488281, 122070, 30518, 7629}
var cellAreas []float64

const (
	D2R = math.Pi / 180
	R2D = 180 / math.Pi
)

type Bounds struct {
	MinLat float64
	MaxLat float64
	MinLon float64
	MaxLon float64
}

type Point struct {
	Lat float64
	Lon float64
}

type Tile struct {
	X int
	Y int
	Z int
}

func init() {
	for _, v := range gridCellAreas {
		cellAreas = append(cellAreas, math.Sqrt(v)/111320)
	}
}

func tile2lon(x int, z int) float64 {
	return float64(x)/math.Pow(2, float64(z))*360 - 180
}

func tile2lat(y int, z int) float64 {
	n := math.Pi - 2*math.Pi*float64(y)/math.Pow(2, float64(z))
	return R2D * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
}

func Lon2tile(lon float64, zoom float64) int {
	return int(math.Floor(((lon + 180) / 360) * math.Pow(2, zoom)))
}

func FromCoordinate(lat, lon, zoom float64) Tile {
	x := Lon2tile(lon, zoom)
	y := Lat2tile(lat, zoom)

	return Tile{
		X: x,
		Y: y,
		Z: int(zoom),
	}
}
func Lat2tile(lat float64, zoom float64) int {
	return int(math.Floor(
		((1 -
			math.Log(
				math.Tan((lat*math.Pi)/180)+1/math.Cos((lat*math.Pi)/180),
			)/
				math.Pi) /
			2) *
			math.Pow(2, zoom),
	))
}

func TileToBBOX(x int, y int, z int) Bounds {
	w := tile2lon(x, z)
	s := tile2lat(y+1, z)
	e := tile2lon(x+1, z)
	n := tile2lat(y, z)
	return Bounds{
		MinLat: s,
		MaxLat: n,
		MinLon: w,
		MaxLon: e,
	}
}

func TileToBBOXProj(x int, y int, z int, cellSize float64) Bounds {
	w := tile2lon(x, z)
	s := tile2lat(y+1, z)
	e := tile2lon(x+1, z)
	n := tile2lat(y, z)

	if math.Mod(w, cellSize) != 0 {
		w = w - math.Mod(w, cellSize)
	}
	if math.Mod(s, cellSize) != 0 {
		s = s - math.Mod(s, cellSize)
	}
	if math.Mod(e, cellSize) != 0 {
		e = e - math.Mod(e, cellSize)
	}
	if math.Mod(n, cellSize) != 0 {
		n = n - math.Mod(n, cellSize)
	}
	return Bounds{
		MinLat: s,
		MaxLat: n,
		MinLon: w,
		MaxLon: e,
	}
}

func Tile2Num(z, x, y int) (int64, error) {
	return int64(x) + int64(y)*int64(math.Pow(2, float64(z))), nil
}

func GetPointFromCell(cell, x, y, z, numCells int) Point {
	cellSize := 1 / math.Pow(2, float64(z-1))
	bounds := TileToBBOX(x, y, z)

	numCellsLat := math.Ceil((bounds.MaxLat - bounds.MinLat) / cellSize)
	numCellsLon := math.Ceil((bounds.MaxLon - bounds.MinLon) / cellSize)
	deltaLat := (bounds.MaxLat - bounds.MinLat) / numCellsLat
	deltaLon := (bounds.MaxLon - bounds.MinLon) / numCellsLon

	numCellsProj := (bounds.MaxLon - bounds.MinLon) / cellSize
	x1 := cell % int(numCellsProj)
	y1 := cell / int(numCellsProj)
	// deltaLat := (bounds.MaxLat - bounds.MinLat) / float64(numCells)
	// deltaLon := (bounds.MaxLon - bounds.MinLon) / float64(numCells)

	lat := bounds.MinLat + float64(y1)*deltaLat + deltaLat/2
	lon := bounds.MinLon + float64(x1)*deltaLon + deltaLon/2
	return Point{Lat: lat, Lon: lon}
}

func checkMaxLon(lon float64) float64 {
	const MIN_LON = -180
	const MAX_LON = 180
	if lon == 0 {
		return 0.00000001
	} else {
		return clip(lon, MIN_LON, MAX_LON)
	}
}

func checkMaxLat(lat float64) float64 {
	const MIN_LAT = -85.051128
	const MAX_LAT = 85.051128
	if lat == 0 {
		return 0.00000001
	} else {
		return clip(lat, MIN_LAT, MAX_LAT)
	}
}

func getTilePosition(lon float64, lat float64, maxZoom int) int64 {
	pos := ""
	newLon := checkMaxLon(lon)
	newLat := checkMaxLat(lat)
	for i := 0; i <= maxZoom; i++ {
		x := Lon2tile(newLon, float64(i))
		y := Lat2tile(newLat, float64(i))
		if i == 0 {
			continue
		} else if i > 1 {
			x = x % 2
			y = y % 2
		}
		pos = fmt.Sprintf("%d%s", (y%2)*2+(x%2), pos)

	}

	posNumber, _ := strconv.ParseInt(pos, 10, 64)
	return posNumber
}

func GetMinMaxPositionByTile(x, y, z, maxZoom int64) (string, string) {
	pos := ""
	newX := x
	newY := y
	for i := z; i > 0; i-- {
		pos = fmt.Sprintf("%d%s", (newY%2)*2+(newX%2), pos)
		newX = newX / 2
		newY = newY / 2
	}

	min, _ := strconv.Atoi(pos)
	zeros := ""
	for i := z + 1; i <= maxZoom; i++ {
		zeros = fmt.Sprintf("%s%d", zeros, 0)
	}

	return fmt.Sprintf("%d%s", min, zeros), fmt.Sprintf("%d%s", min+1, zeros)
}

func GetPolygonFromCell(cell, x, y, z int) orb.Polygon {
	// cellDefinition := []float64{2.8407093, 1.4203547, 0.898311, 0.449156, 0.200868, 0.089833, 0.04492, 0.02008, 0.00898, 0.00492025295, 0.0028407093, 0.0014203547, 0.0006352019, 0.00028407093, 0.000200868, 0.000089833}
	cellSizeLat := cellAreas[z]
	cellSizeLon := cellAreas[z]
	bounds := TileToBBOX(x, y, z)

	numCellsLat := math.Ceil((bounds.MaxLat - bounds.MinLat) / cellSizeLat)
	numCellsLon := math.Ceil((bounds.MaxLon - bounds.MinLon) / cellSizeLon)
	deltaLat := (bounds.MaxLat - bounds.MinLat) / numCellsLat
	deltaLon := (bounds.MaxLon - bounds.MinLon) / numCellsLon

	x1 := cell % int(numCellsLon)
	y1 := cell / int(numCellsLon)
	// deltaLat := (bounds.MaxLat - bounds.MinLat) / float64(numCells)
	// deltaLon := (bounds.MaxLon - bounds.MinLon) / float64(numCells)

	lat := bounds.MinLat + float64(y1)*deltaLat
	lon := bounds.MinLon + float64(x1)*deltaLon

	return orb.Polygon{orb.Ring{orb.Point{lon, lat}, orb.Point{lon + deltaLon, lat}, orb.Point{lon + deltaLon, lat + deltaLat}, orb.Point{lon, lat + deltaLat}, orb.Point{lon, lat}}}

}
