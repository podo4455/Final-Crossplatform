<script setup lang="ts">
import { GetPhotoPreview, GetPhotoTags, AnalyzePhoto, SetPhotoTags } from '../../wailsjs/go/main/App'
import type { database } from '../../wailsjs/go/models'
const props = defineProps<{ photo: database.Photo; analysisBusy: boolean }>()
const emit = defineEmits<{ saved: []; analyzing: [value: boolean]; search: [tag: string] }>()
const preview = ref('')
const previewError = ref('')
const tags = ref<database.Tag[]>([])
const draft = ref('')
const busy = ref(false)
const saving = ref(false)
const error = ref('')
const message = ref('')
const expanded = ref(false)
const draftTags = computed(() => [...new Set(draft.value.split(/[,\n]/).map(tag => tag.trim().toLowerCase()).filter(Boolean))])
async function loadTags() { tags.value = (await GetPhotoTags(props.photo.id)) || [] }
onMounted(async () => {
  await Promise.allSettled([
    GetPhotoPreview(props.photo.id).then(value => { preview.value = value }).catch(cause => { previewError.value = String(cause) }),
    loadTags().catch(cause => { error.value = `อ่าน Tag ไม่สำเร็จ: ${String(cause)}` }),
  ])
})
async function analyze() {
  if (busy.value || props.analysisBusy || saving.value) return
  busy.value = true; emit('analyzing', true); expanded.value = true
  error.value = ''; message.value = ''
  try {
    const values = await AnalyzePhoto(props.photo.id)
    draft.value = (values || []).join(', ')
    message.value = values?.length ? 'วิเคราะห์แล้ว ตรวจแก้ Tag ก่อนกดบันทึก' : 'ไม่พบ Tag ที่โมเดลระบุได้ สามารถเพิ่มเองได้'
  } catch (cause) { error.value = `วิเคราะห์ไม่สำเร็จ: ${String(cause)}` }
  finally { busy.value = false; emit('analyzing', false) }
}
async function save() {
  if (saving.value || busy.value || !draftTags.value.length) return
  if (draftTags.value.length > 20 || draftTags.value.some(tag => [...tag].length > 80)) {
    error.value = 'เพิ่มได้ครั้งละไม่เกิน 20 Tag และแต่ละ Tag ไม่เกิน 80 ตัวอักษร'; return
  }
  saving.value = true; error.value = ''; message.value = ''
  try {
    await SetPhotoTags(props.photo.id, draftTags.value)
    draft.value = ''; message.value = 'บันทึก Tag แล้ว'; emit('saved')
    await loadTags()
  } catch (cause) { error.value = `บันทึกหรือโหลด Tag ไม่สำเร็จ: ${String(cause)}` }
  finally { saving.value = false }
}
</script>
<template>
  <article class="photo-card">
    <div class="preview">
      <img v-if="preview" :src="preview" :alt="photo.fileName" />
      <p v-else class="subtitle">{{ previewError ? 'แสดงภาพตัวอย่างไม่ได้' : 'กำลังโหลดภาพ…' }}</p>
    </div>
    <div class="photo-content">
      <h2>{{ photo.fileName }}</h2><p class="path">{{ photo.destinationPath }}</p>
      <div class="tags" aria-label="Tag ที่บันทึกแล้ว">
        <button v-for="tag in tags" :key="tag.id" class="tag" :disabled="analysisBusy || saving" @click="emit('search', tag.name)">{{ tag.name }}</button>
        <small v-if="!tags.length">ยังไม่มี Tag</small>
      </div>
      <div class="actions">
        <button class="primary-button" :disabled="analysisBusy || saving" @click="analyze">{{ busy ? 'กำลังวิเคราะห์…' : 'วิเคราะห์ภาพ' }}</button>
        <button class="secondary-button" @click="expanded = !expanded">{{ expanded ? 'ซ่อนรายละเอียด' : 'รายละเอียด / เพิ่ม Tag' }}</button>
      </div>
      <div v-if="expanded" class="details">
        <p class="subtitle">ขนาด {{ (photo.fileSize / 1024).toFixed(1) }} KB · สำรอง {{ new Date(String(photo.backedUpAt)).toLocaleString('th-TH') }}</p>
        <p v-if="previewError" class="error">{{ previewError }}</p>
        <label :for="`tags-${photo.id}`">Tag ที่ต้องการเพิ่ม (คั่นด้วยเครื่องหมาย ,)</label>
        <textarea :id="`tags-${photo.id}`" v-model="draft" :disabled="busy || saving" rows="3" placeholder="cat, beach, sky" />
        <p class="subtitle">บันทึกเพิ่มจาก Tag เดิม โดยไม่เพิ่มคำซ้ำ ผลจากโมเดลอาจคลาดเคลื่อน แก้ไขก่อนบันทึกได้</p>
        <button class="secondary-button" :disabled="busy || saving || !draftTags.length" @click="save">{{ saving ? 'กำลังบันทึก…' : 'บันทึก Tag' }}</button>
      </div>
      <p v-if="message" class="success" role="status">{{ message }}</p>
      <p v-if="error" class="error" role="alert">{{ error }}</p>
    </div>
  </article>
</template>
<style scoped>
.photo-card { border: 1px solid #e3e9e5; background: white; border-radius: 14px; overflow: hidden; }
.preview { height: 220px; background: #edf3ef; display: grid; place-items: center; }
.preview img { width: 100%; height: 220px; object-fit: contain; }
.photo-content { padding: 22px; }
h2, .path { overflow-wrap: anywhere; }
.path, small { font-size: 12px; color: #7b8880; }
.tags, .actions { display: flex; gap: 8px; flex-wrap: wrap; margin: 14px 0; }
.tag { border: 1px solid #d4e5d9; background: #edf5ef; color: #367959; padding: 5px 10px; border-radius: 20px; }
.details { border-top: 1px solid #e3e9e5; padding-top: 16px; }
textarea { box-sizing: border-box; width: 100%; font: inherit; padding: 10px; border: 1px solid #d4e5d9; border-radius: 8px; resize: vertical; }
.error, .success { font-size: 13px; overflow-wrap: anywhere; }
</style>
