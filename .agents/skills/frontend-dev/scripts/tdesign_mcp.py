#!/usr/bin/env python3
"""TDesign MCP stdio client — manual start + call (engine MCP mount unavailable).

Usage:
  tdesign_mcp.py list-tools
  tdesign_mcp.py call <tool> '<json-args>'
  tdesign_mcp.py call get-component-docs '{"framework":"vue-next","names":["table","select"]}'

Launch order:
  1) $TDESIGN_MCP_CMD
  2) already-installed package entry (node .../tdesign-mcp-server/dist/stdio.js)
  3) npx -y tdesign-mcp-server@latest

Notes:
  - Protocol: newline-delimited JSON-RPC 2.0 over stdio (MCP).
  - Do not use sudo. Do not leave a server attached to a TTY and expect this client to attach;
    this script owns the child process pipes.
"""

from __future__ import annotations

import json
import os
import select
import shutil
import subprocess
import sys
import threading
import time
from pathlib import Path


PROTOCOL_VERSION = "2024-11-05"
DEFAULT_TIMEOUT = float(os.environ.get("TDESIGN_MCP_TIMEOUT", "90"))
CLIENT_INFO = {"name": "frontend-dev-tdesign-mcp", "version": "1.0.0"}


class McpError(RuntimeError):
    pass


def _find_installed_entry() -> list[str] | None:
    """Locate dist/stdio.js of an already-installed tdesign-mcp-server."""
    candidates: list[Path] = []
    npx_root = Path.home() / ".npm" / "_npx"
    if npx_root.is_dir():
        for p in npx_root.glob("*/node_modules/tdesign-mcp-server/dist/stdio.js"):
            candidates.append(p)
    for root in (
        Path.cwd() / "node_modules" / "tdesign-mcp-server" / "dist" / "stdio.js",
        Path("/usr/local/lib/node_modules/tdesign-mcp-server/dist/stdio.js"),
        Path("/opt/homebrew/lib/node_modules/tdesign-mcp-server/dist/stdio.js"),
    ):
        candidates.append(root)
    for p in candidates:
        if p.is_file():
            node = shutil.which("node") or "node"
            return [node, str(p)]
    return None


def resolve_command() -> list[str]:
    env = os.environ.get("TDESIGN_MCP_CMD", "").strip()
    if env:
        return env.split()
    installed = _find_installed_entry()
    if installed:
        return installed
    return ["npx", "-y", "tdesign-mcp-server@latest"]


