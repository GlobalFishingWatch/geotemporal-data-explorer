package actions

import (
	"fmt"

	"github.com/4wings/cli/internal"
	"github.com/4wings/cli/internal/utils"
	"github.com/4wings/cli/types"
	log "github.com/sirupsen/logrus"
)

func GetTempFiles() ([]types.TempFile, error) {
	return utils.ReadTempFilesFile()

}

func GetTempFile(id string) (*types.TempFile, error) {
	tempFiles, err := GetTempFiles()
	if err != nil {
		log.Errorf("Error reading tempFiles %e", err)
		return nil, err
	}
	for _, d := range tempFiles {
		if d.Name == id {
			return &d, nil
		}
	}

	return nil, nil
}

func DeleteTempFile(id string) error {
	log.Debugf("Deleting tempFile with id %s", id)
	tempFile, _ := GetTempFile(id)
	utils.RemoveFile(fmt.Sprintf("./%s/%s.csv", internal.DATA_FOLDER, tempFile.Name))
	utils.RemoveFile(fmt.Sprintf("./%s/%s.json", internal.DATA_FOLDER, tempFile.Name))

	return utils.DeleteTempFile(id)
}
