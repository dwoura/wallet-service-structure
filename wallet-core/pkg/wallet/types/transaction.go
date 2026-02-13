package types

// UnsignedTransaction represents a transaction waiting to be signed.
// It contains all necessary fields for a cold wallet to sign, plus metadata for the user to verify.
type UnsignedTransaction struct {
	Chain    string `json:"chain"`          // ETH, BTC
	From     string `json:"from"`           // Sender Address
	To       string `json:"to"`             // Recipient Address
	Amount   string `json:"amount"`         // Amount in base unit (Wei, Satoshi)
	Nonce    uint64 `json:"nonce"`          // Account Nonce (ETH)
	GasLimit uint64 `json:"gas_limit"`      // Gas Limit (ETH)
	GasPrice string `json:"gas_price"`      // Gas Price in Wei (ETH)
	Data     string `json:"data,omitempty"` // Contract Data (Hex)

	// DerivationPath is crucial for the signer to know which key to use
	// e.g., "m/44'/60'/0'/0/0"
	DerivationPath string `json:"derivation_path"`

	// ChainID for EIP-155 replay protection
	ChainID int64 `json:"chain_id"`
}

// SignedTransaction represents the result of the signing process.
type SignedTransaction struct {
	TxHash string `json:"tx_hash"` // Transaction Hash
	RawTx  string `json:"raw_tx"`  // RLP Encoded Hex String (ready to broadcast)
}
