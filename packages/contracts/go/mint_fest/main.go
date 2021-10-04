package main

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"

	"encoding/csv"
	"go/common"
	"math"
	"os"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
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

var mintNFTs string = fmt.Sprintf(`
import FungibleToken from 0x%s
import NonFungibleToken from 0x%s
import %s from 0x%s
transaction(recipient: Address, names: [String], descriptions: [String], animationUrls: [String], hashes: [String], types: [String]) {
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
        var size = names.length
        while size > 0 {
            let idx = names.length - size
            self.minter.mintNFT(recipient: receiver, name: names[idx], description: descriptions[idx], animationUrl: animationUrls[idx], hash: hashes[idx], type: types[idx])
            size = size - 1
        }
    }
}`, common.Config.FungibleTokenAddress, common.Config.NonFungibleTokenAddress, common.Config.ContractName, common.Config.ContractAddress, common.Config.ContractName, common.Config.ContractName, common.Config.ContractName, common.Config.ContractName)

var getNFTUUID string = fmt.Sprintf(`
import NonFungibleToken from 0x%s
import %s from 0x%s
pub fun main(id: UInt64) : UInt64{

    let account = getAccount(0x%s)

    let acctCapability = account.getCapability(%s.CollectionPublicPath)
    let receiverRef = acctCapability.borrow<&{NonFungibleToken.CollectionPublic}>()
        ?? panic("Could not borrow account receiver reference")
    return receiverRef.borrowNFT(id:id).uuid
}
`, common.Config.NonFungibleTokenAddress, common.Config.ContractName, common.Config.ContractAddress, common.Config.SingerAddress, common.Config.ContractName)

var getNFTUUIDs string = fmt.Sprintf(`
import NonFungibleToken from 0x%s
import %s from 0x%s
pub fun main(ids: [UInt64]) : [UInt64]{

    let account = getAccount(0x%s)

    let acctCapability = account.getCapability(%s.CollectionPublicPath)
    let receiverRef = acctCapability.borrow<&{NonFungibleToken.CollectionPublic}>()
        ?? panic("Could not borrow account receiver reference")
    let uuids: [UInt64] = []
    var size = ids.length
    while size > 0 {
        let idx = ids.length - size
        uuids.append(receiverRef.borrowNFT(id: ids[idx]).uuid)
        size = size - 1
    }
    return uuids
}
`, common.Config.NonFungibleTokenAddress, common.Config.ContractName, common.Config.ContractAddress, common.Config.SingerAddress, common.Config.ContractName)

func writeLock(fileName string, data [][]string) {
	recordFile, err := os.Create(fileName)
	if err != nil {
		fmt.Println("An error encountered ::", err)
	}

	// 2. Initialize the writer
	writer := csv.NewWriter(recordFile)

	// 3. Write all the records
	err = writer.WriteAll(data) // returns error
	if err != nil {
		fmt.Println("An error encountered ::", err)
	}
}

