# แผน Final Project: Photo Backup — สมาชิก 3 คน

เอกสารนี้เป็นแผนงาน ยังไม่ใช่โปรแกรมที่พัฒนาเสร็จ ใช้ A/B/C แทนชื่อสมาชิก และใช้ milestone แทนวันที่เพราะยังไม่ทราบกำหนดส่ง

## 1. สิ่งที่ต้องทำตามโจทย์

ใช้ภาพ Final เป็นข้อกำหนดหลัก ข้อสอบกลางภาคและ flashMid.zip ใช้เทียบความสามารถ Backup ส่วนบทเรียน Wails ใช้เป็นแนวทางการออกแบบ ไม่ได้รันคำสั่งในเอกสารอ้างอิงโดยอัตโนมัติ

| ข้อกำหนด | แผนรองรับ | ผู้รับผิดชอบหลัก |
|---|---|---|
| Wails + Nuxt | Wails v2 ตามตัวอย่างรายวิชา + Nuxt แบบ SPA | A ตั้งโครง ทุกคนพัฒนาหน้าจอของฟีเจอร์ตัวเอง |
| SQLite + GORM | Model, migration และ repository ผ่าน GORM | B ดูแล schema ทุกคนใช้ repository |
| Backup เทียบกลางภาค | เลือก source/destination, เลือกรูปหรือทั้งหมด, ย้ายข้ามไดรฟ์, ตรวจซ้ำ, ประวัติแยกปลายทาง, จับเวลา, ตรวจไฟล์หาย, ลบและอัปเดตสถานะ | A และ B |
| วิเคราะห์คำอธิบายหรือ Tag | เลือกวิเคราะห์เนื้อหาภาพให้เป็น Tag อัตโนมัติ | C |
| ค้นหาและแสดงหลายรูป | ค้นด้วย Tag และแสดง Gallery | B |
| Tag Cloud เมื่อค้นด้วย Tag | ขนาด Tag สัมพันธ์กับจำนวนรูป คลิกเพื่อกรองได้ | C พัฒนา component, B เชื่อมการค้นหา |
| Windows และ macOS | Build และทดสอบบนเครื่องทั้งสองระบบ | ทุกคนแบ่งทดสอบตามฟีเจอร์ |

ขอบเขตหลักที่เสนอ: Flash drive, รูป JPEG/PNG, วิเคราะห์ Tag และค้นด้วย Tag ให้ครบก่อน คำอธิบายภาพเป็นงานเสริม จึงยังไม่จำเป็นต้องทำช่องค้นคำอธิบาย หากเพิ่มฟีเจอร์นั้นภายหลังต้องมีช่องให้ผู้ใช้พิมพ์คำค้นเองตามภาพโจทย์

Flash drive เป็นหนึ่งในตัวเลือกที่ภาพโจทย์อนุญาต จึงไม่ต้องเริ่มด้วยการรองรับโทรศัพท์โดยตรง หากเพิ่มโทรศัพท์ต้องแยกทดสอบการเชื่อมต่อจริง อย่าถือว่าทุกโทรศัพท์อ่านได้เป็นโฟลเดอร์เหมือน USB

Copy เป็นโหมด Backup เริ่มต้นที่เสนอ และมี Move เพื่อคงความสามารถข้อสอบกลางภาค พร้อมข้อความยืนยันการลบต้นทางใน UI การตีความนี้เป็นข้อเสนอของทีม ควรให้ผู้สอนรับทราบใน Progress Demo

## 2. ผลจากการอ่านงานกลางภาค

อ่าน source ใน ZIP โดยไม่ได้รัน executable และยังไม่ได้ทดสอบพฤติกรรมจริง

