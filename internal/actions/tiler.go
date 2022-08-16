package actions

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/4wings/cli/internal/database"
	tile "github.com/4wings/cli/internal/proto"
	"github.com/4wings/cli/internal/utils"
	"github.com/4wings/cli/types"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	"github.com/paulmach/orb/encoding/mvt"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
)

var cellsByZoom []float64 = []float64{128000000000, 32000000000, 8000000000, 2000000000, 500000000, 125000000, 31250000, 7812500, 1953125, 488281, 122070, 30518, 7629}

func init() {

}

type Cell struct {
	cell    int
	min     int
	max     int
	htime   map[int]float64
	value   float64
	h3index string
}

func maxmin(results [][]*Cell, i int) (int, int, int) {
	max, min, cell := -1, -1, -1
	for j := 0; j < len(results); j++ {
		if results[j] == nil || results[j][i] == nil {
			continue
		}
		if max == -1 || results[j][i].max > max {
			max = results[j][i].max
			cell = results[j][i].cell
		}
		if min == -1 || results[j][i].min < min {
			min = results[j][i].min
		}

	}
	return max, min, cell

}

func maxminValue(results [][]*Cell, temporalAggr bool) (float64, float64) {
	max, min := -1.0, -1.0
	for _, r := range results {
		for _, c := range r {
			if temporalAggr {
				if c.value > max {
					max = c.value
				}
				if (min == -1.0 || c.value < min) && c.value > 0 {
					min = c.value
				}
			} else {
				if c != nil {
					for _, h := range c.htime {
						if h > max {
							max = h
						}
						if (min == -1.0 || h < min) && h > 0 {
							min = h
						}
					}
				}
			}
		}
	}
	return max, min

}

func generateMVT(results [][]*Cell, z, x, y int, numCellsLat, numCellsLon int, temporalAggregation bool) ([]byte, map[string]string, error) {
	log.Debug("Generating mvt")
	totalCells := numCellsLat * numCellsLon
	var feature *geojson.Feature
	fc := geojson.NewFeatureCollection()

	for i := 0; i < totalCells; i++ {
		added := false
		point := utils.GetPolygonFromCell(i, x, y, z)
		if point == nil {
			continue
		}
		feature = geojson.NewFeature(point)

		if !temporalAggregation {
			max, min, _ := maxmin(results, i)
			for h := min; h <= max; h++ {
				if len(results) > 1 {
					count := make([]int, len(results))
					for j := 0; j < len(results); j++ {
						if results[j][i] != nil {
							v, exists := results[j][i].htime[h]
							if exists {
								added = true
								count[j] += int(v)
							}
						}
					}

					feature.Properties[strconv.Itoa(h)] = count
				} else {
					if results[0][i] != nil {
						v, exists := results[0][i].htime[h]
						if exists {
							added = true
							feature.Properties[strconv.Itoa(h)] = int(v)
						}
						// fmt.Println("feature", feature.Properties)
					}
				}
			}
		} else {
			if len(results) > 1 {
				count := make([]int, len(results))
				for j := 0; j < len(results); j++ {
					if results[j][i] != nil {
						count[j] = int(results[j][i].value)
					}
				}
				added = true
				feature.Properties["count"] = count
			} else {
				if results[0][i] != nil {
					added = true
					feature.Properties["count"] = results[0][i].value
				}
			}
		}
		if added {
			fc.Append(feature)
		}

		feature.Properties["cell"] = i
	}

	collections := map[string]*geojson.FeatureCollection{
		"main": fc,
	}
	layers := mvt.NewLayers(collections)
	layers[0].Version = 2
	log.Debugf("Generating tile %d/%d/%d \n", z, x, y)

	layers.ProjectToTile(maptile.New(uint32(x), uint32(y), maptile.Zoom(z)))
	layers.Clip(mvt.MapboxGLDefaultExtentBound)

	data, err := mvt.Marshal(layers)
	return data, nil, err

}

