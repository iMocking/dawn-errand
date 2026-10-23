import re
import sys

path = sys.argv[1] if len(sys.argv) > 1 else r"D:\code\github\dawn-errand\apps\driver\ui.xml"
with open(path, encoding="utf-8") as f:
    x = f.read()
nodes = re.findall(
    r'<node[^>]*text="([^"]*)"[^>]*resource-id="([^"]*)"[^>]*class="([^"]*)"[^>]*bounds="\[(\d+),(\d+)\]\[(\d+),(\d+)\]"[^>]*>',
    x,
)
for t, rid, cls, x1, y1, x2, y2 in nodes:
    if t.strip() or rid:
        print(f"{t!r} | {rid} | {cls.split('.')[-1]} | [{x1},{y1}][{x2},{y2}]")
