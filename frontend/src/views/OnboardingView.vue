<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import BrandLogo from '../components/BrandLogo.vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()

const goal = ref('Pasar entrevistas remotas')
const role = ref('Backend Developer')
const englishLevel = ref('B1')
const mainStack = ref('Go, PostgreSQL, Docker')
const localError = ref('')

const goals = [
  'Preparar entrevistas con Recursos Humanos',
  'Comunicarme en daily meetings',
  'Reforzar vocabulario técnico de IT',
]

const roles = [
  'Backend Developer',
  'Frontend Developer',
  'Fullstack Developer',
  'QA / Tester',
  'DevOps / Cloud',
]

const levels = [
  { value: 'A2', label: 'A2 - básico' },
  { value: 'B1', label: 'B1 - intermedio' },
  { value: 'B2', label: 'B2 - puedo comunicarme' },
  { value: 'C1', label: 'C1 - quiero sonar más natural' },
]

async function submit() {
  localError.value = ''
  try {
    await auth.saveProfile({
      goal: goal.value,
      role: role.value,
      englishLevel: englishLevel.value,
      mainStack: mainStack.value,
    })
    router.push({ name: 'dashboard' })
  } catch (e) {
    localError.value = e.message
  }
}
</script>

<template>
  <div>
    <nav class="mx-auto flex w-full max-w-[1180px] px-4 py-6">
      <BrandLogo />
    </nav>

    <main class="mx-auto flex min-h-[calc(100vh-92px)] max-w-[460px] items-center px-4 py-8">
      <section class="panel w-full p-8">
        <span class="badge"><span class="dot" /> Plan personalizado</span>
        <h2 class="mt-4 text-2xl font-bold">Armemos tu entrenamiento</h2>
        <p class="text-[#94a3b8]">
          Las instrucciones estarán en español, pero vas a practicar frases reales en inglés.
        </p>

        <form class="mt-6 grid gap-4" @submit.prevent="submit">
          <div>
            <label class="mb-2 block text-sm text-[#cbd5e1]">¿Cuál es tu objetivo principal?</label>
            <select v-model="goal" class="select">
              <option v-for="g in goals" :key="g" :value="g">{{ g }}</option>
            </select>
          </div>
          <div>
            <label class="mb-2 block text-sm text-[#cbd5e1]">Tu rol</label>
            <select v-model="role" class="select">
              <option v-for="r in roles" :key="r" :value="r">{{ r }}</option>
            </select>
          </div>
          <div>
            <label class="mb-2 block text-sm text-[#cbd5e1]">Nivel aproximado de inglés</label>
            <select v-model="englishLevel" class="select">
              <option v-for="l in levels" :key="l.value" :value="l.value">{{ l.label }}</option>
            </select>
          </div>
          <div>
            <label class="mb-2 block text-sm text-[#cbd5e1]">Stack principal</label>
            <input v-model="mainStack" class="input" />
          </div>

          <p v-if="localError" class="text-sm text-[#fb7185]">{{ localError }}</p>

          <button type="submit" class="btn btn-primary w-full" :disabled="auth.loading">
            {{ auth.loading ? 'Guardando...' : 'Crear mi dashboard' }}
          </button>
        </form>
      </section>
    </main>
  </div>
</template>
