package middlewares

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/4wings/cli/internal/actions"
	"github.com/4wings/cli/internal/utils"
	"github.com/4wings/cli/types"
	"github.com/blastrain/vitess-sqlparser/sqlparser"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func IsValidZoomInDatasetsMiddleware(c *gin.Context) {
	z, _ := strconv.Atoi(c.Param("z"))
	datasetGroups := c.Keys["datasetGroups"].([][]*types.Dataset)
	var unsupportedDatasets []string
	for _, group := range datasetGroups {
		for _, dataset := range group {
			if dataset.Configuration.MaxZoom != 0 && dataset.Configuration.MaxZoom < z {
				unsupportedDatasets = append(unsupportedDatasets, dataset.ID)
			}
		}
	}
	if len(unsupportedDatasets) > 0 {
		c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
			Title:  "z",
			Detail: fmt.Sprintf("Zoom %d is not supported in datasets %s", z, strings.Join(unsupportedDatasets, ",")),
		}}))
		return
	}
	c.Next()
}

func parseDatasets(c *gin.Context) ([][]*types.Dataset, error) {
	_, existComparisonRange := c.GetQuery("comparison-range")
	datasetGroups := make([][]string, 0)
	datasetList := make([]string, 0)
	if !existComparisonRange {
		for i := 0; ; i++ {
			datasets, exists := c.GetQuery(fmt.Sprintf("datasets[%d]", i))
			if !exists {
				break
			}
			datasetGroups = append(datasetGroups, strings.Split(datasets, ","))
			datasetList = append(datasetList, strings.Split(datasets, ",")...)
		}
	} else {
		datasets, exists := c.GetQuery("datasets[0]")
		if exists {
			datasetGroups = append(datasetGroups, strings.Split(datasets, ","))
			datasetGroups = append(datasetGroups, strings.Split(datasets, ",")) //second group is the same as first
			datasetList = append(datasetList, strings.Split(datasets, ",")...)
		}

	}

	datasetObjs, err := actions.GetDatasets()
	if err != nil {
		return nil, err
	}

	result := make([][]*types.Dataset, len(datasetGroups))
	for gIndex, group := range datasetGroups {
		result[gIndex] = make([]*types.Dataset, len(group))
		for i, d := range group {
			for _, dObj := range datasetObjs {
				if dObj.ID == d {
					result[gIndex][i] = &dObj
					break
				}
			}
			if result[gIndex][i] == nil {
				return nil, types.NewNotFoundStandard(fmt.Sprintf("Dataset with id %s not exists", d))
			}
		}
	}
	return result, nil
}

func parseFilters(c *gin.Context, numGroups int) (filters []string, originalFilters []string) {
	filters = make([]string, numGroups)
	originalFilters = make([]string, numGroups)
	datesArray := []string{}
	existDateFilter := false
	comparison := false
	comparisonRange, existDateFilter := c.GetQuery("comparison-range")
	if !existDateFilter {
		dateRange, existsDate := c.GetQuery("date-range")
		if existsDate {
			//validate dateRange
			startDate, endDate, err := utils.ParseDateRange(dateRange)
			if err != nil {
				c.AbortWithStatusJSON(types.UnprocessableEntityCode, err)
				return
			}
			if startDate.After(endDate) {
				c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
					Title:  "date-range",
					Detail: "Start date is after of end date",
				}}))
				return
			}
			existDateFilter = existsDate
			for i := 0; i < numGroups; i++ {
				datesArray = append(datesArray, dateRange)
			}
		}
	} else {
		comparison = true
		comparisonDates := strings.Split(comparisonRange, ",")
		index := 0
		for i := 0; i < numGroups; i++ {
			datesArray = append(datesArray, fmt.Sprintf("%s,%s", comparisonDates[index], comparisonDates[index+1]))
			index += 2
		}
		if !strings.Contains(comparisonDates[0], "T") {
			comparisonDates[0] = fmt.Sprintf("%sT00:00:00.000Z", comparisonDates[0])
		}

		firstStartRange, err := time.Parse("2006-01-02T15:04:05.999Z", comparisonDates[0])
		if err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
				Title:  "comparison-range",
				Detail: fmt.Sprintf("Format of comparison date (%s) not valid", comparisonDates[0]),
			}}))

			return
		}
		if !strings.Contains(comparisonDates[2], "T") {
			comparisonDates[2] = fmt.Sprintf("%sT00:00:00.000Z", comparisonDates[2])
		}
		secondStartRange, err := time.Parse("2006-01-02T15:04:05.999Z", comparisonDates[2])
		if err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
				Title:  "comparison-range",
				Detail: fmt.Sprintf("Format of comparison date (%s) not valid", comparisonDates[2]),
			}}))

			return
		}
		intervalNum := 60 * 60
		if c.Query("interval") == "day" {
			intervalNum = 24 * 60 * 60
		} else if c.Query("interval") == "10days" {
			intervalNum = 10 * 24 * 60 * 60
		}

		diff := ((secondStartRange.Unix() / int64(intervalNum)) - (firstStartRange.Unix())/int64(intervalNum))
		c.Set("comparisonDiff", int(diff))

	}

	for i := 0; i < numGroups; i++ {
		var filter string
		var existsFilter bool
		if !comparison {
			filter, existsFilter = c.GetQuery(fmt.Sprintf("filters[%d]", i))
		} else {
			filter, existsFilter = c.GetQuery("filters[0]")
		}

		if !existsFilter || filter == "" {
			filters[i] = ""
			originalFilters[i] = ""
		} else {
			fmt.Println("query", "select * from user where "+filter)
			_, err := sqlparser.Parse("select * from user where " + filter)
			if err != nil {
				log.Error(err)
				c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
					Title:  fmt.Sprintf("filters[%d]", i),
					Detail: fmt.Sprintf("Format of where (%s) not valid", filter),
				}}))

				return
			}
			filters[i] = filter
			originalFilters[i] = filter
		}
		if existDateFilter {
			dates := strings.Split(datesArray[i], ",")
			if !existsFilter || filter == "" {
				filters[i] = fmt.Sprintf(" timestamp >= '%s' and timestamp <= '%s'", dates[0], dates[1])
			} else {
				filters[i] = fmt.Sprintf("%s and timestamp >= '%s' and timestamp <= '%s'", filters[i], dates[0], dates[1])
			}
		}
	}
	return filters, originalFilters
}

func DatasetMiddleware(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Next()
		return
	}

	datasetGroups, err := parseDatasets(c)
	if err != nil {
		if appError, ok := err.(*types.AppError); ok {
			c.AbortWithStatusJSON(appError.Code, appError)
			return
		}
		c.AbortWithStatusJSON(404, err)
		return
	}
	if len(datasetGroups) == 0 {
		c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
			Title:  "datasets",
			Detail: "datasets query param is required",
		}}))
		return
	}

	filters, originalFilters := parseFilters(c, len(datasetGroups))
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	temporalAggregation := c.Query("temporal-aggregation")
	if temporalAggregation == "true" {
		c.Keys["temporalAggregation"] = true
	} else if temporalAggregation == "false" {
		c.Keys["temporalAggregation"] = false
	}
	c.Keys["datasetGroups"] = datasetGroups
	c.Keys["filters"] = filters
	c.Keys["originalFilters"] = originalFilters

	c.Next()

}
