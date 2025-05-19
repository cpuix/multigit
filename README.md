# MultiGit

MultiGit, birden fazla GitHub hesabı arasında kolayca geçiş yapmanızı sağlayan bir komut satırı aracıdır. SSH anahtarlarınızı yönetir ve git konfigürasyonlarını otomatik olarak ayarlar.

[![Go](https://github.com/cpuix/multigit/actions/workflows/test.yml/badge.svg)](https://github.com/cpuix/multigit/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/cpuix/multigit/graph/badge.svg?token=YOUR-TOKEN)](https://codecov.io/gh/cpuix/multigit)
[![Go Report Card](https://goreportcard.com/badge/github.com/cpuix/multigit)](https://goreportcard.com/report/github.com/cpuix/multigit)
[![Test Coverage](https://img.shields.io/badge/coverage-53.5%25-green)](https://github.com/cpuix/multigit/actions)

## Özellikler

- 🚀 Birden fazla GitHub hesabı için SSH anahtarı oluşturma ve yönetme
- 🔄 Hesaplar arasında hızlı geçiş
- 📊 Profil yönetimi ile hesapları gruplama
- 🔒 SSH anahtarlarını güvenli bir şekilde yönetme
- ⚡ SSH config dosyasını otomatik olarak yönetme
- 🎨 Renkli ve kullanıcı dostu arayüz
- ✅ %58.8+ test kapsamı (artırılmaya devam ediyor)

## 📦 Kurulum

### macOS (Homebrew)

```bash
# Özel tap'ı ekleyin (sadece ilk kurulumda)
brew tap cpuix/multigit

# MultiGit'i kurun
brew install multigit
```

### Linux (DEB/RPM)

```bash
# DEB (Ubuntu/Debian)
wget https://github.com/cpuix/multigit/releases/latest/download/multigit_linux_amd64.deb
sudo dpkg -i multigit_linux_amd64.deb

# RPM (Fedora/CentOS/RHEL)
wget https://github.com/cpuix/multigit/releases/latest/download/multigit_linux_amd64.rpm
sudo rpm -i multigit_linux_amd64.rpm
```

### Windows

1. [En son sürümü indirin](https://github.com/cpuix/multigit/releases/latest)
2. İndirilen `.msi` dosyasını çalıştırın
3. Kurulum sihirbazını takip edin

### Docker ile Kullanım

```bash
# MultiGit'i çalıştır
docker run --rm -it -v ~/.ssh:/root/.ssh -v $(pwd):/workspace ghcr.io/cpuix/multigit:latest

# Veya bir alias ekleyin
echo 'alias multigit="docker run --rm -it -v ~/.ssh:/root/.ssh -v $(pwd):/workspace ghcr.io/cpuix/multigit:latest"' >> ~/.bashrc
```

### Go ile Kurulum (Geliştiriciler için)

1. Go'yu yükleyin (1.21 veya üzeri)
2. MultiGit'i kurun:

```bash
go install github.com/cpuix/multigit@latest
```

### Manuel Kurulum (Binary)

1. [En son sürümü indirin](https://github.com/cpuix/multigit/releases/latest)
2. İndirilen binary'i PATH'inize ekleyin
3. Çalıştırılabilir yapın:

```bash
chmod +x multigit
sudo mv multigit /usr/local/bin/
```

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

Tüm yeni özellikler için birim testleri eklenmelidir. Mevcut test kapsamı %58.8'tir ve artırılmaya devam edilmektedir. Test kapsamını artırmak için çalışmalar sürmektedir.

Testleri çalıştırmak için:

```bash
# Tüm testleri çalıştır
make test

# Test kapsamını görüntüle
make cover
```

## Lisans

Bu proje [MIT lisansı](LICENSE) altında lisanslanmıştır.
