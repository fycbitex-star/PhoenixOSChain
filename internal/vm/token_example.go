package vm

// BasicTokenExample stores an initial supply in storage slot 0 and logs it.
func BasicTokenExample(initialSupply uint64) []byte {
	return Program(
		Push(initialSupply),
		Push(0),
		Op(OpStore),
		Push(initialSupply),
		Op(OpLog),
		Op(OpStop),
	)
}
