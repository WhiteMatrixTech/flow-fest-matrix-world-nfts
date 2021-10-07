package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/grpc"

	"go/common"
	"math"
	"os"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

var transferNFT string = fmt.Sprintf(`
import FungibleToken from 0xFUNGIBLE_TOKEN_ADDRESS
import NonFungibleToken from 0xNONFUNGIBLE_TOKEN_ADDRESS
import CONTRACT_NAME from 0xCONTRACT_ADDRESS
transaction(recipient: Address, tokenIds: [UInt64]) {
    let senderProvider: &CONTRACT_NAME.Collection
    prepare(signer: AuthAccount) {
        self.senderProvider = signer.borrow<&CONTRACT_NAME.Collection>(from: CONTRACT_NAME.CollectionStoragePath)
            ?? panic("Could not borrow a reference to the NFT senderProvider")
    }
    execute {
        let recipient = getAccount(recipient)

        let receiver = recipient
            .getCapability(CONTRACT_NAME.CollectionPublicPath)!
            .borrow<&{NonFungibleToken.CollectionPublic}>()
            ?? panic("Could not get receiver reference to the NFT Collection")

        var size = tokenIds.length
        while size > 0 {
            let idx = tokenIds.length - size
            let token <- self.senderProvider.withdraw(withdrawID: tokenIds[idx])
            size = size - 1
            receiver.deposit(token:<-token)
        }
    }
}`)

func main() {
	receiverAddress := os.Args[1]
	fromTokenId, err := strconv.Atoi(os.Args[2])
	if err != nil {
		// handle error
		fmt.Println(err)
		os.Exit(2)
	}
	toTokenId, err := strconv.Atoi(os.Args[3])
	if err != nil {
		// handle error
		fmt.Println(err)
		os.Exit(2)
	}
	batchSize := 500

	transferNFT = strings.ReplaceAll(transferNFT, "0xFUNGIBLE_TOKEN_ADDRESS", "0x"+common.Config.FungibleTokenAddress)
	transferNFT = strings.ReplaceAll(transferNFT, "0xNONFUNGIBLE_TOKEN_ADDRESS", "0x"+common.Config.NonFungibleTokenAddress)
	transferNFT = strings.ReplaceAll(transferNFT, "0xCONTRACT_ADDRESS", "0x"+common.Config.ContractAddress)
	transferNFT = strings.ReplaceAll(transferNFT, "CONTRACT_NAME", common.Config.ContractName)

	fmt.Println(common.Config.Node)
	ctx := context.Background()
	flowClient, err := client.New(common.Config.Node, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	acctAddress, acctKey, signer := common.ServiceAccount(flowClient, common.Config.SingerAddress, common.Config.SingerPriv)

	totalNumTokens := toTokenId - fromTokenId
    fmt.Println("totalNumTokens:", totalNumTokens)

	for i := 0; i < totalNumTokens; i += batchSize {
		tokenIds := make([]cadence.Value, 0)

		for j := i; j < int(math.Min(float64(i+batchSize), float64(totalNumTokens))); j++ {
			tokenId := cadence.UInt64(uint64(fromTokenId + j))
			tokenIds = append(tokenIds, tokenId)
		}

		referenceBlock, err := flowClient.GetLatestBlock(ctx, false)
		if err != nil {
			panic(err)
		}

		proposalAccount, err := flowClient.GetAccountAtLatestBlock(ctx, acctAddress)
		if err != nil {
			panic(err)
		}

		tx := flow.NewTransaction().
			SetScript([]byte(transferNFT)).
			SetGasLimit(9999).
			SetProposalKey(acctAddress, acctKey.Index, proposalAccount.Keys[acctKey.Index].SequenceNumber).
			SetReferenceBlockID(referenceBlock.ID).
			SetPayer(acctAddress).
			AddAuthorizer(acctAddress)

		if err := tx.AddArgument(cadence.NewAddress(flow.HexToAddress(receiverAddress))); err != nil {
			panic(err)
		}

		tokenIdsC := cadence.NewArray(tokenIds)
		if err := tx.AddArgument(tokenIdsC); err != nil {
			panic(err)
		}

		if err := tx.SignEnvelope(acctAddress, acctKey.Index, signer); err != nil {
			panic(err)
		}

		if err := flowClient.SendTransaction(ctx, *tx); err != nil {
			panic(err)
		}
		fmt.Println("send tx.ID().String() ---- ", tx.ID().String())
		result := common.WaitForSeal(ctx, flowClient, tx.ID())
		fmt.Println("Transaction complete!")
		if result == nil || result.Error != nil {
			panic("Something is wrong with")
		}
		fmt.Println("Finish", tokenIds)
	}
}
