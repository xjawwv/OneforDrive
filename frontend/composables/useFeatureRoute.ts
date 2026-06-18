export const useFeatureRoute = (path: string) => {
  const enabled = useState<boolean>(`route:${path}`, () => true)
  const loading = useState<boolean>(`route:${path}:loading`, () => true)
  const description = useState<string>(`route:${path}:desc`, () => '')

  const checkRoute = async () => {
    const { apiFetch } = useApi()
    try {
      const data = await apiFetch(`/api/routes${path}`) as any
      enabled.value = data.enabled
      description.value = data.description || ''
    } catch {
      enabled.value = true
    }
    loading.value = false
  }

  return { enabled, loading, description, checkRoute }
}
