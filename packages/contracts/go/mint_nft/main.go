package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"go/common"
)

var mintNFT string = fmt.Sprintf(`
import FungibleToken from 0x%s
import NonFungibleToken from 0x%s
import %s from 0x%s
transaction(recipient: Address, name: String, description: String, animationUrl: String, hash: String, type: String) {
    let minter: &%s.NFTMinter
    prepare(signer: AuthAccount) {
        self.minter = signer.borrow<&%s.NFTMinter>(from: %s.MinterStoragePath)
            ?? panic("Could not borrow a reference to the NFT minter")
    }
    execute {
        let recipient = getAccount(recipient)
        let receiver = recipient
            .getCapability(%s.CollectionPublicPath)!
            .borrow<&{NonFungibleToken.CollectionPublic}>()
            ?? panic("Could not get receiver reference to the NFT Collection")
        self.minter.mintNFT(recipient: receiver, name: name, description: description, animationUrl: animationUrl, hash: hash, type: type)
    }
}`, common.Config.FungibleTokenAddress, common.Config.NonFungibleTokenAddress, common.Config.ContractName, common.Config.ContractAddress, common.Config.ContractName, common.Config.ContractName, common.Config.ContractName, common.Config.ContractName)

func main() {
	fmt.Println(common.Config.Node) // Ahoy, world!
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
		SetScript([]byte(mintNFT)).
		SetGasLimit(100).
		SetProposalKey(acctAddress, acctKey.Index, acctKey.SequenceNumber).
		SetReferenceBlockID(referenceBlock.ID).
		SetPayer(acctAddress).
		AddAuthorizer(acctAddress)

	if err := tx.AddArgument(cadence.NewAddress(flow.HexToAddress(common.Config.SingerAddress))); err != nil {
		panic(err)
	}

	name, _ := cadence.NewString("Example Name")

	if err := tx.AddArgument(name); err != nil {
		panic(err)
	}

	des, _ := cadence.NewString("Example Description")
	if err := tx.AddArgument(des); err != nil {
		panic(err)
	}
	aniUrl, _ := cadence.NewString("Example AnimationUrl")
	if err := tx.AddArgument(aniUrl); err != nil {
		panic(err)
	}

	hash, _ := cadence.NewString("Example Hash")
	if err := tx.AddArgument(hash); err != nil {
		panic(err)
	}

	typeName, _ := cadence.NewString("Example Type")
	if err := tx.AddArgument(typeName); err != nil {
		panic(err)
	}

	if err := tx.SignEnvelope(acctAddress, acctKey.Index, signer); err != nil {
		panic(err)
	}

	if err := flowClient.SendTransaction(ctx, *tx); err != nil {
		panic(err)
	}

	common.WaitForSeal(ctx, flowClient, tx.ID())
	fmt.Println("Transaction complet!")
	fmt.Println("tx.ID().String() ---- ", tx.ID().String())
}
