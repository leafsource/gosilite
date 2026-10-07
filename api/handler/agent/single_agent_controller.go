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
	// TODO: 实现业务逻辑
	list, _, _ := singleAgentDraftService.SingleAgentDraftSVC.List(ctx, 1, 10)
	resp := &agent.AgentDraftResponse{
		Items: toAgentDraftDTOs(list),
	}
	R.SuccessWithData(ctx, c, resp)
}
