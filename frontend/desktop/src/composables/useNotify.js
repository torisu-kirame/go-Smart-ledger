import { useNotifyStore } from '../stores/notify'

export function useNotify() {
  const store = useNotifyStore()
  return {
    success: (text) => store.success(text),
    error: (text) => store.error(text),
    info: (text) => store.info(text),
  }
}
