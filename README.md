# 🗄️ DB Backup Pro

**PostgreSQL, MySQL ve MongoDB veritabanlarınız için profesyonel yedekleme çözümü**

[![Build](https://img.shields.io/github/actions/workflow/status/melihozkara/db-backup-pro/build.yml?branch=main)](https://github.com/melihozkara/db-backup-pro/actions)
[![Release](https://img.shields.io/github/v/release/melihozkara/db-backup-pro)](https://github.com/melihozkara/db-backup-pro/releases)
[![License](https://img.shields.io/github/license/melihozkara/db-backup-pro)](LICENSE)

> **Not:** Bu proje açık kaynak olarak geliştirilmektedir. İşinize yararlı bulduysanız ⭐ **yıldızlamayı unutmayın**, bu beni motive ediyor!

---

## 📖 Türkçe

### 🎯 Nedir?

DB Backup Pro, veritabanı yedeklemelerinizi kolayca yönetmenizi sağlayan cross-platform bir masaüstü uygulamasıdır. Veritabanlarınızı otomatik olarak yedekleyip, yerel disk, FTP, SFTP veya S3 uyumlu depolamaya kaydeder.

### ✨ Özellikler

- 🎯 **Çoklu Veritabanı Desteği**
  - PostgreSQL (pg_dump ile)
  - MySQL/MariaDB (mysqldump ile)
  - MongoDB (mongodump ile)

- 📦 **Esnek Depolama Seçenekleri**
  - Yerel disk (klasör)
  - FTP / FTPS
  - SFTP (SSH)
  - S3 / MinIO (S3 uyumlu depolama)

- ⏰ **Otomatik Zamanlama**
  - Manuel yedekleme
  - Dakika bazlı aralıklı
  - Günlük (belirli saatte)
  - Haftalık (belirli günler ve saatte)

- 🔐 **Güvenlik**
  - AES-256 şifreleme desteği
  - Gzip sıkıştırma
  - Şifreli SQLite veritabanı (yerel veri koruması)

- 🗂️ **Gelişmiş Organizasyon**
  - Tarih bazlı klasör gruplandırma (günlük/aylık/yıllık)
  - Özelleştirilebilir dosya ön eki
  - Otomatik eski yedekleri temizleme (retention)

- 📱 **Bildirimler**
  - Telegram bildirimleri (başarılı/başarısız)
  - Gerçek zamanlı durum güncellemeleri

- 🖥️ **İki Mod**
  - **Desktop Mode:** Wails ile native masaüstü arayüzü
  - **Web Mode:** HTTP server ile tarayıcıdan erişim

- 🌍 **Çoklu Dil Desteği**
  - Türkçe 🇹🇷
  - English 🇬🇧

### 📥 Kurulum

#### Windows
1. [Releases](https://github.com/melihozkara/db-backup-pro/releases) sayfasından en son `dbbackup-windows-amd64.zip` dosyasını indirin
2. ZIP dosyasını çıkartın
3. `dbbackup.exe` dosyasını çalıştırın

> **Not:** MySQL, PostgreSQL ve MongoDB araçları Windows paketi ile birlikte gelir!

#### macOS
1. [Releases](https://github.com/melihozkara/db-backup-pro/releases) sayfasından en son `dbbackup-macos-universal.zip` dosyasını indirin
2. ZIP dosyasını çıkartın
3. `DB Backup Pro.app` uygulamasını Applications klasörüne taşıyın
4. İlk açılışta güvenlik uyarısı alırsanız: **Sistem Tercihleri > Güvenlik ve Gizlilik** bölümünden uygulamaya izin verin

> **Gereksinimler:** Homebrew ile araçları yükleyin:
> ```bash
> brew install postgresql mysql-client mongodb/brew/mongodb-database-tools
> ```

#### Linux
1. [Releases](https://github.com/melihozkara/db-backup-pro/releases) sayfasından en son `dbbackup-linux-amd64.tar.gz` dosyasını indirin
2. TAR dosyasını çıkartın:
   ```bash
   tar -xzf dbbackup-linux-amd64.tar.gz
   cd dbbackup
   ./dbbackup
   ```

> **Gereksinimler:** Veritabanı araçlarını yükleyin:
> ```bash
> # Debian/Ubuntu
> sudo apt install postgresql-client mysql-client mongodb-database-tools
>
> # RHEL/AlmaLinux/Rocky
> sudo dnf install postgresql mysql mongodb-database-tools
> ```

### 🚀 Kullanım

#### Desktop Mode (Varsayılan)
Uygulamayı çift tıklayarak açın. Eğer WebView2 (Windows) bulunamazsa otomatik olarak web moduna geçer.

#### Web Mode (Sunucu)
```bash
# Varsayılan port (8090)
./dbbackup serve

# Özel port
./dbbackup serve -port 3000 -host 0.0.0.0
```

Tarayıcıdan `http://localhost:8090` adresine gidin.

#### Temel İş Akışı
1. **Veritabanı Ekle:** Veritabanları sekmesinden bağlantı bilgilerini girin
2. **Depolama Hedefi Ekle:** Yedeklerin nereye kaydedileceğini belirleyin
3. **Yedekleme Görevi Oluştur:** Veritabanı + Depolama + Zamanlama ayarlayın
4. **Aktif Edin:** Görev otomatik çalışmaya başlasın

### 🛠️ Teknoloji Stack

- **Backend:** Go 1.24, Wails v2
- **Frontend:** React 18, TypeScript, TailwindCSS
- **Veritabanı:** SQLite (SQLCipher şifreli)
- **Zamanlanmış Görevler:** gocron v2
- **Storage:** AWS SDK v2 (S3), pkg/sftp, jlaffaye/ftp

### 🏗️ Geliştirme

#### Gereksinimler
- Go 1.24+
- Node.js 20+
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

#### Yerel Geliştirme
```bash
# Bağımlılıkları yükle
cd frontend && npm install && cd ..
go mod download

# Geliştirme modunda çalıştır
wails dev
```

#### Production Build
```bash
# Windows
wails build -platform windows/amd64

# macOS
wails build -platform darwin/universal

# Linux
wails build -platform linux/amd64
```

### 🤝 Katkıda Bulunma

Katkılarınızı bekliyorum! Lütfen:
1. Fork edin
2. Feature branch oluşturun (`git checkout -b feature/amazing-feature`)
3. Commit edin (`git commit -m 'feat: add amazing feature'`)
4. Push edin (`git push origin feature/amazing-feature`)
5. Pull Request açın

### 📄 Lisans

Bu proje [MIT lisansı](LICENSE) altında lisanslanmıştır.

### 🙏 Teşekkürler

Bu proje açık kaynak yazılım sayesinde mümkün oldu. Kullandığım harika kütüphanelere teşekkürler!

---

## 📖 English

### 🎯 What is it?

DB Backup Pro is a cross-platform desktop application that helps you easily manage your database backups. It automatically backs up your databases and saves them to local disk, FTP, SFTP, or S3-compatible storage.

### ✨ Features

- 🎯 **Multiple Database Support**
  - PostgreSQL (with pg_dump)
  - MySQL/MariaDB (with mysqldump)
  - MongoDB (with mongodump)

- 📦 **Flexible Storage Options**
  - Local disk (folder)
  - FTP / FTPS
  - SFTP (SSH)
  - S3 / MinIO (S3-compatible storage)

- ⏰ **Automatic Scheduling**
  - Manual backup
  - Interval-based (minutes)
  - Daily (at specific time)
  - Weekly (specific days and time)

- 🔐 **Security**
  - AES-256 encryption support
  - Gzip compression
  - Encrypted SQLite database (local data protection)

- 🗂️ **Advanced Organization**
  - Date-based folder grouping (daily/monthly/yearly)
  - Customizable file prefix
  - Automatic old backup cleanup (retention)

- 📱 **Notifications**
  - Telegram notifications (success/failure)
  - Real-time status updates

- 🖥️ **Dual Mode**
  - **Desktop Mode:** Native desktop UI with Wails
  - **Web Mode:** Browser access via HTTP server

- 🌍 **Multi-language Support**
  - Türkçe 🇹🇷
  - English 🇬🇧

### 📥 Installation

#### Windows
1. Download the latest `dbbackup-windows-amd64.zip` from [Releases](https://github.com/melihozkara/db-backup-pro/releases)
2. Extract the ZIP file
3. Run `dbbackup.exe`

> **Note:** MySQL, PostgreSQL, and MongoDB tools are bundled with the Windows package!

#### macOS
1. Download the latest `dbbackup-macos-universal.zip` from [Releases](https://github.com/melihozkara/db-backup-pro/releases)
2. Extract the ZIP file
3. Move `DB Backup Pro.app` to Applications folder
4. On first launch, if you get a security warning: Allow the app in **System Preferences > Security & Privacy**

> **Requirements:** Install tools via Homebrew:
> ```bash
> brew install postgresql mysql-client mongodb/brew/mongodb-database-tools
> ```

#### Linux
1. Download the latest `dbbackup-linux-amd64.tar.gz` from [Releases](https://github.com/melihozkara/db-backup-pro/releases)
2. Extract the TAR file:
   ```bash
   tar -xzf dbbackup-linux-amd64.tar.gz
   cd dbbackup
   ./dbbackup
   ```

> **Requirements:** Install database tools:
> ```bash
> # Debian/Ubuntu
> sudo apt install postgresql-client mysql-client mongodb-database-tools
>
> # RHEL/AlmaLinux/Rocky
> sudo dnf install postgresql mysql mongodb-database-tools
> ```

### 🚀 Usage

#### Desktop Mode (Default)
Double-click the application to open. If WebView2 (Windows) is not found, it automatically switches to web mode.

#### Web Mode (Server)
```bash
# Default port (8090)
./dbbackup serve

# Custom port
./dbbackup serve -port 3000 -host 0.0.0.0
```

Navigate to `http://localhost:8090` in your browser.

#### Basic Workflow
1. **Add Database:** Enter connection details in the Databases tab
2. **Add Storage Target:** Specify where backups will be saved
3. **Create Backup Job:** Configure Database + Storage + Schedule
4. **Activate:** The job will start running automatically

### 🛠️ Technology Stack

- **Backend:** Go 1.24, Wails v2
- **Frontend:** React 18, TypeScript, TailwindCSS
- **Database:** SQLite (encrypted with SQLCipher)
- **Scheduled Tasks:** gocron v2
- **Storage:** AWS SDK v2 (S3), pkg/sftp, jlaffaye/ftp

### 🏗️ Development

#### Requirements
- Go 1.24+
- Node.js 20+
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

#### Local Development
```bash
# Install dependencies
cd frontend && npm install && cd ..
go mod download

# Run in development mode
wails dev
```

#### Production Build
```bash
# Windows
wails build -platform windows/amd64

# macOS
wails build -platform darwin/universal

# Linux
wails build -platform linux/amd64
```

### 🤝 Contributing

I welcome contributions! Please:
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### 📄 License

This project is licensed under the [MIT License](LICENSE).

### 🙏 Acknowledgments

This project is made possible by open source software. Thanks to all the amazing libraries I use!

---

<div align="center">

**⭐ Projeyi faydalı bulduysan yıldızlamayı unutma! / If you find this project useful, don't forget to star it! ⭐**

Made with ❤️ by Melih Özkara

</div>
