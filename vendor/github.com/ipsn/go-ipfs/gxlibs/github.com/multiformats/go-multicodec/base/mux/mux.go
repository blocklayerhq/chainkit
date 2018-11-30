package basemux

import (
	mc "github.com/ipsn/go-ipfs/gxlibs/github.com/multiformats/go-multicodec"
	mux "github.com/ipsn/go-ipfs/gxlibs/github.com/multiformats/go-multicodec/mux"

	b64 "github.com/ipsn/go-ipfs/gxlibs/github.com/multiformats/go-multicodec/base/b64"
	bin "github.com/ipsn/go-ipfs/gxlibs/github.com/multiformats/go-multicodec/base/bin"
	hex "github.com/ipsn/go-ipfs/gxlibs/github.com/multiformats/go-multicodec/base/hex"
)

func AllBasesMux() *mux.Multicodec {
	m := mux.MuxMulticodec([]mc.Multicodec{
		hex.Multicodec(),
		b64.Multicodec(),
		bin.Multicodec(),
	}, mux.SelectFirst)
	m.Wrap = false
	return m
}
