import NonFungibleToken from "./lib/NonFungibleToken.cdc"

pub contract MatrixWorldFlowFestNFTS2: NonFungibleToken {
    pub event ContractInitialized()
    pub event Withdraw(id: UInt64, from: Address?)
    pub event Deposit(id: UInt64, to: Address?)
    pub event Minted(id: UInt64, name: String, description:String, animationUrl:String, hash: String, type: String)

    pub let CollectionStoragePath: StoragePath
    pub let CollectionPublicPath: PublicPath
    pub let MinterStoragePath: StoragePath

    pub var totalSupply: UInt64

    pub resource interface NFTPublic {
        pub let id: UInt64
        pub let metadata: Metadata
    }

    pub struct Metadata {
        pub let name: String
        pub let description: String
        pub let animationUrl: String
        pub let hash: String
        pub let type: String

        init(name: String, description: String, animationUrl: String, hash: String, type: String) {
            self.name = name
            self.description = description
            self.animationUrl = animationUrl
            self.hash = hash
            self.type = type
        }
    }

   pub resource NFT: NonFungibleToken.INFT, NFTPublic {
        pub let id: UInt64
        pub let metadata: Metadata
        init(initID: UInt64,metadata: Metadata) {
            self.id = initID
            self.metadata=metadata
        }
    }

    pub resource interface MatrixWorldFlowFestNFTCollectionPublic {
        pub fun deposit(token: @NonFungibleToken.NFT)
        pub fun getIDs(): [UInt64]
        pub fun borrowNFT(id: UInt64): &NonFungibleToken.NFT
        pub fun borrowArt(id: UInt64): &MatrixWorldFlowFestNFTS2.NFT? {
            post {
                (result == nil) || (result?.id == id):
                    "Cannot borrow MatrixWorldFlowFestNFT reference: The ID of the returned reference is incorrect"
            }
        }
    }

    pub resource Collection: MatrixWorldFlowFestNFTCollectionPublic, NonFungibleToken.Provider, NonFungibleToken.Receiver, NonFungibleToken.CollectionPublic {
        pub var ownedNFTs: @{UInt64: NonFungibleToken.NFT}

        pub fun withdraw(withdrawID: UInt64): @NonFungibleToken.NFT {
            let token <- self.ownedNFTs.remove(key: withdrawID) ?? panic("missing NFT")

            emit Withdraw(id: token.id, from: self.owner?.address)

            return <-token
        }

        pub fun deposit(token: @NonFungibleToken.NFT) {
            let token <- token as! @MatrixWorldFlowFestNFTS2.NFT

            let id: UInt64 = token.id

            let oldToken <- self.ownedNFTs[id] <- token

            emit Deposit(id: id, to: self.owner?.address)

            destroy oldToken
        }


        pub fun getIDs(): [UInt64] {
            return self.ownedNFTs.keys
        }

        pub fun borrowNFT(id: UInt64): &NonFungibleToken.NFT {
            return &self.ownedNFTs[id] as &NonFungibleToken.NFT
        }

        pub fun borrowArt(id: UInt64): &MatrixWorldFlowFestNFTS2.NFT? {
            if self.ownedNFTs[id] != nil {
                let ref = &self.ownedNFTs[id] as auth &NonFungibleToken.NFT
                return ref as! &MatrixWorldFlowFestNFTS2.NFT
            } else {
                return nil
            }
        }

        destroy() {
            destroy self.ownedNFTs
        }

        init () {
            self.ownedNFTs <- {}
        }
    }

    pub fun createEmptyCollection(): @NonFungibleToken.Collection {
        return <- create Collection()
    }

    pub struct NftData {
        pub let metadata: MatrixWorldFlowFestNFTS2.Metadata
        pub let id: UInt64
        init(metadata: MatrixWorldFlowFestNFTS2.Metadata, id: UInt64) {
            self.metadata= metadata
            self.id=id
        }
    }

    pub fun getNft(address:Address) : [NftData] {
        var artData: [NftData] = []
        let account = getAccount(address)

        if let artCollection = account.getCapability(self.CollectionPublicPath).borrow<&{MatrixWorldFlowFestNFTS2.MatrixWorldFlowFestNFTCollectionPublic}>()  {
            for id in artCollection.getIDs() {
                var art = artCollection.borrowArt(id: id)
                artData.append(NftData(metadata: art!.metadata,id: id))
            }
        }
        return artData
    }

	pub resource NFTMinter {
		pub fun mintNFT(
            recipient: &{NonFungibleToken.CollectionPublic},
            name: String,
            description: String,
            animationUrl: String
            hash: String,
            type: String) {
            emit Minted(id: MatrixWorldFlowFestNFTS2.totalSupply, name: name, description: description, animationUrl: animationUrl, hash: hash, type: type)

			recipient.deposit(token: <-create MatrixWorldFlowFestNFTS2.NFT(
			    initID: MatrixWorldFlowFestNFTS2.totalSupply,
			    metadata: Metadata(
                    name: name,
                    description:description,
                    animationUrl: animationUrl,
                    hash: hash,
                    type: type
                )))

            MatrixWorldFlowFestNFTS2.totalSupply = MatrixWorldFlowFestNFTS2.totalSupply + (1 as UInt64)
		}
	}

    init() {
        self.CollectionStoragePath = /storage/MatrixWorldFlowFestNFTS2Collection
        self.CollectionPublicPath = /public/MatrixWorldFlowFestNFTS2Collection
        self.MinterStoragePath = /storage/MatrixWorldFlowFestNFTS2Minter

        self.totalSupply = 0

        let minter <- create NFTMinter()
        self.account.save(<-minter, to: self.MinterStoragePath)

        emit ContractInitialized()
    }
}
