package vm

import (
	"encoding/binary"
	"fmt"

	phxcrypto "github.com/phoenixchain/phoenixchain/internal/crypto"
)

type VM struct {
	state *State
}

type ExecutionEnv struct {
	Contract Address
	Caller   string
	GasLimit uint64
}

func New(state *State) *VM {
	return &VM{state: state}
}

func (v *VM) Deploy(code []byte, deployer string, nonce uint64, gasLimit uint64) (Receipt, error) {
	address := Address("pc" + phxcrypto.HashBytes([]byte(fmt.Sprintf("%s:%d:%x", deployer, nonce, code)))[:40])
	receipt, err := v.ExecuteCode(code, ExecutionEnv{Contract: address, Caller: deployer, GasLimit: gasLimit})
	if err != nil {
		return receipt, err
	}
	if !receipt.Success {
		return receipt, fmt.Errorf(receipt.Error)
	}
	v.state.PutContract(Contract{Address: address, Code: code})
	receipt.ContractAddress = address
	return receipt, nil
}

func (v *VM) Call(address Address, caller string, gasLimit uint64) (Receipt, error) {
	contract, ok := v.state.Contract(address)
	if !ok {
		return Receipt{Success: false, Error: "contract not found"}, fmt.Errorf("contract not found")
	}
	return v.ExecuteCode(contract.Code, ExecutionEnv{Contract: address, Caller: caller, GasLimit: gasLimit})
}

func (v *VM) ExecuteCode(code []byte, env ExecutionEnv) (Receipt, error) {
	var pc int
	var gasUsed uint64
	var stack []uint64
	var logs []Log
	staged := make(map[uint64]uint64)

	charge := func(op OpCode) error {
		gasUsed += gasTable[op]
		if gasUsed > env.GasLimit {
			return fmt.Errorf("out of gas")
		}
		return nil
	}
	pop := func() (uint64, error) {
		if len(stack) == 0 {
			return 0, fmt.Errorf("stack underflow")
		}
		value := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		return value, nil
	}

	for pc < len(code) {
		op := OpCode(code[pc])
		pc++
		if err := charge(op); err != nil {
			return Receipt{GasUsed: gasUsed, Success: false, Error: err.Error(), Logs: logs}, err
		}
		switch op {
		case OpStop:
			for key, value := range staged {
				v.state.Store(env.Contract, key, value)
			}
			return Receipt{GasUsed: gasUsed, Success: true, Logs: logs}, nil
		case OpPush:
			if pc+8 > len(code) {
				return Receipt{GasUsed: gasUsed, Success: false, Error: "truncated push", Logs: logs}, fmt.Errorf("truncated push")
			}
			stack = append(stack, binary.BigEndian.Uint64(code[pc:pc+8]))
			pc += 8
		case OpAdd:
			a, err := pop()
			if err != nil {
				return Receipt{GasUsed: gasUsed, Success: false, Error: err.Error(), Logs: logs}, err
			}
			b, err := pop()
			if err != nil {
				return Receipt{GasUsed: gasUsed, Success: false, Error: err.Error(), Logs: logs}, err
			}
			stack = append(stack, a+b)
		case OpSub:
			a, err := pop()
			if err != nil {
				return Receipt{GasUsed: gasUsed, Success: false, Error: err.Error(), Logs: logs}, err
			}
			b, err := pop()
			if err != nil {
				return Receipt{GasUsed: gasUsed, Success: false, Error: err.Error(), Logs: logs}, err
			}
			stack = append(stack, b-a)
		case OpStore:
			key, err := pop()
			if err != nil {
				return Receipt{GasUsed: gasUsed, Success: false, Error: err.Error(), Logs: logs}, err
			}
			value, err := pop()
			if err != nil {
				return Receipt{GasUsed: gasUsed, Success: false, Error: err.Error(), Logs: logs}, err
			}
			staged[key] = value
		case OpLoad:
			key, err := pop()
			if err != nil {
				return Receipt{GasUsed: gasUsed, Success: false, Error: err.Error(), Logs: logs}, err
			}
			if value, ok := staged[key]; ok {
				stack = append(stack, value)
			} else {
				stack = append(stack, v.state.Load(env.Contract, key))
			}
		case OpLog:
			value, err := pop()
			if err != nil {
				return Receipt{GasUsed: gasUsed, Success: false, Error: err.Error(), Logs: logs}, err
			}
			logs = append(logs, Log{Contract: env.Contract, Value: value})
		case OpReturn:
			value, err := pop()
			if err != nil {
				return Receipt{GasUsed: gasUsed, Success: false, Error: err.Error(), Logs: logs}, err
			}
			for key, stagedValue := range staged {
				v.state.Store(env.Contract, key, stagedValue)
			}
			return Receipt{GasUsed: gasUsed, Success: true, ReturnValue: value, Logs: logs}, nil
		case OpRevert:
			return Receipt{GasUsed: gasUsed, Success: false, Error: "execution reverted", Logs: logs}, fmt.Errorf("execution reverted")
		default:
			return Receipt{GasUsed: gasUsed, Success: false, Error: "invalid opcode", Logs: logs}, fmt.Errorf("invalid opcode 0x%x", byte(op))
		}
	}
	return Receipt{GasUsed: gasUsed, Success: true, Logs: logs}, nil
}
