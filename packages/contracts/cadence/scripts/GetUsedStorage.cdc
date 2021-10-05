pub fun main(address: Address): UInt64 {
    let account = getAccount(address)
    log(account.storageUsed)
    log(account.storageCapacity)
    return account.storageUsed
}
