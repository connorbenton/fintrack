package itemTokens

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	// "fmt"
	"fintrack-go/db"
	"fintrack-go/routes/plaid"
	"fintrack-go/routes/saltedge"
	"fintrack-go/socket"
	"fintrack-go/types"

	_ "github.com/jmoiron/sqlx"
)

type wsMsg struct {
	// type message struct {
	Name string                 `json:"name"`
	Data map[string]interface{} `json:"data"`
}

// type CurrencyRate struct {
// 	Id    int            `json:"id"`
// 	Date  time.Time      `json:"date" db:"date"`
// 	Rates types.JSONText `json:"rates" db:"rates"`

// 	CreatedAt time.Time `json:"created_at" db:"created_at"`
// 	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
// }

func SelectAll() []types.ItemToken {
	dbdata := []types.ItemToken{}
	err := db.DBCon.Select(&dbdata, "SELECT * FROM `item_tokens`")
	if err != nil {
		log.Fatal(err)
	}
	return dbdata
}

func GetFunction() func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {

		dbdata := SelectAll()
		// err := db.DBCon.Select(&dbdata, "SELECT * FROM `item_tokens`")
		// if err != nil {
		// log.Fatal(err)
		// }

		res.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(res).Encode(dbdata); err != nil {
			panic(err)
		}
	}
}

//Need a 'refresh item tokens for new accounts' method to go along with a new button on Accounts tab, instead of refreshing SaltEdge and Plaid connections on each try

func FetchTransactionsFunction() func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {

		// First get all item tokens
		itemTokens := []types.ItemToken{}
		err := db.DBCon.Select(&itemTokens, "SELECT * FROM `item_tokens`")
		if err != nil {
			log.Fatal(err)
		}

		// Make sure currencies are up to date

		// Refresh Plaid and SaltEdge connections

		// Then we iterate through item tokens and process in either saltedge or plaid

		for _, itemToken := range itemTokens {
			// Think these can be done in goroutines
			// Using websocket connection to transmit which item is being currently worked on
			message := []byte(`{ "username": "Booh", }`)
			socket.ExportHub.Broadcast <- message
			if itemToken.Provider == "SaltEdge" {
				saltedge.FetchTransactionsForItemToken(itemToken.ItemID)
				if itemToken.Interactive {
					// Needs to be direct DB call
					itemToken.LastDownloadedTransactions = itemToken.LastRefresh
				} else {
					// Needs to be direct DB call
					itemToken.LastDownloadedTransactions = time.Now()
				}
			} else {
				plaid.FetchTransactionsForItemToken(itemToken.ItemID)
				// Needs direct DB call here to set LastDownloadedTransactions
			}

		}

		// currencyRates := []CurrencyRate{}
		// err2 := db.DBCon.Select(&currencyRates, "SELECT * FROM `currency_rates`")
		// if err2 != nil {
		// 	log.Fatal(err)
		// }
		// rates, _ := string([]byte{json.Marshal(currencyRates[0].DataJSON)})
		// log.Println(rates)

		// res.WriteHeader(http.StatusOK)
		// if err := json.NewEncoder(res).Encode(itemTokens); err != nil {
		// 	panic(err)
		// }
	}
}