package governance

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type ProposalStatus string

const (
	StatusDraft     ProposalStatus = "draft"
	StatusSubmitted ProposalStatus = "submitted"
	StatusActive    ProposalStatus = "active"
	StatusPassed    ProposalStatus = "passed"
	StatusRejected  ProposalStatus = "rejected"
	StatusVetoed    ProposalStatus = "vetoed"
	StatusQueued    ProposalStatus = "queued"
	StatusExecuted  ProposalStatus = "executed"
	StatusExpired   ProposalStatus = "expired"
	StatusCanceled  ProposalStatus = "canceled"
)

type QueueState string

const (
	QueueStateQueued          QueueState = "queued"
	QueueStateTimelockPending QueueState = "timelock_pending"
	QueueStateExecutable      QueueState = "executable"
	QueueStateExecuted        QueueState = "executed"
	QueueStateBlocked         QueueState = "blocked"
	QueueStateFailed          QueueState = "failed"
)

type VoteChoice string

const (
	VoteYes     VoteChoice = "yes"
	VoteNo      VoteChoice = "no"
	VoteAbstain VoteChoice = "abstain"
	VoteVeto    VoteChoice = "veto"
)

type Proposal struct {
	ProposalID         string         `json:"proposal_id"`
	Proposer           string         `json:"proposer"`
	Title              string         `json:"title"`
	Description        string         `json:"description"`
	Type               string         `json:"type"`
	CreatedAt          string         `json:"created_at"`
	VotingStart        string         `json:"voting_start"`
	VotingEnd          string         `json:"voting_end"`
	Status             ProposalStatus `json:"status"`
	QuorumRequired     string         `json:"quorum_required"`
	ThresholdRequired  string         `json:"threshold_required"`
	ExecutionState     string         `json:"execution_state"`
	TimelockState      string         `json:"timelock_state"`
	TreasuryActionID   string         `json:"treasury_action_id,omitempty"`
	ValidatorPolicy    string         `json:"validator_policy,omitempty"`
	ExecutionReference string         `json:"execution_reference,omitempty"`
	AuditTrail         []AuditEntry   `json:"audit_trail,omitempty"`
}

type VoteRecord struct {
	ProposalID         string     `json:"proposal_id"`
	Voter              string     `json:"voter"`
	VoterRole          string     `json:"voter_role"`
	DelegationPosture  string     `json:"delegation_posture"`
	VotingPower        string     `json:"voting_power"`
	Vote               VoteChoice `json:"vote"`
	Timestamp          string     `json:"timestamp"`
	SignatureReference string     `json:"signature_reference,omitempty"`
}

type TreasuryAction struct {
	ActionID          string `json:"action_id"`
	ProposalID        string `json:"proposal_id"`
	Recipient         string `json:"recipient"`
	RecipientValid    bool   `json:"recipient_valid"`
	AllocationReason  string `json:"allocation_reason"`
	SpendingCap       string `json:"spending_cap"`
	RequestedAmount   string `json:"requested_amount"`
	ExecutionDelay    string `json:"execution_delay"`
	OperatorApproval  string `json:"operator_approval"`
	EmergencyBlock    string `json:"emergency_block"`
	AuditState        string `json:"audit_state"`
	ExecutionState    string `json:"execution_state"`
	QueueState        string `json:"queue_state"`
	QueueBlockedReason string `json:"queue_blocked_reason,omitempty"`
}

type AuditEntry struct {
	ProposalID string `json:"proposal_id,omitempty"`
	Timestamp  string `json:"timestamp"`
	Action     string `json:"action"`
	Actor      string `json:"actor"`
	Summary    string `json:"summary"`
	Severity   string `json:"severity"`
	Visibility string `json:"visibility"`
}

type Snapshot struct {
	Proposals       []Proposal       `json:"proposals,omitempty"`
	Votes           []VoteRecord     `json:"votes,omitempty"`
	TreasuryActions []TreasuryAction `json:"treasury_actions,omitempty"`
	AuditTrail      []AuditEntry     `json:"audit_trail,omitempty"`
}

