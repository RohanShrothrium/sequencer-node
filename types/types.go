package types

type Transaction struct {
	SenderAddress   string `json:"sender_address"`
	ContractAddress string `json:"contract_address"`
	GasLimit        uint32 `json:"gas_limit"`
}

type Block struct {
	Height          uint32        `json:"height"`
	LeaderSignature bool          `json:"leaderSignature"`
	QC              []bool        `json:"qc"` // TODO: have to update to BLS agg sig
	TC              []bool        `json:"tc"` // TODO: have to update to BLS agg sig
	Transactions    []Transaction `json:"transactions"`
}

type SignedTimeout struct {
	Height    uint32 `json:"height"`
	PrevHash  string `json:"prevHash"`
	Signature bool   `json:"signature"` // TODO: have to update to sig
}

type SignedResponse struct {
	Height    uint32 `json:"height"`
	PrevHash  string `json:"prevHash"`
	Signature bool   `json:"signature"` // TODO: have to update to sig
}
