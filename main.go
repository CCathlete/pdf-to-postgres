package main

import (
	"log"
	"os"
	pdf2pg "pdf-to-postgres/pdfHandler"
	"strings"
)

func main() {
	docPath := "Parasitology/1 - Introduction to helminthology.pdf"
	pages := pdf2pg.ParsePdf(docPath)
	txtPath := strings.Replace(docPath, "pdf", "txt", -1) // -1 means all instances.

	file, err := os.OpenFile(txtPath, os.O_CREATE|os.O_APPEND, 0644)
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
		file.WriteString(page.Content)
	}
}
