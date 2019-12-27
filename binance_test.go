package binancewebsocket

import (
	"testing"
	"time"
)

func TestBinanceDepthConnection(t *testing.T) {
	ws := NewBinanceWs()
	err, _, close := ws.SubscribeDepth("BTCUSDT")
	if err != nil {
		t.Fatalf("failed to connect to binance @depth websocket")
	}
	close <- struct{}{}
	time.Sleep(time.Duration(1) * time.Second)
}

func TestBinanceDepthMessage(t *testing.T) {
	ws := NewBinanceWs()
	err, messages, close := ws.SubscribeDepth("btcusdt")
	if err != nil {
		t.Fatalf("failed to connect to binance @depth websocket")
	}

	msg := <- messages
	if msg.LastUpdateID == 0 {
		t.Errorf("LastUpdateID should not be 0")
	}
	if len(msg.Bids) == 0 && len(msg.Asks) == 0 {
		t.Errorf("Depth update should contain asks or bids")
	}

	close <- struct{}{}
	time.Sleep(time.Duration(1) * time.Second)
}