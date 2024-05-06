package yamlhandler

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

func ParseYaml(yamlPath string) map[string]interface{} {
	file, err := os.OpenFile(yamlPath, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatalf("Problem with opening the file in path %s: %v", yamlPath, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalf("Problem with closing the file in path %s: %v", yamlPath, err)
		}
	}()
	fileInfo, _ := file.Stat()
	dataBytes := make([]byte, fileInfo.Size())

	// Reading the file we opened.
	_, err = file.Read(dataBytes)
	if err != nil {
		log.Fatalf("Problem with reading the file in path %s: %v", yamlPath, err)
	}
	// Equivalent to make(map[string]interface{})
	configMap := map[string]interface{}{}
	err = yaml.Unmarshal(dataBytes, configMap)
	if err != nil {
		log.Fatalf("Couldn't Unmarshal the yaml file in %s: %v", yamlPath, err)
	}

	return configMap
}

func GetDbName(yamlPath string) string {
	configInfoMap := ParseYaml(yamlPath)
	dbInfoMap := configInfoMap["Database"].(map[interface{}]interface{})
	dbName := dbInfoMap["name"].(string)
	return dbName
}
