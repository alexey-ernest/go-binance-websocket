package ws

// A websocket connection abstraction

import (
	"github.com/json-iterator/go"
	"errors"
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httputil"
	"time"
	"log"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type WsConfig struct {
	WsUrl                 string
	ReqHeaders            map[string][]string
	MessageHandleFunc     func([]byte) error
	ErrorHandleFunc       func(err error)
	IsAutoReconnect       bool
	IsDump                bool
	readDeadLineTime      time.Duration
}

type WsConn struct {
	c                      *websocket.Conn
	WsConfig
	writeBufferChan        chan []byte
	pingMessageBufferChan  chan []byte
	closeMessageBufferChan chan []byte
	subs                   []interface{}
	close                  chan struct{}
	isClosed               bool
}

type WsBuilder struct {
	wsConfig *WsConfig
}

func NewWsBuilder() *WsBuilder {
	return &WsBuilder {
		&WsConfig {
			ReqHeaders: make(map[string][]string, 1),
		},
	}
}

func (b *WsBuilder) WsUrl(wsUrl string) *WsBuilder {
	b.wsConfig.WsUrl = wsUrl
	return b
}

func (b *WsBuilder) ReqHeader(key, value string) *WsBuilder {
	b.wsConfig.ReqHeaders[key] = append(b.wsConfig.ReqHeaders[key], value)
	return b
}

func (b *WsBuilder) AutoReconnect() *WsBuilder {
	b.wsConfig.IsAutoReconnect = true
	return b
}

func (b *WsBuilder) Dump() *WsBuilder {
	b.wsConfig.IsDump = true
	return b
}

func (b *WsBuilder) MessageHandleFunc(f func([]byte) error) *WsBuilder {
	b.wsConfig.MessageHandleFunc = f
	return b
}

func (b *WsBuilder) ErrorHandleFunc(f func(err error)) *WsBuilder {
	b.wsConfig.ErrorHandleFunc = f
	return b
}

func (b *WsBuilder) Build() *WsConn {
	wsConn := &WsConn{
		WsConfig: *b.wsConfig,
	}
	return wsConn.NewWs()
}

func (ws *WsConn) NewWs() *WsConn {

	ws.readDeadLineTime = time.Minute

	if err := ws.connect(); err != nil {
		log.Panic(err)
	}

	ws.close = make(chan struct{}, 1)
	ws.pingMessageBufferChan = make(chan []byte, 10)
	ws.closeMessageBufferChan = make(chan []byte, 1)
	ws.writeBufferChan = make(chan []byte, 10)

	go ws.writeRequest()
	go ws.receiveMessage()

	return ws
}

func (ws *WsConn) connect() error {
	dialer := websocket.DefaultDialer

	wsConn, resp, err := dialer.Dial(ws.WsUrl, http.Header(ws.ReqHeaders))
	if err != nil {
		log.Printf("[ws][%s] %s", ws.WsUrl, err.Error())
		if ws.IsDump && resp != nil {
			dumpData, _ := httputil.DumpResponse(resp, true)
			log.Printf("[ws][%s] %s", ws.WsUrl, string(dumpData))
		}
		return err
	}

	ws.c = wsConn

	if ws.IsDump {
		dumpData, _ := httputil.DumpResponse(resp, true)
		log.Printf("[ws][%s] %s", ws.WsUrl, string(dumpData))
	}

	return nil
}

func (ws *WsConn) reconnect() {
	ws.c.Close()

	var err error
	
	sleep := 1
	for retry := 1; retry <= 10; retry += 1 {
		time.Sleep(time.Duration(sleep) * time.Second)

		err = ws.connect()
		if err != nil {
			log.Printf("[ws][%s] websocket reconnect fail , %s", ws.WsUrl, err.Error())
		} else {
			break
		}

		sleep <<= 1
	}

	if err != nil {
		log.Printf("[ws][%s] retry reconnect fail , begin exiting. ", ws.WsUrl)
		ws.Close()
		if ws.ErrorHandleFunc != nil {
			ws.ErrorHandleFunc(errors.New("retry reconnect fail"))
		}
	} else {
		// re-subscribe
		var subs []interface{}
		copy(subs, ws.subs)
		ws.subs = ws.subs[:0]
		for _, sub := range subs {
			ws.Subscribe(sub)
		}
	}
}

func (ws *WsConn) writeRequest() {
	var err error

	for {
		select {
		case <-ws.close:
			//log.Printf("[ws][%s] close websocket, exiting write message goroutine.", ws.WsUrl)
			return
		case d := <-ws.writeBufferChan:
			err = ws.c.WriteMessage(websocket.TextMessage, d)
		case d := <-ws.pingMessageBufferChan:
			err = ws.c.WriteMessage(websocket.PingMessage, d)
		case d := <-ws.closeMessageBufferChan:
			err = ws.c.WriteMessage(websocket.CloseMessage, d)
		}

		if err != nil {
			log.Printf("[ws][%s] %s", ws.WsUrl, err.Error())
			time.Sleep(time.Second)
		}
	}
}

func (ws *WsConn) Subscribe(sub interface{}) error {
	data, err := json.Marshal(sub)
	if err != nil {
		log.Printf("[ws][%s] json encode error , %s", ws.WsUrl, err)
		return err
	}
	ws.writeBufferChan <- data
	ws.subs = append(ws.subs, sub)
	return nil
}

func (ws *WsConn) SendMessage(msg []byte) {
	ws.writeBufferChan <- msg
}

func (ws *WsConn) SendPingMessage(msg []byte) {
	ws.pingMessageBufferChan <- msg
}

func (ws *WsConn) SendCloseMessage(msg []byte) {
	ws.closeMessageBufferChan <- msg
}

func (ws *WsConn) SendJsonMessage(m interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		log.Printf("[ws][%s] json encode error , %s", ws.WsUrl, err)
		return err
	}
	ws.writeBufferChan <- data
	return nil
}

