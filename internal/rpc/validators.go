package rpc

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/phoenixchain/phoenixchain/internal/chain"
	"github.com/phoenixchain/phoenixchain/internal/governance"
)

type validatorPortalRecord struct {
	Rank                    int
	Name                    string
	Address                 string
	Status                  string
	Health                  string
	Identity                string
	GovernanceIdentity      string
	Bio                     string
	Website                 string
	Social                  string
	Region                  string
	Commission              string
	TotalStake              string
	SelfStake               string
	Delegators              string
	RewardVisibility        string
	BlocksProposed          int
	ExpectedProposals       int
	MissedBlocks            int
	UptimeLabel             string
	LatestActivity          string
	RiskStatus              string
	RecentHeights           []uint64
	GovernanceParticipation string
	GovernanceTrust         string
	GovernanceStance        string
	GovernanceVotes         int
	GovernanceLastVote      string
}

type validatorSecuritySummary struct {
	ActiveValidators string
	TotalStaked      string
	StakingRatio     string
	RewardState      string
	Concentration    string
	NetworkHealth    string
}

func buildValidatorRecords(bc *chain.Blockchain) []validatorPortalRecord {
	validators := bc.ValidatorList()
	blocks := bc.Blocks()
	head := bc.Head().Height
	records := make([]validatorPortalRecord, 0, len(validators))

	for index, address := range validators {
		record := validatorPortalRecord{
			Rank:                    index + 1,
			Name:                    fmt.Sprintf("Genesis Validator %d", index+1),
			Address:                 address,
			Status:                  "active",
			Identity:                fmt.Sprintf("phoenixchain-validator-%d", index+1),
			GovernanceIdentity:      "operator-managed devnet validator",
			Bio:                     "Permissioned validator in the current PhoenixChain PoA devnet set.",
			Website:                 "not published",
			Social:                  "not published",
			Region:                  "not published",
			Commission:              "pending staking runtime support",
			TotalStake:              "pending staking runtime support",
			SelfStake:               "permissioned validator bond not exposed",
			Delegators:              "delegation not active",
			RewardVisibility:        "reward accounting not exposed in live devnet portal",
			LatestActivity:          "no proposal observed yet",
			GovernanceParticipation: "no live governance votes",
			GovernanceTrust:         "operator-managed governance bootstrap",
			GovernanceStance:        "not published",
			GovernanceLastVote:      "not recorded",
		}

		for _, block := range blocks {
			if block.Height == 0 {
				continue
			}
			if validatorExpectedProposer(address, block.Height, validators) {
				record.ExpectedProposals++
			}
			if block.ValidatorAddress == address {
				record.BlocksProposed++
				record.RecentHeights = append(record.RecentHeights, block.Height)
			}
		}

		if len(record.RecentHeights) > 0 {
			latest := record.RecentHeights[len(record.RecentHeights)-1]
			record.LatestActivity = fmt.Sprintf("proposed block #%d", latest)
		}
		record.MissedBlocks = record.ExpectedProposals - record.BlocksProposed
		if record.MissedBlocks < 0 {
			record.MissedBlocks = 0
		}
		record.UptimeLabel = proposalUptimeLabel(record.BlocksProposed, record.ExpectedProposals)
		record.Health = classifyValidatorHealth(record, head, len(validators))
		record.RiskStatus = classifyValidatorRisk(record)
		records = append(records, record)
	}

	return records
}

func (s *Server) governanceAwareValidatorRecords() []validatorPortalRecord {
	records := buildValidatorRecords(s.chain)
	for i := range records {
		participation := s.governance.ValidatorParticipation(records[i].Address)
		records[i].GovernanceVotes = participation.VotesCast
		records[i].GovernanceLastVote = participation.LastVoteAt
		records[i].GovernanceParticipation = fmt.Sprintf("%d proposals | %d votes", participation.ParticipationCount, participation.VotesCast)
		records[i].GovernanceTrust = participation.GovernanceTrust
		records[i].GovernanceStance = stanceFromParticipation(participation)
	}
	return records
}

func stanceFromParticipation(participation governance.ValidatorParticipation) string {
	if participation.VotesCast == 0 {
		return "no live stance recorded"
	}
	switch {
	case participation.YesVotes > participation.NoVotes && participation.YesVotes >= participation.VetoVotes:
		return "yes-leaning from recorded votes"
	case participation.VetoVotes > 0:
		return "veto-capable caution observed"
	case participation.NoVotes > participation.YesVotes:
		return "no-leaning from recorded votes"
	default:
		return "mixed or abstain-heavy posture"
	}
}

