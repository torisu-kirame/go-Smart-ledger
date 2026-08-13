from langchain_core.messages import AIMessage, HumanMessage, SystemMessage
from langchain_openai import ChatOpenAI

from .config import WORKSPACE_DIR
from .models import AgentChatRequest, ChatMessage, TestAgentRequest
from .url import InvalidBaseURL, normalize_llm_base_url


def default_agent_system_prompt() -> str:
    parts: list[str] = []
    for name in ("AGENTS.md", "TOOLS.md"):
        path = WORKSPACE_DIR / name
        if path.is_file():
            parts.append(path.read_text(encoding="utf-8"))
    return "\n\n---\n\n".join(parts).strip()


def _to_lc_messages(messages: list[ChatMessage]) -> list:
    out = []
    has_system = False
    for m in messages:
        role = (m.role or "").lower().strip()
        content = (m.content or "").strip()
        if not content:
            continue
        if role == "system":
            has_system = True
            out.append(SystemMessage(content=content))
        elif role == "assistant":
            out.append(AIMessage(content=content))
        else:
            out.append(HumanMessage(content=content))
    if not has_system:
        prompt = default_agent_system_prompt()
        if prompt:
            out.insert(0, SystemMessage(content=prompt))
    return out


def build_llm(req: AgentChatRequest | TestAgentRequest) -> ChatOpenAI:
    base = normalize_llm_base_url(req.baseUrl)
    kwargs: dict = {
        "base_url": base,
        "model": req.model.strip(),
        "timeout": 120,
    }
    key = (req.apiKey or "").strip()
    if key:
        kwargs["api_key"] = key
    else:
        kwargs["api_key"] = "no-key"
    return ChatOpenAI(**kwargs)


async def test_agent_chat(req: TestAgentRequest) -> None:
    llm = build_llm(req)
    await llm.ainvoke([HumanMessage(content="ping")])


async def chat_completion(req: AgentChatRequest):
    llm = build_llm(req)
    return await llm.ainvoke(_to_lc_messages(req.messages))


async def chat_completion_stream(req: AgentChatRequest):
    llm = build_llm(req)
    async for chunk in llm.astream(_to_lc_messages(req.messages)):
        text = chunk.content if isinstance(chunk.content, str) else ""
        if text:
            yield text


def build_agent_input(messages: list[ChatMessage]) -> str:
    max_turns = 10
    turns: list[str] = []
    for m in messages:
        role = (m.role or "").lower().strip()
        content = (m.content or "").strip()
        if role == "system" or not content:
            continue
        label = "Assistant" if role == "assistant" else "User"
        turns.append(f"{label}: {content}")
    if not turns:
        return ""
    if len(turns) > max_turns:
        turns = turns[-max_turns:]
    last = turns[-1]
    if not last.startswith("User: "):
        return last
    if len(turns) == 1:
        return last.removeprefix("User: ")
    body = "Conversation:\n" + "\n".join(turns[:-1])
    body += "\n\nCurrent question:\n" + last.removeprefix("User: ")
    return body


def build_agent_system_prefix(messages: list[ChatMessage], bound_ledger_id: str) -> str:
    parts: list[str] = []
    for m in messages:
        if (m.role or "").lower().strip() != "system":
            continue
        c = (m.content or "").strip()
        if not c:
            continue
        if "账本上下文" in c or "RAG 导出" in c:
            parts.append(
                "用户已绑定账本；请用 search_ledger_rag / get_ledger_summary 等工具查询最新链上数据。"
            )
            continue
        if len(c) > 4000:
            c = c[:4000] + "…"
        parts.append(c)
    if not parts:
        prompt = default_agent_system_prompt()
        if prompt:
            parts.append(prompt)
    lid = (bound_ledger_id or "").strip()
    if lid:
        parts.append(
            f"当前绑定账本 ID：{lid}。调用需要 ledgerId 的工具时，若用户未指定其他账本，请使用该 ID。"
        )
    parts.append(
        "你可以通过 list_ledgers、get_ledger_summary、search_ledger_rag、verify_ledger、"
        "append_ledger_entry（记流水）、get_financial_reports、get_ledger_budget 等工具操作/查询账本；"
        "不要编造链上数据。记流水时用途要具体（如「午餐-张总客户」），区分 expense/income/transfer。"
        "回复请使用 Markdown（标题、列表、表格、代码块）。"
    )
    return "\n\n".join(parts)
