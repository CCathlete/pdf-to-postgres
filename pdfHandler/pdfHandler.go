package pdfhandler

import (
	"fmt"
	"log"
	"os"

	"github.com/unidoc/unipdf/v3/model"
)

func ParsePdf(pdfPath string) {
	file, err := os.Open(pdfPath)
	if err != nil {
		log.Fatalf("Failed to open PDF: %v\n", err)
	}
	defer file.Close()
	pdfReader, err := model.NewPdfReader(file)
	if err != nil {
		log.Fatalf("Failed to read PDF: %v\n", err)
	}
	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		log.Fatalf("Failed to retrieve the number of pages: %v\n", err)
	}
	fmt.Println("The total number of pages is: ", numPages)
}
