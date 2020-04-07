package binancewebsocket

import (
	"github.com/json-iterator/go"
	"github.com/alexey-ernest/go-binance-websocket/ws"
	"log"
	"fmt"
	"time"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type binanceWs struct {
	baseURL string
	Conn *ws.WsConn
}

func NewBinanceWs() *binanceWs {
	bnWs := &binanceWs{}
	bnWs.baseURL = "wss://stream.binance.com:9443/ws"
	return bnWs
}

func (this *binanceWs) subscribe(endpoint string, handle func(msg []byte) error) {
	wsBuilder := ws.NewWsBuilder().
		WsUrl(endpoint).
		AutoReconnect().
		MessageHandleFunc(handle)
	this.Conn = wsBuilder.Build()
}

func (this *binanceWs) SubscribeDepth(pair string, callback func (*Depth)) (error, chan<- struct{}) {
	endpoint := fmt.Sprintf("%s/%s@depth@100ms", this.baseURL, pair)
	close := make(chan struct{})

	handle := func(msg []byte) error {
		rawDepth := AcquireDepth()
		if err := json.Unmarshal(msg, rawDepth); err != nil {
			log.Printf("json unmarshal error: %s", string(msg))
			return err
		}

		callback(rawDepth)
		return nil
	}
	this.subscribe(endpoint, handle)

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <- close:
				this.Conn.Close()
				return
			case t := <-ticker.C:
				this.Conn.SendPingMessage([]byte(t.String()))
			}
		}
	}()

	return nil, close
}
