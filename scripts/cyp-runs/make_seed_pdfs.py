"""Generate 3 small PDFs (one chapter each) for the beginner run's corpus.
Uses Alice in Wonderland (public domain) chunks so the chat-evidence step
has groundable, citable content."""
from pathlib import Path
from reportlab.lib.pagesizes import letter
from reportlab.pdfgen import canvas

CHAPTERS = {
    "alice-chap1.pdf": (
        "Chapter I: Down the Rabbit-Hole",
        [
            "Alice was beginning to get very tired of sitting by her sister on the bank,",
            "and of having nothing to do: once or twice she had peeped into the book her",
            "sister was reading, but it had no pictures or conversations in it,",
            "'and what is the use of a book,' thought Alice, 'without pictures or conversations?'",
            "So she was considering in her own mind whether the pleasure of making a",
            "daisy-chain would be worth the trouble of getting up and picking the daisies,",
            "when suddenly a White Rabbit with pink eyes ran close by her.",
        ],
    ),
    "alice-chap2.pdf": (
        "Chapter II: The Pool of Tears",
        [
            "'Curiouser and curiouser!' cried Alice (she was so much surprised, that for",
            "the moment she quite forgot how to speak good English).",
            "'Now I'm opening out like the largest telescope that ever was! Good-bye, feet!'",
            "for when she looked down at her feet, they seemed to be almost out of sight,",
            "they were getting so far off.",
            "'Oh, my poor little feet, I wonder who will put on your shoes and stockings",
            "for you now, dears?'",
        ],
    ),
    "alice-chap3.pdf": (
        "Chapter III: A Caucus-Race and a Long Tale",
        [
            "They were indeed a queer-looking party that assembled on the bank--the birds",
            "with draggled feathers, the animals with their fur clinging close to them,",
            "and all dripping wet, cross, and uncomfortable.",
            "The first question of course was, how to get dry again: they had a",
            "consultation about this, and after a few minutes it seemed quite natural to",
            "Alice to find herself talking familiarly with them.",
        ],
    ),
}

OUT = Path(__file__).parent / "seed_pdfs"
OUT.mkdir(exist_ok=True)

for name, (title, lines) in CHAPTERS.items():
    c = canvas.Canvas(str(OUT / name), pagesize=letter)
    c.setFont("Helvetica-Bold", 16)
    c.drawString(72, 720, title)
    c.setFont("Helvetica", 11)
    y = 690
    for line in lines:
        c.drawString(72, y, line)
        y -= 18
    c.showPage()
    c.save()
    print(f"wrote {name}")
