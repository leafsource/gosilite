package dept

import (
	"gosilite/application/framework/dal/model"
)

// toNode 部门模型 → 树节点（负责人字段由 attachLeaders 补充）
func toNode(d *model.SystemDept) *Node {
	return &Node{
		ID: d.ID, ParentID: d.ParentID, Label: d.DeptName,
		OrderNum: d.OrderNum, Status: d.Status,
	}
}
