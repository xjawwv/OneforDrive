export const usePermissions = () => {
  const permissions = useState<string[]>('permissions', () => [])

  const fetchPermissions = async () => {
    const { apiFetch } = useApi()
    try {
      const data = await apiFetch('/api/rbac/me/permissions') as any
      permissions.value = data.permissions || []
    } catch {}
  }

  const can = (permissionKey: string) => permissions.value.includes(permissionKey)

  return { permissions, fetchPermissions, can }
}
