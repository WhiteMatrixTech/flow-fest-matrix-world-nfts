import MatrixWorldFlowFestNFT from 0x2d2750f240198f91
pub fun main(address: Address): [MatrixWorldFlowFestNFT.NftData]{
    return MatrixWorldFlowFestNFT.getNft(address: address);
}
