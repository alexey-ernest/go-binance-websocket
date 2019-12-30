package binancewebsocket

import (
	"github.com/alexey-ernest/go-binance-websocket/m/v2/pool"
	"errors"
)

type Depth struct {
	pool.ReferenceCounter `json:"-"`

	LastUpdateID int64 `json:"u"`
	Bids [][]interface{} `json:"b"`
	Asks [][]interface{} `json:"a"`
}

func (e *Depth) Reset() {
	e.Bids = nil
	e.Asks = nil
	e.LastUpdateID = 0
}

// Used by reference countable pool
func ResetDepth(i interface{}) error {
	obj, ok := i.(*Depth)
	if !ok {
		return errors.New("illegal object sent to ResetDepth")
	}
	obj.Reset()
	return nil
}

// depth pool
var depthPool = pool.NewReferenceCountedPool(
	func(counter pool.ReferenceCounter) pool.ReferenceCountable {
		d := new(Depth)
		d.ReferenceCounter = counter
		return d
	}, ResetDepth)

// Method to get new Depth
func AcquireDepth() *Depth {
	return depthPool.Get().(*Depth)
}
