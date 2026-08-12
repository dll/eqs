#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""EQS 生产环境演示账号实测脚本
用途：验证部署后的生产功能（演示数据保留，验证后管理员可清理）。
凭据为受控演示账号（1390000 前缀 + 固定码），由项目负责人授权本脚本使用。
"""
import json
import urllib.request

BASE = "http://129.211.223.113:8091/api/v1"
OPENER = urllib.request.build_opener(urllib.request.ProxyHandler({}))  # 禁用代理直连


def call(method, path, body=None, token=None, timeout=15):
    req = urllib.request.Request(BASE + path, method=method)
    req.add_header("Content-Type", "application/json")
    if token:
        req.add_header("Authorization", "Bearer " + token)
    data = json.dumps(body).encode() if body is not None else None
    try:
        with OPENER.open(req, data, timeout=timeout) as r:
            return r.status, json.load(r)
    except urllib.error.HTTPError as e:
        try:
            return e.code, json.load(e)
        except Exception:
            return e.code, {}


def login(phone, user_type):
    s, r = call("POST", "/auth/login", {"phone": phone, "code": "123456", "user_type": user_type})
    return s, r.get("token", ""), r


def upload_file(path, token):
    """multipart/form-data 上传文件到 /qualification/upload，返回 file_id"""
    boundary = "----EQSVerifyBoundary123"
    with open(path, "rb") as f:
        content = f.read()
    body = (
        f"--{boundary}\r\n"
        f'Content-Disposition: form-data; name="file"; filename="{path.split(chr(92))[-1].split(chr(47))[-1]}"\r\n'
        f"Content-Type: application/octet-stream\r\n\r\n"
    ).encode() + content + f"\r\n--{boundary}--\r\n".encode()
    req = urllib.request.Request(BASE + "/qualification/upload", data=body, method="POST")
    req.add_header("Content-Type", f"multipart/form-data; boundary={boundary}")
    req.add_header("Authorization", "Bearer " + token)
    try:
        with OPENER.open(req, timeout=15) as r:
            return r.status, json.load(r)
    except urllib.error.HTTPError as e:
        try:
            return e.code, json.load(e)
        except Exception:
            return e.code, {}


def main():
    print("=" * 60)
    print("EQS 生产环境演示账号实测")
    print("=" * 60)

    # ---- 1. 管理员登录 + 播种(保留演示功能) ----
    s, admin, r = login("13900003333", 3)
    print(f"\n[1] 管理员登录: {s} token长度={len(admin)}")
    if not admin:
        print("    登录失败:", r)
        return
    s, r = call("POST", "/admin/demo/seed?mode=demo", {}, admin)
    print(f"    播种HTTP: {s}")
    print("    播种返回:", json.dumps(r, ensure_ascii=False)[:300])
    # 播种可能重建用户导致旧 token 失效，重新登录管理员
    s, admin, r = login("13900003333", 3)
    print(f"    播种后重新登录管理员: {s} token长度={len(admin)}")
    s, r = call("GET", "/admin/demo/status", None, admin)
    print(f"    演示状态: {s} {str(r)[:100]}")

    # ---- 2. 甲方功能 ----
    s, client, _ = login("13900001111", 1)
    print(f"\n[2] 甲方登录: {s}")
    s, r = call("GET", "/project/mine?page=1&size=5", None, client)
    projects = r.get("projects", [])
    print(f"    我的发单: {s} total={r.get('total')} 本页={len(projects)}")
    s, r = call("GET", "/project/checklist?service_type=cost", None, client)
    print(f"    智能清单: {s} 项数={r.get('count')}")
    s, r = call("GET", "/notification/unread-count", None, client)
    print(f"    未读数: {s} {r}")
    # 项目编辑(未签约/草稿可改标题,补全必填字段)
    if projects:
        pid = projects[0]["id"]
        s, r = call("PUT", f"/project/{pid}", {
            "project_type": projects[0].get("project_type", "cost"),
            "service_type": projects[0].get("service_type", "cost"),
            "title": projects[0]["title"] + "·改",
            "description": projects[0].get("description", ""),
            "address": projects[0].get("address", ""),
            "budget_min": projects[0].get("budget_min", 1000),
            "budget_max": projects[0].get("budget_max", 5000),
        }, client)
        print(f"    项目编辑: {s} {r.get('message', r.get('error', ''))[:60]}")

    # ---- 2b. 服务超市 stats(公开接口;从智能派单取服务方 ID) ----
    if projects:
        s, rec = call("GET", f"/project/{projects[0]['id']}/recommend", None, client)
        sups = rec.get("suppliers", [])
        if sups:
            sid = sups[0]["id"]
            s, r = call("GET", f"/provider/{sid}", None, None)
            print(f"    服务超市stats: {s} provider={r.get('provider', {}).get('company_name')} match={r.get('stats')}")

    # ---- 3. 服务方功能 ----
    s, sup, _ = login("13900004444", 2)
    print(f"\n[3] 服务方登录: {s}")
    s, r = call("GET", "/bid/mine?page=1&size=5", None, sup)
    print(f"    我的报价: {s} total={r.get('total')} 本页={len(r.get('bids', []))}")
    s, r = call("GET", "/delivery-templates", None, sup)
    print(f"    交付模板: {s} 模板数={r.get('count')}")
    # 资质提交(服务方本人 13900004444;查其用户ID)
    s, me = call("GET", "/user/info", None, sup)
    sup_id = me.get("user", {}).get("id")
    print(f"    服务方ID: {sup_id}")
    if sup_id:
        s, r = call("POST", f"/supplier/{sup_id}/qualifications", {
            "qualification_type": "工程造价咨询资质",
            "certificate_no": "GC-2026-10001",
            "level": "甲级",
            "issuing_authority": "安徽省住房和城乡建设厅",
            "valid_to": "2030-12-31T00:00:00Z",
        }, sup)
        print(f"    资质提交: {s} {str(r.get('qualification', r.get('error', '')))[:80]}")
        s, r = call("GET", f"/supplier/{sup_id}/qualifications", None, sup)
        quals = r.get("qualifications", [])
        print(f"    资质列表: {s} 数量={len(quals)}")
        # 资质详情
        if quals:
            qid = quals[0]["id"]
            s, r = call("GET", f"/qualification/{qid}", None, sup)
            print(f"    资质详情: {s} 状态={r.get('qualification', {}).get('verification_status')}")

    # ---- 4. 平台方功能 ----
    s, r = call("GET", "/admin/operations-stats", None, admin)
    print(f"\n[4] 运营看板: {s} 漏斗={r.get('funnel')} 活跃7d={r.get('active_suppliers_7d')}")
    # 用户详情:从用户列表取真实 ID
    s, r = call("GET", "/admin/users", None, admin)
    users = r.get("users", [])
    uid = users[0]["id"] if users else None
    if uid:
        s, r = call("GET", f"/admin/users/{uid}", None, admin)
        print(f"    用户详情: {s} id={uid} stats={r.get('stats')}")

    # ---- 5. 智能派单 match_score(公开接口,无需登录) ----
    pid = projects[0]["id"] if projects else None
    if pid:
        s, r = call("GET", f"/project/{pid}/recommend", None, client)
        sups = r.get("suppliers", [])
        print(f"\n[5] 智能派单: {s} 推荐数={len(sups)} 首条match={sups[0].get('match_score') if sups else 'N/A'}")
    else:
        print("\n[5] 智能派单: 无项目可测(播种未生成项目)")

    # ---- 6. 文件预览(上传测试附件→preview) ----
    # 生成 1x1 PNG 测试文件
    import base64
    png = base64.b64decode("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==")
    test_path = "verify_test.png"
    with open(test_path, "wb") as f:
        f.write(png)
    s, r = upload_file(test_path, sup)
    fid = r.get("file_id")
    print(f"    附件上传: {s} file_id={fid}")
    if fid:
        # preview 返回文件二进制,用原始读取
        req = urllib.request.Request(BASE + f"/file/{fid}/preview")
        req.add_header("Authorization", "Bearer " + sup)
        try:
            with OPENER.open(req, timeout=15) as r:
                data = r.read(16)
                ctype = r.headers.get("Content-Type", "")
                print(f"    文件预览: {r.status} (file_id={fid}) type={ctype} 首字节={data[:8].hex()}")
        except urllib.error.HTTPError as e:
            print(f"    文件预览: {e.code} (file_id={fid})")

    # ---- 7. 争议三专家(创建争议→自动指派完整链路) ----
    s, r = call("GET", "/order/list", None, client)
    orders = r.get("orders", [])
    oid = orders[0]["id"] if orders else None
    print(f"    订单列表: 数量={len(orders)} 首个订单ID={oid}")
    if oid:
        s, r = call("POST", "/dispute/create", {"order_id": oid, "reason": "生产实测：验收分歧"}, client)
        print(f"    争议创建: {s} 响应={str(r)[:200]}")
        did = r.get("dispute", {}).get("id")
        if did:
            s, r = call("POST", f"/dispute/{did}/auto-expert", {}, admin)
            print(f"    争议三专家: {s} {r.get('message')} 指派={r.get('assigned')}")
    else:
        print("    争议三专家: 无订单可创建争议")

    print("\n" + "=" * 60)
    print("实测完成")


if __name__ == "__main__":
    main()