func main() {
	fileName := os.Args[1]
	batchSize := 100

	lockFileName := fileName + ".lock"
	_, err := os.Stat(lockFileName) // file does not exist
	if os.IsNotExist(err) {
		fmt.Println("No lock file found")
	} else {
		fileName = lockFileName
		fmt.Println("Found lock file continue last processing")
	}
	// 1. Open the file
	recordFile, err := os.Open(fileName)
	if err != nil {
		fmt.Println("An error encountered ::", err)
	}
	// 2. Initialize the reader
	reader := csv.NewReader(recordFile)
	// 3. Read all the records
	records, _ := reader.ReadAll()

	fmt.Println(len(records) - 1)

	// setup FCL
	fmt.Println(common.Config.Node)
	ctx := context.Background()
	flowClient, err := client.New(common.Config.Node, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	acctAddress, acctKey, signer := common.ServiceAccount(flowClient, common.Config.SingerAddress, common.Config.SingerPriv)

	for i := 1; i < len(records); i += batchSize {
		names := make([]cadence.Value, 0)
		descriptions := make([]cadence.Value, 0)
		aniUrls := make([]cadence.Value, 0)
		hashes := make([]cadence.Value, 0)
		typeNames := make([]cadence.Value, 0)

		fmt.Println(len(records[i]))
		fmt.Println(records[i])

		var result *flow.TransactionResult = nil

		for j := i; j < int(math.Min(float64(i+batchSize), float64(len(records)))); j++ {
			rowInfo := records[j]
			name, _ := cadence.NewValue(rowInfo[2])
			names = append(names, name)
			des, _ := cadence.NewValue(fmt.Sprintf("%s, %s/%s", rowInfo[3], rowInfo[6], rowInfo[7]))
			descriptions = append(descriptions, des)
			aniUrl, _ := cadence.NewValue(rowInfo[4])
			aniUrls = append(aniUrls, aniUrl)
			hash, _ := cadence.NewValue("")
			hashes = append(hashes, hash)
			typeName, _ := cadence.NewValue(rowInfo[5])
			typeNames = append(typeNames, typeName)
		}

		rowInfo := records[i]
		if rowInfo[11] != "" {
			if rowInfo[0] == "" || rowInfo[1] == "" {
				txId := flow.HexToID(rowInfo[len(rowInfo)-1])
				result = common.WaitForSeal(ctx, flowClient, txId)
				fmt.Println("tx.ID().String() ---- ", txId.String())
				fmt.Println("tx.ID().String():Result ---- ", result.Events)
			} else {
				continue
			}
		} else {

			referenceBlock, err := flowClient.GetLatestBlock(ctx, false)
			if err != nil {
				panic(err)
			}

			proposalAccount, err := flowClient.GetAccountAtLatestBlock(ctx, acctAddress)
			tx := flow.NewTransaction().
				SetScript([]byte(mintNFTs)).
				SetGasLimit(9999).
				SetProposalKey(acctAddress, acctKey.Index, proposalAccount.Keys[acctKey.Index].SequenceNumber).
				SetReferenceBlockID(referenceBlock.ID).
				SetPayer(acctAddress).
				AddAuthorizer(acctAddress)

			if err := tx.AddArgument(cadence.NewAddress(flow.HexToAddress(common.Config.SingerAddress))); err != nil {
				panic(err)
			}

			namesC := cadence.NewArray(names)
			if err := tx.AddArgument(namesC); err != nil {
				panic(err)
			}

			desC := cadence.NewArray(descriptions)
			if err := tx.AddArgument(desC); err != nil {
				panic(err)
			}

			aniUrlsC := cadence.NewArray(aniUrls)
			if err := tx.AddArgument(aniUrlsC); err != nil {
				panic(err)
			}

			hashesC := cadence.NewArray(hashes)
			if err := tx.AddArgument(hashesC); err != nil {
				panic(err)
			}

			typeNamesC := cadence.NewArray(typeNames)
			if err := tx.AddArgument(typeNamesC); err != nil {
				panic(err)
			}

			if err := tx.SignEnvelope(acctAddress, acctKey.Index, signer); err != nil {
				panic(err)
			}

			if err := flowClient.SendTransaction(ctx, *tx); err != nil {
				panic(err)
			}

			fmt.Println("send tx.ID().String() ---- ", tx.ID().String())
			for j := 0; j < len(names); j++ {
				records[j+i][11] = tx.ID().String()
			}
			writeLock(lockFileName, records)
			result = common.WaitForSeal(ctx, flowClient, tx.ID())
			fmt.Println("Transaction complete!")
		}

		if result == nil || result.Error != nil {
			panic("Something is wrong with")
		}
		fmt.Println(len(names))
		mintCount := 0
		tokenIds := make([]cadence.Value, 0)
		for k := 0; k < len(result.Events); k++ {
			event := result.Events[k]
			if strings.Contains(event.Type, common.Config.ContractName+".Minted") {
				tokenId := event.Value.Fields[0].(cadence.UInt64)
				records[i+mintCount][0] = tokenId.String()

				tokenIds = append(tokenIds, tokenId)

				mintCount++
			}
		}
		if _, err := flowClient.GetLatestBlock(ctx, false); err != nil {
			panic(err)
		}
		value, err := flowClient.ExecuteScriptAtLatestBlock(ctx, []byte(getNFTUUIDs), []cadence.Value{cadence.NewArray(tokenIds)})
		if err != nil {
			panic("failed to execute script")
		}

		uuids := value.(cadence.Array).Values
		for j := 0; j < len(uuids); j++ {
			records[i+j][1] = uuids[j].String()
		}
		fmt.Println(mintCount)
		fmt.Println(fmt.Sprintf("finish %d-%d\n", i, i+len(names)))
		writeLock(lockFileName, records)
	}
}
