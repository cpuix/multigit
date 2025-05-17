# Multigit Geliştirme Planı

Bu belge, Multigit projesinin geliştirme aşamalarını ve her bir adım için gerekli AI prompt'larını içerir.

## Aşama 1: Temel Altyapı ve SSH Entegrasyonu

### 1.1 SSH Anahtar Yönetimi

**Prompt:**
```
Hedef: internal/ssh/ssh.go dosyasını tamamla.

Gereksinimler:
1. CreateSSHKey fonksiyonunu, belirtilen isim ve e-posta ile RSA 4096 bit SSH anahtarı oluşturacak şekilde implemente et
2. Anahtarlar ~/.ssh/ dizini altında "id_rsa_<account_name>" formatında kaydedilmeli
3. İsteğe bağlı passphrase desteği ekle
4. Hata durumlarını uygun şekilde yönet
5. Başarılı işlem sonrası oluşturulan public key'i ekrana yazdır

Kütüphaneler:
- crypto/rsa, crypto/rand (anahtar oluşturma için)
- encoding/pem (anahtar formatlama için)
- os (dosya işlemleri için)
- path/filepath (platform bağımsız path işlemleri için)
```

### 1.2 SSH Agent Entegrasyonu

**Prompt:**
```
Hedef: SSH anahtarlarını SSH agent'a ekleme işlevselliğini tamamla.

Gereksinimler:
1. AddSSHKeyToAgent fonksiyonunu, belirtilen hesap adına ait SSH anahtarını SSH agent'a ekleyecek şekilde implemente et
2. Eğer SSH agent çalışmıyorsa uygun hata mesajı döndür
3. Anahtar zaten agent'a ekliyse tekrar ekleme
4. İşlem başarılı olursa onay mesajı döndür

Kütüphaneler:
- golang.org/x/crypto/ssh/agent (SSH agent işlemleri için)
- net (Unix domain socket bağlantısı için)
- os/user (kullanıcı bilgilerine erişim için)
```

## Aşama 2: SSH Konfigürasyon Yönetimi

### 2.1 SSH Config Dosya Yönetimi

**Prompt:**
```
Hedef: SSH config dosyasına yeni host girişleri ekleme işlevselliğini tamamla.

Gereksinimler:
1. AddSSHConfigEntry fonksiyonunu, ~/.ssh/config dosyasına yeni bir host girişi ekleyecek şekilde implemente et
2. Host girişi şu şekilde olmalı:
   Host github.com-<account_name>
       HostName github.com
       User git
       IdentityFile ~/.ssh/id_rsa_<account_name>
       IdentitiesOnly yes
3. Eğer config dosyası yoksa oluştur
4. Aynı host zaten varsa üzerine yazma, hata döndür
5. Dosya izinlerini güvenli bir şekilde ayarla (chmod 600)

Kütüphaneler:
- os (dosya işlemleri için)
- path/filepath
- bufio (dosya okuma/yazma için)
```

### 2.2 SSH Config'den Giriş Silme

**Prompt:**
```
Hedef: SSH config dosyasından belirli bir hesaba ait girişi kaldır.

Gereksinimler:
1. RemoveSSHConfigEntry fonksiyonunu, belirtilen hesap adına ait host girişini kaldıracak şekilde implemente et
2. Eğer host bulunamazsa uygun hata döndür
3. Config dosyası yoksa hata döndür
4. İşlem başarılı olursa onay mesajı döndür
```

## Aşama 3: Kullanıcı Arayüzü ve Komutlar

### 3.1 Use Komutu

**Prompt:**
```
Hedef: cmd/use.go dosyasını implemente et.

Gereksinimler:
1. "use" adında yeni bir komut ekle
2. Kullanım: multigit use <account_name>
3. Belirtilen hesabı aktif hale getir:
   - İlgili SSH anahtarını SSH agent'a ekle
   - Global git kullanıcı adı ve e-postasını güncelle
   - Aktif hesap bilgisini kaydet (~/.multigit/active_account)
4. Eğer hesap yoksa hata döndür
5. Başarılı olursa onay mesajı göster

Kullanılacak paketler:
- github.com/spf13/cobra
- os/exec (git komutlarını çalıştırmak için)
- path/filepath
```

### 3.2 List Komutu

**Prompt:**
```
Hedef: cmd/list.go dosyasını oluştur ve list komutunu implemente et.

Gereksinimler:
1. "list" adında yeni bir komut ekle
2. Tüm kayıtlı hesapları listele
3. Aktif hesabı işaretle
4. Her hesap için:
   - Hesap adı
   - İlişkili e-posta
   - SSH anahtar durumu (mevcut/değil)
   - Son kullanım tarihi (eğer mevcutsa)
5. Eğer hiç hesap yoksa bilgilendirici mesaj göster
```

## Aşama 4: Gelişmiş Özellikler

### 4.1 Profil Yönetimi

**Prompt:**
```
Hedef: Profil yönetimi için gerekli yapıyı oluştur.

Gereksinimler:
1. "profile" adında yeni bir komut ekle
2. Alt komutlar:
   - create <profile_name>: Yeni profil oluştur
   - delete <profile_name>: Profili sil
   - list: Tüm profilleri listele
   - add-account <profile_name> <account_name>: Profile hesap ekle
   - remove-account <profile_name> <account_name>: Profilden hesap kaldır
3. Profil bilgilerini ~/.multigit/profiles/ altında JSON formatında sakla
4. "use" komutuna --profile parametresi ekle
```

### 4.2 Otomatik GitHub Entegrasyonu

**Prompt:**
```
Hedef: GitHub API entegrasyonu ile SSH anahtarlarını otomatik yönet.

Gereksinimler:
1. "github" adında yeni bir komut ekle
2. Alt komutlar:
   - add-key <account_name>: Oluşturulan public key'i GitHub hesabına ekle
   - remove-key <account_name>: GitHub'dan SSH anahtarını kaldır
3. GitHub API v3 kullan
4. Personal Access Token ile kimlik doğrulama
5. Kullanıcıdan gerekli izinleri iste
```

## Aşama 5: Test ve Dokümantasyon

### 5.1 Test Kapsamı

**Prompt:**
```
Hedef: Kapsamlı testler yaz.

Gereksinimler:
1. Her public fonksiyon için unit test yaz
2. Entegrasyon testleri ekle
3. Mock kullanarak dış bağımlılıkları yönet
4. Test kapsamını ölç ve en az %80 hedefle
5. GitHub Actions ile CI/CD iş akışı oluştur
```

### 5.2 Kullanım Kılavuzu

**Prompt:**
```
Hedef: Kapsamlı bir README.md dosyası oluştur.

İçerik:
1. Proje açıklaması ve özellikler
2. Kurulum talimatları
3. Kullanım kılavuzu ve örnekler
4. Yapılandırma seçenekleri
5. Geliştirme ve katkı kılavuzu
6. Lisans bilgisi

Ayrıca:
- Komut referansı (her komut için detaylı açıklama)
- Sık karşılaşılan sorunlar
- Katkıda bulunma rehberi
```

## Sonraki Adımlar

1. Her bir aşamayı sırayla tamamla
2. Her özellik için önce testleri yaz, sonra implementasyonu yap
3. Değişiklikleri küçük commit'ler halinde yap
4. Her özellik için dokümantasyonu güncelle
