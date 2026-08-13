/** 轻量 Markdown → 安全 HTML（先转义再解析常用语法） */
export function escapeHtml(text) {
  return String(text || '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function renderInline(text) {
  let s = escapeHtml(text)
  s = s.replace(/`([^`]+)`/g, '<code class="md-code">$1</code>')
  s = s.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
  s = s.replace(/__([^_]+)__/g, '<strong>$1</strong>')
  s = s.replace(/(?<!\*)\*([^*]+)\*(?!\*)/g, '<em>$1</em>')
  s = s.replace(/(?<!_)_([^_]+)_(?!_)/g, '<em>$1</em>')
  s = s.replace(
    /\[([^\]]+)\]\((https?:\/\/[^)\s]+)\)/g,
    '<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>'
  )
  return s
}

/**
 * @param {string} markdown
 * @returns {string} HTML
 */
export function renderMarkdown(markdown) {
  const src = String(markdown || '').replace(/\r\n/g, '\n')
  if (!src.trim()) return ''

  const blocks = []
  const lines = src.split('\n')
  let i = 0

  while (i < lines.length) {
    const line = lines[i]

    // fenced code
    const fence = line.match(/^```(\w*)\s*$/)
    if (fence) {
      const lang = fence[1] || ''
      const code = []
      i++
      while (i < lines.length && !/^```\s*$/.test(lines[i])) {
        code.push(lines[i])
        i++
      }
      i++ // skip closing ```
      blocks.push(
        `<pre class="md-pre"><code class="md-code-block"${lang ? ` data-lang="${escapeHtml(lang)}"` : ''}>${escapeHtml(code.join('\n'))}</code></pre>`
      )
      continue
    }

    // heading
    const heading = line.match(/^(#{1,6})\s+(.+)$/)
    if (heading) {
      const level = heading[1].length
      blocks.push(`<h${level} class="md-h">${renderInline(heading[2].trim())}</h${level}>`)
      i++
      continue
    }

    // hr
    if (/^(-{3,}|\*{3,}|_{3,})\s*$/.test(line)) {
      blocks.push('<hr class="md-hr" />')
      i++
      continue
    }

    // table
    if (line.includes('|') && i + 1 < lines.length && /^\s*\|?[\s:-]+\|/.test(lines[i + 1])) {
      const rows = []
      while (i < lines.length && lines[i].includes('|')) {
        rows.push(lines[i])
        i++
      }
      if (rows.length >= 2) {
        const parseRow = (row) =>
          row
            .replace(/^\s*\|/, '')
            .replace(/\|\s*$/, '')
            .split('|')
            .map((c) => c.trim())
        const header = parseRow(rows[0])
        const body = rows.slice(2).map(parseRow)
        let html = '<table class="md-table"><thead><tr>'
        header.forEach((c) => {
          html += `<th>${renderInline(c)}</th>`
        })
        html += '</tr></thead><tbody>'
        body.forEach((cols) => {
          html += '<tr>'
          cols.forEach((c) => {
            html += `<td>${renderInline(c)}</td>`
          })
          html += '</tr>'
        })
        html += '</tbody></table>'
        blocks.push(html)
        continue
      }
    }

    // unordered list
    if (/^\s*[-*+]\s+/.test(line)) {
      const items = []
      while (i < lines.length && /^\s*[-*+]\s+/.test(lines[i])) {
        items.push(lines[i].replace(/^\s*[-*+]\s+/, ''))
        i++
      }
      blocks.push(
        `<ul class="md-ul">${items.map((it) => `<li>${renderInline(it)}</li>`).join('')}</ul>`
      )
      continue
    }

    // ordered list
    if (/^\s*\d+\.\s+/.test(line)) {
      const items = []
      while (i < lines.length && /^\s*\d+\.\s+/.test(lines[i])) {
        items.push(lines[i].replace(/^\s*\d+\.\s+/, ''))
        i++
      }
      blocks.push(
        `<ol class="md-ol">${items.map((it) => `<li>${renderInline(it)}</li>`).join('')}</ol>`
      )
      continue
    }

    // blank
    if (!line.trim()) {
      i++
      continue
    }

    // paragraph (merge consecutive non-empty non-special lines)
    const para = []
    while (
      i < lines.length &&
      lines[i].trim() &&
      !/^```/.test(lines[i]) &&
      !/^#{1,6}\s/.test(lines[i]) &&
      !/^\s*[-*+]\s+/.test(lines[i]) &&
      !/^\s*\d+\.\s+/.test(lines[i]) &&
      !/^(-{3,}|\*{3,}|_{3,})\s*$/.test(lines[i])
    ) {
      para.push(lines[i])
      i++
    }
    blocks.push(`<p class="md-p">${renderInline(para.join(' '))}</p>`)
  }

  return blocks.join('')
}
