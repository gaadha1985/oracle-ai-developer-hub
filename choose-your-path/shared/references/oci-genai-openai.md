# OCI Generative AI — OpenAI-compatible endpoint

For intermediate / advanced users who want hosted inference instead of local Ollama. Same `OpenAI`-shaped Python client, different `base_url` and auth.

## Endpoint shape

```
base_url = "https://inference.generativeai.us-chicago-1.oci.oraclecloud.com/20231130/actions/openai"
```

The endpoint speaks the OpenAI wire format **but does not accept a bearer-token `api_key`** — every request must be signed with OCI Signature V1. `from openai import OpenAI; OpenAI(api_key="oci")` returns 401.

Use one of the two patterns below.

### Pattern 1: `oci-openai` SDK (the right way)

```bash
pip install oci-openai
```

```python
from oci_openai import OciOpenAI
import oci

config = oci.config.from_file("~/.oci/config", "DEFAULT")
signer = oci.signer.Signer(
    tenancy=config["tenancy"],
    user=config["user"],
    fingerprint=config["fingerprint"],
    private_key_file_location=config["key_file"],
)

client = OciOpenAI(
    base_url="https://inference.generativeai.us-chicago-1.oci.oraclecloud.com/20231130/actions/openai",
    auth=signer,
    compartment_id=os.environ["OCI_COMPARTMENT_ID"],
)
resp = client.chat.completions.create(model="grok-4", messages=[...])
```

`oci-openai` wraps the `openai` Python client and slips the OCI signer onto every request. Same `client.chat.completions.create(...)` surface, but signed.

### Pattern 2: API-key auth (only if your tenancy enables it)

Some OCI tenancies expose a personal-access-style API key for the OpenAI-compat endpoint. If you have one, plain `openai` works:

```python
from openai import OpenAI
client = OpenAI(base_url=BASE_URL, api_key=OCI_API_KEY)
```

If unsure, use Pattern 1 — Signature V1 always works given a valid `~/.oci/config`. Source: `~/git/personal/oci-genai-service/src/oci_genai_service/inference/chat.py:1-95` shows the dual-auth selector pattern.

## Region matrix (where each model lives)

| Model | Regions | Notes |
| --- | --- | --- |
| `grok-4` | `us-chicago-1` only | Default for intermediate path. The skill warns + offers fallback if the user's tenancy isn't in chicago. |
| `cohere.command-r-plus` | most regions | Reliable fallback for chat. |
| `meta.llama-3.3-70b-instruct` | most regions | Open-weight option hosted in OCI. |
| `cohere.embed-english-v3.0` | most regions | 1024-dim. Default embedder for intermediate path. |
| `cohere.embed-multilingual-v3.0` | most regions | 1024-dim. Use when content isn't English-only. |

The model + region matrix changes — when in doubt, link the user to https://docs.oracle.com/en-us/iaas/Content/generative-ai/pretrained-models.htm rather than guessing.

## Auth

Two paths the skill supports.

### 1. API key file (`~/.oci/config`)

Default for laptop development. The user runs `oci setup config` once. The skill checks `~/.oci/config` exists and the `[DEFAULT]` profile has the keys it needs:

```ini
[DEFAULT]
user=ocid1.user.oc1..xxxx
fingerprint=xx:xx:...
key_file=~/.oci/oci_api_key.pem
tenancy=ocid1.tenancy.oc1..xxxx
region=us-chicago-1
```

### 2. Instance principal (no config file)

When the code runs on an OCI compute VM. The skill detects this via "no `~/.oci/config` AND `OCI_RESOURCE_PRINCIPAL_VERSION` env var present" and switches to:

```python
from oci.auth.signers import InstancePrincipalsSecurityTokenSigner
signer = InstancePrincipalsSecurityTokenSigner()
```

Source: `~/git/work/ai-solutions/apps/langgraph_agent_with_genai/src/jlibspython/oci_embedding_utils.py:1-80` — same fallback pattern.

## Embeddings (Cohere via the `oci` SDK directly)

The OpenAI-compat surface covers chat. For embeddings, the cleanest path is the OCI Python SDK:

```python
import oci
from oci.generative_ai_inference import GenerativeAiInferenceClient
from oci.generative_ai_inference.models import (
    EmbedTextDetails, OnDemandServingMode,
)

config = oci.config.from_file("~/.oci/config", "DEFAULT")
client = GenerativeAiInferenceClient(config=config)
resp = client.embed_text(EmbedTextDetails(
    serving_mode=OnDemandServingMode(model_id="cohere.embed-english-v3.0"),
    inputs=["hello world"],
    truncate="NONE",
    compartment_id=os.environ["OCI_COMPARTMENT_ID"],
))
vector = resp.data.embeddings[0]  # list[float], length 1024
```

Source: `~/git/personal/oci-genai-service/src/oci_genai_service/inference/embeddings.py:1-60`.

The intermediate skill wraps this in a LangChain `Embeddings` subclass so the rest of the project speaks LangChain. The wrapper:
- Filters empty strings before the API call (the API rejects them with a 400).
- Batches in chunks of 96 (Cohere's per-call limit).
- Caches the client at module scope.

## Chat via the OpenAI-compat endpoint, in LangChain

```python
from langchain_openai import ChatOpenAI
llm = ChatOpenAI(
    model="grok-4",
    base_url="https://inference.generativeai.us-chicago-1.oci.oraclecloud.com/20231130/actions/openai",
    api_key=os.environ["OCI_API_KEY"],   # see auth section
    temperature=0.2,
)
```

Source: `~/git/personal/oci-genai-service/src/oci_genai_service/inference/chat.py:1-95`. Streaming works the same way as with native OpenAI — `stream=True` on the `OpenAI` client or `llm.stream(...)` in LangChain.

## Cost / quota gotchas the skill should mention once

- OCI GenAI is **not free**. The user pays per token, billed to their tenancy. The intermediate skill says this once during the interview and again in the README.
- Free trial / always-free credits exist but Grok 4 isn't on the always-free list.
- Quota limits are per-region. The skill doesn't try to predict them — it surfaces the OCI error and tells the user what to do (request quota increase via the OCI console).

## What the skill does NOT do

- Doesn't manage the OCI tenancy / compartment / IAM policies. The user has to set those up themselves.
- Doesn't try to programmatically create OCI keys. The user runs `oci setup config` if they don't have one.
- Doesn't pin a specific OCI SDK version beyond `oci>=2.130` — newer is fine.

## Exemplars

| Pattern | File |
| --- | --- |
| Embeddings, clean | `~/git/personal/oci-genai-service/src/oci_genai_service/inference/embeddings.py:1-60` |
| Chat (OpenAI-compat) | `~/git/personal/oci-genai-service/src/oci_genai_service/inference/chat.py:1-95` |
| Embeddings + IP fallback | `~/git/work/ai-solutions/apps/langgraph_agent_with_genai/src/jlibspython/oci_embedding_utils.py:1-80` |
| OracleVS exposed as /chat/completions | `apps/agentic_rag/src/openai_compat.py:54+` |
