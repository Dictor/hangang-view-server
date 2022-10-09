package main

import (
	"net/http"
	"time"

	elogrus "github.com/dictor/echologrus"
	"github.com/labstack/echo/v4"
)

var (
	SymbolList   []*Symbol = []*Symbol{{Kind: "indices", Name: "nq-100-futures"}}
	GlobalLogger elogrus.EchoLogger
)

func main() {
	e := echo.New()
	GlobalLogger = elogrus.Attach(e)
	InitChrome()
	go UpdateSymbolTask(SymbolList) // TODO: in this method, we cannot dynamically control this list. List copied local variable in routine

	e.GET("/symbol", func(c echo.Context) error {
		return c.JSON(http.StatusOK, SymbolList)
	})
	e.Logger.Fatal(e.Start(":80"))
}

func UpdateSymbolTask(list []*Symbol) {
	GlobalLogger.Debugln("start symbol update task")
	for {
		success := 0
		for _, s := range list {
			if err := GetPriceBySymbol(s); err == nil {
				success++
			} else {
				GlobalLogger.WithError(err).Error("symbol update fail")
			}
		}
		GlobalLogger.Debugf("symbol updated, total %d, success %d", len(list), success)
		time.Sleep(10 * time.Second)
	}
}
