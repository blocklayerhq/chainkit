package muxcodec

import (
	mc "github.com/ipsn/go-ipfs/gxlibs/github.com/multiformats/go-multicodec"
	cbor "github.com/ipsn/go-ipfs/gxlibs/github.com/multiformats/go-multicodec/cbor"
	json "github.com/ipsn/go-ipfs/gxlibs/github.com/multiformats/go-multicodec/json"
)

func StandardMux() *Multicodec {
	return MuxMulticodec([]mc.Multicodec{
		cbor.Multicodec(),
		json.Multicodec(false),
		json.Multicodec(true),
	}, SelectFirst)
}
