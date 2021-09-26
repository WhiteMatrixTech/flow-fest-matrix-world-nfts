# Setup env

## Local Env

```bash
export SIGNER_PRIV=a996c6d610d93faf82ad5b15407b66d3a2b72a284b5c2fd4097b5a3e735a79e1
export SIGNER_ADDRESS=01cf0e2f2f715450
export NODE="127.0.0.1:3569"

export FUNTOKEN_ADDRESS=ee82856bf20e2aa6
export NONFUNTOKEN_ADDRESS=f8d6e0586b0a20c7
export FLOWTOKEN_ADDRESS=0ae53cb6e3f42a79
export CONTRACT_NAME=MatrixWorldFlowFestNFT
export CONTRACT_ADDRESS=01cf0e2f2f715450
```

## Mainnet Env
```bash
TBP
```

# GO Interaction

```bash
cd go
```

## Create Collection

```bash
go run create_collection/main.go
```

## Mint Sample NFT

```bash
go run mint_nft/main.go
```
