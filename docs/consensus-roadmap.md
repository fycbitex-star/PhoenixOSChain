# PhoenixChain Consensus Roadmap

Current consensus is fixed-validator PoA for devnet.

## Current

- fixed validator list
- validator signatures
- round-robin proposer validation
- duplicate block rejection
- invalid validator rejection

## BFT-Ready Architecture

Future consensus should add:

- proposal
- prevote
- precommit
- quorum checks
- finality certificates
- validator set updates
- evidence handling

## Do Not Implement Yet

HotStuff or Tendermint-style BFT should not be rushed until networking,
storage, and validator operations are stable.
