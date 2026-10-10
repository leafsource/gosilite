package storage

import (
	storageModel "gosilite/api/model/framework/storage"
	storageService "gosilite/application/framework/service/storage"
)

func toStorageInfoResponse(vo *storageService.InfoVO) *storageModel.StorageInfoResponse {
	return &storageModel.StorageInfoResponse{
		Endpoint:    vo.Endpoint,
		Bucket:      vo.Bucket,
		Domain:      vo.Domain,
		Connected:   vo.Connected,
		ObjectCount: vo.ObjectCount,
		TotalSize:   vo.TotalSize,
		Message:     vo.Message,
	}
}

func toStorageObjectListResponse(vos []*storageService.ObjectVO, total int64) *storageModel.StorageObjectListResponse {
	out := &storageModel.StorageObjectListResponse{
		List:  make([]*storageModel.StorageObjectDTO, 0, len(vos)),
		Total: total,
	}
	for _, vo := range vos {
		out.List = append(out.List, &storageModel.StorageObjectDTO{
			Key:          vo.Key,
			Size:         vo.Size,
			Etag:         vo.Etag,
			LastModified: vo.LastModified,
			Type:         vo.Type,
		})
	}
	return out
}

func toStorageUploadResponse(vo *storageService.UploadVO) *storageModel.StorageUploadResponse {
	return &storageModel.StorageUploadResponse{
		WebURI: vo.WebURI,
		Key:    vo.Key,
	}
}
