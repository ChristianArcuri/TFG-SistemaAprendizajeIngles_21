<script setup>
import { RouterLink, useRoute } from 'vue-router'
import BrandLogo from './BrandLogo.vue'

const route = useRoute()

const links = [
  { to: '/dashboard', label: 'Inicio' },
  { to: '/history', label: 'Progreso' },
]

function isActive(path) {
  if (path === '/dashboard') {
    return (
      route.path === '/dashboard' ||
      route.path.startsWith('/practice') ||
      route.path === '/quiz'
    )
  }
  return route.path === path
}
</script>

<template>
  <aside class="border-r border-white/10 bg-[rgba(5,8,15,0.68)] p-6">
    <BrandLogo to="/dashboard" />
    <nav class="mt-8 grid gap-2">
      <RouterLink
        v-for="link in links"
        :key="link.to"
        :to="link.to"
        class="rounded-[14px] px-[14px] py-[13px] text-[#94a3b8] transition hover:bg-white/[0.07] hover:text-white"
        :class="{ 'bg-white/[0.07] text-white': isActive(link.to) }"
      >
        {{ link.label }}
      </RouterLink>
    </nav>
  </aside>
</template>
