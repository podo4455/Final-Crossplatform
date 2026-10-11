<script setup lang="ts">
import { SelectSource, SelectDestination } from '../../wailsjs/go/main/App'
const sourcePath = ref('')
const destinationPath = ref('')
const selecting = ref(false)
const error = ref('')
async function selectFolder(target: 'source' | 'destination') {
  if (selecting.value) return
  selecting.value = true
  error.value = ''
  try {
    const path = target === 'source' ? await SelectSource() : await SelectDestination()
    if (!path) return
    if (target === 'source') sourcePath.value = path
    else destinationPath.value = path
  } catch { error.value = 'เลือกโฟลเดอร์ไม่สำเร็จ กรุณาเปิดผ่าน wails dev หรือตัวแอปที่ build แล้ว' }
  finally { selecting.value = false }
}
</script>
<template>
  <main class="shell">
    <p class="eyebrow">PHOTO BACKUP</p>
    <h1>สำรองรูปภาพ</h1>
    <p class="subtitle">เริ่มต้นด้วยการเลือกที่อยู่ของรูป และพื้นที่สำหรับจัดเก็บ</p>
    <div class="backup-layout">
      <section class="backup-form" aria-label="เลือกโฟลเดอร์สำหรับสำรองรูป">
        <FolderSelector id="source-folder" step="01" label="โฟลเดอร์ต้นทาง" hint="เลือกโฟลเดอร์รูปภาพใน Flash drive ของคุณ" :path="sourcePath" :disabled="selecting" @select="selectFolder('source')" />
        <FolderSelector id="destination-folder" step="02" label="โฟลเดอร์ปลายทาง" hint="เลือกพื้นที่ในคอมพิวเตอร์ที่ต้องการเก็บรูป" :path="destinationPath" :disabled="selecting" @select="selectFolder('destination')" />
        <div class="backup-status">
          <p v-if="error" class="error" role="alert">{{ error }}</p>
          <p v-if="selecting" class="subtitle" role="status">กำลังเลือกโฟลเดอร์…</p>
          <p v-else-if="sourcePath && destinationPath" class="success" role="status">เลือกต้นทางและปลายทางครบแล้ว</p>
          <p v-else class="subtitle" role="status">เลือกโฟลเดอร์ทั้งสองแห่งเพื่อเตรียมสำรองรูป</p>
          <button class="primary-button" type="button" disabled>เริ่มสำรองรูปภาพ<AppIcon name="arrow" /></button>
          <p class="subtitle">ระบบสำรองไฟล์จะพัฒนาในขั้นถัดไป</p>
        </div>
      </section>
      <aside class="help-card"><AppIcon name="folder" /><h2>จากอุปกรณ์ สู่ที่เก็บของคุณ</h2><p>เชื่อมต่อ Flash drive แล้วเลือกโฟลเดอร์ที่มีรูปภาพ จากนั้นเลือกโฟลเดอร์ปลายทางในเครื่อง</p><p>ต้นแบบนี้เลือกโฟลเดอร์ผ่านแอป Wails ได้แล้ว ส่วนการสำรองและบันทึกประวัติเป็นงานที่จะพัฒนาต่อ</p></aside>
    </div>
    <footer>Photo Backup · เลือกโฟลเดอร์ได้เมื่อเปิดในแอป Wails</footer>
  </main>
</template>
