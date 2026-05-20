package economics

import "fmt"

type ValidatorStatus string

const (
	ValidatorBonded    ValidatorStatus = "bonded"
	ValidatorUnbonding ValidatorStatus = "unbonding"
	ValidatorSlashed   ValidatorStatus = "slashed"
)

type Validator struct {
	Address          string          `json:"address"`
	Identity         string          `json:"identity"`
	Metadata         string          `json:"metadata"`
	Bond             uint64          `json:"bond"`
	Status           ValidatorStatus `json:"status"`
	UnbondStartUnix  int64           `json:"unbond_start_unix,omitempty"`
	RewardsAccrued   uint64          `json:"rewards_accrued"`
	SlashingEvidence []string        `json:"slashing_evidence,omitempty"`
}

func (c *Core) RegisterValidator(address, identity, metadata string, bond uint64) error {
	if address == "" {
		return fmt.Errorf("validator address is required")
	}
	if bond < c.params.MinimumBond {
		return fmt.Errorf("bond below minimum")
	}
	if _, exists := c.validators[address]; exists {
		return fmt.Errorf("validator already registered")
	}
	c.validators[address] = Validator{
		Address:  address,
		Identity: identity,
		Metadata: metadata,
		Bond:     bond,
		Status:   ValidatorBonded,
	}
	return nil
}

func (c *Core) BeginUnbond(address string, nowUnix int64) error {
	validator, ok := c.validators[address]
	if !ok {
		return fmt.Errorf("validator not found")
	}
	if validator.Status != ValidatorBonded {
		return fmt.Errorf("validator is not bonded")
	}
	validator.Status = ValidatorUnbonding
	validator.UnbondStartUnix = nowUnix
	c.validators[address] = validator
	return nil
}

func (c *Core) SlashPlaceholder(address, evidence string) error {
	validator, ok := c.validators[address]
	if !ok {
		return fmt.Errorf("validator not found")
	}
	validator.Status = ValidatorSlashed
	validator.SlashingEvidence = append(validator.SlashingEvidence, evidence)
	c.validators[address] = validator
	return nil
}
