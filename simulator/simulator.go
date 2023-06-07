package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"google.golang.org/grpc"
	"log"
	"simulator/utils"
	"time"
)

type Good struct {
	GoodId   string `json:"goodId"`
	Quantity string `json:"quantity"`
	Unit     string `json:"unit"`
}

const (
	mspID        = "SingleWindowMSP"
	cryptoPath   = "../test-network/organizations/peerOrganizations/singlewindow.example.com"
	certPath     = cryptoPath + "/users/User1@singlewindow.example.com/msp/signcerts/User1@singlewindow.example.com-cert.pem"
	keyPath      = cryptoPath + "/users/User1@singlewindow.example.com/msp/keystore/"
	tlsCertPath  = cryptoPath + "/peers/peer0.singlewindow.example.com/tls/ca.crt"
	peerEndpoint = "localhost:7051"
	gatewayPeer  = "peer0.singlewindow.example.com"

	mspIDTc        = "TradingCompany1MSP"
	cryptoPathTc   = "../test-network/organizations/peerOrganizations/tradingcompany1.example.com"
	certPathTc     = cryptoPathTc + "/users/User1@tradingcompany1.example.com/msp/signcerts/User1@tradingcompany1.example.com-cert.pem"
	keyPathTc      = cryptoPathTc + "/users/User1@tradingcompany1.example.com/msp/keystore/"
	tlsCertPathTc  = cryptoPathTc + "/peers/peer0.tradingcompany1.example.com/tls/ca.crt"
	peerEndpointTc = "localhost:9051"
	gatewayPeerTc  = "peer0.tradingcompany1.example.com"
)

