/** @typedef {{ t: 'path', d: string }} PathShape */
/** @typedef {{ t: 'line', x1: number, y1: number, x2: number, y2: number }} LineShape */
/** @typedef {{ t: 'polyline', points: string }} PolylineShape */
/** @typedef {{ t: 'circle', cx: number, cy: number, r: number }} CircleShape */
/** @typedef {PathShape | LineShape | PolylineShape | CircleShape} IconShape */

/** @param {string} d @returns {PathShape} */
function p(d) {
  return { t: 'path', d }
}

/** @param {number} x1 @param {number} y1 @param {number} x2 @param {number} y2 @returns {LineShape} */
function l(x1, y1, x2, y2) {
  return { t: 'line', x1, y1, x2, y2 }
}

/** @param {string} points @returns {PolylineShape} */
function pl(points) {
  return { t: 'polyline', points }
}

/** @param {number} cx @param {number} cy @param {number} r @returns {CircleShape} */
function c(cx, cy, r) {
  return { t: 'circle', cx, cy, r }
}

/** @type {Record<string, IconShape[]>} */
export const ICON_REGISTRY = {
  brand: [
    p('M4 6.5h16v11H4z'),
    p('M8 6.5V5a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v1.5'),
    l(12, 10, 12, 14),
    l(10, 12, 14, 12),
  ],
  home: [p('M3 10.5 12 4l9 6.5V20a1 1 0 0 1-1 1h-5v-6H9v6H4a1 1 0 0 1-1-1v-9.5z')],
  ledger: [
    p('M6 4h9l3 3v13a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1V5a1 1 0 0 1 1-1z'),
    l(14, 4, 14, 7),
    l(9, 4, 9, 7),
    l(8, 12, 16, 12),
    l(8, 16, 14, 16),
  ],
  template: [
    p('M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z'),
    p('M14 2v6h6'),
    l(8, 13, 16, 13),
    l(8, 17, 13, 17),
  ],
  import: [p('M12 3v12'), p('M7 10l5 5 5-5'), p('M5 21h14')],
  backup: [
    p('M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4'),
    p('M17 8l-5-5-5 5'),
    p('M12 3v12'),
  ],
  friends: [p('M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2'), c(12, 7, 4)],
  teams: [
    p('M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2'),
    c(9, 7, 4),
    p('M23 21v-2a4 4 0 0 0-3-3.87'),
    p('M16 3.13a4 4 0 0 1 0 7.75'),
  ],
  chain: [
    p('M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71'),
    p('M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71'),
  ],
  settings: [
    p('M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z'),
    p(
      'M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z'
    ),
  ],
  eye: [p('M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z'), c(12, 12, 3)],
  logout: [p('M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4'), pl('16 17 21 12 16 7'), l(21, 12, 9, 12)],
  user: [p('M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2'), c(12, 7, 4)],
  grid: [
    p('M3 3h7v7H3z'),
    p('M14 3h7v7h-7z'),
    p('M14 14h7v7h-7z'),
    p('M3 14h7v7H3z'),
  ],
  rows: [
    l(4, 7, 20, 7),
    l(4, 12, 20, 12),
    l(4, 17, 20, 17),
  ],
  palette: [
    p('M12 2a10 10 0 1 0 10 10 4 4 0 0 1-5-5 4 4 0 0 1 5-5'),
    c(8.5, 8.5, 1.5),
    c(15.5, 8.5, 1.5),
    c(15.5, 15.5, 1.5),
    c(8.5, 15.5, 1.5),
  ],
  globe: [c(12, 12, 10), p('M2 12h20'), p('M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z')],
  sparkles: [
    p('M12 3l1.2 4.2L17.5 8.5l-4.3 1.2L12 14l-1.2-4.3L6.5 8.5l4.3-1.3L12 3z'),
    p('M19 14l.8 2.8L22.5 18l-2.7.8L19 21l-.8-2.2L15.5 18l2.7-.7L19 14z'),
    p('M5 16l.6 2.1L7.5 19l-1.9.5L5 22l-.6-2.4L2.5 19l1.9-.4L5 16z'),
  ],
  shield: [p('M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z')],
  'arrow-left': [pl('19 12 5 12'), pl('12 19 5 12 12 5')],
  'arrow-right': [pl('5 12 19 12'), pl('12 5 19 12 12 19')],
  upload: [p('M12 3v12'), p('M7 10l5-5 5 5'), p('M5 21h14')],
  paperclip: [
    p('M21.44 11.05l-9.19 9.19a6 6 0 0 1-8.49-8.49l9.19-9.19a4 4 0 0 1 5.66 5.66l-9.2 9.19a2 2 0 0 1-2.83-2.83l8.49-8.48'),
  ],
  refresh: [p('M21 12a9 9 0 1 1-2.64-6.36'), pl('21 3 21 9 15 9')],
  plus: [pl('12 5 12 19'), pl('5 12 19 12')],
  search: [c(11, 11, 8), l(21, 21, 16.65, 16.65)],
  'chevron-right': [pl('9 18 15 12 9 6')],
  'chevron-down': [pl('6 9 12 15 18 9')],
  check: [pl('20 6 9 17 4 12')],
  x: [pl('18 6 6 18'), pl('6 6 18 18')],
  send: [p('M22 2 11 13'), p('M22 2 15 22 11 13 2 9 22 2z')],
  file: [
    p('M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z'),
    p('M14 2v6h6'),
  ],
  folder: [p('M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z')],
  external: [p('M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6'), pl('15 3 21 3 21 9'), l(10, 14, 21, 3)],
  activity: [pl('22 12 18 12 15 21 9 3 6 12 2 12')],
  layers: [p('M12 2 2 7l10 5 10-5-10-5z'), p('M2 17l10 5 10-5'), p('M2 12l10 5 10-5')],
}

export const NAV_ICON_BY_ROUTE = {
  '/': 'home',
  '/assistant': 'sparkles',
  '/ledgers': 'ledger',
  '/entry-templates': 'template',
  '/backup': 'backup',
  '/friends': 'friends',
  '/teams': 'teams',
  '/chain': 'chain',
  '/logs': 'activity',
  '/settings': 'settings',
}

export const SETTINGS_ICON_BY_SECTION = {
  appearance: 'palette',
  language: 'globe',
  ai: 'sparkles',
  account: 'user',
  security: 'shield',
}