| ไฟล์เดิม | สิ่งที่ใช้เป็นพื้นฐานได้ | สิ่งที่ต้องปรับ |
|---|---|---|
| main.go | ตรวจ state, ตั้ง path, รายการไฟล์, ประวัติและจับเวลา | เปลี่ยน CLI เป็น service/DTO ที่เรียกจาก Nuxt |
| files.go | ตรวจชื่อซ้ำทั้ง DB/ดิสก์, rename และ fallback copy/delete, O_EXCL | กรองรูป, แยก skipped/failed, สถานะงาน, จัดการ failure หลังย้ายไฟล์ |
| database.go | SQLite ผ่าน github.com/glebarez/sqlite, GORM, AutoMigrate | กำหนด lifecycle connection และ migration ของ schema ใหม่ |
| models.go | unique index ของ destination + filename, status และ timestamp | เพิ่มข้อมูลรูปและตาราง Tag |
| intregity.go | ตรวจ record ที่ active เทียบไฟล์จริง | คืนรายงานให้ UI; การตรวจ hash เป็นงานเสริม ไม่ใช่ความสามารถเดิม |
| delete.go | ลบไฟล์และอัปเดตสถานะใน DB | คืนผลรายไฟล์และจัดการกรณีลบสำเร็จแต่ DB ล้มเหลว |

จุดที่ควรแก้ขณะนำมาต่อยอด:

- เดิม fallback copy/delete เมื่อ rename เกิด error ใด ๆ ควรจำแนก error ข้ามไดรฟ์กับ error อื่นให้เหมาะกับ OS
- เดิมไฟล์ถูกย้าย/ต้นทางถูกลบก่อน Create record ถ้า DB ล้มเหลวจะเกิดไฟล์ที่ไม่มีประวัติ ต้องมีสถานะและกระบวนการกู้คืน
- เดิมตรวจ integrity ว่าไฟล์ยังอยู่ ไม่ได้พิสูจน์ว่าเนื้อหาเหมือนต้นฉบับ
- เดิมตรวจซ้ำจากชื่อ ไม่ได้ตรวจเนื้อหารูปด้วย hash และอ่านไฟล์ระดับโฟลเดอร์เดียว
- record ที่ deleted ยังติดเงื่อนไขชื่อซ้ำในโค้ดเดิม ต้องระบุนโยบาย re-import ให้ชัดก่อนแก้ ไม่เปลี่ยนเงียบ ๆ
- ไม่คัดลอก .cache หรือ executable จาก ZIP ไปเป็น source ของ Final

## 3. แบ่งงานให้ทุกคนทำทั้ง Go และ Nuxt

### สมาชิก A — Backup และการเชื่อม Wails

เจ้าของฟีเจอร์นำรูปเข้าโปรแกรมตั้งแต่เลือกโฟลเดอร์จนเห็นผล Backup

- ตั้ง Wails/Nuxt ให้เปิดและ build ได้ตามบทเรียน พร้อมล็อก dependency versions
- Go: source/destination, scan รูป, Backup selected/all, Copy/Move, duplicate protection, elapsed time และผลรายไฟล์
- Nuxt: หน้า Backup, เลือกรูป, progress, สรุป copied/moved/skipped/failed และ error
- เชื่อม Wails bindings/events และออกแบบ DTO ของ Backup ร่วมกับ B
- ทดสอบ USB ข้ามไดรฟ์จริง, ชื่อไทย/ช่องว่าง, ไฟล์ซ้ำ, permission error และถอด USB ระหว่างงาน
- Demo: อธิบายไฟล์ I/O, fallback และ events ของ progress

เจ้าของไฟล์หลัก: service/backup.go, service/source.go, frontend/app/pages/backup.vue, frontend/app/components/backup/, frontend/app/composables/useBackup.ts

### สมาชิก B — ฐานข้อมูล ประวัติ และการค้นหารูป

เจ้าของการเก็บข้อมูลและค้นรูปจาก Tag ที่ C วิเคราะห์ได้

- Go: SQLite/GORM schema/migration, repositories ของ Photo/BackupRecord/Tag, history แบบ destination-scoped
- Go: query รูปตาม Tag, pagination, integrity scan และลบรูปพร้อมเปลี่ยนสถานะประวัติ
- Nuxt: Gallery, ตัวกรอง Tag, history, รายงาน Missing Files และหน้าต่างยืนยันลบ
- เชื่อม Tag Cloud ของ C เข้ากับ Gallery และรีเฟรชข้อมูลหลัง Backup/Analysis
- ทดสอบประวัติปลายทาง A/B ไม่ปะปน, ผลค้นหลายรูป, pagination, ไฟล์หาย, ลบและเปิดโปรแกรมใหม่
- Demo: อธิบาย schema, relationships, query และ integrity

