package main

import (
	pdf2pg "pdf-to-postgres/pdfHandler"
)

func main() {
	docPath := "Parasitology/1 - Introduction to helminthology.pdf"
	pages := pdf2pg.ParsePdf(docPath)
	for _, page := range pages {
		println(page.Content)
	}
}
