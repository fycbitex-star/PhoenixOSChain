package rpc

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/phoenixchain/phoenixchain/internal/governance"
)

type governancePortalData struct {
	Summary        governance.RuntimeSummary
	Proposals      []governance.ProposalView
	Treasury       governance.TreasurySurface
	AuditTrail     []governance.AuditEntry
	Audit          []governance.RuntimeAuditItem
	TreasuryAction []governance.TreasuryAction
}

func (s *Server) governanceData() governancePortalData {
	return governancePortalData{
		Summary:        s.governance.Summary(),
		Proposals:      s.governance.Proposals(),
		Treasury:       s.governance.TreasurySurface(),
		AuditTrail:     governance.SortAuditTrail(s.governance.PublicAuditTrail()),
		Audit:          governance.RuntimeAudit(),
		TreasuryAction: s.governance.TreasuryActions(),
	}
}

func findTreasuryAction(actions []governance.TreasuryAction, proposalID string) (governance.TreasuryAction, bool) {
	for _, action := range actions {
		if strings.EqualFold(action.ProposalID, proposalID) {
			return action, true
		}
	}
	return governance.TreasuryAction{}, false
}

func governanceAnalystSummary(proposal governance.ProposalView) string {
	switch {
	case proposal.Type == "treasury allocation":
		return "Treasury impact exists, but movement remains blocked until recipient validation, queue delay and live voting power all clear together."
	case proposal.Status == governance.StatusActive:
		return "Proposal is structurally active, but the runtime still refuses to manufacture quorum or settlement before voting-power publication."
	case proposal.Status == governance.StatusDraft:
		return "Draft proposal records execution and audit intent early, which is healthier than pretending inactive governance is already binding."
	case proposal.Status == governance.StatusCanceled:
		return "Canceled state protects governance truthfulness by refusing to present inactive ecosystem decisions as live participation."
	default:
		return "Governance analyst sees a runtime-backed record, but not enough live voting evidence to summarize validator stance with confidence."
	}
}

func governanceSecurityNotes() []string {
	return []string{
		"Proposal spam remains blocked at the surface level because no public write endpoint is exposed yet.",
		"Duplicate vote prevention posture is active by design because the runtime refuses vote settlement without published voting power and identity binding.",
		"Treasury recipients are validated before a proposal can even appear queue-eligible.",
		"Execution delay and emergency block posture remain visible even while executable treasury movement is still blocked.",
		"Public audit trail is explorer-readable; private operator notes remain excluded from public pages.",
	}
}

