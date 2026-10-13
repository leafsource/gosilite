package idparse

import (
	"reflect"
	"testing"
)

func TestInt64s(t *testing.T) {
	tests := []struct {
		name    string
		values  []string
		want    []int64
		wantErr bool
	}{
		{name: "nil", values: nil, want: nil},
		{name: "empty", values: []string{}, want: []int64{}},
		{name: "snowflake", values: []string{"6855337641038665531"}, want: []int64{6855337641038665531}},
		{name: "invalid", values: []string{"1.5"}, wantErr: true},
		{name: "overflow", values: []string{"9223372036854775808"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Int64s(tt.values)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Int64s() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Int64s() = %v, want %v", got, tt.want)
			}
		})
	}
}