func (ws *WsConn) ReceiveMessage(msg []byte) {
	ws.MessageHandleFunc(msg)
}

func (ws *WsConn) receiveMessage() {
	ws.c.SetCloseHandler(func(code int, text string) error {
		log.Printf("[ws][%s] websocket exiting [code=%d , text=%s]", ws.WsUrl, code, text)
		ws.Close()
		return nil
	})

	ws.c.SetPongHandler(func(pong string) error {
		ws.c.SetReadDeadline(time.Now().Add(ws.readDeadLineTime))
		return nil
	})

	ws.c.SetPingHandler(func(ping string) error {
		ws.c.SetReadDeadline(time.Now().Add(ws.readDeadLineTime))
		return nil
	})

	for {

		t, msg, err := ws.c.ReadMessage()
		
		if len(ws.close) > 0 || ws.isClosed {
			//log.Printf("[ws][%s] close websocket, exiting receive message goroutine.", ws.WsUrl)
			return
		}

		if err != nil {
			log.Printf("[ws] error: %s", err)
			if ws.IsAutoReconnect {
				log.Printf("[ws][%s] Unexpected Closed, Begin Retry Connect.", ws.WsUrl)
				ws.reconnect()
				continue
			}

			if ws.ErrorHandleFunc != nil {
				ws.ErrorHandleFunc(err)
			}

			return
		}

		ws.c.SetReadDeadline(time.Now().Add(ws.readDeadLineTime))

		switch t {
		case websocket.TextMessage:
			ws.MessageHandleFunc(msg)
		case websocket.BinaryMessage:
			ws.MessageHandleFunc(msg)
		case websocket.CloseMessage:
			ws.Close()
			return
		default:
			log.Printf("[ws][%s] error websocket message type, content is :\n %s \n", ws.WsUrl, string(msg))
		}
	}
}

func (ws *WsConn) Close() {
	ws.isClosed = true
	ws.close <- struct{}{}
	close(ws.close)

	err := ws.c.Close()
	if err != nil {
		log.Println("[ws]", ws.WsUrl, "close websocket error ,", err)
	}
}
