package monitor

// 缓存监控 VO。

type CachePrefixVO struct {
	Prefix string `json:"prefix"`
	Count  int64  `json:"count"`
	Remark string `json:"remark"`
}

type CacheKeyVO struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Field string `json:"field,omitempty"`
}

type CacheValueVO struct {
	Key        string `json:"key"`
	Type       string `json:"type"`
	Value      string `json:"value"`
	DisplayKey string `json:"displayKey"`
}
