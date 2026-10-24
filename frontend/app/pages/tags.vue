<script setup lang="ts">
import { GetPhotoDataURL, GetPhotos, GetTagCloud, SearchPhotosByTag } from '../../wailsjs/go/main/App'
import type { database } from '../../wailsjs/go/models'

const photos = ref<database.Photo[]>([])
const tagCloud = ref<database.TagCloudItem[]>([])
const activeTag = ref('')
const loading = ref(false)
const loadingTagCloud = ref(true)
const error = ref('')
const photoThumbs = ref<Record<number, string>>({})
const loadingThumbs = ref<Record<number, boolean>>({})

async function loadTagCloud() {
  loadingTagCloud.value = true
  try {
    tagCloud.value = await GetTagCloud()
  } catch (cause) {
    console.error('โหลด Tag Cloud ไม่สำเร็จ:', cause)
    tagCloud.value = []
  } finally {
    loadingTagCloud.value = false
  }
}

async function loadThumb(photoId: number) {
  if (photoThumbs.value[photoId] || loadingThumbs.value[photoId]) {
    return
  }

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

async function searchByTag(tagName: string) {
  if (loading.value) return

  loading.value = true
  error.value = ''
  activeTag.value = tagName
  photos.value = []

  try {
    photos.value = await SearchPhotosByTag(tagName)
    await Promise.all(photos.value.map((photo) => loadThumb(photo.id)))
  } catch (cause) {
    error.value = `ค้นหาด้วย Tag ไม่สำเร็จ: ${String(cause)}`
  } finally {
    loading.value = false
  }
}

async function showAll() {
  loading.value = true
  error.value = ''
  activeTag.value = ''
  photos.value = []

  try {
    photos.value = await GetPhotos()
    await Promise.all(photos.value.map((photo) => loadThumb(photo.id)))
  } catch (cause) {
    error.value = `โหลดรูปทั้งหมดไม่สำเร็จ: ${String(cause)}`
  } finally {
    loading.value = false
  }
}

function getTagSize(count: number) {
  if (count >= 10) return 'tag-large'
  if (count >= 5) return 'tag-medium'
  return 'tag-small'
}

function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function formatDate(value: unknown) {
  const date = new Date(String(value))
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleString('th-TH')
}

onMounted(async () => {
  await loadTagCloud()
  await showAll()
})
</script>

<template>
  <main class="shell">
    <p class="eyebrow">TAG SEARCH</p>
    <h1>ค้นหาด้วย Tag</h1>
    <p class="subtitle">ค้นหารูปภาพจาก Tag ที่ได้จากการวิเคราะห์ภาพ</p>

    <section class="tag-cloud-section">
      <div class="section-heading">
        <h2>Tag Cloud</h2>
        <button v-if="activeTag || photos.length > 0" class="secondary-button" @click="showAll">
          ดูรูปทั้งหมด
        </button>
      </div>

      <p v-if="loadingTagCloud" class="subtitle">กำลังโหลด Tag Cloud...</p>

      <div v-else-if="tagCloud.length > 0" class="tag-cloud">
        <button
          v-for="tag in tagCloud"
          :key="tag.id"
          class="tag-cloud-item"
          :class="[getTagSize(tag.photoCount), { active: activeTag === tag.name }]"
          @click="searchByTag(tag.name)"
        >
          {{ tag.name }}
          <span class="tag-count">{{ tag.photoCount }}</span>
        </button>
      </div>

      <p v-else class="tag-empty">ยังไม่มี Tag ในระบบ</p>
      <p class="hint">กด Tag ใดก็ได้เพื่อค้นหารูปที่เกี่ยวข้อง</p>
    </section>

    <div class="section-heading">
      <h2>
        {{ activeTag ? `ผลการค้นหา Tag: "${activeTag}"` : 'รูปทั้งหมดในคลัง' }}
        <span class="count">({{ photos.length }} รูป)</span>
      </h2>
    </div>

    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <p v-else-if="loading" class="subtitle" role="status">กำลังค้นหา...</p>

    <section v-else-if="photos.length" class="photo-grid">
      <article v-for="photo in photos" :key="photo.id" class="photo-card">
        <div class="photo-thumb">
          <img v-if="photoThumbs[photo.id]" :src="photoThumbs[photo.id]" :alt="photo.fileName" />
          <div v-else class="thumb-placeholder">
            <AppIcon name="gallery" />
          </div>
        </div>
        <div class="photo-card-info">
          <h3>{{ photo.fileName }}</h3>
          <p>{{ formatSize(photo.fileSize) }} · {{ formatDate(photo.backedUpAt) }}</p>
        </div>
      </article>
    </section>

    <section v-else class="empty-state">
      <h2>ไม่พบรูปที่มี Tag นี้</h2>
      <p class="subtitle">ลองเลือก Tag อื่น หรือกด "ดูรูปทั้งหมด"</p>
      <button class="primary-button" @click="showAll">ดูรูปทั้งหมด</button>
    </section>
  </main>
</template>

<style scoped>
.tag-cloud-section {
  margin-bottom: 32px;
  padding: 24px;
  background: #fff;
  border: 1px solid #e3e9e5;
  border-radius: 16px;
}

.tag-cloud {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  margin-top: 18px;
}

.tag-cloud-item {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 999px;
  border: 1px solid #d7e3dc;
  background: linear-gradient(180deg, #f7fbf9 0%, #f0f7f3 100%);
  color: #2f6b4a;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
  text-transform: capitalize;
}

.tag-cloud-item:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 10px rgba(54, 121, 89, 0.08);
}

.tag-cloud-item.active {
  background: #367959;
  color: #fff;
  border-color: #367959;
}

.tag-cloud-item.active .tag-count {
  background: rgba(255, 255, 255, 0.25);
  color: #fff;
}

.tag-small { font-size: 12px; }
.tag-medium { font-size: 14px; }
.tag-large { font-size: 16px; }

.tag-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 24px;
  padding: 2px 8px;
  border-radius: 999px;
  background: #eaf3ee;
  color: #367959;
  font-size: 11px;
}

.hint {
  margin-top: 14px;
  font-size: 12px;
  color: #7b8880;
}

.tag-empty {
  margin-top: 16px;
  font-size: 13px;
  color: #7b8880;
}

.count {
  font-weight: 400;
  color: #7b8880;
}

.photo-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 16px;
}

.photo-card {
  background: #fff;
  border: 1px solid #e3e9e5;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 10px 30px rgba(47, 107, 74, 0.02);
}

.photo-thumb {
  width: 100%;
  height: 200px;
  background: #f0f7f3;
}

.photo-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.thumb-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #7b8880;
}

.photo-card-info {
  padding: 14px 18px;
}

.photo-card-info h3 {
  font-size: 14px;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.photo-card-info p {
  font-size: 12px;
  color: #7b8880;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  background: #fff;
  border: 1px solid #e3e9e5;
  border-radius: 16px;
}

.empty-state h2 {
  margin-bottom: 8px;
}
</style>
