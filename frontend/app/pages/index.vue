<script setup lang="ts">
import { Greet } from '../../wailsjs/go/main/App'
const name = ref('ทีม Photo Backup')
const message = ref('')
const error = ref('')
const busy = ref(false)
async function checkConnection() {
  busy.value = true
  message.value = ''; error.value = ''
  try { message.value = await Greet(name.value.trim() || 'ทีม Photo Backup') }
  catch { error.value = 'เรียก Go ไม่สำเร็จ กรุณาเปิดผ่าน wails dev หรือตัวแอปที่ build แล้ว' }
  finally { busy.value = false }
}
</script>
<template>
  <main class="shell">
    <section class="hero" aria-labelledby="welcome-title">
      <div>
        <p class="eyebrow">YOUR MEMORIES, IN ONE PLACE</p>
        <h1 id="welcome-title">เก็บรูปที่สำคัญ<br>ไว้ในที่เดียวกัน</h1>
        <p class="subtitle">สำรองรูปจาก Flash drive สู่คอมพิวเตอร์ของคุณ<br>พร้อมต่อยอดเป็นคลังรูปที่ค้นหาได้ง่าย</p>
        <NuxtLink class="primary-button" to="/backup">เริ่มสำรองรูปภาพ<AppIcon name="arrow" /></NuxtLink>
      </div>
      <PhotoIllustration />
    </section>
    <div class="section-heading"><h2>พื้นที่สำหรับรูปภาพของคุณ</h2><span>เลือกฟังก์ชันที่ต้องการใช้งาน</span></div>
    <section class="roadmap" aria-label="ฟีเจอร์ของโปรเจกต์">
      <NuxtLink class="feature-card" to="/backup">
        <div class="feature-icon"><AppIcon name="folder" /></div>
        <h2>สำรองรูปภาพ</h2>
        <p>เลือกต้นทางและปลายทาง<br>เพื่อจัดเก็บรูปในคอมพิวเตอร์</p>
        <span class="feature-link">ไปหน้าสำรองรูปภาพ <AppIcon name="arrow" /></span>
      </NuxtLink>
      <NuxtLink class="feature-card" to="/gallery">
        <div class="feature-icon"><AppIcon name="gallery" /></div>
        <h2>คลังรูปภาพ</h2>
        <p>ดูรูปที่สำรองไว้ในที่เดียว<br>พร้อมข้อมูลและประวัติการจัดเก็บ</p>
        <span class="feature-link">เปิดคลังรูปภาพ <AppIcon name="arrow" /></span>
      </NuxtLink>
      <NuxtLink class="feature-card" to="/tags">
        <div class="feature-icon"><AppIcon name="sparkles" /></div>
        <h2>ค้นหาด้วย Tag</h2>
        <p>ค้นหาความทรงจำจากคำที่เกี่ยวข้อง<br>ด้วย Tag ที่ได้จากการวิเคราะห์ภาพ</p>
        <span class="feature-link">ค้นหารูปภาพ <AppIcon name="arrow" /></span>
      </NuxtLink>
    </section>
    <details class="connection">
      <summary>เครื่องมือสำหรับทีม · ทดสอบการเชื่อมต่อ Nuxt กับ Go</summary>
      <form @submit.prevent="checkConnection">
        <label for="name">ชื่อสำหรับทดสอบ</label>
        <div class="input-row"><input id="name" v-model="name" maxlength="100" autocomplete="off"><button class="secondary-button" :disabled="busy" type="submit">{{ busy ? 'กำลังทดสอบ…' : 'ทดสอบการเชื่อมต่อ' }}</button></div>
      </form>
      <p v-if="message" class="success" role="status">{{ message }}</p>
      <p v-if="error" class="error" role="alert">{{ error }}</p>
    </details>
    <footer>Photo Backup · Final Project · ต้นแบบหน้าตาและโครงสำหรับพัฒนาต่อ</footer>
  </main>
</template>
