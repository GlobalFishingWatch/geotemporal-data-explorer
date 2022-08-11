package routes

import (
	"fmt"
	"net/http"
	"time"

	"github.com/4wings/cli/internal"
	"github.com/4wings/cli/internal/actions"
	"github.com/4wings/cli/internal/utils"
	"github.com/4wings/cli/types"
	"github.com/gin-gonic/gin"
	validator "github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

func GetAllDatasets(c *gin.Context) {
	log.Debug("Obtaining datasets")
	datasets, err := actions.GetDatasets()
	if err != nil {
		c.AbortWithStatusJSON(types.ServiceUnavailableCode, types.NewServerUnavailableStandard(err.Error()))
		return
	}
	c.JSON(200, datasets)

}

func GetDataset(c *gin.Context) {
	datasetID := c.Param("id")
	log.Debug("Get dataset %s", datasetID)
	dataset, err := actions.GetDataset(datasetID)
	if err != nil {
		c.AbortWithStatusJSON(types.ServiceUnavailableCode, types.NewServerUnavailableStandard(fmt.Sprintf("error %e", err)))
		return
	}
	if dataset == nil {
		c.AbortWithStatusJSON(types.NotFoundCode, types.NewNotFoundStandard(fmt.Sprintf("dataset with id %s not found", datasetID)))
		return
	}
	c.JSON(http.StatusOK, dataset)
}

func DeleteDataset(c *gin.Context) {
	datasetID := c.Param("id")
	log.Debug("Get dataset %s", datasetID)
	dataset, err := actions.GetDataset(datasetID)
	if err != nil {
		c.AbortWithStatusJSON(types.ServiceUnavailableCode, types.NewServerUnavailableStandard(fmt.Sprintf("error %e", err)))
		return
	}
	if dataset == nil {
		c.AbortWithStatusJSON(types.NotFoundCode, types.NewNotFoundStandard(fmt.Sprintf("dataset with id %s not found", datasetID)))
		return
	}
	err = actions.DeleteDataset(datasetID)
	if err != nil {
		c.AbortWithStatusJSON(types.ServiceUnavailableCode, types.NewServerUnavailableStandard(fmt.Sprintf("error %e", err)))
		return
	}
	c.JSON(http.StatusOK, dataset)
}

func CreateDataset(c *gin.Context) {
	log.Debug("Creating dataset")
	var dataset types.Dataset
	if err := c.ShouldBindJSON(&dataset); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
				Title:  "validation",
				Detail: err.Error(),
			}}))
			return
		}
		if _, ok := err.(validator.ValidationErrors); ok {
			errorMessages := make([]types.MessageError, 0)
			for _, err := range err.(validator.ValidationErrors) {
				errorMessages = append(errorMessages, types.MessageError{
					Title:  err.Field(),
					Detail: err.ActualTag(),
				})
			}
			c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard(errorMessages))
			return
		}
		if _, ok := err.(*time.ParseError); ok {
			c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
				Title:  "time",
				Detail: err.Error(),
			}}))

			return
		}
	}
	if dataset.Type == "context" {
		dataset, err := actions.CreateNewContextDataset(dataset)
		if err != nil {
			c.AbortWithStatusJSON(types.ServiceUnavailableCode, types.NewServerUnavailableStandard(err.Error()))
			return
		}
		c.JSON(200, dataset)
		return
	} else if dataset.Type == "4wings" {
		if dataset.Configuration.AggregationOperation == "" {
			c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
				Title:  "aggregationOperation",
				Detail: fmt.Sprintf("aggregationOperation is required. Should be one of these values (SUM, AVG, MIN, MAX)"),
			}}))

			return
		}
		dataset, err := actions.CreateNew4wingsDataset(dataset)
		if err != nil {
			c.AbortWithStatusJSON(types.ServiceUnavailableCode, types.NewServerUnavailableStandard(err.Error()))
			return
		}
		c.JSON(200, dataset)
		return
	} else {
		c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
			Title:  "type",
			Detail: fmt.Sprintf("type %s not supported", dataset.Type),
		}}))
		return
	}

}

func GetFiltersData(c *gin.Context) {
	datasetID := c.Param("id")
	dataset, err := actions.GetDataset(datasetID)
	if err != nil {
		c.AbortWithStatusJSON(types.ServiceUnavailableCode, types.NewServerUnavailableStandard(fmt.Sprintf("error %e", err)))
		return
	}
	if dataset == nil {
		c.AbortWithStatusJSON(types.NotFoundCode, types.NewNotFoundStandard(fmt.Sprintf("dataset with id %s not found", datasetID)))
		return
	}
	if dataset.Type != "4wings" {
		c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
			Title:  "type",
			Detail: "dataset should be of 4wings type",
		}}))
		return
	}
	log.Debug("Getting filters of 4wings dataset %s", datasetID)
	data, err := actions.GetFiltersOfDataset(dataset)
	if err != nil {
		c.AbortWithStatusJSON(types.ServiceUnavailableCode, types.NewServerUnavailableStandard(fmt.Sprintf("error %e", err)))
		return
	}
	c.JSON(http.StatusOK, data)
}

func GetContextData(c *gin.Context) {
	datasetID := c.Param("id")
	log.Debug("Getting content of context dataset %s", datasetID)
	dataset, err := actions.GetDataset(datasetID)
	if err != nil {
		c.AbortWithStatusJSON(types.ServiceUnavailableCode, types.NewServerUnavailableStandard(fmt.Sprintf("error %e", err)))
		return
	}
	if dataset == nil {
		c.AbortWithStatusJSON(types.NotFoundCode, types.NewNotFoundStandard(fmt.Sprintf("dataset with id %s not found", datasetID)))
		return
	}
	if dataset.Type != "context" {
		c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
			Title:  "type",
			Detail: "dataset should be of context type",
		}}))
		return
	}
	data, err := utils.ReadFile(fmt.Sprintf("./%s/%s", internal.DATA_FOLDER, dataset.Configuration.FileID))
	if err != nil {
		c.AbortWithStatusJSON(types.ServiceUnavailableCode, types.NewServerUnavailableStandard(fmt.Sprintf("error %e", err)))
		return
	}
	if data == "" {
		c.AbortWithStatusJSON(types.NotFoundCode, types.NewNotFoundStandard("File not exists"))
		return
	}
	c.Data(http.StatusOK, "application/json", []byte(data))
}
