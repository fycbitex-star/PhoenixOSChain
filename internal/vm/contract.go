package vm

type Address string

type Contract struct {
	Address Address `json:"address"`
	Code    []byte  `json:"code"`
}

type Log struct {
	Contract Address `json:"contract"`
	Value    uint64  `json:"value"`
}

type Receipt struct {
	ContractAddress Address `json:"contract_address,omitempty"`
	GasUsed         uint64  `json:"gas_used"`
	Success         bool    `json:"success"`
	ReturnValue     uint64  `json:"return_value,omitempty"`
	Error           string  `json:"error,omitempty"`
	Logs            []Log   `json:"logs,omitempty"`
}
