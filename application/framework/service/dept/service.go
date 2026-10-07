package dept

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"gosilite/application/framework/dal/model"
	"gosilite/application/framework/dal/repository"
	"gosilite/pkg/errorx"
	"gosilite/types/errno"

	"gorm.io/gorm"
)

// SystemDeptSVC 门面
var SystemDeptSVC *Service

// Service 部门业务
type Service struct {
	repo repository.IDepartmentRepository
}

// NewService 构造部门服务
func NewService(db *gorm.DB) *Service {
	return &Service{repo: repository.NewDepartmentRepository(db)}
}

// InitService 安装包级门面
func InitService(ctx context.Context, db *gorm.DB) *Service {
	SystemDeptSVC = NewService(db)
	return SystemDeptSVC
}

// Node 树节点
type Node struct {
	ID         int64   `json:"id"`
	ParentID   int64   `json:"parentId"`
	Label      string  `json:"label"`
	OrderNum   int32   `json:"orderNum"`
	Status     int16   `json:"status"`
	LeaderID   int64   `json:"leaderId"`
	LeaderName string  `json:"leaderName"`
	Children   []*Node `json:"children,omitempty"`
}

// Tree 返回部门树（附带负责人）
func (s *Service) Tree(ctx context.Context) ([]*Node, error) {
	list, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	leaders, err := s.repo.LeaderInfo(ctx)
	if err != nil {
		return nil, err
	}
	leaderMap := make(map[int64]repository.DeptLeader, len(leaders))
	for _, l := range leaders {
		leaderMap[l.DeptID] = l
	}
	return attachLeaders(buildTree(list), leaderMap), nil
}

func attachLeaders(nodes []*Node, leaders map[int64]repository.DeptLeader) []*Node {
	for _, n := range nodes {
		if l, ok := leaders[n.ID]; ok {
			n.LeaderID = l.LeaderID
			n.LeaderName = l.LeaderName
		}
		if len(n.Children) > 0 {
			attachLeaders(n.Children, leaders)
		}
	}
	return nodes
}

// UpdateLeader 设置/清除部门负责人（leaderUserID=0 清除）
func (s *Service) UpdateLeader(ctx context.Context, deptID, leaderUserID int64) error {
	if deptID <= 0 {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "部门不能为空"))
	}
	if _, err := s.repo.GetByID(ctx, deptID); err != nil {
		return errorx.New(errno.ErrSystemDeptNotFound, errorx.KV("id", itoa(deptID)))
	}
	if leaderUserID > 0 {
		ok, err := s.repo.UserExists(ctx, leaderUserID)
		if err != nil {
			return err
		}
		if !ok {
			return errorx.New(errno.ErrSystemUserNotFound, errorx.KV("id", itoa(leaderUserID)))
		}
	}
	return s.repo.UpdateLeader(ctx, deptID, leaderUserID)
}

// List 扁平列表
func (s *Service) List(ctx context.Context) ([]*model.SystemDept, error) {
	return s.repo.ListAll(ctx)
}

// Create 新建部门
func (s *Service) Create(ctx context.Context, d *model.SystemDept) error {
	if strings.TrimSpace(d.DeptName) == "" {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "部门名称不能为空"))
	}
	parentAncestors := "0"
	if d.ParentID > 0 {
		parent, err := s.repo.GetByID(ctx, d.ParentID)
		if err != nil {
			return errorx.New(errno.ErrSystemDeptNotFound, errorx.KV("id", itoa(d.ParentID)))
		}
		parentAncestors = parent.Ancestors
	}
	d.Ancestors = parentAncestors + "," + itoa(d.ParentID)
	if d.ParentID == 0 {
		d.Ancestors = "0"
	}
	d.DelFlag = 0
	return s.repo.Create(ctx, d)
}

