deploy-to-local:
	bash ./sh/deploy-to-local.sh
.PHONY: deploy-to-local

deploy-to-testnet:
	bash ./sh/deploy-to-testnet.sh
.PHONY: deploy-to-testnet

create-collection-nft:
	cd go && go run create_collection/main.go
.PHONY:create-collection-nft

mint-nft:
	cd go && go run mint_nft/main.go
.PHONY: mint-nft

list-nfts:
	cd go && go run list_nfts/main.go
.PHONY: list-nfts

bootstrap-local:
	bash ./sh/bootstrap-local.sh
.PHONY: bootstrap-local

upgrade-to-main:
	flow accounts update-contract MatrixWorldFlowFestNFT ./cadence/contracts/MatrixWorldFlowFestNFTUpgradeMain.cdc --signer main -f ./flow.json -f ./flow.main.json -n mainnet
