<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '../layouts/AppLayout.vue'
import ScoreMeter from '../components/ScoreMeter.vue'
import { usePracticeStore } from '../stores/practice'
import { fetchSession } from '../api/practice'
import { startRouteForCategory } from '../lib/categoryConfig'
import { overallFromScores } from '../lib/progressStats'

const route = useRoute()
const router = useRouter()
const practice = usePracticeStore()

const loading = ref(true)
const error = ref('')
const result = ref(null)

const overallScore = computed(() => overallFromScores(result.value?.scores))

onMounted(async () => {
  const sessionId = route.params.sessionId
  if (practice.lastResult?.sessionId === sessionId || !sessionId) {
    result.value = practice.lastResult
    loading.value = false
    return
  }
  try {
    const detail = await fetchSession(sessionId)
    const fb = detail.feedback
    if (fb) {
      result.value = {
        sessionId: detail.id,
        categoryCode: detail.categoryCode,
        transcript: detail.userTranscript,
        correctedText: fb.correctedText,
        naturalVersion: fb.naturalVersion,
        explanationEs: fb.explanationEs,
        scores: {
          comprehension: fb.comprehensionScore,
          vocabulary: fb.vocabularyScore,
          communication: fb.communicationScore,
        },
        strengths: [],
        improvements: [],
        suggestedPhrases: [],
      }
    } else {
      error.value = 'Esta sesión no tiene feedback todavía.'
    }
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})

function practiceAgain() {
  const code = result.value?.categoryCode || 'hr_interview'
  router.push(startRouteForCategory(code))
}
</script>

<template>
  <AppLayout>
    <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
      <div>
        <span class="badge"><span class="dot" /> Resultado de la sesión</span>
        <h1 class="mt-3 text-4xl font-bold">Resultado de tu práctica</h1>
      </div>
      <button type="button" class="btn btn-primary" @click="practiceAgain">Practicar otra vez</button>
    </div>

    <p v-if="loading" class="text-[#94a3b8]">Cargando resultado...</p>
    <p v-else-if="error" class="text-[#fb7185]">{{ error }}</p>

    <template v-else-if="result">
      <section class="grid gap-5 md:grid-cols-2">
        <article class="panel p-5">
          <span class="text-[#94a3b8]">Score general</span>
          <div class="text-6xl font-black tracking-tight">{{ overallScore }}%</div>
          <p class="mt-2 text-[#94a3b8]">{{ result.explanationEs }}</p>
          <div class="meter mt-4"><span :style="{ width: `${overallScore}%` }" /></div>
        </article>

        <article class="panel p-5">
          <h3 class="font-semibold">Scores</h3>
          <div class="mt-4 grid gap-3">
            <ScoreMeter
              v-if="typeof result.scores.comprehension === 'number'"
              label="Comprensión"
              :value="result.scores.comprehension"
            />
            <ScoreMeter
              v-if="typeof result.scores.vocabulary === 'number'"
              label="Vocabulario profesional"
              :value="result.scores.vocabulary"
            />
            <ScoreMeter
              v-if="typeof result.scores.communication === 'number'"
              label="Comunicación profesional"
              :value="result.scores.communication"
            />
          </div>
          <p class="mt-3 text-sm text-[#94a3b8]">
            La pronunciación y la entonación no se puntúan a partir del texto. La gramática aparece como observación en la explicación.
          </p>
        </article>

        <article class="panel p-5">
          <h3 class="font-semibold">Tu frase (transcripción)</h3>
          <p class="mt-3 text-[#94a3b8] italic">“{{ result.transcript }}”</p>
          <h3 class="mt-6 font-semibold">Versión corregida</h3>
          <p class="mt-2">{{ result.correctedText }}</p>
        </article>

        <article class="panel p-5">
          <h3 class="font-semibold">Versión más natural</h3>
          <p class="mt-3 text-lg">{{ result.naturalVersion }}</p>

          <template v-if="result.strengths?.length">
            <h3 class="mt-6 font-semibold">Fortalezas</h3>
            <ul class="mt-2 list-disc pl-5 text-[#cbd5e1]">
              <li v-for="(s, i) in result.strengths" :key="i">{{ s }}</li>
            </ul>
          </template>

          <template v-if="result.improvements?.length">
            <h3 class="mt-6 font-semibold">Mejoras</h3>
            <ul class="mt-2 list-disc pl-5 text-[#cbd5e1]">
              <li v-for="(s, i) in result.improvements" :key="i">{{ s }}</li>
            </ul>
          </template>
        </article>

        <article v-if="result.suggestedPhrases?.length" class="panel p-5 md:col-span-2">
          <h3 class="font-semibold">Frases sugeridas</h3>
          <div class="mt-4 grid gap-3 md:grid-cols-2">
            <div
              v-for="(p, i) in result.suggestedPhrases"
              :key="i"
              class="rounded-2xl border border-white/10 bg-white/[0.05] p-4"
            >
              <p class="text-sm text-[#94a3b8]">{{ p.spanish }}</p>
              <p class="mt-1 font-medium">{{ p.english }}</p>
            </div>
          </div>
        </article>
      </section>
    </template>
  </AppLayout>
</template>
