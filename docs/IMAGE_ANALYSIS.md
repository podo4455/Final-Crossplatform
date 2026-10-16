# วิเคราะห์ภาพและบันทึก Tag

ใช้ Ollama + `gemma3:4b` บนเครื่อง ไม่ใช้ API key ตัวแอปส่งภาพเฉพาะ `127.0.0.1:11434` และไม่ตาม HTTP redirect
ต้องดาวน์โหลด Ollama และโมเดลครั้งแรก หลังจากนั้นวิเคราะห์ได้แบบ offline

## เริ่มใช้งาน

Windows เครื่องที่มี portable runtime ใน `.tools/ollama`:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/start-local-model.ps1
wails dev
```

สำหรับเพื่อนในทีมบน Windows/macOS: ติดตั้ง [Ollama](https://ollama.com/download) และเปิดบริการ Ollama
จากนั้นรัน `ollama pull gemma3:4b` แล้ว `wails dev` (หากใช้ CLI ให้เปิด `ollama serve` ในอีก terminal)
ตัว runtime และโมเดลไม่ commit เข้า Git; ต้องติดตั้งแต่ละเครื่อง

1. สำรองรูปเพื่อให้มีรายการใน SQLite
2. เปิด **คลังรูปภาพ** กด **วิเคราะห์ภาพ** ใต้รูป
3. รอผล Tag แล้วตรวจ/แก้คำในช่อง เช่น `cat, beach, sky`
4. กด **บันทึก Tag** เพื่อเรียก `SetPhotoTags(photoId, tags)` จาก Nuxt
5. กดคำใต้รูปหรือเลือกคำจาก **Tag Cloud** เพื่อค้นหา หรือเปิดเมนู **ค้นหาด้วย Tag**

## Data flow

`GetPhotos()` → รูปพร้อม `id` และ `destinationPath` → Nuxt เรียก `AnalyzePhoto(id)`
→ Go อ่าน destinationPath จากฐานข้อมูล → `service.AnalyzeImage(path)` → Ollama → `[]string`
→ Nuxt แสดงร่าง Tag → ผู้ใช้กดบันทึก → `SetPhotoTags(id, tags)` → SQLite

ใช้ ID ที่ขอบเขต Wails เพื่อให้ backend ตรวจว่ารูปอยู่ในฐานข้อมูลจริง ก่อนอ่าน path
ใช้ `GetPhotoTags` แสดง Tag ที่บันทึกแล้ว, `SearchPhotosByTag` ค้นหาตรงคำ,
และ `GetTagCloud` แสดงคำพร้อมจำนวนรูปและปรับขนาดตามความถี่

`SetPhotoTags` ของเดิมเป็นการ **เพิ่ม** Tag ไม่ลบหรือแทนที่ Tag เดิม
ระบบตัดช่องว่างหัวท้าย แปลงเป็นตัวพิมพ์เล็ก และไม่เพิ่มความสัมพันธ์ซ้ำ
การแก้ในช่องร่างจึงมีผลต่อ Tag ที่กำลังจะเพิ่มเท่านั้น ไม่มี UI ลบ Tag ที่บันทึกแล้ว

## ขอบเขตและการจัดการข้อผิดพลาด

- วิเคราะห์และสร้างภาพตัวอย่างจาก JPEG/PNG/GIF (เฟรมแรก) ที่อ่านได้ ขนาดไม่เกิน 32 MB / 40 ล้านพิกเซล
- รูปชนิดอื่นยังคัดลอกและเพิ่ม Tag เองได้ ต้องแปลงเป็น JPEG/PNG ก่อนวิเคราะห์
- ย่อภาพสำหรับโมเดลด้านยาวไม่เกิน 1024 px; ตัวอย่าง 480 px โดยไม่แก้ไฟล์ต้นฉบับ
- ภาพ JPEG ที่ต้องหมุนตาม EXIF อาจแสดงตามแนวพิกเซลเดิม ยังไม่ได้หมุนตาม metadata
- วิเคราะห์ครั้งละรูป มี timeout 5 นาที; แสดงข้อผิดพลาดเมื่อไฟล์หาย โมเดลไม่พร้อม หรือผล JSON ไม่ถูกต้อง
- ผล AI เป็นร่าง ต้องตรวจความถูกต้องก่อนบันทึก ไม่ถือว่าการวิเคราะห์สำเร็จเมื่อบันทึกฐานข้อมูลล้มเหลว
- ไม่มีการสร้าง Tag จากชื่อไฟล์หรือผลจำลองเมื่อ Ollama ใช้ไม่ได้
- หน้า Gallery แบ่ง 12 รูปต่อหน้า จำกัดการอ่านภาพตัวอย่างทีละภาพเพื่อลด RAM
- เปลี่ยน GORM SQLite driver เป็น `github.com/glebarez/sqlite` เพื่อทำงานเมื่อ `CGO_ENABLED=0`; ยังใช้ SQLite และ schema เดิม

## ตรวจสอบ

```sh
go test ./...
go vet ./...
cd frontend
npm run generate
```

ทดสอบโมเดลจริงด้วยรูปตัวอย่างและฐานข้อมูลชั่วคราว (ไม่แก้คลังของผู้ใช้):

```powershell
$env:PHOTO_BACKUP_TEST_IMAGE = 'C:\path\to\cat.jpg'
go test -run TestLocalVisionIntegration -v .
```

เอกสารอ้างอิง: [Ollama Generate API](https://docs.ollama.com/api/generate),
[Vision](https://docs.ollama.com/capabilities/vision), [Gemma 3](https://ollama.com/library/gemma3),
[GORM pure-Go SQLite driver](https://github.com/glebarez/sqlite).

## ผลตรวจวันที่ 16 ตุลาคม 2026

- `go test ./...` และ `go vet ./...` ผ่าน รวมทดสอบ protocol ของ Ollama ด้วย HTTP test server และ workflow ฐานข้อมูลจริงแบบชั่วคราว
- `npm run generate` และ `wails build -s -skipbindings` ผ่านบน Windows
- ทดสอบด้วย headless Chrome ผ่าน: ร่าง Tag ก่อนบันทึก, แก้คำ, บันทึก, จำนวน Tag Cloud, ค้นหา, ผลลัพธ์ว่าง, กรณี Ollama ไม่พร้อม, เมนู และ layout จอแคบ (Wails API จำลองสำหรับ UI test)
- โมเดลจริงยังอยู่ระหว่างดาวน์โหลด จึงยังไม่ยืนยันคุณภาพ Tag จากภาพจริง
- ยังไม่ได้รันแอปบน macOS
