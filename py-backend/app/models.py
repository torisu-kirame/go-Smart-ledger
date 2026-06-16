from typing import Any

from pydantic import BaseModel, Field


class ChatMessage(BaseModel):
    role: str
    content: str


class AgentChatRequest(BaseModel):
    baseUrl: str = ""
    apiKey: str = ""
    model: str = ""
    messages: list[ChatMessage] = Field(default_factory=list)
    stream: bool = False
    useTools: bool = False
    boundLedgerId: str = ""


class TestAgentRequest(BaseModel):
    baseUrl: str = ""
    apiKey: str = ""
    model: str = ""


class AgentStoragePaths(BaseModel):
    agentPath: str
    chatHistoryPath: str


class AgentStorageLoadRequest(AgentStoragePaths):
    loadWorkspace: bool = False


class AgentStorageSaveRequest(AgentStoragePaths):
    messages: list[ChatMessage] = Field(default_factory=list)
    workspaceFiles: dict[str, str] = Field(default_factory=dict)


class AgentStorageLoadResponse(BaseModel):
    messages: list[ChatMessage] = Field(default_factory=list)
    workspaceFiles: dict[str, str] = Field(default_factory=dict)
    updatedAt: str = ""


def parse_agent_chat_request(raw: dict[str, Any]) -> AgentChatRequest:
    req = AgentChatRequest.model_validate(raw)
    if not req.baseUrl.strip():
        legacy_url = (raw.get("gatewayUrl") or "").strip()
        if legacy_url:
            raise ValueError(
                "openclaw gateway is deprecated; configure provider baseUrl/apiKey/model in Settings → AI"
            )
    return req


def parse_test_agent_request(raw: dict[str, Any]) -> TestAgentRequest:
    req = TestAgentRequest.model_validate(raw)
    if not req.baseUrl.strip():
        legacy_url = (raw.get("gatewayUrl") or "").strip()
        if legacy_url:
            raise ValueError(
                "openclaw gateway is deprecated; use provider baseUrl/apiKey/model"
            )
    return req
