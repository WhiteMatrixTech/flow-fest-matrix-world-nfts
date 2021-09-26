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
