package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"encoding/csv"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"go/common"
	"os"
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
`, common.Config.NonFungibleTokenAddress, common.Config.ContractName, common.Config.ContractAddress,common.Config.SingerAddress,common.Config.ContractName,)


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
	lockFileName := fileName  + ".lock"
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

	for i := 1; i < len(records); i++ {
		fmt.Println(len(records[i]))
		fmt.Println(records[i])
		rowInfo := records[i]

        var result *flow.TransactionResult = nil; 

		if rowInfo[11] != "" {
            if rowInfo[0] == "" || rowInfo[1] == "" {
                txId := flow.HexToID(rowInfo[len(rowInfo)-1]) 
                result = common.WaitForSeal(ctx, flowClient, txId)
                fmt.Println("tx.ID().String() ---- ", txId.String())
                fmt.Println("tx.ID().String():Result ---- ", result.Events)
            }else{
                continue
            }
		} else {

			referenceBlock, err := flowClient.GetLatestBlock(ctx, false)
			if err != nil {
				panic(err)
			}

            proposalAccount, err := flowClient.GetAccountAtLatestBlock(ctx, acctAddress)
			tx := flow.NewTransaction().
				SetScript([]byte(mintNFT)).
				SetGasLimit(100).
				SetProposalKey(acctAddress, acctKey.Index, proposalAccount.Keys[acctKey.Index].SequenceNumber).
				SetReferenceBlockID(referenceBlock.ID).
				SetPayer(acctAddress).
				AddAuthorizer(acctAddress)

			if err := tx.AddArgument(cadence.NewAddress(flow.HexToAddress(common.Config.SingerAddress))); err != nil {
				panic(err)
			}

			name, _ := cadence.NewString(rowInfo[2])

			if err := tx.AddArgument(name); err != nil {
				panic(err)
			}

			des, _ := cadence.NewString(fmt.Sprintf("%s, %s/%s", rowInfo[3], rowInfo[6], rowInfo[7]))
			if err := tx.AddArgument(des); err != nil {
				panic(err)
			}
			aniUrl, _ := cadence.NewString(rowInfo[4])
			if err := tx.AddArgument(aniUrl); err != nil {
				panic(err)
			}

			hash, _ := cadence.NewString("")
			if err := tx.AddArgument(hash); err != nil {
				panic(err)
			}

			typeName, _ := cadence.NewString(rowInfo[5])
			if err := tx.AddArgument(typeName); err != nil {
				panic(err)
			}

			if err := tx.SignEnvelope(acctAddress, acctKey.Index, signer); err != nil {
				panic(err)
			}

			if err := flowClient.SendTransaction(ctx, *tx); err != nil {
				panic(err)
			}

			fmt.Println("send tx.ID().String() ---- ", tx.ID().String())
			records[i][11] = tx.ID().String()
			writeLock(lockFileName, records)

            result = common.WaitForSeal(ctx, flowClient, tx.ID())
			fmt.Println("Transaction complet!")
		}

        if result == nil || result.Error != nil {
            panic("Something is wrong with")
        }

        mintEvent := result.Events[0]
        tokenId := mintEvent.Value.Fields[0].(cadence.UInt64)
        records[i][0] = tokenId.String()
        if _, err := flowClient.GetLatestBlock(ctx, false); err != nil {
            panic(err)
        }

        value, err := flowClient.ExecuteScriptAtLatestBlock(ctx, []byte(getNFTUUID),[]cadence.Value{tokenId})

        if err != nil {
            panic("failed to execute script")
        }

        uuid := value.(cadence.UInt64)
        records[i][1] = uuid.String()

        writeLock(lockFileName, records)
	}
}