// Update 更新部门（含换父节点后子树 ancestors 重建）
func (s *Service) Update(ctx context.Context, d *model.SystemDept) error {
	old, err := s.repo.GetByID(ctx, d.ID)
	if err != nil {
		return errorx.New(errno.ErrSystemDeptNotFound, errorx.KV("id", itoa(d.ID)))
	}
	if d.ParentID == d.ID {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "不能将部门移动到自身"))
	}
	newParentAncestors := "0"
	if d.ParentID > 0 {
		parent, err := s.repo.GetByID(ctx, d.ParentID)
		if err != nil {
			return errorx.New(errno.ErrSystemDeptNotFound, errorx.KV("id", itoa(d.ParentID)))
		}
		// 环检测：新父级的祖先链含本部门 id ⇒ 父级在本部门子树内 ⇒ 成环。
		// 未换父时 parent.Ancestors == old.Ancestors（不含本部门自身）→ 正常放行。
		// 基于 DB 中的 ancestors，不信任客户端 payload。
		if isDescendant(d.ID, parent.Ancestors) {
			return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "不能移动到子部门"))
		}
		newParentAncestors = parent.Ancestors
	}
	if strings.TrimSpace(d.DeptName) == "" {
		return errorx.New(errno.ErrSystemInvalidParam, errorx.KV("msg", "部门名称不能为空"))
	}

	// 计算新 ancestors
	oldAncestors := old.Ancestors
	newAncestors := "0"
	if d.ParentID > 0 {
		newAncestors = newParentAncestors + "," + itoa(d.ParentID)
	}

	d.Ancestors = newAncestors
	d.DelFlag = old.DelFlag
	d.CreatedAt = old.CreatedAt
	if err := s.repo.Update(ctx, d); err != nil {
		return err
	}

	if oldAncestors != newAncestors {
		all, err := s.repo.ListAll(ctx)
		if err != nil {
			return err
		}
		moved := "," + itoa(d.ID) + ","
		for _, c := range all {
			if c.ID == d.ID {
				continue
			}
			if strings.Contains(","+c.Ancestors+",", moved) {
				c.Ancestors = strings.Replace(c.Ancestors, oldAncestors, newAncestors, 1)
				if err := s.repo.Update(ctx, c); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Delete 逻辑删除，禁止有子部门或用户
func (s *Service) Delete(ctx context.Context, id int64) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return errorx.New(errno.ErrSystemDeptNotFound, errorx.KV("id", itoa(id)))
	}
	n, err := s.repo.CountChildren(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return errorx.New(errno.ErrSystemDeptHasChildren, errorx.KV("id", itoa(id)))
	}
	u, err := s.repo.CountUsers(ctx, id)
	if err != nil {
		return err
	}
	if u > 0 {
		return errorx.New(errno.ErrSystemDeptHasUsers, errorx.KV("id", itoa(id)))
	}
	return s.repo.SoftDelete(ctx, id)
}

// VisibleDeptIDs 按数据范围计算可见部门 id（管理过滤用）
// dataScope: 1全部 2本部门 3本部门及以下 4仅本人 5自定义
func (s *Service) VisibleDeptIDs(ctx context.Context, dataScope int16, userDeptID int64, customIDs []int64) (all bool, ids []int64, err error) {
	switch dataScope {
	case 1:
		return true, nil, nil
	case 4:
		// 仅本人：返回空部门集（列表按用户 id 收窄，而非本部门）
		return false, nil, nil
	case 5:
		if len(customIDs) == 0 {
			return false, nil, nil
		}
		return false, customIDs, nil
	case 2, 3:
		all, err := s.repo.ListAll(ctx)
		if err != nil {
			return false, nil, err
		}
		self, _ := s.repo.GetByID(ctx, userDeptID)
		if self == nil {
			return false, nil, nil
		}
		if dataScope == 2 {
			return false, []int64{userDeptID}, nil
		}
		// 本部门及以下：ancestors 含 userDeptID
		ids = append(ids, userDeptID)
		for _, d := range all {
			if strings.Contains(","+d.Ancestors+",", ","+itoa(userDeptID)+",") {
				ids = append(ids, d.ID)
			}
		}
		return false, ids, nil
	default:
		// fail-closed：未知 scope 不放行
		return false, nil, nil
	}
}

func buildTree(list []*model.SystemDept) []*Node {
	nodes := map[int64]*Node{}
	for _, d := range list {
		nodes[d.ID] = toNode(d)
	}
	var roots []*Node
	for id, n := range nodes {
		if p, ok := nodes[n.ParentID]; ok && id != n.ParentID {
			p.Children = append(p.Children, n)
		} else {
			roots = append(roots, n)
		}
	}
	var sortFn func([]*Node)
	sortFn = func(ns []*Node) {
		sort.Slice(ns, func(i, j int) bool {
			if ns[i].OrderNum != ns[j].OrderNum {
				return ns[i].OrderNum < ns[j].OrderNum
			}
			return ns[i].ID < ns[j].ID
		})
		for _, n := range ns {
			sortFn(n.Children)
		}
	}
	sortFn(roots)
	return roots
}

func isDescendant(parentID int64, ancestors string) bool {
	return parentID > 0 && strings.Contains(","+ancestors+",", ","+itoa(parentID)+",")
}

func itoa(v int64) string {
	return strconv.FormatInt(v, 10)
}
