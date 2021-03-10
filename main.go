package main

import (
	"context"
	"fmt"
	"github.com/rrrkren/topshot-sales/topshot"
	"github.com/fatih/color"

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
	
	// Run a bigger fetch block the first time, to check more blocks in the past:
	latestBlock, err := flowClient.GetLatestBlock(context.Background(), false)
	handleErr(err)
	fetchBlocks(flowClient, int64(latestBlock.Height - 50), int64(latestBlock.Height), "A.c1e4f4f4c4257510.Market.MomentListed")

	for {
		// fetch latest block
		latestBlock, err := flowClient.GetLatestBlock(context.Background(), false)
		handleErr(err)
		//fmt.Println("current height: ", latestBlock.Height)

		blockSize := 10
		for i := 0; i < blockSize; i+=blockSize {
			//fmt.Println("current block: ", int64(latestBlock.Height) - int64(i))
			fetchBlocks(flowClient, int64(latestBlock.Height) - int64(i) - int64(blockSize), int64(latestBlock.Height) - int64(i), "A.c1e4f4f4c4257510.Market.MomentListed")
			//fetchBlocks(flowClient, int64(latestBlock.Height) - int64(i) - int64(blockSize), int64(latestBlock.Height) - int64(i), "A.c1e4f4f4c4257510.Market.MomentPriceChanged")
		}
		fmt.Print(".")
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
			// fmt.Println(sellerEvent.Value)
			e := topshot.MomentListed(sellerEvent.Value)
			if(e.Price() <= 170){
				saleMoment, err := topshot.GetSaleMomentFromOwnerAtBlock(flowClient, blockEvent.Height, *e.Seller(), e.Id())
				handleErr(err)
				if(shouldPrintPlayer(e, saleMoment)){
					//fmt.Println("be:", sellerEvent)
					fmt.Println("\a")
					//finalString := saleMoment.String()+"\tPrice: "+fmt.Sprintf("%.0f", e.Price())
					c := color.New(color.FgWhite)
					if (isMomentVeryRare(saleMoment)) {
						c = c.Add(color.FgGreen) 
						c = c.Add(color.BgWhite) 	
					}
					if (isMomentRare(saleMoment)) {
						c = c.Add(color.FgGreen) 
					}
					if (isMomentSerialLow(saleMoment)) {
						c = c.Add(color.Bold) 		
					}

					c.Println(saleMoment, "\tPrice: ", e.Price())
						//saleMoment.String()+"\tPrice: "+fmt.Sprintf("%.0f", e.Price()))
					//fmt.Println(saleMoment, "\tPrice: ", e.Price())
				}
			}
		}
	}
}

func shouldPrintPlayer(moment topshot.MomentListed, sale *topshot.SaleMoment) bool {
	if(moment.Price() < 6){
		return true;
	}
	
	if(sale == nil) {
		return false;
	}
	
	if(sale.SetID() != 26 && ( moment.Price() <= 70 || (sale.SetID() == 2 || sale.SetID() == 32 || sale.SetID() == 33 || sale.SetID() == 34) && moment.Price() <= 70)){
		return true;	
	}
	
	if(moment.Price() < 40 && sale.NumMoments() <= 11000){
		return true;	
	}
	
	if(moment.Price() < 10 && sale.NumMoments() <= 15000){
		return true;	
	}
	
	if(moment.Price() < 40 && sale.SerialNumber() <= 200){
		return true;	
	}
	
	if(moment.Price() < 25 && sale.SerialNumber() <= 500){
		return true;	
	}
	
	if(moment.Price() < 20 && sale.SerialNumber() <= 1000){
		return true;
	}
	
	return false;
}

func isMomentVeryRare(sale *topshot.SaleMoment) bool {
	if (sale.NumMoments() <= 3000) {
		return true;
	}
	return false;	
}

func isMomentRare(sale *topshot.SaleMoment) bool {
	if (sale.NumMoments() <= 15000) {
		return true;
	}
	return false;		
}

func isMomentSerialLow(sale *topshot.SaleMoment) bool {
	if (sale.SerialNumber() <= 500) {
		return true;
	}
	return false;
}
