<template>
  <section>
    <h1>403</h1>
    <p>你没有权限访问这个页面。</p>
    <RouterLink to="/chat">返回应用</RouterLink>
  </section>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>
