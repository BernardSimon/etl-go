<template>
  <div class="surface-card" :class="{ 'surface-card--padded': padded }">
    <slot />
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    padded?: boolean;
  }>(),
  {
    padded: true,
  }
);
</script>
