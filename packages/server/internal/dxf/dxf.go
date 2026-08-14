// Package dxf 提供轻量 DXF(ASCII) → SVG 在线预览渲染。
// 无第三方依赖：解析 ENTITIES 段常见实体（LINE/LWPOLYLINE/CIRCLE/ARC/POINT），
// 输出可内联展示的 SVG。文本/块引用/填充等复杂实体暂不渲染（结果中标注）。
package dxf

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// RenderResult DXF 渲染结果
type RenderResult struct {
	SVG         string   `json:"svg"`
	EntityCount int      `json:"entity_count"`
	Unsupported []string `json:"unsupported"` // 未渲染的实体类型（去重）
}

// Render 将 DXF 文本转换为 SVG
func Render(input []byte) (*RenderResult, error) {
	entities, err := parseEntities(input)
	if err != nil {
		return nil, err
	}
	if len(entities) == 0 {
		return nil, fmt.Errorf("DXF 中未找到可渲染实体")
	}

	res := &RenderResult{Unsupported: []string{}}
	seen := map[string]bool{}

	// 第一遍：收集边界与实体几何
	type pt struct{ x, y float64 }
	var lines [][4]float64  // x1,y1,x2,y2
	var polylines [][]pt
	var circles [][3]float64 // cx,cy,r
	var arcs [][6]float64    // cx,cy,r,startDeg,endDeg,sweepDeg
	var points []pt
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)

	extend := func(x, y float64) {
		if math.IsInf(minX, 1) {
			minX, minY, maxX, maxY = x, y, x, y
			return
		}
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x > maxX {
			maxX = x
		}
		if y > maxY {
			maxY = y
		}
	}

	for _, e := range entities {
		res.EntityCount++
		// 组码 → 值列表（同一组码可多次出现，如 LWPOLYLINE 顶点 10/20）
		vals := map[int][]string{}
		for _, t := range e.groups {
			vals[t.code] = append(vals[t.code], t.value)
		}
		first := func(codes ...int) float64 {
			for _, c := range codes {
				if v, ok := vals[c]; ok && len(v) > 0 {
					return f(v[len(v)-1])
				}
			}
			return 0
		}
		switch e.name {
		case "LINE":
			x1, y1 := first(10), first(20)
			x2, y2 := first(11), first(21)
			lines = append(lines, [4]float64{x1, y1, x2, y2})
			extend(x1, y1)
			extend(x2, y2)
		case "LWPOLYLINE":
			// 顶点：10/20 成对出现
			xs, ys := vals[10], vals[20]
			var poly []pt
			for k := 0; k < len(xs) && k < len(ys); k++ {
				xk, yk := f(xs[k]), f(ys[k])
				poly = append(poly, pt{xk, yk})
				extend(xk, yk)
			}
			if len(poly) >= 2 {
				polylines = append(polylines, poly)
			}
		case "CIRCLE":
			cx, cy, r := first(10), first(20), first(40)
			circles = append(circles, [3]float64{cx, cy, r})
			extend(cx-r, cy-r)
			extend(cx+r, cy+r)
		case "ARC":
			cx, cy, r := first(10), first(20), first(40)
			a1, a2 := first(50), first(51)
			sweep := math.Mod(a2-a1, 360)
			if sweep < 0 {
				sweep += 360
			}
			arcs = append(arcs, [6]float64{cx, cy, r, a1, a2, sweep})
			extend(cx-r, cy-r)
			extend(cx+r, cy+r)
		case "POINT":
			x, y := first(10), first(20)
			points = append(points, pt{x, y})
			extend(x, y)
		default:
			if !seen[e.name] {
				seen[e.name] = true
				res.Unsupported = append(res.Unsupported, e.name)
			}
		}
	}

	if math.IsInf(minX, 1) {
		return nil, fmt.Errorf("DXF 中未找到可渲染实体")
	}

	// 边界留 5% 边距
	mx, my := (maxX-minX)*0.05, (maxY-minY)*0.05
	if mx == 0 {
		mx = 1
	}
	if my == 0 {
		my = 1
	}
	minX, minY = minX-mx, minY-my
	maxX, maxY = maxX+mx, maxY+my

	// 第二遍：输出 SVG（DXF Y 向上，SVG Y 向下 → 翻转）
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="%f %f %f %f" width="100%%" height="100%%">`,
		minX, -maxY, maxX-minX, maxY-minY)
	b.WriteString(`<rect x="0" y="0" width="100%" height="100%" fill="#ffffff"/>`)
	fmt.Fprintf(&b, `<g transform="scale(1,-1)" fill="none" stroke="#1e293b" stroke-width="%f" stroke-linecap="round" stroke-linejoin="round">`,
		math.Max(0.5, (maxX-minX)*0.0015))

	flip := func(x, y float64) string {
		return fmt.Sprintf("%f,%f", x, -y)
	}
	for _, l := range lines {
		fmt.Fprintf(&b, `<line x1="%f" y1="%f" x2="%f" y2="%f"/>`, l[0], -l[1], l[2], -l[3])
	}
	for _, p := range polylines {
		var pts []string
		for _, v := range p {
			pts = append(pts, flip(v.x, v.y))
		}
		fmt.Fprintf(&b, `<polyline points="%s"/>`, strings.Join(pts, " "))
	}
	for _, c := range circles {
		fmt.Fprintf(&b, `<circle cx="%f" cy="%f" r="%f"/>`, c[0], -c[1], c[2])
	}
	for _, a := range arcs {
		cx, cy, r, a1, a2, sweep := a[0], a[1], a[2], a[3], a[4], a[5]
		sx, sy := cx+r*math.Cos(a1*math.Pi/180), cy+r*math.Sin(a1*math.Pi/180)
		ex, ey := cx+r*math.Cos(a2*math.Pi/180), cy+r*math.Sin(a2*math.Pi/180)
		large := 0
		if sweep > 180 {
			large = 1
		}
		fmt.Fprintf(&b, `<path d="M %f %f A %f %f 0 %d 1 %f %f"/>`,
			sx, -sy, r, r, large, ex, -ey)
	}
	if len(points) > 0 {
		pr := math.Max(0.8, (maxX-minX)*0.004)
		for _, p := range points {
			fmt.Fprintf(&b, `<circle cx="%f" cy="%f" r="%f" fill="#2563eb"/>`, p.x, -p.y, pr)
		}
	}
	b.WriteString(`</g></svg>`)

	res.SVG = b.String()
	return res, nil
}

// entity 一个 DXF 实体（名称 + 组码序列）
type entity struct {
	name   string
	groups []token
}

// token 一个 DXF 组码/值对
type token struct {
	code  int
	value string
}

// parseEntities 解析 DXF 文件为 ENTITIES 段实体列表
func parseEntities(input []byte) ([]entity, error) {
	raw := strings.Split(string(input), "\n")
	tokens := make([]token, 0, len(raw)/2)
	for idx := 0; idx+1 < len(raw); idx += 2 {
		codeStr := strings.TrimSpace(raw[idx])
		val := strings.TrimRight(raw[idx+1], "\r")
		if codeStr == "" {
			continue
		}
		code, err := strconv.Atoi(codeStr)
		if err != nil {
			// 允许跳过异常行，避免整个文件不可预览
			continue
		}
		tokens = append(tokens, token{code: code, value: strings.TrimSpace(val)})
	}

	var entities []entity
	inEntities := false
	i := 0
	for i < len(tokens) {
		t := tokens[i]
		i++
		if t.code != 0 {
			continue
		}
		switch t.value {
		case "SECTION":
			// 下一组码 2 为段名
			if i+1 < len(tokens) && tokens[i].code == 2 {
				inEntities = tokens[i].value == "ENTITIES"
				i++
			}
		case "ENDSEC":
			inEntities = false
		default:
			if !inEntities {
				continue
			}
			// 实体开始：收集到下一个 code==0 为止
			groups := []token{}
			for i < len(tokens) && tokens[i].code != 0 {
				groups = append(groups, tokens[i])
				i++
			}
			entities = append(entities, entity{name: t.value, groups: groups})
		}
	}
	return entities, nil
}

func f(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}
