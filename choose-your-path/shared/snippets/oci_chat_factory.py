"""Dual-auth chat client for OCI's OpenAI-compatible endpoint.

Source: ~/git/personal/oci-genai-service/src/oci_genai_service/inference/chat.py:1-95

WHY THIS EXISTS
---------------
OCI's OpenAI-compat endpoint speaks the OpenAI wire format but does NOT accept
a bearer-token `api_key` for most tenancies — every request must be signed with
OCI Signature V1. Plain `OpenAI(api_key="oci")` returns 401.

`get_chat_client()` picks the right auth pattern:
  * If `OCI_API_KEY` is set: bearer-token mode via the upstream `openai` client.
    (Only some tenancies enable this.)
  * Else: SigV1 via `oci-openai` SDK, which wraps `openai` and slips the OCI
    signer onto every request. Reads `~/.oci/config` profile DEFAULT, or falls
    back to InstancePrincipals when running on OCI compute.

USAGE
-----
    client = get_chat_client(region="us-chicago-1",
                             compartment_id=os.environ["OCI_COMPARTMENT_ID"])
    resp = client.chat.completions.create(
        model="grok-4",
        messages=[{"role": "user", "content": "hi"}],
    )
"""

from __future__ import annotations

import os
from typing import Any


def get_chat_base_url(region: str = "us-chicago-1") -> str:
    return (
        f"https://inference.generativeai.{region}.oci.oraclecloud.com"
        f"/20231130/actions/openai"
    )


def get_chat_client(
    region: str = "us-chicago-1",
    compartment_id: str | None = None,
    api_key: str | None = None,
) -> Any:
    """Return an OpenAI-compatible client signed appropriately for OCI."""
    base_url = get_chat_base_url(region)
    api_key = api_key or os.environ.get("OCI_API_KEY")

    if api_key:
        # Pattern 2: bearer-token mode (only if your tenancy enables it).
        from openai import OpenAI
        return OpenAI(base_url=base_url, api_key=api_key)

    # Pattern 1 (default): SigV1 via oci-openai SDK.
    import oci
    from oci_openai import OciOpenAI

    try:
        config = oci.config.from_file("~/.oci/config", "DEFAULT")
        signer = oci.signer.Signer(
            tenancy=config["tenancy"],
            user=config["user"],
            fingerprint=config["fingerprint"],
            private_key_file_location=config["key_file"],
        )
    except Exception:
        # On OCI compute with instance principals, no config file exists.
        from oci.auth.signers import InstancePrincipalsSecurityTokenSigner
        signer = InstancePrincipalsSecurityTokenSigner()

    if not compartment_id:
        compartment_id = os.environ.get("OCI_COMPARTMENT_ID")
    if not compartment_id:
        raise RuntimeError(
            "OCI_COMPARTMENT_ID is required for SigV1 auth — set the env var "
            "or pass compartment_id=..."
        )

    return OciOpenAI(base_url=base_url, auth=signer, compartment_id=compartment_id)
