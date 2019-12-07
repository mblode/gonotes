package server

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
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

// WordCount implementation
func WordCount(s string) int {

	words := strings.Fields(s)
	wordCountMap := make(map[string]int)

	for _, word := range words {
		wordCountMap[word]++
	}

	return len(wordCountMap)
}

// // GetIndex is for the index
// func (e *Env) GetIndex(c *gin.Context) {
// 	var items []Item

// 	files, err := ioutil.ReadDir(e.dest)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	for _, file := range files {
// 		if !strings.HasPrefix(file.Name(), ".") {
// 			items = append(items, Item{Slug: file.Name(), Title: file.Name()})
// 		}
// 	}

// 	c.HTML(http.StatusOK, "index.html", gin.H{
// 		"Title": "Home",
// 		"Items": items,
// 	})
// }

// GetNotes is for the notes
func (e *Env) GetNotes(c *gin.Context) {
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
			c.HTML(http.StatusNotFound, "error.html", nil)
			return
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
						c.HTML(http.StatusNotFound, "error.html", nil)
						return
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
			c.HTML(http.StatusNotFound, "error.html", nil)
			return
		}

		wordCount = WordCount(string(file))

		var buf bytes.Buffer
		if err := goldmark.Convert(file, &buf); err != nil {
			panic(err)
		}

		content = template.HTML(buf.String())
	}

	item := Item{Title: fileTitle, Content: content, FileType: fileType, WordCount: wordCount}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"Items":     items,
		"Title":     item.Title,
		"Content":   item.Content,
		"FileType":  item.FileType,
		"WordCount": item.WordCount,
	})
}

// Process is the main function for Gin
func Process(dest string) error {
	router := gin.Default()
	router.Use(gin.Logger())
	router.Delims("{{", "}}")

	router.Use(static.Serve("/assets", static.LocalFile("/assets", false)))
	router.LoadHTMLGlob("./templates/*.html")

	env := &Env{dest: dest}
	router.GET("/", env.GetNotes)
	router.GET("/:path1", env.GetNotes)
	router.GET("/:path1/:path2", env.GetNotes)
	router.GET("/:path1/:path2/:path3", env.GetNotes)
	router.GET("/:path1/:path2/:path3/:path4", env.GetNotes)

	router.Run()

	return nil
}
