package idparse

import "strconv"

func Int64s(values []string) ([]int64, error) {
	if values == nil {
		return nil, nil
	}
	ids := make([]int64, len(values))
	for i, value := range values {
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		ids[i] = id
	}
	return ids, nil
}
