package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/resty.v1"
)

/* CONSTRANT */
const (
	SERVER_URL = "https://data-seed-prebsc-1-s1.binance.org:8545/"
)

/* STRUCTURE */
type RPCResp struct {
	Jsonrpc string      `json:"jsonrpc"`
	Id      int         `json:"id"`
	Result  interface{} `json:"result"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

/* DATABASE */
func createDB() {
	log.Println("Creating collector.db...")
	db, _ := sql.Open("sqlite3", "db/collector.sqlite")

	stmt, _ := db.Prepare(`
		CREATE TABLE IF NOT EXISTS "Owner" (
			"new_owner_address"	TEXT,
			"timestamp"	TEXT,
			"transaction_hash"	TEXT,
			"block_number"	TEXT,
			"id"	INTEGER NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`)

	stmt.Exec()
	db.Close()
}

/* PRINTER */
func printOutput(resp *resty.Response, err error) {
	fmt.Println(resp, err)
}

/* CLIENT REST */
/* --POST-- */
func simplePost(requestJsonBody string) {
	// POST JSON string
	// No need to set content type, if you have client level setting

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestJsonBody).
		Post(SERVER_URL)

	body := resp.Body()

	var response RPCResp

	if err := json.Unmarshal(body, &response); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}

	// fmt.Printf("id %s", string(response.result))
	fmt.Printf("%+v\n", response.Error)
	printOutput(resp, err)
}

/* --GET-- */
func simpleGet() {
	resp, err := resty.R().Get("http://httpbin.org/get")

	fmt.Printf("\nError: %v", err)
	fmt.Printf("\nResponse Status Code: %v", resp.StatusCode())
	fmt.Printf("\nResponse Status: %v", resp.Status())
	fmt.Printf("\nResponse Body: %v", resp)
	fmt.Printf("\nResponse Time: %v", resp.Time())
	fmt.Printf("\nResponse Received At: %v", resp.ReceivedAt())
}

/* CONTROLLER */
var test = "<html><header></header><body><button>%s</button></body></html>"

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, test, r.URL.Path[1:])
	log.Println(r.URL.Path[1:])
	switch r.URL.Path[1:] {
	case "/":
		{
			fmt.Fprintf(w, test, r.URL.Path[1:])
			break
		}
	case "/get":
		{
			simpleGet()
			break
		}
	}
}

/* MAIN PROGRAM */
func main() {
	if _, err := os.Stat("collector.db"); errors.Is(err, os.ErrNotExist) {
		createDB()
	}
	var jsonrpc = `{
		"jsonrpc":"2.0",
		"method":"eth_getFilterLogs","params":["0xbd7a033ee864955fef747df2fd8a9c6d"],
		"id":1
	}`

	simplePost(jsonrpc)
	fmt.Println("\nServer Blockchain Collector Owner is running ...")
	http.HandleFunc("/", handler)
	var port string = ":8080"
	fmt.Printf("\nServer Blockchain listening on port %s...", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
