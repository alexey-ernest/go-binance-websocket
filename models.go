package binancewebsocket

import (
	"github.com/alexey-ernest/go-binance-websocket/pool"
	//"errors"
)

type RawDepth struct {
	LastUpdateID int64 `json:"u"`
	Bids [][2]string `json:"b"`
	Asks [][2]string `json:"a"`
}

type Depth struct {
	pool.ReferenceCounter `json:"-"`
	RawDepth
}

func (d *Depth) Reset() {
	d.Bids = nil
	d.Asks = nil
	d.LastUpdateID = 0
}

// Used by reference countable pool
func ResetDepth(i interface{}) error {
	return nil
	// obj, ok := i.(*Depth)
	// if !ok {
	// 	return errors.New("illegal object sent to ResetDepth")
	// }
	// obj.Reset()
	// return nil
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
