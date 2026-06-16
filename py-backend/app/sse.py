import json
from typing import AsyncIterator

from starlette.responses import StreamingResponse


def chunk_text(text: str, size: int = 64) -> list[str]:
    if size <= 0 or len(text) <= size:
        return [text]
    runes = list(text)
    out: list[str] = []
    for i in range(0, len(runes), size):
        out.append("".join(runes[i : i + size]))
    return out


async def sse_stream(chunks: AsyncIterator[str]) -> AsyncIterator[str]:
    async for delta in chunks:
        if not delta.strip():
            continue
        payload = json.dumps(
            {"choices": [{"index": 0, "delta": {"content": delta}}]},
            ensure_ascii=False,
        )
        yield f"data: {payload}\n\n"
    yield "data: [DONE]\n\n"


async def sse_from_text(text: str) -> AsyncIterator[str]:
    async def gen() -> AsyncIterator[str]:
        for part in chunk_text(text):
            yield part

    return sse_stream(gen())


def sse_error_response(message: str) -> StreamingResponse:
    payload = json.dumps({"error": {"message": message}}, ensure_ascii=False)

    async def body() -> AsyncIterator[str]:
        yield f"data: {payload}\n\n"
        yield "data: [DONE]\n\n"

    return StreamingResponse(
        body(),
        media_type="text/event-stream",
        headers={"Cache-Control": "no-cache", "Connection": "keep-alive"},
    )
