package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/resty.v1"
)

/* CONSTRANT */
const (
	SERVER_URL = "https://data-seed-prebsc-1-s1.binance.org:8545/"
)

const (
	CONTRACT_NUM        = 14599911
	CONTRACT_ADDRESS    = "0x98b3f2219a2b7a047B6234c19926673ad4aac83A"
	TOPIC_CHANGE_OWNER  = "0x342827c97908e5e2f71151c08502a66d44b6f758e3ac2f1de95f02eb95f0a735"
	NEW_FILTER          = "eth_newFilter"
	GET_FILTERLOGS      = "eth_getFilterLogs"
	GET_BLOCK_BY_NUMBER = "eth_getBlockByNumber"
)

/* STRUCTURE */
type ErrorResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RPCResp struct {
	Jsonrpc string         `json:"jsonrpc"`
	Id      int            `json:"id"`
	Result  []FilterParams `json:"result"`
	Error   ErrorResp      `json:"error"`
}

type CustomRPCResp struct {
	Jsonrpc string    `json:"jsonrpc"`
	Id      int       `json:"id"`
	Result  BlockInfo `json:"result"`
	Error   ErrorResp `json:"error"`
}

type iRPCResp struct {
	Jsonrpc string      `json:"jsonrpc"`
	Id      int         `json:"id"`
	Result  interface{} `json:"result"`
	Error   ErrorResp   `json:"error"`
}

type RPCReq struct {
	Jsonrpc string         `json:"jsonrpc"`
	Id      int            `json:"id"`
	Method  string         `json:"method"`
	Params  []FilterParams `json:"params"`
}

type FilterParams struct {
	FromBlock       string   `json:"fromBlock"`
	ToBlock         string   `json:"toBlock"`
	Address         string   `json:"address"`
	Topics          []string `json:"topics"`
	BlockNumber     string   `json:"blockNumber"`
	TransactionHash string   `json:"transactionHash"`
}

type BlockInfo struct {
	Timestamp string `json:"timestamp"`
}

type Owner struct {
	NewOwnerAddress string
	Timestamp       string
	TransactionHash string
	BlockNumber     string
}

/* DATABASE */
func createDB() {
	log.Println("Creating collector.db...")
	db, _ := sql.Open("sqlite3", "db/collector.sqlite")

	statement, _ := db.Prepare(`
		CREATE TABLE IF NOT EXISTS "Owner" (
			"new_owner_address"	TEXT,
			"timestamp"	TEXT,
			"transaction_hash"	TEXT,
			"block_number"	TEXT,
			"id"	INTEGER NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`)

	statement.Exec()
	db.Close()
}

func truncateTable() {
	db, _ := sql.Open("sqlite3", "db/collector.sqlite")
	statement, _ := db.Prepare(`DELETE FROM "Owner"`)
	statement.Exec()
	db.Close()
}

func insertData(newOwner Owner) {
	log.Println("\nInserting OWNER record ...")
	db, _ := sql.Open("sqlite3", "db/collector.sqlite")
	insertOwnerSQL := `INSERT INTO Owner(new_owner_address, timestamp, transaction_hash, block_number) VALUES (?, ?, ?, ?)`
	statement, err := db.Prepare(insertOwnerSQL)
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(newOwner.NewOwnerAddress, newOwner.Timestamp, newOwner.TransactionHash, newOwner.BlockNumber)
	if err != nil {
		log.Fatalln(err.Error())
	}
	db.Close()
}

