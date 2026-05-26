import { computed, unref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from './useI18n'

/**
 * @param {import('vue').MaybeRef<string>|import('vue').ComputedRef<string>|null} tailLabel 二级页标题（如账本名、团队名）
 */
export function usePageCrumbs(tailLabel = null) {
  const route = useRoute()
  const { t } = useI18n()

  const crumbs = computed(() => {
    const list = []
    const parent = route.meta?.breadcrumbParent
    const titleKey = route.meta?.titleKey

    if (parent && titleKey) {
      list.push({ label: t(titleKey), to: parent })
    } else if (titleKey) {
      list.push({ label: t(titleKey) })
    }

    const tail = unref(tailLabel)
    if (tail) {
      list.push({ label: tail })
    }

    return list
  })

  return { crumbs }
}
