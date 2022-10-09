package main

import (
	"net/http"

	elogrus "github.com/dictor/echologrus"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	elogrus.Attach(e)
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Logger.Fatal(e.Start(":80"))
}
