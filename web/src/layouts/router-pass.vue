<template>
  <!-- 仅透传子路由，避免在 LAYOUT 内再嵌一层完整 Layout -->
  <router-view />
</template>

<script setup lang="ts">
defineOptions({ name: 'RouterPass' });
</script>