func filterAndSortValidators(records []validatorPortalRecord, query, sortBy, status string) []validatorPortalRecord {
	query = strings.TrimSpace(strings.ToLower(query))
	status = strings.TrimSpace(strings.ToLower(status))
	filtered := make([]validatorPortalRecord, 0, len(records))
	for _, record := range records {
		if query != "" {
			haystack := strings.ToLower(record.Name + " " + record.Address + " " + record.Identity)
			if !strings.Contains(haystack, query) {
				continue
			}
		}
		if status != "" && status != "all" && strings.ToLower(record.Health) != status && strings.ToLower(record.Status) != status {
			continue
		}
		filtered = append(filtered, record)
	}

	sort.Slice(filtered, func(i, j int) bool {
		switch sortBy {
		case "uptime":
			return filtered[i].BlocksProposed > filtered[j].BlocksProposed
		case "active":
			return filtered[i].Health < filtered[j].Health
		case "top-stake":
			return filtered[i].Rank < filtered[j].Rank
		default:
			return filtered[i].BlocksProposed > filtered[j].BlocksProposed
		}
	})
	return filtered
}

func findValidatorRecord(records []validatorPortalRecord, address string) (validatorPortalRecord, bool) {
	for _, record := range records {
		if strings.EqualFold(record.Address, address) {
			return record, true
		}
	}
	return validatorPortalRecord{}, false
}

func validatorSecurity(records []validatorPortalRecord) validatorSecuritySummary {
	healthy := 0
	for _, record := range records {
		if record.Health == "active" || record.Health == "healthy" {
			healthy++
		}
	}
	concentration := "high concentration risk"
	if len(records) >= 4 {
		concentration = "moderate concentration risk"
	}
	networkHealth := "degraded"
	if healthy == len(records) && healthy > 0 {
		networkHealth = "healthy"
	} else if healthy > 0 {
		networkHealth = "mixed"
	}
	return validatorSecuritySummary{
		ActiveValidators: fmt.Sprintf("%d / %d", healthy, len(records)),
		TotalStaked:      "not published in current PoA devnet",
		StakingRatio:     "staking not active",
		RewardState:      "rewards not active in live validator portal",
		Concentration:    concentration,
		NetworkHealth:    networkHealth,
	}
}

func proposalUptimeLabel(proposed, expected int) string {
	if expected <= 0 {
		return "not enough proposal history"
	}
	return fmt.Sprintf("%.0f%%", (float64(proposed)/float64(expected))*100)
}

func classifyValidatorHealth(record validatorPortalRecord, head uint64, validatorCount int) string {
	if record.ExpectedProposals == 0 {
		return "delayed"
	}
	if record.BlocksProposed == 0 {
		return "offline"
	}
	if record.MissedBlocks == 0 {
		return "active"
	}
	latest := uint64(0)
	if len(record.RecentHeights) > 0 {
		latest = record.RecentHeights[len(record.RecentHeights)-1]
	}
	if latest > 0 && head-latest <= uint64(max(validatorCount*2, 2)) {
		return "delayed"
	}
	if record.MissedBlocks > record.BlocksProposed {
		return "degraded"
	}
	return "unhealthy"
}

func classifyValidatorRisk(record validatorPortalRecord) string {
	switch record.Health {
	case "active":
		return "low current protocol risk"
	case "delayed":
		return "watch proposal continuity"
	case "degraded", "unhealthy":
		return "elevated continuity risk"
	case "offline":
		return "offline validator risk"
	default:
		return "pending runtime support"
	}
}

