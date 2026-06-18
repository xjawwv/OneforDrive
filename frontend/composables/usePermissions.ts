export const usePermissions = () => {
  const permissions = useState<string[]>('permissions', () => {
    if (import.meta.client) {
      try {
        const stored = localStorage.getItem('permissions')
        return stored ? JSON.parse(stored) : []
      } catch { return [] }
    }
    return []
  })

  const fetchPermissions = async () => {
    const { apiFetch } = useApi()
    try {
      const data = await apiFetch('/api/rbac/me/permissions') as any
      permissions.value = data.permissions || []
      if (import.meta.client) {
        localStorage.setItem('permissions', JSON.stringify(permissions.value))
      }
    } catch {}
  }

  const can = (permissionKey: string) => permissions.value.includes(permissionKey)

  return { permissions, fetchPermissions, can }
}
