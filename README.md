[![Build Status](https://travis-ci.com/alexey-ernest/go-binance-websocket.svg?branch=master)](https://travis-ci.com/alexey-ernest/go-binance-websocket)

# go-binance-websocket
Binance websocket client with optimized latency

## Optimized latency
Leveraging fast json deserializer and object pool for good base performance around ~1500 ns/op or ~600K op/s
```
$ go test --bench=. --benchtime 30s --benchmem

BenchmarkBinanceMessageHandling-4 21613016  1575 ns/op  128 B/op  8 allocs/op
```