เจ้าของไฟล์หลัก: model/, repository/, service/library.go, service/integrity.go, frontend/app/pages/library.vue, frontend/app/pages/history.vue, frontend/app/components/gallery/

### สมาชิก C — วิเคราะห์ภาพและ Tag Cloud

เจ้าของฟีเจอร์เพิ่มจากกลางภาค: วิเคราะห์ภาพจริงและนำผลไปใช้ค้นหา

- ทดลองตัววิเคราะห์เนื้อหาภาพหนึ่งตัวกับ JPEG/PNG ตั้งแต่ milestone แรก เลือกวิธีและล็อก dependency หลังพิสูจน์บนสอง OS
- Go: Analyzer interface, implementation จริง, normalize Tag, บันทึกผลผ่าน repository ของ B และ re-analyze
- แยก analysis status จาก backup status เพื่อให้วิเคราะห์ล้มเหลวแล้วไฟล์ Backup ยังใช้งานได้
- Nuxt: หน้ารายละเอียดรูป, ปุ่มวิเคราะห์/ลองใหม่, สถานะวิเคราะห์ และ Tag Cloud component
- Go: นับจำนวนรูปต่อ Tag เฉพาะปลายทางปัจจุบัน โดยไม่รวมรูป deleted/missing ในรายการที่แสดงได้
- ทดสอบภาพหลายประเภท, ไฟล์เสีย, analyzer timeout, การวิเคราะห์ซ้ำไม่สร้าง Tag ซ้ำ และขนาด Tag Cloud ตามจำนวน
- Demo: อธิบายการวิเคราะห์, normalization, ความสัมพันธ์รูปกับ Tag และการคำนวณ Cloud

เจ้าของไฟล์หลัก: analysis/, service/analysis.go, service/tags.go, frontend/app/components/TagCloud.vue, frontend/app/components/PhotoDetail.vue, frontend/app/composables/useAnalysis.ts

การพิมพ์ Tag เองหรือ mock ใช้ได้ระหว่างพัฒนา แต่การส่งจริงต้องมีผลวิเคราะห์เนื้อหาภาพอัตโนมัติ ไม่ใช้ชื่อไฟล์หรือนามสกุลแทนการวิเคราะห์ภาพ หากเลือกบริการภายนอกต้องระบุ internet requirement, เก็บ key ฝั่ง Go และไม่นำ key ขึ้น GitHub

ไม่กำหนดให้คนใดทำเฉพาะเอกสารหรือทดสอบ ทุกคนมี implementation, verification และ Technical Demo ของตัวเอง งานวิเคราะห์ของ C มีความไม่แน่นอนสูงจึงต้องเริ่มทดลองก่อน ไม่รอ Backup เสร็จ

## 4. โครงสร้างและจุดเชื่อม

ยึดบทเรียน Layered Architecture และ IPC:

```text
Nuxt pages/components → Wails binding → Go service → repository → SQLite/GORM
                                        ├→ file I/O
                                        └→ Analyzer
Go events → Nuxt progress/status
```

โครงสร้างที่เสนอ (ยังไม่ได้สร้างตัวแอป):

```text
main.go / app.go          composition และ binding
model/                   GORM models
dto/                     รูปแบบ request/response ที่ทีมตกลงร่วมกัน
repository/              database connection และ query
service/                 backup, library, integrity, analysis, tags
analysis/                interface และ implementation ตัววิเคราะห์
frontend/app/pages/      backup, library, history
frontend/app/components/ backup, gallery, TagCloud, PhotoDetail
frontend/app/composables/ ตัวเรียก binding และ event subscriptions
frontend/wailsjs/        generated bindings ไม่แก้ด้วยมือ
docs/                    แผน, API contract, test evidence, demo script
```

