import { defineStore } from 'pinia'

/** Workspace pages that can open in editor-style left/right split panes. */
export const SPLITTABLE_TABS = [
  { id: 'assistant', path: '/assistant', labelKey: 'layout.nav.assistant', icon: 'sparkles' },
  { id: 'ledgers', path: '/ledgers', labelKey: 'layout.nav.ledgers', icon: 'ledger' },
]

export function tabFromPath(path) {
  if (!path) return null
  if (path === '/assistant' || path.startsWith('/assistant/')) {
    return SPLITTABLE_TABS.find((t) => t.id === 'assistant')
  }
  if (path === '/ledgers' || path.startsWith('/ledgers/')) {
    // Detail routes exit split; only list participates as a split pane root
    if (path === '/ledgers') return SPLITTABLE_TABS.find((t) => t.id === 'ledgers')
    return null
  }
  return null
}

export const useWorkspacePanes = defineStore('workspacePanes', {
  state: () => ({
    split: false,
    leftId: null,
    rightId: null,
    focus: 'left',
    /** @type {string[]} */
    openIds: [],
  }),
  getters: {
    leftTab(state) {
      return SPLITTABLE_TABS.find((t) => t.id === state.leftId) || null
    },
    rightTab(state) {
      return SPLITTABLE_TABS.find((t) => t.id === state.rightId) || null
    },
    openTabs(state) {
      return state.openIds
        .map((id) => SPLITTABLE_TABS.find((t) => t.id === id))
        .filter(Boolean)
    },
  },
  actions: {
    touchTab(id) {
      if (!SPLITTABLE_TABS.some((t) => t.id === id)) return
      if (!this.openIds.includes(id)) this.openIds = [...this.openIds, id]
    },
    setFocus(side) {
      this.focus = side === 'right' ? 'right' : 'left'
    },
    openSingle(id) {
      this.touchTab(id)
      this.split = false
      this.leftId = id
      this.rightId = null
      this.focus = 'left'
    },
    placeOnSide(id, side) {
      this.touchTab(id)
      if (side === 'right') {
        this.rightId = id
        if (!this.leftId || this.leftId === id) {
          const other = SPLITTABLE_TABS.find((t) => t.id !== id)
          this.leftId = other?.id || id
        }
        this.focus = 'right'
      } else {
        this.leftId = id
        if (!this.rightId || this.rightId === id) {
          const other = SPLITTABLE_TABS.find((t) => t.id !== id)
          this.rightId = other?.id || null
        }
        this.focus = 'left'
      }
      this.split = !!(this.leftId && this.rightId && this.leftId !== this.rightId)
      if (!this.split) {
        this.leftId = id
        this.rightId = null
        this.focus = 'left'
      }
    },
    exitSplit(keepId) {
      const id = keepId || this.leftId || this.rightId
      this.split = false
      this.leftId = id
      this.rightId = null
      this.focus = 'left'
    },
    syncFromRoute(path) {
      const tab = tabFromPath(path)
      if (!tab) {
        if (this.split && path.startsWith('/ledgers/')) {
          this.exitSplit('ledgers')
        }
        return
      }
      this.touchTab(tab.id)
      if (!this.split) {
        this.leftId = tab.id
      } else if (this.focus === 'right') {
        this.rightId = tab.id
      } else {
        this.leftId = tab.id
      }
    },
  },
})
