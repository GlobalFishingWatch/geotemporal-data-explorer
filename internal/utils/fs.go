package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/4wings/cli/internal"
	"github.com/4wings/cli/types"
	log "github.com/sirupsen/logrus"
)

func CreateFolder(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return os.Mkdir(path, os.ModePerm)
	}
	return nil
}

func ExistFile(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func ReadFile(path string) (string, error) {
	if !ExistFile(path) {
		log.Debugf("not exist file in path %s", path)
		return "", nil
	}
	file, err := os.Open(path)
	if err != nil {
		log.Errorf("Error opening file %e", err)
		return "", err
	}
	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("Error reading file %e", err)
		return "", err
	}
	return string(byteValue), nil
}

func RemoveFile(path string) error {
	if !ExistFile(path) {
		return nil
	}
	return os.Remove(path)
}

func ReadDatasetFile() ([]types.Dataset, error) {
	if !ExistFile(internal.DATASETS_PATH) {
		return nil, nil
	}
	jsonFile, err := os.Open(internal.DATASETS_PATH)
	if err != nil {
		log.Errorf("Error opening json dataset %e", err)
		return nil, err
	}
	defer jsonFile.Close()
	var datasets []types.Dataset
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Errorf("Error reading json dataset %e", err)
		return nil, err
	}
	err = json.Unmarshal(byteValue, &datasets)
	if err != nil {
		log.Errorf("Error unmarshal json dataset %e", err)
		return nil, err
	}
	return datasets, nil
}

func WriteDataset(dataset types.Dataset) error {
	datasets, err := ReadDatasetFile()
	if err != nil {
		return err
	}
	datasets = append(datasets, dataset)
	file, _ := json.MarshalIndent(datasets, "", " ")

	err = ioutil.WriteFile(internal.DATASETS_PATH, file, 0644)
	return err
}

func DeleteDataset(datasetID string) error {
	datasets, err := ReadDatasetFile()
	if err != nil {
		return err
	}
	toSave := make([]types.Dataset, 0)
	for _, d := range datasets {
		if d.ID != datasetID {
			toSave = append(toSave, d)
		}
	}

	file, _ := json.MarshalIndent(toSave, "", " ")

	err = ioutil.WriteFile(internal.DATASETS_PATH, file, 0644)
	return err
}
