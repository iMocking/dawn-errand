package com.dawnerrand.client.ui.theme

/**
 * 设计令牌：与 design/client 设计稿及 docs/客户端交付说明.md 保持一致。
 * 8 位十六进制色值超出 Int.MAX，统一 .toInt() 转 Int。
 */
object Design {

    // 品牌色
    val PRIMARY: Int = 0xFFFF6B2C.toInt()      // 曙光橙
    val SECONDARY: Int = 0xFFFF8A3D.toInt()    // 浅橙
    val DEEP_ORANGE: Int = 0xFFFF5A1F.toInt()  // 深橙
    val REWARD: Int = 0xFFFF4D4F.toInt()       // 价格红
    val SUCCESS: Int = 0xFF00B578.toInt()      // 完成绿

    // 文字
    val TEXT_PRIMARY: Int = 0xFF1D2129.toInt()
    val TEXT_SECONDARY: Int = 0xFF4E5969.toInt()
    val TEXT_TERTIARY: Int = 0xFF86909C.toInt()
    val TEXT_FAINT: Int = 0xFFC9CDD4.toInt()

    // 线 / 背景
    val DIVIDER: Int = 0xFFF2F3F5.toInt()
    val BACKGROUND: Int = 0xFFF5F6F8.toInt()
    val MAP_BG: Int = 0xFFE8ECEB.toInt()

    // 白色
    val WHITE: Int = 0xFFFFFFFF.toInt()
    val WHITE_20: Int = 0x33FFFFFF.toInt()

    // 评分星
    val STAR: Int = 0xFFFFB020.toInt()

    // 其他
    val PURPLE: Int = 0xFF722ED1.toInt()
    val ORANGE_DEEP: Int = 0xFFFF7D00.toInt()
    val INK: Int = 0xFF141414.toInt()
}
