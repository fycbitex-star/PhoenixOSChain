package economics

import "fmt"

type Treasury struct {
	Address string `json:"address"`
	Balance uint64 `json:"balance"`
}

type Core struct {
	params     Params
	treasury   Treasury
	validators map[string]Validator
}

func NewCore(params Params, treasuryAddress string) *Core {
	return &Core{
		params:     params,
		treasury:   Treasury{Address: treasuryAddress},
		validators: make(map[string]Validator),
	}
}

func (c *Core) Validator(address string) (Validator, bool) {
	validator, ok := c.validators[address]
	return validator, ok
}

func (c *Core) Validators() []Validator {
	out := make([]Validator, 0, len(c.validators))
	for _, validator := range c.validators {
		out = append(out, validator)
	}
	return out
}

func (c *Core) Treasury() Treasury {
	return c.treasury
}

func (c *Core) DistributeFees(proposer string, fees uint64) error {
	validator, ok := c.validators[proposer]
	if !ok {
		return fmt.Errorf("validator not found")
	}
	if validator.Status != ValidatorBonded {
		return fmt.Errorf("validator is not bonded")
	}
	validatorShare := fees * c.params.ValidatorFeeShare / 100
	treasuryShare := fees - validatorShare
	validator.RewardsAccrued += validatorShare + c.params.BlockReward
	c.treasury.Balance += treasuryShare
	c.validators[proposer] = validator
	return nil
}
