"""骑手端截图验证助手：解析 uiautomator XML + 采样像素颜色。"""
import sys, re, subprocess, os
from xml.etree import ElementTree as ET
from PIL import Image

ADB = r"C:\Users\jsong\AppData\Local\Android\Sdk\platform-tools\adb.exe"
SHOTS = r"D:\code\github\dawn-errand\apps\driver\shots"

def dump_hierarchy():
    subprocess.run([ADB, "shell", "uiautomator", "dump", "/sdcard/ui.xml"], check=False, capture_output=True)
    subprocess.run([ADB, "pull", "/sdcard/ui.xml", os.path.join(SHOTS, "_ui.xml")], check=False, capture_output=True)
    p = os.path.join(SHOTS, "_ui.xml")
    if not os.path.exists(p):
        print("NO_HIERARCHY"); return []
    tree = ET.parse(p)
    out = []
    for node in tree.iter("node"):
        text = node.get("text", "")
        cls = node.get("class", "")
        b = node.get("bounds", "")
        if text.strip():
            m = re.match(r"\[(\d+),(\d+)\]\[(\d+),(\d+)\]", b)
            if m:
                x1, y1, x2, y2 = map(int, m.groups())
                out.append((text, cls, x1, y1, x2, y2))
    return out

def sample(path, points):
    im = Image.open(path).convert("RGB")
    res = {}
    for name, (x, y) in points.items():
        if 0 <= x < im.width and 0 <= y < im.height:
            res[name] = "#%02X%02X%02X" % im.getpixel((x, y))
        else:
            res[name] = "OOB"
    return res

if __name__ == "__main__":
    mode = sys.argv[1]
    if mode == "dump":
        for t, c, x1, y1, x2, y2 in dump_hierarchy():
            print(f"{t} | {c.split('.')[-1]} | {x1},{y1} {x2},{y2} | dp=({x1/2.625:.0f},{y1/2.625:.0f})-({x2/2.625:.0f},{y2/2.625:.0f})")
    elif mode == "pixel":
        path = sys.argv[2]
        pts = {}
        for spec in sys.argv[3:]:
            name, xy = spec.split("=")
            x, y = map(int, xy.split(","))
            pts[name] = (x, y)
        for k, v in sample(path, pts).items():
            print(f"{k}={v}")