func generateCustomPBF(results [][]*Cell, numCellsLat, numCellsLon int, temporalAggregation bool, resolution string) ([]byte, map[string]string, error) {
	log.Debug("Generating pbf")
	headers := map[string]string{
		"num-cells-lat":      fmt.Sprintf("%d", numCellsLat),
		"num-cells-lon":      fmt.Sprintf("%d", numCellsLon),
		"decimal-precission": "2",
	}
	totalCells := numCellsLat * numCellsLon
	lengthArray := 2
	total := 0
	for i := 0; i < totalCells; i++ {
		if !temporalAggregation {
			max, min, _ := maxmin(results, i)
			if max != -1 {
				lengthArray += ((max - min + 1) * len(results)) + 3
				total++
			}
		} else {
			for j := 0; j < len(results); j++ {
				if results[j] != nil && results[j][i] != nil {
					lengthArray += 1 + len(results)
					break
				}
			}

		}
	}

	if lengthArray == 0 {
		return nil, nil, nil
	}
	if resolution == "high" || resolution == "" {
		data := make([]uint32, lengthArray)
		data[0] = uint32(numCellsLat)
		data[1] = uint32(numCellsLon)
		index := 2
		for i := 0; i < totalCells; i++ {
			max, min, cell := maxmin(results, i)

			if max == -1 {
				continue
			}
			data[index] = uint32(cell)
			index++
			if !temporalAggregation {
				data[index] = uint32(min)
				index++
				data[index] = uint32(max)
				index++

				for h := min; h <= max; h++ {
					for j := 0; j < len(results); j++ {
						if results[j] != nil && results[j][i] != nil {
							v, exists := results[j][i].htime[h]
							if !exists {
								data[index] = 0
							} else {
								data[index] = uint32(v)
							}
						} else {
							data[index] = 0
						}

						index++
					}

				}
			} else {
				for j := 0; j < len(results); j++ {
					if results[j] != nil {
						data[index] = uint32(results[j][i].value)
						index++
					} else {
						data[index] = 0
						index++
					}
				}
			}
		}

		tileTo := &tile.Tile{
			Data: data,
		}
		bytes, err := proto.Marshal(tileTo)
		if err != nil {
			return nil, nil, fmt.Errorf("Error generating protobuffer from array: %e", err)
		}
		return bytes, headers, nil
	} else {
		maxValue, minValue := maxminValue(results, temporalAggregation)
		var delta float64
		if resolution == "poor" {
			delta = (maxValue - minValue) / 255
		} else {
			delta = (maxValue - minValue) / 65535
		}
		headers["min-value"] = fmt.Sprintf("%f", minValue)
		headers["delta"] = fmt.Sprintf("%f", delta)

		fmt.Printf("max: %f, min: %f, Delta: %f", maxValue, minValue, delta)
		data := make([]uint32, lengthArray)
		data[0] = uint32(numCellsLat)
		data[1] = uint32(numCellsLon)
		index := 2
		for i := 0; i < totalCells; i++ {
			max, min, cell := maxmin(results, i)

			if max == -1 {
				continue
			}
			data[index] = uint32(cell)
			index++
			if !temporalAggregation {
				data[index] = uint32(min)
				index++
				data[index] = uint32(max)
				index++

				for h := min; h <= max; h++ {
					for j := 0; j < len(results); j++ {
						if results[j] != nil && results[j][i] != nil {
							v, exists := results[j][i].htime[h]
							if !exists {
								data[index] = 0
							} else {
								data[index] = uint32(((v - minValue) / delta) + 1)
							}
						} else {
							data[index] = 0
						}

						index++
					}

				}
			} else {
				for j := 0; j < len(results); j++ {
					if results[j] != nil {
						data[index] = uint32(((results[j][i].value - minValue) / delta) + 1)
						index++
					} else {
						data[index] = 0
						index++
					}
				}
			}
		}
		tileTo := &tile.Tile{
			Data: data,
		}
		bytes, err := proto.Marshal(tileTo)
		if err != nil {
			return nil, nil, fmt.Errorf("Error generating protobuffer from array: %e", err)
		}
		return bytes, headers, nil
	}

}

