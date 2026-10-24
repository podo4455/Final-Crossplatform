<script setup lang="ts">
import { AnalyzePhoto, AnalyzePhotos, GetPhotoDataURL, GetPhotoTags, GetPhotos } from '../../wailsjs/go/main/App'
import type { database } from '../../wailsjs/go/models'

const photos = ref<database.Photo[]>([])
const loading = ref(true)
const error = ref('')
const analyzing = ref<Record<number, boolean>>({})
const analyzingAll = ref(false)
const photoTags = ref<Record<number, database.Tag[]>>({})
const photoThumbs = ref<Record<number, string>>({})
const loadingThumbs = ref<Record<number, boolean>>({})

async function loadTags(photoId: number) {
  try {
    photoTags.value[photoId] = await GetPhotoTags(photoId)
  } catch (cause) {
    console.error('โหลด Tag ไม่สำเร็จ:', cause)
    photoTags.value[photoId] = []
  }
}

async function loadThumb(photoId: number) {
  if (photoThumbs.value[photoId] || loadingThumbs.value[photoId]) return
  loadingThumbs.value[photoId] = true
  try {
    const dataUrl = await GetPhotoDataURL(photoId)
    photoThumbs.value[photoId] = dataUrl
  } catch (cause) {
    console.error('โหลดรูปไม่สำเร็จ:', cause)
  } finally {
    loadingThumbs.value[photoId] = false
  }
}

async function loadPhotos() {
  loading.value = true
  error.value = ''
  try {
    photos.value = await GetPhotos()
    await Promise.all([
      ...photos.value.map((photo) => loadTags(photo.id)),
      ...photos.value.map((photo) => loadThumb(photo.id)),
    ])
  } catch (cause) {
    error.value = `อ่านคลังรูปไม่สำเร็จ: ${String(cause)}`
  } finally {
    loading.value = false
  }
}

async function analyzePhoto(photo: database.Photo) {
  if (analyzing.value[photo.id] || analyzingAll.value) return
  analyzing.value[photo.id] = true
  try {
    const tags = await AnalyzePhoto(photo.id)
    photoTags.value[photo.id] = tags
  } catch (cause) {
    alert(`วิเคราะห์รูปไม่สำเร็จ: ${String(cause)}`)
  } finally {
    analyzing.value[photo.id] = false
  }
}

async function analyzeAll() {
  if (analyzingAll.value || photos.value.length === 0) return
  analyzingAll.value = true
  try {
    const ids = photos.value.map((photo) => photo.id)
    const result = await AnalyzePhotos(ids)
    for (const photo of photos.value) {
      const tags = result[photo.id]
      if (tags) {
        photoTags.value[photo.id] = tags
      } else {
        await loadTags(photo.id)
      }
    }
  } catch (cause) {
    alert(`วิเคราะห์รูปทั้งหมดไม่สำเร็จ: ${String(cause)}`)
  } finally {
    analyzingAll.value = false
  }
}

function formatDate(value: unknown) {
  const date = new Date(String(value))
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleString('th-TH')
}

function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

onMounted(loadPhotos)
</script>

