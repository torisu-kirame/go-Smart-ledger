import json
from typing import Any, AsyncIterator

from fastapi import FastAPI, HTTPException, Request
from fastapi.responses import JSONResponse, StreamingResponse

from . import agent as agent_runner
from . import llm
from .models import (
    AgentStorageLoadRequest,
    AgentStorageSaveRequest,
    AgentChatRequest,
    parse_agent_chat_request,
    parse_test_agent_request,
)
from .sse import chunk_text, sse_error_response, sse_stream
from .storage import (
    AgentPathInvalid,
    AgentStorageError,
    default_workspace_files,
    load_agent_storage,
    save_agent_storage,
)
from .tools import LedgerClient
from .url import InvalidBaseURL

app = FastAPI(title="Smart Ledger AI API", version="1.0.0")


def _user_id(request: Request) -> str:
    return (request.headers.get("X-User-Id") or "").strip()


def _auth_header(request: Request) -> str:
    return (request.headers.get("Authorization") or "").strip()


@app.get("/api/v1/health")
async def health() -> dict[str, bool]:
    return {"ok": True}


@app.post("/api/v1/ai/chat")
async def ai_chat(request: Request) -> Any:
    uid = _user_id(request)
    if not uid:
        raise HTTPException(status_code=401, detail="unauthorized")
    try:
        raw = await request.json()
        req = parse_agent_chat_request(raw)
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    except Exception as exc:
        raise HTTPException(status_code=400, detail="invalid body") from exc

    if not req.baseUrl.strip() or not req.model.strip():
        raise HTTPException(status_code=400, detail="baseUrl and model required")
    if not req.messages:
        raise HTTPException(status_code=400, detail="messages required")

    try:
        if req.useTools and uid:
            return await _chat_with_tools(request, req, uid)
        if req.stream:
            return await _chat_stream(req)
        return await _chat_json(req)
    except InvalidBaseURL as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    except Exception as exc:
        raise HTTPException(status_code=502, detail=str(exc)) from exc


async def _chat_json(req: AgentChatRequest) -> JSONResponse:
    resp = await llm.chat_completion(req)
    content = resp.content if isinstance(resp.content, str) else str(resp.content or "")
    return JSONResponse(
        {
            "choices": [
                {
                    "index": 0,
                    "message": {"role": "assistant", "content": content},
                    "finish_reason": "stop",
                }
            ]
        }
    )


async def _chat_stream(req: AgentChatRequest) -> StreamingResponse:
    async def gen() -> AsyncIterator[str]:
        started = False
        try:
            async for delta in llm.chat_completion_stream(req):
                if not delta.strip():
                    continue
                started = True
                payload = json.dumps(
                    {"choices": [{"index": 0, "delta": {"content": delta}}]},
                    ensure_ascii=False,
                )
                yield f"data: {payload}\n\n"
            if not started:
                raise ValueError("LLM returned empty streaming response")
            yield "data: [DONE]\n\n"
        except Exception as exc:
            if not started:
                raise
            payload = json.dumps({"error": {"message": str(exc)}}, ensure_ascii=False)
            yield f"data: {payload}\n\n"
            yield "data: [DONE]\n\n"

    return StreamingResponse(
        gen(),
        media_type="text/event-stream",
        headers={"Cache-Control": "no-cache", "Connection": "keep-alive"},
    )


async def _chat_with_tools(request: Request, req: AgentChatRequest, uid: str) -> Any:
    user_input = llm.build_agent_input(req.messages)
    if not user_input:
        raise HTTPException(status_code=400, detail="no user message for agent")
    prefix = llm.build_agent_system_prefix(req.messages, req.boundLedgerId)
    client = LedgerClient(_auth_header(request), uid)
    model = llm.build_llm(req)
    content = await agent_runner.run_ledger_agent(
        model, client, req.boundLedgerId, user_input, prefix
    )
    if req.stream:

        async def gen() -> AsyncIterator[str]:
            for part in chunk_text(content):
                payload = json.dumps(
                    {"choices": [{"index": 0, "delta": {"content": part}}]},
                    ensure_ascii=False,
                )
                yield f"data: {payload}\n\n"
            yield "data: [DONE]\n\n"

        return StreamingResponse(
            gen(),
            media_type="text/event-stream",
            headers={"Cache-Control": "no-cache", "Connection": "keep-alive"},
        )
    return JSONResponse(
        {
            "choices": [
                {
                    "index": 0,
                    "message": {"role": "assistant", "content": content},
                    "finish_reason": "stop",
                }
            ]
        }
    )


@app.post("/api/v1/ai/test")
async def ai_test(request: Request) -> dict[str, bool]:
    if not _user_id(request):
        raise HTTPException(status_code=401, detail="unauthorized")
    try:
        raw = await request.json()
        req = parse_test_agent_request(raw)
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    except Exception as exc:
        raise HTTPException(status_code=400, detail="invalid body") from exc
    if not req.baseUrl.strip() or not req.model.strip():
        raise HTTPException(status_code=400, detail="baseUrl and model required")
    try:
        await llm.test_agent_chat(req)
    except InvalidBaseURL as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    except Exception as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    return {"ok": True}


@app.post("/api/v1/ai/agent/load")
async def ai_agent_load(request: Request) -> Any:
    if not _user_id(request):
        raise HTTPException(status_code=401, detail="unauthorized")
    try:
        raw = await request.json()
        req = AgentStorageLoadRequest.model_validate(raw)
    except Exception as exc:
        raise HTTPException(status_code=400, detail="invalid json") from exc
    if not req.agentPath.strip() or not req.chatHistoryPath.strip():
        raise HTTPException(status_code=400, detail="agentPath and chatHistoryPath required")
    try:
        out = load_agent_storage(req)
    except AgentPathInvalid as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    except AgentStorageError as exc:
        raise HTTPException(status_code=500, detail=str(exc)) from exc
    if req.loadWorkspace and not out.workspaceFiles:
        out.workspaceFiles = default_workspace_files()
    return out


@app.post("/api/v1/ai/agent/save")
async def ai_agent_save(request: Request) -> dict[str, bool]:
    if not _user_id(request):
        raise HTTPException(status_code=401, detail="unauthorized")
    try:
        raw = await request.json()
        req = AgentStorageSaveRequest.model_validate(raw)
    except Exception as exc:
        raise HTTPException(status_code=400, detail="invalid json") from exc
    if not req.agentPath.strip() or not req.chatHistoryPath.strip():
        raise HTTPException(status_code=400, detail="agentPath and chatHistoryPath required")
    if not req.messages and not req.workspaceFiles:
        raise HTTPException(status_code=400, detail="nothing to save")
    try:
        save_agent_storage(req)
    except AgentPathInvalid as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    except AgentStorageError as exc:
        raise HTTPException(status_code=500, detail=str(exc)) from exc
    return {"ok": True}
