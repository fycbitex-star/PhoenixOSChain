package state

func (db *StateDB) BalanceOf(address string) uint64 {
	return db.GetAccount(address).Balance
}

func (db *StateDB) NonceOf(address string) uint64 {
	return db.GetAccount(address).Nonce
}
