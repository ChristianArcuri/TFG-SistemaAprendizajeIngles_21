<script setup>
import { onMounted, ref, watch } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import AppLayout from '../layouts/AppLayout.vue'
import { usePracticeStore } from '../stores/practice'
import { usePracticePrompts } from '../composables/usePracticePrompts'
import { useAudioRecorder } from '../composables/useAudioRecorder'
import hrInterviewBank from '../data/prompts/hr_interview.json'

const router = useRouter()
const practice = usePracticeStore()
const recorder = useAudioRecorder()
const categoryCode = 'hr_interview'

const {
  subtypes,
  activeSubtypeId,
  activeSubtype,
  currentPrompt,
  pool,
  progressLabel,
  isSeen,
  init,
  selectSubtype: pickSubtype,
  nextPrompt,
  completeCurrent,
} = usePracticePrompts(hrInterviewBank, categoryCode)

const localError = ref('')
const analyzing = ref(false)

onMounted(() => {
  init()
})

watch(
  () => currentPrompt.value?.id,
  () => {
    recorder.reset()
    localError.value = ''
  },
)

function countForSubtype(subtypeId) {
  return pool.value.filter((p) => p.subtypeId === subtypeId).length
}

function selectSubtype(subtypeId) {
  pickSubtype(subtypeId)
  recorder.reset()
  localError.value = ''
}

function goNext() {
  nextPrompt()
  recorder.reset()
  localError.value = ''
}

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
    localError.value = 'Grabá una respuesta antes de enviar.'
    return
  }
  if (!currentPrompt.value) return

  localError.value = ''
  analyzing.value = true
  try {
    const ctx = activeSubtype.value
      ? `[${activeSubtype.value.label}] ${currentPrompt.value.title}`
      : currentPrompt.value.title
    const result = await practice.analyzeAudio({
      audioBlob: recorder.audioBlob.value,
      categoryCode,
      promptText: `${ctx}: ${currentPrompt.value.promptText}`,
    })
    completeCurrent()
    router.push({ name: 'results', params: { sessionId: result.sessionId } })
  } catch (e) {
    localError.value = e.message
  } finally {
    analyzing.value = false
  }
}
</script>

<template>
  <AppLayout>
    <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
      <div>
        <span class="badge"><span class="dot" /> Entrevista HR</span>
        <h1 class="mt-3 text-4xl font-bold">Practicar hablando</h1>
        <p class="text-[#94a3b8]">
          Preguntas clásicas de recursos humanos — respondé en inglés y recibí feedback.
        </p>
      </div>
      <RouterLink :to="{ name: 'dashboard' }" class="btn btn-secondary">Volver</RouterLink>
    </div>

    <div class="grid gap-5 lg:grid-cols-[1fr_360px]">
      <section class="panel flex min-h-[520px] flex-col justify-between p-6">
        <div>
          <div class="flex flex-wrap items-center justify-between gap-2">
            <span class="tag">{{ activeSubtype?.label ?? 'Pregunta HR' }}</span>
            <span v-if="progressLabel" class="text-sm text-[#94a3b8]">
              {{ progressLabel }}
              <span
                v-if="currentPrompt && isSeen(currentPrompt.id)"
                class="ml-1 text-[#22d3ee]"
              >
                · practicado
              </span>
            </span>
          </div>
          <h2 v-if="currentPrompt" class="mt-3 text-lg font-semibold text-[#cbd5e1]">
            {{ currentPrompt.title }}
          </h2>
          <p class="mt-3 text-2xl leading-tight font-semibold tracking-tight md:text-3xl">
            “{{ currentPrompt?.promptText }}”
          </p>
          <p v-if="currentPrompt?.hintEs" class="mt-3 text-sm text-[#94a3b8]">
            Tip: {{ currentPrompt.hintEs }}
          </p>
        </div>

        <div class="text-center">
          <div
            class="mx-auto my-8 grid h-[126px] w-[126px] place-items-center rounded-full border border-[rgba(251,113,133,0.4)] bg-[radial-gradient(circle,rgba(251,113,133,0.28),rgba(124,58,237,0.18))] text-4xl shadow-[0_0_60px_rgba(251,113,133,0.25)]"
            :class="{ 'animate-pulse': recorder.isRecording.value }"
          >
            🎙️
          </div>
          <div class="flex flex-wrap justify-center gap-3">
            <button type="button" class="btn btn-primary" @click="toggleRecord">
              {{ recorder.isRecording.value ? 'Detener grabación' : 'Grabar respuesta' }}
            </button>
            <button
              type="button"
              class="btn btn-secondary"
              :disabled="!recorder.audioBlob.value || analyzing || practice.loading"
              @click="submitAudio"
            >
              {{ analyzing || practice.loading ? 'Analizando...' : 'Enviar y ver feedback' }}
            </button>
            <button
              type="button"
              class="btn btn-ghost"
              :disabled="!currentPrompt"
              @click="goNext"
            >
              Siguiente pregunta
            </button>
          </div>
          <p v-if="recorder.audioBlob.value" class="mt-3 text-sm text-[#22c55e]">
            Audio listo para enviar
          </p>
          <p v-if="localError" class="mt-3 text-sm text-[#fb7185]">{{ localError }}</p>
        </div>
      </section>

      <aside class="panel p-5">
        <h3 class="font-semibold">Tipo de pregunta</h3>
        <p class="mt-1 text-sm text-[#94a3b8]">Elegí un bloque típico de entrevistas con HR.</p>
        <div class="mt-3 grid gap-2">
          <button
            v-for="st in subtypes"
            :key="st.id"
            type="button"
            class="rounded-2xl border p-3 text-left text-sm transition"
            :class="
              activeSubtypeId === st.id
                ? 'border-[#7c3aed]/60 bg-[rgba(124,58,237,0.15)] text-white'
                : 'border-white/10 bg-white/[0.05] text-[#cbd5e1] hover:border-white/20'
            "
            @click="selectSubtype(st.id)"
          >
            {{ st.label }}
            <span class="mt-1 block text-xs text-[#94a3b8]">
              {{ countForSubtype(st.id) }} preguntas
            </span>
          </button>
        </div>

        <h3 class="mt-6 font-semibold">Qué corrige la IA</h3>
        <ul class="mt-3 grid gap-2 text-sm text-[#94a3b8]">
          <li class="rounded-2xl border border-white/10 bg-white/[0.05] p-3">Fluidez y naturalidad</li>
          <li class="rounded-2xl border border-white/10 bg-white/[0.05] p-3">Estructura STAR en respuestas</li>
          <li class="rounded-2xl border border-white/10 bg-white/[0.05] p-3">Tono profesional y claro</li>
          <li class="rounded-2xl border border-white/10 bg-white/[0.05] p-3">Versión más natural en inglés</li>
        </ul>
      </aside>
    </div>
  </AppLayout>
</template>
