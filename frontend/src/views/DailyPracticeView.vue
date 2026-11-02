<script setup>
import { computed, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import AppLayout from '../layouts/AppLayout.vue'
import { usePracticeStore } from '../stores/practice'
import { useAudioRecorder } from '../composables/useAudioRecorder'
import { defaultPrompts } from '../lib/categoryPrompts'

const route = useRoute()
const router = useRouter()
const practice = usePracticeStore()
const recorder = useAudioRecorder()

const categoryCode = computed(() => 'daily_meeting')
const promptText = ref(defaultPrompts.daily_meeting)
const userText = ref('')
const localError = ref('')
const analyzingAudio = ref(false)
const analyzingText = ref(false)

async function toggleRecord() {
  localError.value = ''
  if (recorder.isRecording.value) {
    recorder.stop()
    return
  }
  try {
    await recorder.start()
  } catch {
    localError.value = recorder.error.value
  }
}

async function submitAudio() {
  if (!recorder.audioBlob.value) {
    localError.value = 'Grabá tu update en inglés o usá el cuadro de texto.'
    return
  }
  localError.value = ''
  analyzingAudio.value = true
  try {
    const result = await practice.analyzeAudio({
      audioBlob: recorder.audioBlob.value,
      categoryCode: categoryCode.value,
      promptText: promptText.value,
    })
    router.push({ name: 'results', params: { sessionId: result.sessionId } })
  } catch (e) {
    localError.value = e.message
  } finally {
    analyzingAudio.value = false
  }
}

async function submitText() {
  if (!userText.value.trim()) {
    localError.value = 'Escribí lo que dirías en la daily en inglés.'
    return
  }
  localError.value = ''
  analyzingText.value = true
  try {
    const result = await practice.analyzeText({
      categoryCode: categoryCode.value,
      promptText: promptText.value,
      userText: userText.value,
    })
    router.push({ name: 'results', params: { sessionId: result.sessionId } })
  } catch (e) {
    localError.value = e.message
  } finally {
    analyzingText.value = false
  }
}
</script>

<template>
  <AppLayout>
    <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
      <div>
        <span class="badge"><span class="dot" /> Daily meeting</span>
        <h1 class="mt-3 text-4xl font-bold">Empezar daily</h1>
        <p class="text-[#94a3b8]">
          Practica qué hiciste ayer, qué harás hoy y bloqueos — hablando o por escrito.
        </p>
      </div>
      <RouterLink :to="{ name: 'dashboard' }" class="btn btn-secondary">Volver</RouterLink>
    </div>

    <div class="grid gap-5 lg:grid-cols-[1fr_360px]">
      <section class="panel flex flex-col gap-6 p-6">
        <div>
          <span class="tag">Consigna</span>
          <p class="mt-3 text-2xl font-semibold leading-tight">“{{ promptText }}”</p>
        </div>

        <div>
          <h3 class="font-semibold">Escribí tu update en inglés</h3>
          <p class="mt-1 text-sm text-[#94a3b8]">
            La IA lo corrige y te da alternativas para presentar en la reunión.
          </p>
          <textarea
            v-model="userText"
            class="input mt-3 min-h-[120px] resize-y"
            placeholder="Yesterday I fixed the login bug. Today I will work on the API tests. I have no blockers."
          />
          <button
            type="button"
            class="btn btn-primary mt-3"
            :disabled="analyzingText || practice.loading"
            @click="submitText"
          >
            {{ analyzingText ? 'Corrigiendo texto...' : 'Corregir texto con IA' }}
          </button>
        </div>

        <div class="border-t border-white/10 pt-6 text-center">
          <p class="text-sm text-[#94a3b8]">O practicá hablando como en la reunión real</p>
          <div
            class="mx-auto my-6 grid h-[100px] w-[100px] place-items-center rounded-full border border-[rgba(251,113,133,0.4)] bg-[radial-gradient(circle,rgba(251,113,133,0.28),rgba(124,58,237,0.18))] text-3xl"
            :class="{ 'animate-pulse': recorder.isRecording.value }"
          >
            🎙️
          </div>
          <div class="flex flex-wrap justify-center gap-3">
            <button type="button" class="btn btn-secondary" @click="toggleRecord">
              {{ recorder.isRecording.value ? 'Detener' : 'Grabar daily' }}
            </button>
            <button
              type="button"
              class="btn btn-primary"
              :disabled="!recorder.audioBlob.value || analyzingAudio"
              @click="submitAudio"
            >
              {{ analyzingAudio ? 'Analizando audio...' : 'Enviar grabación' }}
            </button>
          </div>
        </div>
        <p v-if="localError" class="text-sm text-[#fb7185]">{{ localError }}</p>
      </section>

      <aside class="panel p-5">
        <h3 class="font-semibold">Estructura sugerida</h3>
        <ul class="mt-4 grid gap-3 text-sm text-[#cbd5e1]">
          <li class="rounded-2xl border border-white/10 bg-white/[0.05] p-3">
            <strong>Yesterday:</strong> qué terminaste
          </li>
          <li class="rounded-2xl border border-white/10 bg-white/[0.05] p-3">
            <strong>Today:</strong> en qué vas a trabajar
          </li>
          <li class="rounded-2xl border border-white/10 bg-white/[0.05] p-3">
            <strong>Blockers:</strong> impedimentos (si hay)
          </li>
        </ul>
      </aside>
    </div>
  </AppLayout>
</template>
