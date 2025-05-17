# MultiGit

MultiGit, birden fazla GitHub hesabı arasında kolayca geçiş yapmanızı sağlayan bir komut satırı aracıdır. SSH anahtarlarınızı yönetir ve git konfigürasyonlarını otomatik olarak ayarlar.

[![Go](https://github.com/cpuix/multigit/actions/workflows/test.yml/badge.svg)](https://github.com/cpuix/multigit/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/cpuix/multigit/graph/badge.svg?token=YOUR-TOKEN)](https://codecov.io/gh/cpuix/multigit)
[![Go Report Card](https://goreportcard.com/badge/github.com/cpuix/multigit)](https://goreportcard.com/report/github.com/cpuix/multigit)

## Özellikler

- 🚀 Birden fazla GitHub hesabı için SSH anahtarı oluşturma ve yönetme
- 🔄 Hesaplar arasında hızlı geçiş
- 📊 Profil yönetimi ile hesapları gruplama
- 🔒 SSH anahtarlarını güvenli bir şekilde yönetme
- ⚡ SSH config dosyasını otomatik olarak yönetme
- 🎨 Renkli ve kullanıcı dostu arayüz
- ✅ %85+ test kapsamı

## Kurulum

### Go ile Kurulum (Geliştiriciler için)

1. Go'yu yükleyin (1.21 veya üzeri)
2. MultiGit'i kurun:

```bash
go install github.com/cpuix/multigit@latest
```

### Binary İndir (Kullanıcılar için)

[En son sürümü indirin](https://github.com/cpuix/multigit/releases/latest) ve PATH'inize ekleyin.

## Hızlı Başlangıç

### Yeni bir hesap ekleme

```bash
multigit create <hesap_adi> <email@example.com>
```

Örnek:
```bash
multigit create is-hesabi isim.soyisim@firma.com
multigit create kisisel ben@mailim.com -p "güçlü-şifre"
```

### Hesaplar arasında geçiş yapma

```bash
multigit use <hesap_adi>
```

### Profil Yönetimi

```bash
# Yeni profil oluştur
multigit profile create <profil_adi>

# Profilleri listele
multigit profile list

# Profil kullan
multigit profile use <profil_adi>

# Profil sil
multigit profile delete <profil_adi>
```

### Diğer Komutlar

```bash
# Hesapları listele
multigit list

# Aktif hesabı göster
multigit status

# Hesap sil
multigit delete <hesap_adi>
```

## Geliştirme

### Testleri Çalıştırma

```bash
# Tüm testleri çalıştır
make test

# Test kapsamını görüntüle
make cover

# Lint kontrolü
make lint
```

### Yapılandırma

MultiGit yapılandırması `~/.config/multigit/config.json` dosyasında saklanır. Bu dosyayı manuel olarak düzenleyebilir veya komut satırı arayüzünü kullanabilirsiniz.

## Katkıda Bulunma

Katkılarınızı bekliyoruz! Lütfen önce bir konu açarak neyi değiştirmek istediğinizi tartışın.

1. Fork'layın
2. Yeni bir branch oluşturun (`git checkout -b yeni-ozellik`)
3. Değişikliklerinizi commit'leyin (`git commit -am 'Yeni özellik eklendi'`)
4. Branch'e push'layın (`git push origin yeni-ozellik`)
5. Pull Request açın

### Test Kapsamı

Tüm yeni özellikler için birim testleri eklenmelidir. Test kapsamı en az %85 olmalıdır.

## Lisans

Bu proje [MIT lisansı](LICENSE) altında lisanslanmıştır.
