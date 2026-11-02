<script setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BrandLogo from '../components/BrandLogo.vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const mode = ref('login')
const email = ref('')
const password = ref('')
const fullName = ref('')
const localError = ref('')

async function submit() {
  localError.value = ''
  try {
    if (mode.value === 'register') {
      await auth.register({
        email: email.value,
        password: password.value,
        fullName: fullName.value,
      })
    } else {
      await auth.login({ email: email.value, password: password.value })
    }
    const redirect = route.query.redirect
    if (typeof redirect === 'string' && redirect.startsWith('/')) {
      router.push(redirect)
    } else if (auth.hasProfile) {
      router.push({ name: 'dashboard' })
    } else {
      router.push({ name: 'onboarding' })
    }
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
        <span class="badge"><span class="dot" /> Tu entrenamiento continúa</span>
        <h2 class="mt-4 text-2xl font-bold">
          {{ mode === 'login' ? 'Ingresá a Inglés Profesional IT' : 'Creá tu cuenta' }}
        </h2>
        <p class="text-[#94a3b8]">
          Practicá entrevistas, dailies y comunicación laboral en inglés.
        </p>

        <div class="mt-4 flex gap-2">
          <button
            type="button"
            class="btn flex-1"
            :class="mode === 'login' ? 'btn-primary' : 'btn-secondary'"
            @click="mode = 'login'"
          >
            Ingresar
          </button>
          <button
            type="button"
            class="btn flex-1"
            :class="mode === 'register' ? 'btn-primary' : 'btn-secondary'"
            @click="mode = 'register'"
          >
            Registrarse
          </button>
        </div>

        <form class="mt-6 grid gap-4" @submit.prevent="submit">
          <div v-if="mode === 'register'">
            <label class="mb-2 block text-sm text-[#cbd5e1]">Nombre</label>
            <input v-model="fullName" class="input" placeholder="Tu nombre" />
          </div>
          <div>
            <label class="mb-2 block text-sm text-[#cbd5e1]">Email</label>
            <input v-model="email" type="email" class="input" placeholder="tu@email.com" required />
          </div>
          <div>
            <label class="mb-2 block text-sm text-[#cbd5e1]">Contraseña</label>
            <input
              v-model="password"
              type="password"
              class="input"
              placeholder="Mínimo 8 caracteres"
              required
              minlength="8"
            />
          </div>

          <p v-if="mode === 'login'" class="text-sm text-[#94a3b8]">
            Demo: admintfg@gmail.com · AdminTFG
          </p>

          <p v-if="localError" class="text-sm text-[#fb7185]">{{ localError }}</p>

          <button type="submit" class="btn btn-primary w-full" :disabled="auth.loading">
            {{ auth.loading ? 'Cargando...' : mode === 'login' ? 'Ingresar' : 'Crear cuenta' }}
          </button>
        </form>
      </section>
    </main>
  </div>
</template>
