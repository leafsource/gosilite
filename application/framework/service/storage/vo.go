package storage

// InfoVO 存储概览：配置 + 探活 + 用量聚合
type InfoVO struct {
	Endpoint    string
	Bucket      string
	Domain      string
	Connected   bool
	ObjectCount int64
	TotalSize   int64
	Message     string
}

// ObjectVO 对象列表条目
type ObjectVO struct {
	Key          string
	Size         int64
	Etag         string
	LastModified string
	Type         string
}

// UploadVO 上传结果
type UploadVO struct {
	WebURI string
	Key    string
}