func validatorExpectedProposer(address string, height uint64, validators []string) bool {
	if len(validators) == 0 {
		return false
	}
	return validators[int(height%uint64(len(validators)))] == address
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (s *Server) handleValidatorDetail(w http.ResponseWriter, r *http.Request) {
	address, ok := normalizeAddress(r.PathValue("address"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid validator address")
		return
	}
	data := s.explorerData("PhoenixOS Chain Validator")
	records := s.governanceAwareValidatorRecords()
	record, found := findValidatorRecord(records, address)
	if !found {
		writeError(w, http.StatusNotFound, "validator not found")
		return
	}
	data["ValidatorRecord"] = record
	data["ValidatorSecurity"] = validatorSecurity(records)
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>{{.ValidatorRecord.Name}}</h1><span class="muted">{{.ValidatorRecord.GovernanceIdentity}}</span></div>
<section class="grid">
  <div class="metric"><div class="label">Validator Address</div><div class="value hash">{{short .ValidatorRecord.Address}}</div></div>
  <div class="metric"><div class="label">Status</div><div class="value">{{.ValidatorRecord.Status}}</div></div>
  <div class="metric"><div class="label">Health</div><div class="value">{{.ValidatorRecord.Health}}</div></div>
  <div class="metric"><div class="label">Risk Status</div><div class="value">{{.ValidatorRecord.RiskStatus}}</div></div>
  <div class="metric"><div class="label">Blocks Proposed</div><div class="value">{{.ValidatorRecord.BlocksProposed}}</div></div>
  <div class="metric"><div class="label">Missed Blocks</div><div class="value">{{.ValidatorRecord.MissedBlocks}}</div></div>
  <div class="metric"><div class="label">Uptime</div><div class="value">{{.ValidatorRecord.UptimeLabel}}</div></div>
  <div class="metric"><div class="label">Latest Activity</div><div class="value">{{.ValidatorRecord.LatestActivity}}</div></div>
</section>
<div class="section-title"><h2>Overview</h2><span class="muted">Governance-ready validator identity</span></div>
<div class="panel">
  <p class="hash">Validator address: {{.ValidatorRecord.Address}}</p>
  <p>Identity: {{.ValidatorRecord.Identity}}</p>
  <p>Bio: {{.ValidatorRecord.Bio}}</p>
  <p>Website: {{.ValidatorRecord.Website}}</p>
  <p>Social: {{.ValidatorRecord.Social}}</p>
  <p>Region: {{.ValidatorRecord.Region}}</p>
  <p>Governance profile: {{.ValidatorRecord.GovernanceIdentity}}</p>
  <p class="muted">This validator identity is surfaced as infrastructure profile metadata. Signed live governance voting is still blocked until voting power and vote submission runtime are published.</p>
</div>
<div class="section-title"><h2>Delegations</h2><span class="muted">Delegation readiness</span></div>
<div class="panel">
  <p>Total delegated: {{.ValidatorRecord.TotalStake}}</p>
  <p>Self bonded: {{.ValidatorRecord.SelfStake}}</p>
  <p>Delegator count: {{.ValidatorRecord.Delegators}}</p>
  <p>Commission: {{.ValidatorRecord.Commission}}</p>
  <p class="muted">Delegation accounting is not active in the current PoA devnet, so this page does not fabricate live stake values.</p>
</div>
<div class="section-title"><h2>Rewards</h2><span class="muted">Reward visibility</span></div>
<div class="panel">
  <p>{{.ValidatorRecord.RewardVisibility}}</p>
  <p>Claim status: pending staking runtime support</p>
  <p>Reward history: pending staking runtime support</p>
</div>
<div class="section-title"><h2>Activity</h2><span class="muted">Proposal continuity</span></div>
<div class="panel">
  <p>Expected proposals from observed chain history: {{.ValidatorRecord.ExpectedProposals}}</p>
  <p>Observed proposals: {{.ValidatorRecord.BlocksProposed}}</p>
  <p>Missed proposal slots: {{.ValidatorRecord.MissedBlocks}}</p>
  <p>Recent proposal heights: {{range $index, $item := .ValidatorRecord.RecentHeights}}{{if $index}}, {{end}}#{{$item}}{{else}}none observed yet{{end}}</p>
</div>
<div class="section-title"><h2>Governance</h2><span class="muted">Validator participation surface</span></div>
<div class="panel">
  <p>Validator governance identity: {{.ValidatorRecord.GovernanceIdentity}}</p>
  <p>Proposal voting: {{.ValidatorRecord.GovernanceParticipation}}</p>
  <p>Votes cast: {{.ValidatorRecord.GovernanceVotes}} | Last vote: {{.ValidatorRecord.GovernanceLastVote}}</p>
  <p>Community trust: {{.ValidatorRecord.GovernanceTrust}}</p>
  <p>Observed stance: {{.ValidatorRecord.GovernanceStance}}</p>
  <p>Infrastructure profile: validator-secured PoA runtime participant</p>
</div>
<div class="section-title"><h2>Analytics</h2><span class="muted">Network security context</span></div>
<div class="panel">
  <p>Active validator summary: {{.ValidatorSecurity.ActiveValidators}}</p>
  <p>Network concentration risk: {{.ValidatorSecurity.Concentration}}</p>
  <p>Network health: {{.ValidatorSecurity.NetworkHealth}}</p>
  <p>Nakamoto-style decentralization: placeholder until richer validator/runtime metrics exist</p>
</div>
{{template "modulePreview" dict "Title" "What remains before live staking" "Body" "Still missing for full staking parity: stake transactions, undelegate flow, redelegation flow, reward claim transactions, validator commission runtime, slashing enforcement and wallet-signed confirmations."}}
{{end}}`, data)
}
