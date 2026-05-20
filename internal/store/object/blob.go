package object

import (
	"bytes"
	"compress/gzip"
	"io"
)

const minCompressionBytes = 1024

func MaybeCompressBlob(raw []byte) ([]byte, bool, error) {
	if len(raw) < minCompressionBytes {
		return raw, false, nil
	}
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		return nil, false, err
	}
	if _, err := zw.Write(raw); err != nil {
		if closeErr := zw.Close(); closeErr != nil {
			return nil, false, closeErr
		}
		return nil, false, err
	}
	if err := zw.Close(); err != nil {
		return nil, false, err
	}
	compressed := buf.Bytes()
	if len(compressed) >= len(raw) {
		return raw, false, nil
	}
	return compressed, true, nil
}

func MaybeDecompressBlob(raw []byte, compressed bool) ([]byte, error) {
	if !compressed {
		return raw, nil
	}
	gr, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	decompressed, err := io.ReadAll(gr)
	if closeErr := gr.Close(); closeErr != nil {
		if err == nil {
			return nil, closeErr
		}
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return decompressed, nil
}