func (s *Server) handleGovernanceProposals(w http.ResponseWriter, r *http.Request) {
	data := s.explorerData("PhoenixOS Chain Governance Proposals")
	runtime := s.governanceData()
	data["GovernanceRuntime"] = runtime
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Governance Proposals</h1><span class="muted">Runtime-backed proposal explorer</span></div>
<section class="grid">
  <div class="metric"><div class="label">Draft</div><div class="value">{{index .GovernanceRuntime.Summary.Counts "draft"}}</div></div>
  <div class="metric"><div class="label">Submitted</div><div class="value">{{index .GovernanceRuntime.Summary.Counts "submitted"}}</div></div>
  <div class="metric"><div class="label">Active</div><div class="value">{{index .GovernanceRuntime.Summary.Counts "active"}}</div></div>
  <div class="metric"><div class="label">Queued</div><div class="value">{{index .GovernanceRuntime.Summary.Counts "queued"}}</div></div>
  <div class="metric"><div class="label">Executed</div><div class="value">{{index .GovernanceRuntime.Summary.Counts "executed"}}</div></div>
  <div class="metric"><div class="label">Voting Power</div><div class="value">{{.GovernanceRuntime.Summary.PowerState}}</div></div>
</section>
<div class="panel">
  <p>This explorer now reads persisted governance runtime records rather than helper-only architecture examples.</p>
  <p class="muted">What remains blocked is truthful too: live voting power, signed vote submission and treasury execution are still withheld until the supporting runtime exists.</p>
</div>
<div class="portal-actions" style="margin-bottom:12px"><a class="portal-button" href="/governance">Governance Home</a><a class="portal-button secondary" href="/governance/treasury">Treasury Safety</a><a class="portal-button ghost" href="/governance/audit">Audit Trail</a></div>
<table><tr><th>ID</th><th>Title</th><th>Type</th><th>Status</th><th>Quorum Required</th><th>Participation</th><th>Queue</th><th>Execution</th></tr>{{range .GovernanceRuntime.Proposals}}<tr><td><a href="/governance/proposals/{{.ProposalID}}">{{.ProposalID}}</a></td><td><a href="/governance/proposals/{{.ProposalID}}">{{.Title}}</a></td><td>{{.Type}}</td><td><span class="tag">{{.Status}}</span></td><td>{{.QuorumRequired}}</td><td>{{.ParticipationRatio}}</td><td>{{.QueueState}}</td><td>{{.ExecutionState}}</td></tr>{{else}}<tr><td colspan="8" class="muted">No governance runtime records available.</td></tr>{{end}}</table>
{{end}}`, data)
}

func (s *Server) handleGovernanceProposalDetail(w http.ResponseWriter, r *http.Request) {
	data := s.explorerData("PhoenixOS Chain Governance Proposal")
	proposal, ok := s.governance.Proposal(r.PathValue("proposalId"))
	if !ok {
		writeError(w, http.StatusNotFound, "proposal not found")
		return
	}
	runtime := s.governanceData()
	action, hasAction := findTreasuryAction(runtime.TreasuryAction, proposal.ProposalID)
	data["GovernanceProposal"] = proposal
	data["GovernanceRuntime"] = runtime
	data["GovernanceAnalyst"] = governanceAnalystSummary(proposal)
	data["TreasuryAction"] = action
	data["HasTreasuryAction"] = hasAction
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>{{.GovernanceProposal.Title}}</h1><span class="muted">{{.GovernanceProposal.ProposalID}} | {{.GovernanceProposal.Type}}</span></div>
<section class="grid">
  <div class="metric"><div class="label">Status</div><div class="value">{{.GovernanceProposal.Status}}</div></div>
  <div class="metric"><div class="label">Execution</div><div class="value">{{.GovernanceProposal.ExecutionState}}</div></div>
  <div class="metric"><div class="label">Queue</div><div class="value">{{.GovernanceProposal.QueueState}}</div></div>
  <div class="metric"><div class="label">Timelock</div><div class="value">{{.GovernanceProposal.TimelockState}}</div></div>
  <div class="metric"><div class="label">Voting Start</div><div class="value">{{.GovernanceProposal.VotingStart}}</div></div>
  <div class="metric"><div class="label">Voting End</div><div class="value">{{.GovernanceProposal.VotingEnd}}</div></div>
  <div class="metric"><div class="label">Remaining Time</div><div class="value">{{.GovernanceProposal.RemainingTime}}</div></div>
  <div class="metric"><div class="label">Power State</div><div class="value">{{.GovernanceProposal.PowerState}}</div></div>
</section>
<div class="panel">
  <p>Proposer: {{.GovernanceProposal.Proposer}}</p>
  <p>Created at: {{.GovernanceProposal.CreatedAt}}</p>
  <p>Description: {{.GovernanceProposal.Description}}</p>
  <p>Quorum required: {{.GovernanceProposal.QuorumRequired}}</p>
  <p>Threshold required: {{.GovernanceProposal.ThresholdRequired}}</p>
  <p>Quorum reached: {{.GovernanceProposal.QuorumReached}}</p>
  <p>Veto threshold: {{.GovernanceProposal.VetoThreshold}}</p>
</div>
<div class="section-title"><h2>Vote Runtime</h2><span class="muted">Truth-first vote visibility</span></div>
<table><tr><th>Yes</th><th>No</th><th>Abstain</th><th>Veto</th><th>Validator Votes</th><th>Community Votes</th></tr><tr><td>{{.GovernanceProposal.Yes}}</td><td>{{.GovernanceProposal.No}}</td><td>{{.GovernanceProposal.Abstain}}</td><td>{{.GovernanceProposal.Veto}}</td><td>{{.GovernanceProposal.ValidatorVotes}}</td><td>{{.GovernanceProposal.CommunityVotes}}</td></tr></table>
<div class="panel">
  <p>Participation ratio: {{.GovernanceProposal.ParticipationRatio}}</p>
  <p>Voting power: {{.GovernanceProposal.PowerState}}</p>
  <p class="muted">No vote is fabricated here. This proposal remains visible even when vote settlement is blocked.</p>
</div>
{{if .HasTreasuryAction}}
<div class="section-title"><h2>Treasury Safety</h2><span class="muted">Action preview, not fund movement</span></div>
<div class="panel">
  <p>Recipient: {{.TreasuryAction.Recipient}}</p>
  <p>Recipient valid: {{.TreasuryAction.RecipientValid}}</p>
  <p>Reason: {{.TreasuryAction.AllocationReason}}</p>
  <p>Spending cap: {{.TreasuryAction.SpendingCap}}</p>
  <p>Requested amount: {{.TreasuryAction.RequestedAmount}}</p>
  <p>Execution delay: {{.TreasuryAction.ExecutionDelay}}</p>
  <p>Operator approval posture: {{.TreasuryAction.OperatorApproval}}</p>
  <p>Emergency block: {{.TreasuryAction.EmergencyBlock}}</p>
  <p>Queue blocked reason: {{.TreasuryAction.QueueBlockedReason}}</p>
</div>
{{end}}
<div class="section-title"><h2>Audit Trail</h2><span class="muted">Public governance movement history</span></div>
<table><tr><th>Time</th><th>Action</th><th>Actor</th><th>Severity</th><th>Summary</th></tr>{{range .GovernanceProposal.AuditTrail}}<tr><td>{{.Timestamp}}</td><td>{{.Action}}</td><td>{{.Actor}}</td><td>{{.Severity}}</td><td>{{.Summary}}</td></tr>{{else}}<tr><td colspan="5" class="muted">No public audit entries recorded.</td></tr>{{end}}</table>
<div class="section-title"><h2>AI Governance Analyst</h2><span class="muted">No fake certainty</span></div>
<div class="panel">
  <p>{{.GovernanceAnalyst}}</p>
  <p class="muted">This summary is limited to runtime-backed proposal state, queue posture and visible public audit entries.</p>
</div>
{{end}}`, data)
}

