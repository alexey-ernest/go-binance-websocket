[![Build Status](https://travis-ci.com/alexey-ernest/go-binance-websocket.svg?branch=master)](https://travis-ci.com/alexey-ernest/go-binance-websocket)

# go-binance-websocket
Binance websocket client with optimized latency

## Optimized latency
Leveraging fast json deserializer and object pool for good base performance (i5 6-core): ~1000 ns/op or ~1M op/s
```
$ go test --bench=. --benchtime 30s --benchmem

BenchmarkBinanceMessageHandling-6  38260412  946 ns/op  128 B/op  8 allocs/op
```

## Example

```
import (
	. "github.com/alexey-ernest/go-binance-websocket"
	"log"
)

func main() {
	ws := NewBinanceWs()
	messages := make(chan *Depth, 10)
	err, _ := ws.SubscribeDepth("btcusdt", func (d *Depth) {
		messages <- d
	})

	if err != nil {
		log.Fatalf("failed to connect to binance @depth websocket")
	}

	for m := range messages {
		log.Printf("%+v\n", m.RawDepth)
		m.DecrementReferenceCount()
	}
}
```