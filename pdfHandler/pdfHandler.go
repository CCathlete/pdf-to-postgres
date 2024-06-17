package pdfhandler

import (
	"fmt"
	"log"
	"maps"
	"os"
	"os/exec"
	"regexp"
	"strings"

	pdf2txt "github.com/heussd/pdftotext-go"
	license "github.com/unidoc/unipdf/v3/common/license"
	extractor "github.com/unidoc/unipdf/v3/extractor"
	model "github.com/unidoc/unipdf/v3/model"
)

type ParasiteInfo map[string]string

func (p *ParasiteInfo) Init() {
	initialPattern := ParasiteInfo{
		"Scientific name": "",
		"Common name":     "",
		"Size":            "",
		"Importance":      "",
		"Diagnosis":       "",
		"Treatment":       "",
	}
	maps.Copy(*p, initialPattern)
}

func ConvertToText_crap(pdfPath string) {
	cmd := exec.Command("bash", "-c", "pdftotext "+pdfPath)
	stdOutErr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(stdOutErr)
		log.Fatalf("Failed to run pdftotext using the path %s, the error was: %v\n", pdfPath, err)
	}
}

func ConvertToText(pdfPath string) {
	err := license.SetMeteredKey(os.Getenv(`UNIDOC_LICENSE_API_KEY`))
	if err != nil {
		panic(err)
	}
	f, err := os.Open(pdfPath)
	if err != nil {
		log.Fatalf("Failed to open the file in path %s, the error was: %v\n", pdfPath, err)
	}

	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		log.Fatalf("Failed to createnannew pdf reader for path %s, the error was: %v\n", pdfPath, err)
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		log.Fatalf("Failed to get number of pages, the error was: %v\n", err)
	}

	fmt.Printf("--------------------\n")
	fmt.Printf("PDF to text extraction:\n")
	fmt.Printf("--------------------\n")
	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			log.Fatalf("Failed to get a page pbject for page num %d, the error was: %v\n", pageNum, err)
		}

		ex, err := extractor.New(page)
		if err != nil {
			log.Fatalf("Failed to create a new extractor for page num %d, the error was: %v\n", pageNum, err)
		}

		text, err := ex.ExtractText()
		if err != nil {
			log.Fatalf("Failed to extract text from page num %d, the error was: %v\n", pageNum, err)
		}

		txtPath := strings.Replace(pdfPath, "pdf", "txt", -1) // -1 means all instances.
		file, err := os.OpenFile(txtPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
		if err != nil {
			log.Fatalf("Failed to a file in path %s, the error was: %v\n", txtPath, err)
		}
		defer file.Close()

		_, err = file.WriteString(text)
		if err != nil {
			log.Fatalf("Failed to write the text into a txt file, the error was: %v\n", err)
		}

		// fmt.Println("------------------------------")
		// fmt.Printf("Page %d:\n", pageNum)
		// fmt.Printf("\"%s\"\n", text)
		// fmt.Println("------------------------------")
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
		case "Scientific name":
			start = `^`
			end = `Common`

		case "Common name":
			start = fmt.Sprintf("%s: ", key)
			end = `Size`

		case "Size":
			start = `Common name.*\n`
			end = `Importance`

		case "Importance":
			start = fmt.Sprintf("%s: ", key)
			end = `Diagnosis`

		case "Diagnosis":
			start = fmt.Sprintf("%s: ", key)
			end = `Treatment`

		// Treatment also includes prevention and notes if these
		// fields exist for this parasite.
		case "Treatment":
			start = fmt.Sprintf("%s: ", key)
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