Schema เริ่มต้นให้ B นำไปตกลงกับทีม:

- Destination: normalized absolute path ที่ unique
- Photo: destination ID, relative filename, size, backed-up time, file status, analysis status
- BackupRecord: photo ID, source path, operation Copy/Move, result, duration และเวลา
- Tag: normalized name ที่ unique
- PhotoTag: photo ID + tag ID ที่ unique เพื่อไม่ผูก Tag ซ้ำ

อย่าเพิ่มระบบสมัครสมาชิก/JWT/bcrypt เพียงเพราะมีในตัวอย่างบทเรียน งานนี้ยังไม่มีเงื่อนไข login; ใช้ GORM Gen ได้ถ้าทีมต้องการตามบทเรียน แต่โจทย์บังคับ GORM ไม่ได้ระบุว่าต้อง Gen

API contract ที่ตกลงก่อนแยกงาน (เป็นชื่อเสนอ ไม่ใช่ฟังก์ชันที่มีแล้ว):

| API | เจ้าของ | ผลที่ผู้เรียกต้องรู้ |
|---|---|---|
| SelectSource / SelectDestination | A | path หรือผลการยกเลิก dialog |
| ScanSource | A | source image IDs, filename, size, preview |
| StartBackup | A | job ID; events มี completed/total และผลรายไฟล์ |
| ListPhotos / SearchPhotosByTag | B | PhotoDTO[], total, page |
| ListHistory / CheckIntegrity / DeletePhotos | B | ประวัติ/ผลตรวจ/ผลลบรายไฟล์ตาม destination |
| AnalyzePhotos | C | job ID; events มี photo ID และ analysis status |
| GetTagCloud | C ผ่าน repository B | tag ID, name, distinct photo count |
| GetPhotoPreview | A/B ร่วมกัน | thumbnail จาก photo ID สำหรับ Gallery/รายละเอียด |

ทุก query ของ library/history/cloud ต้องระบุ destination ID ทุก operation ตรวจ state ที่ Go แม้ UI ปิดปุ่มไว้แล้ว ใช้ DTO แทนส่งข้อมูล DB ทั้งก้อน และตกลง error messages/รูปแบบ event ให้ตรงกัน

ทดสอบการแสดงรูปจากดิสก์ตั้งแต่ milestone แรก ข้อเสนอเริ่มต้นคือส่ง thumbnail ขนาดจำกัดผ่าน binding ตาม ID หากเลือก asset handler ต้องพิสูจน์กับ Nuxt/Vite รุ่นที่ใช้ทั้ง dev และ build ก่อน เพราะเอกสาร Wails v2 มีข้อจำกัดกับ Vite รุ่นใหม่

Backup ขั้นต่ำต้องไม่เขียนทับไฟล์เดิม: copy ไป temporary file → ตรวจ copy/close → publish แบบไม่ทับของเดิม → บันทึก DB/status → ในโหมด Move จึงลบต้นทาง ต้องมีผลกรณีลบต้นทางไม่ได้และกรณี DB ล้มเหลวด้วย การเขียนไฟล์กับ DB ไม่ใช่ transaction เดียวกัน จึงต้องออกแบบ recovery; A/B ทำจุดนี้ร่วมกัน

## 5. ลำดับงาน ไม่ผูกกับวันที่ที่ยังไม่ทราบ

| Milestone | งานคู่ขนาน | เงื่อนไขผ่าน |
|---|---|---|
| M0: พิสูจน์โครง | A เปิด Wails/Nuxt, B schema/DTO, C ทดลอง analyzer | เปิด/build ทั้งสอง OS, แสดงรูปจากดิสก์, วิเคราะห์ภาพจริงอย่างน้อยหนึ่งรูป |
| M1: Backup พื้นฐาน | A scan/transfer/UI, B DB/history, C analysis/status/UI | รูปจาก USB เข้าเครื่องและมีประวัติ; เปิดใหม่ยังเห็นข้อมูล |
| M2: เพิ่มจากกลางภาค | A progress/errors, B Gallery/search, C Tag Cloud/retry | Backup → วิเคราะห์ → Tag Cloud → คลิกค้นหลายรูปได้ |
| M3: ความสามารถครบ | A duplicate/cross-drive, B integrity/delete, C analysis failure | checklist กลางภาคและข้อกำหนด Final ผ่าน |
| M4: ส่งงาน | ทุกคนตรวจ OS/แก้บั๊ก/ซ้อม Demo | executable ของแต่ละ OS, ผลทดสอบจริง และทุกคนอธิบายโค้ดตัวเองได้ |

