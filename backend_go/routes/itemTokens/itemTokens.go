package itemTokens

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"

	// "fmt"
	"fintrack-go/db"
	"fintrack-go/routes/plaid"
	"fintrack-go/routes/saltedge"
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
		panic(err)
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
//Actually don't, we'll do refresh toks/accs on every trans fetch

func FetchTransactionsFunction() func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {

		baseCurrency := strings.ToUpper(os.Getenv("BASE_CURRENCY"))

		// First get all item tokens
		// itemTokens := []types.ItemToken{}
		// err := db.DBCon.Select(&itemTokens, "SELECT * FROM `item_tokens`")
		// if err != nil {
		// 	log.Fatal(err)
		// }

		itemTokens := SelectAll()

		txn := db.DBCon.MustBegin()

		var wgPre sync.WaitGroup
		wgPre.Add(1)
		go func() {
			defer wgPre.Done()
			saltedge.RefreshConnectionsFunction(txn)
		}()
		wgPre.Add(1)
		go func() {
			defer wgPre.Done()
			db.GetNewXML()
		}()
		wgPre.Wait()

		// Make sure currencies are up to date

		// Refresh Plaid and SaltEdge connections

		// Then we iterate through item tokens and process in either saltedge or plaid

		var wg sync.WaitGroup
		for _, itemTok := range itemTokens {
			wg.Add(1)
			go func(itemToken types.ItemToken) {
				defer wg.Done()
				// Think these can be done in goroutines
				// Using websocket connection to transmit which item is being currently worked on
				// message := []byte(`{ "username": "Booh", }`)
				// socket.ExportHub.Broadcast <- message

				if itemToken.Provider == "SaltEdge" {
					saltedge.FetchTransactionsForItemToken(itemToken, txn, baseCurrency)
					// if itemToken.Interactive {
					// 	// Needs to be direct DB call
					// 	itemToken.LastDownloadedTransactions = itemToken.LastRefresh
					// } else {
					// 	// Needs to be direct DB call
					// 	itemToken.LastDownloadedTransactions = time.Now()
					// }
				} else {
					plaid.FetchTransactionsForItemToken(itemToken, txn, baseCurrency)
					// Needs direct DB call here to set LastDownloadedTransactions
				}
			}(itemTok)

		}
		wg.Wait()

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