func getResultsDB(group []*types.Dataset, x, y, z int, pos int64, interval, filter string, numCellsLat, numCellsLon int, temporalAggregation bool) ([]*Cell, error) {
	fmt.Println(numCellsLat, numCellsLon)
	datasets := make([]string, len(group))
	relationalSQLGroup := make([]*types.Dataset, 0)
	bqGroup := make([]*types.Dataset, 0)
	for i, d := range group {
		if d.Configuration.Source == "bigquery" {
			bqGroup = append(bqGroup, d)
		} else {
			relationalSQLGroup = append(relationalSQLGroup, d)
		}
		datasets[i] = d.ID
	}
	log.Debugf("Obtaining results of datasets %s", strings.Join(datasets, ","))
	rows := types.NewRows()
	if len(relationalSQLGroup) > 0 {
		log.Debug("Obtaining results of postgres datasets")
		err := database.LocalDB.HeatmapQuery(relationalSQLGroup, int64(x), int64(y), int64(z), pos, interval, filter, temporalAggregation, rows)
		if err != nil {
			return nil, err
		}
	}
	// if len(bqGroup) > 0 {
	// 	err := database.BQDB.HeatmapQuery(bqGroup, int64(x), int64(y), int64(z), pos, interval, filter, temporalAggregation, vesselGroupIds, rows)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	defer rows.Close()

	readed := 0

	results := make([]*Cell, int(numCellsLat)*int(numCellsLon))

	for rows.Next() {
		readed++
		var cell, htime int
		var count float64
		var err error
		var obj types.RowObj
		if temporalAggregation {
			obj, err = rows.Scan(2)
		} else {
			obj, err = rows.Scan(3)
		}
		if err != nil {
			log.Debugf("Error doing query: %e", err)
			return nil, err
		}
		cell = int(obj.Cell)
		htime = int(obj.Htime)

		count = obj.Value
		if results[cell] == nil {
			results[cell] = &Cell{
				cell:  cell,
				min:   htime,
				max:   htime,
				htime: map[int]float64{},
				value: 0,
			}
		}
		if !temporalAggregation {
			_, exist := results[cell].htime[htime]
			if !exist {
				results[cell].htime[htime] = 0
			}

			results[cell].htime[htime] = results[cell].htime[htime] + (count * 100)
			if results[cell].min > htime {
				results[cell].min = htime
			}
			if results[cell].max < htime {
				results[cell].max = htime
			}
		} else {
			results[cell].value = results[cell].value + (count * 100)
		}

	}

	if readed == 0 {
		return nil, nil
	}
	return results, nil

}

