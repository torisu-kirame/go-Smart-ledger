import { defineStore } from 'pinia'

let toastSeq = 0

/**
 * @typedef {'success' | 'error' | 'info'} ToastType
 * @typedef {{ id: number, type: ToastType, text: string, createdAt: number }} ToastItem
 * @typedef {{ id: number, level: string, text: string, at: string }} LogEntry
 */

export const useNotifyStore = defineStore('notify', {
  state: () => ({
    /** @type {ToastItem[]} newest first */
    toasts: [],
    /** @type {LogEntry[]} newest first */
    logs: [],
  }),

  actions: {
    pushToast(type, text) {
      const t = String(text || '').trim()
      if (!t) return
      const id = ++toastSeq
      this.toasts.unshift({ id, type, text: t, createdAt: Date.now() })
      this.addLog(type === 'success' ? 'success' : 'error', t)
      window.setTimeout(() => this.dismissToast(id), 1000)
    },

    dismissToast(id) {
      this.toasts = this.toasts.filter((x) => x.id !== id)
    },

    success(text) {
      this.pushToast('success', text)
    },

    error(text) {
      this.pushToast('error', text)
    },

    info(text) {
      this.pushToast('info', text)
    },

    addLog(level, text) {
      const t = String(text || '').trim()
      if (!t) return
      this.logs.unshift({
        id: ++toastSeq,
        level,
        text: t,
        at: new Date().toISOString(),
      })
      if (this.logs.length > 500) {
        this.logs.length = 500
      }
    },

    clearLogs() {
      this.logs = []
    },
  },
})
