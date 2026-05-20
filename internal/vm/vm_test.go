package vm_test

import (
	"testing"

	"github.com/phoenixchain/phoenixchain/internal/vm"
)

func TestDeployTokenContractStoresSupplyAndLogs(t *testing.T) {
	state := vm.NewState()
	machine := vm.New(state)

	receipt, err := machine.Deploy(vm.BasicTokenExample(1000), "phxdev", 0, 100)
	if err != nil {
		t.Fatal(err)
	}
	if !receipt.Success {
		t.Fatalf("deployment failed: %s", receipt.Error)
	}
	if got := state.Load(receipt.ContractAddress, 0); got != 1000 {
		t.Fatalf("supply slot = %d, want 1000", got)
	}
	if len(receipt.Logs) != 1 || receipt.Logs[0].Value != 1000 {
		t.Fatalf("unexpected logs: %+v", receipt.Logs)
	}
}

func TestCallReturnsStoredValue(t *testing.T) {
	state := vm.NewState()
	machine := vm.New(state)
	code := vm.Program(vm.Push(77), vm.Push(1), vm.Op(vm.OpStore), vm.Push(1), vm.Op(vm.OpLoad), vm.Op(vm.OpReturn))

	receipt, err := machine.Deploy(code, "phxdev", 1, 100)
	if err != nil {
		t.Fatal(err)
	}
	callReceipt, err := machine.Call(receipt.ContractAddress, "phxcaller", 100)
	if err != nil {
		t.Fatal(err)
	}
	if callReceipt.ReturnValue != 77 {
		t.Fatalf("return = %d, want 77", callReceipt.ReturnValue)
	}
}

func TestOutOfGasRevertsState(t *testing.T) {
	state := vm.NewState()
	machine := vm.New(state)
	code := vm.Program(vm.Push(55), vm.Push(0), vm.Op(vm.OpStore), vm.Op(vm.OpStop))

	receipt, err := machine.ExecuteCode(code, vm.ExecutionEnv{Contract: "pc1", Caller: "phxdev", GasLimit: 5})
	if err == nil {
		t.Fatal("expected out of gas")
	}
	if receipt.Success {
		t.Fatal("expected failed receipt")
	}
	if got := state.Load("pc1", 0); got != 0 {
		t.Fatalf("state changed after failed execution: %d", got)
	}
}

func TestRevertDoesNotCommitStorage(t *testing.T) {
	state := vm.NewState()
	machine := vm.New(state)
	code := vm.Program(vm.Push(55), vm.Push(0), vm.Op(vm.OpStore), vm.Op(vm.OpRevert))

	_, err := machine.ExecuteCode(code, vm.ExecutionEnv{Contract: "pc1", Caller: "phxdev", GasLimit: 100})
	if err == nil {
		t.Fatal("expected revert")
	}
	if got := state.Load("pc1", 0); got != 0 {
		t.Fatalf("state changed after revert: %d", got)
	}
}
