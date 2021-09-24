package common

import (
	"os"
)

type FlowConfig struct {
	Node                    string
	FungibleTokenAddress    string
	NonFungibleTokenAddress string
	FlowTokenAddress        string

	SingerAddress string
	SingerPriv    string
	ContractName  string
	ContractAddress string
}

var (
	// Test
	Config = FlowConfig{
		Node:                    os.Getenv("NODE"),
		FungibleTokenAddress:    os.Getenv("FUNTOKEN_ADDRESS"),
		NonFungibleTokenAddress: os.Getenv("NONFUNTOKEN_ADDRESS"),
		FlowTokenAddress:        os.Getenv("FLOWTOKEN_ADDRESS"),
		SingerAddress:           os.Getenv("SIGNER_ADDRESS"),
		SingerPriv:              os.Getenv("SIGNER_PRIV"),
		ContractName:            os.Getenv("CONTRACT_NAME"),
		ContractAddress:         os.Getenv("CONTRACT_ADDRESS"),
	}
)
