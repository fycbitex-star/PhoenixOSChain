package economics

type Params struct {
	MinimumBond        uint64
	BlockReward        uint64
	ValidatorFeeShare  uint64
	TreasuryFeeShare   uint64
	UnbondingPeriodSec uint64
}

func DefaultParams() Params {
	return Params{
		MinimumBond:        1000,
		BlockReward:        10,
		ValidatorFeeShare:  80,
		TreasuryFeeShare:   20,
		UnbondingPeriodSec: 7 * 24 * 60 * 60,
	}
}
