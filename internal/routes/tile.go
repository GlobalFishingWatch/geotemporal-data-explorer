package routes

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"strconv"

	"github.com/4wings/cli/internal/actions"
	"github.com/4wings/cli/internal/utils"
	"github.com/4wings/cli/types"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Tile(c *gin.Context) {

	z, _ := strconv.Atoi(c.Param("z"))
	x, _ := strconv.Atoi(c.Param("x"))
	y, _ := strconv.Atoi(c.Param("y"))

	if z > 12 {
		c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
			Title:  "z",
			Detail: "The tiler does not support zoom levels greather than 12",
		}}))
		return
	}
	temporalAggr := c.Keys["temporalAggregation"].(bool)

	format := c.Query("format")

	datasetGroups := c.Keys["datasetGroups"].([][]*types.Dataset)
	filters := c.Keys["filters"].([]string)
	interval := c.Query("interval")
	if interval == "" && temporalAggr {
		interval = "hour"
	} else if interval == "" {
		c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
			Title:  "interval",
			Detail: "Interval is required. Possible values: hour, day, 10days, month",
		}}))
		return
	}
	_, isComparison := c.GetQuery("comparison-range")
	comparisonDiff := -1
	if isComparison {
		comparisonDiff = c.Keys["comparisonDiff"].(int)
	}

	var err error

	log.Debugf("Coords %i/%i/%i", z, x, y)
	var data []byte
	var headers map[string]string
	if datasetGroups[0][0].Source != "GFW" && datasetGroups[0][0].Source != "GEE" {

		data, headers, err = actions.GenerateTile4wings(datasetGroups, z, x, y, interval, filters, format, temporalAggr, comparisonDiff, c.Query("resolution"))

	} else {
		if comparisonDiff > 0 {
			c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
				Title:  "dataset",
				Detail: "The dataset/s not support comparison ranges",
			}}))
			return
		}
		dateRange := c.Query("date-range")
		if datasetGroups[0][0].Source == "GEE" {

			if len(datasetGroups) > 1 || len(datasetGroups[0]) > 1 {
				c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
					Title:  "dataset",
					Detail: "GEE Datasets only support one layer with one dataset",
				}}))
				return
			}
			if !viper.GetBool("gee") {
				c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
					Title:  "dataset",
					Detail: "GEE Datasets not enabled. You need to configure a service account",
				}}))
				return
			}
			if dateRange == "" {
				dateRange = fmt.Sprintf("%s,%s", datasetGroups[0][0].StartDate.Format("2006-01-02"), datasetGroups[0][0].EndDate.Format("2006-01-02"))
			} else {
				startDate, endDate, err := utils.ParseDateRange(dateRange)
				if err != nil {
					c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
						Title:  "date-range",
						Detail: "invalid date-range",
					}}))
				}

				if datasetGroups[0][0].StartDate.After(startDate) {
					startDate = datasetGroups[0][0].StartDate
				}
				if datasetGroups[0][0].EndDate.Before(endDate) {
					endDate = datasetGroups[0][0].EndDate
				}
				dateRange = fmt.Sprintf("%s,%s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
			}

			if interval == "10days" || datasetGroups[0][0].Configuration.Images[utils.INTERVALS_GEE[interval]] == "" {
				c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
					Title:  "interval",
					Detail: fmt.Sprintf("Interval %s not supported for dataset %s", interval, datasetGroups[0][0].ID),
				}}))
				return
			}
			data, headers, err = actions.GenerateTileGEE(datasetGroups[0][0], z, x, y, dateRange, format, temporalAggr, interval)
		} else if datasetGroups[0][0].Source == "GFW" {
			if !viper.GetBool("gfw") {
				c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
					Title:  "dataset",
					Detail: "GFW Datasets not enabled. You need to configure your GFW token",
				}}))
				return
			}
			actions.GenerateTileGFW(c)
			return
		}
	}
	if err != nil {
		c.AbortWithStatusJSON(types.ServiceUnavailableCode, types.NewServerUnavailableStandard(err.Error()))
		return
	}
	if data == nil {
		// c.Writer.WriteHeader(types.NoContent)
		c.AbortWithStatusJSON(types.NotFoundCode, types.NewNotFoundStandard("Tile empty"))
		return
	}
	for k, v := range headers {
		c.Header(k, v)
	}

	var b bytes.Buffer
	gz, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		c.AbortWithStatusJSON(types.ServiceUnavailableCode, types.NewServerUnavailableStandard("Error compressing"))
		return
	}
	if _, err := gz.Write(data); err != nil {
		c.AbortWithStatusJSON(types.ServiceUnavailableCode, types.NewServerUnavailableStandard("Error compressing"))
		return
	}
	if err := gz.Close(); err != nil {
		c.AbortWithStatusJSON(types.ServiceUnavailableCode, types.NewServerUnavailableStandard("Error compressing"))
		return
	}
	c.Header("content-encoding", "gzip")
	c.Header("Cache-Control", "private, max-age=86400")
	if format == "mvt" {
		c.Data(http.StatusOK, "application/vnd.mapbox-vector-tile", b.Bytes())
	} else {
		c.Data(http.StatusOK, "application/x-protobuf", b.Bytes())
	}

}
