package topshot

import (
	"fmt"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)


// pub event MomentListed(id: UInt64, price: UFix64, seller: Address?)
type MomentListed cadence.Event

// pub event MomentPriceChanged(id: UInt64, newPrice: UFix64, seller: Address?)
type MomentPriceChanged cadence.Event


func (evt MomentListed) Id() uint64 {
	return uint64(evt.Fields[0].(cadence.UInt64))
}

func (evt MomentListed) Price() float64 {
	return float64(evt.Fields[1].(cadence.UFix64).ToGoValue().(uint64))/1e8 // ufixed 64 have 8 digits of precision
}

func (evt MomentListed) Seller() *flow.Address {
	optionalAddress := (evt.Fields[2]).(cadence.Optional)
	if cadenceAddress, ok := optionalAddress.Value.(cadence.Address); ok {
		sellerAddress := flow.BytesToAddress(cadenceAddress.Bytes())
		return &sellerAddress
	}
	return nil
}

func (evt MomentListed) String() string {
	return fmt.Sprintf("moment: momentid: %d, price: %f, seller: %s",
		evt.Id(), evt.Price(), evt.Seller())
}
