## Deploy Project to Emulator

Start local emulator

`flow emulator`

Deploy all project

`bash ./sh/deploy-to-local-emulator.sh`

## Create Collection
Setup go script env
```bash
export SIGNER_PRIV=
export SIGNER_ADDRESS=
export NODE=

export FUNTOKEN_ADDRESS=
export NONFUNTOKEN_ADDRESS=
export FLOWTOKEN_ADDRESS=
export CONTRACT_NAME=MatrixWorldFlowFestNFT
export CONTRACT_ADDRESS=

make create-collection-nft
```
