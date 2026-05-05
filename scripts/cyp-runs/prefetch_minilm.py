"""Pre-cache sentence-transformers/all-MiniLM-L6-v2 (~90MB) so beginner run
doesn't stall on first-time HF download. Idempotent."""
from sentence_transformers import SentenceTransformer

print("downloading sentence-transformers/all-MiniLM-L6-v2 ...")
model = SentenceTransformer("sentence-transformers/all-MiniLM-L6-v2")
v = model.encode(["dim check"])
assert v.shape == (1, 384), f"unexpected shape: {v.shape}"
print(f"cached. dim={v.shape[1]}")
