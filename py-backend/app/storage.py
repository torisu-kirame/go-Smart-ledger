import json
import os
from datetime import datetime, timezone
from pathlib import Path

from .config import AGENT_CONFIG_PATH, WORKSPACE_DIR
from .models import (
    AgentStorageLoadRequest,
    AgentStorageLoadResponse,
    AgentStorageSaveRequest,
    ChatMessage,
)

CHAT_MESSAGES_FILE = "messages.json"

ALLOWED_WORKSPACE_FILES = {
    "AGENTS.md",
    "BOOTSTRAP.md",
    "HEARTBEAT.md",
    "IDENTITY.md",
    "SOUL.md",
    "TOOLS.md",
    "USER.md",
}


class AgentPathInvalid(Exception):
    pass


class AgentStorageError(Exception):
    pass


def _config_root() -> Path:
    return Path(AGENT_CONFIG_PATH)


def _resolve_safe_rel_path(base_dir: Path, rel_path: str) -> Path:
    rel_path = rel_path.strip().replace("\\", "/")
    if not rel_path or rel_path.startswith("/") or ".." in rel_path.split("/"):
        raise AgentPathInvalid(f"agent path not allowed: {rel_path}")
    full = (base_dir / rel_path).resolve()
    base = base_dir.resolve()
    if full != base and base not in full.parents:
        raise AgentPathInvalid(f"agent path not allowed: {rel_path}")
    return full


def _workspace_dir_for_agent(config_root: Path, agent_path: str) -> Path:
    normalized = agent_path.strip("/\\")
    if normalized == "agents/main":
        legacy = config_root / "workspace" / "workspace-smart-ledger"
        if legacy.is_dir():
            return legacy
    return config_root / normalized / "workspace"


def default_workspace_files() -> dict[str, str]:
    out: dict[str, str] = {}
    for name in ("AGENTS.md", "TOOLS.md"):
        path = WORKSPACE_DIR / name
        if path.is_file():
            out[name] = path.read_text(encoding="utf-8")
    return out


def load_agent_storage(req: AgentStorageLoadRequest) -> AgentStorageLoadResponse:
    root = _config_root()
    chat_dir = _resolve_safe_rel_path(root, req.chatHistoryPath)
    out = AgentStorageLoadResponse(messages=[], workspaceFiles={})
    chat_path = chat_dir / CHAT_MESSAGES_FILE
    if chat_path.is_file():
        try:
            payload = json.loads(chat_path.read_text(encoding="utf-8"))
            msgs = payload.get("messages") or []
            out.messages = [ChatMessage.model_validate(m) for m in msgs]
            out.updatedAt = payload.get("updatedAt") or ""
        except (json.JSONDecodeError, ValueError) as exc:
            raise AgentStorageError(f"read chat: {exc}") from exc
    if req.loadWorkspace:
        ws_dir = _workspace_dir_for_agent(root, req.agentPath)
        for name in ALLOWED_WORKSPACE_FILES:
            fp = ws_dir / name
            if fp.is_file():
                out.workspaceFiles[name] = fp.read_text(encoding="utf-8")
    return out


def save_agent_storage(req: AgentStorageSaveRequest) -> None:
    root = _config_root()
    if req.messages:
        chat_dir = _resolve_safe_rel_path(root, req.chatHistoryPath)
        chat_dir.mkdir(parents=True, exist_ok=True)
        payload = {
            "version": 1,
            "updatedAt": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
            "messages": [m.model_dump() for m in req.messages],
        }
        chat_path = chat_dir / CHAT_MESSAGES_FILE
        try:
            chat_path.write_text(
                json.dumps(payload, ensure_ascii=False, indent=2) + "\n",
                encoding="utf-8",
            )
        except OSError as exc:
            raise AgentStorageError(f"write chat: {exc}") from exc
    if req.workspaceFiles:
        ws_dir = _workspace_dir_for_agent(root, req.agentPath)
        ws_dir.mkdir(parents=True, exist_ok=True)
        for name, content in req.workspaceFiles.items():
            if name not in ALLOWED_WORKSPACE_FILES:
                continue
            try:
                (ws_dir / name).write_text(content, encoding="utf-8")
            except OSError as exc:
                raise AgentStorageError(f"write {name}: {exc}") from exc