ใน M0 ให้รวมเฉพาะ scaffold, DTO/interface และข้อมูลตัวอย่าง แล้วทุกคนทำกับ mock ระหว่างรอส่วนอื่น โดย M2 ต้องเปลี่ยนเป็น implementation จริงครบ

## 6. ใช้ GitHub อย่างไร

1. ใช้ repository เดียว: https://github.com/podo4455/Final-Crossplatform เพิ่มเพื่อนทั้งสองคนเป็น collaborator; ยังไม่ได้สร้าง Issue บน GitHub
2. main เก็บงานที่รวมแล้วเปิดใช้งานได้ สร้าง branch ต่อ Issue เช่น feat/backup-transfer, feat/library-search, feat/image-analysis แทน branch ส่วนตัวที่เปิดยาวจนจบงาน
3. เปิด Issue จาก ISSUE_BACKLOG.md แต่ละ Issue มีเจ้าของ 1 คนและ reviewer 1 คน งานเชื่อมอาจช่วยกันแต่ระบุผู้รับผิดชอบหลัก
4. ใช้บอร์ด Backlog → Ready → In progress → Review → Done; Ready ต้องมี contract/dependency ที่พร้อม
5. ส่ง PR ขนาดเล็ก มีสิ่งที่เปลี่ยน วิธีทดสอบ และภาพ/คลิปถ้าเป็น UI ระบุ Closes #เลขIssueจริง
6. สลับ review A→B, B→C, C→A และเพิ่มเจ้าของส่วนที่ถูกกระทบเมื่อแก้ schema/DTO
7. ตั้ง main ให้ merge ผ่าน PR และมี review 1 คน หาก repository/แผน GitHub รองรับ ไม่ push ทับงานเพื่อน
8. แต่ละคนตั้ง Git name/email ของตัวเอง เพื่อให้ commit และ PR เป็นหลักฐานผลงานของสมาชิกคนนั้น
9. owner ของ main.go/app.go/wails.json คือ A; schema/migration คือ B; shared layout คือ C; ก่อนแก้ไฟล์ร่วมให้แจ้งใน Issue/PR ไม่ห้ามคนอื่นแตะ
10. commit source และ lockfiles; ไม่ commit ภาพส่วนตัว, DB ทดสอบ, .env/key, .cache, node_modules, .nuxt, .output หรือ build artifacts; generated bindings ให้ตกลงนโยบายร่วมกันและ regenerate ตาม API

คำสั่งตัวอย่างเมื่อมี repository แล้ว:

```sh
git switch main
git pull --ff-only
git switch -c feat/backup-transfer
# ทำงานและตรวจไฟล์ที่จะ commit
git add service/backup.go frontend/app/pages/backup.vue
git commit -m "feat: add photo backup transfer"
git push -u origin feat/backup-transfer
# เปิด PR บน GitHub และขอ review
```

## 7. เกณฑ์ตรวจรับร่วมกัน