func displayData() string {
	db, _ := sql.Open("sqlite3", "db/collector.sqlite")
	row, err := db.Query(`SELECT * FROM Owner ORDER BY id DESC`)
	if err != nil {
		log.Fatal(err)
	}
	var table = `<table> <tr><th>ID</th> <th>New Owner</th> <th>Timestamp</th> <th>Transaction Hash</th> <th>Block Number</th></tr>`
	for row.Next() { // Iterate and fetch the records from result cursor
		var id int
		var newOwnerAddress string
		var timestamp string
		var transactionHash string
		var blockNumber string
		row.Scan(&newOwnerAddress, &timestamp, &transactionHash, &blockNumber, &id)
		// printOutStruct(row)
		log.Println("New Owner: ", id, ", ", newOwnerAddress, ", ", timestamp, ", ", transactionHash, ", ", blockNumber)
		var tr = `<tr><td>%d</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`
		i, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			panic(err)
		}
		tm := time.Unix(i, 0)
		rowDate := fmt.Sprintf(tr, id, newOwnerAddress, tm, transactionHash, blockNumber)
		table += rowDate
	}
	table += `</table>`
	db.Close()
	return table
}

/* PRINTER */
// func printOutput(resp *resty.Response, err error) {
// 	fmt.Println(resp, err)
// }
func printErr(err error) {
	fmt.Println("\nERR:%w\n", err)
}
func printOutStruct(s interface{}) {
	fmt.Printf("%+v\n", s)
}

/* CONVERTER */
func hex2Dec(hexaString string) int {
	// replace 0x or 0X with empty String
	numberStr := strings.Replace(hexaString, "0x", "", -1)
	numberStr = strings.Replace(numberStr, "0X", "", -1)
	output, err := strconv.ParseInt(numberStr, 16, 64)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	n, err := strconv.Atoi(strconv.FormatInt(output, 10))
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	return n
}

func dec2Hex(dec int) string {
	// replace 0x or 0X with empty String
	output := fmt.Sprintf("0x%x", dec)
	return output
}

/* BODY REQUEST */
func getBodyJSONRPC() {
	var bodyRPC = ``
	printOutStruct(bodyRPC)
}

func getNewFilterLogs(fromBlock string, toBlock string) {
	//get new filter
	var rawJsonRPC = `{
		"jsonrpc":"2.0",
		"method":"%s",
		"params":[%s],
		"id":1
	}`
	var rawJsonRPCFilter = fmt.Sprintf(rawJsonRPC, NEW_FILTER, `{
		"fromBlock":"%s",
		"toBlock":"%s",
		"address":"%s",
		"topics":["%s"]
	}`)
	jsonFilter := fmt.Sprintf(rawJsonRPCFilter, fromBlock, toBlock, CONTRACT_ADDRESS, TOPIC_CHANGE_OWNER)
	// printOutStruct(respFilter)
	var responseNewFilter = simplePostWithIRPCResp(jsonFilter)
	var filterID = responseNewFilter.Result
	// printOutStruct(filterID)
	var rawJsonRPCFilterLogs = fmt.Sprintf(rawJsonRPC, GET_FILTERLOGS, `"%s"`)
	jsonLogs := fmt.Sprintf(rawJsonRPCFilterLogs, filterID)
	// printOutStruct(respLogs)
	var responseLogs = simplePost(jsonLogs)
	printOutStruct("\n")
	for _, log := range responseLogs.Result {
		// fmt.Printf("\n%d. %+v", i, log)
		var rawNewOwner = log.Topics[2]
		var newOwner = strings.Replace(rawNewOwner, "000000000000000000000000", "", -1)

		var rawJsonRPCBlockInfo = fmt.Sprintf(rawJsonRPC, GET_BLOCK_BY_NUMBER, `"%s", false`)
		jsonBlockInfo := fmt.Sprintf(rawJsonRPCBlockInfo, log.BlockNumber)
		var responseBlockInfo = CustomPost(jsonBlockInfo)
		var timestamp = responseBlockInfo.Result.Timestamp
		var sTime = strconv.Itoa(hex2Dec(timestamp))
		i, err := strconv.ParseInt(sTime, 10, 64)
		if err != nil {
			panic(err)
		}
		tm := time.Unix(i, 0)
		fmt.Printf("\nNew Owner: %s", newOwner)
		fmt.Printf("\nTimestamp: %s", tm)
		fmt.Printf("\nBlock Number: %s", log.BlockNumber)
		fmt.Printf("\nTransaction Hash: %s", log.TransactionHash)
		fmt.Printf("\n___________")
		var owner = Owner{newOwner, sTime, log.BlockNumber, log.TransactionHash}
		insertData(owner)
	}
	// printOutStruct(responseLogs.Result)

}

