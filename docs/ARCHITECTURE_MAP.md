# Architecture Map

```mermaid
flowchart LR
    subgraph PhoenixOS["PhoenixOS Operating Layer"]
        O1["Execution + OMS"]
        O2["Wallet OS"]
        O3["AI + Replay"]
        O4["Operator Control"]
    end

    subgraph PhoenixChain["PhoenixChain Network Layer"]
        C1["Chain Core"]
        C2["Validator Runtime"]
        C3["Governance Runtime"]
        C4["Treasury Safety"]
        C5["PHX / PHX-20"]
        C6["Explorer + RPC"]
    end

    O1 --> C1
    O1 --> C6
    O2 --> C5
    O2 --> C3
    O3 --> C3
    O4 --> C2
    O4 --> C4
```

## Reading

The topology is straightforward:

- PhoenixOS is the operator environment
- PhoenixChain is the sovereign trust environment

PhoenixOS can coordinate.
PhoenixChain must eventually legitimize.
