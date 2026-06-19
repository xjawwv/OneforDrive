export const useFeatureRoute = (path: string) => {
  const enabled = useState<boolean>(`route:${path}`, () => true)
  const loading = useState<boolean>(`route:${path}:loading`, () => true)
  const description = useState<string>(`route:${path}:desc`, () => '')

  const checkRoute = async () => {
    const { apiFetch } = useApi()
    try {
      const data = await apiFetch(`/api/routes${path}`) as any
      let isEnabled = data.enabled

      // If route is disabled, check if user's role is exempt
      if (!isEnabled && data.exempt_role_ids?.length) {
        try {
          const roleData = await apiFetch('/api/rbac/me/roles') as any
          const userRoleIDs = roleData.role_ids || []
          const isExempt = userRoleIDs.some((id: number) => data.exempt_role_ids.includes(id))
          if (isExempt) isEnabled = true
        } catch {}
      }

      enabled.value = isEnabled
      description.value = data.description || ''
    } catch {
      enabled.value = true
    }
    loading.value = false
  }

  return { enabled, loading, description, checkRoute }
}
