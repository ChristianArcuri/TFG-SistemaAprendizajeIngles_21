<script setup>
import { ref } from 'vue'

const tab = ref('interview')

const demos = {
  interview: {
    label: 'Entrevista RRHH',
    title: '“Tell me about yourself and your experience.”',
    body: 'Tu respuesta',
    draft: '“I am developer and I work with backend...”',
    improved: '“I am a backend developer and I have worked with APIs and databases.”',
    meter: 78,
  },
  daily: {
    label: 'Daily meeting',
    title: 'Contá qué hiciste, qué vas a hacer y si tenés un bloqueo',
    body: 'Tu respuesta',
    draft: '“Yesterday I work in the API and today I continue.”',
    improved: '“Yesterday I worked on the API. Today I will finish the tests. I have no blockers.”',
    meter: 74,
  },
  quiz: {
    label: 'Quiz IT',
    title: 'Elegí la frase más natural',
    body: 'Situación',
    draft: 'Estoy trabajando en un bug de base de datos',
    improved: 'B. I am working on a database bug.',
    meter: 100,
  },
}
</script>

<template>
  <section id="how" class="mx-auto max-w-[1180px] px-4 py-16">
    <h2 class="text-center text-3xl font-bold">Cómo funciona</h2>
    <p class="mx-auto mt-3 max-w-2xl text-center text-[#94a3b8]">
      Elegís el entrenamiento, respondés como en el trabajo real y la IA te corrige con
      explicación en español.
    </p>

    <div class="mt-10 grid gap-4 md:grid-cols-3">
      <article class="rounded-[22px] border border-white/10 bg-white/[0.06] p-5">
        <h3 class="font-semibold">1. Elegís qué practicar</h3>
        <p class="mt-2 text-sm text-[#94a3b8]">
          Entrevista con Recursos Humanos, daily meeting o quiz de inglés IT.
        </p>
      </article>
      <article class="rounded-[22px] border border-white/10 bg-white/[0.06] p-5">
        <h3 class="font-semibold">2. Respondés en serio</h3>
        <p class="mt-2 text-sm text-[#94a3b8]">
          Podés hablar, escribir o elegir opciones según el ejercicio.
        </p>
      </article>
      <article class="rounded-[22px] border border-white/10 bg-white/[0.06] p-5">
        <h3 class="font-semibold">3. La IA te corrige</h3>
        <p class="mt-2 text-sm text-[#94a3b8]">
          Versión natural, explicación en español y próximos pasos recomendados.
        </p>
      </article>
    </div>

    <div class="mt-12 flex flex-wrap justify-center gap-2">
      <button
        v-for="(demo, key) in demos"
        :key="key"
        type="button"
        class="btn text-sm"
        :class="tab === key ? 'btn-primary' : 'btn-ghost'"
        @click="tab = key"
      >
        {{ demo.label }}
      </button>
    </div>

    <div class="panel mx-auto mt-6 max-w-2xl p-4">
      <div class="overflow-hidden rounded-[18px] border border-white/10 bg-[rgba(7,10,18,0.62)]">
        <div class="flex gap-2 border-b border-white/10 p-3">
          <span class="h-2.5 w-2.5 rounded-full bg-white/20" />
          <span class="h-2.5 w-2.5 rounded-full bg-white/20" />
          <span class="h-2.5 w-2.5 rounded-full bg-white/20" />
        </div>
        <div class="p-5">
          <p class="text-sm text-[#94a3b8]">{{ demos[tab].body }}</p>
          <h3 class="mt-2 text-xl font-semibold">{{ demos[tab].title }}</h3>
          <div class="mt-4 rounded-[18px] border border-white/10 bg-white/[0.06] p-4">
            <p class="text-sm text-[#94a3b8]">Borrador / opción</p>
            <p class="mt-1">{{ demos[tab].draft }}</p>
          </div>
          <div class="mt-4 rounded-[18px] border border-white/10 bg-white/[0.06] p-4">
            <p class="text-sm text-[#94a3b8]">Versión mejorada</p>
            <p class="mt-1">{{ demos[tab].improved }}</p>
          </div>
          <p class="mt-4 text-sm text-[#94a3b8]">Claridad</p>
          <div class="meter mt-2"><span :style="{ width: `${demos[tab].meter}%` }" /></div>
        </div>
      </div>
    </div>
  </section>
</template>
