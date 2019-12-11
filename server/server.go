package server

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/yuin/goldmark"
)

// Env variables for Gin
type Env struct {
	dest string
}

// Item struct for markdown
type Item struct {
	Slug      string
	Path      string
	Title     string
	Content   template.HTML
	FileType  string
	WordCount int
}

// Breadcrumb struct for items
type Breadcrumb struct {
	Title string
	Slug  string
}

// Template struct for HTML
type Template struct {
	templates *template.Template
}

// Render the HTML template
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// WordCount implementation
func WordCount(str string) int {
	// Match non-space character sequences.
	re := regexp.MustCompile(`[\S]+`)

	// Find all matches and return count.
	results := re.FindAllString(str, -1)
	return len(results)
}

// GetIndex is for the index
func (e *Env) GetIndex(c echo.Context) error {
	var items []Item

	files, err := ioutil.ReadDir(e.dest)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if !strings.HasPrefix(file.Name(), ".") {
			items = append(items, Item{Slug: "notes/" + file.Name(), Title: file.Name()})
		}
	}

	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Title": "Home",
		"Items": items,
	})
}

// GetNotes is for the notes
func (e *Env) GetNotes(c echo.Context) error {
	var items []Item
	var breadcrumbs []Breadcrumb

	var wordCount int
	param := c.Param("*")
	filePath := e.dest + "/" + param
	fileType := "File"
	content := template.HTML("")
	fileTitle := filepath.Base(filePath)
	fileExtension := filepath.Ext(filePath)
	fileTitle = fileTitle[0 : len(fileTitle)-len(fileExtension)]

	breadcrumbSplit := strings.Split(param, "/")
	breadcrumbPath := "/notes"
	breadcrumbs = append(breadcrumbs, Breadcrumb{Title: "Home", Slug: breadcrumbPath})
	breadcrumbPath = breadcrumbPath + "/"

	for index, item := range breadcrumbSplit {
		if item != "" {
			if index == len(breadcrumbSplit)-1 {
				breadcrumbs = append(breadcrumbs, Breadcrumb{Title: item, Slug: ""})
				breadcrumbPath += item + "/"
			} else {
				breadcrumbs = append(breadcrumbs, Breadcrumb{Title: item, Slug: breadcrumbPath + item})
				breadcrumbPath += item + "/"
			}
		}
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Println(err)
	}

	switch fileInfo.Mode() & os.ModeType {
	case os.ModeDir:
		fileType = "Directory"

		files, err := ioutil.ReadDir(filePath)
		if err != nil {
			fmt.Println(err)
			c.HTML(http.StatusNotFound, "error.html")
			return err
		}

		for _, file := range files {
			if !strings.HasPrefix(file.Name(), ".") {
				var itemWordCount int
				itemPath := filePath + "/" + file.Name()
				itemSlug := strings.Replace(itemPath, e.dest, "", -1)
				itemSlug = strings.Replace(itemSlug, "//", "/", -1)
				itemFileType := "File"

				itemFileInfo, err := os.Stat(itemPath)
				if err != nil {
					fmt.Println(err)
				}

				switch mode := itemFileInfo.Mode(); {
				case mode.IsDir():
					itemFileType = "Directory"
				case mode.IsRegular():
					itemFile, err := ioutil.ReadFile(itemPath)

					if err != nil {
						fmt.Println(err)
						c.HTML(http.StatusNotFound, "error.html")
						return err
					}

					itemWordCount = WordCount(string(itemFile))
				}

				items = append(items, Item{Slug: "/notes" + itemSlug, Path: itemPath, Title: file.Name(), Content: content, FileType: itemFileType, WordCount: itemWordCount})
			}
		}

		item := Item{Path: filePath, Title: fileTitle, Content: content, FileType: fileType, WordCount: wordCount}

		return c.Render(http.StatusOK, "directory.html", map[string]interface{}{
			"Items":       items,
			"Path":        item.Path,
			"Title":       item.Title,
			"Content":     item.Content,
			"FileType":    item.FileType,
			"WordCount":   item.WordCount,
			"Breadcrumbs": breadcrumbs,
		})
	default:
		file, err := ioutil.ReadFile(filePath)

		if err != nil {
			fmt.Println(err)
			c.HTML(http.StatusNotFound, "error.html")
			return err
		}

		wordCount = WordCount(string(file))

		var buf bytes.Buffer
		if err := goldmark.Convert(file, &buf); err != nil {
			panic(err)
		}

		content = template.HTML(buf.String())

		item := Item{Path: filePath, Title: fileTitle, Content: content, FileType: fileType, WordCount: wordCount}

		return c.Render(http.StatusOK, "file.html", map[string]interface{}{
			"Items":       items,
			"Path":        item.Path,
			"Title":       item.Title,
			"Content":     item.Content,
			"FileType":    item.FileType,
			"WordCount":   item.WordCount,
			"Breadcrumbs": breadcrumbs,
		})
	}
}

// Process is the main function for Gin
func Process(dest string) error {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Templates
	t := &Template{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}

	e.Renderer = t

	// Routing
	env := &Env{dest: dest}
	e.GET("/", env.GetIndex)
	e.GET("/notes", env.GetNotes)
	e.GET("/notes/*", env.GetNotes)

	e.Logger.Fatal(e.Start(":3000"))

	return nil
}
