package dxf

import (
	"strings"
	"testing"
)

const sampleDXF = `0
SECTION
2
ENTITIES
0
LINE
8
0
10
0.0
20
0.0
11
100.0
21
50.0
0
CIRCLE
8
0
10
50.0
20
25.0
40
20.0
0
LWPOLYLINE
8
0
90
2
70
1
10
0.0
20
0.0
10
10.0
20
10.0
0
ARC
8
0
10
0.0
20
0.0
40
30.0
50
0.0
51
90.0
0
TEXT
8
0
10
5.0
20
5.0
40
2.5
1
HELLO
0
POINT
8
0
10
99.0
20
99.0
0
ENDSEC
0
EOF
`

func TestRender(t *testing.T) {
	res, err := Render([]byte(sampleDXF))
	if err != nil {
		t.Fatalf("Render 失败: %v", err)
	}
	if res.EntityCount != 6 {
		t.Errorf("实体数 = %d, 期望 6", res.EntityCount)
	}
	for _, want := range []string{"<svg", "<line", "<circle", "<polyline", "<path"} {
		if !strings.Contains(res.SVG, want) {
			t.Errorf("SVG 缺少 %s", want)
		}
	}
	// TEXT 应列为未支持
	found := false
	for _, u := range res.Unsupported {
		if u == "TEXT" {
			found = true
		}
	}
	if !found {
		t.Errorf("Unsupported 应包含 TEXT，实际 %v", res.Unsupported)
	}
}

func TestRenderInvalid(t *testing.T) {
	if _, err := Render([]byte("not a dxf file")); err == nil {
		t.Errorf("非法输入应返回错误")
	}
}
