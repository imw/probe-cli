package quicdialer_test

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"testing"

	"github.com/lucas-clemente/quic-go"
	"github.com/ooni/probe-cli/v3/internal/engine/netx/errorx"
	"github.com/ooni/probe-cli/v3/internal/engine/netx/quicdialer"
)

func TestErrorWrapperFailure(t *testing.T) {
	ctx := context.Background()
	d := quicdialer.ErrorWrapperDialer{
		Dialer: MockDialer{Sess: nil, Err: io.EOF}}
	sess, err := d.DialContext(
		ctx, "udp", "www.google.com:443", &tls.Config{}, &quic.Config{})
	if sess != nil {
		t.Fatal("expected a nil sess here")
	}
	errorWrapperCheckErr(t, err, errorx.QUICHandshakeOperation)
}

func errorWrapperCheckErr(t *testing.T, err error, op string) {
	if !errors.Is(err, io.EOF) {
		t.Fatal("expected another error here")
	}
	var errWrapper *errorx.ErrWrapper
	if !errors.As(err, &errWrapper) {
		t.Fatal("cannot cast to ErrWrapper")
	}
	if errWrapper.Operation != op {
		t.Fatal("unexpected Operation")
	}
	if errWrapper.Failure != errorx.FailureEOFError {
		t.Fatal("unexpected failure")
	}
}

func TestErrorWrapperInvalidCertificate(t *testing.T) {
	nextprotos := []string{"h3"}
	servername := "example.com"
	tlsConf := &tls.Config{
		NextProtos: nextprotos,
		ServerName: servername,
	}

	dlr := quicdialer.ErrorWrapperDialer{Dialer: &quicdialer.SystemDialer{
		QUICListener: &quicdialer.QUICListenerStdlib{},
	}}
	// use Google IP
	sess, err := dlr.DialContext(context.Background(), "udp",
		"216.58.212.164:443", tlsConf, &quic.Config{})
	if err == nil {
		t.Fatal("expected an error here")
	}
	if sess != nil {
		t.Fatal("expected nil sess here")
	}
	if err.Error() != errorx.FailureSSLInvalidCertificate {
		t.Fatal("unexpected failure")
	}
}

func TestErrorWrapperSuccess(t *testing.T) {
	ctx := context.Background()
	tlsConf := &tls.Config{
		NextProtos: []string{"h3"},
		ServerName: "www.google.com",
	}
	d := quicdialer.ErrorWrapperDialer{Dialer: quicdialer.SystemDialer{
		QUICListener: &quicdialer.QUICListenerStdlib{},
	}}
	sess, err := d.DialContext(ctx, "udp", "216.58.212.164:443", tlsConf, &quic.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if sess == nil {
		t.Fatal("expected non-nil sess here")
	}
}
