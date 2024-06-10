package pdfhandler

import (
	"fmt"
	"log"
	"maps"
	"os"
	"os/exec"
	"regexp"

	pdf2txt "github.com/heussd/pdftotext-go"
)

type ParasiteInfo map[string]string

func (p *ParasiteInfo) Init() {
	initialPattern := ParasiteInfo{
		"Scientific Name": "",
		"Common Name":     "",
		"Size":            "",
		"Importance":      "",
		"Diagnosis":       "",
		"Treatment":       "",
	}
	maps.Copy(*p, initialPattern)
}

func ConvertToText(pdfPath string) {
	cmd := exec.Command("bash", "-c", "pdftotext "+pdfPath)
	stdOutErr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(stdOutErr)
		log.Fatalf("Failed to run pdftotext using the path %s, the error was: %v\n", pdfPath, err)
	}
}

func ExtractParasitesInfo(txtPath string) (output []ParasiteInfo) {
	txtBytes, err := os.ReadFile(txtPath)
	if err != nil {
		log.Fatalf("Failed to read txt document: %v\n", err)
	}
	txtString := string(txtBytes)
	pattern := `((.*\n*)*)Fig` // Catch one parasite. Prototype pattern, change in the future.
	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(txtString, -1)
	for _, match := range matches {
		pInfo := ParasiteInfo{}
		pInfo.Init()
		processMatchInfo(&pInfo, match)
		output = append(output, pInfo)
	}
	return output
}

func processMatchInfo(ppInfo *ParasiteInfo, regexpMatch string) {
	/*
		A match is the text with the parasiteInfo keys ("categories") in it.
		We want ot take the content of each category and put it uner the correct key
		in ppInfo - pointer to a parasiteInfo variable.
	*/
	for key, _ := range *ppInfo {
		var start, end string
		body := `((.*|\n)*?)`

		switch key {
		case "Scientific Name":
			start = `^`
			end = `Common`

		case "Common Name":
			start = key
			end = `Size`

		case "Size":
			start = `Common name.*\n`
			end = `Importance`

		case "Importance":
			start = key
			end = `Diagnosis`

		case "Diagnosis":
			start = key
			end = `Treatment`

		// Treatment also includes prevention and notes if these
		// fields exist for this parasite.
		case "Treatment":
			start = key
			body = `((.*|\n)*)` // Without the `prefer fewer`.
			end = `\n`
		}
		pattern := start + body + end
		re := regexp.MustCompile(pattern)
		// Note:
		// categoryRaw = [entire_match capturing_group_1 capturing_group_2 ...]
		categoryRaw := re.FindStringSubmatch(regexpMatch)

		var categoryContent string
		if categoryRaw != nil {
			categoryContent = categoryRaw[1]
		} else {
			categoryContent = ""
		}

		mapy := *ppInfo // Dereferencing the pointer for assignment
		mapy[key] = categoryContent
		*ppInfo = mapy
	}
}

func ParsePdf(pdfPath string) []pdf2txt.PdfPage {
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		log.Fatalf("Failed to read PDF: %v\n", err)
	}

	/*
		Note:
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
