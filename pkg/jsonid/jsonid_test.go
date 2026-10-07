package jsonid

import (
	"strings"
	"testing"
)

func TestTransformOutQuotesIDs(t *testing.T) {
	in := []byte(`{"code":200,"msg":"ok 123","data":{"id":1961643802838147072,"parentId":100,"total":5,"roleIds":[1,2],"name":"z"}}`)
	got := string(TransformOut(in))
	for _, want := range []string{
		`"id":"1961643802838147072"`,
		`"parentId":"100"`,
		`"roleIds":["1","2"]`,
		`"code":200`,
		`"total":5`,
		`"msg":"ok 123"`,
		`"name":"z"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("want %s in %s", want, got)
		}
	}
}

func TestTransformInUnquotesIDs(t *testing.T) {
	in := []byte(`{"id":"1961643802838147072","deptId":"0","roleIds":["1","2"],"mobilePhone":"17637474888","name":"z"}`)
	got := string(TransformIn(in))
	for _, want := range []string{
		`"id":1961643802838147072`,
		`"deptId":0`,
		`"roleIds":[1,2]`,
		`"mobilePhone":"17637474888"`,
		`"name":"z"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("want %s in %s", want, got)
		}
	}
}

func TestRoundTripKeepsExactDigits(t *testing.T) {
	big := "1961643802838147072"
	resp := TransformOut([]byte(`{"data":{"id":` + big + `}}`))
	if !strings.Contains(string(resp), `"`+big+`"`) {
		t.Fatalf("out want quoted digits: %s", resp)
	}
	back := TransformIn(resp)
	if !strings.Contains(string(back), `{"data":{"id":`+big+`}}`) {
		t.Fatalf("back want exact number: %s", back)
	}
}
