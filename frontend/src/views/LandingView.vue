<script setup>
import { RouterLink } from 'vue-router'
import BrandLogo from '../components/BrandLogo.vue'
import LandingDemo from '../components/LandingDemo.vue'
</script>

<template>
  <div>
    <nav class="mx-auto flex w-full max-w-[1180px] items-center justify-between px-4 py-6">
      <BrandLogo />
      <div class="hidden items-center gap-1 text-sm md:flex">
        <a class="rounded-full px-3 py-2 font-medium text-white transition hover:bg-white/10 hover:text-[#a5f3fc]" href="#features">Entrenamientos</a>
        <a class="rounded-full px-3 py-2 font-medium text-white transition hover:bg-white/10 hover:text-[#a5f3fc]" href="#how">Cómo funciona</a>
      </div>
      <div class="flex items-center gap-3">
        <RouterLink to="/login" class="btn border border-[#22d3ee]/70 bg-[#22d3ee]/15 text-[#a5f3fc] hover:bg-[#22d3ee]/25">Ingresar</RouterLink>
        <RouterLink to="/login" class="btn btn-primary">Empezar gratis</RouterLink>
      </div>
    </nav>

    <div class="mx-auto grid max-w-[1180px] items-center gap-10 px-4 py-16 md:grid-cols-2">
      <div>
        <span class="badge"><span class="dot" /> Inglés profesional para el trabajo en IT</span>
        <h1 class="mt-5 text-5xl leading-[0.96] font-bold tracking-tight md:text-7xl">
          Entrená inglés para
          <span class="gradient-text">entrevistas, dailies y vocabulario técnico.</span>
        </h1>
        <p class="mt-4 max-w-xl text-lg leading-relaxed text-[#cbd5e1]">
          Tres módulos: entrevista con Recursos Humanos, daily meeting y quiz de inglés IT.
          El sistema guarda tu historial y recomienda refuerzo según comprensión, vocabulario
          profesional y comunicación profesional.
        </p>
        <div class="mt-8 flex flex-wrap gap-3">
          <RouterLink to="/login" class="btn btn-primary">Crear mi plan</RouterLink>
          
        </div>
      </div>

      <div class="panel p-4">
        <div class="overflow-hidden rounded-[18px] border border-white/10 bg-[rgba(7,10,18,0.62)]">
          <div class="flex gap-2 border-b border-white/10 p-3">
            <span class="h-2.5 w-2.5 rounded-full bg-white/20" />
            <span class="h-2.5 w-2.5 rounded-full bg-white/20" />
            <span class="h-2.5 w-2.5 rounded-full bg-white/20" />
          </div>
          <div class="p-5">
            <p class="text-sm text-[#94a3b8]">Tu coach IA pregunta:</p>
            <h3 class="mt-2 text-xl font-semibold">
              “Tell me about a bug you fixed recently.”
            </h3>
            <div class="mt-4 rounded-[18px] border border-white/10 bg-white/[0.06] p-4">
              <p class="text-sm text-[#94a3b8]">Versión más natural</p>
              <p class="mt-1">
                “I worked on a database-related bug and fixed it yesterday.”
              </p>
            </div>
            <p class="mt-4 text-sm text-[#94a3b8]">Claridad</p>
            <div class="meter mt-2"><span style="width: 78%" /></div>
          </div>
        </div>
      </div>
    </div>

    <section id="features" class="mx-auto grid max-w-[1180px] gap-4 px-4 py-8 md:grid-cols-3">
      <article class="rounded-[22px] border border-white/10 bg-white/[0.06] p-5">
        <span class="tag">Entrevista</span>
        <h3 class="mt-2 text-lg font-semibold">Recursos Humanos</h3>
        <p class="text-[#94a3b8]">Respondé preguntas de entrevista y recibí feedback lingüístico.</p>
      </article>
      <article class="rounded-[22px] border border-white/10 bg-white/[0.06] p-5">
        <span class="tag">Reunión</span>
        <h3 class="mt-2 text-lg font-semibold">Daily meeting</h3>
        <p class="text-[#94a3b8]">Explicá avances, próximas tareas y bloqueos en inglés.</p>
      </article>
      <article class="rounded-[22px] border border-white/10 bg-white/[0.06] p-5">
        <span class="tag">Quiz</span>
        <h3 class="mt-2 text-lg font-semibold">Situaciones IT</h3>
        <p class="text-[#94a3b8]">Elegí la mejor frase para bugs, deploys, APIs y dailies.</p>
      </article>
    </section>

    <LandingDemo />
  </div>
</template>
