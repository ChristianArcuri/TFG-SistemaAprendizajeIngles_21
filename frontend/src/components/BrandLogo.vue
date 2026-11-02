<script setup>
import { RouterLink } from 'vue-router'

defineProps({
  to: { type: String, default: '/' },
})
</script>

<template>
  <RouterLink :to="to" class="flex items-center gap-3 font-extrabold tracking-tight text-white">
    <span
      class="grid h-[38px] w-[38px] place-items-center rounded-[14px] bg-linear-to-br from-[#7c3aed] to-[#22d3ee] shadow-[0_15px_45px_rgba(124,58,237,0.35)] text-sm"
    >IT</span>
    <span>Inglés Profesional IT</span>
  </RouterLink>
</template>
