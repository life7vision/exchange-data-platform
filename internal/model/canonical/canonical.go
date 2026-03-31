package canonical

type EnvelopeRow struct {
	Exchange         string `parquet:"name=exchange, type=BYTE_ARRAY, convertedtype=UTF8"`
	Region           string `parquet:"name=region, type=BYTE_ARRAY, convertedtype=UTF8"`
	Market           string `parquet:"name=market, type=BYTE_ARRAY, convertedtype=UTF8"`
	Dataset          string `parquet:"name=dataset, type=BYTE_ARRAY, convertedtype=UTF8"`
	Symbol           string `parquet:"name=symbol, type=BYTE_ARRAY, convertedtype=UTF8"`
	EventTimeUnixMs  int64  `parquet:"name=event_time_unix_ms, type=INT64"`
	IngestTimeUnixMs int64  `parquet:"name=ingest_time_unix_ms, type=INT64"`
	JobID            string `parquet:"name=job_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	ProducerVersion  string `parquet:"name=producer_version, type=BYTE_ARRAY, convertedtype=UTF8"`
	PayloadJSON      string `parquet:"name=payload_json, type=BYTE_ARRAY, convertedtype=UTF8"`
}
