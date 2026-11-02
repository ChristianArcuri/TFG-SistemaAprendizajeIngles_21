<script setup>
defineProps({
  streakDays: { type: Number, default: 0 },
  comprehensionPct: { type: Number, default: 0 },
  vocabularyPct: { type: Number, default: 0 },
  communicationPct: { type: Number, default: 0 },
})
</script>

<template>
  <section class="mb-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
    <article class="panel p-5">
      <span class="text-sm text-[#94a3b8]">Días de práctica</span>
      <p class="mt-2 text-3xl font-bold">{{ streakDays }} {{ streakDays === 1 ? 'día' : 'días' }}</p>
    </article>
    <article class="panel p-5">
      <span class="text-sm text-[#94a3b8]">Comprensión</span>
      <p class="mt-2 text-3xl font-bold">{{ comprehensionPct }}%</p>
    </article>
    <article class="panel p-5">
      <span class="text-sm text-[#94a3b8]">Vocabulario profesional</span>
      <p class="mt-2 text-3xl font-bold">{{ vocabularyPct }}%</p>
    </article>
    <article class="panel p-5">
      <span class="text-sm text-[#94a3b8]">Comunicación profesional</span>
      <p class="mt-2 text-3xl font-bold">{{ communicationPct }}%</p>
    </article>
  </section>
</template>
