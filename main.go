package main

import (
	//"time"
	"encoding/json"
	"runtime"
	"strings"
	"strconv"
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

var gameData topshot.Data

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

	gameData = topshot.LoadGameData()	

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
	// latestBlock, err := flowClient.GetLatestBlock(context.Background(), false)
	// handleErr(err)

	//fetchBlocks(flowClient, int64(latestBlock.Height - 50), int64(latestBlock.Height), "A.c1e4f4f4c4257510.Market.MomentListed")

	for {
		// fetch latest block
		latestBlock, err := flowClient.GetLatestBlock(context.Background(), false)
		handleErr(err)
		//fmt.Println("current height: ", latestBlock.Height)

		blockSize := 5
		
		//start := time.Now()
		for i := 0; i < blockSize; i+=blockSize {
			//fmt.Println("current block: ", int64(latestBlock.Height) - int64(i))
			
			fetchBlocks(flowClient, int64(latestBlock.Height) - int64(i) - int64(blockSize), int64(latestBlock.Height) - int64(i), "A.c1e4f4f4c4257510.Market.MomentListed")

			//fetchBlocks(flowClient, int64(latestBlock.Height) - int64(i) - int64(blockSize), int64(latestBlock.Height) - int64(i), "A.c1e4f4f4c4257510.Market.MomentPriceChanged")
		}
		// elapsed := time.Since(start)
		// fmt.Println("Fetch block took %s", elapsed)
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
			
			if(e.Price() <= 60){
				// start := time.Now()

				saleMoment, err := topshot.GetSaleMomentFromOwnerAtBlock(flowClient, blockEvent.Height, *e.Seller(), e.Id())
				handleErr(err)
				// elapsed := time.Since(start)
    // 			fmt.Println("Getting sale moment took %s", elapsed)
    			
				if(shouldPrintPlayer(e, saleMoment)){
					//start := time.Now()

					printPlayer(saleMoment, true)

					// elapsed := time.Since(start)
    	// 			fmt.Println("Print player took %s", elapsed)
				}
			}
		}
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

func shouldPrintPlayer(moment topshot.MomentListed, sale *topshot.SaleMoment) bool {
	// if(moment.Price() < 11){
	// 	return true;
	// }

	if(sale == nil) {
		return false;
	}
	
	if(sale.SerialNumber() == sale.JerseyNumber()){
		return true;	
	}
	
	if(sale.SetID() != 26 && ( moment.Price() <= 70 || (sale.SetID() == 2 || sale.SetID() == 32 || sale.SetID() == 33 || sale.SetID() == 34) && moment.Price() <= 70)){
		return true;	
	}
	
	if(moment.Price() <= 25 && sale.SerialNumber() < 300 && sale.NumMoments() <= 12000) {
		return true;	
	}

	if(moment.Price() <= 12 && sale.SerialNumber() < 1000 && sale.NumMoments() <= 12000) {
		return true;	
	}
	
	if(moment.Price() <= 10 && sale.SerialNumber() <= 2500 && sale.NumMoments() <= 15000){
		return true;	
	}
	
	if(moment.Price() <= 15 && sale.SerialNumber() <= 250 && sale.NumMoments() < 35000){
		return true;	
	}
	
	if(moment.Price() <= 10 && sale.SerialNumber() <= 500 && sale.NumMoments() < 35000){
		return true;	
	}
	
	if(moment.Price() <= 10 && sale.SerialNumber() <= 1000){
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
	if (sale.SerialNumber() <= 200) {
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
		printPlayer(saleMoment, true)
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

func printPlayer(saleMoment *topshot.SaleMoment, printURL bool) {
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
	
	// start := time.Now()
	if (printURL) {
		url := getPlayerURL(saleMoment)

		// elapsed := time.Since(start)
		// fmt.Println("Getting url took %s", elapsed)

		c.Println(url)
		

		// start = time.Now()
		
		shoutSale(saleMoment)
		
		// elapsed = time.Since(start)
		// fmt.Println("Shouting took %s", elapsed)


		// start = time.Now()

		openURLOnChrome(url)

		// elapsed = time.Since(start)
		// fmt.Println("Opening on chrome took %s", elapsed)		
	}

	fmt.Println("")
}

func getMomentInfoFromPlayerID(playerId int, momentsCount uint32, price float64) []byte {
	playerIdStr := strconv.Itoa(playerId)
	momentsCountStr := strconv.Itoa(int(momentsCount))
	priceStr := fmt.Sprintf("%.0f", price)

	queryData := "{\"operationName\":\"SearchMomentListingsDefault\",\"variables\":{\"byPrice\":{\"min\":null,\"max\":\"" + priceStr + "\"},\"byPower\":{\"min\":null,\"max\":null},\"bySerialNumber\":{\"min\":null,\"max\":\"" + momentsCountStr + "\"},\"byGameDate\":{\"start\":null,\"end\":null},\"byCreatedAt\":{\"start\":null,\"end\":null},\"byPrimaryPlayerPosition\":[],\"bySets\":[],\"bySeries\":[],\"bySetVisuals\":[],\"byPlayStyle\":[],\"bySkill\":[],\"byPlayers\":[\"" + playerIdStr + "\"],\"byTagNames\":[],\"byTeams\":[],\"byListingType\":[\"BY_USERS\"],\"searchInput\":{\"pagination\":{\"cursor\":\"\",\"direction\":\"RIGHT\",\"limit\":12}},\"orderBy\":\"UPDATED_AT_DESC\"},\"query\":\"query SearchMomentListingsDefault($byPlayers: [ID], $byTagNames: [String!], $byTeams: [ID], $byPrice: PriceRangeFilterInput, $orderBy: MomentListingSortType, $byGameDate: DateRangeFilterInput, $byCreatedAt: DateRangeFilterInput, $byListingType: [MomentListingType], $bySets: [ID], $bySeries: [ID], $bySetVisuals: [VisualIdType], $byPrimaryPlayerPosition: [PlayerPosition], $bySerialNumber: IntegerRangeFilterInput, $searchInput: BaseSearchInput!, $userDapperID: ID) {\n  searchMomentListings(input: {filters: {byPlayers: $byPlayers, byTagNames: $byTagNames, byGameDate: $byGameDate, byCreatedAt: $byCreatedAt, byTeams: $byTeams, byPrice: $byPrice, byListingType: $byListingType, byPrimaryPlayerPosition: $byPrimaryPlayerPosition, bySets: $bySets, bySeries: $bySeries, bySetVisuals: $bySetVisuals, bySerialNumber: $bySerialNumber}, sortBy: $orderBy, searchInput: $searchInput, userDapperID: $userDapperID}) {\n    data {\n      filters {\n        byPlayers\n        byTagNames\n        byTeams\n        byPrimaryPlayerPosition\n        byGameDate {\n          start\n          end\n          __typename\n        }\n        byCreatedAt {\n          start\n          end\n          __typename\n        }\n        byPrice {\n          min\n          max\n          __typename\n        }\n        bySerialNumber {\n          min\n          max\n          __typename\n        }\n        bySets\n        bySeries\n        bySetVisuals\n        __typename\n      }\n      searchSummary {\n        count {\n          count\n          __typename\n        }\n        pagination {\n          leftCursor\n          rightCursor\n          __typename\n        }\n        data {\n          ... on MomentListings {\n            size\n            data {\n              ... on MomentListing {\n                id\n                version\n                circulationCount\n                flowRetired\n                set {\n                  id\n                  flowName\n                  setVisualId\n                  flowSeriesNumber\n                  __typename\n                }\n                play {\n                  description\n                  id\n                  stats {\n                    playerName\n                    dateOfMoment\n                    playCategory\n                    teamAtMomentNbaId\n                    teamAtMoment\n                    __typename\n                  }\n                  __typename\n                }\n                assetPathPrefix\n                priceRange {\n                  min\n                  max\n                  __typename\n                }\n                momentListingCount\n                listingType\n                userOwnedSetPlayCount\n                __typename\n              }\n              __typename\n            }\n            __typename\n          }\n          __typename\n        }\n        __typename\n      }\n      __typename\n    }\n    __typename\n  }\n}\n\"}"
	//queryData = strings.ReplaceAll(queryData, "\\n", "\n")
	queryData = strings.Replace(queryData, "\n",`\n`, -1)

	args := []string{"--header", "Content-Type: application/json", "--data", queryData, "https://api.nbatopshot.com/marketplace/graphql?SearchMomentListingsDefault"}
	c := exec.Command("curl",args...)

	output,_ := c.Output()

	return output
}

func getPlayerURL(saleMoment *topshot.SaleMoment) string {
		playData := saleMoment.Play()
		playerIdStr := gameData.GetPlayerIDForName(playData["FullName"])
		
		playerId, _ := strconv.Atoi(playerIdStr)

		//fmt.Println("https://www.nbatopshot.com/search?byPlayers="+gameData.GetPlayerIDForName(playData["FullName"]))
		jsonBytes := getMomentInfoFromPlayerID(playerId, saleMoment.NumMoments(), saleMoment.Price())

		var postData topshot.POSTData	
		err :=json.Unmarshal(jsonBytes, &postData)
		if err != nil {
			fmt.Println("error:", err)
		}	

		momentListings := postData.GetMomentListings()
		momentCount := len(momentListings)

		if(momentCount > 1) {
			args := []string{"Warning!","Found", "moment", "with", "more", "than", "1", "listing"}
			shoutStrings(args)

			for i := 0; i < momentCount; i+=1 {
			  // Each value is an interface{} type, that is type asserted as a string
			  moment := momentListings[i]

			  if ((int(moment.Count) == int(saleMoment.NumMoments())) && (moment.SetData.Name == saleMoment.SetName())) {
				return "https://www.nbatopshot.com/listings/p2p/"+moment.GetURLHash()+"?serialNumber="+strconv.FormatUint(uint64(saleMoment.SerialNumber()), 10)
			  }
			}

			return "no url found :("

		} else if (momentCount == 1) {
			moment := momentListings[0]
			return "https://www.nbatopshot.com/listings/p2p/"+moment.GetURLHash()+"?serialNumber="+strconv.FormatUint(uint64(saleMoment.SerialNumber()), 10)
		}

		return "no url found :("
}

func openURLOnChrome(url string) {
	if runtime.GOOS == "windows" {
		args := []string{url}
		c := exec.Command("C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe",args...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stdout
		c.Run()	
		
	} else {
		args := []string{"--new", "-a", "Google Chrome", "--args",url}
		c := exec.Command("open",args...)
		c.Stdout = os.Stdout
		c.Run()	
	}
}

func shoutSale(saleMoment *topshot.SaleMoment) {
	if runtime.GOOS == "windows" {
		return;
	}

	serialStr := strconv.FormatUint(uint64(saleMoment.SerialNumber()), 10)
	totalStr := strconv.FormatUint(uint64(saleMoment.NumMoments()), 10)
	priceStr := fmt.Sprintf("%.0f", saleMoment.Price()) + "$"

	args := []string{"serial", serialStr, "of", totalStr, "price", priceStr}
	shoutStrings(args)
}

func shoutStrings(args []string) {
	c := exec.Command("say",args...)
	c.Stdout = os.Stdout
	c.Run()		
}