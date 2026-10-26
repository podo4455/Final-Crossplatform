<script setup lang="ts">
import { GetPhotoDataURL, GetPhotos, GetTagCloud, SearchPhotosByTags } from '../../wailsjs/go/main/App'
import type { database } from '../../wailsjs/go/models'

const photos = ref<database.Photo[]>([])
const tagCloud = ref<database.TagCloudItem[]>([])
const selectedTags = ref<string[]>([])
const matchAll = ref(false)
const loading = ref(false)
const loadingTagCloud = ref(true)
const error = ref('')
const photoThumbs = ref<Record<number, string>>({})
const loadingThumbs = ref<Record<number, boolean>>({})

// ใช้กันผลลัพธ์ของคำขอเก่ามาทับคำขอล่าสุด เมื่อผู้ใช้กด Tag เร็ว ๆ
let searchId = 0

const currentPage = ref(1)
const itemsPerPage = 12

const totalPages = computed(() => Math.max(1, Math.ceil(photos.value.length / itemsPerPage)))

const paginatedPhotos = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage
  const end = start + itemsPerPage
  return photos.value.slice(start, end)
})

watch(photos, () => {
  currentPage.value = 1
})

function goToPage(page: number) {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

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

async function loadCurrentPageThumbs() {
  const current = paginatedPhotos.value
  const needLoad = current.filter((p) => !photoThumbs.value[p.id] && !loadingThumbs.value[p.id])
  await Promise.all(needLoad.map((p) => loadThumb(p.id)))
}

watch(paginatedPhotos, () => {
  loadCurrentPageThumbs()
})

async function toggleTag(tagName: string) {
  const index = selectedTags.value.indexOf(tagName)
  if (index >= 0) {
    selectedTags.value.splice(index, 1)
  } else {
    selectedTags.value.push(tagName)
  }

  await searchPhotos()
}

async function searchPhotos() {
  const id = ++searchId
  loading.value = true
  error.value = ''
  photos.value = []

  try {
    const result = selectedTags.value.length > 0
      ? await SearchPhotosByTags([...selectedTags.value], matchAll.value)
      : await GetPhotos()

    if (id !== searchId) return
    photos.value = result
  } catch (cause) {
    if (id !== searchId) return
    error.value = `ค้นหาด้วย Tag ไม่สำเร็จ: ${String(cause)}`
  } finally {
    if (id === searchId) {
      loading.value = false
      loadCurrentPageThumbs()
    }
  }
}

async function setMatchAll(value: boolean) {
  if (matchAll.value === value) return
  matchAll.value = value
  await searchPhotos()
}

async function showAll() {
  selectedTags.value = []
  matchAll.value = false
  await searchPhotos()
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
        <button v-if="selectedTags.length > 0" class="secondary-button" @click="showAll">
          ล้างการเลือก
        </button>
      </div>

      <p v-if="loadingTagCloud" class="subtitle">กำลังโหลด Tag Cloud...</p>

      <div v-else-if="tagCloud.length > 0" class="tag-cloud-wrapper">
        <div class="tag-cloud">
          <button
            v-for="tag in tagCloud"
            :key="tag.id"
            class="tag-cloud-item"
            :class="[getTagSize(tag.photoCount), { active: selectedTags.includes(tag.name) }]"
            @click="toggleTag(tag.name)"
          >
            {{ tag.name }}
            <span class="tag-count">{{ tag.photoCount }}</span>
          </button>
        </div>
      </div>

      <p v-else class="tag-empty">ยังไม่มี Tag ในระบบ</p>
      <p class="hint">กด Tag เพื่อเลือกได้หลายคำ — เลือกแล้วระบบจะกรองรูปให้อัตโนมัติ</p>

      <div v-if="selectedTags.length > 0" class="selected-bar">
        <span class="selected-label">Tag ที่เลือก:</span>
        <button
          v-for="tag in selectedTags"
          :key="tag"
          class="selected-chip"
          :title="`เอา Tag ${tag} ออก`"
          @click="toggleTag(tag)"
        >
          {{ tag }} <span aria-hidden="true">×</span>
        </button>

        <div v-if="selectedTags.length > 1" class="match-mode" role="group" aria-label="เงื่อนไขการค้นหา">
          <button
            class="match-mode-button"
            :class="{ active: !matchAll }"
            @click="setMatchAll(false)"
          >
            มี Tag ใดก็ได้
          </button>
          <button
            class="match-mode-button"
            :class="{ active: matchAll }"
            @click="setMatchAll(true)"
          >
            ต้องมีทุก Tag
          </button>
        </div>
      </div>
    </section>

    <div class="section-heading">
      <h2>
        {{ selectedTags.length > 0 ? `ผลการค้นหา ${selectedTags.length} Tag` : 'รูปทั้งหมดในคลัง' }}
        <span class="count">({{ photos.length }} รูป)</span>
      </h2>
    </div>

    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <p v-else-if="loading" class="subtitle" role="status">กำลังค้นหา...</p>

    <section v-else-if="photos.length">
      <div class="photo-grid">
        <article v-for="photo in paginatedPhotos" :key="photo.id" class="photo-card">
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
      </div>

      <div class="pagination" v-if="totalPages > 1">
        <button class="pagination-button" :disabled="currentPage === 1" @click="goToPage(currentPage - 1)">ก่อนหน้า</button>
        <div class="pagination-pages">
          <button
            v-for="page in totalPages"
            :key="page"
            class="pagination-page"
            :class="{ active: currentPage === page }"
            @click="goToPage(page)"
          >
            {{ page }}
          </button>
        </div>
        <button class="pagination-button" :disabled="currentPage === totalPages" @click="goToPage(currentPage + 1)">ถัดไป</button>
      </div>
      <p class="pagination-info">แสดง {{ paginatedPhotos.length }} จาก {{ photos.length }} รูป (หน้า {{ currentPage }} / {{ totalPages }})</p>
    </section>

    <section v-else class="empty-state">
      <h2>ไม่พบรูปที่ตรงกับ Tag ที่เลือก</h2>
      <p class="subtitle">
        {{ matchAll && selectedTags.length > 1 ? 'ลองเปลี่ยนเป็น "มี Tag ใดก็ได้" หรือลด Tag ที่เลือก' : 'ลองเลือก Tag อื่น หรือล้างการเลือก' }}
      </p>
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

.tag-cloud-wrapper {
  margin-top: 18px;
  max-height: 220px;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 8px 12px;
  border: 1px solid #e3e9e5;
  border-radius: 12px;
  background: linear-gradient(180deg, #ffffff 0%, #f9fbfa 100%);
}

.tag-cloud {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
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

.selected-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px dashed #dde6e0;
}

.selected-label {
  font-size: 12px;
  font-weight: 600;
  color: #4f5d57;
}

.selected-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  border-radius: 999px;
  border: 1px solid #367959;
  background: #367959;
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  text-transform: capitalize;
}

.selected-chip:hover {
  background: #2f6b4a;
}

.match-mode {
  display: inline-flex;
  gap: 4px;
  margin-left: auto;
  padding: 3px;
  border-radius: 10px;
  background: #eef3f0;
}

.match-mode-button {
  padding: 6px 12px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #5c6c64;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.match-mode-button.active {
  background: #fff;
  color: #2f6b4a;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.05);
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

.tag-cloud-wrapper::-webkit-scrollbar {
  width: 8px;
}

.tag-cloud-wrapper::-webkit-scrollbar-track {
  background: transparent;
}

.tag-cloud-wrapper::-webkit-scrollbar-thumb {
  background: rgba(54, 121, 89, 0.4);
  border-radius: 999px;
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

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-top: 24px;
  flex-wrap: wrap;
}

.pagination-button {
  padding: 8px 14px;
  border-radius: 999px;
  border: 1px solid #d7e3dc;
  background: #fff;
  color: #2f6b4a;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}

.pagination-button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.pagination-pages {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.pagination-page {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 1px solid transparent;
  background: transparent;
  color: #4f5d57;
  font-size: 12px;
  cursor: pointer;
}

.pagination-page.active {
  background: #367959;
  color: #fff;
}

.pagination-info {
  text-align: center;
  margin-top: 10px;
  font-size: 12px;
  color: #7b8880;
}
</style>
