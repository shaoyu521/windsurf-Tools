#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
批量走 _quick_key.py 同款链路：Firebase 登录 → Codeium register_user → GetUserJwt，
得到 Windsurf 用的 JWT（非 Firebase idToken），供应用内「JWT」批量导入。

输入：与 normalize 脚本一致，每行
  账号: email@x.com 密码: xxx [ 日期 ] [ 密码不对就这个 备用 ]

用法:
  python tools/batch_quick_jwt.py tools/normalized-accounts-output.txt -o tools/jwt-windsurf.txt
  python tools/batch_quick_jwt.py accounts.txt -o out.txt -j 6

仅依赖 Python 3 标准库（HTTPS + gRPC-over-HTTP 帧，与 backend/services/windsurf.go 对齐）。
"""
from __future__ import annotations

import argparse
import http.client
import json
import re
import ssl
import sys
import threading
from concurrent.futures import ThreadPoolExecutor, as_completed
from typing import List, Optional, Tuple
from urllib.parse import urlparse

# 与 backend/services/windsurf.go 一致
FK = "AIzaSyDsOl-1XpT5err0Tcnx8FFod1H8gVGIycY"
WINDSURF_APP = "windsurf"
WINDSURF_VERSION = "1.48.2"
WINDSURF_CLIENT = "1.9566.11"
GRPC_UPSTREAM_HOST = "server.self-serve.windsurf.com"
GRPC_UPSTREAM_IP = "34.49.14.144"
GRPC_JWT_PATH = "/exa.auth_pb.AuthService/GetUserJwt"

FIREBASE_URL = f"https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key={FK}"
REGISTER_URL = "https://api.codeium.com/register_user/"


def _stdout_utf8() -> None:
    try:
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
        sys.stderr.reconfigure(encoding="utf-8", errors="replace")
    except AttributeError:
        pass


def ssl_ctx_insecure() -> ssl.SSLContext:
    ctx = ssl.create_default_context()
    ctx.check_hostname = False
    ctx.verify_mode = ssl.CERT_NONE
    return ctx


def encode_varint(n: int) -> bytes:
    buf = bytearray()
    while n >= 0x80:
        buf.append((n & 0x7F) | 0x80)
        n >>= 7
    buf.append(n & 0x7F)
    return bytes(buf)


def encode_string_field(field_num: int, value: str) -> bytes:
    b = value.encode("utf-8")
    tag = (field_num << 3) | 2
    return bytes([tag]) + encode_varint(len(b)) + b


def build_api_key_metadata(api_key: str) -> bytes:
    return (
        encode_string_field(1, WINDSURF_APP)
        + encode_string_field(2, WINDSURF_VERSION)
        + encode_string_field(3, api_key)
        + encode_string_field(4, "en")
        + encode_string_field(5, "windows")
        + encode_string_field(7, WINDSURF_CLIENT)
        + encode_string_field(12, WINDSURF_APP)
    )


def build_grpc_envelope(message: bytes) -> bytes:
    inner = bytes([0x0A]) + encode_varint(len(message)) + message
    return bytes([0x00]) + len(inner).to_bytes(4, "big") + inner


def read_varint(data: bytes, pos: int) -> Tuple[int, int]:
    result = 0
    shift = 0
    while pos < len(data):
        b = data[pos]
        pos += 1
        result |= (b & 0x7F) << shift
        shift += 7
        if (b & 0x80) == 0:
            return result, pos
        if shift >= 64:
            break
    return 0, pos


def find_jwt_in_protobuf(data: bytes) -> Optional[str]:
    """移植自 backend/utils/proto.go FindJWTInProtobuf"""
    pos = 0
    while pos < len(data):
        tag, pos = read_varint(data, pos)
        if pos > len(data):
            break
        wire = tag & 7
        if wire == 0:
            _, pos = read_varint(data, pos)
        elif wire == 2:
            ln, pos = read_varint(data, pos)
            if pos + ln > len(data):
                return None
            field_data = data[pos : pos + ln]
            pos += ln
            try:
                s = field_data.decode("utf-8", errors="strict")
            except UnicodeDecodeError:
                s = ""
                nested = find_jwt_in_protobuf(field_data)
                if nested:
                    return nested
                continue
            if len(s) > 100 and s.startswith("eyJ"):
                return s
            if ln > 2:
                nested = find_jwt_in_protobuf(field_data)
                if nested:
                    return nested
        elif wire == 5:
            pos += 4
        elif wire == 1:
            pos += 8
        else:
            return None
    return None


def http_json_post(url: str, payload: dict, ctx: ssl.SSLContext, timeout: float = 30) -> dict:
    u = urlparse(url)
    host = u.netloc
    path = u.path or "/"
    if u.query:
        path = f"{path}?{u.query}"
    body = json.dumps(payload).encode("utf-8")
    conn = http.client.HTTPSConnection(host, 443, context=ctx, timeout=timeout)
    try:
        conn.request("POST", path, body=body, headers={"Content-Type": "application/json"})
        resp = conn.getresponse()
        raw = resp.read()
        if resp.status != 200:
            raise RuntimeError(f"HTTP {resp.status}: {raw[:400]!r}")
        return json.loads(raw.decode())
    finally:
        conn.close()


def firebase_sign_in(email: str, password: str, ctx: ssl.SSLContext) -> Tuple[str, str]:
    data = http_json_post(
        FIREBASE_URL,
        {
            "email": email,
            "password": password,
            "returnSecureToken": True,
            "clientType": "CLIENT_TYPE_WEB",
        },
        ctx,
    )
    return data["idToken"], data["refreshToken"]


def register_user(id_token: str, ctx: ssl.SSLContext) -> Tuple[str, str]:
    data = http_json_post(
        REGISTER_URL, {"firebase_id_token": id_token}, ctx
    )
    api_key = data.get("api_key") or ""
    if not api_key:
        raise RuntimeError("register_user 未返回 api_key")
    name = data.get("name") or ""
    return api_key, name


def get_user_jwt_grpc(api_key: str, ctx: ssl.SSLContext) -> str:
    meta = build_api_key_metadata(api_key)
    envelope = build_grpc_envelope(meta)
    conn = http.client.HTTPSConnection(
        GRPC_UPSTREAM_IP, 443, context=ctx, timeout=30
    )
    try:
        conn.request(
            "POST",
            GRPC_JWT_PATH,
            body=envelope,
            headers={
                "Content-Type": "application/grpc",
                "Authorization": api_key,
                "Host": GRPC_UPSTREAM_HOST,
            },
        )
        resp = conn.getresponse()
        raw = resp.read()
        if resp.status != 200:
            gs = resp.getheader("grpc-status", "")
            gm = resp.getheader("grpc-message", "")
            raise RuntimeError(f"gRPC HTTP {resp.status} status={gs} msg={gm} body={raw[:200]!r}")
        if len(raw) < 6:
            raise RuntimeError(f"响应体过短 {len(raw)}")
        payload = raw[5:]
        jwt = find_jwt_in_protobuf(payload)
        if not jwt:
            raise RuntimeError("响应中未解析出 JWT")
        return jwt
    finally:
        conn.close()


def parse_account_line(line: str) -> Optional[Tuple[str, List[str]]]:
    s = line.strip()
    if not s or s.startswith("#"):
        return None
    m = re.match(r"^账号:\s*(\S+@\S+)\s*密码:\s*(.+)$", s, re.I)
    if not m:
        return None
    rest = m.group(2).strip()
    alt = ""
    alt_m = re.search(r"\s+密码不对就这个\s*(.+)$", rest, re.I)
    if alt_m:
        alt = alt_m.group(1).strip()
        rest = rest[: alt_m.start()].strip()
    rest = re.sub(r"\s+\d{4}/\d{1,2}/\d{1,2}\s*$", "", rest).strip()
    pws = [rest] + ([alt] if alt else [])
    pws = [p for p in pws if p]
    if not pws:
        return None
    return m.group(1), pws


def pipeline_one(
    email: str, passwords: List[str], ctx: ssl.SSLContext
) -> Tuple[str, str, str, str]:
    """返回 (jwt, email, remark_suffix, refresh_token)"""
    last_err: Optional[Exception] = None
    for i, pw in enumerate(passwords):
        try:
            id_token, refresh = firebase_sign_in(email, pw, ctx)
            api_key, _name = register_user(id_token, ctx)
            jwt = get_user_jwt_grpc(api_key, ctx)
            suffix = " (备用密码)" if i > 0 else ""
            return jwt, email, suffix, refresh
        except Exception as e:
            last_err = e
    raise last_err or RuntimeError("登录失败")


_thread_local = threading.local()


def get_ctx() -> ssl.SSLContext:
    c = getattr(_thread_local, "ctx", None)
    if c is None:
        c = ssl_ctx_insecure()
        _thread_local.ctx = c
    return c


def main() -> int:
    _stdout_utf8()
    ap = argparse.ArgumentParser(description="批量 Firebase → RegisterUser → GetUserJwt")
    ap.add_argument("input", help="账号密码文件（每行 账号:… 密码:…）")
    ap.add_argument("-o", "--output", required=True, help="输出：每行「JWT 备注」")
    ap.add_argument(
        "-j",
        "--jobs",
        type=int,
        default=4,
        help="并发数（默认 4，避免触发限流）",
    )
    ap.add_argument(
        "--write-refresh",
        metavar="PATH",
        help="可选：同时写入 refresh_token 行（邮箱密码登录得到的 refresh，供 Refresh Token 导入）",
    )
    args = ap.parse_args()
    jobs = max(1, min(16, args.jobs))

    with open(args.input, encoding="utf-8", errors="replace") as f:
        rows: List[Tuple[str, List[str]]] = []
        for line in f:
            p = parse_account_line(line)
            if p:
                rows.append(p)

    if not rows:
        print("未解析到任何「账号: … 密码: …」行", file=sys.stderr)
        return 1

    results_ok: List[Tuple[str, str, str, str]] = []  # jwt, email, suffix, refresh
    results_fail: List[Tuple[str, str]] = []
    lock = threading.Lock()

    def task(item: Tuple[str, List[str]]) -> None:
        email, passwords = item
        try:
            jwt, em, suffix, refresh = pipeline_one(email, passwords, get_ctx())
            with lock:
                results_ok.append((jwt, em, suffix, refresh))
        except Exception as e:
            with lock:
                results_fail.append((email, str(e)))

    print(f"# 共 {len(rows)} 条，并发 {jobs} …", file=sys.stderr)
    with ThreadPoolExecutor(max_workers=jobs) as ex:
        futs = [ex.submit(task, r) for r in rows]
        done = 0
        for _ in as_completed(futs):
            done += 1
            if done % 10 == 0 or done == len(rows):
                print(f"# 进度 {done}/{len(rows)}", file=sys.stderr)

    results_ok.sort(key=lambda x: x[1].lower())

    refresh_lines: List[str] = []
    with open(args.output, "w", encoding="utf-8") as out:
        for jwt, email, suffix, refresh in results_ok:
            remark = f"{email}{suffix}"
            out.write(f"{jwt} {remark}\n")
            refresh_lines.append(f"{refresh} {remark}")

    if args.write_refresh:
        with open(args.write_refresh, "w", encoding="utf-8") as rf:
            rf.write("\n".join(refresh_lines) + ("\n" if refresh_lines else ""))

    print(f"\n# 成功 {len(results_ok)} / {len(rows)}，失败 {len(results_fail)}", file=sys.stderr)
    for em, err in results_fail:
        print(f"# FAIL {em}: {err}", file=sys.stderr)
    if args.write_refresh:
        print(f"# 已写 refresh: {args.write_refresh}", file=sys.stderr)
    return 1 if results_fail else 0


if __name__ == "__main__":
    raise SystemExit(main())
