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
	Title     string
	Content   template.HTML
	FileType  string
	WordCount int
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// WordCount implementation
func WordCount(s string) int {

	words := strings.Fields(s)
	wordCountMap := make(map[string]int)

	for _, word := range words {
		wordCountMap[word]++
	}

	return len(wordCountMap)
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

	var wordCount int

	path1 := c.Param("path1")
	path2 := c.Param("path2")
	path3 := c.Param("path3")
	path4 := c.Param("path4")
	fileType := "File"

	fileName := e.dest + "/" + path1
	fileTitle := "Home"

	if path1 != "" {
		fileName = path1
		fileTitle = path1
	}

	if path2 != "" {
		fileName += "/" + path2
		fileTitle = path2
	}

	if path3 != "" {
		fileName += "/" + path3
		fileTitle = path3
	}

	if path4 != "" {
		fileName += "/" + path4
		fileTitle = path4
	}

	content := template.HTML("")

	fi, err := os.Stat(fileName)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(fileName)

	switch mode := fi.Mode(); {
	case mode.IsDir():
		fileType = "Directory"

		files, err := ioutil.ReadDir(fileName)
		if err != nil {
			fmt.Println(err)
			c.HTML(http.StatusNotFound, "error.html")
			return err
		}

		for _, file := range files {
			fmt.Println(file.Name())

			if !strings.HasPrefix(file.Name(), ".") {
				itemSlug := strings.Replace(fileName+"/"+file.Name(), e.dest, "", -1)

				if fileName == "/" {
					itemSlug = strings.Replace(file.Name(), "//", "/", -1)
				}

				var itemWordCount int

				itemFileType := "File"
				switch mode := fi.Mode(); {
				case mode.IsDir():
					itemFileType = "Directory"
				case mode.IsRegular():
					file, err := ioutil.ReadFile(file.Name())

					if err != nil {
						fmt.Println(err)
						c.HTML(http.StatusNotFound, "error.html")
						return err
					}

					itemWordCount = WordCount(string(file))
				}

				items = append(items, Item{Slug: itemSlug, Title: file.Name(), Content: content, FileType: itemFileType, WordCount: itemWordCount})
			}
		}
	case mode.IsRegular():
		file, err := ioutil.ReadFile(fileName)

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
	}

	item := Item{Title: fileTitle, Content: content, FileType: fileType, WordCount: wordCount}

	return c.Render(http.StatusOK, "notes.html", map[string]interface{}{
		"Items":     items,
		"Title":     item.Title,
		"Content":   item.Content,
		"FileType":  item.FileType,
		"WordCount": item.WordCount,
	})
}

// Process is the main function for Gin
func Process(dest string) error {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())

	e.Use(middleware.Recover())

	// e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
	// 	Root:   "assets",
	// 	Browse: false,
	// }))

	// Templates
	t := &Template{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}

	e.Renderer = t

	// Routing
	env := &Env{dest: dest}
	e.GET("/", env.GetIndex)
	e.GET("/notes/*", env.GetNotes)

	e.Logger.Fatal(e.Start(":8000"))

	return nil
}
