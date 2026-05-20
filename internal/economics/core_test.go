package economics_test

import (
	"testing"

	"github.com/phoenixchain/phoenixchain/internal/economics"
)

func TestValidatorRegistrationAndRewards(t *testing.T) {
	core := economics.NewCore(economics.DefaultParams(), "phxtreasury")
	if err := core.RegisterValidator("phxvalidator", "validator-1", "local devnet", 1000); err != nil {
		t.Fatal(err)
	}
	if err := core.DistributeFees("phxvalidator", 100); err != nil {
		t.Fatal(err)
	}
	validator, ok := core.Validator("phxvalidator")
	if !ok {
		t.Fatal("validator missing")
	}
	if validator.RewardsAccrued != 90 {
		t.Fatalf("rewards = %d, want 90", validator.RewardsAccrued)
	}
	if core.Treasury().Balance != 20 {
		t.Fatalf("treasury = %d, want 20", core.Treasury().Balance)
	}
}

func TestRejectsLowBond(t *testing.T) {
	core := economics.NewCore(economics.DefaultParams(), "phxtreasury")
	if err := core.RegisterValidator("phxvalidator", "validator-1", "", 999); err == nil {
		t.Fatal("expected low bond rejection")
	}
}

func TestUnbondAndSlashPlaceholder(t *testing.T) {
	core := economics.NewCore(economics.DefaultParams(), "phxtreasury")
	if err := core.RegisterValidator("phxvalidator", "validator-1", "", 1000); err != nil {
		t.Fatal(err)
	}
	if err := core.BeginUnbond("phxvalidator", 123); err != nil {
		t.Fatal(err)
	}
	validator, _ := core.Validator("phxvalidator")
	if validator.Status != economics.ValidatorUnbonding {
		t.Fatalf("status = %s, want unbonding", validator.Status)
	}
	if err := core.SlashPlaceholder("phxvalidator", "double-sign placeholder evidence"); err != nil {
		t.Fatal(err)
	}
	validator, _ = core.Validator("phxvalidator")
	if validator.Status != economics.ValidatorSlashed {
		t.Fatalf("status = %s, want slashed", validator.Status)
	}
}