class TDesignMcp:
    def __init__(self, cmd: list[str] | None = None, timeout: float = DEFAULT_TIMEOUT):
        self.cmd = cmd or resolve_command()
        self.timeout = timeout
        self.proc: subprocess.Popen | None = None
        self._next_id = 1
        self._stderr_lines: list[str] = []

    def __enter__(self) -> "TDesignMcp":
        self.start()
        return self

    def __exit__(self, *exc) -> None:
        self.close()

    def _drain_stderr(self) -> None:
        assert self.proc and self.proc.stderr
        for line in self.proc.stderr:
            self._stderr_lines.append(line.rstrip("\n"))
            if len(self._stderr_lines) > 50:
                self._stderr_lines.pop(0)

    def start(self) -> None:
        self.proc = subprocess.Popen(
            self.cmd,
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            bufsize=1,
        )
        threading.Thread(target=self._drain_stderr, daemon=True).start()
        resp = self._request(
            "initialize",
            {
                "protocolVersion": PROTOCOL_VERSION,
                "capabilities": {},
                "clientInfo": CLIENT_INFO,
            },
        )
        if "error" in resp:
            raise McpError(f"initialize failed: {resp['error']}\nstderr:\n" + "\n".join(self._stderr_lines[-10:]))
        self._notify("notifications/initialized")

    def _send(self, payload: dict) -> None:
        assert self.proc and self.proc.stdin
        self.proc.stdin.write(json.dumps(payload, ensure_ascii=False) + "\n")
        self.proc.stdin.flush()

    def _notify(self, method: str) -> None:
        self._send({"jsonrpc": "2.0", "method": method})

    def _request(self, method: str, params: dict | None = None) -> dict:
        rid = self._next_id
        self._next_id += 1
        payload: dict = {"jsonrpc": "2.0", "id": rid, "method": method}
        if params is not None:
            payload["params"] = params
        self._send(payload)
        return self._wait_result(rid)

    def _wait_result(self, expected_id: int) -> dict:
        assert self.proc and self.proc.stdout
        deadline = time.time() + self.timeout
        stdout = self.proc.stdout
        fd = stdout.fileno()
        while time.time() < deadline:
            remaining = max(0.05, deadline - time.time())
            if self.proc.poll() is not None:
                # process died — try to read any remaining output once
                line = stdout.readline()
                if line and line.strip().startswith("{"):
                    msg = json.loads(line)
                    if msg.get("id") == expected_id:
                        return msg
                raise McpError(
                    f"server exited early (code={self.proc.returncode}); stderr:\n"
                    + "\n".join(self._stderr_lines[-15:])
                )
            ready, _, _ = select.select([fd], [], [], remaining)
            if not ready:
                continue
            line = stdout.readline()
            if not line:
                continue
            line = line.strip()
            if not line or not line.startswith("{"):
                continue
            try:
                msg = json.loads(line)
            except json.JSONDecodeError:
                continue
            if msg.get("id") == expected_id:
                return msg
        raise McpError(
            f"timeout waiting for response id={expected_id}; cmd={self.cmd}; stderr:\n"
            + "\n".join(self._stderr_lines[-15:])
        )

    def list_tools(self) -> list[dict]:
        resp = self._request("tools/list")
        if "error" in resp:
            raise McpError(str(resp["error"]))
        return resp.get("result", {}).get("tools", [])

    def call_tool(self, name: str, arguments: dict) -> dict:
        resp = self._request("tools/call", {"name": name, "arguments": arguments})
        if "error" in resp:
            raise McpError(str(resp["error"]))
        return resp.get("result", resp)

    def close(self) -> None:
        if not self.proc:
            return
        try:
            if self.proc.stdin:
                self.proc.stdin.close()
        except Exception:
            pass
        try:
            self.proc.terminate()
            self.proc.wait(timeout=3)
        except Exception:
            try:
                self.proc.kill()
            except Exception:
                pass
        self.proc = None


def _print_result(result: dict) -> int:
    content = result.get("content")
    if isinstance(content, list):
        for item in content:
            if isinstance(item, dict) and item.get("type") == "text":
                text = item.get("text", "")
                try:
                    print(json.dumps(json.loads(text), ensure_ascii=False, indent=2))
                except Exception:
                    print(text)
            else:
                print(json.dumps(item, ensure_ascii=False, indent=2))
        if result.get("isError"):
            print("[isError=true]", file=sys.stderr)
            return 1
        return 0
    print(json.dumps(result, ensure_ascii=False, indent=2))
    return 0


def main(argv: list[str]) -> int:
    if len(argv) < 2 or argv[1] in {"-h", "--help", "help"}:
        print(__doc__.strip())
        return 2
    action = argv[1]
    cmd = resolve_command()
    print(f"# launch: {' '.join(cmd)}", file=sys.stderr)
    with TDesignMcp(cmd=cmd) as client:
        if action == "list-tools":
            for t in client.list_tools():
                name = t.get("name")
                desc = (t.get("description") or "").split("\n")[0]
                schema = t.get("inputSchema") or {}
                props = sorted((schema.get("properties") or {}).keys())
                req = schema.get("required") or []
                print(f"- {name}: {desc}")
                print(f"  params: {', '.join(props) or '(none)'}; required: {', '.join(req) or '(none)'}")
            return 0
        if action == "call":
            if len(argv) < 3:
                print("usage: tdesign_mcp.py call <tool> '<json-args>'", file=sys.stderr)
                return 2
            tool = argv[2]
            raw = argv[3] if len(argv) > 3 else "{}"
            try:
                args = json.loads(raw) if raw.strip() else {}
            except json.JSONDecodeError as e:
                print(f"invalid JSON args: {e}", file=sys.stderr)
                return 2
            return _print_result(client.call_tool(tool, args))
        print(f"unknown action: {action}", file=sys.stderr)
        return 2


if __name__ == "__main__":
    try:
        raise SystemExit(main(sys.argv))
    except McpError as e:
        print(f"MCP error: {e}", file=sys.stderr)
        raise SystemExit(1)
