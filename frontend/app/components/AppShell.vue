<script setup lang="ts">
const route = useRoute();

const pageTitle = computed(() => {
  if (route.path === "/backup") return "สำรองรูปภาพ";
  if (route.path === "/gallery") return "คลังรูปภาพ";
  if (route.path === "/tags") return "ค้นหาด้วย Tag";
  return "ภาพรวม";
});
</script>
<template>
  <div class="app-frame">
    <aside class="sidebar">
      <NuxtLink class="brand" to="/"
        ><span class="brand-mark"><AppIcon name="gallery" /></span>Photo
        Backup</NuxtLink
      >
      <p class="nav-caption">WORKSPACE</p>
      <nav class="nav-items" aria-label="เมนูหลัก">
        <NuxtLink
          class="nav-item"
          :class="{ active: route.path === '/' }"
          to="/"
          ><AppIcon name="home" />ภาพรวม</NuxtLink
        >
        <NuxtLink
          class="nav-item"
          :class="{ active: route.path === '/backup' }"
          to="/backup"
          ><AppIcon name="folder" />สำรองรูปภาพ</NuxtLink
        >
        <NuxtLink
          class="nav-item"
          :class="{ active: route.path === '/gallery' }"
          to="/gallery"
        >
          <AppIcon name="gallery" />คลังรูปภาพ
        </NuxtLink>
        <NuxtLink class="nav-item" :class="{ active: route.path === '/tags' }" to="/tags"><AppIcon name="sparkles" />ค้นหาด้วย Tag</NuxtLink>
      </nav>
      <div class="sidebar-note">
        พื้นที่เล็ก ๆ สำหรับ<br />ความทรงจำที่สำคัญของคุณ
      </div>
    </aside>
    <div class="workspace">
      <header class="topbar">
        <span>พื้นที่ทำงาน / {{ pageTitle }}</span>
        <span class="preview-badge">ต้นแบบสำหรับทีม</span>
      </header>
      <slot />
    </div>
  </div>
</template>