<template>
  <main class="shell">
    <p class="eyebrow">PHOTO LIBRARY</p>
    <h1>คลังรูปภาพ</h1>
    <p class="subtitle">รายการรูปที่ระบบบันทึกไว้หลังสำรองสำเร็จ</p>

    <div class="section-heading">
      <h2>{{ photos.length }} รูปในคลัง</h2>
      <div class="header-actions">
        <button class="primary-button" :disabled="analyzingAll || loading || photos.length === 0" @click="analyzeAll">
          {{ analyzingAll ? 'กำลังวิเคราะห์ทั้งหมด…' : 'วิเคราะห์รูปทั้งหมด' }}
        </button>
        <button class="secondary-button" :disabled="loading || analyzingAll" @click="loadPhotos">
          {{ loading ? 'กำลังโหลด…' : 'รีเฟรชรายการ' }}
        </button>
      </div>
    </div>

    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <p v-else-if="loading" class="subtitle" role="status">กำลังอ่านข้อมูลจากฐานข้อมูล…</p>

    <section v-else-if="photos.length" class="photo-list">
      <article v-for="photo in photos" :key="photo.id" class="photo-row">
        <div class="photo-thumb">
          <img v-if="photoThumbs[photo.id]" :src="photoThumbs[photo.id]" :alt="photo.fileName" />
          <div v-else class="thumb-placeholder"><AppIcon name="gallery" /></div>
        </div>
        <div class="photo-info">
          <h2>{{ photo.fileName }}</h2>
          <p>{{ photo.destinationPath }}</p>
          <small>{{ formatSize(photo.fileSize) }} · สำรองเมื่อ {{ formatDate(photo.backedUpAt) }}</small>
          <div class="tag-section">
            <p class="tag-label">Tag</p>
            <div v-if="photoTags[photo.id]?.length" class="tag-list">
              <span v-for="tag in photoTags[photo.id]" :key="tag.id" class="tag-chip">{{ tag.name }}</span>
            </div>
            <p v-else class="tag-empty">ยังไม่มี Tag</p>
          </div>
        </div>
        <div class="photo-actions">
          <button class="analyze-button" :disabled="analyzing[photo.id] || analyzingAll" @click="analyzePhoto(photo)">
            {{ analyzing[photo.id] ? 'กำลังวิเคราะห์…' : 'วิเคราะห์รูป' }}
          </button>
          <span class="photo-status">{{ photo.backupStatus }}</span>
        </div>
      </article>
    </section>

    <section v-else class="connection">
      <h2>ยังไม่มีรูปในคลัง</h2>
      <p class="subtitle">ไปที่หน้าสำรองรูปภาพ เลือกรูป และคัดลอกให้สำเร็จ แล้วกลับมาหน้านี้</p>
      <NuxtLink class="primary-button" to="/backup">ไปหน้าสำรองรูปภาพ <AppIcon name="arrow" /></NuxtLink>
    </section>
  </main>
</template>

<style scoped>
.photo-list { display: grid; gap: 12px; }
.photo-row { display: flex; align-items: flex-start; gap: 18px; padding: 18px; background: #fff; border: 1px solid #e3e9e5; border-radius: 14px; }
.photo-thumb { width: 120px; height: 90px; border-radius: 12px; overflow: hidden; background: #f0f7f3; flex-shrink: 0; display: flex; align-items: center; justify-content: center; }
.photo-thumb img { width: 100%; height: 100%; object-fit: cover; }
.thumb-placeholder { color: #7b8880; }
.photo-info { min-width: 0; flex: 1; }
.photo-info h2 { font-size: 15px; margin-bottom: 4px; }
.photo-info p { overflow-wrap: anywhere; color: #7b8880; font-size: 12px; margin: 0 0 6px; }
.photo-info small { color: #89968f; font-size: 11px; }
.tag-section { margin-top: 12px; display: flex; flex-direction: column; gap: 8px; }
.tag-label { margin: 0; font-size: 12px; font-weight: 600; color: #4f5d57; }
.tag-list { display: flex; flex-wrap: wrap; gap: 8px; }
.tag-chip { display: inline-flex; align-items: center; padding: 4px 10px; border-radius: 999px; background: #f0f7f3; color: #367959; font-size: 11px; font-weight: 600; text-transform: capitalize; }
.tag-empty { margin: 0; font-size: 11px; color: #89968f; }
.photo-actions { display: flex; flex-direction: column; align-items: flex-end; gap: 10px; }
.header-actions { display: flex; flex-wrap: wrap; gap: 10px; }
.analyze-button { padding: 8px 14px; border-radius: 999px; border: 1px solid #d7e3dc; background: linear-gradient(180deg, #f7fbf9 0%, #eef6f1 100%); color: #2f6b4a; font-size: 12px; font-weight: 600; cursor: pointer; }
.analyze-button:disabled { opacity: 0.6; cursor: not-allowed; }
.photo-status { padding: 6px 10px; border-radius: 20px; background: #eaf3ee; color: #367959; font-size: 11px; font-weight: 600; }
@media (max-width: 700px) {
  .photo-row { flex-direction: column; align-items: flex-start; }
  .photo-actions { width: 100%; flex-direction: row; justify-content: space-between; align-items: center; }
  .header-actions { width: 100%; justify-content: flex-start; }
  .photo-thumb { width: 100%; height: 180px; }
}
</style>
