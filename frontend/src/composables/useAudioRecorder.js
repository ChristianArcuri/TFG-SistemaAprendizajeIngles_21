import { ref } from 'vue'

export function useAudioRecorder() {
  const isRecording = ref(false)
  const audioBlob = ref(null)
  const error = ref('')
  let mediaRecorder = null
  let chunks = []

  async function start() {
    error.value = ''
    audioBlob.value = null
    chunks = []

    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      const mimeType = MediaRecorder.isTypeSupported('audio/webm;codecs=opus')
        ? 'audio/webm;codecs=opus'
        : 'audio/webm'

      mediaRecorder = new MediaRecorder(stream, { mimeType })
      mediaRecorder.ondataavailable = (e) => {
        if (e.data.size > 0) chunks.push(e.data)
      }
      mediaRecorder.onstop = () => {
        stream.getTracks().forEach((t) => t.stop())
        const rawType = mediaRecorder.mimeType || 'audio/webm'
        const type = rawType.split(';')[0] || 'audio/webm'
        audioBlob.value = new Blob(chunks, { type })
        audioBlob.value.name = 'recording.webm'
      }
      mediaRecorder.start()
      isRecording.value = true
    } catch (e) {
      error.value =
        'No pudimos acceder al micrófono. Verificá permisos del navegador.'
      throw e
    }
  }

  function stop() {
    if (mediaRecorder && isRecording.value) {
      mediaRecorder.stop()
      isRecording.value = false
    }
  }

  function reset() {
    stop()
    audioBlob.value = null
    error.value = ''
    chunks = []
  }

  return { isRecording, audioBlob, error, start, stop, reset }
}
