package binancewebsocket

import (
	"github.com/json-iterator/go"
	"github.com/alexey-ernest/go-binance-websocket/m/v2/ws"
	"log"
	"fmt"
	//"sync"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type binanceWs struct {
	baseURL string
	//depthPool *sync.Pool
	Conn *ws.WsConn
}

func NewBinanceWs() *binanceWs {
	bnWs := &binanceWs{}
	bnWs.baseURL = "wss://stream.binance.com:9443/ws"
	// bnWs.depthPool = &sync.Pool {
	// 	New: func() interface{} {
	// 		return new(Depth)
	// 	},
	// }
	return bnWs
}

type Depth struct {
	//pool *sync.Pool

	LastUpdateID int64 `json:"u"`
	Bids [][]interface{} `json:"b"`
	Asks [][]interface{} `json:"a"`
}

// return instance to the pool
// func (d *Depth) Done() {
// 	d.pool.Put(d)
// }

func (this *binanceWs) subscribe(endpoint string, handle func(msg []byte) error) {
	wsBuilder := ws.NewWsBuilder().
		WsUrl(endpoint).
		AutoReconnect().
		MessageHandleFunc(handle)
	this.Conn = wsBuilder.Build()
}

func (this *binanceWs) SubscribeDepth(pair string) (error, <-chan *Depth, chan<- struct{}) {
	endpoint := fmt.Sprintf("%s/%s@depth@100ms", this.baseURL, pair)
	messages := make(chan *Depth, 10)
	close := make(chan struct{})

	handle := func(msg []byte) error {
		rawDepth := Depth{}
		if err := json.Unmarshal(msg, &rawDepth); err != nil {
			log.Printf("json unmarshal error: %s", string(msg))
			return err
		}

		// send message down to the channel
		messages <- &rawDepth

		return nil
	}
	this.subscribe(endpoint, handle)

	go func() {
		<- close
		this.Conn.Close()
	}()

	return nil, messages, close
}
