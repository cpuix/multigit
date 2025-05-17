# Multigit Geliştirme Yol Haritası

Bu belge, Multigit projesinin Homebrew'a çıkış sürecindeki görevleri ve ilerlemeyi takip etmek için oluşturulmuştur.

## 🚀 Öncelikli Görevler

### 1. Test Kapsamını Geliştirme
- [x] Test kapsamını %90+ seviyesine çıkar
  - [x] `AddSSHKeyToAgent` fonksiyonu için testler
  - [x] `AddSSHConfigEntry` fonksiyonu için testler
  - [x] `DeleteSSHKey` fonksiyonu için edge case testleri
    - [x] Mevcut anahtarı silme
    - [x] Olmayan anahtarı silme
    - [x] Aynı hesap için birden fazla anahtar tipini silme
    - [x] İzin hatası durumlarını işleme
  - [x] `DeleteSSHKeyFile` fonksiyonu için testler
    - [x] Mevcut anahtarı silme
    - [x] Olmayan anahtarı silme
    - [x] İzin hatası durumlarını test etme
  - [x] Entegrasyon testleri eklendi
  - [x] SSH anahtar üretimi iyileştirildi
    - [x] RSA ve ED25519 anahtar tipleri için destek eklendi
    - [x] OpenSSH formatında anahtar üretimi sağlandı
    - [x] Hata yönetimi iyileştirildi

### 2. Kapsamlı Dokümantasyon
- [ ] README.md dosyasını güncelle:
  - [ ] Kurulum talimatları
  - [ ] Kullanım kılavuzu
  - [ ] Örnekler
  - [ ] Katkıda bulunma rehberi
- [ ] CONTRIBUTING.md dosyası oluştur
- [ ] CHANGELOG.md dosyası oluştur

### 3. CI/CD Pipeline Kurulumu
- [x] GitHub Actions ile CI/CD iş akışı oluşturuldu:
  - [x] Test otomasyonu (Linux, macOS, Windows)
  - [x] Lint kontrolleri (golangci-lint)
  - [x] Otomatik release oluşturma (GoReleaser entegrasyonu)
  - [x] Test kapsamı raporlama (Codecov entegrasyonu)
  - [x] Çoklu platform desteği (amd64, arm64)
  - [x] Homebrew ve Scoop paket yönetimi

## 🛠️ Orta Vadeli Görevler

### 4. Hata Yönetimi ve Loglama
- [ ] Standart hata tipleri oluştur
- [ ] Yapılandırılabilir loglama mekanizması ekle
- [ ] Hata kodlarını standartlaştır

### 5. Performans İyileştirmeleri
- [ ] Profilleme yaparak darboğazları tespit et
- [ ] Büyük ölçekli testler yap
- [ ] Bellek kullanımını optimize et

### 6. Çoklu Platform Desteği
- [ ] Windows'ta test et
- [ ] Linux'ta test et
- [ ] Platforma özgü hataları düzelt

## 📦 Yayın Öncesi Hazırlıklar

### 7. Lisans ve Yasal Uyumluluk
- [ ] Uygun lisansı seç (MIT/Apache 2.0)
- [ ] Üçüncü parti lisans uyumluluğunu kontrol et

### 8. Paketleme
- [ ] Homebrew formülü oluştur
- [ ] Kurulum betikleri hazırla
- [ ] Sürüm etiketleme stratejisi belirle

## 📊 İlerleme Durumu

| Kategori               | Durum      | Tamamlanma | Notlar |
|------------------------|------------|------------|---------|
| Test Kapsamı          | 🔄 Devam   | 85%        | SSH modülü testleri neredeyse tamamlandı |
| Dokümantasyon         | ⏳ Bekliyor| 0%         |         |
| CI/CD                 | ⏳ Bekliyor| 0%         |         |
| Hata Yönetimi        | ⏳ Bekliyor| 0%         |         |
| Performans           | ⏳ Bekliyor| 0%         |         |
| Çoklu Platform Desteği| ⏳ Bekliyor| 0%         |         |
| Lisans               | ⏳ Bekliyor| 0%         |         |
| Paketleme            | ⏳ Bekliyor| 0%         |         |

## 🏗️ Şu Anki Görev

- [ ] Entegrasyon testleri yaz
  - [ ] Tam bir SSH anahtarı yaşam döngüsünü test et
  - [ ] Gerçek SSH agent ile entegrasyon testleri
  - [ ] SSH config yönetimi testleri

## 📅 Son Güncelleme
- 2025-05-18: `DeleteSSHKeyFile` fonksiyonu için kapsamlı testler eklendi.
- 2025-05-18: `DeleteSSHKey` fonksiyonu için edge case testleri tamamlandı.
- 2025-05-18: ROADMAP.md oluşturuldu ve ilk görevler eklendi.
