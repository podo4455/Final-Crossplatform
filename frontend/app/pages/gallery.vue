<script setup lang="ts">
import { AddPhotoTag, AnalyzePhoto, AnalyzePhotos, AnalyzeUnanalyzedPhotos, CleanupMissingPhotos, DeletePhoto, GetAllPhotoTags, GetPhotoDataURL, GetPhotoTags, GetPhotos, RemovePhotoTag, SearchPhotosByName } from '../../wailsjs/go/main/App'
import type { database } from '../../wailsjs/go/models'

const photos = ref<database.Photo[]>([])
const loading = ref(true)
const error = ref('')
const searching = ref(false)
const analyzing = ref<Record<number, boolean>>({})
const analyzingAll = ref(false)
const analyzingNew = ref(false)
const cleaningUp = ref(false)
const photoTags = ref<Record<number, database.Tag[]>>({})
const photoThumbs = ref<Record<number, string>>({})
const loadingThumbs = ref<Record<number, boolean>>({})

const addingTag = ref<Record<number, boolean>>({})
const newTagName = ref<Record<number, string>>({})

const searchKeyword = ref('')

const activeTab = ref<'all' | 'pending' | 'done'>('all')

const currentPage = ref(1)
const itemsPerPage = 12

const filteredPhotos = computed(() => {
  if (activeTab.value === 'pending') {
    return photos.value.filter((photo) => !(photoTags.value[photo.id]?.length))
  }
  if (activeTab.value === 'done') {
    return photos.value.filter((photo) => (photoTags.value[photo.id]?.length))
  }
  return photos.value
})

const totalPages = computed(() => Math.max(1, Math.ceil(filteredPhotos.value.length / itemsPerPage)))

const paginatedPhotos = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage
  const end = start + itemsPerPage
  return filteredPhotos.value.slice(start, end)
})

const pendingCount = computed(() => photos.value.filter((photo) => !(photoTags.value[photo.id]?.length)).length)
const doneCount = computed(() => photos.value.filter((photo) => (photoTags.value[photo.id]?.length)).length)

watch([photos, photoTags, activeTab], () => {
  currentPage.value = 1
})

