# ผลตรวจโครงเริ่มต้น

วันที่ 10 ตุลาคม 2026 บน Windows/amd64

- wails doctor: พร้อมพัฒนา มี WebView2
- npm run generate: ผ่าน Nuxt 4.6.1 / Vue 3.5.43 และสร้าง frontend/dist
- go test ./...: ผ่านการ compile package; ยังไม่มี unit tests ของฟีเจอร์
- wails build -s: ผ่าน โดยใช้ frontend ที่ generate แล้ว ได้ build/bin/photo-backup.exe
- npm run dev -- --host 127.0.0.1: พร้อมที่ port 3000; หยุด server หลังตรวจ
- เปิด executable จริงและกด “ทดสอบการเชื่อมต่อ”: ได้ “สวัสดี ทีม Photo Backup — Nuxt เรียก Go ผ่าน Wails ได้แล้ว”

ยังไม่ได้ทดสอบ macOS, Backup, SQLite/GORM หรือ Analyzer ใน milestone นี้

ข้อสังเกต: npm audit ยังมี 14 รายการตาม README ต้องติดตาม dependency patches ต่อ

## ตรวจซ้ำให้ตรงบทเรียน

- คัดลอก nuxt.config.ts จาก 05-Wails Application/01-Initial Wails Project.md โดยเพิ่มเฉพาะ css ของหน้าทดสอบ และตรวจข้อความเทียบต้นฉบับแล้ว
- forceHmrLocalhost, DevTools enabled, SSR false, production publicDir ../frontend/dist และ hashMode true ตรงบทเรียน
- wails.json ใช้ npm install, npm run generate, npm run dev และ serverUrl auto ตรงบทเรียน
- package.json ใช้ dependency ranges ^4.4.7 / ^3.5.35 / ^5.1.0 ตรงตัวอย่าง; lockfile resolve Nuxt 4.6.1 / Vue 3.5.43 / Router 5.1.0
- Windows ใช้ dev: nuxt dev; เก็บคำสั่ง macOS TMPDIR=/tmp nuxt dev ใน dev:mac ตามหัวข้อที่ระบุ macOS โดยตรง ไม่มี wrapper script ที่เพิ่มเองแล้ว
- wails dev ผ่าน และคลิกปุ่มเรียก Go ในหน้าต่าง dev ได้ข้อความภาษาไทยจริง
- แก้หัวข้อชั่วคราวเป็น HMR check แล้วเห็นหน้าต่างอัปเดตโดยไม่เปิดใหม่ จากนั้นคืนข้อความเดิมแล้ว
- wails build แบบเต็ม (ไม่ใช้ -s) ผ่าน ได้ build/bin/photo-backup.exe
- DevTools มี warning เรื่อง cross-origin ของ Wails; ไม่ได้ยืนยันว่า DevTools ทุกเมนูใช้ได้
- .gitkeep เป็นข้อเพิ่มสำหรับ Go embed บน clone ใหม่ ไม่ใช่โค้ดในบทเรียน
- ชื่อโปรเจกต์ ผู้เขียน และหน้าทดสอบใช้ของกลุ่ม ไม่ใช่ foodie-app ตัวอย่าง

## UI และโครงเลือกโฟลเดอร์ — 11 ตุลาคม 2026

- มีหน้าภาพรวมและหน้า /backup แบบ hash router; AppShell, AppIcon, PhotoIllustration และ FolderSelector แยกเป็น components
- เพิ่ม Go SelectSource / SelectDestination พร้อม generated bindings; UI จัดการการยกเลิก การกดซ้ำ และข้อความ error
- npm run generate ผ่านหลังปรับ UI สร้าง frontend/dist แล้ว มี warning cache-driver ของ Nitro; ยังไม่ได้แก้ toolchain ในงาน UI นี้
- เปิด Nuxt preview ใน browser ตรวจหน้าจอจริงและกดเปลี่ยนหน้าระหว่างภาพรวม/สำรองรูปได้ บันทึกภาพใน docs/previews/
- คง .gitkeep สำหรับ clone ใหม่; ไม่ commit generated dist อื่น ๆ
- ยังไม่ได้ทดสอบ dialog ของ UI รุ่นนี้ใน native app ซ้ำ และไม่ได้ build executable รุ่น UI นี้ซ้ำ ผล native/build ด้านบนเป็นผลตรวจโครงก่อนปรับหน้าตา
- ยังไม่มีการ scan, copy/move, SQLite/GORM, คลังรูป, วิเคราะห์ภาพ หรือค้นหา Tag; ปุ่มเริ่มสำรองถูกปิดไว้
- ตรวจ Go ก่อน push: go test ./... ผ่านการ compile; package ยังไม่มี unit tests
