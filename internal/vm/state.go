package vm

type State struct {
	contracts map[Address]Contract
	storage   map[Address]map[uint64]uint64
}

func NewState() *State {
	return &State{
		contracts: make(map[Address]Contract),
		storage:   make(map[Address]map[uint64]uint64),
	}
}

func (s *State) PutContract(contract Contract) {
	s.contracts[contract.Address] = contract
	if _, ok := s.storage[contract.Address]; !ok {
		s.storage[contract.Address] = make(map[uint64]uint64)
	}
}

func (s *State) Contract(address Address) (Contract, bool) {
	contract, ok := s.contracts[address]
	return contract, ok
}

func (s *State) Store(address Address, key, value uint64) {
	if _, ok := s.storage[address]; !ok {
		s.storage[address] = make(map[uint64]uint64)
	}
	s.storage[address][key] = value
}

func (s *State) Load(address Address, key uint64) uint64 {
	if s.storage[address] == nil {
		return 0
	}
	return s.storage[address][key]
}

func (s *State) SnapshotStorage(address Address) map[uint64]uint64 {
	copy := make(map[uint64]uint64)
	for key, value := range s.storage[address] {
		copy[key] = value
	}
	return copy
}
