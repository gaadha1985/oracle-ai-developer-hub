"""Direct OCI Generative AI chat client (canonical recipe).

WHY THIS RECIPE
---------------
The OpenAI-compatible OCI Generative AI endpoint speaks the OpenAI wire format
but is unstable across `openai` SDK versions. With `openai>=1.x`, the
`oci-openai` shim raises `APIConnectionError` whose underlying cause is
`AttributeError: 'URL' object has no attribute 'decode'` — `httpx.URL` reaches
a `urllib.parse._decode_args` that expects a string. See friction P0-2.

The direct OCI SDK path (`oci.generative_ai_inference`) is stable and
recommended at all tiers of choose-your-path. It uses `GenericChatRequest` +
`OnDemandServingMode` and SigV1 auth via `~/.oci/config` (or
InstancePrincipals when running on OCI compute).

USAGE
-----
    client = get_chat_client()
    out = chat_complete([{"role": "user", "content": "Reply OK."}])
    # `out` is a string (the assistant message text).

The model id for Grok 4 is `xai.grok-4` (not `grok-4`). Pass via
`OCI_LLM_MODEL` env var; default is `xai.grok-4`.

REQUIRED ENV
------------
OCI_COMPARTMENT_ID    OCID of the compartment that has GenAI enabled.
OCI_REGION            Defaults to us-chicago-1 (Grok 4 only ships there).
OCI_LLM_MODEL         Defaults to xai.grok-4.

DEPS
----
oci>=2.130   (the only chat dep needed; `oci-openai` and `openai` are NOT used)
"""

from __future__ import annotations

import os
from typing import Any

import oci


_chat_client: Any = None


def get_chat_client() -> Any:
    """Return an `oci.generative_ai_inference.GenerativeAiInferenceClient`
    pinned to `OCI_REGION`. Module-scoped + cached. SigV1 via `~/.oci/config`,
    InstancePrincipals if config is absent."""
    global _chat_client
    if _chat_client is not None:
        return _chat_client

    region = os.environ.get("OCI_REGION", "us-chicago-1")
    try:
        config = oci.config.from_file()
        config["region"] = region
    except Exception:
        from oci.auth.signers import InstancePrincipalsSecurityTokenSigner
        signer = InstancePrincipalsSecurityTokenSigner()
        endpoint = f"https://inference.generativeai.{region}.oci.oraclecloud.com"
        _chat_client = oci.generative_ai_inference.GenerativeAiInferenceClient(
            config={"region": region}, signer=signer, service_endpoint=endpoint
        )
        return _chat_client

    endpoint = f"https://inference.generativeai.{region}.oci.oraclecloud.com"
    _chat_client = oci.generative_ai_inference.GenerativeAiInferenceClient(
        config=config, service_endpoint=endpoint
    )
    return _chat_client


def chat_complete(
    messages: list[dict],
    temperature: float = 0.2,
    max_tokens: int = 512,
) -> str:
    """One-shot chat completion. Returns the assistant message text."""
    client = get_chat_client()
    model_id = os.environ.get("OCI_LLM_MODEL", "xai.grok-4")
    if model_id == "grok-4":
        model_id = "xai.grok-4"

    compartment_id = os.environ.get("OCI_COMPARTMENT_ID")
    if not compartment_id:
        raise RuntimeError(
            "OCI_COMPARTMENT_ID is required for OCI Generative AI calls."
        )

    sdk_messages = []
    for m in messages:
        content = oci.generative_ai_inference.models.TextContent(text=m["content"])
        sdk_messages.append(
            oci.generative_ai_inference.models.Message(
                role=m["role"].upper(), content=[content]
            )
        )

    chat_request = oci.generative_ai_inference.models.GenericChatRequest(
        api_format=oci.generative_ai_inference.models.BaseChatRequest.API_FORMAT_GENERIC,
        messages=sdk_messages,
        max_tokens=max_tokens,
        temperature=temperature,
    )
    chat_details = oci.generative_ai_inference.models.ChatDetails(
        serving_mode=oci.generative_ai_inference.models.OnDemandServingMode(
            model_id=model_id
        ),
        compartment_id=compartment_id,
        chat_request=chat_request,
    )

    resp = client.chat(chat_details)
    return resp.data.chat_response.choices[0].message.content[0].text
