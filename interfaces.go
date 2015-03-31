package kala

// A Minter provides methods for minting unique IDs
type Minter interface {
	Mint() (string, error)
}
