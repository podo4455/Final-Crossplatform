<script setup lang="ts">
defineProps<{ id: string; label: string; hint: string; step: string; path: string; disabled: boolean }>()
const emit = defineEmits<{ select: [] }>()
</script>
<template>
  <div class="folder-selector">
    <div class="folder-title"><span class="folder-number">{{ step }}</span><label :for="id">{{ label }}</label></div>
    <p class="folder-hint">{{ hint }}</p>
    <div class="input-row"><input :id="id" :value="path" :aria-label="label" placeholder="ยังไม่ได้เลือกโฟลเดอร์" readonly><button class="secondary-button" type="button" :disabled="disabled" @click="emit('select')">เลือกโฟลเดอร์</button></div>
  </div>
</template>
