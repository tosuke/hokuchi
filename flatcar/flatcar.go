package flatcar

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"braces.dev/errtrace"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/tosuke/hokuchi"
	"golang.org/x/sync/errgroup"
)

type Fetcher struct {
	http *http.Client
	sema chan struct{}
}
type Option struct {
	RequestConcurrency int
	HTTP *http.Client
}

func New(option Option) *Fetcher {
	var sema chan struct{}
	if option.RequestConcurrency > 0 {
		sema = make(chan struct{}, option.RequestConcurrency)
	}

	hc := http.DefaultClient
	if option.HTTP != nil {
		hc = option.HTTP
	}

	return &Fetcher{
		http: hc,
		sema: sema,
	}
}

func (f *Fetcher) ResolveKey(ctx context.Context, channel, arch, version string) (Key, error) {
	k := Key{
		channel: channel,
		arch:    hokuchi.NormalizeArch(arch),
		version: version,
	}

	if !IsValidChannel(k.channel) {
		return Key{}, errtrace.New("invalid channel")
	}
	if !IsValidArch(k.arch) {
		return Key{}, errtrace.New("invalid arch")
	}
	if k.version == "current" {
		currentVer, err := f.fetchVersion(ctx, k)
		if err != nil {
			return Key{}, errtrace.Wrap(err)
		}
		k.version = currentVer
	}
	if !IsValidVersion(k.version) {
		return Key{}, errtrace.New("invalid version")
	}

	return k, nil
}

func (f *Fetcher) FetchKernel(ctx context.Context, w io.Writer, key Key) error {
	return f.fetchKernel(ctx, w, key)
}

func (f *Fetcher) fetchVersion(ctx context.Context, key Key) (string, error) {
	var versionBuf bytes.Buffer
	var versionSigBuf bytes.Buffer

	eg, egctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		if err := f.fetchData(egctx, &versionBuf, key, "/version.txt", 2048); err != nil {
			return errtrace.Wrap(err)
		}
		return nil
	})
	eg.Go(func() error {
		if err := f.fetchData(egctx, &versionSigBuf, key, "/version.txt.sig", 2048); err != nil {
			return errtrace.Wrap(err)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return "", errtrace.Wrap(err)
	}

	message := crypto.NewPlainMessage(versionBuf.Bytes())
	signature := crypto.NewPGPSignature(versionSigBuf.Bytes())

	if err := signKeyring.VerifyDetached(message, signature, crypto.GetUnixTime()); err != nil {
		return "", errtrace.Wrap(err)
	}

	sc := bufio.NewScanner(&versionBuf)
	for sc.Scan() {
		line := sc.Text()
		if after, ok := strings.CutPrefix(line, "FLATCAR_VERSION="); ok {
			version := strings.TrimSpace(after)
			return version, nil
		}
	}
	return "", errtrace.New("failed to find version")
}

func (f *Fetcher) fetchKernel(ctx context.Context, w io.Writer, key Key) error {
	eg, ectx := errgroup.WithContext(ctx)
	pr, pw := io.Pipe()

	eg.Go(func() error {
		writer := io.MultiWriter(w, pw)
		defer pw.Close()
		if err := f.fetchData(ectx, writer, key, "/flatcar_production_pxe.vmlinuz", 0); err != nil {
			return errtrace.Wrap(err)
		}
		return nil
	})
	eg.Go(func() error {

		var kernelSigBuf bytes.Buffer
		if err := f.fetchData(ctx, &kernelSigBuf, key, "/flatcar_production_pxe.vmlinuz.sig", 2048); err != nil {
			return errtrace.Wrap(err)
		}
		signature := crypto.NewPGPSignature(kernelSigBuf.Bytes())

		if err := signKeyring.VerifyDetachedStream(pr, signature, crypto.GetUnixTime()); err != nil {
			return errtrace.Wrap(err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return errtrace.Wrap(err)
	}

	return nil
}

func (f *Fetcher) fetchData(ctx context.Context, w io.Writer, key Key, subpath string, limit int64) error {
	url := key.baseURL() + subpath
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return errtrace.Wrap(err)
	}

	if f.sema != nil {
		f.sema <- struct{}{}
		defer func() {
			<-f.sema
		}()
	}

	slog.DebugContext(ctx, "request", slog.String("url", url))
	resp, err := f.http.Do(req)
	if err != nil {
		return errtrace.Wrap(err)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	slog.DebugContext(ctx, "processing response", slog.String("url", url))
	if resp.StatusCode != http.StatusOK {
		return errtrace.Errorf("invalid status from flatcar: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var r io.Reader = resp.Body
	if limit > 0 {
		r = io.LimitReader(r, limit)
	}
	if _, err := io.Copy(w, r); err != nil {
		return errtrace.Wrap(err)
	}
	return nil
}

//go:embed Flatcar_Image_Signing_key.asc
var flatcarGPGKey string

var signKeyring *crypto.KeyRing = mustKeyringFromArmored(flatcarGPGKey)

func mustKeyringFromArmored(pubKey string) *crypto.KeyRing {
	pubKeyObj, err := crypto.NewKeyFromArmored(pubKey)
	if err != nil {
		panic(err)
	}
	keyring, err := crypto.NewKeyRing(pubKeyObj)
	if err != nil {
		panic(err)
	}
	return keyring
}
