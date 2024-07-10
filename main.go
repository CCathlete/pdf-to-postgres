package main

import (
	"errors"
	"fmt"
	"os"
	dbH "pdf-to-postgres/dbHandler"
	pdf2txt "pdf-to-postgres/pdfHandler"
	ymlH "pdf-to-postgres/yamlHandler"
	"strings"
)

func main() {
	// docPath := "Parasitology/dummy_parasite.pdf"
	docPath := "Parasitology/Parasitology_book_2.pdf"
	var configYml ymlH.Config
	yamlPath := "config.yaml"
	configYml = ymlH.ParseYaml(yamlPath)
	fmt.Printf("Our config is: %v\n", configYml)
	needProcessing := configYml.NeedProcessing

	// Checking if the text file already exists.
	inputDirPath := "Parasitology/book_jpgs"
	txtPath := strings.Replace(docPath, "pdf", "txt", -1) // -1 means all instances.
	if _, err := os.Stat(txtPath); errors.Is(err, os.ErrNotExist) {
		pdf2txt.ConvertToTextWhenNotScanned(docPath)
		pdf2txt.ImagesToText(inputDirPath, docPath, needProcessing)
		fmt.Println("PDF doc was converted to text at path: " + txtPath)
	} else {
		fmt.Println("The text file exists.")
	}

	animalNames := configYml.AnimalNames
	parasites := pdf2txt.ExtractParasitesInfo(txtPath, animalNames)

	dbName := ymlH.GetDbName(yamlPath)
	dbName = strings.ToLower(dbName)
	fmt.Printf("Our DB name is: %s\n", dbName)

	dbInfo := configYml.DbInfo
	dbPointer := dbH.DbInit(dbInfo, animalNames)

	for _, animal := range animalNames {
		for _, parasiteInfo := range parasites[animal] {
			dbH.AddToTable(dbPointer, animal, parasiteInfo)
		}
	}
}

/*
In case we want to use the golang wrapper of poppler (pdftotext). It seemed to work better from the cli.
	pages := pdf2txt.ParsePdf(docPath)
	file, err := os.OpenFile(txtPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	defer func() {
		err := file.Close()
		if err != nil {
			log.Fatalf("Failed to close the txt file: %v.", err)
		}
	}()
	if err != nil {
		log.Fatalf("Failed to append or create the txt file: %v.", err)
	}

	for _, page := range pages {
		_, err := file.WriteString(page.Content + "\n")
		if err != nil {
			log.Fatalf("Failed to write to the txt file: %v.", err)
		}
	}
*/
