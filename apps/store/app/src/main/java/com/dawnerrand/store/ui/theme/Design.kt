package com.dawnerrand.store.ui.theme

/**
 * 设计令牌：与 design/store 设计稿（index.html 色板）保持一致。
 * 8 位十六进制色值超出 Int.MAX，统一 .toInt() 转 Int。
 */
object Design {

    // 品牌色
    val PRIMARY: Int = 0xFFFF6B2C.toInt()      // 主色
    val PRIMARY_LIGHT: Int = 0xFFFF8A3D.toInt() // 浅橙
    val PRIMARY_DEEP: Int = 0xFFFF5A1F.toInt()  // 深橙
    val REWARD: Int = 0xFFFF4D4F.toInt()        // 警示/待接单
    val SUCCESS: Int = 0xFF00B578.toInt()       // 成功
    val SUCCESS_LIGHT: Int = 0xFF36CFC9.toInt() // 青绿
    val BLUE: Int = 0xFF2B7FFF.toInt()          // 制作中/配送中
    val PURPLE: Int = 0xFF722ED1.toInt()        // 经营数据图标
    val GOLD: Int = 0xFFD48806.toInt()          // 菜品管理图标
    val GOLD_2: Int = 0xFFFA8C16.toInt()        // 排行第3
    val GOLD_3: Int = 0xFFFAAD14.toInt()        // 西红柿炒蛋
    val GREEN: Int = 0xFF52C41A.toInt()         // 手撕包菜
    val RED: Int = 0xFFF5222D.toInt()           // 麻婆豆腐

    // 文字
    val TEXT_PRIMARY: Int = 0xFF1D2129.toInt()
    val TEXT_SECONDARY: Int = 0xFF4E5969.toInt()
    val TEXT_TERTIARY: Int = 0xFF86909C.toInt()
    val TEXT_FAINT: Int = 0xFFC9CDD4.toInt()

    // 线/背景
    val DIVIDER: Int = 0xFFF2F3F5.toInt()
    val BACKGROUND: Int = 0xFFF5F6F8.toInt()
    val WHITE: Int = 0xFFFFFFFF.toInt()

    // 浅色功能底
    val ORANGE_LIGHT: Int = 0xFFFFF3E8.toInt()
    val BLUE_LIGHT: Int = 0xFFE8F0FF.toInt()
    val GREEN_LIGHT: Int = 0xFFE8F7F0.toInt()
    val GRAY_LIGHT: Int = 0xFFF2F3F5.toInt()
    val GOLD_LIGHT: Int = 0xFFFFF8E0.toInt()
    val RED_LIGHT: Int = 0xFFFFE8E8.toInt()
    val PURPLE_LIGHT: Int = 0xFFF0E8FF.toInt()
}
