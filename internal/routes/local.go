package routes

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/4wings/cli/internal"
	"github.com/4wings/cli/internal/actions"
	"github.com/4wings/cli/internal/database"
	"github.com/4wings/cli/internal/utils"
	"github.com/4wings/cli/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func UploadFile(c *gin.Context) {
	log.Debug("Uploading file")
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.AbortWithStatusJSON(types.ServiceUnavailableCode, types.NewServerUnavailableStandard(err.Error()))
		return
	}
	typeDataset := c.Request.Form.Get("type")
	if typeDataset != "4wings" && typeDataset != "context" {
		c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
			Title:  "context",
			Detail: "Invalid context. Values allowed [context, 4wings]",
		}}))
		return
	}
	utils.CreateFolder(fmt.Sprintf("./%s", internal.DATA_FOLDER))
	extension := "csv"
	if typeDataset == "context" {
		extension = "json"
	}

	filename := uuid.New().String()
	out, err := os.Create(fmt.Sprintf("./%s/%s.%s", internal.DATA_FOLDER, filename, extension))
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		log.Fatal(err)
	}
	utils.WriteTempFile(types.TempFile{
		Name:   filename,
		Status: types.Created,
	})
	if typeDataset == "4wings" {
		go func() {
			log.Debugf("Starting import of file %s", filename)

			err := utils.WriteTempFile(types.TempFile{
				Name:   filename,
				Status: types.Importing,
			})
			if err != nil {
				log.Errorf("error updating temp_dataset %e", err)
			}
			path := fmt.Sprintf("./%s/%s.csv", internal.DATA_FOLDER, filename)

			err = database.LocalDB.IngestFile(path, filename, true)
			if err != nil {
				log.Errorf("error importing file %e", err)
				err := utils.WriteTempFile(types.TempFile{
					Name:    filename,
					Status:  types.Error,
					Message: err.Error(),
				})
				if err != nil {
					log.Errorf("error updating temp_file %e", err)
				}
			} else {
				err := utils.WriteTempFile(types.TempFile{
					Name:   filename,
					Status: types.Completed,
				})
				if err != nil {
					log.Errorf("error updating temp_file %e", err)
				}
				log.Infof("Imported file %s", filename)
			}

		}()
	}
	c.JSON(http.StatusOK, gin.H{"filename": filename})

}

func GettingFieldsFile(c *gin.Context) {
	log.Debug("Getting fields of file")
	filename := c.Param("filename")
	if filename == "" {
		c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
			Title:  "filename",
			Detail: "filename empty",
		}}))
		return
	}
	tempFile, err := actions.GetTempFile(filename)
	if err != nil || tempFile == nil {
		c.AbortWithStatusJSON(types.NotFoundCode, types.NewNotFoundStandard(fmt.Sprintf("filename %s not found", filename)))
		return
	}
	if tempFile.Status != types.Completed {
		c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
			Title:  "temp-file",
			Detail: fmt.Sprintf("file not completed. Status %s", tempFile.Status),
		}}))
		return
	}
	columns, err := database.LocalDB.GetSchema(filename)
	if err != nil {
		if err != nil {
			c.AbortWithStatusJSON(types.ServiceUnavailableCode, types.NewServerUnavailableStandard(err.Error()))
			return
		}
	}

	c.JSON(http.StatusOK, columns)
}

func GetTempFile(c *gin.Context) {
	log.Debug("Getting status of temp file")
	filename := c.Param("filename")
	if filename == "" {
		c.AbortWithStatusJSON(types.UnprocessableEntityCode, types.NewUnprocessableEntityStandard([]types.MessageError{{
			Title:  "filename",
			Detail: "filename empty",
		}}))
		return
	}
	tempFile, err := actions.GetTempFile(filename)
	if err != nil || tempFile == nil {
		c.AbortWithStatusJSON(types.NotFoundCode, types.NewNotFoundStandard(fmt.Sprintf("filename %s not found", filename)))
		return
	}
	c.JSON(http.StatusOK, tempFile)

}
