export const MODEL_SERIES = ['deepseek', 'qwen', 'kimi', 'glm', 'other'] as const
export type ModelSeries = (typeof MODEL_SERIES)[number]

/** 系列只用于展示筛选；接入平台、调用名和价格仍由原响应持有。 */
export function modelSeries(name: string): ModelSeries {
  const model = name.trim().toLowerCase().split('/').pop() ?? ''
  if (/^deepseek(?:[-._\d]|$)/.test(model)) return 'deepseek'
  if (/^qwen(?:[-._\d]|$)/.test(model)) return 'qwen'
  if (/^kimi(?:[-._\d]|$)/.test(model)) return 'kimi'
  if (/^glm(?:[-._\d]|$)/.test(model)) return 'glm'
  return 'other'
}
