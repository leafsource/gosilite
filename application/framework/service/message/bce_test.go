package message

import (
	"strings"
	"testing"
	"time"
)

// 官方向量：https://cloud.baidu.com/doc/Reference/s/njwvz1yfu 「认证字符串示例」
// Access Key ID = 32 个 a；Secret Access Key = 32 个 b；
// 时间 2015-04-27T08:23:49Z，有效期 1800s；
// PUT /v1/test/myfolder/readme.txt?partNumber=9&uploadId=...，body "Example"。

func TestBceSigningKeyOfficialVector(t *testing.T) {
	sk := strings.Repeat("b", 32)
	prefix := "bce-auth-v1/" + strings.Repeat("a", 32) + "/2015-04-27T08:23:49Z/1800"
	got := bceSigningKey(sk, prefix)
	want := "1d5ce5f464064cbee060330d973218821825ac6952368a482a592e6615aef479"
	if got != want {
		t.Fatalf("signingKey = %s, want %s", got, want)
	}
}

func TestBceSignatureOfficialVector(t *testing.T) {
	sk := strings.Repeat("b", 32)
	prefix := "bce-auth-v1/" + strings.Repeat("a", 32) + "/2015-04-27T08:23:49Z/1800"

	headers := map[string]string{
		"content-length": "8",
		"content-md5":    "NFzcPqhviddjRNnSOGo4rw==",
		"content-type":   "text/plain",
		"host":           "bj.bcebos.com",
		"x-bce-date":     "2015-04-27T08:23:49Z",
	}
	canonicalHeaders, signedHeaders := bceCanonicalHeaders(headers)
	wantCH := "content-length:8\n" +
		"content-md5:NFzcPqhviddjRNnSOGo4rw%3D%3D\n" +
		"content-type:text%2Fplain\n" +
		"host:bj.bcebos.com\n" +
		"x-bce-date:2015-04-27T08%3A23%3A49Z"
	if canonicalHeaders != wantCH {
		t.Fatalf("canonicalHeaders =\n%s\nwant\n%s", canonicalHeaders, wantCH)
	}
	if signedHeaders != "content-length;content-md5;content-type;host;x-bce-date" {
		t.Fatalf("signedHeaders = %s", signedHeaders)
	}

	query := "partNumber=9&uploadId=a44cc9bab11cbd156984767aad637851"
	cr, _ := bceCanonicalRequest("PUT", "/v1/test/myfolder/readme.txt", query, headers)
	wantCR := "PUT\n" +
		"/v1/test/myfolder/readme.txt\n" +
		query + "\n" +
		wantCH
	if cr != wantCR {
		t.Fatalf("canonicalRequest =\n%s\nwant\n%s", cr, wantCR)
	}

	got := bceSignature(bceSigningKey(sk, prefix), cr)
	want := "d74a04362e6a848f5b39b15421cb449427f419c95a480fd6b8cf9fc783e2999e"
	if got != want {
		t.Fatalf("signature = %s, want %s", got, want)
	}
}

func TestBceAuthStringFormat(t *testing.T) {
	ts := time.Date(2019, 12, 4, 6, 53, 12, 0, time.UTC)
	headers := map[string]string{"host": "sms.bj.baidubce.com", "x-bce-date": "2019-12-04T06:53:12Z"}
	auth := bceAuthString("ak-1", "sk-1", ts, 3600, "POST", "/api/v3/sendSms", "", headers)
	parts := strings.Split(auth, "/")
	if len(parts) != 6 {
		t.Fatalf("auth parts = %d, %q", len(parts), auth)
	}
	if parts[0] != "bce-auth-v1" || parts[1] != "ak-1" || parts[2] != "2019-12-04T06:53:12Z" || parts[3] != "3600" {
		t.Fatalf("auth prefix wrong: %s", auth)
	}
	if parts[4] != "host;x-bce-date" {
		t.Fatalf("signedHeaders = %s", parts[4])
	}
	if len(parts[5]) != 64 {
		t.Fatalf("signature should be 64 hex chars, got %q", parts[5])
	}
	for _, r := range parts[5] {
		if !strings.ContainsRune("0123456789abcdef", r) {
			t.Fatalf("signature should be lowercase hex: %q", parts[5])
		}
	}
}

func TestBceUriEncode(t *testing.T) {
	if got := uriEncodeExceptSlash("/v1/test/测试"); got != "/v1/test/%E6%B5%8B%E8%AF%95" {
		t.Fatalf("uriEncodeExceptSlash = %s", got)
	}
	if got := uriEncode("a b/c", true); got != "a%20b%2Fc" {
		t.Fatalf("uriEncode = %s", got)
	}
	if got := uriEncode("_.-~.", false); got != "_.-~." {
		t.Fatalf("unreserved chars should stay: %s", got)
	}
}
