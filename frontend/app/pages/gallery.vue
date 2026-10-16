<script setup lang="ts">
import { GetPhotos } from '../../wailsjs/go/main/App'
import type { database } from '../../wailsjs/go/models'

const photos = ref<database.Photo[]>([])
const loading = ref(true)
const error = ref('')

async function loadPhotos() {
  loading.value = true
  error.value = ''

  try {
    photos.value = await GetPhotos()
  } catch (cause) {
    error.value = `อ่านคลังรูปไม่สำเร็จ: ${String(cause)}`
  } finally {
    loading.value = false
  }
}

function formatDate(value: unknown) {
  const date = new Date(String(value))

  if (Number.isNaN(date.getTime())) {
    return '-'
  }

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
      <button class="secondary-button" :disabled="loading" @click="loadPhotos">
        {{ loading ? 'กำลังโหลด…' : 'รีเฟรชรายการ' }}
      </button>
    </div>

    <p v-if="error" class="error" role="alert">{{ error }}</p>

    <p v-else-if="loading" class="subtitle" role="status">
      กำลังอ่านข้อมูลจากฐานข้อมูล…
    </p>

    <section v-else-if="photos.length" class="photo-list" aria-label="รายการรูปที่สำรอง">
      <article v-for="photo in photos" :key="photo.id" class="photo-row">
        <div class="feature-icon">
          <AppIcon name="gallery" />
        </div>

        <div class="photo-info">
          <h2>{{ photo.fileName }}</h2>
          <p>{{ photo.destinationPath }}</p>
          <small>
            {{ formatSize(photo.fileSize) }} · สำรองเมื่อ {{ formatDate(photo.backedUpAt) }}
          </small>
        </div>

        <span class="photo-status">{{ photo.backupStatus }}</span>
      </article>
    </section>

    <section v-else class="connection">
      <h2>ยังไม่มีรูปในคลัง</h2>
      <p class="subtitle">
        ไปที่หน้าสำรองรูปภาพ เลือกรูป และคัดลอกให้สำเร็จ
        แล้วกลับมาหน้านี้เพื่อดูข้อมูลจาก SQLite
      </p>
      <NuxtLink class="primary-button" to="/backup">
        ไปหน้าสำรองรูปภาพ
        <AppIcon name="arrow" />
      </NuxtLink>
    </section>
  </main>
</template>

<style scoped>
.photo-list {
  display: grid;
  gap: 12px;
}

.photo-row {
  display: flex;
  align-items: center;
  gap: 18px;
  padding: 18px;
  background: #fff;
  border: 1px solid #e3e9e5;
  border-radius: 14px;
}

.photo-info {
  min-width: 0;
  flex: 1;
}

.photo-info h2 {
  font-size: 15px;
  margin-bottom: 4px;
}

.photo-info p {
  overflow-wrap: anywhere;
  color: #7b8880;
  font-size: 12px;
  margin: 0 0 6px;
}

.photo-info small {
  color: #89968f;
  font-size: 11px;
}

.photo-status {
  padding: 6px 10px;
  border-radius: 20px;
  background: #eaf3ee;
  color: #367959;
  font-size: 11px;
  font-weight: 600;
}

@media (max-width: 600px) {
  .photo-row {
    align-items: flex-start;
  }

  .photo-status {
    display: none;
  }
}
</style>