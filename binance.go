package binancewebsocket

import (
	"github.com/json-iterator/go"
	"github.com/alexey-ernest/go-binance-websocket/m/v2/ws"
	"log"
	"fmt"
	"sync"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type binanceWs struct {
	baseURL string
	Conn *ws.WsConn
	depthpool *sync.Pool
}

func NewBinanceWs() *binanceWs {
	bnWs := &binanceWs{}
	bnWs.baseURL = "wss://stream.binance.com:9443/ws"
	bnWs.depthpool = &sync.Pool {
		New: func() interface{} {
			return new(RawDepth)
		},
	}
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
		//this.depthpool.Put(rawDepth)
		return nil
	}
	this.subscribe(endpoint, handle)

	go func() {
		<- close
		this.Conn.Close()
	}()

	return nil, close
}
