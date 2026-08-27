<template>
  <div class="stat-card">
    <p class="stat-card__label">{{ label }}</p>
    <div class="stat-card__value-row">
      <strong class="stat-card__value">{{ value }}</strong>
      <span v-if="hint" class="stat-card__hint">{{ hint }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  label: string;
  value: string | number;
  hint?: string;
}>();
</script>
