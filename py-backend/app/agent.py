from langchain.agents import AgentExecutor, create_tool_calling_agent
from langchain_core.prompts import ChatPromptTemplate, MessagesPlaceholder
from langchain_openai import ChatOpenAI

from .tools import LedgerClient, build_ledger_tools

AGENT_MAX_ITERATIONS = 6


async def run_ledger_agent(
    llm: ChatOpenAI,
    client: LedgerClient,
    bound_ledger_id: str,
    user_input: str,
    system_prefix: str,
) -> str:
    tools = build_ledger_tools(client, bound_ledger_id.strip())
    prompt = ChatPromptTemplate.from_messages(
        [
            ("system", "{system_message}"),
            ("human", "{input}"),
            MessagesPlaceholder("agent_scratchpad"),
        ]
    )
    agent = create_tool_calling_agent(llm, tools, prompt)
    executor = AgentExecutor(
        agent=agent,
        tools=tools,
        max_iterations=AGENT_MAX_ITERATIONS,
        verbose=False,
    )
    result = await executor.ainvoke(
        {"input": user_input, "system_message": system_prefix.strip()}
    )
    content = (result.get("output") or "").strip()
    if not content:
        raise ValueError("agent returned empty response")
    return content
