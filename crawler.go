package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/sirupsen/logrus"
)

type (
	// symbol definition for investing.com requesting
	Symbol struct {
		Kind        string `json:"kind"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
	}

	// Symbol for publishing to MQTT
	SymbolTopic struct {
		Name       string `json:"name"`
		Price      int    `json:"price"`
		Percentile int    `json:"percentile"`
	}

	Price struct {
		Price      float64 `json:"price"`
		Percentile float64 `json:"percentile"`
	}
)

func InitChrome() {
}

func GetPrice(symbolKind string, symbolName string) (Price, error) {
	var (
		result            Price
		price, percentile string
		err               error
		sanitize          func(s string) string = func(s string) string {
			s = strings.ReplaceAll(s, "(", "")
			s = strings.ReplaceAll(s, ")", "")
			s = strings.ReplaceAll(s, ",", "")
			s = strings.ReplaceAll(s, "%", "")
			return s
		}
	)

	parentCtx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	ctx, _ := chromedp.NewContext(parentCtx)

	url := fmt.Sprintf("https://kr.investing.com/%s/%s", symbolKind, symbolName)
	err = chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Text(`[data-test="instrument-price-last"]`, &price, chromedp.NodeVisible),
		chromedp.Text(`[data-test="instrument-price-change-percent"]`, &percentile, chromedp.NodeVisible))
	if err != nil {
		GlobalLogger.WithError(err).Error("fail to crawl information")
		return result, err
	}

	price = sanitize(price)
	percentile = sanitize(percentile)

	if result.Price, err = strconv.ParseFloat(price, 64); err != nil {
		GlobalLogger.WithError(err).WithField("input", price).Error("fail to parse price string")
		return result, err
	}
	if result.Percentile, err = strconv.ParseFloat(percentile, 64); err != nil {
		GlobalLogger.WithError(err).WithField("input", percentile).Error("fail to parse percentile string")
		return result, err
	}

	GlobalLogger.WithFields(logrus.Fields{"kind": symbolKind,
		"name":       symbolName,
		"price":      result.Price,
		"percentile": result.Percentile,
	}).Debug("GetPrice")
	return result, err
}

func GetPriceBySymbol(symbol Symbol) (SymbolTopic, error) {
	price, err := GetPrice(symbol.Kind, symbol.Name)
	res := SymbolTopic{}
	if err == nil {
		// look hangang view firmware's json comment for diving by 100
		res.Price = int(float32(price.Price) * 100)
		res.Percentile = int(float32(price.Percentile) * 100)
		res.Name = symbol.DisplayName
	}
	return res, err
}
