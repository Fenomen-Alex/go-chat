package network

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"

	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
)

func TestIdentityPrivateKeyBytesExtraction(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	libp2pPriv, err := libp2pcrypto.UnmarshalEd25519PrivateKey(priv)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	raw := ed25519PrivateKeyBytes(libp2pPriv)
	if len(raw) != ed25519.PrivateKeySize {
		t.Fatalf("expected raw key length %d, got %d", ed25519.PrivateKeySize, len(raw))
	}
}

func TestSignAndVerifyMessageRoundTrip(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	libp2pPriv, err := libp2pcrypto.UnmarshalEd25519PrivateKey(priv)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	rawPriv := ed25519PrivateKeyBytes(libp2pPriv)
	if rawPriv == nil {
		t.Fatal("identityPrivateKey extraction returned nil")
	}

	msg := &Message{
		Type:      "message",
		SenderID:  "12D3Test",
		ChannelID: "ch_1",
		MessageID: "msg_1",
		Content:   "hello",
		Timestamp: time.Now().UnixMilli(),
	}
	SignMessage(rawPriv, msg)
	if len(msg.Signature) == 0 {
		t.Fatal("signature empty after signing")
	}
	if !VerifyMessageSignature(pub, msg) {
		t.Fatal("signature verification failed")
	}
}

func TestPeerstorePublicKeyIsRetrievable(t *testing.T) {
	_, alicePriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate alice key: %v", err)
	}
	aliceKey, err := libp2pcrypto.UnmarshalEd25519PrivateKey(alicePriv)
	if err != nil {
		t.Fatalf("unmarshal alice: %v", err)
	}
	_, bobPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate bob key: %v", err)
	}
	bobKey, err := libp2pcrypto.UnmarshalEd25519PrivateKey(bobPriv)
	if err != nil {
		t.Fatalf("unmarshal bob: %v", err)
	}

	ctx := context.Background()
	aliceHost, err := libp2p.New(libp2p.Identity(aliceKey), libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
	if err != nil {
		t.Fatalf("create alice host: %v", err)
	}
	bobHost, err := libp2p.New(libp2p.Identity(bobKey), libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
	if err != nil {
		t.Fatalf("create bob host: %v", err)
	}
	defer aliceHost.Close()
	defer bobHost.Close()

	if err := bobHost.Connect(ctx, peer.AddrInfo{ID: aliceHost.ID(), Addrs: aliceHost.Addrs()}); err != nil {
		t.Fatalf("connect bob to alice: %v", err)
	}

	pub, ok := GetPublicKeyFromPeerstore(aliceHost, bobHost.ID().String())
	if !ok || len(pub) != ed25519.PublicKeySize {
		t.Fatalf("could not retrieve bob public key from peerstore ok=%v len=%d", ok, len(pub))
	}

	msg := &Message{
		Type:      "message",
		SenderID:  bobHost.ID().String(),
		ChannelID: "ch_1",
		MessageID: "msg_1",
		Content:   "hello",
		Timestamp: time.Now().UnixMilli(),
	}
	SignMessage(bobPriv, msg)
	if len(msg.Signature) == 0 {
		t.Fatal("signature empty after signing")
	}
	if !VerifyMessageSignature(pub, msg) {
		t.Fatal("signature verification failed")
	}
}
