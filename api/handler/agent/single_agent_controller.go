package agent

import (
	"context"
	R "gosilite/pkg/utils/response"

	agent "gosilite/api/model/agent"
	singleAgentDraftService "gosilite/application/agent/service/single_agent_draft"

	"github.com/cloudwego/hertz/pkg/app"
)

// GetPageList .
// @router /agent/list [GET]
func GetPageList(ctx context.Context, c *app.RequestContext) {
	resp := &agent.AgentDraftResponse{
		Items: []*agent.AgentDraft{},
	}
	// TODO: 实现业务逻辑
	list, _, _ := singleAgentDraftService.SingleAgentDraftSVC.List(ctx, 1, 10)
	for _, item := range list {
		resp.Items = append(resp.Items, &agent.AgentDraft{
			Id:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			IconUri:     item.IconURI,
			ModelInfo:   item.ModelInfo,
		})
	}
	R.SuccessWithData(ctx, c, resp)
}
