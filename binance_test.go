package binancewebsocket

import (
	"testing"
)

func TestBinanceDepthConnection(t *testing.T) {
	ws := NewBinanceWs()
	err, close := ws.SubscribeDepth("btcusdt", func (d *Depth) {})
	if err != nil {
		t.Fatalf("failed to connect to binance @depth websocket")
	}
	close <- struct{}{}
}

func TestBinanceDepthMessage(t *testing.T) {
	ws := NewBinanceWs()
	messages := make(chan *Depth, 10)
	err, close := ws.SubscribeDepth("btcusdt", func (d *Depth) {
		messages <- d
	})

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
}

func BenchmarkBinanceMessageHandling(b *testing.B) {
	ws := NewBinanceWs()
	//messages := make(chan *Depth, 10)
	err, close := ws.SubscribeDepth("btcusdt2", func (d *Depth) {
		d.DecrementReferenceCount()
	})
	if err != nil {
		b.Fatalf("failed to connect to binance @depth websocket")
	}

	// go func() {
	// 	for {
	// 		d := <- messages
	// 		d.DecrementReferenceCount()
	// 	}
	// }()

	b.ResetTimer()
	msg := []byte("{\"e\":\"depthUpdate\",\"E\":1577485630559,\"s\":\"BTCUSDT\",\"U\":1627259958,\"u\":1627259960,\"b\":[[\"7246.02000000\",\"0.00000000\"],[\"7246.00000000\",\"0.02930400\"],[\"7245.75000000\",\"0.00000000\"],[\"7239.18000000\",\"0.00000000\"]],\"a\":[]}")
	for i := 0; i < b.N; i += 1 {
		ws.Conn.ReceiveMessage(msg)
	}

	close <- struct{}{}
}