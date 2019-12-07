package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/yuin/goldmark"
)

// Post struct for markdown
type Post struct {
	Slug    string
	Title   string
	Content template.HTML
}

// Directory struct for markdown
type Directory struct {
	Slug  string
	Title string
}

// CopyDirectory will copy the contents of a directory
func CopyDirectory(src, dest string) error {
	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CreateIfNotExists(destPath, os.ModePerm); err != nil {
				return err
			}
			if err := CopyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := CopySymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := Copy(sourcePath, destPath); err != nil {
				return err
			}
		}

		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return err
		}

		isSymlink := entry.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, entry.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

// Copy will copy the file
func Copy(srcFile, destFile string) error {
	out, err := os.Create(destFile)
	defer out.Close()
	if err != nil {
		return err
	}

	in, err := os.Open(srcFile)
	defer in.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

// Exists will chek if the file exists
func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

// CreateIfNotExists will create a directory
func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

// CopySymLink will copy sym link files
func CopySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	// srcDir := "/Users/mblode/Library/Mobile Documents/27N4MQEA55~pro~writer/Documents/"
	srcDir := "/Users/mblode/Google Drive/Backups/Notes"
	destDir := "/Users/mblode/Google Drive/Backups/Notes2"

	src := flag.String("src", srcDir, "The folder where the files should be copied from")
	dest := flag.String("dest", destDir, "The folder that the files should be copied to")
	flag.Parse()

	if err := CreateIfNotExists(*dest, os.ModePerm); err != nil {
		fmt.Println(err)
	}

	err := CopyDirectory(*src, *dest)
	check(err)

	fmt.Println("Successfully copied files")

	r := gin.Default()
	r.Use(gin.Logger())
	r.Delims("{{", "}}")

	r.Use(static.Serve("/assets", static.LocalFile("/assets", false)))
	r.LoadHTMLGlob("./templates/*.html")

	r.GET("/", func(c *gin.Context) {
		var directories []Directory

		files, err := ioutil.ReadDir(destDir)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			if !strings.HasPrefix(file.Name(), ".") {
				directories = append(directories, Directory{Slug: file.Name(), Title: file.Name()})
			}
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"directories": directories,
		})
	})

	r.GET("/:dirName", func(c *gin.Context) {
		var posts []Post

		dirName := c.Param("dirName")

		files, err := ioutil.ReadDir(destDir + "/" + dirName)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			if !strings.HasPrefix(file.Name(), ".") {
				posts = append(posts, Post{Slug: dirName + "/" + file.Name(), Title: file.Name()})
			}
		}

		// if the folder can not be found
		if err != nil {
			fmt.Println(err)
			c.HTML(http.StatusNotFound, "error.html", nil)
			return
		}

		dir := Directory{Slug: dirName, Title: dirName}

		c.HTML(http.StatusOK, "directory.html", gin.H{
			"ParentSlug": dirName,
			"Slug":       dir.Slug,
			"Title":      dir.Title,
			"posts":      posts,
		})
	})

	r.GET("/:dirName/:postName", func(c *gin.Context) {
		dirName := c.Param("dirName")
		postName := c.Param("postName")

		mdfile, err := ioutil.ReadFile(destDir + "/" + dirName + "/" + postName)

		// if the file can not be found
		if err != nil {
			fmt.Println(err)
			c.HTML(http.StatusNotFound, "error.html", nil)
			return
		}

		var buf bytes.Buffer
		if err := goldmark.Convert(mdfile, &buf); err != nil {
			panic(err)
		}

		result := template.HTML(buf.String())

		post := Post{Slug: postName, Title: postName, Content: result}

		c.HTML(http.StatusOK, "post.html", gin.H{
			"Slug":    post.Slug,
			"Title":   post.Title,
			"Content": post.Content,
		})
	})

	r.Run()
}