/* CLIENT REST */
/* --POST-- */
func simplePost(requestJsonBody string) RPCResp {
	// POST JSON string
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestJsonBody).
		Post(SERVER_URL)
	if err != nil {
		printErr(err)
	}
	body := resp.Body()

	var response RPCResp

	if err := json.Unmarshal(body, &response); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	return response
}

/* use to get jsonrpc has simple result */
func simplePostWithIRPCResp(requestJsonBody string) iRPCResp {
	// POST JSON string
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestJsonBody).
		Post(SERVER_URL)
	if err != nil {
		printErr(err)
	}
	body := resp.Body()

	var response iRPCResp

	if err := json.Unmarshal(body, &response); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}

	return response
}

func CustomPost(requestJsonBody string) CustomRPCResp {
	// POST JSON string
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestJsonBody).
		Post(SERVER_URL)
	if err != nil {
		printErr(err)
	}
	body := resp.Body()

	var response CustomRPCResp

	if err := json.Unmarshal(body, &response); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	return response
}

/* --GET-- */
func simpleGet() {
}

/* CONTROLLER */

func handler(w http.ResponseWriter, r *http.Request) {
	var title = "<html><header></header><h2>%s</h2></body></html>"
	var href = "<br><a href='%s'>%s</a>"
	fmt.Fprintf(w, title, r.URL.Path[1:])

	fmt.Fprintf(w, href, "/", "Home Page")
	fmt.Fprintf(w, href, "/refresh", "Refresh")
	fmt.Fprintf(w, href, "/get-owner", "Get Data Table")
	log.Println(r.URL.Path[1:])
	switch r.URL.Path[0:] {
	case "/":
		{
			break
		}
	case "/refresh":
		{
			fmt.Fprintf(w, title, "Refresing data ...")
			progress := updateRecord()
			fmt.Fprint(w, progress)
			break
		}
	case "/get-owner":
		{
			var style = `<style>table, td, th, tr {border: 1px solid black}</style>`
			fmt.Fprint(w, style)
			var table = displayData()
			fmt.Fprint(w, table)
			break
		}
	}
}

/* BUSINESS PROGESS */
func businessProgress() {
	if _, err := os.Stat("collector.db"); errors.Is(err, os.ErrNotExist) {
		createDB()
	}
	updateRecord()
}

func updateRecord() string {
	//Clear exist data
	printOutStruct("\n Clean Exists Data ...")
	truncateTable()
	// get latest block
	var jsonRPC = `{
		"jsonrpc":"2.0",
		"method":"eth_blockNumber",
		"params":[],
		"id":67
	}`
	var latestBlockResp = simplePostWithIRPCResp(jsonRPC)
	var latestBlock = latestBlockResp.Result.(string)

	var latestNum = hex2Dec(latestBlock)

	for i := CONTRACT_NUM; i < latestNum; i += 5000 {
		if latestNum-i < 5000 {
			// fmt.Printf("\n@ %d - %d", CONTRACT_NUM, i)
			var toBlock = dec2Hex(latestNum)
			var fromBlock = dec2Hex(i)
			getNewFilterLogs(fromBlock, toBlock)
		} else {
			var toBlock = dec2Hex(i + 4999)
			var fromBlock = dec2Hex(i)
			// fmt.Printf("\n@ %d - %d", hex2Dec(fromBlock), hex2Dec(toBlock))
			getNewFilterLogs(fromBlock, toBlock)
		}

	}

	log.Println("===================")
	displayData()
	return "Done"
}

/* MAIN PROGRAM */
func main() {
	/* Start Progress */
	businessProgress()

	/* Server Information */
	fmt.Println("\nServer Blockchain Collector Owner is running ...")
	http.HandleFunc("/", handler)
	var port string = ":8080"
	fmt.Printf("\nServer Blockchain listening on port %s...", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
