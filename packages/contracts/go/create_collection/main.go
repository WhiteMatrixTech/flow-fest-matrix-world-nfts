package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"

	"go/common"
)

var createStorage string = fmt.Sprintf(`
import NonFungibleToken from 0x%s
import %s from 0x%s
transaction {
    prepare(signer: AuthAccount) {
        if signer.borrow<&%s.Collection>(from: %s.CollectionStoragePath) == nil {
            let collection <- %s.createEmptyCollection()

            signer.save(<-collection, to: %s.CollectionStoragePath)
            signer.link<&%s.Collection{NonFungibleToken.CollectionPublic, %s.%sCollectionPublic}>(%s.CollectionPublicPath, target: %s.CollectionStoragePath)
        }
    }
}`, common.Config.NonFungibleTokenAddress, common.Config.ContractName, common.Config.ContractAddress, common.Config.ContractName, common.Config.ContractName, common.Config.ContractName, common.Config.ContractName, common.Config.ContractName, common.Config.ContractName, common.Config.ContractName, common.Config.ContractName, common.Config.ContractName)

func main() {
	ctx := context.Background()
	flowClient, err := client.New(common.Config.Node, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	referenceBlock, err := flowClient.GetLatestBlock(ctx, false)
	if err != nil {
		panic(err)
	}

	acctAddress, acctKey, signer := common.ServiceAccount(flowClient, common.Config.SingerAddress, common.Config.SingerPriv)
	tx := flow.NewTransaction().
		SetScript([]byte(createStorage)).
		SetGasLimit(100).
		SetProposalKey(acctAddress, acctKey.Index, acctKey.SequenceNumber).
		SetReferenceBlockID(referenceBlock.ID).
		SetPayer(acctAddress).
		AddAuthorizer(acctAddress)

	if err := tx.SignEnvelope(acctAddress, acctKey.Index, signer); err != nil {
		panic(err)
	}

	if err := flowClient.SendTransaction(ctx, *tx); err != nil {
		panic(err)
	}

	common.WaitForSeal(ctx, flowClient, tx.ID())
}
