class_name DesignTokens
extends RefCounted
## 设计规范（对齐 design/client 设计稿）：
## 主色 #FF6B2C、价格 #FF4D4F、成功 #00B578，字号 10-28，圆角 8-24。

const PRIMARY := Color("FF6B2C")
const PRIMARY_LIGHT := Color("FF8A3D")
const PRIMARY_DARK := Color("FF5A1F")
const PRIMARY_BG := Color("FFF3E8")        # 主色浅底（标签/激活态）
const PRICE := Color("FF4D4F")
const SUCCESS := Color("00B578")
const SUCCESS_BG := Color("E8F7F0")
const TEXT := Color("1D2129")
const TEXT_SEC := Color("4E5969")
const TEXT_MUTED := Color("86909C")
const TEXT_FAINT := Color("C9CDD4")
const BORDER := Color("E5E6EB")
const DIVIDER := Color("F2F3F5")
const BG := Color("F5F6F8")
const STAR := Color("FFB020")
const WARN_BG := Color("FFF8E0")
const WARN_LINE := Color("FFD66B")

## 渐变
const GRAD_PRIMARY := [Color("FF8A3D"), Color("FF6B2C")]
const GRAD_PRIMARY_BTN := [Color("FF8A3D"), Color("FF5A1F")]
const GRAD_SCENE_EAT := [Color("FF6B2C"), Color("FF5A1F")]
const GRAD_SCENE_RUN := [Color("FF9F43"), Color("FF7A2C")]

const FONT_XS := 10
const FONT_SM := 11
const FONT_MD := 12
const FONT_LG := 13
const FONT_XL := 14
const FONT_XXL := 15
const FONT_TITLE := 16
const FONT_H1 := 17
const FONT_H2 := 20
const FONT_H3 := 22
const FONT_HERO := 28

const RADIUS_SM := 8.0
const RADIUS_MD := 12.0
const RADIUS_LG := 16.0
const RADIUS_XL := 24.0

const STATUS_BAR_H := 44.0
const TABBAR_H := 64.0
const NAV_H := 44.0
