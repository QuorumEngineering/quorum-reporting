package rpc

import (
	"math/big"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"quorumengineering/quorum-report/database/memory"
	"quorumengineering/quorum-report/types"
)

const validABI = `
[
	{"constant":true,"inputs":[],"name":"storedData","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},
	{"constant":false,"inputs":[{"name":"_x","type":"uint256"}],"name":"set","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},
	{"constant":true,"inputs":[],"name":"get","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},
	{"inputs":[{"name":"_initVal","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"},
	{"anonymous":false,"inputs":[{"indexed":false,"name":"_value","type":"uint256"}],"name":"valueSet","type":"event"}
]`

var (
	dummyReq = &http.Request{}

	addr  = types.NewAddress("0x0000000000000000000000000000000000000001")
	block = &types.Block{
		Hash:   types.NewHash("0xc7fd1915b4b8ac6344e750e4eaeacf9114d4e185f9c10b6b3bc7049511a96998"),
		Number: 1,
		Transactions: []types.Hash{
			types.NewHash("0x1a6f4292bac138df9a7854a07c93fd14ca7de53265e8fe01b6c986f97d6c1ee7"), types.NewHash("0xbc77a72b3409ba3e098cb45bac1b7727b59dae9a05f37a0dbc61007949c8cede"), types.NewHash("0xb2d58900a820afddd1d926845e7655d445885524b9af1cc946b45949be74cc08"),
		},
		ParentHash:  types.NewHash(""),
		ReceiptRoot: types.NewHash(""),
		TxRoot:      types.NewHash(""),
		StateRoot:   types.NewHash(""),
	}
	tx1 = &types.Transaction{ // deployment
		Hash:            types.NewHash("0x1a6f4292bac138df9a7854a07c93fd14ca7de53265e8fe01b6c986f97d6c1ee7"),
		BlockNumber:     1,
		From:            types.NewAddress("0x0000000000000000000000000000000000000009"),
		To:              "",
		Data:            types.NewHexData("0x608060405234801561001057600080fd5b506040516020806101a18339810180604052602081101561003057600080fd5b81019080805190602001909291905050508060008190555050610149806100586000396000f3fe608060405234801561001057600080fd5b506004361061005e576000357c0100000000000000000000000000000000000000000000000000000000900480632a1afcd91461006357806360fe47b1146100815780636d4ce63c146100af575b600080fd5b61006b6100cd565b6040518082815260200191505060405180910390f35b6100ad6004803603602081101561009757600080fd5b81019080803590602001909291905050506100d3565b005b6100b7610114565b6040518082815260200191505060405180910390f35b60005481565b806000819055507fefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36816040518082815260200191505060405180910390a150565b6000805490509056fea165627a7a7230582061f6956b053dbf99873b363ab3ba7bca70853ba5efbaff898cd840d71c54fc1d0029000000000000000000000000000000000000000000000000000000000000002a"),
		CreatedContract: addr,
	}
	tx2 = &types.Transaction{ // set
		Hash:            types.NewHash("0xbc77a72b3409ba3e098cb45bac1b7727b59dae9a05f37a0dbc61007949c8cede"),
		BlockNumber:     1,
		From:            types.NewAddress("0x0000000000000000000000000000000000000009"),
		To:              addr,
		Data:            types.NewHexData("0x60fe47b100000000000000000000000000000000000000000000000000000000000003e7"),
		CreatedContract: "",
	}
	tx3 = &types.Transaction{ // private
		Hash:            types.NewHash("0xb2d58900a820afddd1d926845e7655d445885524b9af1cc946b45949be74cc08"),
		BlockNumber:     1,
		From:            types.NewAddress("0x0000000000000000000000000000000000000009"),
		To:              addr,
		PrivateData:     types.NewHexData("0x60fe47b100000000000000000000000000000000000000000000000000000000000003e8"),
		CreatedContract: "",
		Events: []*types.Event{
			{
				Data:    types.NewHexData("0x00000000000000000000000000000000000000000000000000000000000003e8"),
				Address: addr,
				Topics:  []types.Hash{types.NewHash("0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36")},
			},
		},
		InternalCalls: []*types.InternalCall{
			{
				Type: "CALL",
				To:   addr,
			},
		},
	}
)

