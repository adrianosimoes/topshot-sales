package main

import (
	"context"
	"fmt"
	"github.com/rrrkren/topshot-sales/topshot"

	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

func handleErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	// connect to flow
	flowClient, err := client.New("access.mainnet.nodes.onflow.org:9000", grpc.WithInsecure())
	handleErr(err)
	err = flowClient.Ping(context.Background())
	handleErr(err)

	for {
		// fetch latest block
		latestBlock, err := flowClient.GetLatestBlock(context.Background(), false)
		handleErr(err)
		fmt.Println("current height: ", latestBlock.Height)

		blockSize := 20
		for i := 0; i < 20; i+=blockSize {
			//fmt.Println("current block: ", int64(latestBlock.Height) - int64(i))
			fetchBlocks(flowClient, int64(latestBlock.Height) - int64(i) - int64(blockSize), int64(latestBlock.Height) - int64(i), "A.c1e4f4f4c4257510.Market.MomentListed")
			fetchBlocks(flowClient, int64(latestBlock.Height) - int64(i) - int64(blockSize), int64(latestBlock.Height) - int64(i), "A.c1e4f4f4c4257510.Market.MomentPriceChanged")
		}
	}
}

func fetchBlocks(flowClient *client.Client, startBlock int64, endBlock int64, typeStr string) {
	// fetch block events of topshot Market.MomentListed/PriceChanged events for the past 1000 blocks
	blockEvents, err := flowClient.GetEventsForHeightRange(context.Background(), client.EventRangeQuery{
		Type:        typeStr,
		StartHeight: uint64(startBlock),
		EndHeight:   uint64(endBlock),
	})
	handleErr(err)

	for _, blockEvent := range blockEvents {
		for _, sellerEvent := range blockEvent.Events {
			// loop through the Market.MomentListed/PriceChanged events in this blockEvent
			//fmt.Println(sellerEvent.Value)
			e := topshot.MomentListed(sellerEvent.Value)
			if(e.Price() < 35){
				saleMoment, err := topshot.GetSaleMomentFromOwnerAtBlock(flowClient, blockEvent.Height, *e.Seller(), e.Id())
				handleErr(err)
				if(e.Price()<= 9 || (saleMoment != nil && (saleMoment.SerialNumber() <= 500 || (e.Price() < 25 && saleMoment.SerialNumber() <= 1000)))) {
					fmt.Println(saleMoment, "\tPrice: ", e.Price())
				}
			}
		}
	}
}