type ProposalView struct {
	Proposal
	Yes                int
	No                 int
	Abstain            int
	Veto               int
	ParticipationRatio string
	QuorumReached      string
	VetoThreshold      string
	RemainingTime      string
	PowerState         string
	ValidatorVotes     int
	CommunityVotes     int
	QueueState         string
	QueueBlockedReason string
}

type TreasurySurface struct {
	Wallet             string
	GovernanceFunds    string
	EcosystemReserve   string
	ValidatorReserve   string
	GrantPool          string
	OperationalReserve string
	TreasuryHistory    string
	ActionCount        int
	SafetyNotes        []string
}

type ValidatorParticipation struct {
	Address              string
	ParticipationCount   int
	VotesCast            int
	YesVotes             int
	NoVotes              int
	AbstainVotes         int
	VetoVotes            int
	GovernanceReliability string
	GovernanceTrust       string
	LastVoteAt            string
}

type RuntimeAuditItem struct {
	Risk          string
	Severity      string
	Blocker       string
	RuntimeImpact string
	RequiredFix   string
}

type RuntimeSummary struct {
	Counts            map[string]int
	QueueCounts       map[string]int
	PowerState        string
	GovernanceHealth  string
	ValidatorActivity string
	TreasurySafety    string
}

type Manager struct {
	mu              sync.RWMutex
	proposals       map[string]Proposal
	order           []string
	votes           []VoteRecord
	treasuryActions []TreasuryAction
	auditTrail      []AuditEntry
}

func NewManager() *Manager {
	m := &Manager{}
	m.seed()
	return m
}

func (m *Manager) LoadSnapshot(snapshot Snapshot) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(snapshot.Proposals) == 0 && len(snapshot.AuditTrail) == 0 && len(snapshot.TreasuryActions) == 0 {
		m.seedLocked()
		return nil
	}
	m.proposals = make(map[string]Proposal, len(snapshot.Proposals))
	m.order = make([]string, 0, len(snapshot.Proposals))
	for _, proposal := range snapshot.Proposals {
		m.proposals[proposal.ProposalID] = proposal
		m.order = append(m.order, proposal.ProposalID)
	}
	m.votes = append([]VoteRecord(nil), snapshot.Votes...)
	m.treasuryActions = append([]TreasuryAction(nil), snapshot.TreasuryActions...)
	m.auditTrail = append([]AuditEntry(nil), snapshot.AuditTrail...)
	return nil
}

func (m *Manager) Snapshot() Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	proposals := make([]Proposal, 0, len(m.order))
	for _, id := range m.order {
		proposals = append(proposals, m.proposals[id])
	}
	return Snapshot{
		Proposals:       proposals,
		Votes:           append([]VoteRecord(nil), m.votes...),
		TreasuryActions: append([]TreasuryAction(nil), m.treasuryActions...),
		AuditTrail:      append([]AuditEntry(nil), m.auditTrail...),
	}
}

func (m *Manager) Proposals() []ProposalView {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]ProposalView, 0, len(m.order))
	for _, id := range m.order {
		out = append(out, m.buildProposalViewLocked(m.proposals[id]))
	}
	return out
}

func (m *Manager) Proposal(id string) (ProposalView, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, proposalID := range m.order {
		if strings.EqualFold(proposalID, id) {
			return m.buildProposalViewLocked(m.proposals[proposalID]), true
		}
	}
	return ProposalView{}, false
}

func (m *Manager) PublicAuditTrail() []AuditEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]AuditEntry, 0, len(m.auditTrail))
	for _, item := range m.auditTrail {
		if item.Visibility == "public" {
			out = append(out, item)
		}
	}
	return out
}

func (m *Manager) TreasuryActions() []TreasuryAction {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]TreasuryAction(nil), m.treasuryActions...)
}

