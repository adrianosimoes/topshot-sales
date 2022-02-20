package topshot

import (
    "encoding/json"
    "os"
    "fmt"
    "io/ioutil"
)

type Data struct {
    AllPlayers AllPlayers `json:"allPlayers"`
    AllTeams AllTeams `json:"allTeams"`
    AllSets AllSets `json:"allSets"`
}

func (data Data) GetPlayerIDForName(name string) string {
   
   for _, player := range data.AllPlayers.Data {
      if (player.Name == name) {
         return player.ID   
      }
      
   }

   return ""
}

func (data Data) GetSetIdByName(setName string, setSimpleId string) string {

	// Quick hack for the Sets that have the same Name ("Base Set", "Rising Stars") but different ids: 
	if( setSimpleId=="2" || setSimpleId=="26" ||  setSimpleId=="51" || setSimpleId=="35" || setSimpleId=="60"){
		for _, set := range data.AllSets.Data {
		  if (set.SimpleId == setSimpleId) {
			 return set.ID   
		  }
	   }
	}
   
   for _, set := range data.AllSets.Data {
      if (set.Name == setName) {
         return set.ID   
      }
   }

   return ""
}

type AllPlayers struct {
   size   int `json:"size"`
   Data   []Player `json:"data"`
}

type AllTeams struct {
   size   int `json:"size"`
   Data   []Team `json:"data"`
}

type AllSets struct {
   Data   []Set `json:"data"`
}

type Player struct {
   ID   string `json:"id"`
   Name   string `json:"name"`
   Type    string    `json:"__typename"`
}

type Team struct {
   ID   string `json:"id"`
   Name   string `json:"name"`
   Type    string    `json:"__typename"`
}

type Set struct {
   ID   string `json:"id"`
   SimpleId   string `json:"simpleId"`
   Name   string `json:"flowName"`
   VisualID   string `json:"setVisualId"`
   SerialNumber   string `json:"flowSerialNumber"`
   Type    string    `json:"__typename"`
}

func LoadGameData() Data {
   // Open our jsonFile
   jsonFile, err := os.Open("topshot/gameData.json")
   // if we os.Open returns an error then handle it
   if err != nil {
       fmt.Println(err)
   }

   fmt.Println("Successfully Opened gameData.json")
   // defer the closing of our jsonFile so that we can parse it later on
   defer jsonFile.Close()

   // read our opened jsonFile as a byte array.
   byteValue, _ := ioutil.ReadAll(jsonFile)

   // we initialize our Users array
   var gameData Data

   // we unmarshal our byteArray which contains our
   // jsonFile's content into 'users' which we defined above
   json.Unmarshal(byteValue, &gameData)

   // we iterate through every user within our users array and
   // print out the user Type, their name, and their facebook url
   // as just an example
   // for i := 0; i < len(gameData.AllPlayers.Data); i++ {
   //     fmt.Println("Player ID: " + gameData.AllPlayers.Data[i].ID)
   //     fmt.Println("Player Name: " + gameData.AllPlayers.Data[i].Name)
   // }

   return gameData
}

type POSTData struct {
    Data ListingData `json:"data"`
}

func (post POSTData) GetMomentListings() []MomentListing {
   return post.Data.Listings.Data.Summary.Data.Results
}

type ListingData struct {
   Listings SearchMomentListings `json:"searchMomentListings"`
}

type SearchMomentListings struct {
   Data SearchListingContents `json:"data"`  
}

type SearchListingContents struct {
   Summary SearchSummary `json:"searchSummary"`
}

type SearchSummary struct {
   Data SearchSummaryData `json:"data"`
}

type SearchSummaryData struct {
   Results []MomentListing `json:"data"`
   Count int `json:"size"`
}

type MomentListing struct {
   Id string `json:"id"`
   SetData SetData `json:"set"`
   PlayData PlayData `json:"play"`
   Count int `json:"circulationCount"`
}

func (moment MomentListing) GetURLHash() string {
   return moment.SetData.Id+"+"+moment.PlayData.Id
}

type SetData struct {
   Id string `json:"id"`
   Name string `json:"flowName"`
   SeriesNumber int `json:"flowSeriesNumber"`
}

type PlayData struct {
   Id string `json:"id"`
}