- ยังไม่เลือก destination: แจ้งเตือน ไม่ panic; path ว่าง/ภาษาไทย/มีช่องว่างใช้ได้
- Backup selected และ all ได้; ข้ามไฟล์ที่ไม่ใช่รูปตามขอบเขตที่ประกาศ
- Move USB→เครื่องจริงได้และลบต้นทางหลังยืนยันผล; Copy คงต้นทาง
- ชื่อซ้ำใน DB หรือปลายทางไม่ถูกทับ และแสดงเหตุผล
- แสดงเวลารวม และนับสำเร็จ/ข้าม/ผิดพลาดแยกกัน
- history/library/tag cloud แยก destination; ปิดเปิดแอปแล้วข้อมูลยังอยู่
- ลบรูปภายนอกแล้ว Check พบชื่อรูปที่หาย; รูปที่ผู้ใช้ลบผ่านแอปไม่รายงานเป็น missing ที่ไม่คาดหมาย
- ลบผ่านแอปแล้วไฟล์จริงและสถานะ DB สอดคล้อง หรือรายงาน partial failure
- วิเคราะห์จากเนื้อหาภาพจริง; failure retry ได้และไม่ทำ Backup เสีย
- Cloud นับจำนวนรูป distinct; คลิก Tag แล้วเห็นหลายรูปที่ตรง; empty state ชัดเจน
- ทดสอบ Windows และ macOS จาก build จริง โดยบันทึก OS, dependency versions, ขั้นตอน, expected/actual และหลักฐาน
- หากทำ CI ให้มี Go tests และ frontend generate ก่อน แล้วเพิ่ม build matrix หลังพิสูจน์ scaffold; CI build ผ่านยังไม่แทนการทดสอบ USB/WebView บนเครื่องจริง

ใช้ภาพทดสอบที่แชร์ได้ และไฟล์ทดสอบสำหรับ Move/Delete เท่านั้น ไม่ใช้รูปส่วนตัวชุดเดียวที่ยังไม่มีสำเนา

## 8. Demo และการกระจายการอธิบาย

Progress Demo ไม่เกิน 15 นาทีตามภาพ: 1 นาทีภาพรวม, A 4 นาที Backup, B 4 นาที DB/Gallery, C 4 นาที Analysis/Cloud, 2 นาทีผลทดสอบและสิ่งที่เหลือ

Technical Demo: ทุกคนอธิบายฟีเจอร์ที่ตัวเองพัฒนา โดย A แสดง Copy/Move และ duplicate, B แสดง destination-scoped history/missing/delete, C แสดงวิเคราะห์จริงและ Tag Cloud แล้ว B เชื่อมค้นรูป จัดเวลาตาม slot ที่อาจารย์กำหนด; ข้อความจำนวนกลุ่ม/5 ชั่วโมงในภาพยังไม่ใช่เวลาของกลุ่มนี้แน่นอน

เตรียม test images, USB, database และผลวิเคราะห์สำหรับสำรองการนำเสนอ แต่ต้องแยกข้อมูลที่เตรียมไว้กับผลวิเคราะห์สดให้ชัด ไม่แสดง mock เป็นฟีเจอร์ที่ทำเสร็จ

## 9. แหล่งอ้างอิง

บทเรียนใน Wails ที่ใช้เป็นหลัก:

- 03-Front End/01-Basic of Nuxt/01-Nuxt Project in Wails.md: SPA, hash routing, generate, frontend/dist
- 03-Front End/02-Vue Engine/: binding, events, dynamic content, composables
- 02-Back end (Go)/02-GORM/: model, query และ CRUD
- 04-Wails-Nuxt-IPC/01-IPC Concept.md และ 02-Bindings.md: bindings, errors และ progress events
- 04-Wails-Nuxt-IPC/03-Layer Architeccture.md และ 05-Wails Application/06-Layers.md: service/repository/model
- 05-Wails Application/01-Initial Wails Project.md, 02-Frontend.md, 04-Backend.md, 07-Form Data.md: scaffold, DaisyUI, SQLite และ forms

อ้างอิงภายนอกประกอบการทำงาน:

- [GitHub Flow](https://docs.github.com/en/get-started/using-github/github-flow): branch ต่อชุดงาน, PR และ review
- [Wails v2 Dynamic Assets](https://v2.wails.io/docs/guides/dynamic-assets/): asset handler และข้อจำกัด compatibility ที่ต้องทดลองก่อนเลือกใช้

เรื่องที่ยังต้องระบุเมื่อเริ่มจริง: ชื่อ/GitHub ของ A/B/C, วัน Progress/Final, เครื่อง macOS ที่ใช้ทดสอบ และผลทดลองเลือก analyzer ใน M0
