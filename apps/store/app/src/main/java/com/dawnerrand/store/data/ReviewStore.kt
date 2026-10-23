package com.dawnerrand.store.data

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData

data class Review(
    val id: String,
    val customer: String,
    val rating: Int,
    val time: String,
    val content: String,
    val tags: List<String> = emptyList(),
    val reply: String? = null,
)

/**
 * 顾客评价（演示数据）：评分汇总 + 好评/中评/差评筛选 + 商家回复。
 * 回复保存在内存共享状态，页面间切换不丢失。
 */
object ReviewStore {

    val rating = "4.8"
    val total = 326
    val goodRate = "98%"
    val goodCount = 312
    val midCount = 9
    val badCount = 5

    private val _reviews = MutableLiveData(
        listOf(
            Review(
                "r1", "王建国", 5, "今天 12:40",
                "农家小炒肉很下饭，分量足，米饭也热乎。骑手送得很快，下次还点。",
                listOf("味道赞", "分量足"),
                "谢谢支持，欢迎常来！"
            ),
            Review(
                "r2", "李秀英", 5, "今天 11:52",
                "红烧土鸡块炖得入味，就是有点辣，备注少辣做到了。",
                listOf("味道赞")
            ),
            Review(
                "r3", "张大山", 4, "昨天 19:20",
                "手撕包菜有点油，其他都挺好。出餐速度可以再快一点。",
                listOf("出餐稍慢")
            ),
            Review(
                "r4", "刘春梅", 5, "昨天 18:05",
                "西红柿炒蛋酸甜合适，孩子很喜欢，包装也没有洒。",
                listOf("包装完好", "孩子爱吃")
            ),
            Review(
                "r5", "赵铁柱", 3, "10月21日 12:30",
                "麻婆豆腐偏咸，米饭分量可以。希望注意一下口味。",
                listOf("口味偏咸")
            ),
            Review(
                "r6", "孙秀兰", 5, "10月20日 19:10",
                "第三次点了，出餐快，骑手也准时，五星好评。",
                listOf("出餐快", "准时送达")
            ),
        )
    )
    val reviews: LiveData<List<Review>> = _reviews

    fun filtered(filter: Int): List<Review> {
        val all = _reviews.value ?: emptyList()
        return when (filter) {
            1 -> all.filter { it.rating >= 4 }  // 好评
            2 -> all.filter { it.rating == 3 }  // 中评
            3 -> all.filter { it.rating <= 2 }  // 差评
            else -> all
        }
    }

    fun count(filter: Int): Int {
        val all = _reviews.value ?: emptyList()
        return when (filter) {
            1 -> all.count { it.rating >= 4 }
            2 -> all.count { it.rating == 3 }
            3 -> all.count { it.rating <= 2 }
            else -> all.size
        }
    }

    fun reply(id: String, text: String) {
        val clean = text.trim()
        if (clean.isEmpty()) return
        _reviews.value = (_reviews.value ?: emptyList()).map {
            if (it.id == id) it.copy(reply = clean) else it
        }
    }
}