func main() {
	swGrpcConnection := utils.NewGrpcConnection(mspID, cryptoPath, certPath, keyPath, tlsCertPath, peerEndpoint, gatewayPeer)
	defer func(swGrpcConnection *grpc.ClientConn) {
		err := swGrpcConnection.Close()
		if err != nil {
			panic(err)
		}
	}(swGrpcConnection)

	swId := utils.NewIdentity(certPath, mspID)
	swSign := utils.NewSign(keyPath)

	swGw, err := client.Connect(
		swId,
		client.WithSign(swSign),
		client.WithClientConnection(swGrpcConnection),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer func(swGw *client.Gateway) {
		err := swGw.Close()
		if err != nil {
			panic(err)
		}
	}(swGw)

	chaincodeName := "do_chaincode"
	channelName := "singlewindow"

	network := swGw.GetNetwork(channelName)
	contract := network.GetContract(chaincodeName)

	tcGrpcConnection := utils.NewGrpcConnection(mspIDTc, cryptoPathTc, certPathTc, keyPathTc, tlsCertPathTc, peerEndpointTc, gatewayPeerTc)
	defer func(tcGrpcConnection *grpc.ClientConn) {
		err := tcGrpcConnection.Close()
		if err != nil {
			panic(err)
		}
	}(tcGrpcConnection)

	tcId := utils.NewIdentity(certPathTc, mspIDTc)
	tcSign := utils.NewSign(keyPathTc)

	tcGw, err := client.Connect(
		tcId,
		client.WithSign(tcSign),
		client.WithClientConnection(tcGrpcConnection),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer func(tcGw *client.Gateway) {
		err := tcGw.Close()
		if err != nil {
			panic(err)
		}
	}(tcGw)

	tcNetwork := tcGw.GetNetwork(channelName)
	tcOrderContract := tcNetwork.GetContractWithName(chaincodeName, "OrderContract")
	tcGoodContract := tcNetwork.GetContractWithName(chaincodeName, "GoodContract")

	initGoods(contract)
	time.Sleep(5 * time.Second)
	readGoods(contract)
	time.Sleep(5 * time.Second)
	createOrderSuccess(tcOrderContract)
	time.Sleep(5 * time.Second)
	readOrder(tcOrderContract)
	time.Sleep(5 * time.Second)
	createOrderFailed(tcOrderContract)
	time.Sleep(5 * time.Second)
	createGoodUnauthorized(tcGoodContract)
	time.Sleep(5 * time.Second)

	// congruence test
	createOrderConcurrently(tcOrderContract, 9)
	cleanLedger(tcOrderContract, "order-")
	createOrderConcurrently(tcOrderContract, 10)
	cleanLedger(tcOrderContract, "order-")
	createOrderConcurrently(tcOrderContract, 100)
	cleanLedger(tcOrderContract, "order-")
	createOrderConcurrently(tcOrderContract, 1000)
	cleanLedger(tcOrderContract, "order-")
	createOrderConcurrently(tcOrderContract, 2000)
	cleanLedger(tcOrderContract, "order-")
	createOrderConcurrently(tcOrderContract, 3000)
	cleanLedger(tcOrderContract, "order-")
	createOrderConcurrently(tcOrderContract, 4000)
	cleanLedger(tcOrderContract, "order-")
	createOrderConcurrently(tcOrderContract, 5000)
	cleanLedger(tcOrderContract, "order-")
	createOrderConcurrently(tcOrderContract, 6000)
	cleanLedger(tcOrderContract, "order-")
	createOrderConcurrently(tcOrderContract, 7000)
	cleanLedger(tcOrderContract, "order-")
	createOrderConcurrently(tcOrderContract, 8000)
	cleanLedger(tcOrderContract, "order-")
	createOrderConcurrently(tcOrderContract, 9000)
	cleanLedger(tcOrderContract, "order-")
	createOrderConcurrently(tcOrderContract, 10000)
	cleanLedger(tcOrderContract, "order-")
	createOrderConcurrently(tcOrderContract, 100000)
	cleanLedger(tcOrderContract, "order-")
	cleanLedger(contract, "good-")
}

func initGoods(contract *client.Contract) {
	fmt.Printf("\n--> Init Goods: Initiate 10 goods \n")
	prefixId := "good-"
	for i := 0; i < 10; i++ {

		goodId := fmt.Sprintf("%s%06d", prefixId, i)
		goodName := fmt.Sprintf("Good number %d", i)
		fmt.Printf("\n*** Submitting Transaction: CreateGood with params: goodId: %s, goodName: %s, unit: kg, importLimit: 100.00, exportLimit 200.0\n\n", goodId, goodName)
		timeStart := time.Now()
		_, err := contract.SubmitTransaction("CreateGood", goodId, goodName, "kg", "100.00", "200.0")
		secondDuration := time.Now().Sub(timeStart).Seconds()
		if err != nil {
			fmt.Println("\n*** Submit Transaction error")
			fmt.Println(err)
		}
		fmt.Printf("%s created in %f seconds\n", goodId, secondDuration)
	}
}

func readGoods(contract *client.Contract) {
	fmt.Println("\n--> Read Goods: read previously created 10 goods\n")
	for i := 0; i < 10; i++ {
		timeStart := time.Now()
		goodId := fmt.Sprintf("good-%06d", i)
		fmt.Printf("\n*** Reading good %s\n", goodId)
		evaluateResult, err := contract.EvaluateTransaction("GetGoodById", goodId)
		secondsDuration := time.Now().Sub(timeStart).Seconds()
		if err != nil {
			fmt.Println("\n*** Failed to get data")
			fmt.Println(err)
		}

		result := string(evaluateResult)
		fmt.Println("\n*** Retrieved good: " + result)
		fmt.Printf("time to read: %f seconds\n", secondsDuration)
	}
}

func createOrderSuccess(contract *client.Contract) {

	fmt.Printf("\n--> Submit Transaction: CreateOrder, Creating an order \n")
	log.Println("INFO: creating order with 5 goods that comply with the rule")
	goods := []Good{
		{
			GoodId:   "good-000000",
			Quantity: "5",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000001",
			Quantity: "6",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000002",
			Quantity: "2",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000003",
			Quantity: "8.5",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000004",
			Quantity: "3.25",
			Unit:     "kg",
		},
	}
	for _, good := range goods {
		fmt.Println(good)
	}

	goodJson, _ := json.Marshal(goods)
	goodStr := string(goodJson)
	fmt.Println("\n*** Submitting transaction with params: orderId: order-000000, orderType: import, country: US, name: First Order, description: An order description,\ngoods: " + goodStr)
	timeStart := time.Now()
	_, err := contract.SubmitTransaction("CreateOrder", "order-000000", "import", "US", "First Order", "An order description", goodStr)
	timeDuration := time.Now().Sub(timeStart)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("\n*** Transaction committed successfully")
	fmt.Printf("\n*** Create Order duration: %f seconds\n", timeDuration.Seconds())
}

func createOrderFailed(contract *client.Contract) {
	fmt.Printf("\n--> Submit Transaction: CreateOrder, Creating an order \n")
	log.Println("INFO: creating order with goods that does not comply with rules")

	goods := []Good{
		{
			GoodId:   "good-000000",
			Quantity: "500",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000001",
			Quantity: "6",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000002",
			Quantity: "2",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000003",
			Quantity: "8.5",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000004",
			Quantity: "3.25",
			Unit:     "kg",
		},
	}
	for _, good := range goods {
		fmt.Println(good)
	}

	goodJson, _ := json.Marshal(goods)
	goodStr := string(goodJson)

	log.Println("INFO: Submitting invalid order")

	_, err := contract.SubmitTransaction("CreateOrder", "order-000001", "import", "US", "First Order", "An order description", goodStr)
	if err != nil {
		fmt.Println(err)
		fmt.Println("\n*** Transaction not committed")
		return
	}
	fmt.Println("\n*** Transaction committed successfully")
}

func readOrder(contract *client.Contract) {
	fmt.Println("\n--> Evaluate Transaction: ReadOrder: retrieving previously created order")
	res, _ := contract.EvaluateTransaction("ReadOrder", "order-000000")
	fmt.Println(string(res))
}

func createGoodUnauthorized(contract *client.Contract) {
	fmt.Println("\n--> Submit Transaction: CreateGood: creating good from unauthorized organization: tradingcompany1.example.com\n")
	_, err := contract.SubmitTransaction("CreateGood", "good-100123", "Good number 100123", "kg", "100.00", "200.0")
	if err != nil {
		fmt.Println(err)
		fmt.Println("*** Transaction failed")
	}
}

func createOrder10NGoods(contract *client.Contract, orderId string, n int) {
	fmt.Printf("\n--> Submit Transaction: CreateOrder: create order with %d goods\n\n", 10*n)
	templateGoods := []Good{
		{
			GoodId:   "good-000000",
			Quantity: "5",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000001",
			Quantity: "6",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000002",
			Quantity: "2",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000003",
			Quantity: "8.5",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000004",
			Quantity: "3.25",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000005",
			Quantity: "5",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000006",
			Quantity: "6",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000007",
			Quantity: "2",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000008",
			Quantity: "8.5",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000009",
			Quantity: "3.25",
			Unit:     "kg",
		},
	}
	var goods []Good
	for i := 0; i < n; i++ {
		goods = append(goods, templateGoods...)
	}
	goodsJson, _ := json.Marshal(goods)
	timeStart := time.Now()
	_, err := contract.SubmitTransaction("CreateOrder", orderId, "import", "US", "N-good Ooder", "An order description", string(goodsJson))
	secondsDuration := time.Now().Sub(timeStart).Seconds()
	if err != nil {
		fmt.Println(err)
		fmt.Println("***Submit Transaction failed")
		return
	}
	fmt.Printf("\n*** Transaction Committed: %s created with %d good(s) in %f seconds\n", orderId, 10*n, secondsDuration)
}

func createOrderConcurrently(contract *client.Contract, n int) {
	fmt.Printf("\n--> Submit Transaction: submitting %d concurrent requests\n", n)
	c := make(chan bool)
	goods := []Good{
		{
			GoodId:   "good-000000",
			Quantity: "5",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000001",
			Quantity: "6",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000002",
			Quantity: "2",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000003",
			Quantity: "8.5",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000004",
			Quantity: "3.25",
			Unit:     "kg",
		},
		{
			GoodId:   "good-000005",
			Quantity: "5",
			Unit:     "kg",
		},
	}
	goodsJson, _ := json.Marshal(goods)
	goodsStr := string(goodsJson)
	timeStart := time.Now()
	for i := 0; i < n; i++ {
		orderId := fmt.Sprintf("order-%06d", i)
		go func(contract *client.Contract, orderId, goods string, c chan bool) {
			_, err := contract.SubmitTransaction("CreateOrder", orderId, "import", "US", "concurrent order", "Order description", goods)
			if err != nil {
				c <- false
				return
			}

			c <- true
		}(contract, orderId, goodsStr, c)
	}
	failedTx := 0
	for i := 0; i < n; i++ {
		res := <-c
		if res == false {
			failedTx++
		}
	}
	secondsDuration := time.Now().Sub(timeStart).Seconds()
	if failedTx > 0 {
		fmt.Printf("*** There are %d failed request(s) on %d concurrent request(s)\n", failedTx, n)
	}
	fmt.Printf("*** %d concurrent request(s) finished in %f seconds\n", n, secondsDuration)
}

func cleanLedger(contract *client.Contract, prefix string) {
	_, err := contract.SubmitTransaction("ClearData", prefix+"000000", prefix+"order-999999")
	if err != nil {
		fmt.Println(err)
	}
}
