package main

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/sirupsen/logrus"
)

var (
	globalChromeContext context.Context
	globalChromeCancler context.CancelFunc
)

type (
	Symbol struct {
		Kind          string    `json:"kind"`
		Name          string    `json:"name"`
		Result        *Price    `json:"result"`
		LatestUpdated time.Time `json:"latest_updated"`
	}
	Price struct {
		Price      string `json:"price"`
		Percentile string `json:"percentile"`
	}
)

func InitChrome() {
	globalChromeContext, globalChromeCancler = chromedp.NewContext(context.Background())
}

func GetPrice(symbolKind string, symbolName string) (Price, error) {
	var (
		result Price
	)

	url := fmt.Sprintf("https://kr.investing.com/%s/%s", symbolKind, symbolName)
	err := chromedp.Run(globalChromeContext,
		chromedp.Navigate(url),
		chromedp.Text(`[data-test="instrument-price-last"]`, &result.Price, chromedp.NodeVisible),
		chromedp.Text(`[data-test="instrument-price-change-percent"]`, &result.Percentile, chromedp.NodeVisible))

	GlobalLogger.WithFields(logrus.Fields{"kind": symbolKind,
		"name":       symbolName,
		"price":      result.Price,
		"percentile": result.Percentile,
	}).Debug("GetPrice")
	return result, err
}

func GetPriceBySymbol(symbol *Symbol) error {
	res, err := GetPrice(symbol.Kind, symbol.Name)
	if err == nil {
		symbol.Result = &res
		symbol.LatestUpdated = time.Now()
	}
	return err
}
