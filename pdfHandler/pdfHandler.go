package pdfhandler

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	pdf2txt "github.com/heussd/pdftotext-go"
)

func ParsePdf(pdfPath string) []pdf2txt.PdfPage {
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		log.Fatalf("Failed to read PDF: %v\n", err)
	}

	/*
		pages = []PdfPage
		type PdfPage struct{
			Content string -> page text content
			Number int -> page number
		}
	*/
	pages, err := pdf2txt.Extract(pdfBytes)
	if err != nil {
		log.Fatalf("Failed to extract text from pages: %v\n", err)
	}

	return pages
}

func ConvertToText(pdfPath string) {
	cmd := exec.Command("bash", "-c", "pdftotext "+pdfPath)
	stdOutErr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(stdOutErr)
		log.Fatalf("Failed to run pdftotext using the path %s, the error was: %v\n", pdfPath, err)
	}
}
