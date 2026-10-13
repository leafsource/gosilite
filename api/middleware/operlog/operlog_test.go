package operlog

import (
	"strings"
	"testing"
)

func TestPeekEnvelope(t *testing.T) {
	code, msg, ok := peekEnvelope([]byte(`{"code":200,"msg":"success","data":{"id":1}}`))
	if !ok || code != 200 || msg != "success" {
		t.Fatalf("got code=%d msg=%q ok=%v", code, msg, ok)
	}

	// 中文 msg 的 \uXXXX 转义
	code, msg, ok = peekEnvelope([]byte(`{"code":700012006,"msg":"\u767b\u5f55\u5931\u8d25"}`))
	if !ok || code != 700012006 || msg != "登录失败" {
		t.Fatalf("got code=%d msg=%q ok=%v", code, msg, ok)
	}

	// 非 JSON（如 404 纯文本）
	_, _, ok = peekEnvelope([]byte("404 page not found"))
	if ok {
		t.Fatal("plain text must not be treated as envelope")
	}

	// code 不在首位也能取到
	code, _, ok = peekEnvelope([]byte(`{"data":{},"code":500,"msg":"x"}`))
	if !ok || code != 500 {
		t.Fatalf("got code=%d ok=%v", code, ok)
	}
}

func TestClassify(t *testing.T) {
	// 成功
	st, toErr := classify(200, 200, true, false)
	if st != 0 || toErr {
		t.Fatalf("success: st=%d toErr=%v", st, toErr)
	}
	// 业务失败 → 操作失败 + 异常表
	st, toErr = classify(200, 701000028, true, false)
	if st != 1 || !toErr {
		t.Fatalf("business fail: st=%d toErr=%v", st, toErr)
	}
	// 401 鉴权拒绝 → 只进操作失败
	st, toErr = classify(401, 401, true, false)
	if st != 1 || toErr {
		t.Fatalf("401: st=%d toErr=%v", st, toErr)
	}
	// 404 无路由 → 不进异常表
	_, toErr = classify(404, 0, false, false)
	if toErr {
		t.Fatal("404 must not go to error log")
	}
	// ErrUserAuthenticationFailed → 不进异常表
	_, toErr = classify(200, 700012006, true, false)
	if toErr {
		t.Fatal("auth failed code must not go to error log")
	}
	// panic → 进异常表
	st, toErr = classify(500, 500, true, true)
	if st != 1 || !toErr {
		t.Fatalf("panic: st=%d toErr=%v", st, toErr)
	}
	// 500 → 异常表
	_, toErr = classify(500, 500, true, false)
	if !toErr {
		t.Fatal("500 must go to error log")
	}
}

func TestMaskSecrets(t *testing.T) {
	src := []byte(`{"username":"alice","password":"Admin@123","oldPassword":"x","note":"ok"}`)
	out := string(maskSecrets(src))
	if strings.Contains(out, "Admin@123") || !strings.Contains(out, "******") {
		t.Fatalf("password not masked: %s", out)
	}
	if !strings.Contains(out, `"note":"ok"`) {
		t.Fatalf("normal field damaged: %s", out)
	}
	// 数字型敏感值
	out = string(maskSecrets([]byte(`{"token":12345,"password":67890}`)))
	if strings.Contains(out, "67890") {
		t.Fatalf("numeric password not masked: %s", out)
	}
	// 无敏感字段不分配不改动
	src = []byte(`{"a":1}`)
	if string(maskSecrets(src)) != `{"a":1}` {
		t.Fatal("plain body must stay unchanged")
	}
}

func TestOperTypeOf(t *testing.T) {
	cases := []struct {
		method, path string
		want         int8
	}{
		{"GET", "/system/operlog/list", 4},
		{"POST", "/system/notice", 1},
		{"PUT", "/system/notice", 2},
		{"DELETE", "/system/notice", 3},
		{"POST", "/system/user/import", 5},
		{"GET", "/system/user/export", 6},
		{"POST", "/user/update/upload_avatar", 7},
	}
	for _, c := range cases {
		if got := operTypeOf(c.method, c.path); got != c.want {
			t.Fatalf("%s %s: got %d want %d", c.method, c.path, got, c.want)
		}
	}
}


func TestSkipPath(t *testing.T) {
	if !skipPath("/ping", "GET") || !skipPath("/swagger/index.html", "GET") || !skipPath("/any", "OPTIONS") {
		t.Fatal("health/doc/preflight must be skipped")
	}
	if skipPath("/system/operlog/list", "GET") {
		t.Fatal("api must not be skipped")
	}
}

func TestPeekLoginUsername(t *testing.T) {
	if got := peekLoginUsername([]byte(`{"username":"skyx","password":"x"}`)); got != "skyx" {
		t.Fatalf("json body: %q", got)
	}
	if got := peekLoginUsername([]byte(`username=skyx&password=x`)); got != "skyx" {
		t.Fatalf("form body: %q", got)
	}
}
