Add-Type -AssemblyName System.Drawing
Add-Type @'
using System;
using System.Runtime.InteropServices;
public class W32 {
    [DllImport("user32.dll")] public static extern bool PrintWindow(IntPtr hWnd, IntPtr hdcBlit, uint nFlags);
    [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr hWnd, out RECT r);
    [DllImport("user32.dll")] public static extern bool IsWindowVisible(IntPtr hWnd);
    [DllImport("user32.dll")] public static extern bool IsIconic(IntPtr hWnd);
    [StructLayout(LayoutKind.Sequential)] public struct RECT { public int Left, Top, Right, Bottom; }
}
'@
$p = Get-Process | Where-Object { $_.ProcessName -like "Godot*" -and $_.MainWindowHandle -ne 0 } | Select-Object -First 1
if (-not $p) { "NO_WINDOW_PROCESS"; exit 1 }
$hwnd = $p.MainWindowHandle
"hwnd=$hwnd title=$($p.MainWindowTitle)"
$rect = New-Object W32+RECT
[W32]::GetWindowRect($hwnd, [ref]$rect) | Out-Null
"rect=$($rect.Left),$($rect.Top),$($rect.Right),$($rect.Bottom) vis=$([W32]::IsWindowVisible($hwnd)) icon=$([W32]::IsIconic($hwnd))"
$w = $rect.Right - $rect.Left; $h = $rect.Bottom - $rect.Top
if ($w -le 0 -or $h -le 0) { "BAD_SIZE"; exit 1 }
$bmp = New-Object System.Drawing.Bitmap($w, $h)
$g = [System.Drawing.Graphics]::FromImage($bmp)
$hdc = $g.GetHdc()
$ok = [W32]::PrintWindow($hwnd, $hdc, 2)
$g.ReleaseHdc($hdc)
$g.Dispose()
"printwindow_ok=$ok"
$bmp.Save("$env:TEMP\godot_real.png", [System.Drawing.Imaging.ImageFormat]::Png)
$bmp.Dispose()
"SAVED $env:TEMP\godot_real.png"