function goToPage(page: number) {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function isAnalyzed(photoId: number) {
  return !!photoTags.value[photoId]?.length
}

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

async function loadCurrentPageThumbs() {
  const current = paginatedPhotos.value
  const needLoad = current.filter((p) => !photoThumbs.value[p.id] && !loadingThumbs.value[p.id])
  await Promise.all(needLoad.map((p) => loadThumb(p.id)))
}

watch(paginatedPhotos, () => {
  loadCurrentPageThumbs()
})

async function loadPhotos() {
  loading.value = true
  searching.value = false
  error.value = ''
  searchKeyword.value = ''
  try {
    const [fetchedPhotos, fetchedTags] = await Promise.all([
      GetPhotos(),
      GetAllPhotoTags().catch(() => ({})),
    ])
    photos.value = fetchedPhotos || []
    photoTags.value = fetchedTags || {}
  } catch (cause) {
    error.value = `อ่านคลังรูปไม่สำเร็จ: ${String(cause)}`
  } finally {
    loading.value = false
    loadCurrentPageThumbs()
  }
}

async function searchByName() {
  const keyword = searchKeyword.value.trim()
  loading.value = true
  error.value = ''
  searching.value = keyword !== ''
  try {
    const [fetchedPhotos, fetchedTags] = await Promise.all([
      SearchPhotosByName(keyword),
      GetAllPhotoTags().catch(() => ({})),
    ])
    photos.value = fetchedPhotos || []
    photoTags.value = fetchedTags || {}
  } catch (cause) {
    error.value = `ค้นหาชื่อไฟล์ไม่สำเร็จ: ${String(cause)}`
  } finally {
    loading.value = false
    loadCurrentPageThumbs()
  }
}

function clearSearch() {
  searchKeyword.value = ''
  loadPhotos()
}

async function analyzePhoto(photo: database.Photo) {
  if (analyzing.value[photo.id] || analyzingAll.value || analyzingNew.value) return
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
  if (analyzingAll.value || photos.value.length === 0 || analyzingNew.value) return
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

async function analyzeNewOnly() {
  if (analyzingNew.value || pendingCount === 0 || analyzingAll.value) return
  analyzingNew.value = true
  try {
    const result = await AnalyzeUnanalyzedPhotos()
    const updated = Object.keys(result).length
    for (const photo of photos.value) {
      const tags = result[photo.id]
      if (tags) {
        photoTags.value[photo.id] = tags
      }
    }
    if (updated === 0) {
      alert('ไม่มีรูปที่ยังไม่ได้วิเคราะห์')
    } else {
      alert(`วิเคราะห์เสร็จ ${updated} รูป`)
    }
  } catch (cause) {
    alert(`วิเคราะห์รูปที่ยังไม่มี Tag ไม่สำเร็จ: ${String(cause)}`)
  } finally {
    analyzingNew.value = false
  }
}

async function cleanupMissing() {
  if (cleaningUp.value) return
  cleaningUp.value = true
  try {
    const removed = await CleanupMissingPhotos()
    if (removed === 0) {
      alert('ไม่พบรูปที่หายไป — ข้อมูลตรงกับไฟล์จริงแล้ว')
    } else {
      alert(`ลบข้อมูลรูปที่ไฟล์ไม่อยู่แล้ว ${removed} รายการ`)
      await loadPhotos()
    }
  } catch (cause) {
    alert(`ทำความสะอาดข้อมูลไม่สำเร็จ: ${String(cause)}`)
  } finally {
    cleaningUp.value = false
  }
}

async function deleteSinglePhoto(photo: database.Photo) {
  if (!confirm(`ต้องการลบข้อมูลรูป "${photo.fileName}" ออกจากระบบหรือไม่?`)) return
  try {
    await DeletePhoto(photo.id)
    photos.value = photos.value.filter((p) => p.id !== photo.id)
    delete photoTags.value[photo.id]
    delete photoThumbs.value[photo.id]
  } catch (cause) {
    alert(`ลบข้อมูลรูปไม่สำเร็จ: ${String(cause)}`)
  }
}

async function addTagToPhoto(photoId: number) {
  const name = (newTagName.value[photoId] || '').trim()
  if (!name) return
  if (addingTag.value[photoId]) return

  addingTag.value[photoId] = true
  try {
    const tags = await AddPhotoTag(photoId, name)
    photoTags.value[photoId] = tags
    newTagName.value[photoId] = ''
  } catch (cause) {
    alert(`เพิ่ม Tag ไม่สำเร็จ: ${String(cause)}`)
  } finally {
    addingTag.value[photoId] = false
  }
}

async function removeTagFromPhoto(photoId: number, tagId: number) {
  try {
    await RemovePhotoTag(photoId, tagId)
    photoTags.value[photoId] = (photoTags.value[photoId] || []).filter((tag) => tag.id !== tagId)
  } catch (cause) {
    alert(`ลบ Tag ไม่สำเร็จ: ${String(cause)}`)
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
      <h2>{{ photos.length }} {{ searching ? 'ผลการค้นหา' : 'รูปในคลัง' }}</h2>
      <div class="header-actions">
        <button class="secondary-button" :disabled="analyzingNew || analyzingAll || loading || photos.length === 0" @click="analyzeNewOnly">
          {{ analyzingNew ? 'กำลังวิเคราะห์รูปใหม่…' : 'วิเคราะห์เฉพาะรูปที่ยังไม่มี Tag' }}
        </button>
        <button class="primary-button" :disabled="analyzingAll || analyzingNew || loading || photos.length === 0" @click="analyzeAll">
          {{ analyzingAll ? 'กำลังวิเคราะห์ทั้งหมด…' : 'วิเคราะห์รูปทั้งหมด' }}
        </button>
        <button class="secondary-button" :disabled="loading || analyzingAll || analyzingNew || cleaningUp" @click="loadPhotos">
          {{ loading ? 'กำลังโหลด…' : 'รีเฟรชรายการ' }}
        </button>
        <button class="secondary-button cleanup-button" :disabled="loading || analyzingAll || analyzingNew || cleaningUp || photos.length === 0" @click="cleanupMissing">
          {{ cleaningUp ? 'กำลังตรวจสอบ…' : 'ลบรูปที่หายไป' }}
        </button>
      </div>
    </div>

    <div class="search-bar">
      <input
        class="search-input"
        v-model="searchKeyword"
        placeholder="ค้นหาจากชื่อไฟล์รูปภาพ..."
        @keyup.enter="searchByName"
      />
      <button class="search-button" :disabled="loading" @click="searchByName">ค้นหา</button>
      <button v-if="searching" class="secondary-button" @click="clearSearch">แสดงทั้งหมด</button>
    </div>

    <div v-if="photos.length" class="tabs-bar">
      <div class="tabs">
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'all' }"
          @click="activeTab = 'all'"
        >
          ทั้งหมด ({{ photos.length }})
        </button>
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'pending' }"
          @click="activeTab = 'pending'"
        >
          ยังไม่มี Tag ({{ pendingCount }})
        </button>
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'done' }"
          @click="activeTab = 'done'"
        >
          วิเคราะห์แล้ว ({{ doneCount }})
        </button>
      </div>
    </div>

    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <p v-else-if="loading" class="subtitle" role="status">กำลังอ่านข้อมูลจากฐานข้อมูล…</p>

    <section v-else-if="filteredPhotos.length">
      <div class="photo-list">
        <article v-for="photo in paginatedPhotos" :key="photo.id" class="photo-row">
          <div class="photo-thumb">
            <img v-if="photoThumbs[photo.id]" :src="photoThumbs[photo.id]" :alt="photo.fileName" />
            <div v-else class="thumb-placeholder"><AppIcon name="gallery" /></div>
          </div>
          <div class="photo-info">
            <h2>{{ photo.fileName }}</h2>
            <p class="photo-path">{{ photo.destinationPath }}</p>
            <small>{{ formatSize(photo.fileSize) }} · สำรองเมื่อ {{ formatDate(photo.backedUpAt) }}</small>
            <div class="tag-section">
              <p class="tag-label">Tag</p>
              <div v-if="photoTags[photo.id]?.length" class="tag-list">
                <span v-for="tag in photoTags[photo.id]" :key="tag.id" class="tag-chip">
                  {{ tag.name }}
                  <button class="tag-remove" @click="removeTagFromPhoto(photo.id, tag.id)" title="ลบ Tag นี้">×</button>
                </span>
              </div>
              <p v-else class="tag-empty">ยังไม่มี Tag</p>
              <div class="tag-add-row">
                <input
                  class="tag-input"
                  v-model="newTagName[photo.id]"
                  placeholder="เพิ่ม Tag เอง..."
                  @keyup.enter="addTagToPhoto(photo.id)"
                />
                <button class="tag-add-button" :disabled="addingTag[photo.id] || !(newTagName[photo.id] || '').trim()" @click="addTagToPhoto(photo.id)">
                  {{ addingTag[photo.id] ? 'กำลังเพิ่ม...' : 'เพิ่ม' }}
                </button>
              </div>
            </div>
          </div>
          <div class="photo-actions">
            <button class="analyze-button" :disabled="analyzing[photo.id] || analyzingAll" @click="analyzePhoto(photo)">
              {{ analyzing[photo.id] ? 'กำลังวิเคราะห์…' : 'วิเคราะห์รูป' }}
            </button>
            <button class="delete-icon-button" title="ลบข้อมูลรูปนี้ออกจากฐานข้อมูล" @click="deleteSinglePhoto(photo)">
              ลบ
            </button>
            <span class="photo-status">{{ photo.backupStatus }}</span>
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
      <p class="pagination-info">แสดง {{ paginatedPhotos.length }} จาก {{ filteredPhotos.length }} รูป (หน้า {{ currentPage }} / {{ totalPages }})</p>
    </section>

    <section v-else-if="photos.length" class="connection">
      <h2>ไม่พบรูปในหมวดหมู่นี้</h2>
      <p class="subtitle">ลองเลือกแถบ "ทั้งหมด" เพื่อดูรูปภาพที่มีอยู่</p>
      <button class="secondary-button" @click="activeTab = 'all'">ดูรูปทั้งหมด</button>
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
.tag-chip { display: inline-flex; align-items: center; gap: 6px; padding: 4px 10px; border-radius: 999px; background: #f0f7f3; color: #367959; font-size: 11px; font-weight: 600; text-transform: capitalize; }
.tag-remove { border: none; background: transparent; color: #5f9f7b; cursor: pointer; font-size: 12px; line-height: 1; }
.tag-empty { margin: 0; font-size: 11px; color: #89968f; }
.tag-add-row { display: flex; gap: 8px; }
.tag-input { flex: 1; border-radius: 999px; border: 1px solid #d7e3dc; padding: 6px 12px; font-size: 12px; }
.tag-add-button { border-radius: 999px; border: 1px solid #d7e3dc; background: #fff; padding: 6px 12px; font-size: 12px; color: #2f6b4a; cursor: pointer; }
.tag-add-button:disabled { opacity: 0.5; cursor: not-allowed; }
.photo-actions { display: flex; flex-direction: column; align-items: flex-end; gap: 10px; }
.header-actions { display: flex; flex-wrap: wrap; gap: 10px; }
.analyze-button { padding: 8px 14px; border-radius: 999px; border: 1px solid #d7e3dc; background: linear-gradient(180deg, #f7fbf9 0%, #eef6f1 100%); color: #2f6b4a; font-size: 12px; font-weight: 600; cursor: pointer; }
.analyze-button:disabled { opacity: 0.6; cursor: not-allowed; }
.primary-button:disabled,
.secondary-button:disabled { opacity: 0.6; cursor: not-allowed; }
.photo-status { padding: 6px 10px; border-radius: 20px; background: #eaf3ee; color: #367959; font-size: 11px; font-weight: 600; }
.delete-icon-button {
  padding: 6px 12px;
  border-radius: 999px;
  border: 1px solid #ebd0d0;
  background: #fff5f5;
  color: #b54747;
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}
.delete-icon-button:hover {
  background: #fde8e8;
  border-color: #e5b3b3;
}
.cleanup-button {
  border-color: #edd5d5;
  color: #a43a3a;
}
.cleanup-button:hover:not(:disabled) {
  background: #fdf2f2;
}

.search-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin: 18px 0 24px;
}

.search-input {
  flex: 1 1 280px;
  border-radius: 999px;
  border: 1px solid #d7e3dc;
  padding: 9px 18px;
  font-size: 13px;
  background: #fff;
}

.search-button {
  padding: 9px 18px;
  border-radius: 999px;
  border: 1px solid #d7e3dc;
  background: linear-gradient(180deg, #f7fbf9 0%, #eef6f1 100%);
  color: #2f6b4a;
  font-weight: 600;
  font-size: 13px;
  cursor: pointer;
}

.search-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.tabs-bar {
  margin-bottom: 20px;
}
.tabs {
  display: inline-flex;
  gap: 6px;
  padding: 4px;
  background: #eef3f0;
  border-radius: 12px;
}
.tab-btn {
  padding: 7px 16px;
  border: none;
  background: transparent;
  color: #5c6c64;
  font-size: 13px;
  font-weight: 600;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
}
.tab-btn.active {
  background: #fff;
  color: #2f6b4a;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.05);
}
.photo-path {
  font-family: monospace;
  font-size: 11px;
  color: #7b8880;
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

.gallery-toolbar {
  margin-top: 24px;
  display: flex;
  justify-content: center;
}

.header-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: center;
}

.primary-button:disabled,
.secondary-button:disabled { opacity: 0.6; cursor: not-allowed; }

@media (max-width: 700px) {
  .photo-row { flex-direction: column; align-items: flex-start; }
  .photo-actions { width: 100%; flex-direction: row; justify-content: space-between; align-items: center; }
  .header-actions { width: 100%; justify-content: flex-start; }
  .photo-thumb { width: 100%; height: 180px; }
}
</style>
