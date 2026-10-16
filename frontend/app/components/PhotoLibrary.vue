<script setup lang="ts">
import { GetPhotos, SearchPhotosByTag, GetTagCloud } from '../../wailsjs/go/main/App'
import type { database } from '../../wailsjs/go/models'
defineProps<{ searchPage?: boolean }>()
const photos = ref<database.Photo[]>([])
const cloud = ref<database.TagCloudItem[]>([])
const query = ref('')
const activeTag = ref('')
const loading = ref(false)
const analysisBusy = ref(false)
const error = ref('')
const page = ref(1)
const pageCount = computed(() => Math.max(1, Math.ceil(photos.value.length / 12)))
const visiblePhotos = computed(() => photos.value.slice((page.value-1)*12, page.value*12))
const maxCount = computed(() => Math.max(1, ...cloud.value.map(item => item.photoCount)))
async function refreshCloud() { cloud.value = (await GetTagCloud()) || [] }
async function load(tag = '') {
  if (loading.value || analysisBusy.value) return
  loading.value = true; error.value = ''
  try {
    const normalized = tag.trim().toLowerCase()
    const values = normalized ? await SearchPhotosByTag(normalized) : await GetPhotos()
    photos.value = values || []
    activeTag.value = normalized; query.value = normalized; page.value = 1
    await refreshCloud()
  } catch (cause) { error.value = `อ่านข้อมูลไม่สำเร็จ: ${String(cause)}` }
  finally { loading.value = false }
}
async function saved() {
  try { await refreshCloud() }
  catch (cause) { error.value = `บันทึกแล้ว แต่รีเฟรช Tag Cloud ไม่สำเร็จ: ${String(cause)}` }
}
onMounted(() => load())
onBeforeRouteLeave(() => !analysisBusy.value)
</script>
<template>
  <main class="shell">
    <p class="eyebrow">{{ searchPage ? 'FIND YOUR PHOTOS' : 'PHOTO LIBRARY' }}</p>
    <h1>{{ searchPage ? 'ค้นหาด้วย Tag' : 'คลังรูปภาพ' }}</h1>
    <p class="subtitle">วิเคราะห์รูปด้วยโมเดลในเครื่อง ตรวจแก้ Tag แล้วบันทึกเพื่อค้นหาภายหลัง</p>
    <details class="connection setup">
      <summary>เตรียมโมเดลวิเคราะห์ภาพในเครื่อง</summary>
      <p>ติดตั้งและเปิด Ollama แล้วดาวน์โหลดโมเดลด้วย <code>ollama pull gemma3:4b</code> เพียงครั้งแรก จากนั้นวิเคราะห์ได้โดยไม่ใช้ API key</p>
      <p>รองรับการวิเคราะห์ JPEG, PNG และ GIF (เฟรมแรก) ไม่เกิน 32 MB / 40 ล้านพิกเซล รูปชนิดอื่นยังสำรองได้และเพิ่ม Tag เองได้</p>
      <p>การวิเคราะห์อาจใช้เวลาหลายนาที โดยเฉพาะครั้งแรกที่โหลดโมเดล ภาพส่งให้ Ollama บนเครื่องนี้เท่านั้น</p>
    </details>
    <form class="search-form" @submit.prevent="load(query)">
      <label for="tag-search">ค้นหาด้วย Tag หนึ่งคำ เช่น cat หรือ beach</label>
      <div class="input-row">
        <input id="tag-search" v-model="query" :disabled="loading || analysisBusy" placeholder="พิมพ์ Tag ที่ต้องการค้นหา" />
        <button class="primary-button" :disabled="loading || analysisBusy">ค้นหา</button>
        <button type="button" class="secondary-button" :disabled="loading || analysisBusy" @click="load()">ดูทั้งหมด</button>
      </div>
    </form>
    <section class="cloud-panel" aria-label="Tag Cloud">
      <h2>Tag Cloud</h2>
      <div v-if="cloud.length" class="cloud">
        <button v-for="tag in cloud" :key="tag.id" :disabled="loading || analysisBusy" :aria-pressed="activeTag === tag.name" :style="{ fontSize: `${14 + 14*tag.photoCount/maxCount}px` }" @click="load(tag.name)">{{ tag.name }} <small>({{ tag.photoCount }})</small></button>
      </div>
      <p v-else class="subtitle">ยังไม่มี Tag — วิเคราะห์หรือเพิ่ม Tag ให้รูปในคลังก่อน</p>
    </section>
    <div class="section-heading"><h2>{{ photos.length }} รูป{{ activeTag ? ` · Tag: ${activeTag}` : '' }}</h2>
      <button class="secondary-button" :disabled="loading || analysisBusy" @click="load(activeTag)">รีเฟรช</button>
    </div>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <p v-if="loading" class="subtitle" role="status">กำลังโหลด…</p>
    <section v-else-if="photos.length" class="photo-grid" aria-label="รูปภาพ">
      <PhotoCard v-for="photo in visiblePhotos" :key="photo.id" :photo="photo" :analysis-busy="analysisBusy" @analyzing="analysisBusy = $event" @saved="saved" @search="load" />
    </section>
    <section v-else-if="!error" class="connection">
      <h2>{{ activeTag ? 'ไม่พบรูปที่มี Tag นี้' : 'ยังไม่มีรูปในคลัง' }}</h2>
      <p class="subtitle">{{ activeTag ? 'ลองเลือกคำใน Tag Cloud หรือกดดูทั้งหมด' : 'สำรองรูปภาพก่อน แล้วกลับมาวิเคราะห์และเพิ่ม Tag' }}</p>
      <NuxtLink v-if="!activeTag" class="primary-button" to="/backup">ไปหน้าสำรองรูป</NuxtLink>
    </section>
    <div v-if="pageCount > 1" class="pagination">
      <button class="secondary-button" :disabled="page === 1 || analysisBusy" @click="page--">ก่อนหน้า</button><span>หน้า {{ page }} / {{ pageCount }}</span>
      <button class="secondary-button" :disabled="page === pageCount || analysisBusy" @click="page++">ถัดไป</button>
    </div>
  </main>
</template>
<style scoped>
.setup { margin: 20px 0; }
.search-form { margin: 24px 0; }
.cloud-panel { border: 1px solid #e3e9e5; border-radius: 14px; padding: 22px; background: white; }
.cloud { display: flex; flex-wrap: wrap; align-items: center; gap: 12px; }
.cloud button { color: #367959; background: #edf5ef; border: 1px solid #d4e5d9; border-radius: 20px; padding: 6px 12px; }
.cloud button[aria-pressed="true"] { color: white; background: #367959; }
.cloud small { font-size: 12px; }
.photo-grid { display: grid; grid-template-columns: repeat(2,minmax(0,1fr)); gap: 20px; }
.pagination { display: flex; align-items: center; justify-content: center; gap: 16px; margin-top: 24px; }
@media (max-width: 900px) { .photo-grid { grid-template-columns: 1fr; } }
</style>