func (s *Server) handleGovernanceTreasury(w http.ResponseWriter, r *http.Request) {
	data := s.explorerData("PhoenixOS Chain Treasury Safety")
	runtime := s.governanceData()
	data["GovernanceRuntime"] = runtime
	data["GovernanceSecurityNotes"] = governanceSecurityNotes()
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Treasury Safety</h1><span class="muted">Governance-controlled fund discipline</span></div>
<section class="grid">
  <div class="metric"><div class="label">Treasury Overview</div><div class="value">{{.GovernanceRuntime.Treasury.GovernanceFunds}}</div></div>
  <div class="metric"><div class="label">Action Previews</div><div class="value">{{.GovernanceRuntime.Treasury.ActionCount}}</div></div>
  <div class="metric"><div class="label">Queue Blocked</div><div class="value">{{index .GovernanceRuntime.Summary.QueueCounts "blocked"}}</div></div>
  <div class="metric"><div class="label">Treasury Safety</div><div class="value">{{.GovernanceRuntime.Summary.TreasurySafety}}</div></div>
</section>
<div class="panel">
  <p>Treasury wallet: {{.GovernanceRuntime.Treasury.Wallet}}</p>
  <p>Governance-controlled funds: {{.GovernanceRuntime.Treasury.GovernanceFunds}}</p>
  <p>Ecosystem reserve: {{.GovernanceRuntime.Treasury.EcosystemReserve}}</p>
  <p>Validator reserve: {{.GovernanceRuntime.Treasury.ValidatorReserve}}</p>
  <p>Grant pool: {{.GovernanceRuntime.Treasury.GrantPool}}</p>
  <p>Operational reserve: {{.GovernanceRuntime.Treasury.OperationalReserve}}</p>
  <p>Treasury history: {{.GovernanceRuntime.Treasury.TreasuryHistory}}</p>
</div>
<table><tr><th>Action</th><th>Proposal</th><th>Recipient</th><th>Recipient Valid</th><th>Cap</th><th>Delay</th><th>Queue</th><th>Execution</th></tr>{{range .GovernanceRuntime.TreasuryAction}}<tr><td>{{.ActionID}}</td><td><a href="/governance/proposals/{{.ProposalID}}">{{.ProposalID}}</a></td><td class="hash">{{.Recipient}}</td><td>{{.RecipientValid}}</td><td>{{.SpendingCap}}</td><td>{{.ExecutionDelay}}</td><td>{{.QueueState}}</td><td>{{.ExecutionState}}</td></tr>{{else}}<tr><td colspan="8" class="muted">No treasury governance actions recorded.</td></tr>{{end}}</table>
<div class="section-title"><h2>Security Posture</h2><span class="muted">No fake treasury movement</span></div>
<ul class="portal-list">{{range .GovernanceSecurityNotes}}<li>{{.}}</li>{{end}}</ul>
{{end}}`, data)
}

func (s *Server) handleGovernanceAudit(w http.ResponseWriter, r *http.Request) {
	data := s.explorerData("PhoenixOS Chain Governance Audit")
	runtime := s.governanceData()
	data["GovernanceRuntime"] = runtime
	renderExplorer(w, `{{define "content"}}
<div class="section-title"><h1>Governance Audit Trail</h1><span class="muted">Public governance actions and runtime blockers</span></div>
<div class="panel">
  <p>Public audit trail is explorer-readable. Sensitive operator notes remain admin-gated and are intentionally excluded from this surface.</p>
</div>
<table><tr><th>Time</th><th>Proposal</th><th>Action</th><th>Actor</th><th>Severity</th><th>Summary</th></tr>{{range .GovernanceRuntime.AuditTrail}}<tr><td>{{.Timestamp}}</td><td>{{if .ProposalID}}<a href="/governance/proposals/{{.ProposalID}}">{{.ProposalID}}</a>{{else}}runtime{{end}}</td><td>{{.Action}}</td><td>{{.Actor}}</td><td>{{.Severity}}</td><td>{{.Summary}}</td></tr>{{else}}<tr><td colspan="6" class="muted">No public governance audit entries recorded.</td></tr>{{end}}</table>
<div class="section-title"><h2>Runtime Audit</h2><span class="muted">Hardening blockers</span></div>
<table><tr><th>Risk</th><th>Severity</th><th>Blocker</th><th>Runtime Impact</th><th>Required Fix</th></tr>{{range .GovernanceRuntime.Audit}}<tr><td>{{.Risk}}</td><td>{{.Severity}}</td><td>{{.Blocker}}</td><td>{{.RuntimeImpact}}</td><td>{{.RequiredFix}}</td></tr>{{else}}<tr><td colspan="5" class="muted">No governance audit findings recorded.</td></tr>{{end}}</table>
{{end}}`, data)
}

func governancePortalMetrics(summary governance.RuntimeSummary) map[string]string {
	return map[string]string{
		"active":     fmt.Sprintf("%d", summary.Counts["active"]),
		"passed":     fmt.Sprintf("%d", summary.Counts["passed"]),
		"rejected":   fmt.Sprintf("%d", summary.Counts["rejected"]),
		"queued":     fmt.Sprintf("%d", summary.QueueCounts["queued"]+summary.QueueCounts["timelock_pending"]+summary.QueueCounts["executable"]),
		"blocked":    fmt.Sprintf("%d", summary.QueueCounts["blocked"]),
		"executed":   fmt.Sprintf("%d", summary.Counts["executed"]),
	}
}
