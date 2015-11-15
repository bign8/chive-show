package main

import (
	"bytes"
	"compress/gzip"

	"github.com/golang/protobuf/proto"
)

const shouldGZIP = true

// DecodeSerialGraph converts a byte string back into a hydrated SerialGraph.
func DecodeSerialGraph(data []byte) (g *SerialGraph, err error) {
	if shouldGZIP {
		if data, err = decompress(data); err != nil {
			return nil, err
		}
	}

	// log.Printf("DecodeSerialGraph: %q", data)

	g = &SerialGraph{}
	if err := proto.Unmarshal(data, g); err != nil {
		return nil, err
	}
	return g, nil
}

// Bytes converts a serial graph to a gzipped graph (used for storage)
func (g *SerialGraph) Bytes() (data []byte, err error) {
	data, err = proto.Marshal(g)
	if err != nil {
		return nil, err
	}

	// log.Printf("      Graph.Bytes: %q", data)

	if shouldGZIP {
		if data, err = compress(data); err != nil {
			return nil, err
		}
	}
	return data, nil
}

// Simple GZIP decompression
func decompress(garbage []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(garbage))
	if err != nil {
		return nil, err
	}
	var buff bytes.Buffer
	if _, err := buff.ReadFrom(gz); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

// Simple GZIP compression
func compress(data []byte) ([]byte, error) {
	var buff bytes.Buffer
	gz := gzip.NewWriter(&buff)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Flush(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}
