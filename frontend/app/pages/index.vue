<script setup lang="ts">
import { Greet } from '../../wailsjs/go/main/App'

const name = ref('ทีม Photo Backup')
const message = ref('')
const error = ref('')
const busy = ref(false)

async function checkConnection() {
  busy.value = true
  message.value = ''
  error.value = ''
  try {
    message.value = await Greet(name.value.trim() || 'ทีม Photo Backup')
  } catch {
    error.value = 'เรียก Go ไม่สำเร็จ กรุณาเปิดผ่าน wails dev หรือตัวแอปที่ build แล้ว'
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <main class="shell">
    <header>
      <p class="eyebrow">FINAL PROJECT · GO / WAILS / NUXT</p>
      <h1>Photo Backup</h1>
      <p class="subtitle">เริ่มสร้างพื้นที่เก็บความทรงจำของคุณ</p>
    </header>

    <section class="connection" aria-labelledby="connection-title">
      <span class="step">ขั้นที่ 1 · ตั้งโครงโปรเจกต์</span>
      <h2 id="connection-title">ทดสอบ Nuxt ↔ Go</h2>
      <p>ปุ่มนี้เรียกฟังก์ชัน Greet ใน app.go ผ่าน Wails binding จริง</p>
      <form @submit.prevent="checkConnection">
        <label for="name">ชื่อสำหรับทดสอบ</label>
        <div class="input-row">
          <input id="name" v-model="name" maxlength="100" autocomplete="off">
          <button :disabled="busy" type="submit">{{ busy ? 'กำลังทดสอบ…' : 'ทดสอบการเชื่อมต่อ' }}</button>
        </div>
      </form>
      <p v-if="message" class="success" role="status">{{ message }}</p>
      <p v-if="error" class="error" role="alert">{{ error }}</p>
    </section>

    <section class="roadmap" aria-label="แผนพัฒนาของทีม">
      <article><span>01 · สมาชิก A</span><h2>Backup</h2><p>เลือกโฟลเดอร์ สำรองรูป และแสดงความคืบหน้า</p><small>ยังไม่ได้พัฒนา</small></article>
      <article><span>02 · สมาชิก B</span><h2>Library</h2><p>ฐานข้อมูล Gallery ประวัติ และค้นรูป</p><small>ยังไม่ได้พัฒนา</small></article>
      <article><span>03 · สมาชิก C</span><h2>Image Analysis</h2><p>วิเคราะห์ภาพเป็น Tag และแสดง Tag Cloud</p><small>ยังไม่ได้พัฒนา</small></article>
    </section>
    <footer>โครงโปรเจกต์สำหรับทำงานร่วมกัน 3 คน · SQLite/GORM เป็นขั้นถัดไป</footer>
  </main>
</template>
