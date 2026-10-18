from pathlib import Path
import sys
from reportlab.pdfgen import canvas
from reportlab.lib.utils import ImageReader
from pypdf import PdfReader

root = Path(__file__).resolve().parent
output = Path(sys.argv[1])
times = [4, 11, 26, 33, 42, 50, 60]
if output.exists():
    raise FileExistsError(output)
pdf = canvas.Canvas(str(output), pagesize=(960, 540), pageCompression=1)
pdf.setTitle("把温柔，写成一朵玫瑰")
pdf.setAuthor("Animation")
pdf.setSubject("七幕动画关键画面，静态分镜")
for index, at in enumerate(times):
    source = root / f"frame-{index + 1}.png"
    image = ImageReader(str(source))
    if image.getSize() != (3840, 2160):
        raise ValueError(f"Unexpected frame size at {at}s: {image.getSize()}")
    pdf.bookmarkPage(f"scene-{index + 1}")
    pdf.drawImage(image, 0, 0, width=960, height=540)
    pdf.showPage()
pdf.save()
result = PdfReader(str(output))
assert len(result.pages) == 7
assert all(float(p.mediabox.width) == 960 and float(p.mediabox.height) == 540 for p in result.pages)
print(f"PDF: {output}")
print(f"Pages: {len(result.pages)}; page size: 960 x 540 pt")