func (m *Manager) TreasurySurface() TreasurySurface {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return TreasurySurface{
		Wallet:             "not published in current devnet runtime",
		GovernanceFunds:    "movement blocked until proposal power, quorum and queue execution are live",
		EcosystemReserve:   "allocation architecture visible, fund movement not active",
		ValidatorReserve:   "reserved in monetary architecture, not treasury-executable",
		GrantPool:          "grant pool requires queued and delayed governance execution",
		OperationalReserve: "operational reserve requires treasury recipient validation",
		TreasuryHistory:    fmt.Sprintf("%d treasury action previews recorded", len(m.treasuryActions)),
		ActionCount:        len(m.treasuryActions),
		SafetyNotes: []string{
			"Treasury balance is intentionally not fabricated before a live treasury wallet and accounting runtime exist.",
			"Every treasury action remains blocked behind recipient validation, spending cap review and execution delay.",
			"No action is marked executed unless the governance queue explicitly reaches executable and then executed state.",
		},
	}
}

func (m *Manager) ValidatorParticipation(address string) ValidatorParticipation {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := ValidatorParticipation{
		Address:               address,
		GovernanceReliability: "no live validator votes recorded yet",
		GovernanceTrust:       "operator-managed governance bootstrap",
		LastVoteAt:            "not recorded",
	}
	proposalsSeen := map[string]struct{}{}
	for _, vote := range m.votes {
		if !strings.EqualFold(vote.Voter, address) {
			continue
		}
		result.VotesCast++
		proposalsSeen[vote.ProposalID] = struct{}{}
		result.LastVoteAt = vote.Timestamp
		switch vote.Vote {
		case VoteYes:
			result.YesVotes++
		case VoteNo:
			result.NoVotes++
		case VoteAbstain:
			result.AbstainVotes++
		case VoteVeto:
			result.VetoVotes++
		}
	}
	result.ParticipationCount = len(proposalsSeen)
	if result.VotesCast > 0 {
		result.GovernanceReliability = "observed vote history available"
		result.GovernanceTrust = "validator governance record available"
	}
	return result
}

func (m *Manager) Summary() RuntimeSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()
	counts := map[string]int{
		"draft": 0, "submitted": 0, "active": 0, "passed": 0, "rejected": 0,
		"vetoed": 0, "queued": 0, "executed": 0, "expired": 0, "canceled": 0,
	}
	queueCounts := map[string]int{
		"queued": 0, "timelock_pending": 0, "executable": 0, "executed": 0, "blocked": 0, "failed": 0,
	}
	for _, id := range m.order {
		view := m.buildProposalViewLocked(m.proposals[id])
		counts[string(view.Status)]++
		if view.QueueState != "" {
			queueCounts[view.QueueState]++
		}
	}
	validatorVotes := 0
	for _, vote := range m.votes {
		if vote.VoterRole == "validator" {
			validatorVotes++
		}
	}
	health := "governance runtime active with blocked voting power"
	if len(m.votes) > 0 {
		health = "governance runtime active with recorded vote history"
	}
	return RuntimeSummary{
		Counts:            counts,
		QueueCounts:       queueCounts,
		PowerState:        "voting power runtime not published; quorum remains blocked",
		GovernanceHealth:  health,
		ValidatorActivity: fmt.Sprintf("%d validator vote records", validatorVotes),
		TreasurySafety:    fmt.Sprintf("%d treasury action previews blocked behind queue safety", len(m.treasuryActions)),
	}
}

