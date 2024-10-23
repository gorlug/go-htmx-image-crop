package main

import (
	"errors"
	"fmt"
	. "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"
)

const imagePath = "upload/image.jpg"

func main() {
	e := New()
	e.Use(middleware.Logger())

	e.GET("/", getIndex)
	e.POST("/upload", uploadForm)
	e.GET("/image", getImage)
	e.GET("/crop", getCrop)
	e.GET("/crop/cancel", getCropCancel)
	e.PUT("/crop", cropImage)

	e.Logger.Fatal(e.Start(":8080"))
}

type ImageDisplayData struct {
	// This is appended so that a reload of the images is triggered after cropping
	DateQueryParam string
}

func getIndex(c Context) error {
	return RenderGoHtml(c, "index", ImageDisplayData{})
}

func RenderGoHtml(c Context, name string, data any) error {
	tmpl := template.Must(template.ParseGlob("views/*.gohtml"))
	return tmpl.ExecuteTemplate(c.Response().Writer, name, data)
}

func uploadForm(c Context) error {
	err := saveImageFromForm(c, "image", imagePath)
	if err != nil {
		println("Failed to save image from form with error", err)
	}
	return getIndex(c)
}

func saveImageFromForm(c Context, formName string, targetPath string) error {
	file, err := c.FormFile(formName)
	if errors.Is(err, http.ErrMissingFile) {
		println("No file uploaded")
		return err
	}
	if err != nil {
		println("Failed to save image from input with error", err)
		return err
	}
	src, err := file.Open()
	if err != nil {
		println("Failed to open the image file with error", err)
		return err
	}
	defer src.Close()

	dst, err := os.Create(targetPath)
	if err != nil {
		println("Failed to create the target path with error", err)
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		println("Failed to copy the image file with error", err)
		return err
	}
	return nil
}

/**
 * Returns the image. This way we have full caching control.
 * In this case we don't want any caching.
 */
func getImage(c Context) error {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return err
	}
	mimeType := http.DetectContentType(data)
	return c.Blob(http.StatusOK, mimeType, data)
}

func cropImage(c Context) error {
	err := saveImageFromForm(c, "image", imagePath)
	if err != nil {
		println("Failed to save image from form with error", err)
	}
	return RenderGoHtml(c, "imageDisplay", ImageDisplayData{
		DateQueryParam: fmt.Sprint("?date=", time.Now().Unix()),
	})
}

func getCropCancel(c Context) error {
	return RenderGoHtml(c, "imageDisplay", ImageDisplayData{})
}

func getCrop(c Context) error {
	return RenderGoHtml(c, "cropImage", nil)
}
