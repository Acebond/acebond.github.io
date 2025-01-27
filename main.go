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
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

var (
	blogTitles = []string{}
)

type Post struct {
	Title string
	Post  template.HTML
}

func reverse(arr []string) {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
}

func get_meta(meta map[string]interface{}, fileName string, key string) string {
	value, defined := meta[key].(string)

	if !defined {
		fmt.Printf("%v is missing from: %v\n", key, fileName)
	}

	return value
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

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			meta.Meta,
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
	context := parser.NewContext()
	err = md.Convert(content, &htmlOutput, parser.WithContext(context))
	if err != nil {
		return err
	}

	metaData := meta.Get(context)
	title := get_meta(metaData, fileName, "title")

	// Add title for index page
	blogTitles = append(blogTitles, fmt.Sprintf("- [%v](%v)", title, "./"+fileName))

	tpl, err := template.ParseFiles("./templates/post.html")
	if err != nil {
		return err
	}

	templateData := Post{
		Title: title,
		Post:  template.HTML(htmlOutput.String()),
	}

	err = tpl.Execute(htmlFile, templateData)
	if err != nil {
		return err
	}

	return nil
}

func GenerateIndexPage() error {

	md := goldmark.New()

	savePath := "./site/index.html"
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

	tpl, err := template.ParseFiles("./templates/post.html")
	if err != nil {
		return err
	}

	templateData := Post{
		Title: "Posts",
		Post:  template.HTML(htmlOutput.String()),
	}

	err = tpl.Execute(htmlFile, templateData)
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
