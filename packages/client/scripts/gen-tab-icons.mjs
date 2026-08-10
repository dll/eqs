// 生成 tabBar 图标 PNG（纯 Node，无依赖）
// 输出：src/static/tab/*.png（home/project/order/mine 各 2 种颜色）
import { deflateSync } from 'node:zlib'
import { writeFileSync, mkdirSync } from 'node:fs'
import { join } from 'node:path'

const SIZE = 81
const OUT = join(import.meta.dirname, '../src/static/tab')
mkdirSync(OUT, { recursive: true })
const CRC_TABLE = (() => {
  const t = new Uint32Array(256)
  for (let n = 0; n < 256; n++) {
    let c = n
    for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1
    t[n] = c >>> 0
  }
  return t
})()

function crc32(buf) {
  let c = 0xffffffff
  for (let i = 0; i < buf.length; i++) c = CRC_TABLE[(c ^ buf[i]) & 0xff] ^ (c >>> 8)
  return (c ^ 0xffffffff) >>> 0
}

function chunk(type, data) {
  const len = Buffer.alloc(4)
  len.writeUInt32BE(data.length)
  const body = Buffer.concat([Buffer.from(type, 'ascii'), data])
  const crc = Buffer.alloc(4)
  crc.writeUInt32BE(crc32(body))
  return Buffer.concat([len, body, crc])
}

function encodePng(rgba) {
  const ihdr = Buffer.alloc(13)
  ihdr.writeUInt32BE(SIZE, 0)
  ihdr.writeUInt32BE(SIZE, 4)
  ihdr[8] = 8  // bit depth
  ihdr[9] = 6  // color type RGBA
  const raw = Buffer.alloc((SIZE * 4 + 1) * SIZE)
  for (let y = 0; y < SIZE; y++) {
    raw[y * (SIZE * 4 + 1)] = 0
    for (let x = 0; x < SIZE; x++) {
      const src = (y * SIZE + x) * 4
      const dst = y * (SIZE * 4 + 1) + 1 + x * 4
      raw[dst] = rgba[src]
      raw[dst + 1] = rgba[src + 1]
      raw[dst + 2] = rgba[src + 2]
      raw[dst + 3] = rgba[src + 3]
    }
  }
  return Buffer.concat([
    Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]),
    chunk('IHDR', ihdr),
    chunk('IDAT', deflateSync(raw, { level: 9 })),
    chunk('IEND', Buffer.alloc(0)),
  ])
}

// 简单矢量光栅化：基于采样点的 SDF 判定图元
function render(fns) {
  const N = 8 // 超采样
  const rgba = new Uint8Array(SIZE * SIZE * 4)
  const isIn = (x, y) => fns.some((fn) => fn(x, y))
  for (let y = 0; y < SIZE; y++) {
    for (let x = 0; x < SIZE; x++) {
      let hit = 0
      for (let dy = 0; dy < N; dy++) {
        for (let dx = 0; dx < N; dx++) {
          const u = x + (dx + 0.5) / N
          const v = y + (dy + 0.5) / N
          if (isIn(u, v)) hit++
        }
      }
      const a = hit / (N * N)
      const i = (y * SIZE + x) * 4
      rgba[i] = 255
      rgba[i + 1] = 255
      rgba[i + 2] = 255
      rgba[i + 3] = Math.round(a * 255)
    }
  }
  return rgba
}

// 各图标绘制函数（81x81 栅格坐标）
const icons = {
  home: () => [
    // 屋顶三角
    (x, y) => pointInTriangle(x, y, 40, 10, 10, 38, 70, 38),
    // 房身
    (x, y) => pointInRect(x, y, 14, 38, 66, 70),
  ],
  project: () => [
    (x, y) => pointInRect(x, y, 8, 20, 72, 60),
    (x, y) => pointInRect(x, y, 24, 8, 60, 30),
  ],
  order: () => [
    (x, y) => pointInRect(x, y, 16, 8, 64, 72),
    (x, y) => pointInRect(x, y, 26, 26, 54, 32),
    (x, y) => pointInRect(x, y, 26, 42, 54, 48),
    (x, y) => pointInRect(x, y, 26, 58, 44, 64),
  ],
  mine: () => [
    // 头
    (x, y) => pointInCircle(x, y, 40, 24, 14),
    // 身体
    (x, y) => pointInRect(x, y, 24, 40, 56, 72),
  ],
}

export function pointInCircle(x, y, cx, cy, r) {
  return (x - cx) ** 2 + (y - cy) ** 2 <= r * r
}
export function pointInRect(x, y, x0, y0, x1, y1) {
  return x >= x0 && x <= x1 && y >= y0 && y <= y1
}
export function pointInEllipse(x, y, cx, cy, rx, ry) {
  return ((x - cx) / rx) ** 2 + ((y - cy) / ry) ** 2 <= 1
}
export function pointInTriangle(x, y, x1, y1, x2, y2, x3, y3) {
  const d1 = sign(x, y, x1, y1, x2, y2)
  const d2 = sign(x, y, x2, y2, x3, y3)
  const d3 = sign(x, y, x3, y3, x1, y1)
  const neg = d1 < 0 || d2 < 0 || d3 < 0
  const pos = d1 > 0 || d2 > 0 || d3 > 0
  return !(neg && pos)
}
export function sign(x, y, x1, y1, x2, y2) {
  return (x - x2) * (y1 - y2) - (x1 - x2) * (y - y2)
}

export function genCategoryIcons() {
  const dir = join(import.meta.dirname, '../src/static/category')
  mkdirSync(dir, { recursive: true })
  const catIcons = {
    price: () => [
      (x, y) => pointInRect(x, y, 10, 10, 70, 70),
    ],
    supervise: () => [
      (x, y) => pointInCircle(x, y, 40, 40, 28),
      (x, y) => pointInRect(x, y, 38, 22, 42, 58),
      (x, y) => pointInRect(x, y, 22, 38, 58, 42),
    ],
    survey: () => [
      (x, y) => pointInTriangle(x, y, 40, 8, 8, 72, 72, 72),
    ],
    design: () => [
      (x, y) => pointInCircle(x, y, 40, 40, 25),
      (x, y) => pointInCircle(x, y, 40, 40, 10),
    ],
  }
  for (const [name, fn] of Object.entries(catIcons)) {
    writeFileSync(join(dir, `${name}.png`), encodePng(render(fn())))
  }
}

for (const [name, fn] of Object.entries(icons)) {
  writeFileSync(join(OUT, `${name}.png`), encodePng(render(fn())))
  writeFileSync(join(OUT, `${name}-active.png`), encodePng(render(fn())))
}
genCategoryIcons()
console.log('done')