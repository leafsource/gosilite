package monitor

import (
	monitor "gosilite/api/model/framework/monitor"
	userService "gosilite/application/framework/service/user"
)

const timeLayout = "2006-01-02 15:04:05"

func toOnlineDTO(r userService.OnlineRow) *monitor.OnlineDTO {
	return &monitor.OnlineDTO{
		ID:       r.AccessHash,
		UserId:   r.UserID,
		UserName: r.Account,
		DeptName: r.DeptName,
		ClientIp: r.IP,
		Browser:  r.Browser,
		Os:       r.OS,
		LoginAt:  r.LoginAt.Format(timeLayout),
	}
}

func toOnlineDTOs(rows []userService.OnlineRow) []*monitor.OnlineDTO {
	out := make([]*monitor.OnlineDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toOnlineDTO(r))
	}
	return out
}
