import { defineStore } from 'pinia'
import { ref } from 'vue'
import { APIInfo } from '../api/wails'

export const useSystemStore = defineStore('system', () => {
  const windsurfPath = ref('')
  const patchStatus = ref(false)
  const isGlobalLoading = ref(false)

  const initSystemEnvironment = async () => {
    try {
      windsurfPath.value = await APIInfo.findWindsurfPath()
      if (windsurfPath.value) {
        patchStatus.value = await APIInfo.checkPatchStatus(windsurfPath.value)
      }
    } catch (e) {
      console.error('Error init system store:', e)
    }
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
    patchStatus,
    isGlobalLoading,
    initSystemEnvironment,
    detectWindsurfPath,
    applySeamlessPatch,
    restoreSeamlessPatch,
  }
})