func RuntimeAudit() []RuntimeAuditItem {
	return []RuntimeAuditItem{
		{
			Risk:          "proposal storage was preview-only",
			Severity:      "high",
			Blocker:       "no persisted proposal lifecycle",
			RuntimeImpact: "proposal pages could show architecture records without real state transitions",
			RequiredFix:   "persist proposal, queue and audit state in chain snapshot",
		},
		{
			Risk:          "voting storage absent",
			Severity:      "critical",
			Blocker:       "no vote record registry",
			RuntimeImpact: "validator participation and vote auditability could not be derived truthfully",
			RequiredFix:   "introduce immutable vote records with duplicate-vote guard posture",
		},
		{
			Risk:          "quorum logic unavailable",
			Severity:      "high",
			Blocker:       "voting power not published",
			RuntimeImpact: "proposal state could not safely advance to passed or rejected from live power data",
			RequiredFix:   "expose voting-power runtime before enabling live quorum settlement",
		},
		{
			Risk:          "treasury action execution unsafe",
			Severity:      "critical",
			Blocker:       "recipient validation and delayed queue not enforced",
			RuntimeImpact: "treasury governance could misrepresent safety or execution readiness",
			RequiredFix:   "attach treasury action previews to queue states and delay posture",
		},
		{
			Risk:          "execution queue missing",
			Severity:      "high",
			Blocker:       "no timelock or blocked-state runtime",
			RuntimeImpact: "passed proposals could appear directly executable",
			RequiredFix:   "derive queue state with timelock_pending, executable and blocked posture",
		},
		{
			Risk:          "fake proposal or vote perception",
			Severity:      "high",
			Blocker:       "static helper data in public explorer",
			RuntimeImpact: "operators could mistake preview content for live governance execution",
			RequiredFix:   "render runtime-backed state and clearly block unpublished power or movement",
		},
	}
}

func (m *Manager) seed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seedLocked()
}

