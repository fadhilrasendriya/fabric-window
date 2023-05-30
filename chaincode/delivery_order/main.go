package main

import (
	"delivery_order/do_smartcontract"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
)

func main() {
	doCC, err := contractapi.NewChaincode(&do_smartcontract.GoodContract{}, &do_smartcontract.OrderContract{})
	if err != nil {
		log.Panicf("Error creating delivery_order chaincode: %v", err)
	}

	if err := doCC.Start(); err != nil {
		log.Panicf("Error starting delivery_order chaincode: %v", err)
	}
}
