param(
    [int]$Grabs = 14,
    [int]$IntervalMs = 6000,
    [int]$FirstDelayMs = 6000,
    [string]$OutDir = "$env:TEMP\real_captures"
)
Add-Type -AssemblyName System.Drawing
Add-Type @'
using System;
using System.Runtime.InteropServices;
using System.Text;
public class W32e {
    public delegate bool EnumProc(IntPtr hWnd, IntPtr lParam);
    [DllImport("user32.dll")] public static extern bool EnumWindows(EnumProc cb, IntPtr lParam);
    [DllImport("user32.dll")] public static extern int GetWindowText(IntPtr hWnd, StringBuilder text, int count);
    [DllImport("user32.dll")] public static extern bool IsWindowVisible(IntPtr hWnd);
    [DllImport("user32.dll")] public static extern bool IsIconic(IntPtr hWnd);
    [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr hWnd, out RECT r);
    [DllImport("user32.dll")] public static extern bool PrintWindow(IntPtr hWnd, IntPtr hdcBlit, uint nFlags);
    [StructLayout(LayoutKind.Sequential)] public struct RECT { public int Left, Top, Right, Bottom; }
}
'@
function Get-GodotHwndByTitle {
    $found = [IntPtr]::Zero
    $cb = [W32e+EnumProc]{
        param($h, $lp)
        if (-not [W32e]::IsWindowVisible($h)) { return $true }
        $sb = New-Object System.Text.StringBuilder 256
        [void][W32e]::GetWindowText($h, $sb, 256)
        $t = $sb.ToString()
        if ($t -like "*曙光*") { $script:foundHwnd = $h; return $false }
        return $true
    }
    $script:foundHwnd = [IntPtr]::Zero
    [void][W32e]::EnumWindows($cb, [IntPtr]::Zero)
    return $script:foundHwnd
}
function Save-WindowCapture([IntPtr]$hwnd, [string]$path) {
    $rect = New-Object W32e+RECT
    [W32e]::GetWindowRect($hwnd, [ref]$rect) | Out-Null
    $w = $rect.Right - $rect.Left; $h = $rect.Bottom - $rect.Top
    if ($w -le 0 -or $h -le 0) { return $false }
    $bmp = New-Object System.Drawing.Bitmap($w, $h)
    $g = [System.Drawing.Graphics]::FromImage($bmp)
    $hdc = $g.GetHdc()
    $ok = [W32e]::PrintWindow($hwnd, $hdc, 2)
    $g.ReleaseHdc($hdc); $g.Dispose()
    if ($ok) { $bmp.Save($path, [System.Drawing.Imaging.ImageFormat]::Png) }
    $bmp.Dispose()
    return $ok
}
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null
$hwnd = [IntPtr]::Zero
$deadline = (Get-Date).AddSeconds(30)
while ($hwnd -eq [IntPtr]::Zero -and (Get-Date) -lt $deadline) {
    $hwnd = Get-GodotHwndByTitle
    if ($hwnd -eq [IntPtr]::Zero) { Start-Sleep -Milliseconds 400 }
}
"hwnd=$hwnd"
Start-Sleep -Milliseconds $FirstDelayMs
for ($i = 0; $i -lt $Grabs; $i++) {
    $hwnd = Get-GodotHwndByTitle
    if ($hwnd -eq [IntPtr]::Zero) { "GRAB[$i] no-hwnd"; Start-Sleep -Milliseconds $IntervalMs; continue }
    $fname = ("{0:D2}" -f $i) + ".png"
    $ok = Save-WindowCapture $hwnd "$OutDir\$fname"
    "GRAB[$i] $fname ok=$ok"
    Start-Sleep -Milliseconds $IntervalMs
}
"DONE"