func TestAPIValidation(t *testing.T) {
	db := memory.NewMemoryDB()
	apis := NewRPCAPIs(db, NewDefaultContractManager(db))

	err := apis.AddAddress(dummyReq, &AddressWithOptionalBlock{}, nil)
	assert.EqualError(t, err, "address not provided")
}

func TestAPIParsing(t *testing.T) {
	db := memory.NewMemoryDB()
	apis := NewRPCAPIs(db, NewDefaultContractManager(db))
	err := apis.AddAddress(dummyReq, &AddressWithOptionalBlock{Address: &addr}, nil)
	assert.Nil(t, err)

	// Test AddABI string to ABI parsing.
	err = apis.AddABI(dummyReq, &AddressWithData{&addr, "hello"}, nil)
	assert.EqualError(t, err, "invalid character 'h' looking for beginning of value")

	err = apis.AddABI(dummyReq, &AddressWithData{&addr, validABI}, nil)
	assert.Nil(t, err)

	// Set up test data.
	err = db.WriteTransactions([]*types.Transaction{tx1, tx2, tx3})
	assert.Nil(t, err)

	err = db.WriteBlocks([]*types.Block{block})
	assert.Nil(t, err)
	// Test GetTransaction parse transaction data.
	parsedTx1 := &types.ParsedTransaction{}
	err = apis.GetTransaction(dummyReq, &tx1.Hash, parsedTx1)
	assert.Nil(t, err)
	assert.Equal(t, "constructor(uint256 _initVal)", parsedTx1.Sig)
	assert.Equal(t, big.NewInt(42), parsedTx1.ParsedData["_initVal"])

	parsedTx2 := &types.ParsedTransaction{}
	err = apis.GetTransaction(dummyReq, &tx2.Hash, parsedTx2)
	assert.Nil(t, err)
	assert.Equal(t, "set(uint256 _x)", parsedTx2.Sig)
	assert.Equal(t, big.NewInt(999), parsedTx2.ParsedData["_x"])
	assert.Equal(t, "0x60fe47b1", parsedTx2.Func4Bytes.String())

	parsedTx3 := &types.ParsedTransaction{}
	err = apis.GetTransaction(dummyReq, &tx3.Hash, parsedTx3)
	assert.Nil(t, err)
	assert.Equal(t, "event valueSet(uint256 _value)", parsedTx3.ParsedEvents[0].Sig)
	assert.Equal(t, big.NewInt(1000), parsedTx3.ParsedEvents[0].ParsedData["_value"])

	// Test GetAllEventsFromAddress parse event.
	err = db.IndexBlocks([]types.Address{addr}, []*types.Block{block})
	assert.Nil(t, err)

	eventsResp := &EventsResp{}
	err = apis.GetAllEventsFromAddress(dummyReq, &AddressWithOptions{Address: &addr}, eventsResp)
	assert.Nil(t, err)
	assert.Equal(t, "event valueSet(uint256 _value)", eventsResp.Events[0].Sig)
	assert.Equal(t, big.NewInt(1000), eventsResp.Events[0].ParsedData["_value"])
}

func TestAddAddressWithFrom(t *testing.T) {
	db := memory.NewMemoryDB()
	apis := NewRPCAPIs(db, NewDefaultContractManager(db))
	from := uint64(100)

	params := &AddressWithOptionalBlock{
		Address:     &addr,
		BlockNumber: &from,
	}

	err := apis.AddAddress(dummyReq, params, nil)
	assert.Nil(t, err)

	lastFiltered, err := db.GetLastFiltered(addr)
	assert.Nil(t, err)
	assert.Equal(t, from-1, lastFiltered)
}
