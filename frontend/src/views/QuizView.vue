<script setup>
import { onMounted, ref, watch } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import AppLayout from '../layouts/AppLayout.vue'
import { useQuizSession } from '../composables/useQuizSession'
import { submitQuizResult } from '../api/practice'
import quizBank from '../data/prompts/it_quiz.json'

const router = useRouter()

const {
  subtypes,
  sessionSize,
  activeSubtypeId,
  sessionQuestions,
  sessionIndex,
  sessionStarted,
  sessionFinished,
  question,
  selected,
  answered,
  correctCount,
  answersLog,
  isLastInSession,
  progressPct,
  pool,
  startSession,
  selectSubtype,
  chooseOption,
  nextAfterAnswer,
  skipQuestion,
} = useQuizSession(quizBank)

const saving = ref(false)
const localError = ref('')
const savedSessionId = ref('')
const progressNote = ref('')

onMounted(() => {
  startSession()
})

watch(sessionFinished, (done) => {
  if (!done) {
    savedSessionId.value = ''
    progressNote.value = ''
    localError.value = ''
    return
  }
  saveQuiz()
})

function missedTopics() {
  const seen = new Set()
  const topics = []
  for (const answer of answersLog.value) {
    if (answer.ok) continue
    const label = topicLabel(answer.subtypeId)
    if (seen.has(label)) continue
    seen.add(label)
    topics.push(label)
  }
  return topics
}

async function saveQuiz() {
  const correct = correctCount.value
  const total = sessionQuestions.value.length
  const topics = missedTopics()
  saving.value = true
  localError.value = ''
  try {
    const result = await submitQuizResult({
      correct,
      total,
      missedTopics: topics,
    })
    if (!sessionFinished.value) return
    savedSessionId.value = result.sessionId
    progressNote.value = result.explanationEs
  } catch (e) {
    if (!sessionFinished.value) return
    localError.value = e.message
  } finally {
    if (sessionFinished.value) saving.value = false
  }
}

function openResult() {
  if (!savedSessionId.value) return
  router.push({ name: 'results', params: { sessionId: savedSessionId.value } })
}

function countForSubtype(subtypeId) {
  return pool.value.filter((q) => q.subtypeId === subtypeId).length
}

function topicLabel(subtypeId) {
  return subtypes.find((s) => s.id === subtypeId)?.label ?? 'Quiz IT'
}

</script>

<template>
  <AppLayout>
    <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
      <div>
        <span class="badge"><span class="dot" /> Entrenamiento rápido</span>
        <h1 class="mt-3 text-4xl font-bold">Quiz de inglés IT</h1>
        <p class="text-[#94a3b8]">
          Sesiones de {{ sessionSize }} preguntas — elegí la frase más natural en inglés.
        </p>
      </div>
      <RouterLink :to="{ name: 'dashboard' }" class="btn btn-secondary">Volver</RouterLink>
    </div>

    <section v-if="sessionFinished" class="panel mx-auto max-w-xl p-8 text-center">
      <h2 class="text-3xl font-bold">Resumen de la sesión</h2>
      <p class="mt-3 text-5xl font-bold text-[#22d3ee]">
        {{ correctCount }} / {{ sessionQuestions.length }}
      </p>
      <p class="mt-2 text-[#94a3b8]">
        {{ answersLog.filter((a) => a.skipped).length }} omitidas ·
        {{ quizBank.questions.length }} preguntas en el banco
      </p>
      <p v-if="saving" class="mt-4 text-[#94a3b8]">Guardando el resultado en tu progreso...</p>
      <p v-else-if="progressNote" class="mt-4 text-[#cbd5e1]">{{ progressNote }}</p>
      <div class="mt-6 flex flex-wrap justify-center gap-3">
        <button type="button" class="btn btn-primary" :disabled="!savedSessionId" @click="openResult">
          Ver resultado
        </button>
        <button v-if="localError" type="button" class="btn btn-secondary" @click="saveQuiz">
          Reintentar
        </button>
        <button type="button" class="btn btn-secondary" :disabled="saving" @click="startSession()">
          Nueva sesión
        </button>
      </div>
      <p v-if="localError" class="mt-4 text-sm text-[#fb7185]">{{ localError }}</p>
    </section>

    <div v-else-if="sessionStarted && question" class="grid gap-5 lg:grid-cols-[1fr_360px]">
      <section class="panel p-6">
        <div class="mb-4">
          <div class="mb-2 flex items-center justify-between text-sm text-[#94a3b8]">
            <span>Pregunta {{ sessionIndex + 1 }} de {{ sessionQuestions.length }}</span>
            <span>Aciertos: {{ correctCount }}</span>
          </div>
          <div class="meter"><span :style="{ width: `${progressPct}%` }" /></div>
        </div>

        <span class="tag">{{ topicLabel(question.subtypeId) }}</span>
        <p class="mt-3 text-2xl font-semibold">{{ question.situationEs }}</p>
        <p class="mt-2 text-sm text-[#94a3b8]">{{ question.hintEs }}</p>

        <div class="mt-6 grid gap-3">
          <button
            v-for="(opt, i) in question.options"
            :key="i"
            type="button"
            class="rounded-2xl border p-4 text-left transition"
            :class="[
              answered && i === question.correctIndex
                ? 'border-[#22c55e]/50 bg-[#22c55e]/10'
                : answered && i === selected && i !== question.correctIndex
                  ? 'border-[#fb7185]/50 bg-[#fb7185]/10'
                  : 'border-white/10 bg-white/[0.05] hover:border-white/20',
              !answered ? 'cursor-pointer' : 'cursor-default',
            ]"
            :disabled="answered"
            @click="chooseOption(i)"
          >
            {{ String.fromCharCode(65 + i) }}. {{ opt }}
          </button>
        </div>

        <div class="mt-6 flex flex-wrap gap-3">
          <button v-if="answered" type="button" class="btn btn-primary" @click="nextAfterAnswer()">
            {{ isLastInSession ? 'Ver resumen' : 'Siguiente pregunta' }}
          </button>
          <button v-if="!answered" type="button" class="btn btn-ghost" @click="skipQuestion()">
            Siguiente pregunta
          </button>
        </div>
      </section>

      <aside v-if="answered" class="panel p-5">
        <h3 class="font-semibold">
          {{ selected === question.correctIndex ? '¡Correcto!' : 'Respuesta correcta' }}
        </h3>
        <p class="mt-3 font-medium">
          {{ String.fromCharCode(65 + question.correctIndex) }}.
          {{ question.options[question.correctIndex] }}
        </p>
        <p class="mt-3 text-sm text-[#94a3b8]">{{ question.explanationEs }}</p>
      </aside>
      <aside v-else class="panel p-5">
        <h3 class="font-semibold">Tema del quiz</h3>
        <p class="mt-1 text-sm text-[#94a3b8]">Cambiá el tema para otra sesión de 10 preguntas.</p>
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
            <span class="mt-1 block text-xs text-[#94a3b8]">{{ countForSubtype(st.id) }} preguntas</span>
          </button>
        </div>
      </aside>
    </div>

    <p v-else class="text-[#94a3b8]">Cargando sesión...</p>
  </AppLayout>
</template>
