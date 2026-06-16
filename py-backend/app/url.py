import ipaddress
from urllib.parse import urlparse

ALLOWED_LOCAL_HOSTS = {
    "127.0.0.1",
    "localhost",
    "::1",
    "ollama",
    "host.docker.internal",
}


class InvalidBaseURL(Exception):
    pass


def normalize_llm_base_url(raw: str) -> str:
    raw = (raw or "").strip()
    if not raw:
        raise InvalidBaseURL("ai base url not allowed")

    parsed = urlparse(raw)
    host = (parsed.hostname or "").lower()
    scheme = (parsed.scheme or "").lower()

    if scheme == "https":
        if _is_blocked_host(host):
            raise InvalidBaseURL("ai base url not allowed")
    elif scheme == "http":
        if host not in ALLOWED_LOCAL_HOSTS:
            raise InvalidBaseURL("ai base url not allowed")
    else:
        raise InvalidBaseURL("ai base url not allowed")

    return _ensure_openai_v1_suffix(raw)


def _ensure_openai_v1_suffix(raw: str) -> str:
    raw = raw.rstrip("/")
    if raw.endswith("/v1"):
        return raw
    return raw + "/v1"


def _is_blocked_host(host: str) -> bool:
    if not host:
        return True
    if host in ALLOWED_LOCAL_HOSTS:
        return False
    try:
        ip = ipaddress.ip_address(host)
    except ValueError:
        return False
    return (
        ip.is_loopback
        or ip.is_private
        or ip.is_link_local
        or ip.is_multicast
    )