func (m *Manager) seedLocked() {
	now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	m.order = []string{"GOV-101", "GOV-102", "GOV-103", "GOV-104"}
	m.proposals = map[string]Proposal{
		"GOV-101": {
			ProposalID:        "GOV-101",
			Proposer:          "PhoenixChain Operator Council",
			Title:             "Treasury Allocation Guardrails for Ecosystem Grant Queue",
			Description:       "Defines recipient validation, allocation reason disclosure and spending-cap review before any treasury proposal can ever move into executable state.",
			Type:              "treasury allocation",
			CreatedAt:         now.Format(time.RFC3339),
			VotingStart:       now.Add(2 * time.Hour).Format(time.RFC3339),
			VotingEnd:         now.Add(50 * time.Hour).Format(time.RFC3339),
			Status:            StatusSubmitted,
			QuorumRequired:    "blocked until validator/community voting power is published",
			ThresholdRequired: "yes threshold blocked until live voting power exists",
			ExecutionState:    "treasury preview blocked",
			TimelockState:     "execution delay required before any treasury movement",
			TreasuryActionID:  "TREASURY-001",
			AuditTrail: []AuditEntry{
				{ProposalID: "GOV-101", Timestamp: now.Format(time.RFC3339), Action: "proposal_created", Actor: "PhoenixChain Operator Council", Summary: "Treasury guardrail proposal registered in governance runtime.", Severity: "info", Visibility: "public"},
				{ProposalID: "GOV-101", Timestamp: now.Add(30 * time.Minute).Format(time.RFC3339), Action: "treasury_action_previewed", Actor: "PhoenixChain Treasury Safety Layer", Summary: "Recipient validation and spending-cap preview attached before any queue eligibility.", Severity: "info", Visibility: "public"},
			},
		},
		"GOV-102": {
			ProposalID:        "GOV-102",
			Proposer:          "PhoenixChain Validator Working Group",
			Title:             "Validator Governance Participation Policy",
			Description:       "Introduces validator governance participation records, abstain and veto visibility and governance reliability posture before live weighted voting is enabled.",
			Type:              "validator policy",
			CreatedAt:         now.Add(15 * time.Minute).Format(time.RFC3339),
			VotingStart:       now.Add(3 * time.Hour).Format(time.RFC3339),
			VotingEnd:         now.Add(51 * time.Hour).Format(time.RFC3339),
			Status:            StatusActive,
			QuorumRequired:    "validator quorum blocked until voting power is published",
			ThresholdRequired: "validator threshold blocked until signed live votes exist",
			ExecutionState:    "policy runtime active, settlement blocked",
			TimelockState:     "no queue execution until quorum can be derived",
			ValidatorPolicy:   "active governance participation surface",
			AuditTrail: []AuditEntry{
				{ProposalID: "GOV-102", Timestamp: now.Add(15 * time.Minute).Format(time.RFC3339), Action: "proposal_created", Actor: "PhoenixChain Validator Working Group", Summary: "Validator participation policy registered.", Severity: "info", Visibility: "public"},
				{ProposalID: "GOV-102", Timestamp: now.Add(3 * time.Hour).Format(time.RFC3339), Action: "proposal_activated", Actor: "PhoenixChain Governance Runtime", Summary: "Proposal entered active state, awaiting live voting-power publication.", Severity: "info", Visibility: "public"},
			},
		},
		"GOV-103": {
			ProposalID:        "GOV-103",
			Proposer:          "PhoenixChain Runtime Maintainers",
			Title:             "Runtime Upgrade Timelock and Queue Discipline",
			Description:       "Creates execution-queue posture, timelock pending state and manual recovery notes for future runtime upgrade proposals.",
			Type:              "runtime upgrade",
			CreatedAt:         now.Add(30 * time.Minute).Format(time.RFC3339),
			VotingStart:       "not scheduled",
			VotingEnd:         "not scheduled",
			Status:            StatusDraft,
			QuorumRequired:    "blocked until draft is submitted and power is published",
			ThresholdRequired: "blocked until runtime upgrade voting weights exist",
			ExecutionState:    "draft only",
			TimelockState:     "timelock model attached, not counting down",
			AuditTrail: []AuditEntry{
				{ProposalID: "GOV-103", Timestamp: now.Add(30 * time.Minute).Format(time.RFC3339), Action: "proposal_created", Actor: "PhoenixChain Runtime Maintainers", Summary: "Runtime upgrade governance record created in draft state.", Severity: "info", Visibility: "public"},
			},
		},
		"GOV-104": {
			ProposalID:         "GOV-104",
			Proposer:           "PhoenixChain Governance Ops",
			Title:              "PHX-20 Ecosystem Decision Archive Placeholder",
			Description:        "Reserved lifecycle slot for future PHX-20 ecosystem decisions; canceled to avoid fake vote history before the voting runtime is published.",
			Type:               "PHX-20 ecosystem decision",
			CreatedAt:          now.Add(45 * time.Minute).Format(time.RFC3339),
			VotingStart:        "not scheduled",
			VotingEnd:          "not scheduled",
			Status:             StatusCanceled,
			QuorumRequired:     "not evaluated",
			ThresholdRequired:  "not evaluated",
			ExecutionState:     "canceled before activation",
			TimelockState:      "not applicable",
			ExecutionReference: "canceled before queue",
			AuditTrail: []AuditEntry{
				{ProposalID: "GOV-104", Timestamp: now.Add(45 * time.Minute).Format(time.RFC3339), Action: "proposal_created", Actor: "PhoenixChain Governance Ops", Summary: "Placeholder proposal created to reserve archive namespace.", Severity: "info", Visibility: "public"},
				{ProposalID: "GOV-104", Timestamp: now.Add(50 * time.Minute).Format(time.RFC3339), Action: "proposal_canceled", Actor: "PhoenixChain Governance Ops", Summary: "Proposal canceled to avoid presenting inactive PHX-20 governance as live voting.", Severity: "warn", Visibility: "public"},
			},
		},
	}
	m.votes = nil
	m.treasuryActions = []TreasuryAction{
		{
			ActionID:           "TREASURY-001",
			ProposalID:         "GOV-101",
			Recipient:          "phx1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			RecipientValid:     false,
			AllocationReason:   "ecosystem grant queue bootstrap policy preview",
			SpendingCap:        "cap review required before activation",
			RequestedAmount:    "not executable",
			ExecutionDelay:     "48h timelock required after queue eligibility",
			OperatorApproval:   "operator override not granted",
			EmergencyBlock:     "standby",
			AuditState:         "recorded",
			ExecutionState:     "blocked",
			QueueState:         string(QueueStateBlocked),
			QueueBlockedReason: "recipient validation failed and voting power runtime unavailable",
		},
	}
	m.auditTrail = []AuditEntry{
		{Timestamp: now.Format(time.RFC3339), Action: "governance_runtime_seeded", Actor: "PhoenixChain Governance Runtime", Summary: "Governance lifecycle, queue and treasury safety runtime seeded from private operator bootstrap.", Severity: "info", Visibility: "public"},
		{Timestamp: now.Add(20 * time.Minute).Format(time.RFC3339), Action: "voting_power_blocked", Actor: "PhoenixChain Governance Runtime", Summary: "Voting power publication is still blocked, so quorum remains intentionally unresolved.", Severity: "warn", Visibility: "public"},
		{Timestamp: now.Add(40 * time.Minute).Format(time.RFC3339), Action: "private_operator_note", Actor: "PhoenixChain Governance Admin", Summary: "Private governance operations note retained for admin-only incident follow-up.", Severity: "info", Visibility: "private"},
	}
}

