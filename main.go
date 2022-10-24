package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/mochi-co/mqtt/server"
	"github.com/mochi-co/mqtt/server/events"
	"github.com/mochi-co/mqtt/server/listeners"
	"github.com/mochi-co/mqtt/server/listeners/auth"
	"github.com/namsral/flag"
	"github.com/sirupsen/logrus"
)

var (
	SymbolList   []Symbol = []Symbol{{Kind: "indices", Name: "nq-100-futures"}}
	GlobalLogger *logrus.Logger
)

func main() {
	var (
		port                int
		id                  string
		symbolPublishBridge chan SymbolTopic = make(chan SymbolTopic, 100)
	)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	GlobalLogger = logrus.New()
	GlobalLogger.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	GlobalLogger.SetLevel(logrus.DebugLevel)

	flag.StringVar(&id, "id", "chinchister", "mqtt broker id")
	flag.IntVar(&port, "port", 1888, "mqtt listening port")
	flag.Parse()

	GlobalLogger.Infof("start MQTT server at port %d, id %s", port, id)
	mqttServer := mqtt.NewServer(nil)
	mqttListner := listeners.NewTCP(id, fmt.Sprintf(":%d", port))
	err := mqttServer.AddListener(mqttListner, &listeners.Config{
		Auth: new(auth.Allow),
	})
	if err != nil {
		GlobalLogger.WithError(err).Error("failed to initiate MQTT server")
		return
	}

	mqttServer.Events.OnDisconnect = func(cl events.Client, err error) {
		GlobalLogger.WithError(err).Infof("cliend disconnect : %v", cl)
	}

	mqttServer.Events.OnConnect = func(cl events.Client, p events.Packet) {
		GlobalLogger.Infof("cliend connect : %v", cl)
	}

	mqttServer.Events.OnError = func(cl events.Client, err error) {
		GlobalLogger.WithError(err).Errorf("cliend error : %v", cl)
	}

	InitChrome()
	go UpdateSymbolTask(SymbolList, symbolPublishBridge) // TODO: in this method, we cannot dynamically control this list. List copied local variable in routine
	go PublishSymbolTask(mqttServer, symbolPublishBridge)

	if err = mqttServer.Serve(); err != nil {
		GlobalLogger.Fatal(err)
	}
	<-sigs
	GlobalLogger.Info("halted by sig")
}

func UpdateSymbolTask(list []Symbol, symbolChan chan<- SymbolTopic) {
	GlobalLogger.Debugln("start symbol update task")
	for {
		success := 0
		for _, s := range list {
			if res, err := GetPriceBySymbol(s); err == nil {
				success++
				symbolChan <- res
			} else {
				GlobalLogger.WithError(err).Error("symbol update fail")
			}
		}
		GlobalLogger.Debugf("symbol updated, total %d, success %d", len(list), success)
		time.Sleep(10 * time.Second)
	}
}

func PublishSymbolTask(server *mqtt.Server, symbolChan <-chan SymbolTopic) {
	var (
		payload       []byte
		err           error
		results       map[string]SymbolTopic = map[string]SymbolTopic{}
		publishTicker *time.Ticker           = time.NewTicker(10 * time.Second)
	)

	for {
		select {
		case s := <-symbolChan:
			results[s.Name] = s
		case <-publishTicker.C:
			resultToList := []SymbolTopic{}
			for _, v := range results {
				resultToList = append(resultToList, v)
			}

			if payload, err = json.Marshal(&resultToList); err != nil {
				GlobalLogger.WithError(err).Error("fail to marshar symbol results")
				continue
			}
			err := server.Publish("symbol", []byte(payload), false)
			if err != nil {
				GlobalLogger.WithError(err).Error("MQTT publish failure")
			}
			GlobalLogger.Debug("symbol published")
		}
	}

}
