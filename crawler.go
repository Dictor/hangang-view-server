package main

import (
	"context"
	"fmt"

	"github.com/chromedp/chromedp"
)

var (
	globalChromeContext context.Context
	globalChromeCancler context.CancelFunc
)

type Price struct {
	Price string
}

func InitChrome() {
	globalChromeContext, globalChromeCancler = chromedp.NewContext(context.Background())
}

func GetPrice(isWorldwide bool, symbolName string) (Price, error) {
	var (
		positionDirectory string
		result            Price
	)

	if isWorldwide {
		positionDirectory = "worldwide"
	} else {
		positionDirectory = "domestic"
	}
	url := fmt.Sprintf("https://m.stock.naver.com/%s/stock/%s/total", positionDirectory, symbolName)

	err := chromedp.Run(globalChromeContext,
		chromedp.Navigate(`https://pkg.go.dev/time`),
		chromedp.Text(`.Documentation-overview`, &res, chromedp.NodeVisible),
	)
}