func (m *Manager) buildProposalViewLocked(proposal Proposal) ProposalView {
	view := ProposalView{Proposal: proposal}
	hasPublishedPower := false
	validatorVotes := 0
	communityVotes := 0
	for _, vote := range m.votes {
		if !strings.EqualFold(vote.ProposalID, proposal.ProposalID) {
			continue
		}
		if vote.VoterRole == "validator" {
			validatorVotes++
		} else {
			communityVotes++
		}
		if vote.VotingPower != "" && vote.VotingPower != "not published" {
			hasPublishedPower = true
		}
		switch vote.Vote {
		case VoteYes:
			view.Yes++
		case VoteNo:
			view.No++
		case VoteAbstain:
			view.Abstain++
		case VoteVeto:
			view.Veto++
		}
	}
	view.ValidatorVotes = validatorVotes
	view.CommunityVotes = communityVotes
	if hasPublishedPower {
		total := view.Yes + view.No + view.Abstain + view.Veto
		view.ParticipationRatio = fmt.Sprintf("%d recorded votes", total)
		view.QuorumReached = "pending weighted settlement"
		view.PowerState = "partial voting-power publication"
	} else {
		view.ParticipationRatio = "blocked until voting power is published"
		view.QuorumReached = "not computable"
		view.PowerState = "voting power runtime unavailable"
	}
	view.VetoThreshold = "blocked until live voting power exists"
	view.RemainingTime = remainingTime(proposal)
	queueState, blockedReason := m.queueStateLocked(proposal)
	view.QueueState = queueState
	view.QueueBlockedReason = blockedReason
	return view
}

func (m *Manager) queueStateLocked(proposal Proposal) (string, string) {
	for _, action := range m.treasuryActions {
		if action.ProposalID == proposal.ProposalID {
			return action.QueueState, action.QueueBlockedReason
		}
	}
	switch proposal.Status {
	case StatusPassed:
		return string(QueueStateQueued), ""
	case StatusQueued:
		return string(QueueStateTimelockPending), ""
	case StatusExecuted:
		return string(QueueStateExecuted), ""
	default:
		return string(QueueStateBlocked), "proposal not eligible for execution queue"
	}
}

func remainingTime(proposal Proposal) string {
	if proposal.VotingEnd == "" || proposal.VotingEnd == "not scheduled" {
		return "not scheduled"
	}
	end, err := time.Parse(time.RFC3339, proposal.VotingEnd)
	if err != nil {
		return "not derivable"
	}
	now := time.Now().UTC()
	if proposal.Status == StatusCanceled || proposal.Status == StatusExpired || proposal.Status == StatusExecuted {
		return "closed"
	}
	if end.Before(now) {
		return "window elapsed"
	}
	return end.Sub(now).Round(time.Minute).String()
}

func SortAuditTrail(items []AuditEntry) []AuditEntry {
	out := append([]AuditEntry(nil), items...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Timestamp > out[j].Timestamp
	})
	return out
}