func GenerateTileGEE(dataset *types.Dataset, z, x, y int, dateRange string, format string, temporalAggregation bool, interval string) ([]byte, map[string]string, error) {
	log.Debug("Generating tile from GEE")
	pos, err := utils.Tile2Num(z, x, y)
	if err != nil {
		log.Error("Error obtaining pos", err)
		return nil, nil, err
	}
	list, err := utils.ReadGEE(dataset, z, x, y, temporalAggregation, dateRange, interval)
	if err != nil {
		return nil, nil, err
	}
	totalResults := make([][]*Cell, 1)
	numCells := utils.GetCellsByPos(int(pos), z)
	numCellsLat, numCellsLon := utils.GetCellsLatLonByPos(int(pos), z)
	numGroup := 0
	totalResults[0] = make([]*Cell, numCells)
	startDate, _, err := utils.ParseDateRange(dateRange)
	if err != nil {
		return nil, nil, err
	}
	var start int
	if interval == "month" {
		start = (startDate.Year()-1970)*12 + int(startDate.Month()) - 1
	} else if interval == "day" {
		start = int(startDate.UnixMilli() / (24 * 60 * 60 * 1000))
	}

	var value float64
	hasData := false
	for i, val := range list {

		if val[2] == 0 {
			value = 0
			continue
		} else {
			value = float64(val[0]) / float64(val[1])
		}
		hasData = true
		if temporalAggregation {
			totalResults[numGroup][i] = &Cell{
				cell:  i,
				value: value,
			}
		} else {
			if totalResults[numGroup][i%numCells] == nil {
				totalResults[numGroup][i%numCells] = &Cell{
					cell:  i % numCells,
					htime: map[int]float64{},
					max:   0,
					min:   0,
				}
			}
			step := i / numCells
			totalResults[numGroup][i%numCells].htime[start+step] = value
			if totalResults[numGroup][i%numCells].max == 0 || totalResults[numGroup][i%numCells].max < start+step {
				totalResults[numGroup][i%numCells].max = start + step
			}
			if totalResults[numGroup][i%numCells].min == 0 || totalResults[numGroup][i%numCells].min > start+step {
				totalResults[numGroup][i%numCells].min = start + step
			}
		}

	}
	if !hasData {
		return nil, nil, nil
	}
	if format == "mvt" {
		return generateMVT(totalResults, z, x, y, int(numCellsLat), int(numCellsLon), temporalAggregation)
	} else {
		return generateCustomPBF(totalResults, int(numCellsLat), int(numCellsLon), temporalAggregation, "high")
	}
}

func GenerateTile4wings(datasetGroups [][]*types.Dataset, z, x, y int, interval string, filters []string, format string, temporalAggregation bool, comparisonDiff int, resolution string) ([]byte, map[string]string, error) {
	pos, err := utils.Tile2Num(z, x, y)
	if err != nil {
		log.Debugf("Error obtaining pos from tile: %e", err)
		return nil, nil, err
	}

	bounds := utils.TileToBBOX(x, y, z)
	valueCellByZoom := math.Sqrt(cellsByZoom[z]) / 111320
	cellSizeLat := valueCellByZoom
	cellSizeLon := valueCellByZoom
	numCellsLat := math.Ceil(
		(bounds.MaxLat - bounds.MinLat) / cellSizeLat,
	)
	numCellsLon := math.Ceil(
		(bounds.MaxLon - bounds.MinLon) / cellSizeLon,
	)
	totalResults := make([][]*Cell, len(datasetGroups))
	var wg sync.WaitGroup
	var generalError error
	for i, group := range datasetGroups {
		wg.Add(1)
		go func(group []*types.Dataset, filter string, numGroup int, wg *sync.WaitGroup) {
			results, err := getResultsDB(group, x, y, z, pos, interval, filter, int(numCellsLat), int(numCellsLon), temporalAggregation)
			if err != nil {
				log.Debugf("Error generating group: %e", err)
				totalResults[numGroup] = nil
				generalError = err
			} else {
				if results != nil {
					totalResults[numGroup] = results
				}
			}
			wg.Done()
		}(group, filters[i], i, &wg)

	}
	wg.Wait()
	if generalError != nil {
		return nil, nil, generalError
	}
	// check if all data is nil
	notNil := false
	for _, d := range totalResults {
		if d != nil {
			notNil = true
			break
		}
	}
	if !notNil {
		return nil, nil, nil
	}

	if comparisonDiff > 0 {
		log.Debug("Modifying result cells of group 2 because is a comparison feature")
		for _, cell := range totalResults[1] {
			if cell != nil {
				cell.max = cell.max - comparisonDiff
				cell.min = cell.min - comparisonDiff
				newHTime := map[int]float64{}
				for k, v := range cell.htime {
					newHTime[k-comparisonDiff] = v
				}
				cell.htime = newHTime
			}
		}
	}
	if format == "mvt" {
		return generateMVT(totalResults, z, x, y, int(numCellsLat), int(numCellsLon), temporalAggregation)
	} else {
		return generateCustomPBF(totalResults, int(numCellsLat), int(numCellsLon), temporalAggregation, resolution)
	}

}
