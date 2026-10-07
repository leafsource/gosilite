package agent

import (
	agent "gosilite/api/model/agent"
	"gosilite/application/agent/dal/model"
)

func toAgentDraftDTO(m *model.SingleAgentDraft) *agent.AgentDraft {
	return &agent.AgentDraft{
		Id:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		IconUri:     m.IconURI,
		ModelInfo:   m.ModelInfo,
	}
}

func toAgentDraftDTOs(list []*model.SingleAgentDraft) []*agent.AgentDraft {
	out := make([]*agent.AgentDraft, 0, len(list))
	for _, m := range list {
		out = append(out, toAgentDraftDTO(m))
	}
	return out
}
