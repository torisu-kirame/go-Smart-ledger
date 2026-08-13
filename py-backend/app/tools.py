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
        headers: dict[str, str] = {"Content-Type": "application/json"}
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

    async def _post(self, path: str, body: dict[str, Any]) -> Any:
        async with httpx.AsyncClient(timeout=120.0) as client:
            resp = await client.post(
                f"{self.base}{path}", headers=self._headers(), json=body
            )
            resp.raise_for_status()
            if resp.status_code == 204 or not resp.content:
                return {"ok": True}
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
                    "bookkeepingMode": m.get("bookkeepingMode"),
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
            "bookkeepingMode": m.get("bookkeepingMode"),
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

    async def append_entry(
        self,
        ledger_id: str,
        *,
        amount: str,
        note: str = "",
        category: str = "",
        entry_type: str = "expense",
        date: str = "",
        table_id: str = "",
    ) -> dict[str, Any]:
        signer = self.user_id or ""
        body = {
            "entry": {
                "signerId": signer,
                "tableId": table_id or "",
                "amount": str(amount).strip(),
                "note": (note or "").strip(),
                "category": (category or "").strip(),
                "type": (entry_type or "expense").strip() or "expense",
                "date": (date or "").strip(),
            }
        }
        data = await self._post(f"/api/v1/ledgers/{ledger_id}/entries", body)
        return {
            "ok": True,
            "ledgerId": ledger_id,
            "seq": data.get("seq"),
            "hash": data.get("hash"),
            "type": data.get("type"),
            "createdAt": data.get("createdAt"),
        }

    async def get_accounting_reports(self, ledger_id: str, period: str = "") -> Any:
        q = f"?period={period}" if period else ""
        return await self._get(f"/api/v1/ledgers/{ledger_id}/accounting/reports{q}")

    async def get_budget(self, ledger_id: str, period: str) -> Any:
        return await self._get(
            f"/api/v1/ledgers/{ledger_id}/accounting/budget?period={period}"
        )

    async def get_budget_analysis(self, ledger_id: str, period: str) -> Any:
        return await self._get(
            f"/api/v1/ledgers/{ledger_id}/accounting/budget/analysis?period={period}"
        )


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


def _truncate_json(obj: Any, limit: int = 16000) -> str:
    s = json.dumps(obj, ensure_ascii=False, default=str)
    if len(s) > limit:
        return s[:limit] + "...(truncated)"
    return s


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

    class AppendArgs(BaseModel):
        ledgerId: str = Field(default="", description="Ledger ID; uses bound ledger when empty")
        amount: str = Field(description="Amount as string, e.g. 200 or -50")
        note: str = Field(
            default="",
            description="Purpose detail, e.g. 午餐-张总客户 (be specific)",
        )
        category: str = Field(
            default="",
            description="Category tag, e.g. 餐饮/工作餐",
        )
        entryType: str = Field(
            default="expense",
            description="expense | income | transfer (transfer is not expense)",
        )
        date: str = Field(default="", description="YYYY-MM-DD optional")
        tableId: str = Field(default="", description="Optional multi-table id")

    async def append_ledger_entry(
        ledgerId: str = "",
        amount: str = "",
        note: str = "",
        category: str = "",
        entryType: str = "expense",
        date: str = "",
        tableId: str = "",
    ) -> str:
        try:
            if not str(amount).strip():
                return "error: amount required"
            lid = _resolve_ledger_id(ledgerId, default_ledger_id)
            out = await client.append_entry(
                lid,
                amount=amount,
                note=note,
                category=category,
                entry_type=entryType,
                date=date,
                table_id=tableId,
            )
            return json.dumps(out, ensure_ascii=False)
        except Exception as exc:
            return f"error: {exc}"

    class ReportArgs(BaseModel):
        ledgerId: str = Field(default="", description="Ledger ID")
        period: str = Field(default="", description="Optional period e.g. 2026-05")

    async def get_financial_reports(ledgerId: str = "", period: str = "") -> str:
        try:
            lid = _resolve_ledger_id(ledgerId, default_ledger_id)
            out = await client.get_accounting_reports(lid, period=period)
            return _truncate_json(out)
        except Exception as exc:
            return f"error: {exc}"

    class BudgetArgs(BaseModel):
        ledgerId: str = Field(default="", description="Ledger ID")
        period: str = Field(description="Budget period e.g. 2026-05")
        analysis: bool = Field(
            default=True, description="If true, return budget analysis vs spend"
        )

    async def get_ledger_budget(
        ledgerId: str = "", period: str = "", analysis: bool = True
    ) -> str:
        try:
            lid = _resolve_ledger_id(ledgerId, default_ledger_id)
            if not (period or "").strip():
                return "error: period required (YYYY-MM)"
            if analysis:
                out = await client.get_budget_analysis(lid, period.strip())
            else:
                out = await client.get_budget(lid, period.strip())
            return _truncate_json(out)
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
        StructuredTool.from_function(
            coroutine=append_ledger_entry,
            name="append_ledger_entry",
            description=(
                "Append a cash-flow entry (记流水). Prefer specific notes like "
                '"午餐-张总客户". Distinguish expense/income/transfer. '
                'JSON: {"amount":"200","note":"...","category":"餐饮","entryType":"expense","date":"YYYY-MM-DD"}.'
            ),
            args_schema=AppendArgs,
        ),
        StructuredTool.from_function(
            coroutine=get_financial_reports,
            name="get_financial_reports",
            description='Get accounting financial reports for a professional ledger. JSON: {"ledgerId":"...","period":"2026-05"}.',
            args_schema=ReportArgs,
        ),
        StructuredTool.from_function(
            coroutine=get_ledger_budget,
            name="get_ledger_budget",
            description='Get budget or budget-vs-spend analysis. JSON: {"period":"2026-05","analysis":true}.',
            args_schema=BudgetArgs,
        ),
    ]
