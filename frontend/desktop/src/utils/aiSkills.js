/**
 * AI 助手 Skills（豆包式技能芯片）
 * action: prompt = 注入引导话术并发送；export = 前端直接导出
 */

export const AI_SKILLS = [
  {
    id: 'record',
    icon: 'plus',
    labelKey: 'assistant.skillRecord',
    hintKey: 'assistant.skillRecordHint',
    action: 'prompt',
    needsLedger: true,
    prompt: `请帮我记一笔流水。我会按「时间、金额、用途」说明；用途尽量具体（如「午餐-张总客户」），并区分支出/收入/转账（转账不算支出）。

若信息不全请逐项追问，确认后调用工具写入当前绑定账本。回复请用 Markdown，写清分类建议。`,
  },
  {
    id: 'reconcile',
    icon: 'check',
    labelKey: 'assistant.skillReconcile',
    hintKey: 'assistant.skillReconcileHint',
    action: 'prompt',
    needsLedger: true,
    prompt: `请帮我对余额。核心是「账实相符」：对照账本汇总与银行卡/微信/支付宝余额。

请先拉取账本摘要与近期流水，给出各账户/分类余额概览，并列出可能漏记或手续费差异。用 Markdown 表格展示。`,
  },
  {
    id: 'tag',
    icon: 'layers',
    labelKey: 'assistant.skillTag',
    hintKey: 'assistant.skillTagHint',
    action: 'prompt',
    needsLedger: true,
    prompt: `请帮我贴标签（分类）。一级分类宜粗（餐饮、交通、房租、采购等），二级宜细（工作餐、家庭买菜、请客招待）。

请分析当前账本流水中的用途/备注，给出分类体系建议与待重分类清单，Markdown 输出。`,
  },
  {
    id: 'trend',
    icon: 'activity',
    labelKey: 'assistant.skillTrend',
    hintKey: 'assistant.skillTrendHint',
    action: 'prompt',
    needsLedger: true,
    prompt: `请帮我看趋势（复盘）。重点：①大额异常项（环比突然升高）；②固定开支占比（房租+工资等是否过高）。

请结合账本事件与财务报表（若有），用 Markdown 给出分类汇总、异常项与决策建议。`,
  },
  {
    id: 'budget',
    icon: 'shield',
    labelKey: 'assistant.skillBudget',
    hintKey: 'assistant.skillBudgetHint',
    action: 'prompt',
    needsLedger: true,
    prompt: `请帮我做预算管控。为主要分类设定本月上限，对照已发生支出，提示接近上限的分类。

请查询预算与分析接口（若有），用 Markdown 给出预算表、执行进度与预警。`,
  },
  {
    id: 'summary',
    icon: 'ledger',
    labelKey: 'assistant.skillSummary',
    hintKey: 'assistant.skillSummaryHint',
    action: 'prompt',
    needsLedger: true,
    prompt: `请对当前绑定账本做一份账目总结：元数据、近期流水要点、收入/支出结构、待核对项。使用 Markdown（标题、列表、表格）。`,
  },
  {
    id: 'report',
    icon: 'template',
    labelKey: 'assistant.skillReport',
    hintKey: 'assistant.skillReportHint',
    action: 'prompt',
    needsLedger: true,
    prompt: `请生成或解读当前账本的财务报表（若为专业记账模式请调用报表工具）。用 Markdown 呈现关键科目与结论，并说明数据来源。`,
  },
  {
    id: 'export',
    icon: 'download',
    labelKey: 'assistant.skillExport',
    hintKey: 'assistant.skillExportHint',
    action: 'export',
    needsLedger: true,
    prompt: '',
  },
]

export function getSkillById(id) {
  return AI_SKILLS.find((s) => s.id === id) || null
}
