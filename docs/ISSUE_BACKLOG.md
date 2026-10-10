# รายการ Issue สำหรับนำไปสร้างใน GitHub

รหัส A1/B1/C1 ด้านล่างเป็นรหัสแผน ไม่ใช่เลข Issue จริง เรียงตาม milestone ใน TEAM_PLAN.md

| รหัส | ชื่อ Issue | Owner | Reviewer | ต้องรอ | เกณฑ์ Done |
|---|---|---|---|---|---|
| A1 | ตั้ง Wails + Nuxt และ build สอง OS | A | C | — | เปิดหน้าจอจริงและ build ได้; ระบุ version/คำสั่ง |
| B1 | ออกแบบ schema, DTO และ API contract | B | A,C | — | ทุกคนตกลง destination/photo/history/tag/status และรูปแบบ error/events |
| C1 | ทดลองวิเคราะห์เนื้อหาภาพจริง | C | B | — | JPEG/PNG ให้ Tag ที่มาจากภาพ; พิสูจน์วิธีใช้บนสอง OS และระบุข้อจำกัด |
| B2 | SQLite/GORM migration และ repository | B | A | B1 | สร้าง/อ่าน/update ข้อมูลได้; destination แยกกัน; เปิดใหม่ข้อมูลอยู่ |
| A2 | เลือก source/destination และ scan รูป | A | B | A1,B1 | เลือกโฟลเดอร์/ยกเลิกได้, ตรวจ path และแสดงรายการรูป |
| A3 | ส่ง thumbnail เข้า Nuxt | A | B | A1,B1 | Gallery แสดงรูปจากดิสก์ทั้ง dev/build บนสอง OS |
| A4 | Backup selected/all + Copy/Move | A | B | A2,B2 | ข้ามไดรฟ์ได้, ไม่เขียนทับ, DB failure ไม่ลบต้นทาง; ผลรายไฟล์ชัด |
| A5 | หน้า Backup, progress และสรุปเวลา | A | C | A4 | UI รับ events และแสดง success/skipped/failed/elapsed; cleanup listener |
| B3 | Gallery และประวัติแยก destination | B | C | B2,A3 | แสดงหลายรูปพร้อม pagination; เปลี่ยนปลายทางแล้วข้อมูลไม่ปะปน |
| C2 | Analysis service และบันทึก Tag | C | B | C1,B2 | ใช้ analyzer จริง, status/retry, normalize และไม่สร้างความสัมพันธ์ซ้ำ |
| C3 | รายละเอียดรูปและ UI วิเคราะห์ | C | A | A3,C2 | แสดงสถานะ/Tag และลองใหม่ได้; error ไม่ทำให้รูป Backup หาย |
| B4 | ค้นรูปตาม Tag | B | C | B3,C2 | เลือก Tag แล้วได้รูปตรงเงื่อนไขหลายรูป; empty state/pagination |
| C4 | Tag aggregation และ Tag Cloud | C | B | B2,C2 | ขนาดตาม distinct photo count, กรอง destination/status; ส่ง selection ให้ Gallery |
| B5 | Integrity และ Delete UI/Go | B | A | B2,B3 | พบ missing files จริง; ลบแล้ว DB สอดคล้อง/แจ้ง partial failure |
| A6 | ทดสอบ Backup failure และข้ามไดรฟ์จริง | A | B | A4,A5 | หลักฐาน USB, duplicate, disk/permission/DB failure, ชื่อไทย; แก้บั๊กที่พบ |
| B6 | ทดสอบ DB/search/integrity สอง OS | B | C | B4,B5,C4 | ประวัติ,ค้น,cloud หลัง restart/delete ตรงกับข้อมูลจริง |
| C5 | ทดสอบ analyzer และ workflow ครบ | C | A | C3,C4,B4 | Backup→analysis→cloud→search ทำงานจริง; invalid image/timeout/retry ผ่าน |
| R1 | รวม release และซ้อม Demo | A ประสาน ทุกคนร่วม | B,C | A6,B6,C5 | build ของสอง OS, README, test evidence และทุกคนอธิบายฟีเจอร์ตัวเอง |

อย่าเปรียบเทียบผลงานจากจำนวน Issue อย่างเดียว A ดูแลงาน file transfer/integration, B ดูแลข้อมูลและ query, C ดูแลงานวิเคราะห์ที่ต้องทดลอง ทุกคนร่วมทดสอบและทำ Demo

ตัวอย่าง Labels: area:backup, area:library, area:analysis, type:feature, type:bug, type:test, priority:must, priority:optional

Milestones: M0 Foundation, M1 Backup, M2 Analysis & Search, M3 Requirements Complete, M4 Release

งานเสริมหลังเกณฑ์หลักผ่าน: caption + ช่องค้นข้อความ, recursive scan, content hash, HEIC, รองรับโทรศัพท์โดยตรง และ thumbnail cache ไม่รวมใน Done ของขอบเขตหลัก
