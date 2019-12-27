package binancewebsocket

import (
	"testing"
)

func TestBinanceDepthConnection(t *testing.T) {
	ws := NewBinanceWs()
	err, messages, close := ws.SubscribeDepth("btc-usdt")
	if err != nil {
		t.Fatalf("failed to connect to binance @depth websocket")
	}
	
}