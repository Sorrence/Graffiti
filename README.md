# Graffiti

Graffiti, **yetkilendirilmiş Active Directory güvenlik değerlendirmelerinde** SMB parola denemelerini daha kontrollü yürütmek için yazılmış bir Go aracıdır. `enum4linux-ng` JSON/YAML çıktısından hedefi ve parola politikasını alır; parola listesini politika ile uyumlu girdilere indirger, kilitlenme eşiğine yaklaşan kullanıcıları pas geçer ve başarılı kimlik bilgilerini kaydeder.

> Yalnızca sahibi olduğunuz veya açık, yazılı yetkiye sahip olduğunuz sistemlerde kullanın. Yetkisiz parola denemeleri hesap kilitlenmesine, hizmet kesintisine ve hukuki sonuçlara yol açabilir.

## Özellikler

- `enum4linux-ng` tarafından üretilen JSON, YAML ve YML sonuçlarını ayrıştırır.
- Parola politikası üzerinden minimum uzunluk, karmaşıklık ve kilitlenme eşiğini çıkarır.
- Uygun olmayan parolaları denemeden önce eler.
- Kilitlenme eşiğinden bir deneme önce kullanıcıyı atlayarak riski azaltır.
- Kilitlenme gözlem penceresinden denemeler arası gecikme hesaplar.
- Başarılı kullanıcı adı/parola çiftlerini seçilen çıktı dosyasına ekler.

## Gereksinimler

- Go 1.23 veya daha yeni
- Erişilebilir SMB hizmeti (TCP/445)
- Hedef için alınmış `enum4linux-ng` JSON veya YAML çıktısı
- Kullanıcı ve parola listeleri

## Kurulum

```bash
git clone https://github.com/sorrence/graffiti.git
cd graffiti
go build -o graffiti .
```

İsterseniz bağımlılıkları önceden indirebilirsiniz:

```bash
go mod download
```

## Kullanım

Önce, yetkili kapsamınızdaki hedef için enum4linux-ng sonucu üretin:

```bash
enum4linux-ng -A -oJ enum.json <hedef>
```

Ardından kullanıcı ve parola listelerini Graffiti’ye verin:

```bash
./graffiti \
  --input enum.json \
  --userlist users.txt \
  --passlist passwords.txt \
  --output successful-logins.txt
```

Araç, `enum.json` içindeki hedef ana makinesine TCP/445 üzerinden bağlanır. Çıktı dosyasına her başarılı giriş `kullanici:parola` biçiminde eklenir.

### Seçenekler

| Seçenek | Açıklama |
| --- | --- |
| `-i`, `--input` | Zorunlu. enum4linux-ng JSON/YAML/YML çıktısı. |
| `-u`, `--userlist` | Zorunlu. Satır başına bir kullanıcı adı içeren dosya. |
| `-p`, `--passlist` | Zorunlu. Satır başına bir parola içeren dosya. |
| `-o`, `--output` | Başarılı girişlerin yazılacağı dosya. Varsayılan: `output.txt`. |
| `--min-length` | Algılanan minimum parola uzunluğunu geçersiz kılar. |
| `--lockout-threshold` | Algılanan hesap kilitleme eşiğini geçersiz kılar. |
| `--delay` | Denemeler arasındaki gecikmeyi geçersiz kılar; ör. `30s`, `2m`. |
| `-v`, `--verbose` | Başarısız denemeleri de ekrana yazar. |

Örnek: değerlendirme planınızın gerektirdiği gecikmeyi açıkça ayarlamak için:

```bash
./graffiti -i enum.yaml -u users.txt -p passwords.txt --delay 45s -v
```

## Parola politikası ve hız kontrolü

Varsayılan olarak Graffiti, enum4linux-ng çıktısındaki aşağıdaki alanları kullanır:

- Minimum parola uzunluğu
- Parola karmaşıklığı
- Hesap kilitleme eşiği
- Kilitlenme gözlem penceresi

Karmaşıklık etkinse parola en az üç kategoriyi içermelidir: büyük harf, küçük harf, rakam ve sembol. Gecikme otomatik belirlendiğinde, gözlem penceresi ve kilitlenme eşiğine göre %20 güvenlik marjı uygulanır. Açıkça verilen CLI seçenekleri bu değerlerin önüne geçer.

Bu kontroller bir güvenlik garantisi değildir. Özellikle üretim ortamında kapsam, izin, bakım penceresi ve hesap kilitlenmesi geri-dönüş planını değerlendirmeden çalıştırmayın.

## Geliştirme

```bash
go test ./...
go vet ./...
```

## Lisans

Bu depoda henüz bir lisans dosyası bulunmuyor. Dağıtım veya yeniden kullanım öncesinde proje sahibiyle lisans koşullarını netleştirin.
