package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

var blogTitles = []string{}

func reverse(arr []string) {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
}

func GenerateBlogPage(file *os.File, path string) error {

	saveDir := "./site/"
	fileName := filepath.Base(path)
	fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))
	fileName = fileName + ".html"

	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	titleStartIndex := 2
	titleEndIndex := bytes.Index(content, []byte{'\r', '\n'})
	title := string(content[titleStartIndex:titleEndIndex])

	// Add title for index page
	blogTitles = append(blogTitles, fmt.Sprintf("- [%v](%v)", title, "/site/"+fileName))

	md := goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithStyle("dracula"),
			),
		),
	)

	htmlFile, err := os.Create(filepath.Join(saveDir, fileName))
	if err != nil {
		return err
	}
	defer htmlFile.Close()

	var htmlOutput bytes.Buffer
	err = md.Convert(content, &htmlOutput)
	if err != nil {
		return err
	}

	tpl, err := template.ParseFiles("post.html")
	if err != nil {
		return err
	}

	err = tpl.Execute(htmlFile, template.HTML(htmlOutput.String()))
	if err != nil {
		return err
	}

	return nil
}

func GenerateIndexPage() error {

	md := goldmark.New()

	savePath := "./index.html"
	htmlFile, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer htmlFile.Close()

	reverse(blogTitles)
	indexMarkdown := strings.Join(blogTitles, "\n")
	var htmlOutput bytes.Buffer
	err = md.Convert([]byte(indexMarkdown), &htmlOutput)
	if err != nil {
		return err
	}

	tpl, err := template.ParseFiles("index.tmpl")
	if err != nil {
		return err
	}

	err = tpl.Execute(htmlFile, template.HTML(htmlOutput.String()))
	if err != nil {
		return err
	}

	return nil
}

func main() {

	postsDir := "./posts/"

	err := filepath.Walk(postsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		return GenerateBlogPage(file, path)
	})

	if err != nil {
		log.Println(err)
	}

	GenerateIndexPage()

}
