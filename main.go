package main

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"os"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

func main() {

	file, err := os.Open("2024-10-10-whoops-i-dropped-my-system-thread-handle.markdown")
	if err != nil {
		log.Fatalf("failed to open markdown file: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("failed to read markdown file: %v", err)
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithStyle("dracula"),
			),
		),
	)

	htmlFile, err := os.Create("output.html")
	if err != nil {
		log.Fatalf("failed to create html file: %v", err)
	}
	defer htmlFile.Close()

	var htmlOutput bytes.Buffer
	err = md.Convert(content, &htmlOutput)
	if err != nil {
		log.Fatalf("failed to convert markdown to HTML: %v", err)
	}

	tpl, err := template.ParseFiles("post.html")
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	err = tpl.Execute(htmlFile, template.HTML(htmlOutput.String()))
	if err != nil {
		log.Fatalf("Error executeing template: %v", err)
	}

	//err = md.Convert(content, htmlFile)
	//if err != nil {
	//	log.Fatalf("failed to convert markdown to HTML: %v", err)
	//}
}
