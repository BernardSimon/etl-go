<template>
  <div class="flex items-center justify-center w-full h-screen bg-gray-100">
    <a-result status="500" title="404" :sub-title="$t('404.message')">
      <template #extra>
        <a-button type="primary" @click="goHome">{{$t('404.back')}}</a-button>
      </template>
    </a-result>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from "vue-router";

const router = useRouter();

const goHome = () => {
  router.push("/");
};
</script>
