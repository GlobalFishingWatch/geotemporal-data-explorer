package actions

import (
	"fmt"
	"time"

	"github.com/4wings/cli/internal"
	"github.com/4wings/cli/internal/database"
	"github.com/4wings/cli/internal/utils"
	"github.com/4wings/cli/types"
	"github.com/ettle/strcase"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func GetDatasets() ([]types.Dataset, error) {
	localDatasets, err := utils.ReadDatasetFile()
	if err != nil {
		log.Errorf("Error reading datasets %e", err)
		return nil, err
	}
	internalDataset := []types.Dataset{}
	if viper.GetBool("gee") {
		internalDataset = append(internalDataset, internal.GEE_DATASETS...)
	}
	if viper.GetBool("gfw") {
		internalDataset = append(internalDataset, internal.GFW_DATASETS...)
	}
	if len(localDatasets) > 0 {
		return append(internalDataset, localDatasets...), nil
	}
	return internalDataset, nil
}

func GetDataset(id string) (*types.Dataset, error) {
	datasets, err := GetDatasets()
	if err != nil {
		log.Errorf("Error reading datasets %e", err)
		return nil, err
	}
	for _, d := range datasets {
		if d.ID == id {
			return &d, nil
		}
	}

	return nil, nil
}

func DeleteDataset(id string) error {
	log.Debugf("Deleting dataset with id %s", id)
	dataset, _ := GetDataset(id)
	utils.RemoveFile(fmt.Sprintf("./%s/%s", internal.DATA_FOLDER, dataset.Configuration.FileID))
	if dataset.Type == "4wings" {
		err := database.LocalDB.DropTable(dataset.Configuration.FileID)
		if err != nil {
			log.Errorf("error removing table %s: %e", dataset.Configuration.FileID, err)
		}
		database.LocalDB.DropTable(dataset.Configuration.Table)
		for _, interval := range dataset.Configuration.Intervals {
			for i := 0; i <= 12; i++ {
				database.LocalDB.DropTable(fmt.Sprintf("%s_%s_z%d", dataset.Configuration.Table, interval, i))
			}
		}

	}
	return utils.DeleteDataset(id)
}

func CreateNewContextDataset(dataset types.Dataset) (types.Dataset, error) {
	log.Debug("Creating new context dataset")
	dataset.Configuration.FileID = fmt.Sprintf("%s.json", dataset.Configuration.FileID)
	dataset.ID = fmt.Sprintf("%s-%d", strcase.ToKebab(dataset.Name), time.Now().UnixMicro())

	err := utils.WriteDataset(dataset)
	return dataset, err
}

func CreateNew4wingsDataset(dataset types.Dataset) (types.Dataset, error) {
	log.Debug("Creating new 4wings dataset")
	dataset.ID = fmt.Sprintf("%s-%d", strcase.ToKebab(dataset.Name), time.Now().UnixMicro())
	if dataset.Configuration.Fields.Resolution == "hour" {
		dataset.Configuration.Intervals = []string{"hour", "day", "month"}
		// create daily and monthly tables
		err := database.LocalDB.CreateGroupedTables(dataset, "month")
		if err != nil {
			return types.Dataset{}, err
		}
		err = database.LocalDB.CreateGroupedTables(dataset, "day")
		if err != nil {
			return types.Dataset{}, err
		}
	} else if dataset.Configuration.Fields.Resolution == "day" {
		dataset.Configuration.Intervals = []string{"day", "month"}
		// create monthly table
		err := database.LocalDB.CreateGroupedTables(dataset, "month")
		if err != nil {
			return types.Dataset{}, err
		}
	}
	dataset.Status = types.Created
	dataset.Configuration.Table = fmt.Sprintf("4wings_%s", dataset.Configuration.FileID)
	// err := database.LocalDB.CreateRawTable(dataset)
	// if err != nil {
	// 	return types.Dataset{}, err
	// }
	err := utils.WriteDataset(dataset)
	database.LocalDB.IngestDataset(dataset)
	return dataset, err
}

func GetFiltersOfDataset(dataset *types.Dataset) (map[string][]interface{}, error) {
	data := map[string][]interface{}{}
	for _, f := range dataset.Configuration.Fields.Filters {
		values, err := database.LocalDB.GetDistinctValuesOfColumn(dataset.Configuration.Table, f)
		if err != nil {
			return nil, err
		}
		data[f] = values
	}

	return data, nil
}
