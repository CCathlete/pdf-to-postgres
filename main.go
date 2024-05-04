package main

import (
	"fmt"
	pdf2txt "pdf-to-postgres/pdfHandler"
	"strings"
)

func main() {
	docPath := "Parasitology/Parasitology_book.pdf"
	txtPath := strings.Replace(docPath, "pdf", "txt", -1) // -1 means all instances.
	pdf2txt.ConvertToText(docPath)
	fmt.Println("PDF doc was converted to text at path: " + txtPath)
	parasiteInfo := pdf2txt.ExtractParasitesInfo(txtPath)
	fmt.Printf("The info: %v\n", parasiteInfo)
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
