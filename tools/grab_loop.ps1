param(
    [string]$LogFile = "$env:TEMP\vcapture15.log",
    [string]$OutDir = "$env:TEMP\real_captures"
)
Add-Type -AssemblyName System.Drawing
Add-Type @'
using System;
using System.Runtime.InteropServices;
public class W32b {
    [DllImport("user32.dll")] public static extern bool PrintWindow(IntPtr hWnd, IntPtr hdcBlit, uint nFlags);
    [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr hWnd, out RECT r);
    [StructLayout(LayoutKind.Sequential)] public struct RECT { public int Left, Top, Right, Bottom; }
}
'@
function Get-GodotHwnd {
    $p = Get-Process | Where-Object { $_.ProcessName -like "Godot*" -and $_.MainWindowHandle -ne 0 } | Select-Object -First 1
    if ($p) { return [IntPtr]$p.MainWindowHandle } else { return [IntPtr]::Zero }
}
function Save-WindowCapture([IntPtr]$hwnd, [string]$path) {
    $rect = New-Object W32b+RECT
    [W32b]::GetWindowRect($hwnd, [ref]$rect) | Out-Null
    $w = $rect.Right - $rect.Left; $h = $rect.Bottom - $rect.Top
    if ($w -le 0 -or $h -le 0) { return $false }
    $bmp = New-Object System.Drawing.Bitmap($w, $h)
    $g = [System.Drawing.Graphics]::FromImage($bmp)
    $hdc = $g.GetHdc()
    $ok = [W32b]::PrintWindow($hwnd, $hdc, 2)
    $g.ReleaseHdc($hdc); $g.Dispose()
    if ($ok) { $bmp.Save($path, [System.Drawing.Imaging.ImageFormat]::Png) }
    $bmp.Dispose()
    return $ok
}
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null
$seen = 0
$deadline = (Get-Date).AddSeconds(130)
$hwnd = [IntPtr]::Zero
while ($seen -lt 16 -and (Get-Date) -lt $deadline) {
    $all = @()
    if (Test-Path -LiteralPath $LogFile) { $all = @(Get-Content -LiteralPath $LogFile -ErrorAction SilentlyContinue) }
    $target = $null
    $n = 0
    foreach ($ln in $all) {
        if ($ln -like "SHOT: *") {
            if ($n -eq $seen) { $target = $ln; break }
            $n = $n + 1
        }
    }
    if ($null -ne $target) {
        if ($hwnd -eq [IntPtr]::Zero) { $hwnd = Get-GodotHwnd }
        if ($hwnd -eq [IntPtr]::Zero) { Start-Sleep -Milliseconds 300; continue }
        $name = $target.Trim().Substring(6)
        Start-Sleep -Milliseconds 900
        $fname = ("{0:D2}" -f $seen) + "_" + $name + ".png"
        $ok = Save-WindowCapture $hwnd "$OutDir\$fname"
        "GRAB[$seen] $fname ok=$ok"
        $seen = $seen + 1
        continue
    }
    Start-Sleep -Milliseconds 250
}
"DONE grabbed=$seen"
