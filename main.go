package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

var (
	blogTitles = []string{}
	blogURLS   = []string{}
	baseURL    = "https://acebond.github.io"
	basePath   = "./site/"
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

func GenerateBlogPage(file *os.File, path string, info os.FileInfo) error {
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
	if !strings.HasPrefix(info.Name(), "HIDDEN") {
		blogTitles = append(blogTitles, fmt.Sprintf("- [%v](%v)", title, "/"+fileName))
		blogURLS = append(blogURLS, baseURL+"/"+fileName)
	}

	tpl, err := template.ParseFiles("./page.template.html")
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

func GeneratePage(content string, saveName string, title string) error {
	md := goldmark.New()

	savePath := path.Join(basePath, saveName)
	htmlFile, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer htmlFile.Close()

	var htmlOutput bytes.Buffer
	err = md.Convert([]byte(content), &htmlOutput)
	if err != nil {
		return err
	}

	tpl, err := template.ParseFiles("./page.template.html")
	if err != nil {
		return err
	}

	templateData := Post{
		Title: title,
		Post:  template.HTML(htmlOutput.String()),
	}

	return tpl.Execute(htmlFile, templateData)
}

func GenerateSitemap() error {
	savePath := "./site/sitemap.xml"
	f, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer f.Close()

	now := time.Now().Format("2006-01-02")

	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	sb.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	// Add index page
	sb.WriteString("  <url>\n")
	sb.WriteString("    <loc>" + baseURL + "/</loc>\n")
	sb.WriteString("    <lastmod>" + now + "</lastmod>\n")
	sb.WriteString("  </url>\n")

	for _, url := range blogURLS {
		sb.WriteString("  <url>\n")
		sb.WriteString("    <loc>" + url + "</loc>\n")
		sb.WriteString("    <lastmod>" + now + "</lastmod>\n")
		sb.WriteString("  </url>\n")
	}

	sb.WriteString(`</urlset>` + "\n")
	_, err = f.WriteString(sb.String())
	return err
}

func main() {
	toolsPageMarkdown := `
### [ReverseSocks5](https://github.com/Acebond/ReverseSocks5)
Single executable reverse SOCKS5 proxy written in Golang.

### [PPLKiller](https://github.com/RedCursorSecurityConsulting/PPLKiller)
PPLKiller leverages a trusted MSI driver to disable LSA Protection; allowing credentials to be dumped from memory. The tool supports removing the Protected Process Light (PPL) attributes from any process and manipulating process tokens.

### [GhostWrite64](https://github.com/Acebond/GhostWrite64)
Process injection technique that uses only Get/SetThreadContext.

### [CoughLoader](https://github.com/Acebond/CoughLoader)
CoughLoader is a COFF loader, allowing malware to load and execute their COFFs in memory for educational purposes.

### [CLRHost](https://github.com/Acebond/CLRHost)
Host the .NET CLR in native code.

### [GoPastAV](https://github.com/Acebond/GoPastAV)
Generate tailored MSBuild compatible project files for executing shellcode on endpoints with application whitelisting solutions. The tool employs a number of antivirus and EDR evasion techniques, including encryption of the shellcode, sandbox detection, environment keying and multiple shellcode injection methods.

### [NTFSCopy](https://github.com/RedCursorSecurityConsulting/NTFSCopy)
An execute-assembly compatible tool that can copy in-use files such as ntds.dit using NTFS structure parsing. The tool simply a wrapper for [NtfsLib](https://github.com/LordMike/NtfsLib).

### [LSASecretsTool](https://github.com/Acebond/LSASecretsTool)
An execute-assembly compatible tool for manipulating LSA secrets using the undocumented but official LSASS API calls. This includes reading, writing, creating and deleting LSA secrets.

### [CVE-2020-0668](https://github.com/RedCursorSecurityConsulting/CVE-2020-0668)
Implementation of CVE-2020-0668 which leverages symbolic links to perform a privileged file move operation that can lead to privilege escalation on all versions of Windows from Vista to 10, including server editions.

### [SharpHashSpray](https://github.com/RedCursorSecurityConsulting/SharpHashSpray)
An execute-assembly compatible tool for spraying local admin hashes (NTLM). By default the tool will automatically fetch a list of all domain joined hosts to check. Alternatively a target range can be provided.

### [GetAdDecodedPassword](https://github.com/RedCursorSecurityConsulting/GetAdDecodedPassword)
This tool queries Active Directory for users with the UnixUserPassword, UserPassword, unicodePwd, or msSFU30Password properties populated. It then decodes those password fields and displays them to the user.
`

	whoamiPageMarkdown := `Acebond`

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

		return GenerateBlogPage(file, path, info)
	})

	if err != nil {
		log.Println(err)
	}

	reverse(blogTitles)
	indexMarkdown := strings.Join(blogTitles, "\n")

	if err := GeneratePage(indexMarkdown, "index.html", "Posts"); err != nil {
		log.Println("Error generating index:", err)
	}

	if err := GenerateSitemap(); err != nil {
		log.Println("Error generating sitemap:", err)
	}

	if err := GeneratePage(toolsPageMarkdown, "tools.html", "Tools"); err != nil {
		log.Println("Error generating tools page:", err)
	}

	if err := GeneratePage(whoamiPageMarkdown, "whoami.html", "whoami"); err != nil {
		log.Println("Error generating whoami page:", err)
	}

}
