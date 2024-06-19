package pdfhandler

import (
	"fmt"
	"image/jpeg"
	"log"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
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

func ConvertToTextWhenNotScanned(pdfPath string) {
	cmd := exec.Command("bash", "-c", "pdftotext "+pdfPath)
	stdOutErr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(stdOutErr)
		log.Fatalf("Failed to run pdftotext using the path %s, the error was: %v\n", pdfPath, err)
	}
}

func runTesseractOCR(imagepath string) string {
	tempOutFile := "output" // Tesseract automatically adds .txt
	cmd := exec.Command("/bin/tesseract", imagepath, tempOutFile)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to convert image with tesseract %v\n", err)
		return ""
	}

	text, err := os.ReadFile(tempOutFile + ".txt")
	if err != nil {
		log.Fatalf("Failed to convert image with tesseract %v\n", err)
		return ""
	}

	os.Remove(tempOutFile + ".txt")

	return string(text)
}

func ImagesToText(inputDir, pdfPath string) {
	files, err := os.ReadDir(inputDir)
	if err != nil {
		log.Fatalf("Failed to read directory %s: %v\n", inputDir, err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".jpg" || filepath.Ext(file.Name()) == ".jpeg" {
			imagePath := filepath.Join(inputDir, file.Name())
			fmt.Printf("Using tesseract on image %s.\n", imagePath)

			text := runTesseractOCR(imagePath)

			// Adding the text to the txt file.
			txtPath := strings.Replace(pdfPath, "pdf", "txt", -1) // -1 means all instances.
			file, err := os.OpenFile(txtPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
			if err != nil {
				log.Fatalf("Failed to a file in path %s, the error was: %v\n", txtPath, err)
			}
			defer file.Close()

			numOfBits, err := file.WriteString(text)
			if err != nil {
				log.Fatalf("Failed to write the text into a txt file, the error was: %v\n", err)
			} else {
				fmt.Printf("%d bits were written.\n", numOfBits)
			}

			if err := file.Sync(); err != nil {
				log.Fatalf("Failed to sync file: %v", err)
			}
		}
	}
}

func splitTextByAnimals(text string, animals []string) map[string]string {
	splittedText := map[string]string{}
	pattern := ""

	for i, animal := range animals {
		if i < len(animals)-1 {
			nextAnimal := animals[i+1]
			// Note: for extra security it's better to use regexp.QuteMeta(animal).
			// it makes sure to treat the animal names as raw strings with no special characters.
			pattern = fmt.Sprintf(`(?s)%s.*?(?=%s|$)`, animal, nextAnimal)
		} else {
			// There is no next animal for the last animal.
			pattern = fmt.Sprintf(`(?s)%s.*`, animal)
		}
		re := regexp.MustCompile(pattern)
		match := re.FindString(text)
		if match != "" {
			splittedText[animal] = match
		}
	}

	return splittedText
}

func ExtractParasitesInfo(txtPath string, animals []string) map[string][]ParasiteInfo {
	/*
		The function returns a map with names of animals and a slice of parasites for each animal.
	*/
	// output := make(map[string][]ParasiteInfo)
	output := map[string][]ParasiteInfo{}
	txtBytes, err := os.ReadFile(txtPath)
	if err != nil {
		log.Fatalf("Failed to read txt document: %v\n", err)
	}
	txtString := string(txtBytes)

	splittedByAnimals := splitTextByAnimals(txtString, animals)

	pattern := `((.*\n*)*)Fig` // Catch one parasite. Prototype pattern, change in the future.
	re := regexp.MustCompile(pattern)
	for _, animal := range animals {
		matches := re.FindAllString(splittedByAnimals[animal], -1)
		for _, match := range matches {
			pInfo := ParasiteInfo{}
			pInfo.Init()
			processMatchInfo(&pInfo, match)
			output[animal] = append(output[animal], pInfo)
		}
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

func ConvertWithUniPDF(pdfPath, outputDirPath string) {
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

		images, err := ex.ExtractPageImages(nil)
		if err != nil {
			log.Fatalf("Failed to extract an image from page num %d, the error was: %v\n",
				pageNum, err)
		}

		for j, image := range images.Images {
			// Extracting text from the am image with OCR.
			imagePath := fmt.Sprintf("%s/page_%d_image_%d.jpg", outputDirPath, i, j)
			imageFile, err := os.Create(imagePath)
			if err != nil {
				log.Fatalf("Failed to create a new extractor for page num %d, the error was: %v\n", pageNum, err)
			}
			defer imageFile.Close()

			goImage, err := image.Image.ToGoImage()
			if err != nil {
				log.Fatalf("Failed to convert to Go image %d on page %d: %v", j, i, err)
			}

			err = jpeg.Encode(imageFile, goImage, nil)
			if err != nil {
				log.Fatalf("Failed to encode image %d on page %d: %v", j, i, err)
			}

			// Adding the text to the txt file.
			txtPath := strings.Replace(pdfPath, "pdf", "txt", -1) // -1 means all instances.
			file, err := os.OpenFile(txtPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
			if err != nil {
				log.Fatalf("Failed to a file in path %s, the error was: %v\n", txtPath, err)
			}
			defer file.Close()

			text := runTesseractOCR(imagePath)
			numOfBits, err := file.WriteString(text)
			if err != nil {
				log.Fatalf("Failed to write the text into a txt file, the error was: %v\n", err)
			} else {
				fmt.Printf("%d bits were written.\n", numOfBits)
			}

			if err := file.Sync(); err != nil {
				log.Fatalf("Failed to sync file: %v", err)
			}
		}

		// fmt.Println("------------------------------")
		// fmt.Printf("Page %d:\n", pageNum)
		// fmt.Printf("\"%s\"\n", text)
		// fmt.Println("------------------------------")
	}
}
