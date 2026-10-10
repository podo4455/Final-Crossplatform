# Photo Backup — Final Crossplatform

โปรเจกต์กลุ่ม 3 คน: Desktop app สำหรับสำรองรูปจาก Flash drive ด้วย Go/Wails, Nuxt และ SQLite/GORM

## สถานะปัจจุบัน

เริ่มต้นโครง Wails v2 + Nuxt 4 และหน้าทดสอบ Nuxt เรียก Go ผ่าน Greet แล้ว ฟีเจอร์ Backup, SQLite/GORM, Gallery และ Analysis ยังไม่ได้พัฒนา

## เครื่องมือ

- Go 1.25 ขึ้นไป ตาม go.mod (เครื่องเริ่มต้นใช้ 1.26.4)
- Node.js 24 LTS (เครื่องเริ่มต้นใช้ 24.18.0) และ npm
- Wails CLI v2.13.0
- Windows: WebView2; macOS: Xcode Command Line Tools ตาม Wails installation guide
- Nuxt 4.6.1, Vue 3.5.43 และ Vue Router 5.1.0 (อัปเดตภายใน major เดิมจากตัวอย่างบทเรียนและล็อกใน lockfile)

## เริ่มใช้งาน

รันจาก root ของ repository:

```sh
wails doctor
cd frontend
npm ci
cd ..
go mod download
wails dev
```

กดปุ่มทดสอบการเชื่อมต่อในหน้าต่าง Wails ต้องได้รับข้อความภาษาไทยจาก Go การเปิดเฉพาะหน้าเว็บใน browser จะไม่มี Wails binding สำหรับเรียก Go

Build ตัวแอป:

```sh
wails build
```

ผลลัพธ์อยู่ใน build/bin ของ OS ที่ build ต้องทดสอบ macOS บนเครื่องจริงแยกจาก Windows

## จุดที่ควรอ่านก่อนแยกงาน

- [แผนแบ่งงาน](docs/TEAM_PLAN.md)
- [รายการ Issue](docs/ISSUE_BACKLOG.md)
- [คู่มือเริ่ม GitHub](docs/GETTING_STARTED.md)
- main.go: เปิดหน้าต่างและ Bind app
- app.go: เมธอด Go ที่ Frontend เรียกได้
- frontend/app/pages/index.vue: async/await เรียก generated binding
- frontend/nuxt.config.ts: SPA, hash routing และ generate ไป frontend/dist
- frontend/wailsjs/: Wails สร้างให้ ไม่แก้ด้วยมือ

Commit lockfiles และ source; ไม่ commit build/bin, node_modules, DB, key หรือรูปส่วนตัว

## Dependency audit

ตรวจ npm audit วันที่ 10 ตุลาคม 2026: หลังอัปเดต Nuxt/Vue ยังมี 14 รายการที่ npm รายงาน (7 high, 7 critical รวม dependencies ที่ส่งผลต่อกัน) ใน Nuxt/toolchain เช่น simple-git, node-forge และ braces ไม่ได้ยืนยันว่าแอป SPA นี้มีช่องทางโจมตีครบตาม advisory ทุกข้อ และไม่ได้ใช้ npm audit fix --force เพราะเสนอให้ downgrade Nuxt ข้าม major ต้องติดตามรุ่นแก้ต่อก่อนส่งงานจริง

## อ้างอิง

ยึดบทเรียน Wails ของรายวิชาเป็นหลัก โดยตรวจเพิ่มเติมจาก [Wails installation](https://v2.wails.io/docs/gettingstarted/installation/) และ [GitHub Flow](https://docs.github.com/en/get-started/using-github/github-flow)
