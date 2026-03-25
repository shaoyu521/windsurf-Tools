import { defineStore } from 'pinia'
import { ref } from 'vue'
import { APIInfo } from '../api/wails'

export const useSystemStore = defineStore('system', () => {
  const windsurfPath = ref('')
  /** 号池 accounts.json / settings.json 所在目录（跨平台 WindsurfTools） */
  const appStoragePath = ref('')
  const patchStatus = ref(false)
  const isGlobalLoading = ref(false)
  /** 当前 windsurf_auth.json 中的邮箱（与号池比对用于「在线」高亮） */
  const currentAuthEmail = ref('')
  let authFetchInFlight: Promise<void> | null = null
  let initInFlight: Promise<void> | null = null
  let lastInitAt = 0

  const fetchCurrentAuth = async (force = false) => {
    if (!force && authFetchInFlight) {
      return authFetchInFlight
    }
    authFetchInFlight = (async () => {
      try {
        const auth = await APIInfo.getCurrentWindsurfAuth()
        currentAuthEmail.value = auth?.email ?? ''
      } catch {
        currentAuthEmail.value = ''
      } finally {
        authFetchInFlight = null
      }
    })()
    return authFetchInFlight
  }

  const initSystemEnvironment = async (force = false) => {
    const now = Date.now()
    if (!force && initInFlight) {
      return initInFlight
    }
    if (
      !force &&
      now-lastInitAt < 2500 &&
      (appStoragePath.value || windsurfPath.value || currentAuthEmail.value)
    ) {
      return
    }
    initInFlight = (async () => {
      try {
        const [storagePath, detectedWindsurfPath] = await Promise.all([
          APIInfo.getAppStoragePath().catch(() => ''),
          APIInfo.findWindsurfPath().catch(() => ''),
        ])
        appStoragePath.value = storagePath || ''
        windsurfPath.value = detectedWindsurfPath || ''
        if (windsurfPath.value) {
          patchStatus.value = await APIInfo.checkPatchStatus(windsurfPath.value)
        } else {
          patchStatus.value = false
        }
        await fetchCurrentAuth(force)
        lastInitAt = Date.now()
      } catch (e) {
        console.error('Error init system store:', e)
      } finally {
        initInFlight = null
      }
    })()
    return initInFlight
  }

  const detectWindsurfPath = async () => {
    try {
      windsurfPath.value = await APIInfo.findWindsurfPath()
      if (windsurfPath.value) {
        patchStatus.value = await APIInfo.checkPatchStatus(windsurfPath.value)
      }
      return windsurfPath.value
    } catch (e) {
      console.error('detectWindsurfPath:', e)
      return ''
    }
  }

  const applySeamlessPatch = async () => {
    if (!windsurfPath.value) return false
    try {
      isGlobalLoading.value = true
      await APIInfo.applySeamlessPatch(windsurfPath.value)
      patchStatus.value = await APIInfo.checkPatchStatus(windsurfPath.value)
      return true
    } catch (e) {
      console.error('Patch failed:', e)
      return false
    } finally {
      isGlobalLoading.value = false
    }
  }

  const restoreSeamlessPatch = async () => {
    if (!windsurfPath.value) return false
    try {
      isGlobalLoading.value = true
      await APIInfo.restoreSeamlessPatch(windsurfPath.value)
      patchStatus.value = await APIInfo.checkPatchStatus(windsurfPath.value)
      return true
    } catch (e) {
      console.error('Restore failed:', e)
      return false
    } finally {
      isGlobalLoading.value = false
    }
  }

  return {
    windsurfPath,
    appStoragePath,
    patchStatus,
    isGlobalLoading,
    currentAuthEmail,
    fetchCurrentAuth,
    initSystemEnvironment,
    detectWindsurfPath,
    applySeamlessPatch,
    restoreSeamlessPatch,
  }
})
