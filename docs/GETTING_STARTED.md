# เริ่มทำงานร่วมกันบน GitHub

## 1. ผู้สร้าง repository

repository ของกลุ่ม: https://github.com/podo4455/Final-Crossplatform

เชื่อม remote ของเครื่องเริ่มต้นแล้ว และต่อประวัติจาก README ที่สร้างไว้บน GitHub

เพิ่มเพื่อนทั้งสองคนใน Settings → Collaborators (ชื่อเมนูอาจต่างตามประเภทบัญชี) และให้เพื่อนรับคำเชิญ

## 2. Push โครงเริ่มต้น

เมื่อทดสอบโครงผ่านและตรวจไฟล์ที่จะ commit แล้ว:

```sh
git add .
git diff --cached --stat
git commit -m "chore: initialize Wails Nuxt project and team plan"
# remote origin ของเครื่องเริ่มต้นตั้งไว้แล้ว
git push -u origin main
```

การ sign in GitHub หรือกรอก credentials ให้เจ้าของบัญชีทำเอง อย่าส่ง token/password ในแชต

## 3. เพื่อนทั้งสองคน

```sh
git clone https://github.com/podo4455/Final-Crossplatform.git
cd Final-Crossplatform
cd frontend
npm ci
cd ..
go mod download
wails dev
```

ติดตั้ง Go, Node และ Wails ก่อนตาม README แล้วกดทดสอบ Nuxt ↔ Go บนเครื่องตัวเอง

## 4. งานแรกของแต่ละคน

- A: ใช้ scaffold นี้เป็นฐาน ทำ Issue A2 เลือก source/destination และ scan รูป หลังตกลง DTO ใน B1
- B: ทำ Issue B1 ตกลง schema/DTO ร่วมกับทีม แล้ว B2 SQLite/GORM
- C: เริ่ม Issue C1 ทดลองวิเคราะห์รูปจริงทันที ระหว่างรอฐานข้อมูล

แต่ละคนเปิด branch ต่อ Issue ไม่ push งานฟีเจอร์ตรง main และส่ง PR ให้เพื่อน review

## 5. ตรวจโครงบนเครื่อง

```sh
go test ./...
cd frontend
npm run generate
cd ..
wails build
```

สำหรับ scaffold ไม่มี unit test ของฟีเจอร์ ยังต้องเปิดแอปจริงและกดปุ่มเชื่อม Go ให้ผ่านด้วย รวมถึงให้สมาชิกที่ใช้ macOS ทดลองแยก
