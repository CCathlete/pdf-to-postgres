package yamlhandler

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func ParseYaml(yamlPath string) (configMap map[string]map[string]string) {
	file, err := os.OpenFile(yamlPath, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatalf("Problem with opening the file in path %s: %v", yamlPath, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalf("Problem with closing the file in path %s: %v", yamlPath, err)
		}
	}()
	dataBytes := []byte{}
	_, err = file.Read(dataBytes)
	if err != nil {
		log.Fatalf("Problem with reading the file in path %s: %v", yamlPath, err)
	}

	err = yaml.Unmarshal(dataBytes, configMap)
	if err != nil {
		log.Fatalf("Couldn't Unmarshal the yaml file in %s: %v", yamlPath, err)
	}

	return configMap
}
