<script setup>
import AppSidebar from '../components/AppSidebar.vue'
</script>

<template>
  <div class="grid min-h-screen grid-cols-1 md:grid-cols-[260px_1fr]">
    <AppSidebar class="hidden md:block" />
    <main class="p-6 md:p-8">
      <slot />
    </main>
  </div>
</template>
