# Architecture Overview

- Her exchange ayri worker servisi olarak calisir.
- Worker veriyi temp spool'a yazar.
- Parquet ve manifest basariyla yazilinca temp ham dosya silinir.
- Manifest ve checkpoint kalici olarak saklanir.
- Bir worker'in cokmesi diger exchange servislerini durdurmaz.
