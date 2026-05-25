/** 由团队 ID 的 SHA-256 生成 5×5 identicon（与后端用户头像算法风格一致）。 */

async function hashTeamId(teamId) {
  const data = new TextEncoder().encode(String(teamId))
  const buf = await crypto.subtle.digest('SHA-256', data)
  return new Uint8Array(buf)
}

function pickColor(bytes, offset) {
  const hue = ((bytes[offset] << 8) | bytes[offset + 1]) % 360
  const sat = 45 + (bytes[offset + 2] % 30)
  const light = 38 + (bytes[offset + 3] % 18)
  return `hsl(${hue} ${sat}% ${light}%)`
}

/**
 * 在 canvas 上绘制团队 identicon。
 * @param {HTMLCanvasElement} canvas
 * @param {string} teamId
 */
export async function drawTeamIdenticon(canvas, teamId) {
  const size = canvas.width || 48
  const bytes = await hashTeamId(teamId)
  const ctx = canvas.getContext('2d')
  const bg = pickColor(bytes, 0)
  const fg = pickColor(bytes, 4)
  ctx.fillStyle = bg
  ctx.fillRect(0, 0, size, size)
  const cell = size / 5
  const grid = 5
  for (let y = 0; y < grid; y++) {
    for (let x = 0; x < Math.ceil(grid / 2); x++) {
      const i = y * 3 + x
      const on = bytes[8 + (i % 24)] % 2 === 0
      if (!on) continue
      ctx.fillStyle = fg
      ctx.fillRect(x * cell, y * cell, cell, cell)
      const mx = grid - 1 - x
      if (mx !== x) {
        ctx.fillRect(mx * cell, y * cell, cell, cell)
      }
    }
  }
}
