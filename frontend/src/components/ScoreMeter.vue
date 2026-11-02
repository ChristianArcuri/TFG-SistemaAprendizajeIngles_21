<script setup>
defineProps({
  value: { type: Number, default: 0 },
  label: { type: String, default: '' },
})
</script>

<template>
  <div>
    <div v-if="label" class="mb-2 text-sm text-[#94a3b8]">{{ label }}</div>
    <div class="meter"><span :style="{ width: `${Math.min(100, Math.max(0, value))}%` }" /></div>
  </div>
</template>
