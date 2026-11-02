<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '../layouts/AppLayout.vue'
import ProgressStats from '../components/ProgressStats.vue'
import { useAuthStore } from '../stores/auth'
import { usePracticeStore } from '../stores/practice'
import { getCategoryConfig, startRouteForCategory } from '../lib/categoryConfig'
import { computeProgressStats } from '../lib/progressStats'
import { fetchRecommendation } from '../api/practice'

const auth = useAuthStore()
const practice = usePracticeStore()
const router = useRouter()
const recommendation = ref(null)

onMounted(async () => {
  await practice.loadCategories()
  await practice.loadHistory()
  try {
    recommendation.value = await fetchRecommendation()
  } catch {
    recommendation.value = null
  }
})

const moduleOrder = ['hr_interview', 'daily_meeting', 'it_quiz']

const stats = computed(() => computeProgressStats(practice.sessions))

const modules = computed(() =>
  [...practice.categories].sort((a, b) => {
    const left = moduleOrder.indexOf(a.code)
    const right = moduleOrder.indexOf(b.code)
    return (left === -1 ? 99 : left) - (right === -1 ? 99 : right)
  }),
)

function startPractice(code) {
  router.push(startRouteForCategory(code))
}

function btnClass(code) {
  return getCategoryConfig(code).btnVariant === 'secondary' ? 'btn btn-secondary' : 'btn btn-primary'
}

function ctaFor(code) {
  return getCategoryConfig(code).cta
}

function logout() {
  auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <AppLayout>
    <div class="mb-7 flex flex-wrap items-center justify-between gap-4">
      <div>
        <span class="badge"><span class="dot" /> Plan: {{ auth.profile?.goal || 'entrenamiento IT' }}</span>
        <h1 class="mt-3 text-4xl font-bold tracking-tight md:text-5xl">
          ¿Qué te gustaría practicar hoy?
        </h1>
        <p class="mt-2 text-[#94a3b8]">Hola, {{ auth.user?.fullName || auth.user?.email }}</p>
      </div>
      <button
        type="button"
        class="btn border border-[#22d3ee]/70 bg-[#22d3ee]/20 text-[#a5f3fc] hover:bg-[#22d3ee]/30"
        @click="logout"
      >
        Salir
      </button>
    </div>

    <ProgressStats
      :streak-days="stats.streakDays"
      :comprehension-pct="stats.comprehensionPct"
      :vocabulary-pct="stats.vocabularyPct"
      :communication-pct="stats.communicationPct"
    />

    <article v-if="recommendation" class="panel mb-8 p-5">
      <span class="tag">Refuerzo</span>
      <h2 class="mt-2 text-xl font-semibold">
        {{ recommendation.competencyLabel || 'Seguimiento por competencia' }}
      </h2>
      <p class="mt-2 text-[#cbd5e1]">{{ recommendation.message }}</p>
      <button
        v-if="recommendation.categoryCode"
        type="button"
        class="btn btn-primary mt-4"
        @click="startPractice(recommendation.categoryCode)"
      >
        Practicar {{ recommendation.categoryName }}
      </button>
    </article>

    <h2 class="mb-4 text-xl font-semibold">Módulos de práctica</h2>

    <p v-if="practice.error" class="mb-4 text-[#fb7185]">{{ practice.error }}</p>
    <div v-if="practice.loading && !practice.categories.length" class="text-[#94a3b8]">
      Cargando categorías...
    </div>

    <section v-else class="grid gap-4 lg:grid-cols-3">
      <article
        v-for="cat in modules"
        :key="cat.id"
        class="panel flex min-h-[160px] flex-col justify-between p-5"
      >
        <div>
          <span class="tag">{{ cat.code }}</span>
          <h3 class="mt-2 text-lg font-semibold">{{ cat.name }}</h3>
          <p class="mt-1 text-sm text-[#94a3b8]">{{ cat.description }}</p>
        </div>
        <button
          type="button"
          :class="[btnClass(cat.code), 'mt-4 w-fit']"
          @click="startPractice(cat.code)"
        >
          {{ ctaFor(cat.code) }}
        </button>
      </article>
    </section>
  </AppLayout>
</template>
