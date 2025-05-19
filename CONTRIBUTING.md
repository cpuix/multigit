# Katkıda Bulunma Rehberi

MultiGit projesine katkıda bulunmak istediğiniz için teşekkür ederiz! İşte projeye nasıl katkıda bulunabileceğinize dair bir rehber.

## Geliştirme Ortamı Kurulumu

1. Go'yu yükleyin (1.21 veya üzeri)
2. Depoyu forklayın ve klonlayın:
   ```bash
   git clone https://github.com/sizinkullaniciadiniz/multigit.git
   cd multigit
   ```
3. Bağımlılıkları yükleyin:
   ```bash
   go mod download
   ```
4. Geliştirme araçlarını yükleyin:
   ```bash
   make setup
   ```

## Kodlama Standartları

- Go standart kodlama kurallarını takip edin
- Yeni özellikler için test yazın
- Kod değişikliklerinizi mantıksal commit mesajları ile kaydedin
- GoDoc yorumlarını ekleyin
- Değişikliklerinizi küçük, odaklı PR'lar halinde gönderin

## Pull Request Süreci

1. Güncel `main` branch'inden yeni bir branch oluşturun:
   ```bash
   git checkout main
   git pull origin main
   git checkout -b feature/benim-yeni-ozelligim
   ```
2. Değişikliklerinizi yapın ve commit edin
3. Testleri çalıştırın:
   ```bash
   make test
   make lint
   ```
4. Değişikliklerinizi push edin:
   ```bash
   git push -u origin feature/benim-yeni-ozelligim
   ```
5. GitHub üzerinden yeni bir Pull Request açın

## Test Yazma Rehberi

- Her yeni özellik için test yazın
- Hata düzeltmeleri için regresyon testleri ekleyin
- Test kapsamını %90'ın üzerinde tutmaya çalışın
- Tablo testleri (table-driven tests) kullanın
- Testlerin bağımsız ve tekrarlanabilir olduğundan emin olun

## Hata Bildirme

1. Önce benzer bir hata olup olmadığını kontrol edin
2. Hatanın nasıl tekrarlanacağını açıklayın
3. Beklenen ve gerçekleşen davranışı belirtin
4. İşletim sistemi, Go sürümü gibi ilgili detayları ekleyin

## Soru Sorma

Sorularınız için yeni bir Issue açabilir veya tartışmalar bölümünü kullanabilirsiniz.

## Lisans

Bu projeye yapılan tüm katkılar [MIT Lisansı](LICENSE) altında lisanslanacaktır.
