<script setup lang="ts">
import { SelectDestination, SelectPhotos, BackupSelectedPhotos } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import type { service } from '../../wailsjs/go/models'
const selectedPhotos = ref<string[]>([])
const hasSource = computed(() => selectedPhotos.value.length > 0)
const destinationPath = ref('')
const selecting = ref(false)
const error = ref('')
const copying = ref(false)
const result = ref<service.Result | null>(null)
const progress = ref({ total: 0, completed: 0, current: '' })
let unsubscribe: (() => void) | undefined
onMounted(() => {
  try { unsubscribe = EventsOn('backup:progress', (value) => { progress.value = value }) }
  catch { /* Normal browser preview has no Wails runtime. */ }
})
onBeforeUnmount(() => unsubscribe?.())
onBeforeRouteLeave(() => !copying.value)
async function backup() {
  if (copying.value || selecting.value || !hasSource.value || !destinationPath.value) return
  copying.value = true
  error.value = ''
  result.value = null
  progress.value = { total: 0, completed: 0, current: '' }
  try { result.value = await BackupSelectedPhotos(selectedPhotos.value, destinationPath.value) }
  catch (cause) { error.value = `สำรองไม่สำเร็จ: ${String(cause)}` }
  finally { copying.value = false }
}
async function selectDestination() {
  if (selecting.value || copying.value) return
  selecting.value = true
  error.value = ''
  try {
    const path = await SelectDestination()
    if (!path) return
    result.value = null
    destinationPath.value = path
  } catch { error.value = 'เลือกโฟลเดอร์ไม่สำเร็จ กรุณาเปิดผ่าน wails dev หรือตัวแอปที่ build แล้ว' }
  finally { selecting.value = false }
}
async function selectPhotos() {
  if (selecting.value || copying.value) return
  selecting.value = true
  error.value = ''
  try {
    const paths = await SelectPhotos()
    if (!paths?.length) return
    selectedPhotos.value = [...new Set(paths)]
    result.value = null
  } catch (cause) { error.value = `เลือกรูปไม่สำเร็จ: ${String(cause)}` }
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
        <div class="photo-picker">
          <h2>เลือกรูปที่ต้องการคัดลอก</h2>
          <button class="secondary-button" type="button" :disabled="selecting || copying" @click="selectPhotos">{{ hasSource ? 'เปลี่ยนรูปที่เลือก' : 'เลือกรูปภาพ (เลือกได้หลายรูป)' }}</button>
          <p class="subtitle">กด Ctrl บน Windows หรือ Command บน macOS ค้างไว้แล้วคลิกรูปที่ต้องการ หรือใช้ Shift เลือกเป็นช่วง</p>
          <p v-if="hasSource" role="status">เลือกแล้ว {{ selectedPhotos.length }} รูป</p>
          <ul v-if="hasSource" class="file-results"><li v-for="path in selectedPhotos" :key="path">{{ path }}</li></ul>
        </div>
        <FolderSelector id="destination-folder" step="02" label="โฟลเดอร์ปลายทาง" hint="เลือกพื้นที่ในคอมพิวเตอร์ที่ต้องการเก็บรูป" :path="destinationPath" :disabled="selecting || copying" @select="selectDestination" />
        <div class="backup-status">
          <p v-if="error" class="error" role="alert">{{ error }}</p>
          <p v-if="copying" class="subtitle" role="status">{{ progress.total ? `คัดลอก ${progress.completed} / ${progress.total} รูป` : 'กำลังค้นหารูปภาพ…' }} {{ progress.current }}</p>
          <p v-else-if="selecting" class="subtitle" role="status">กำลังเลือก…</p>
          <p v-else-if="hasSource && destinationPath" class="success" role="status">เลือกต้นทางและปลายทางครบแล้ว</p>
          <p v-else class="subtitle" role="status">เลือกรูปภาพที่ต้องการ และเลือกโฟลเดอร์ปลายทาง</p>
          <progress v-if="copying && progress.total" :value="progress.completed" :max="progress.total" aria-label="ความคืบหน้าการสำรอง" />
          <button class="primary-button" type="button" :disabled="copying || selecting || !hasSource || !destinationPath" @click="backup">{{ copying ? 'กำลังสำรอง…' : 'เริ่มสำรองรูปภาพ' }}<AppIcon name="arrow" /></button>
          <p class="subtitle">คัดลอกเฉพาะรูปที่เลือก เก็บต้นฉบับไว้ และข้ามไฟล์ที่มีชื่อเดิมในปลายทาง</p>
          <section v-if="result" aria-live="polite" class="backup-results">
            <h2>{{ result.total === 0 ? 'ไม่พบไฟล์รูปภาพที่รองรับ' : 'ดำเนินการเสร็จแล้ว' }}</h2>
            <p>พบ {{ result.total }} รูป · คัดลอก {{ result.copied }} · ข้าม {{ result.skipped }} · ไม่สำเร็จ {{ result.failed }}</p>
            <p v-if="result.skipped" class="subtitle">ไฟล์ที่ข้ามมีชื่ออยู่แล้ว ระบบยังไม่ได้ตรวจว่าเนื้อหาเหมือนกันหรือไม่</p>
            <details v-if="result.files.length"><summary>ผลลัพธ์รายไฟล์</summary>
              <ul class="file-results"><li v-for="file in result.files" :key="file.source" :class="{ error: file.status === 'failed' }">
                {{ file.status === 'copied' ? 'คัดลอกแล้ว' : file.status === 'skipped' ? 'ข้ามชื่อซ้ำ' : 'ไม่สำเร็จ' }} — {{ file.destination }}
                <span v-if="file.error">: {{ file.error }}</span>
              </li></ul>
            </details>
          </section>
        </div>
      </section>
      <aside class="help-card"><AppIcon name="folder" /><h2>จากอุปกรณ์ สู่ที่เก็บของคุณ</h2><p>เลือกรูปที่ต้องการจาก Flash drive ได้หลายรูป แล้วเลือกโฟลเดอร์ปลายทางในเครื่อง ระบบคัดลอกเฉพาะรายการที่เลือก</p><p>รองรับ JPG, JPEG, PNG, GIF, WebP, BMP, TIFF, HEIC, HEIF และ AVIF ตามนามสกุลไฟล์</p><p>โทรศัพท์ต้องแสดงเป็นโฟลเดอร์ที่ระบบอ่านได้ หากเชื่อมต่อแบบ MTP ให้ใช้เครื่องมือนำเข้ารูปของระบบก่อน</p></aside>
    </div>
    <footer>Photo Backup · เลือกโฟลเดอร์ได้เมื่อเปิดในแอป Wails</footer>
  </main>
</template>
<style scoped>
.photo-picker { margin-bottom: 24px; }
progress { display: block; width: 100%; margin: 16px 0; accent-color: #367959; }
.backup-results { margin-top: 24px; }
.file-results { max-height: 260px; overflow: auto; padding-left: 20px; font-size: 12px; overflow-wrap: anywhere; }
.file-results li { margin: 10px 0; }
</style>
