<script setup lang="ts" generic="T extends Record<string, unknown>">
import Empty from './Empty.vue'

export interface Column {
  key: string
  title: string
  width?: string
  align?: 'left' | 'center' | 'right'
  /** 插槽名，默认 `col-${key}` */
  slot?: string
}

withDefaults(
  defineProps<{
    columns: Column[]
    rows: T[]
    loading?: boolean
    rowKey?: string
  }>(),
  { loading: false, rowKey: 'id' },
)

function cellText(row: T, key: string): string | number {
  const v = row[key as keyof T]
  if (v === null || v === undefined) return ''
  if (typeof v === 'object') return JSON.stringify(v)
  return v as string | number
}
</script>

<template>
  <div class="w-full overflow-x-auto">
    <table class="w-full border-collapse text-sm">
      <thead class="radius-tl radius-tr bg-white text-ink-2">
        <tr>
          <th
            v-for="c in columns"
            :key="c.key"
            class="whitespace-nowrap border-b border-line px-4 py-3 text-left text-[13px] font-medium text-ink-2"
            :style="{ width: c.width, textAlign: c.align ?? 'left' }"
          >
            {{ c.title }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td :colspan="columns.length" class="px-4 py-10 text-center text-ink-3">加载中…</td>
        </tr>
        <template v-else>
          <tr v-if="rows.length === 0">
            <td :colspan="columns.length" class="px-4 py-10">
              <Empty />
            </td>
          </tr>
          <tr
            v-for="(row, i) in rows"
            :key="String(row[rowKey as keyof T] ?? i)"
            class="border-b border-line last:border-b-0 hover:bg-page"
          >
            <td
              v-for="c in columns"
              :key="c.key"
              class="whitespace-nowrap px-4 py-3 text-ink-2"
              :style="{ textAlign: c.align ?? 'left' }"
            >
              <slot :name="c.slot ?? `col-${c.key}`" :row="row" :index="i">
                {{ cellText(row, c.key) }}
              </slot>
            </td>
          </tr>
        </template>
      </tbody>
    </table>
  </div>
</template>
