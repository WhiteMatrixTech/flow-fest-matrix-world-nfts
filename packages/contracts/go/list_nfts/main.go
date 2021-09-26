package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk/client"

	"go/common"
)

var listIds string = fmt.Sprintf(`
import NonFungibleToken from 0x%s
import %s from 0x%s
pub fun main() : [UInt64]{

    let account = getAccount(0x%s)

    let acctCapability = account.getCapability(%s.CollectionPublicPath)
    let receiverRef = acctCapability.borrow<&{NonFungibleToken.CollectionPublic}>()
        ?? panic("Could not borrow account receiver reference")
    return receiverRef.getIDs()
}
`, common.Config.NonFungibleTokenAddress, common.Config.ContractName, common.Config.ContractAddress,common.Config.SingerAddress,common.Config.ContractName,)

func main() {
	ctx := context.Background()
	flowClient, err := client.New(common.Config.Node, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	if _, err := flowClient.GetLatestBlock(ctx, false); err != nil {
		panic(err)
	}

	value, err := flowClient.ExecuteScriptAtLatestBlock(ctx, []byte(listIds),[]cadence.Value{})

	if err != nil {
		panic("failed to execute script")
	}

	ID := value.(cadence.Array)

	fmt.Println(ID.ToGoValue())
}
