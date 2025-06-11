# TOPSIS API Documentation

## Overview

TOPSIS (Technique for Order Preference by Similarity to Ideal Solution) adalah metode pengambilan keputusan multi-kriteria yang digunakan untuk mengidentifikasi solusi dari sejumlah alternatif berdasarkan kedekatan dengan solusi ideal positif dan negatif.

## Fitur Utama

1. Kalkulasi TOPSIS
2. Penyimpanan Hasil
3. Update Alternatif dengan Kalkulasi Ulang
4. Riwayat Perhitungan
5. Autentikasi User

## Endpoints

### 1. Autentikasi

```http
POST /api/login
```

Login user dan mendapatkan token untuk akses API.

#### Request Body

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

#### Response

```json
{
  "token": "test_token_user@example.com"
}
```

#### register request

POST /api/signup

````json
{
    "email": "test1@example4.com",
    "password": "password123",
    "confirm_password": "password123",
    "nama_lengkap": "John Doe"
}```

```json
{
    "message": "Succes Create User"
}
```
### logout
http://localhost:8080/api/logout
{
    "message": "Logged out successfully"
}

### 2. Kalkulasi TOPSIS
```http
POST /api/topsis
````

Melakukan perhitungan TOPSIS berdasarkan input yang diberikan.

#### Request Body

```json
{
  "criteria": [
    {
      "name": "cost",
      "weight": 0.5,
      "type": "cost"
    },
    {
      "name": "quality",
      "weight": 0.5,
      "type": "benefit"
    }
  ],
  "alternatives": [
    {
      "name": "Alternative 1",
      "values": {
        "cost": 100,
        "quality": 8
      }
    }
  ]
}
```

### 3. Simpan Hasil TOPSIS

```http
POST /api/topsis/save
```

Menyimpan hasil perhitungan TOPSIS ke database.

#### Request Body

```json
{
  "name": "Product Selection Analysis",
  "data": {
    "idealPositive": {
      "cost": 0.2,
      "quality": 0.8
    },
    "idealNegative": {
      "cost": 0.8,
      "quality": 0.2
    },
    "results": [
      {
        "name": "Product A",
        "closenessvalue": 0.75,
        "rank": 1,
        "normalizedvalues": {
          "cost": 0.5,
          "quality": 0.5
        },
        "WeightedValues": {
          "cost": 0.4,
          "quality": 0.25
        }
      }
    ]
  },
  "raw_input": {
    "alternatives": ["Product A"],
    "criteria": {
      "cost": "cost",
      "quality": "benefit"
    },
    "values": [[100, 8]],
    "weights": [0.5, 0.5]
  }
}
```

### 4. Update Alternatif dengan Kalkulasi Ulang

```http
PUT /api/topsis/{id}
```

Mengupdate alternatif yang ada dan melakukan kalkulasi ulang TOPSIS.

#### Request Body

```json
{
  "alternatives": [
    {
      "name": "Alternative 1",
      "values": [
        {
          "criteria_name": "cost",
          "value": 100
        },
        {
          "criteria_name": "quality",
          "value": 8
        }
      ]
    }
  ]
}
```

#### Response

```json
{
  "message": "Topsis calculation updated successfully",
  "calculation_id": 1,
  "result": {
    "idealPositive": {
      "cost": 0.2,
      "quality": 0.8
    },
    "idealNegative": {
      "cost": 0.8,
      "quality": 0.2
    },
    "results": [
      {
        "name": "Alternative 1",
        "closenessvalue": 0.75,
        "rank": 1,
        "normalizedvalues": {
          "cost": 0.5,
          "quality": 0.5
        },
        "WeightedValues": {
          "cost": 0.4,
          "quality": 0.25
        }
      }
    ]
  }
}
```

### 5. Riwayat Perhitungan

```http
GET /api/topsis/history
```

Mendapatkan riwayat perhitungan TOPSIS untuk user yang sedang login.

#### Response

```json
{
  "message": "Topsis history fetched successfully",
  "data": [
    {
      "id": 1,
      "name": "Product Selection Analysis",
      "raw_data": {
        "alternatives": ["Product A", "Product B"],
        "criteria": {
          "cost": "cost",
          "quality": "benefit"
        },
        "values": [
          [100, 8],
          [150, 9]
        ],
        "weights": [0.5, 0.5]
      },
      "ideal_solutions": [
        {
          "criteria_name": "cost",
          "ideal_positive": 0.2,
          "ideal_negative": 0.8
        }
      ],
      "alternatives": [
        {
          "name": "Product A",
          "closeness_value": 0.75,
          "rank": 1,
          "criteria_values": [
            {
              "criteria_name": "cost",
              "normalized_value": 0.5,
              "weighted_value": 0.4
            }
          ]
        }
      ]
    }
  ]
}
```

## Cara Penggunaan

### 1. Autentikasi

1. Login dengan email dan password
2. Simpan token yang diterima
3. Gunakan token untuk semua request berikutnya

### 2. Kalkulasi Awal

1. Siapkan data kriteria dan alternatif
2. Kirim request POST ke `/api/topsis`
3. Simpan hasil kalkulasi dengan request POST ke `/api/topsis/save`

### 3. Update Alternatif

1. Siapkan data alternatif yang akan diupdate
2. Kirim request PUT ke `/api/topsis/{id}`
3. Sistem akan:
   - Memvalidasi input
   - Mengupdate data alternatif
   - Melakukan kalkulasi ulang TOPSIS
   - Menyimpan hasil kalkulasi terbaru
   - Mengembalikan hasil kalkulasi dalam response

