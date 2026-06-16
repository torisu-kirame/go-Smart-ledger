import os
from pathlib import Path

APP_DIR = Path(__file__).resolve().parent
WORKSPACE_DIR = APP_DIR / "workspace"

AGENT_CONFIG_PATH = os.getenv(
    "AGENT_CONFIG_PATH",
    str(Path.cwd() / "data" / "agent" / "config"),
)
LEDGER_API_URL = os.getenv("LEDGER_API_URL", "http://ledger-api:28888").rstrip("/")
PORT = int(os.getenv("PORT", "28891"))
