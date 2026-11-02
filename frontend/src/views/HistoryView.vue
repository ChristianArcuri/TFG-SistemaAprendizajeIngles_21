<script setup>
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '../layouts/AppLayout.vue'
import ProgressStats from '../components/ProgressStats.vue'
import { usePracticeStore } from '../stores/practice'
import { categoryTitles } from '../lib/categoryPrompts'
import { computeProgressStats } from '../lib/progressStats'

const practice = usePracticeStore()
const router = useRouter()

onMounted(() => {
  practice.loadHistory()
})

const stats = computed(() => computeProgressStats(practice.sessions))

function openSession(id) {
  router.push({ name: 'results', params: { sessionId: id } })
}

function labelFor(code) {
  return categoryTitles[code] || code
}
</script>

<template>
  <AppLayout>
    <div class="mb-6">
      <span class="badge"><span class="dot" /> Progreso</span>
      <h1 class="mt-3 text-4xl font-bold">Tu progreso</h1>
      <p class="text-[#94a3b8]">Métricas y sesiones de práctica recientes.</p>
    </div>

    <ProgressStats
      :streak-days="stats.streakDays"
      :comprehension-pct="stats.comprehensionPct"
      :vocabulary-pct="stats.vocabularyPct"
      :communication-pct="stats.communicationPct"
    />

    <h2 class="mb-4 text-xl font-semibold">Sesiones recientes</h2>

    <p v-if="practice.loading" class="text-[#94a3b8]">Cargando historial...</p>
    <p v-else-if="practice.error" class="text-[#fb7185]">{{ practice.error }}</p>
    <p v-else-if="!practice.sessions.length" class="rounded-[22px] border border-white/10 bg-white/[0.04] p-6 text-[#94a3b8]">
      Todavía no tenés sesiones guardadas. Elegí un entrenamiento en el dashboard para empezar.
    </p>

    <section v-else class="grid gap-4">
      <article
        v-for="s in practice.sessions"
        :key="s.id"
        class="panel flex flex-wrap items-center justify-between gap-4 p-5"
      >
        <div>
          <span class="tag">{{ s.categoryCode }}</span>
          <h3 class="mt-2 font-semibold">{{ labelFor(s.categoryCode) }}</h3>
          <p class="mt-1 text-sm text-[#94a3b8]">
            {{ new Date(s.createdAt).toLocaleString('es-AR') }}
          </p>
          <p v-if="s.scores" class="mt-2 text-sm text-[#22d3ee]">
            Comprensión {{ s.scores.comprehension ?? '—' }}% · Vocabulario
            {{ s.scores.vocabulary ?? '—' }}%
            <template v-if="typeof s.scores.communication === 'number'">
              · Comunicación {{ s.scores.communication }}%
            </template>
          </p>
          <p v-if="s.userTranscript" class="mt-2 max-w-2xl truncate text-sm italic text-[#cbd5e1]">
            “{{ s.userTranscript }}”
          </p>
        </div>
        <button type="button" class="btn btn-secondary" @click="openSession(s.id)">Ver feedback</button>
      </article>
    </section>
  </AppLayout>
</template>
