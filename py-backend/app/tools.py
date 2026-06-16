import json
from typing import Any

import httpx
from langchain_core.tools import StructuredTool
from pydantic import BaseModel, Field

from .config import LEDGER_API_URL


class LedgerClient:
    def __init__(self, auth_header: str, user_id: str):
        self.auth_header = auth_header
        self.user_id = user_id
        self.base = LEDGER_API_URL

    def _headers(self) -> dict[str, str]:
        headers: dict[str, str] = {}
        if self.auth_header:
            headers["Authorization"] = self.auth_header
        if self.user_id:
            headers["X-User-Id"] = self.user_id
        return headers

    async def _get(self, path: str) -> Any:
        async with httpx.AsyncClient(timeout=120.0) as client:
            resp = await client.get(f"{self.base}{path}", headers=self._headers())
            resp.raise_for_status()
            return resp.json()

    async def list_ledgers(self) -> list[dict[str, Any]]:
        data = await self._get("/api/v1/ledgers")
        items = data if isinstance(data, list) else data.get("ledgers") or data.get("items") or []
        out = []
        for m in items:
            out.append(
                {
                    "id": m.get("id"),
                    "name": m.get("name"),
                    "type": m.get("type"),
                    "latestSeq": m.get("latestSeq"),
                    "anchorStatus": m.get("anchorStatus"),
                }
            )
        return out

    async def get_ledger_summary(self, ledger_id: str) -> dict[str, Any]:
        m = await self._get(f"/api/v1/ledgers/{ledger_id}")
        return {
            "id": m.get("id"),
            "name": m.get("name"),
            "type": m.get("type"),
            "latestSeq": m.get("latestSeq"),
            "latestRoot": m.get("latestRoot"),
            "anchorStatus": m.get("anchorStatus"),
            "memberCount": len(m.get("members") or []),
            "createdAt": m.get("createdAt"),
            "updatedAt": m.get("updatedAt"),
        }

    async def export_rag(
        self, ledger_id: str, limit: int = 40, query: str = ""
    ) -> dict[str, Any]:
        export = await self._get(f"/api/v1/ledgers/{ledger_id}/rag-export")
        chunks = export.get("chunks") or []
        q = query.lower().strip()
        if q:
            chunks = [c for c in chunks if q in (c.get("text") or "").lower()]
        if limit <= 0:
            limit = 40
        if limit > 120:
            limit = 120
        chunks = chunks[:limit]
        rows = [
            {"seq": c.get("seq"), "type": c.get("type"), "text": c.get("text")}
            for c in chunks
        ]
        out = {
            "ledgerId": export.get("ledgerId"),
            "ledgerName": export.get("ledgerName"),
            "total": len(export.get("chunks") or []),
            "returned": len(rows),
            "chunks": rows,
        }
        s = json.dumps(out, ensure_ascii=False)
        if len(s) > 16000:
            s = s[:16000] + "...(truncated)"
            return json.loads(s)
        return out

    async def verify_ledger(self, ledger_id: str) -> dict[str, Any]:
        await self._get(f"/api/v1/ledgers/{ledger_id}")
        data = await self._get(f"/api/v1/ledgers/{ledger_id}/verify")
        valid = data if isinstance(data, bool) else data.get("valid", data.get("ok"))
        return {"ledgerId": ledger_id, "valid": bool(valid)}


def _resolve_ledger_id(raw: str, default_id: str) -> str:
    raw = (raw or "").strip()
    if raw:
        if raw.startswith("{"):
            try:
                obj = json.loads(raw)
                lid = (obj.get("ledgerId") or "").strip()
                if lid:
                    return lid
            except json.JSONDecodeError:
                pass
        else:
            return raw
    if default_id:
        return default_id
    raise ValueError("ledgerId required")


def build_ledger_tools(client: LedgerClient, default_ledger_id: str) -> list[StructuredTool]:
    async def list_ledgers(_input: str = "") -> str:
        try:
            rows = await client.list_ledgers()
            return json.dumps(rows, ensure_ascii=False)
        except Exception as exc:
            return f"error: {exc}"

    class SummaryArgs(BaseModel):
        ledgerId: str = Field(default="", description="Ledger ID; uses bound ledger when empty")

    async def get_ledger_summary(ledgerId: str = "") -> str:
        try:
            lid = _resolve_ledger_id(ledgerId, default_ledger_id)
            summary = await client.get_ledger_summary(lid)
            return json.dumps(summary, ensure_ascii=False)
        except Exception as exc:
            return f"error: {exc}"

    class RagArgs(BaseModel):
        ledgerId: str = Field(default="", description="Ledger ID")
        limit: int = Field(default=40, description="Max chunks")
        query: str = Field(default="", description="Optional keyword filter")

    async def search_ledger_rag(
        ledgerId: str = "", limit: int = 40, query: str = ""
    ) -> str:
        try:
            lid = _resolve_ledger_id(ledgerId, default_ledger_id)
            out = await client.export_rag(lid, limit=limit, query=query)
            return json.dumps(out, ensure_ascii=False)
        except Exception as exc:
            return f"error: {exc}"

    class VerifyArgs(BaseModel):
        ledgerId: str = Field(default="", description="Ledger ID")

    async def verify_ledger(ledgerId: str = "") -> str:
        try:
            lid = _resolve_ledger_id(ledgerId, default_ledger_id)
            out = await client.verify_ledger(lid)
            return json.dumps(out, ensure_ascii=False)
        except Exception as exc:
            return f"error: {exc}"

    return [
        StructuredTool.from_function(
            coroutine=list_ledgers,
            name="list_ledgers",
            description="List ledgers the current user can access. Input is optional JSON {} or empty.",
        ),
        StructuredTool.from_function(
            coroutine=get_ledger_summary,
            name="get_ledger_summary",
            description='Get metadata for one ledger. Input JSON: {"ledgerId":"..."} or plain ledger id.',
            args_schema=SummaryArgs,
        ),
        StructuredTool.from_function(
            coroutine=search_ledger_rag,
            name="search_ledger_rag",
            description='Export ledger events as text chunks. Input JSON: {"ledgerId":"...","limit":40,"query":"keyword"}.',
            args_schema=RagArgs,
        ),
        StructuredTool.from_function(
            coroutine=verify_ledger,
            name="verify_ledger",
            description='Verify Merkle integrity. Input JSON: {"ledgerId":"..."} or plain ledger id.',
            args_schema=VerifyArgs,
        ),
    ]
