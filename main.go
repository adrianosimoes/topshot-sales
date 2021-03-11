package main

import (
	"flag"
	"os"
	"os/exec"
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

	boolPtr := flag.Bool("unique",false,"whether entries are merged till they're unique or not")
	flag.Parse()

	if (*boolPtr) {
		fmt.Println("using unique mode!")
	} else {
		fmt.Println("NOT using unique mode!")
	}

	if (*boolPtr) {
		uniqueModeLogic()
	} else {
		nonUniqueModeLogic()
	}
}

func nonUniqueModeLogic() {
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

func uniqueModeLogic() {
	// connect to flow
	flowClient, err := client.New("access.mainnet.nodes.onflow.org:9000", grpc.WithInsecure())
	handleErr(err)
	err = flowClient.Ping(context.Background())
	handleErr(err)
	
	// Run a bigger fetch block the first time, to check more blocks in the past:
	latestBlock, err := flowClient.GetLatestBlock(context.Background(), false)
	handleErr(err)
	
	filteredMoments := make([]*topshot.SaleMoment,0)
	momentCount := 0

	blocks :=fetchFilteredBlocks(flowClient, int64(latestBlock.Height - 50), int64(latestBlock.Height), "A.c1e4f4f4c4257510.Market.MomentListed")
	filteredMoments = mergeUniqueBlocks(blocks, filteredMoments)

	if (len(filteredMoments) > momentCount) {
		dumpMoments(filteredMoments)	
		momentCount = len(filteredMoments)
	}

	for {
		// fetch latest block
		latestBlock, err := flowClient.GetLatestBlock(context.Background(), false)
		handleErr(err)
		//fmt.Println("current height: ", latestBlock.Height)

		blockSize := 10
		
		for i := 0; i < blockSize; i+=blockSize {
			//fmt.Println("current block: ", int64(latestBlock.Height) - int64(i))
			
			//fetchBlocks(flowClient, int64(latestBlock.Height) - int64(i) - int64(blockSize), int64(latestBlock.Height) - int64(i), "A.c1e4f4f4c4257510.Market.MomentListed")
			//fetchBlocks(flowClient, int64(latestBlock.Height) - int64(i) - int64(blockSize), int64(latestBlock.Height) - int64(i), "A.c1e4f4f4c4257510.Market.MomentPriceChanged")

			blocks = fetchFilteredBlocks(flowClient, int64(latestBlock.Height) - int64(i) - int64(blockSize), int64(latestBlock.Height) - int64(i), "A.c1e4f4f4c4257510.Market.MomentListed")

			filteredMoments = mergeUniqueBlocks(blocks, filteredMoments)

			// blocks = fetchFilteredBlocks(flowClient, int64(latestBlock.Height) - int64(i) - int64(blockSize), int64(latestBlock.Height) - int64(i), "A.c1e4f4f4c4257510.Market.MomentPurchased")
			// filteredMoments = removePurchasedBlocks(blocks, filteredMoments)	

			if (len(filteredMoments) > momentCount) {
				dumpMoments(filteredMoments)	
				momentCount = len(filteredMoments)
			}
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
			if(e.Price() <= 300){
				saleMoment, err := topshot.GetSaleMomentFromOwnerAtBlock(flowClient, blockEvent.Height, *e.Seller(), e.Id())
				handleErr(err)
				if(shouldPrintPlayer(e, saleMoment)){
					//fmt.Println("be:", sellerEvent)
					fmt.Println("\a")

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
	
	if(sale.SerialNumber() == sale.JerseyNumber()){
		return true;	
	}
	
	if(sale.SetID() != 26 && ( moment.Price() <= 70 || (sale.SetID() == 2 || sale.SetID() == 32 || sale.SetID() == 33 || sale.SetID() == 34) && moment.Price() <= 70)){
		return true;	
	}
	
	if(moment.Price() < 40 && sale.NumMoments() <= 11000){
		return true;	
	}
	
	if(moment.Price() < 15 && sale.NumMoments() <= 15000 && sale.SerialNumber() <= 2500){
		return true;	
	}
	
	if(moment.Price() < 40 && sale.SerialNumber() <= 200){
		return true;	
	}
	
	if(moment.Price() < 25 && sale.SerialNumber() <= 500){
		return true;	
	}
	
	if(moment.Price() <= 15 && sale.SerialNumber() <= 1000){
		return true;
	}
	
	return false;
}

func isMomentVeryRare(sale *topshot.SaleMoment) bool {
	if (sale.NumMoments() <= 3000 || sale.SerialNumber() == sale.JerseyNumber() ) {
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

func fetchFilteredBlocks(flowClient *client.Client, startBlock int64, endBlock int64, typeStr string) []*topshot.SaleMoment {
	// fetch block events of topshot Market.MomentListed/PriceChanged events for the past 1000 blocks
	blockEvents, err := flowClient.GetEventsForHeightRange(context.Background(), client.EventRangeQuery{
		Type:        typeStr,
		StartHeight: uint64(startBlock),
		EndHeight:   uint64(endBlock),
	})
	handleErr(err)

	filteredMoments := make([]*topshot.SaleMoment,0)

	for _, blockEvent := range blockEvents {
		for _, sellerEvent := range blockEvent.Events {
			// loop through the Market.MomentListed/PriceChanged events in this blockEvent
			// fmt.Println(sellerEvent.Value)
			e := topshot.MomentListed(sellerEvent.Value)
			if(e.Price() <= 300){
				saleMoment, err := topshot.GetSaleMomentFromOwnerAtBlock(flowClient, blockEvent.Height, *e.Seller(), e.Id())
				handleErr(err)
				if(shouldPrintPlayer(e, saleMoment)){
					filteredMoments = append(filteredMoments, saleMoment)
				}
			}
		}
	}

	return filteredMoments
}

func clearScreen() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	c.Run()
}

func dumpMoments(moments []*topshot.SaleMoment) {
	//first clear
	clearScreen()

	//new list needs to warn
	fmt.Println("\a")
	fmt.Println("---------------------------------------------")
	for _, saleMoment := range moments {
		
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

		c.Println(saleMoment, "\tPrice: ", fmt.Sprintf("%.0f", saleMoment.Price()))	
		fmt.Println("")
	}
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++")
}

func mergeUniqueBlocks(blocks []*topshot.SaleMoment, filteredMoments []*topshot.SaleMoment) []*topshot.SaleMoment {
	mergedMoments := make([]*topshot.SaleMoment, 0)
	mergedMoments = append(mergedMoments,filteredMoments...)

	for _, newMoment := range blocks {
		found := false
		for _, existingMoment := range filteredMoments {
			if (existingMoment.SetID() == newMoment.SetID() &&
				existingMoment.PlayID() == newMoment.PlayID() &&
				existingMoment.Price() == newMoment.Price() &&
				existingMoment.SerialNumber() == newMoment.SerialNumber()) {
					found = true
					break
				} 
		}

		if (found == false) {
			mergedMoments = append(mergedMoments, newMoment)
		}
	}

	return mergedMoments
}

func removePurchasedBlocks(blocks []*topshot.SaleMoment, filteredMoments []*topshot.SaleMoment) []*topshot.SaleMoment {
	mergedMoments := make([]*topshot.SaleMoment, 0)

	for _, existingMoment := range filteredMoments {
		found := false

		for _, goneMoment := range blocks {
			
		
			if (existingMoment.SetID() == goneMoment.SetID() &&
				existingMoment.PlayID() == goneMoment.PlayID() &&
				existingMoment.Price() == goneMoment.Price()) {
					found = true
					break
			} 
		}

		if (found == false) {
			mergedMoments = append(mergedMoments, existingMoment)
		}
	}

	return mergedMoments
}