### 4. Melihat Riwayat

1. Kirim request GET ke `/api/topsis/history`
2. Sistem akan mengembalikan semua perhitungan TOPSIS yang pernah dilakukan

## Catatan Penting

1. Setiap update alternatif akan memicu kalkulasi ulang TOPSIS
2. Hasil kalkulasi terbaru akan otomatis tersimpan di database
3. Riwayat perhitungan akan selalu menampilkan hasil kalkulasi terbaru
4. Pastikan format input sesuai dengan yang diharapkan
5. Bobot kriteria harus berjumlah 1
6. Nilai alternatif harus sesuai dengan tipe kriteria (benefit/cost)
7. Semua endpoint (kecuali login) memerlukan token autentikasi

## Troubleshooting

### Error: Invalid Input

- Pastikan format JSON sesuai dengan yang diharapkan
- Periksa tipe data nilai (harus number)
- Pastikan nama kriteria sesuai dengan yang ada di database

### Error: Calculation Failed

- Periksa jumlah nilai alternatif (harus sama dengan jumlah kriteria)
- Pastikan bobot kriteria berjumlah 1
- Periksa tipe kriteria (benefit/cost)

### Error: Database Error

- Periksa koneksi database
- Pastikan user memiliki akses ke database
- Periksa struktur tabel di database

### Error: Authentication

- Pastikan token valid dan belum expired
- Pastikan token dikirim dalam header Authorization
- Format token harus: "Bearer <token>"

## Contoh Penggunaan

### Login

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Update Alternatif

```bash
curl -X PUT http://localhost:8080/api/topsis/1 \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "alternatives": [
      {
        "name": "Product A",
        "values": [
          {
            "criteria_name": "cost",
            "value": 100
          },
          {
            "criteria_name": "quality",
            "value": 8
          }
        ]
      }
    ]
  }'
```

### Lihat Riwayat

```bash
curl -X GET http://localhost:8080/api/topsis/history \
  -H "Authorization: Bearer your_token"
```

## Status Test

Semua endpoint telah diuji dan berhasil:

1. ✅ Autentikasi (Login)
2. ✅ Kalkulasi TOPSIS
3. ✅ Simpan Hasil
4. ✅ Update Alternatif
5. ✅ Riwayat Perhitungan

Test mencakup:

- Validasi input
- Kalkulasi ulang
- Penyimpanan data
- Autentikasi
- Error handling

## Contoh Penggunaan API dengan httpie

### 1. Signup

```sh
http POST :8080/api/signup \
  nama_lengkap='Test User' \
  email='testuser@example.com' \
  password='password123' \
  confirm_password='password123'
```

### 2. Login (simpan session/cookie)

```sh
http --session=auth POST :8080/api/login \
  email='testuser@example.com' \
  password='password123' \
  confirm_password='password123'
```

### 3. Kalkulasi TOPSIS

```sh
http --session=auth POST :8080/api/topsis/ \
  criteria:='[{"name":"C1","weight":0.5,"type":"benefit"},{"name":"C2","weight":0.5,"type":"cost"}]' \
  alternatives:='[{"name":"A1","values":{"C1":1.0,"C2":2.0}},{"name":"A2","values":{"C1":3.0,"C2":4.0}}]'
```

### 4. Simpan Hasil Kalkulasi

```sh
http --session=auth POST :8080/api/topsis/save \
  name='Test Calculation' \
  data:='{"idealPositive":{"C1":0.4743416490252569,"C2":0.22360679774997896},"idealNegative":{"C1":0.15811388300841897,"C2":0.4472135954999579},"results":[{"name":"A2","closenessvalue":0.585786437626905,"rank":1,"normalizedvalues":{"C1":0.9486832980505138,"C2":0.8944271909999159},"WeightedValues":{"C1":0.4743416490252569,"C2":0.4472135954999579}},{"name":"A1","closenessvalue":0.4142135623730951,"rank":2,"normalizedvalues":{"C1":0.31622776601683794,"C2":0.4472135954999579},"WeightedValues":{"C1":0.15811388300841897,"C2":0.22360679774997896}}]}' \
  raw_input:='{"alternatives":["A1","A2"],"criteria":{"C1":"benefit","C2":"cost"},"values":[[1.0,2.0],[3.0,4.0]],"weights":[0.5,0.5]}'
```

### 5. Lihat Riwayat Perhitungan

```sh
http --session=auth GET :8080/api/topsis/history
```

### 6. Update Alternatif & Kalkulasi Ulang

```sh
http --session=auth PUT :8080/api/topsis/18 \
  alternatives:='[{"name":"A1","values":[{"criteria_name":"C1","value":2.0},{"criteria_name":"C2","value":1.0}]},{"name":"A2","values":[{"criteria_name":"C1","value":4.0},{"criteria_name":"C2","value":3.0}]}]'
```

### 7. Lihat Riwayat Setelah Update

```sh
http --session=auth GET :8080/api/topsis/history
```

---

**Catatan:**

- Gunakan `--session=auth` agar httpie otomatis menyimpan dan mengirim cookie autentikasi.
- ID pada endpoint update (`/api/topsis/18`) sesuaikan dengan ID hasil simpan kalkulasi.
- Setelah update, hasil kalkulasi dan riwayat akan langsung terupdate sesuai data terbaru.

---

Untuk penjelasan detail setiap endpoint, format request/response, dan troubleshooting, lihat bagian dokumentasi di atas.
# benewtopsis
